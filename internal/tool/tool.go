package tool

import (
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/constants"
	log "github.com/sirupsen/logrus"
	"os"
)

func RunTools(conf *config.ToolConfig) {
	pid := conf.Pid

	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalf("Process with pid %d not found: %v", pid, err)
		return
	}

	if conf.Reload {
		if err := proc.Signal(constants.SignalReload); err != nil {
			log.Fatalf("Failed to send signal to process with pid %d: %v", pid, err)
			return
		}
	} else {
		log.Warningf("No action was performed")
		return
	}
}
