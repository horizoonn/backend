// Package main is the entry point for the recap service.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/app"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/config"
	applogger "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/logger"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("load config: %v", err)

		return 1
	}

	logger, err := applogger.New(cfg.Logger)
	if err != nil {
		log.Printf("initialize logger: %v", err)

		return 1
	}
	defer func() {
		if err := applogger.Sync(logger); err != nil {
			log.Printf("%v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := app.Run(ctx, logger, cfg); err != nil {
		logger.Error("recap service failed", zap.Error(err))

		return 1
	}

	return 0
}
