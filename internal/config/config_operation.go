package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func LoadConfigFromFlags(flags *cliFlags) (*Config, error) {
	conf := &Config{
		Tunnels: make(map[string]*Tunnel),
	}
	conf.Mode = flags.Mode
	conf.Debug = flags.Debug

	loadTunnelCommon := func() error {
		conf.Crypt = flags.Crypt
		conf.Key = flags.Key
		conf.Cork = flags.Cork

		for i, tunStr := range flags.Tunnels {
			id, protocol, address, err := ParseTunnel(tunStr, fmt.Sprintf("tunnel-%d", i))
			if err != nil {
				return err
			}

			tun := Tunnel{
				Id:       id,
				Protocol: protocol.Name,
			}
			switch conf.Mode {
			case ModeServer:
				tun.Target = address
			case ModeClient:
				tun.Listen = address
			}

			conf.Tunnels[id] = &tun
		}
		return nil
	}

	switch conf.Mode {
	case ModeServer:
		conf.Listen = flags.Listen
		if err := loadTunnelCommon(); err != nil {
			return nil, err
		}

	case ModeClient:
		conf.Server = flags.Server
		if err := loadTunnelCommon(); err != nil {
			return nil, err
		}

	case ModeTool:
		conf.ToolConf.Pid = flags.ToolPid
		conf.ToolConf.Reload = flags.ToolReload
	}
	return conf, nil
}

func LoadConfigFromFile(configPath string) (*Config, error) {
	conf := &Config{}
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", configPath, err)
	}
	if err := yaml.Unmarshal(buf, conf); err != nil {
		return nil, fmt.Errorf("failed to parse yaml from config file %s: %v", configPath, err)
	}

	for id, tun := range conf.Tunnels {
		tun.Id = id
	}

	return conf, nil
}

func ValidateConfig(conf *Config) error {
	if len(conf.Mode) == 0 {
		return errors.New("missing mode")
	}

	validateTunnelCommon := func() error {
		if err := ValidateCrypt(conf.Crypt); err != nil {
			return err
		}
		if len(conf.Key) == 0 {
			return fmt.Errorf("empty key")
		}
		if len(conf.Tunnels) == 0 {
			return errors.New("no tunnels are defined")
		}
		for _, tun := range conf.Tunnels {
			if err := tun.Validate(conf.Mode); err != nil {
				return fmt.Errorf("invalid tunnel definition %+v: %v", tun, err)
			}
		}
		return nil
	}

	switch conf.Mode {
	case ModeServer:
		if err := validateTunnelCommon(); err != nil {
			return err
		}
		if err := ValidateAddress(conf.Listen); err != nil {
			return fmt.Errorf("invalid listen adderss %s: %v", conf.Listen, err)
		}
	case ModeClient:
		if err := validateTunnelCommon(); err != nil {
			return err
		}
		if err := ValidateAddress(conf.Server); err != nil {
			return fmt.Errorf("invalid server adderss %s: %v", conf.Server, err)
		}
	case ModeTool:
		if conf.ToolConf.Pid == 0 {
			return errors.New("pid unset")
		}
	}

	return nil
}

func WriteConfigToFile(conf *Config, configPath string) error {
	var buf bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buf)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(&conf); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %v", configPath, err)
	}
	return nil
}

func ParseCommunicateAddress(pa string) (protocol, address string, err error) {
	t := strings.Split(pa, "://")
	if len(t) == 1 {
		protocol = proto.Tcp // use tcp by default
	} else if len(t) == 2 {
		protocol = t[0]
		if ok := proto.CheckProtocol(protocol); !ok {
			return "", "", fmt.Errorf("invalid protocol %s", protocol)
		}
	} else {
		return "", "", errors.New("invalid protocol separation, check your '://' divider")
	}

	address = t[len(t)-1]
	if err := ValidateAddress(address); err != nil {
		return "", "", err
	}
	return protocol, address, nil
}
