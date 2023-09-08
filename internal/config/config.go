package config

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	log "github.com/sirupsen/logrus"
)

type Tunnel struct {
	Name     string `yaml:"name"`
	Protocol string `yaml:"protocol"`
	Listen   string `yaml:"listen"`
	Target   string `yaml:"target"`
}

type Config struct {
	// common
	Mode  string `yaml:"mode"`
	Debug bool   `yaml:"debug"`

	Crypt string `yaml:"crypt"`
	Key   string `yaml:"key"`
	Cork  bool   `yaml:"cork,omitempty"`

	// server
	Listen string `yaml:"listen,omitempty"`

	// client
	Server  string   `yaml:"server,omitempty"`
	Tunnels []Tunnel `yaml:"tunnels,omitempty"`

	// tool
	ToolConf *ToolConfig `yaml:"-"`
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
