package tunnel

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/proto/header"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

// reference: github.com/shadowsocks/go-shadowsocks2/tcp.go tcpLocal
func (t *tunnelHandlerImpl) runStreamTunnel() {
	listener, err := net.Listen(t.protocol, t.listen)
	if err != nil {
		log.Errorf("Failed to listen on %s: %v", t.listen, err)
		return
	}
	defer func() { _ = listener.Close() }()
	log.Infof("Stream tunnel (%s) start: -> %s -> %s -> %s", t.protocol, t.listen, t.serverAddr, t.target)
	go func() {
		<-t.stopCh
		_ = listener.Close()
	}()

	for {
		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept: %v", err)
			continue
		}

		log.Debugf("Accepted connection from %s", cliConn.RemoteAddr())

		go t.handleStreamConnection(cliConn.(conn.StreamConn))
	}
}

func (t *tunnelHandlerImpl) handleStreamConnection(cliConn conn.StreamConn) {
	defer func() { _ = cliConn.Close() }()

	sc, err := net.Dial("tcp", t.serverAddr)
	if err != nil {
		log.Errorf("Failed to connect to server %s: %v", t.serverAddr, err)
		return
	}
	svrConn := sc.(conn.StreamConn)
	defer func() { _ = svrConn.Close() }()

	if t.corking {
		svrConn = conn.NewTimedCorkConn(svrConn, 10*time.Millisecond, 1280)
	}
	svrConn = conn.NewEncryptedStreamConn(svrConn, t.cipher)

	head := header.Header{
		Protocol: t.protocol,
		Target:   t.target,
	}
	if err = head.MarshalTo(svrConn); err != nil {
		log.Errorf("Failed to write header: %v", err)
		return
	}

	log.Infof("TCP relay start: %s -[ %s -> %s ]-> %s", cliConn.RemoteAddr(), t.listen, t.serverAddr, t.target)
	send, recv := relayConnection(cliConn, svrConn)
	flow := ""
	if log.StandardLogger().Level >= log.DebugLevel {
		flow = fmt.Sprintf(" (send %d, recv %d)", send, recv)
	}
	log.Infof("TCP relay end: %s -[ %s -> %s ]-> %s%s", cliConn.RemoteAddr(), t.listen, t.serverAddr, t.target, flow)
}
