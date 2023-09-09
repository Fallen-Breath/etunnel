package tunnel

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/constants"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Manager interface {
	Run()
}

type managerImpl struct {
	tunnel ITunnel
}

var _ Manager = &managerImpl{}

func NewManager(conf *config.Config) (Manager, error) {
	cipher, err := sscore.PickCipher(strings.ToUpper(conf.Crypt), []byte{}, conf.Key)
	if err != nil {
		return nil, err
	}

	base := &base{
		conf:   conf,
		cipher: cipher,
		stopCh: make(chan int, 1),
	}

	var t ITunnel
	switch conf.Mode {
	case config.ModeServer:
		t, err = newServer(base)
		if err != nil {
			return nil, err
		}
	case config.ModeClient:
		t, err = newClient(base)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported mode %s", conf.Mode)
	}

	return &managerImpl{
		tunnel: t,
	}, nil
}

func (m *managerImpl) Run() {
	reloadCh := make(chan os.Signal, 1)
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reloadCh, syscall.SIGHUP)

	go m.tunnel.start()
	go func() {
		for {
			switch <-reloadCh {
			case constants.SignalReload:
				log.Infof("%s reloading", constants.Name)
				m.tunnel.reload()
			case nil:
				return
			}
		}
	}()

	sig := <-stopCh
	log.Infof("Terminating by signal %s", sig)
	reloadCh <- nil
	m.tunnel.stop()
	log.Infof("%s stopped", constants.Name)
}
