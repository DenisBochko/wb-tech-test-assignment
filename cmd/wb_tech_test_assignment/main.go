package main

import (
	"context"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"wb-tech-test-assignment/internal/app"
	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/pkg/logger"
)

func main() {
	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	errors := make(chan error)

	application := app.MustNew(ctx, cfg, log)
	defer func() {
		if err := application.Shutdown(); err != nil {
			log.Error("Failed to shutdown application", zap.Error(err))
		}

		if err := log.Sync(); err != nil {
			log.Warn("Failed to sync logger", zap.Error(err))
		}

		log.Info("Application shutdowned")
	}()

	go func() { errors <- application.Run(ctx) }()

	select {
	case err := <-errors:
		if err != nil {
			log.Error("Server error, shutting down...", zap.Error(err))
		}
	case <-ctx.Done():
		log.Info("Received stop signal, shutting down...")
	}
}
