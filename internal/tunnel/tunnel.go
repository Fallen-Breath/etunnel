package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	"sync"
	"sync/atomic"
)

type Tunnel struct {
	conf    *config.Config
	cipher  sscore.Cipher
	stopCh  chan int
	stopped atomic.Bool

	// use in client only
	cliWg       sync.WaitGroup
	cliMutex    sync.Mutex
	cliHandlers map[string]tunnelHandler // tunnel definition -> handler
}

type tunnelHandler interface {
	GetDefinition() string
	Start()
	Stop()
}

func NewTunnel(conf *config.Config) (*Tunnel, error) {
	cipher, err := sscore.PickCipher(conf.Crypt, []byte{}, conf.Key)
	if err != nil {
		return nil, err
	}

	t := &Tunnel{
		conf:        conf,
		cipher:      cipher,
		stopCh:      make(chan int, 1),
		cliHandlers: make(map[string]tunnelHandler),
	}
	return t, nil
}

func (t *Tunnel) Start() {
	switch t.conf.Mode {
	case config.ModeClient:
		t.runClient()
	case config.ModeServer:
		t.runServer()
	}
}

func (t *Tunnel) Stop() {
	t.stopped.Store(true)
	t.stopCh <- 0
}

// Reload support reload clientside tunnels only
func (t *Tunnel) Reload() {
	if t.conf.Mode == config.ModeClient {
		t.reloadClient()
	}
}
