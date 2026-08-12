//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

type profileFixture struct {
	ID           uuid.UUID
	Name         string
	Surname      string
	AvatarURL    *string
	Hint         *string
	RegisteredAt time.Time
}

func insertProfile(t *testing.T, fixture profileFixture) entity.Profile {
	t.Helper()

	_, err := testEnv.pool.Exec(
		testContext(t),
		`
			INSERT INTO recap.users (
				id,
				name,
				surname,
				avatar_url,
				hint,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
		fixture.ID,
		fixture.Name,
		fixture.Surname,
		fixture.AvatarURL,
		fixture.Hint,
		fixture.RegisteredAt,
	)
	require.NoError(t, err)

	hint := ""
	if fixture.Hint != nil {
		hint = *fixture.Hint
	}

	return entity.Profile{
		ID:           fixture.ID,
		Name:         fixture.Name,
		Surname:      fixture.Surname,
		AvatarURL:    fixture.AvatarURL,
		Hint:         hint,
		RegisteredAt: fixture.RegisteredAt,
	}
}
