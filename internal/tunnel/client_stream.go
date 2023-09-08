package tunnel

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/proto/header"
	log "github.com/sirupsen/logrus"
	"net"
)

// reference: github.com/shadowsocks/go-shadowsocks2/tcp.go tcpLocal
func (t *tunnelHandlerImpl) runStreamTunnel() {
	listener, err := net.Listen(t.tunnel.Protocol, t.tunnel.Listen)
	if err != nil {
		t.logger.Errorf("Failed to listen on %s: %v", t.tunnel.Listen, err)
		return
	}
	defer func() { _ = listener.Close() }()
	t.logger.Infof("Stream tunnel (%s) start: -> %s -> %s -> %s", t.tunnel.Protocol, t.tunnel.Listen, t.serverAddr, t.tunnel.Target)
	go func() {
		<-t.stopCh
		_ = listener.Close()
	}()

	cid := 0
	for {
		cliConn, err := listener.Accept()
		if err != nil {
			t.logger.Errorf("Failed to accept: %v", err)
			continue
		}

		t.logger.Debugf("Accepted connection from %s", cliConn.RemoteAddr())

		go t.handleStreamConnection(cliConn.(conn.StreamConn), t.logger.WithField("cid", cid))
		cid++
	}
}

func (t *tunnelHandlerImpl) handleStreamConnection(cliConn conn.StreamConn, logger *log.Entry) {
	defer func() { _ = cliConn.Close() }()

	head := header.Header{
		Protocol: t.tunnel.Protocol,
		Target:   t.tunnel.Target,
	}
	svrConn, err := t.connectToServer(&head)
	if err != nil {
		logger.Errorf("Failed to connect to the server: %v", err)
	}

	logger.Infof("Relay start: %s --[ %s -> %s ]-> %s", cliConn.RemoteAddr(), t.tunnel.Listen, t.serverAddr, t.tunnel.Target)
	send, recv := relayConnection(cliConn, svrConn, logger)
	flow := ""
	if log.GetLevel() >= log.DebugLevel {
		flow = fmt.Sprintf(" (send %d, recv %d)", send, recv)
	}
	logger.Infof("Relay end: %s --[ %s -> %s ]-> %s%s", cliConn.RemoteAddr(), t.tunnel.Target, t.serverAddr, t.tunnel.Target, flow)
}
