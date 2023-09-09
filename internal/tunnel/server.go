package tunnel

import (
	"errors"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/Fallen-Breath/etunnel/internal/proto/header"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	*base
}

func newServer(base *base) (ITunnel, error) {
	return &Server{
		base: base,
	}, nil
}

// Start reference: github.com/shadowsocks/go-shadowsocks2/tcp.go tcpRemote
func (t *Server) start() {
	listener, err := net.Listen("tcp", t.conf.Listen)
	if err != nil {
		// server fails hard
		log.Fatalf("failed to listen on %s: %v", t.conf.Listen, err)
		return
	}

	var once sync.Once
	defer once.Do(func() { _ = listener.Close() })

	log.Infof("Listening on %s", t.conf.Listen)
	go func() {
		<-t.stopCh
		log.Infof("Closing connection listener")
		once.Do(func() { _ = listener.Close() })
	}()

	cid := 0
	for {
		cliConn, err := listener.Accept()
		if t.stopped.Load() {
			_ = cliConn.Close()
			break
		}
		if err != nil {
			log.Errorf("Failed to accept: %v", err)
			continue
		}

		log.Debugf("Accepted connection from %s", cliConn.RemoteAddr())
		cid_ := cid
		cid++
		go t.handleConnection(cliConn.(conn.StreamConn), log.WithField("cid", cid_))
	}
}

func (t *Server) handleConnection(cliConn conn.StreamConn, logger *log.Entry) {
	defer func() { _ = cliConn.Close() }()
	originConn := cliConn
	_ = originConn.SetDeadline(time.Now().Add(30 * time.Second))

	if t.conf.Cork {
		cliConn = conn.NewTimedCorkConn(cliConn, 10*time.Millisecond, 1280)
	}
	cliConn = conn.NewEncryptedStreamConn(originConn, t.cipher)

	var reqHead header.ReqHead
	if err := reqHead.UnmarshalFrom(cliConn); err != nil {
		logger.Errorf("Failed to read header: %v", err)
		// drain originConn to avoid leaking server behavioral features
		// see https://www.ndss-symposium.org/ndss-paper/detecting-probe-resistant-proxies/
		if _, err = io.Copy(io.Discard, originConn); err != nil {
			logger.Errorf("Discard error: %v", err)
		}
		return
	}

	sendResponse := func(code uint8) {
		rspHead := header.RspHead{Code: code}
		if err := rspHead.MarshalTo(cliConn); err != nil {
			logger.Errorf("Send response header error: %v", err)
		}
	}

	tun, ok := t.conf.Tunnels[reqHead.Id]
	if !ok {
		sendResponse(header.CodeBadId)
		return
	}
	protoMeta, _ := proto.GetProtocolMeta(tun.Protocol)
	if protoMeta.Kind != reqHead.Kind {
		sendResponse(header.CodeBadKind)
		return
	}
	logger = logger.WithField("tid", tun.Id)

	targetConn, err := net.Dial(tun.Protocol, tun.Target)
	if err != nil {
		logger.Errorf("Failed to connect to target %s://%s: %v", tun.Protocol, tun.Target, err)
		return
	}
	defer func() { _ = targetConn.Close() }()

	sendResponse(header.CodeOk)
	_ = originConn.SetDeadline(time.Time{})

	logger.Infof("Relay start %s <-> %s", cliConn.RemoteAddr(), tun.Target)
	var send, recv int64
	switch protoMeta.Kind {
	case proto.KindStream:
		send, recv, _ = relayStreamConnection(cliConn, targetConn, logger)

	case proto.KindPacket:
		send, recv, err = t.relayPacketConnection(cliConn, targetConn, tun, logger)
		if err != nil {
			logger.Errorf("Relay packet connection error: %v", err)
		}

	default:
		logger.Errorf("Unsupported protocol to relay: %s", tun.Protocol)
		return
	}
	logger.Infof("Relay end %s -> %s%s", cliConn.RemoteAddr(), tun.Target, makeFlowTail(send, recv))
}

func (t *Server) relayPacketConnection(cliConn conn.StreamConn, targetConn net.Conn, tun *config.Tunnel, logger *log.Entry) (send, recv int64, _ error) {
	buf, err := receivePacket(cliConn)
	if err != nil {
		return 0, 0, fmt.Errorf("receive packet from client failed: %v", err)
	}

	if _, err := targetConn.Write(buf); err != nil {
		return 0, 0, fmt.Errorf("send packet to server failed: %v", err)
	}

	send, recv = int64(len(buf)), 0
	buf = make([]byte, proto.MaxUdpPacketSize)

	for {
		if err := targetConn.SetReadDeadline(time.Now().Add(tun.Meta.TimeToLive)); err != nil {
			return send, recv, fmt.Errorf("SetReadDeadline failed: %v", err)
		}

		n, err := targetConn.Read(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				logger.Debugf("ttl timeout, closing connection")
				return send, recv, nil
			} else {
				return send, recv, fmt.Errorf("receive packet from server failed: %v", err)
			}
		}
		recv += int64(n)
		logger.Debugf("Received packet with size %d, forwarding", n)

		if err := sendPacket(cliConn, buf[:n]); err != nil {
			return send, recv, fmt.Errorf("send packet to client failed: %v", err)
		}

		if !tun.Meta.KeepAlive {
			logger.Debugf("no keep alive, closing connection")
			return send, recv, nil
		}
	}
}
