//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	recaprepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/recap"
)

func TestRecapRepository_Create(t *testing.T) {
	testEnv.resetDatabase(t)

	profile := insertProfile(t, profileFixture{
		ID:           uuid.New(),
		Name:         "Анна",
		Surname:      "Иванова",
		RegisteredAt: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
	})
	repository := recaprepo.New(testEnv.pool, operationTimeout)
	recap := newRecapFixture(profile.ID)

	first, err := repository.Create(testContext(t), recap)
	require.NoError(t, err)
	require.True(t, first.Created)
	require.NotEqual(t, uuid.Nil, first.ID)

	second, err := repository.Create(testContext(t), recap)
	require.NoError(t, err)
	require.False(t, second.Created)
	require.Equal(t, first.ID, second.ID)

	stored, err := repository.GetByProfileAndYear(
		testContext(t),
		profile.ID,
		recap.Year,
	)
	require.NoError(t, err)
	require.Equal(t, first.ID, stored.ID)
	require.Equal(t, recap.UserID, stored.UserID)
	require.Equal(t, recap.Year, stored.Year)
	require.Equal(t, recap.Archetype, stored.Archetype)
	require.JSONEq(t, string(recap.Slides), string(stored.Slides))
	require.False(t, stored.GeneratedAt.IsZero())

	storedByID, err := repository.GetByID(testContext(t), first.ID)
	require.NoError(t, err)
	require.Equal(t, stored, storedByID)
}

func TestRecapRepository_Create_ProfileNotFound(t *testing.T) {
	testEnv.resetDatabase(t)

	repository := recaprepo.New(testEnv.pool, operationTimeout)

	_, err := repository.Create(testContext(t), newRecapFixture(uuid.New()))

	require.ErrorIs(t, err, entity.ErrProfileNotFound)
}

func TestRecapRepository_GetByID_RecapNotFound(t *testing.T) {
	testEnv.resetDatabase(t)
	repository := recaprepo.New(testEnv.pool, operationTimeout)

	_, err := repository.GetByID(testContext(t), uuid.New())

	require.ErrorIs(t, err, entity.ErrRecapNotFound)
}

func TestRecapRepository_GetByProfileAndYear_RecapNotFound(t *testing.T) {
	testEnv.resetDatabase(t)
	repository := recaprepo.New(testEnv.pool, operationTimeout)

	_, err := repository.GetByProfileAndYear(
		testContext(t),
		uuid.New(),
		2025,
	)

	require.ErrorIs(t, err, entity.ErrRecapNotFound)
}
