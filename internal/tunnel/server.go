package tunnel

import (
	"fmt"
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
		go t.handleConnection(cliConn.(conn.StreamConn), log.WithField("cid", cid))
		cid++
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
	if protoMeta, _ := proto.GetProtocolMeta(tun.Protocol); protoMeta.Kind != reqHead.Kind {
		sendResponse(header.CodeBadKind)
		return
	}

	targetConn, err := net.Dial(tun.Protocol, tun.Target)
	if err != nil {
		logger.Errorf("Failed to connect to target %s://%s: %v", tun.Protocol, tun.Target, err)
		return
	}
	defer func() { _ = targetConn.Close() }()

	sendResponse(header.CodeOk)
	_ = originConn.SetDeadline(time.Time{})

	logger.Infof("Relay start %s <-> %s", cliConn.RemoteAddr(), tun.Target)
	flow := ""
	switch tun.Protocol {
	case proto.Tcp:
		send, recv := relayConnection(cliConn, targetConn, logger)
		if log.GetLevel() >= log.DebugLevel {
			flow = fmt.Sprintf(" (send %d, recv %d)", send, recv)
		}

	case proto.Udp:
		size := relayUdpConnection(cliConn, targetConn, logger)
		if log.GetLevel() >= log.DebugLevel {
			flow = fmt.Sprintf(" (packet size %d)", size)
		}

	default:
		logger.Errorf("Unsupported protocol to relay: %s", tun.Protocol)
		return
	}
	logger.Infof("Relay end %s -> %s%s", cliConn.RemoteAddr(), tun.Target, flow)
}
