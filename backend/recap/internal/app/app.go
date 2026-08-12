// Package app manages the recap service lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/config"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/database"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

func Run(
	ctx context.Context,
	logger *zap.Logger,
	cfg config.Config,
) error {
	pool, err := database.NewPostgresPool(ctx, cfg.PostgreSQL)
	if err != nil {
		return fmt.Errorf("initialize postgres pool: %w", err)
	}
	defer pool.Close()

	if migrationErr := database.RunMigrations(ctx, pool); migrationErr != nil {
		return fmt.Errorf("run database migrations: %w", migrationErr)
	}
	logger.Info("database migrations applied")

	handler, err := buildHTTPHandler(pool, cfg, logger)
	if err != nil {
		return fmt.Errorf("build http handler: %w", err)
	}

	listenerConfig := net.ListenConfig{}
	listener, err := listenerConfig.Listen(ctx, "tcp", cfg.HTTP.Address)
	if err != nil {
		return fmt.Errorf("listen http: %w", err)
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	logger.Info("recap service started", zap.String("address", cfg.HTTP.Address))

	if err := runHTTPServer(ctx, server, listener, cfg.App.ShutdownTimeout); err != nil {
		return err
	}

	logger.Info("recap service stopped")

	return nil
}

func runHTTPServer(
	ctx context.Context,
	server *http.Server,
	listener net.Listener,
	shutdownTimeout time.Duration,
) error {
	serveErr := make(chan error, 1)

	go func() {
		serveErr <- server.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		return normalizeHTTPServeError(err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		shutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		shutdownErr := fmt.Errorf("shutdown http server: %w", err)
		closeErr := closeHTTPServer(server)
		serveResultErr := normalizeHTTPServeError(<-serveErr)

		return errors.Join(shutdownErr, closeErr, serveResultErr)
	}

	return normalizeHTTPServeError(<-serveErr)
}

func closeHTTPServer(server *http.Server) error {
	if err := server.Close(); err != nil {
		return fmt.Errorf("close http server: %w", err)
	}

	return nil
}

func normalizeHTTPServeError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) ||
		errors.Is(err, net.ErrClosed) {
		return nil
	}

	return fmt.Errorf("serve http: %w", err)
}
