package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/constants"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
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

func (t *Tunnel) Run() {
	reloadCh := make(chan os.Signal, 1)
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reloadCh, syscall.SIGHUP)

	go t.Start()
	go func() {
		for {
			switch <-reloadCh {
			case constants.SignalReload:
				log.Infof("%s reloading", constants.Name)
				t.Reload()
			case nil:
				return
			}
		}
	}()

	sig := <-stopCh
	log.Infof("Terminating by signal %s", sig)
	reloadCh <- nil
	t.Stop()
	log.Infof("%s stopped", constants.Name)
}

func (t *Tunnel) Start() {
	switch t.conf.Mode {
	case config.ModeServer:
		t.runServer()
	case config.ModeClient:
		t.runClient()
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
