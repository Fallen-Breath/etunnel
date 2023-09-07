package tunnel

import (
	log "github.com/sirupsen/logrus"
	"math"
	"net"
)

func (t *tunnelHandlerImpl) runPacketTunnel() {
	if true {
		log.Errorf("Packet tunnel has not implemented yet")
		return
	}

	listener, err := net.ListenPacket(t.protocol, t.listen)
	if err != nil {
		log.Errorf("Failed to listen on %s: %v", t.listen, err)
		return
	}
	defer func() { _ = listener.Close() }()
	log.Infof("Packet tunnel (%s) start: -> %s -> %s -> %s", t.protocol, t.listen, t.serverAddr, t.target)
	go func() {
		<-t.stopCh
		_ = listener.Close()
	}()

	buf := make([]byte, math.MaxUint16)

	for {
		n, addr, err := listener.ReadFrom(buf)
		if err != nil {
			log.Errorf("Failed to accept: %v", err)
			continue
		}

		log.Debugf("Accepted packet with size %d from %s", n, addr)

		t.handlePacketConnection(buf[:n], addr)
	}
}

func (t *tunnelHandlerImpl) handlePacketConnection(packet []byte, addr net.Addr) {
}
