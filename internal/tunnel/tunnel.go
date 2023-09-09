package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	"sync"
	"sync/atomic"
)

type ITunnel interface {
	start()
	reload()
	stop()
}

type base struct {
	conf    *config.Config
	mutex   sync.RWMutex // TODO: make conf access concurrency safe
	cipher  sscore.Cipher
	stopCh  chan int
	stopped atomic.Bool
}

var _ ITunnel = &base{}

func (t *base) start() {
	panic("unimplement")
}

func (t *base) reload() {
}

func (t *base) stop() {
	t.stopped.Store(true)
	t.stopCh <- 0
}
