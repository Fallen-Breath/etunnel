package main

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/constants"
	"github.com/Fallen-Breath/etunnel/internal/tunnel"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conf := config.CreateConfigOrDie()

	log.StandardLogger().SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000"})

	log.Infof("%s v%s starting, mode %s", constants.Name, constants.Version, conf.Mode)
	tun, err := tunnel.NewTunnel(conf)
	if err != nil {
		log.Fatalf("Failed to initialize %s: %v", constants.Name, err)
	}

	reloadCh := make(chan os.Signal, 1)
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reloadCh, syscall.SIGHUP)

	go tun.Start()
	go func() {
		for {
			switch <-reloadCh {
			case syscall.SIGHUP:
				log.Infof("%s reloading", constants.Name)
				tun.Reload()
			case syscall.SIGTERM:
				return
			}
		}
	}()

	sig := <-stopCh
	log.Infof("Terminating by signal %s", sig)
	reloadCh <- syscall.SIGTERM
	tun.Stop()
	log.Infof("%s stopped", constants.Name)
}
