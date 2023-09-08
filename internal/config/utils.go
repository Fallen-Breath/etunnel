package config

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func LoadConfigFromFlags(flags *cliFlags) (*Config, error) {
	conf := &Config{ToolConf: &ToolConfig{}}
	conf.Mode = flags.Mode
	conf.Debug = flags.Debug

	loadTunnelCommon := func() {
		conf.Crypt = flags.Crypt
		conf.Key = flags.Key
		conf.Cork = flags.Cork
	}

	switch conf.Mode {
	case ModeServer:
		loadTunnelCommon()
		conf.Listen = flags.Listen
	case ModeClient:
		loadTunnelCommon()
		for i, tunStr := range flags.Tunnels {
			protocol, listen, target, err := ParseTunnel(tunStr)
			if err != nil {
				return nil, err
			}
			conf.Tunnels = append(conf.Tunnels, Tunnel{
				Name:     fmt.Sprintf("Tunnel-%d", i),
				Protocol: protocol.Name,
				Listen:   listen,
				Target:   target,
			})
		}
		conf.Server = flags.Server
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
		if len(conf.Tunnels) == 0 {
			return errors.New("no tunnels are defined")
		}
		for _, t := range conf.Tunnels {
			if err := t.Validate(); err != nil {
				return fmt.Errorf("invalid tunnel definition %s: %v", t, err)
			}
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
