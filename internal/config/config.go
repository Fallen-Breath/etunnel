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
	Crypt string `yaml:"crypt"`
	Key   string `yaml:"key"`
	Cork  bool   `yaml:"cork"`
	Debug bool   `yaml:"debug"`

	// server
	Listen string `yaml:"listen"`

	// client
	Tunnels []Tunnel `yaml:"tunnels"`
	Server  string   `yaml:"server"`

	// tool
	Pid int `yaml:"-"`
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

	log.StandardLogger().SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
}
