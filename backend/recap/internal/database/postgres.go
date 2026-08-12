// Package database initializes shared database infrastructure.
package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/config"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/migrations"
)

func NewPostgresPool(
	ctx context.Context,
	cfg config.PostgreSQLConfig,
) (*pgxpool.Pool, error) {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)

	connectionURL := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host: net.JoinHostPort(
			cfg.Host,
			strconv.FormatUint(uint64(cfg.Port), 10),
		),
		Path:     cfg.Database,
		RawQuery: query.Encode(),
	}).String()

	poolConfig, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres pool config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) (runErr error) {
	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := db.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close migration database: %w", err))
		}
	}()

	migrator, err := migrations.New(db)
	if err != nil {
		return fmt.Errorf("initialize migrator: %w", err)
	}

	if err := migrator.Up(ctx); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
