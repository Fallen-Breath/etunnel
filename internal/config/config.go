package config

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	log "github.com/sirupsen/logrus"
)

type Tunnel struct {
	// common
	Id       string `yaml:"-"`
	Protocol string `yaml:"protocol"`

	// client
	Listen string `yaml:"listen,omitempty"`

	// server
	Target string `yaml:"target,omitempty"`
}

type Config struct {
	// common
	Mode  string `yaml:"mode"`
	Debug bool   `yaml:"debug"`

	Crypt   string             `yaml:"crypt"`
	Key     string             `yaml:"key"`
	Cork    bool               `yaml:"cork,omitempty"`
	Listen  string             `yaml:"listen,omitempty"` // server only
	Server  string             `yaml:"server,omitempty"` // client only
	Tunnels map[string]*Tunnel `yaml:"tunnels"`

	// tool
	ToolConf ToolConfig `yaml:"-"`
}

type ToolConfig struct {
	Pid    int
	Reload bool
}

func (t *Tunnel) GetDefinition() string {
	return fmt.Sprintf("%s://%s/%s", t.Protocol, t.Listen, t.Target)
}

func (t *Tunnel) Validate() error {
	if _, ok := proto.GetProtocolMeta(t.Protocol); !ok {
		return fmt.Errorf("unknown protocol %s", t.Protocol)
	}
	if err := ValidateAddress(t.Listen); err != nil {
		return err
	}
	if err := ValidateAddress(t.Target); err != nil {
		return err
	}
	return nil
}

func (c *Config) Apply() {
	if c.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
