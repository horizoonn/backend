// Package listing reads listing previews used by personal recap recommendations.
package listing

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool      *pgxpool.Pool
	opTimeout time.Duration
}

func New(pool *pgxpool.Pool, opTimeout time.Duration) *Repository {
	return &Repository{
		pool:      pool,
		opTimeout: opTimeout,
	}
}
