package config

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	log "github.com/sirupsen/logrus"
	"time"
)

type TunnelParams struct {
	// for udp / unixgram
	KeepAlive  bool          `yaml:"keep_alive"`
	TimeToLive time.Duration `yaml:"ttl"`
}

type Tunnel struct {
	Id       string       `yaml:"-"`
	Protocol string       `yaml:"protocol"`
	Listen   string       `yaml:"listen,omitempty"`
	Target   string       `yaml:"target,omitempty"`
	Params   TunnelParams `yaml:"params"`
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

func (t *Tunnel) Validate(mode string) error {
	protoMeta, ok := proto.GetProtocolMeta(t.Protocol)
	if !ok {
		return fmt.Errorf("unknown protocol %s", t.Protocol)
	}

	if protoMeta.Addr == proto.AddrRegular {
		switch mode {
		case ModeServer:
			if err := ValidateAddress(t.Target); err != nil {
				return fmt.Errorf("invalid target address %s: %v", t.Target, err)
			}
		case ModeClient:
			if err := ValidateAddress(t.Listen); err != nil {
				return fmt.Errorf("invalid listen address %s: %v", t.Listen, err)
			}
		}
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
