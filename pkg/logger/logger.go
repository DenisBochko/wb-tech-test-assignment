package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level      string
	FormatJSON bool
	Rotation   Rotation
}

type Rotation struct {
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

func MustSetupLogger(cfg *Config) *zap.Logger {
	logger, err := SetupLogger(cfg)
	if err != nil {
		panic(err)
	}

	return logger
}

func SetupLogger(cfg *Config) (*zap.Logger, error) {
	stdout := zapcore.AddSync(os.Stdout)

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Rotation.File,
		MaxSize:    cfg.Rotation.MaxSize, // megabytes
		MaxBackups: cfg.Rotation.MaxBackups,
		MaxAge:     cfg.Rotation.MaxAge, // days
	})

	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	developmentCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var consoleEncoder zapcore.Encoder
	if cfg.FormatJSON {
		consoleEncoder = zapcore.NewJSONEncoder(productionCfg)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(developmentCfg)
	}

	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)

	return zap.New(core), nil
}
