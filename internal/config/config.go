package config

import (
	"encoding/json"
	"flag"
)

type Config struct {
	// common
	Mode  string `yaml:"mode"` // client, server
	Crypt string `yaml:"crypt"`
	Key   string `yaml:"key"`

	// client
	Tunnels stringList `yaml:"tunnels"`
	Server  string     `yaml:"server"`

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
