package config

import (
	"errors"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

func CreateConfigOrDie() *Config {
	conf := &Config{}

	help := flag.Bool("h", false, "Show help and exit")
	configPath := flag.String("conf", "", "Set the file to load config from. Existing arguments from command line will be override")
	flag.StringVar(&conf.Mode, "m", "", "The mode of etunnel. Options: client, server")
	flag.StringVar(&conf.Crypt, "c", Crypts[0], "The encryption method to use")
	flag.StringVar(&conf.Key, "k", "hidden secret", "The secret password for encryption")
	flag.StringVar(&conf.Server, "s", "127.0.0.1:12000", "(client) The address of the etunnel server")
	flag.Var(&conf.Tunnels, "t", "(client) A list of encrypted tunnels")
	flag.StringVar(&conf.Listen, "l", "127.0.0.1:12000", "(server) The address to listen to")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if len(*configPath) > 0 {
		if err := readConfig(conf, *configPath); err != nil {
			log.Fatalf("Read config file failed: %v", err)
		}
		log.Infof("Loaded config from %s", *configPath)
	}

	if err := validateConfig(conf); err != nil {
		log.Fatalf("Validate config failed: %v", err)
	}

	return conf
}

func readConfig(config *Config, configPath string) error {
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", configPath, err)
	}
	if err := yaml.Unmarshal(buf, &config); err != nil {
		return fmt.Errorf("failed to parse yaml from config file %s: %v", configPath, err)
	}
	return nil
}

func validateConfig(conf *Config) error {
	if len(conf.Mode) == 0 {
		return errors.New("missing mode")
	}
	if conf.Mode != ModeClient && conf.Mode != ModeServer {
		return fmt.Errorf("invalid mode '%s'", conf.Mode)
	}
	if err := ValidateCrypt(conf.Crypt); err != nil {
		return err
	}
	if len(conf.Key) == 0 {
		return fmt.Errorf("empty key")
	}
	if err := ValidateAddress(conf.Listen); err != nil {
		return err
	}

	if conf.Mode == ModeClient {
		if err := ValidateAddress(conf.Server); err != nil {
			return fmt.Errorf("invalid server adderss %s: %v", conf.Server, err)
		}
		if len(conf.Tunnels) == 0 {
			return errors.New("no tunnels are defined")
		}
		for _, t := range conf.Tunnels {
			if err := ValidateTunnel(t); err != nil {
				return fmt.Errorf("invalid tunnel definition %s: %v", t, err)
			}
		}
	}
	if conf.Mode == ModeServer {
		if err := ValidateAddress(conf.Listen); err != nil {
			return fmt.Errorf("invalid listen adderss %s: %v", conf.Listen, err)
		}
	}

	return nil
}
