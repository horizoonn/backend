//go:build integration

package integration

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func mustExec(t *testing.T, query string, args ...any) pgconn.CommandTag {
	t.Helper()

	commandTag, err := testEnv.pool.Exec(testContext(t), query, args...)
	require.NoError(t, err)

	return commandTag
}

func ptr[T any](value T) *T {
	return &value
}
