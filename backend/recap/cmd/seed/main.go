// Package main generates and stores deterministic demo activity.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/config"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/database"
	applogger "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/logger"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/seed"
)

const defaultSeed uint64 = 20250807

type options struct {
	year   int
	seed   uint64
	reset  bool
	dryRun bool
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	opts, err := parseOptions(args)
	if err != nil {
		log.Printf("parse options: %v", err)

		return 2
	}

	dataset, err := generateDataset(opts)
	if err != nil {
		log.Printf("generate demo dataset: %v", err)

		return 1
	}

	if opts.dryRun {
		if err := printSummary(dataset); err != nil {
			log.Printf("print dataset summary: %v", err)

			return 1
		}

		return 0
	}

	return writeDataset(opts, dataset)
}

func parseOptions(args []string) (options, error) {
	result := options{}
	flags := flag.NewFlagSet("seed", flag.ContinueOnError)
	flags.IntVar(&result.year, "year", time.Now().UTC().Year()-1, "recap activity year")
	flags.Uint64Var(&result.seed, "seed", defaultSeed, "non-zero deterministic random seed")
	flags.BoolVar(&result.reset, "reset", false, "replace existing demo profiles and their related data")
	flags.BoolVar(&result.dryRun, "dry-run", false, "preview generated data without connecting to PostgreSQL")

	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}

	return result, nil
}

func generateDataset(options options) (seed.Dataset, error) {
	generator, err := seed.NewGenerator(options.year, options.seed, seed.DefaultCatalog())
	if err != nil {
		return seed.Dataset{}, err
	}

	dataset, err := generator.Generate(seed.DefaultScenarios())
	if err != nil {
		return seed.Dataset{}, err
	}

	return dataset, nil
}

func writeDataset(options options, dataset seed.Dataset) int {
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
		if syncErr := applogger.Sync(logger); syncErr != nil {
			log.Printf("%v", syncErr)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.NewPostgresPool(ctx, cfg.PostgreSQL)
	if err != nil {
		logger.Error("initialize postgres pool", zap.Error(err))

		return 1
	}
	defer pool.Close()

	if migrationErr := database.RunMigrations(ctx, pool); migrationErr != nil {
		logger.Error("run database migrations", zap.Error(migrationErr))

		return 1
	}

	err = seed.NewWriter(pool).Write(ctx, dataset, options.reset)
	switch {
	case err == nil:
		logSummary(logger, "demo dataset written", dataset)

		return 0
	case errors.Is(err, seed.ErrAlreadySeeded):
		logger.Info(
			"demo dataset already exists; use --reset to replace it",
			zap.Int("year", dataset.Year),
			zap.Uint64("seed", dataset.Seed),
		)

		return 0
	default:
		logger.Error("write demo dataset", zap.Error(err))

		return 1
	}
}

func printSummary(dataset seed.Dataset) error {
	summary := dataset.Summary()
	if _, err := fmt.Fprintf(
		os.Stdout,
		"demo dataset generated: year=%d seed=%d users=%d categories=%d subcategories=%d listings=%d views=%d favorites=%d messages=%d deals=%d\n",
		dataset.Year,
		dataset.Seed,
		summary.Users,
		summary.Categories,
		summary.Subcategories,
		summary.Listings,
		summary.Views,
		summary.Favorites,
		summary.Messages,
		summary.Deals,
	); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}

	return nil
}

func logSummary(logger *zap.Logger, message string, dataset seed.Dataset) {
	summary := dataset.Summary()
	logger.Info(
		message,
		zap.Int("year", dataset.Year),
		zap.Uint64("seed", dataset.Seed),
		zap.Int("users", summary.Users),
		zap.Int("categories", summary.Categories),
		zap.Int("subcategories", summary.Subcategories),
		zap.Int("listings", summary.Listings),
		zap.Int("views", summary.Views),
		zap.Int("favorites", summary.Favorites),
		zap.Int("messages", summary.Messages),
		zap.Int("deals", summary.Deals),
	)
}
