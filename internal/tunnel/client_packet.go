package tunnel

import (
	"math"
	"net"
)

func (t *tunnelHandlerImpl) runPacketTunnel() {
	if true {
		t.logger.Errorf("Packet tunnel has not implemented yet")
		return
	}

	listener, err := net.ListenPacket(t.tunnel.Protocol, t.tunnel.Listen)
	if err != nil {
		t.logger.Errorf("Failed to listen on %s: %v", t.tunnel.Listen, err)
		return
	}
	defer func() { _ = listener.Close() }()
	t.logger.Infof("Packet tunnel (%s) start: -> %s -> %s -> %s", t.tunnel.Protocol, t.tunnel.Listen, t.serverAddr, t.tunnel.Target)
	go func() {
		<-t.stopCh
		_ = listener.Close()
	}()

	buf := make([]byte, math.MaxUint16)

	for {
		n, addr, err := listener.ReadFrom(buf)
		if err != nil {
			t.logger.Errorf("Failed to accept: %v", err)
			continue
		}

		t.logger.Debugf("Accepted packet with size %d from %s", n, addr)

		t.handlePacketConnection(buf[:n], addr)
	}
}

func (t *tunnelHandlerImpl) handlePacketConnection(packet []byte, addr net.Addr) {
}
