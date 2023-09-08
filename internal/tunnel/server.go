package tunnel

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/conn"
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
		go t.handleConnection(cliConn.(conn.StreamConn))
	}
}
func (t *Tunnel) handleConnection(cliConn conn.StreamConn) {
	defer func() { _ = cliConn.Close() }()
	originConn := cliConn

	if t.conf.Cork {
		cliConn = conn.NewTimedCorkConn(cliConn, 10*time.Millisecond, 1280)
	}
	cliConn = conn.NewEncryptedStreamConn(originConn, t.cipher)

	var head header.Header
	if err := head.UnmarshalFrom(cliConn); err != nil {
		log.Errorf("Failed to read header: %v", err)
		// drain originConn to avoid leaking server behavioral features
		// see https://www.ndss-symposium.org/ndss-paper/detecting-probe-resistant-proxies/
		_ = originConn.SetDeadline(time.Now().Add(30 * time.Second)) // TODO: find a nice way to close connection
		_, err = io.Copy(io.Discard, originConn)
		if err != nil {
			log.Errorf("Discard error: %v", err)
		}
		return
	}

	target := head.Target
	targetConn, err := net.Dial(head.Protocol, target)
	if err != nil {
		log.Errorf("Failed to connect to target %s://%s: %v", head.Protocol, head.Target, err)
		return
	}
	defer func() { _ = targetConn.Close() }()

	log.Infof("Relay start %s <-> %s", cliConn.RemoteAddr(), target)
	send, recv := relayConnection(cliConn, targetConn)
	flow := ""
	if log.StandardLogger().Level >= log.DebugLevel {
		flow = fmt.Sprintf(" (send %d, recv %d)", send, recv)
	}
	log.Infof("Relay end %s -> %s%s", cliConn.RemoteAddr(), target, flow)
}
