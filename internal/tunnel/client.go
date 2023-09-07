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
		expectedTunnels[tun] = true
		if _, ok := t.cliHandlers[tun]; ok {
			// already exists, skipped
			continue
		}

		handler, err := newTunnelHandler(t.conf, tun, t.cipher)
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
	cipher sscore.Cipher
	stopCh chan int
	tunnel string

	// configs
	serverAddr string
	protocol   string
	listen     string
	target     string
	corking    bool
}

func (t *tunnelHandlerImpl) GetDefinition() string {
	return t.tunnel
}

func newTunnelHandler(conf *config.Config, tun string, cipher sscore.Cipher) (tunnelHandler, error) {
	protocol, listen, target, err := config.ParseTunnel(tun)
	if err != nil { // should already be validated in config.CreateConfigOrDie
		return nil, err
	}
	return &tunnelHandlerImpl{
		cipher: cipher,
		stopCh: make(chan int, 1),
		tunnel: tun,

		serverAddr: conf.Server,
		protocol:   protocol,
		listen:     listen,
		target:     target,
		corking:    conf.Cork,
	}, nil
}

var _ tunnelHandler = &tunnelHandlerImpl{}

func (t *tunnelHandlerImpl) Start() {
	switch t.protocol {
	case proto.Tcp, proto.Unix:
		t.runStreamTunnel()
	case proto.Udp, proto.UnixGram:
		t.runPacketTunnel()
	default:
		log.Errorf("Invalid protocol %s", t.protocol)
	}
}

func (t *tunnelHandlerImpl) Stop() {
	t.stopCh <- 0
}
