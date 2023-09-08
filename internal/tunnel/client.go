package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	log "github.com/sirupsen/logrus"
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
		expectedTunnels[tun.GetDefinition()] = true
		if _, ok := t.cliHandlers[tun.GetDefinition()]; ok {
			// already exists, skipped
			continue
		}

		handler, err := newTunnelHandler(t.conf, tun, t.cipher)
		if err != nil {
			log.Errorf("invalid tunnel %s", tun)
			continue
		}

		log.Infof("Creating tunnel %+v", tun)
		t.cliWg.Add(1)
		t.cliHandlers[tun.GetDefinition()] = handler
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
	cipher sscore.Cipher
	stopCh chan int
	tunnel config.Tunnel
	logger *log.Entry

	// configs
	serverAddr string
	corking    bool
}

func (t *tunnelHandlerImpl) GetDefinition() string {
	return t.tunnel.GetDefinition()
}

func newTunnelHandler(conf *config.Config, tun config.Tunnel, cipher sscore.Cipher) (tunnelHandler, error) {
	return &tunnelHandlerImpl{
		cipher: cipher,
		stopCh: make(chan int, 1),
		tunnel: tun,
		logger: log.WithField("tunnel", tun.Name),

		serverAddr: conf.Server,
		corking:    conf.Cork,
	}, nil
}

var _ tunnelHandler = &tunnelHandlerImpl{}

func (t *tunnelHandlerImpl) Start() {
	switch t.tunnel.Protocol {
	case proto.Tcp:
		t.runStreamTunnel()
	case proto.Udp:
		t.runPacketTunnel()
	default:
		t.logger.Errorf("Unsupported protocol %s", t.tunnel.Protocol)
	}
}

func (t *tunnelHandlerImpl) Stop() {
	t.stopCh <- 0
}
