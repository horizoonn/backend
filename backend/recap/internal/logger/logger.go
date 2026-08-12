// Package logger configures application logging.
package logger

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"go.uber.org/zap"
)

const serviceName = "recap"

func New(cfg Config) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("parse logger level: %w", err)
	}

	var zapConfig zap.Config
	if cfg.Development {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.Level = level
	zapConfig.InitialFields = map[string]any{
		"service": serviceName,
	}

	log, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return log, nil
}

func Sync(log *zap.Logger) error {
	err := log.Sync()

	switch {
	case err == nil:
		return nil
	case errors.Is(err, syscall.EINVAL):
		return nil
	case errors.Is(err, os.ErrInvalid):
		return nil
	default:
		return fmt.Errorf("sync logger: %w", err)
	}
}
