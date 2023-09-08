package main

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/constants"
	"github.com/Fallen-Breath/etunnel/internal/tool"
	"github.com/Fallen-Breath/etunnel/internal/tunnel"
	log "github.com/sirupsen/logrus"
)

func main() {
	initLog()
	conf := config.CliEntry()
	conf.Apply()

	switch conf.Mode {
	case config.ModeServer, config.ModeClient:
		log.Infof("%s v%s starting, mode %s", constants.Name, constants.Version, conf.Mode)
		tun, err := tunnel.NewTunnel(conf)
		if err != nil {
			log.Fatalf("Failed to initialize %s: %v", constants.Name, err)
		}
		tun.Run()

	case config.ModeTool:
		tool.RunTools(conf.ToolConf)
	}
}

func initLog() {
	log.StandardLogger().SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})
}
