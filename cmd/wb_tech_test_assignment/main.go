package main

import (
	"fmt"

	"go.uber.org/zap"

	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/pkg/logger"
)

func main() {
	cfg := config.MustLoadConfig()
	config.MustPrintConfig(cfg)

	loggerCfg := &logger.Config{
		Level:      cfg.Level,
		FormatJSON: cfg.FormatJSON,
		Rotation: logger.Rotation{
			File:       cfg.Rotation.File,
			MaxSize:    cfg.Rotation.MaxSize,
			MaxBackups: cfg.Rotation.MaxBackups,
			MaxAge:     cfg.Rotation.MaxAge,
		},
	}

	log := logger.MustSetupLogger(loggerCfg)
	defer func() {
		if err := log.Sync(); err != nil {
			log.Error("failed to sync logger", zap.Error(err))
		}
	}()

	log.Debug("start")
	log.Info("start")
	log.Warn("start")

	fmt.Println("start")
}
