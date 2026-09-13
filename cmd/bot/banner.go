package main

import (
	"fmt"

	"aurumflow/config"
	"aurumflow/internal/logger"
)

func printStartupBanner(cfg *config.Config, epic, accountRedacted string, killActive bool) {
	ks := "OFF"
	if killActive {
		ks = "ON"
	}
	journal := "OFF"
	if cfg.Logging.JournalEnabled {
		journal = "ON"
	}
	msg := fmt.Sprintf("\nAURUMFLOW\n\nAPI ENVIRONMENT: %s\nEXECUTION MODE: %s\nACCOUNT: %s\nEPIC: %s\nJOURNAL: %s\nKILL SWITCH: %s\n",
		cfg.API.Environment, cfg.ExecMode(), accountRedacted, epic, journal, ks)
	logger.Info("%s", msg)
}
