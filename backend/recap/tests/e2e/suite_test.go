//go:build e2e

package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/config"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/database"
)

const (
	testTimeout      = time.Minute
	teardownTimeout  = 30 * time.Second
	operationTimeout = 5 * time.Second
	postgresImage    = "postgres:18.1-bookworm"
	postgresDatabase = "recap"
	postgresUser     = "recap"
	postgresPassword = "recap"
)

var testEnv *environment

type environment struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
}

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	env, err := newEnvironment(ctx)
	cancel()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "set up e2e environment: %v\n", err)

		return 1
	}
	testEnv = env

	code := m.Run()

	teardownCtx, teardownCancel := context.WithTimeout(context.Background(), teardownTimeout)
	err = env.close(teardownCtx)
	teardownCancel()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "tear down e2e environment: %v\n", err)

		return 1
	}

	return code
}

func newEnvironment(ctx context.Context) (*environment, error) {
	container, err := postgres.Run(
		ctx,
		postgresImage,
		postgres.WithDatabase(postgresDatabase),
		postgres.WithUsername(postgresUser),
		postgres.WithPassword(postgresPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	success := false
	defer func() {
		if success {
			return
		}

		teardownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), teardownTimeout)
		defer cancel()
		_ = container.Terminate(teardownCtx)
	}()

	databaseConfig, err := postgresConfig(ctx, container)
	if err != nil {
		return nil, err
	}

	pool, err := database.NewPostgresPool(ctx, databaseConfig)
	if err != nil {
		return nil, err
	}

	if err := database.RunMigrations(ctx, pool); err != nil {
		pool.Close()

		return nil, err
	}

	success = true

	return &environment{container: container, pool: pool}, nil
}

func postgresConfig(
	ctx context.Context,
	container *postgres.PostgresContainer,
) (config.PostgreSQLConfig, error) {
	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return config.PostgreSQLConfig{}, fmt.Errorf("get postgres connection string: %w", err)
	}

	parsed, err := url.Parse(connectionString)
	if err != nil {
		return config.PostgreSQLConfig{}, fmt.Errorf("parse postgres connection string: %w", err)
	}

	port, err := strconv.ParseUint(parsed.Port(), 10, 16)
	if err != nil {
		return config.PostgreSQLConfig{}, fmt.Errorf("parse postgres port: %w", err)
	}

	password, _ := parsed.User.Password()

	return config.PostgreSQLConfig{
		Host:           parsed.Hostname(),
		Port:           uint16(port),
		User:           parsed.User.Username(),
		Password:       password,
		Database:       strings.TrimPrefix(parsed.Path, "/"),
		SSLMode:        parsed.Query().Get("sslmode"),
		MaxConns:       4,
		MinConns:       1,
		ConnectTimeout: operationTimeout,
	}, nil
}

func (e *environment) close(ctx context.Context) error {
	var closeErr error
	if e.pool != nil {
		e.pool.Close()
	}
	if e.container != nil {
		closeErr = errors.Join(closeErr, e.container.Terminate(ctx))
	}

	return closeErr
}

func testContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	t.Cleanup(cancel)

	return ctx
}
