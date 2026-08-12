// Package migrations contains database migration support.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Migrator struct {
	provider *goose.Provider
}

func New(db *sql.DB) (*Migrator, error) {
	files, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("open embedded migrations: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		files,
		goose.WithTableName("public.goose_db_version"),
	)
	if err != nil {
		return nil, fmt.Errorf("create goose provider: %w", err)
	}

	return &Migrator{provider: provider}, nil
}

func (m *Migrator) Up(ctx context.Context) error {
	if _, err := m.provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
