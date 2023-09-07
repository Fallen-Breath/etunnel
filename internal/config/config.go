package config

import (
	"encoding/json"
	"flag"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	// common
	Mode  string `yaml:"mode"` // client, server
	Crypt string `yaml:"crypt"`
	Key   string `yaml:"key"`
	Cork  bool   `yaml:"cork"`
	Debug bool   `yaml:"debug"`

	// client
	Tunnels stringList `yaml:"tunnels"`
	Server  string     `yaml:"server"`
	PidFile string     `yaml:"pid_file"`

	// server
	Listen string `yaml:"listen"`
}

type stringList []string

var _ flag.Value = &stringList{}

func (l *stringList) String() string {
	v, _ := json.Marshal(l)
	return string(v)
}

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func (c *Config) Apply() {
	if c.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
