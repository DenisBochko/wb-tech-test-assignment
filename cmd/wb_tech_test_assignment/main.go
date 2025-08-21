package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/pkg/logger"
	"wb-tech-test-assignment/pkg/postgres"
)

func main() {
	ctx := context.Background()

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

	postgresCfg := &postgres.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
		MaxConns: cfg.Database.MaxConns,
		MinConns: cfg.Database.MinConns,
		Migration: postgres.Migration{
			Path:      cfg.Migration.Path,
			AutoApply: cfg.Migration.AutoApply,
		},
	}

	database, err := postgres.New(postgresCfg)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	defer database.Close()

	log.Debug("start", zap.Any("database", database.Pool().Ping(ctx)))
	log.Info("start")
	log.Warn("start")

	fmt.Println("start")
}
