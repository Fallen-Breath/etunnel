package tunnel

import (
	"encoding/binary"
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

// reference: github.com/shadowsocks/go-shadowsocks2/tcp.go tcpRemote
func (t *Tunnel) runServer() {
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

func (t *Tunnel) handleConnection(cliConn conn.StreamConn, logger *log.Entry) {
	defer func() { _ = cliConn.Close() }()
	originConn := cliConn

	if t.conf.Cork {
		cliConn = conn.NewTimedCorkConn(cliConn, 10*time.Millisecond, 1280)
	}
	cliConn = conn.NewEncryptedStreamConn(originConn, t.cipher)

	var head header.Header
	if err := head.UnmarshalFrom(cliConn); err != nil {
		logger.Errorf("Failed to read header: %v", err)
		// drain originConn to avoid leaking server behavioral features
		// see https://www.ndss-symposium.org/ndss-paper/detecting-probe-resistant-proxies/
		_ = originConn.SetDeadline(time.Now().Add(30 * time.Second)) // TODO: find a nice way to close connection
		_, err = io.Copy(io.Discard, originConn)
		if err != nil {
			logger.Errorf("Discard error: %v", err)
		}
		return
	}

	target := head.Target
	targetConn, err := net.Dial(head.Protocol, target)
	if err != nil {
		logger.Errorf("Failed to connect to target %s://%s: %v", head.Protocol, head.Target, err)
		return
	}
	defer func() { _ = targetConn.Close() }()

	logger.Infof("Relay start %s <-> %s", cliConn.RemoteAddr(), target)
	flow := ""
	switch head.Protocol {
	case proto.Tcp:
		send, recv := relayConnection(cliConn, targetConn, logger)
		if logger.Level >= log.DebugLevel {
			flow = fmt.Sprintf(" (send %d, recv %d)", send, recv)
		}

	case proto.Udp:
		size := relayUdpConnection(cliConn, targetConn, logger)
		if logger.Level >= log.DebugLevel {
			flow = fmt.Sprintf(" (packet size %d)", size)
		}

	default:
		logger.Errorf("Unsupported protocol to relay: %s", head.Protocol)
		return
	}
	logger.Infof("Relay end %s -> %s%s", cliConn.RemoteAddr(), target, flow)
}

func relayUdpConnection(cliConn conn.StreamConn, targetConn net.Conn, logger *log.Entry) int64 {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(cliConn, buf); err != nil {
		logger.Errorf("Failed to receive udp packet length")
		return 0
	}

	length := binary.BigEndian.Uint16(buf)
	buf = make([]byte, length)
	if _, err := io.ReadFull(cliConn, buf); err != nil {
		logger.Errorf("Failed to receive udp packet body")
		return 0
	}

	if _, err := targetConn.Write(buf); err != nil {
		logger.Errorf("Failed to send udp packet")
		return 0
	}

	return int64(length)
}
