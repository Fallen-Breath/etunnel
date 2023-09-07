package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/protocol/header"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	log "github.com/sirupsen/logrus"
	"net"
)

func (t *Tunnel) runClient() {
	t.reloadClient()
	t.cliWg.Wait()
}

func (t *Tunnel) reloadClient() {
	t.cliMutex.Lock()
	defer t.cliMutex.Unlock()

	expectedTunnels := make(map[string]bool)
	for _, tun := range t.conf.Tunnels {
		expectedTunnels[tun] = true
		if _, ok := t.cliHandlers[tun]; ok {
			// already exists, skipped
			continue
		}

		handler, err := newTunnelHandler(t.conf.Server, tun, t.cipher)
		if err != nil {
			log.Errorf("invalid tunnel %s", tun)
			continue
		}

		log.Infof("Creating tunnel %s", tun)
		t.cliWg.Add(1)
		t.cliHandlers[tun] = handler
		go func() {
			defer t.cliWg.Done()
			handler.Start()
		}()
	}

	var tunnelsToRemove []string
	for tun := range t.cliHandlers {
		if _, ok := expectedTunnels[tun]; !ok {
			tunnelsToRemove = append(tunnelsToRemove, tun)
		}
	}

	for _, tun := range tunnelsToRemove {
		if handler, ok := t.cliHandlers[tun]; ok {
			log.Infof("Removing tunnel %s", tun)
			handler.Stop()
			delete(t.cliHandlers, tun)
		}
	}
}

type tunnelHandlerImpl struct {
	tunnel     string
	serverAddr string
	protocol   string
	listen     string
	target     string
	cipher     sscore.Cipher
	stopCh     chan int
}

func (t *tunnelHandlerImpl) GetDefinition() string {
	return t.tunnel
}

func newTunnelHandler(serverAddr string, tun string, cipher sscore.Cipher) (tunnelHandler, error) {
	protocol, listen, target, err := config.ParseTunnel(tun)
	if err != nil { // should already be validated in config.CreateConfigOrDie
		return nil, err
	}
	return &tunnelHandlerImpl{
		tunnel:     tun,
		serverAddr: serverAddr,
		protocol:   protocol,
		listen:     listen,
		target:     target,
		cipher:     cipher,
		stopCh:     make(chan int, 1),
	}, nil
}

var _ tunnelHandler = &tunnelHandlerImpl{}

func (t *tunnelHandlerImpl) Start() {
	switch t.protocol {
	case "tcp":
		t.startTcpTunnel()
	case "udp":
		t.startUdpTunnel()
	default:
		log.Errorf("Invalid protocol %s", t.protocol)
	}
}

func (t *tunnelHandlerImpl) Stop() {
	t.stopCh <- 0
}

// reference: github.com/shadowsocks/go-shadowsocks2/tcp.go tcpLocal
func (t *tunnelHandlerImpl) startTcpTunnel() {
	listener, err := net.Listen(t.protocol, t.listen)
	if err != nil {
		log.Errorf("Failed to listen on %s: %v", t.listen, err)
		return
	}
	defer doClose(listener)
	log.Infof("TCP tunnel start: -> %s -> %s -> %s", t.listen, t.serverAddr, t.target)
	go func() {
		<-t.stopCh
		doClose(listener)
	}()

	head := header.Header{
		Protocol: t.protocol,
		Target:   t.target,
	}

	for {
		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept: %v", err)
			continue
		}

		log.Infof("Accepted connection from %s", cliConn.RemoteAddr())

		go func() {
			defer doClose(cliConn)
			log.Infof("Dial server %s start", t.serverAddr)
			svrConn, err := net.Dial("tcp", t.serverAddr)
			if err != nil {
				log.Errorf("Failed to connect to server %s: %v", t.serverAddr, err)
				return
			}
			defer doClose(svrConn)
			log.Infof("Dial server %s done", t.serverAddr)

			// TODO: TCP cork support
			svrConn = conn.NewEncryptedStreamConn(svrConn.(conn.StreamConn), t.cipher)

			if err = head.MarshalTo(svrConn); err != nil {
				log.Errorf("Failed to write header: %v", err)
				return
			}

			log.Infof("TCP relay start: %s <-[ %s <-> %s ]-> %s", cliConn.RemoteAddr(), t.listen, t.serverAddr, t.target)
			relayConnection(cliConn, svrConn)
			log.Infof("TCP relay end: %s <-[ %s <-> %s ]-> %s", cliConn.RemoteAddr(), t.listen, t.serverAddr, t.target)
		}()
	}

}

// reference: github.com/shadowsocks/go-shadowsocks2/udp.go udpLocal
func (t *tunnelHandlerImpl) startUdpTunnel() {
	// TODO
	log.Errorf("UDP tunnel has not implemented yet")
}
