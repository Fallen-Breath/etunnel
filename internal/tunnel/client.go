package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
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
	serverAddr string
	protocol   string
	listen     string
	target     string
	cipher     sscore.Cipher
	stopCh     chan int
}

func newTunnelHandler(serverAddr string, tun string, cipher sscore.Cipher) (tunnelHandler, error) {
	protocol, listen, target, err := config.ParseTunnel(tun)
	if err != nil { // should already be validated in config.CreateConfigOrDie
		return nil, err
	}
	return &tunnelHandlerImpl{
		serverAddr: serverAddr,
		protocol:   protocol,
		listen:     listen,
		target:     target,
		cipher:     cipher,
		stopCh:     make(chan int),
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
	defer closeWhatever(listener)
	log.Infof("TCP tunnel start: -> %s -> %s -> %s", t.listen, t.serverAddr, t.target)

	targetSock := socks.ParseAddr(t.target)
	if targetSock == nil {
		log.Errorf("Failed to parse socket addr from %s", t.target)
		return
	}

	for {
		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept: %v", err)
			continue
		}

		go func() {
			defer closeWhatever(cliConn)
			svrConn, err := net.Dial("tcp", t.serverAddr)
			if err != nil {
				log.Errorf("Failed to connect to server %s: %v", t.serverAddr, err)
				return
			}
			defer closeWhatever(svrConn)

			// TODO: TCP cork support
			svrConn = t.cipher.StreamConn(svrConn)

			if _, err = svrConn.Write(targetSock); err != nil {
				log.Errorf("Failed to send target address: %v", err)
				return
			}

			log.Infof("TCP relay start: %s <-> %s <-> %s <-> %s", cliConn.RemoteAddr(), t.listen, t.serverAddr, t.target)
			relayTcp(cliConn, svrConn)
		}()
	}

}

// reference: github.com/shadowsocks/go-shadowsocks2/udp.go udpLocal
func (t *tunnelHandlerImpl) startUdpTunnel() {
	// TODO
	log.Errorf("UDP tunnel has not implemented yet")
}
