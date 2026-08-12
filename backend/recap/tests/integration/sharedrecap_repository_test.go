//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	recaprepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/recap"
	sharedrecaprepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/sharedrecap"
)

func TestSharedRecapRepository_Create(t *testing.T) {
	testEnv.resetDatabase(t)

	profile := insertProfile(t, profileFixture{
		ID:           uuid.New(),
		Name:         "Анна",
		Surname:      "Иванова",
		RegisteredAt: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
	})
	recapCreation, err := recaprepo.New(testEnv.pool, operationTimeout).Create(
		testContext(t),
		newRecapFixture(profile.ID),
	)
	require.NoError(t, err)

	repository := sharedrecaprepo.New(testEnv.pool, operationTimeout)
	sharedRecap := newSharedRecapFixture(recapCreation.ID)

	first, err := repository.Create(testContext(t), sharedRecap)
	require.NoError(t, err)
	require.True(t, first.Created)
	require.Equal(t, sharedRecap.Token, first.Token)
	require.False(t, first.CreatedAt.IsZero())

	duplicate := sharedRecap
	duplicate.Token = "zyxwvutsrqponmlkjihgfe"
	second, err := repository.Create(testContext(t), duplicate)
	require.NoError(t, err)
	require.False(t, second.Created)
	require.Equal(t, first.Token, second.Token)
	require.True(t, first.CreatedAt.Equal(second.CreatedAt))

	stored, err := repository.GetByToken(testContext(t), first.Token)
	require.NoError(t, err)
	require.Equal(t, sharedRecap.Token, stored.Token)
	require.Equal(t, sharedRecap.RecapID, stored.RecapID)
	require.Equal(t, sharedRecap.Year, stored.Year)
	require.Equal(t, sharedRecap.DisplayName, stored.DisplayName)
	require.Equal(t, sharedRecap.Archetype, stored.Archetype)
	require.Equal(t, sharedRecap.ActiveDays, stored.ActiveDays)
	require.Equal(t, sharedRecap.Views, stored.Views)
	require.Equal(t, sharedRecap.TopCategory, stored.TopCategory)
	require.Equal(t, sharedRecap.InterestSummary, stored.InterestSummary)
	require.Equal(t, sharedRecap.Badges, stored.Badges)
	require.True(t, first.CreatedAt.Equal(stored.CreatedAt))

	storedByRecapID, err := repository.GetByRecapID(
		testContext(t),
		sharedRecap.RecapID,
	)
	require.NoError(t, err)
	require.Equal(t, stored, storedByRecapID)
}

func TestSharedRecapRepository_Create_RecapNotFound(t *testing.T) {
	testEnv.resetDatabase(t)

	repository := sharedrecaprepo.New(testEnv.pool, operationTimeout)

	_, err := repository.Create(
		testContext(t),
		newSharedRecapFixture(uuid.New()),
	)

	require.ErrorIs(t, err, entity.ErrRecapNotFound)
}

func TestSharedRecapRepository_GetByToken_SharedRecapNotFound(t *testing.T) {
	testEnv.resetDatabase(t)
	repository := sharedrecaprepo.New(testEnv.pool, operationTimeout)

	_, err := repository.GetByToken(
		testContext(t),
		"abcdefghijklmnopqrstuv",
	)

	require.ErrorIs(t, err, entity.ErrSharedRecapNotFound)
}

func TestSharedRecapRepository_GetByRecapID_SharedRecapNotFound(t *testing.T) {
	testEnv.resetDatabase(t)
	repository := sharedrecaprepo.New(testEnv.pool, operationTimeout)

	_, err := repository.GetByRecapID(testContext(t), uuid.New())

	require.ErrorIs(t, err, entity.ErrSharedRecapNotFound)
}
