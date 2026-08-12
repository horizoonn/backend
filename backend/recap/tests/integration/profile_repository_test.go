//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	profilerepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/profile"
)

func TestProfileRepository_GetByID(t *testing.T) {
	repository := profilerepo.New(testEnv.pool, operationTimeout)
	registeredAt := time.Date(2025, time.January, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		avatarURL *string
		hint      *string
	}{
		{
			name: "null values",
		},
		{
			name:      "empty values",
			avatarURL: ptr(""),
			hint:      ptr(""),
		},
		{
			name:      "populated values",
			avatarURL: ptr("https://example.test/avatar.jpg"),
			hint:      ptr("Активный покупатель"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEnv.resetDatabase(t)

			expected := insertProfile(t, profileFixture{
				ID:           uuid.New(),
				Name:         "Анна",
				Surname:      "Иванова",
				AvatarURL:    tt.avatarURL,
				Hint:         tt.hint,
				RegisteredAt: registeredAt,
			})

			actual, err := repository.GetByID(testContext(t), expected.ID)

			require.NoError(t, err)
			requireProfileEqual(t, expected, actual)
		})
	}
}

func TestProfileRepository_GetByID_ProfileNotFound(t *testing.T) {
	testEnv.resetDatabase(t)
	repository := profilerepo.New(testEnv.pool, operationTimeout)

	_, err := repository.GetByID(testContext(t), uuid.New())

	require.ErrorIs(t, err, entity.ErrProfileNotFound)
}

func TestProfileRepository_List_Empty(t *testing.T) {
	testEnv.resetDatabase(t)
	repository := profilerepo.New(testEnv.pool, operationTimeout)

	profiles, err := repository.List(testContext(t))

	require.NoError(t, err)
	require.NotNil(t, profiles)
	require.Empty(t, profiles)
}

func TestProfileRepository_List_OrdersByRegisteredAtAndID(t *testing.T) {
	testEnv.resetDatabase(t)
	repository := profilerepo.New(testEnv.pool, operationTimeout)
	registeredAt := time.Date(2025, time.January, 2, 12, 0, 0, 0, time.UTC)

	later := insertProfile(t, profileFixture{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		Name:         "Пётр",
		Surname:      "Сидоров",
		RegisteredAt: registeredAt.Add(time.Hour),
	})
	second := insertProfile(t, profileFixture{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		Name:         "Иван",
		Surname:      "Петров",
		RegisteredAt: registeredAt,
	})
	first := insertProfile(t, profileFixture{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:         "Анна",
		Surname:      "Иванова",
		AvatarURL:    ptr("https://example.test/avatar.jpg"),
		Hint:         ptr("Активный покупатель"),
		RegisteredAt: registeredAt,
	})

	profiles, err := repository.List(testContext(t))

	require.NoError(t, err)
	require.Len(t, profiles, 3)
	for index, expected := range []entity.Profile{first, second, later} {
		requireProfileEqual(t, expected, profiles[index])
	}
}

func requireProfileEqual(t *testing.T, expected, actual entity.Profile) {
	t.Helper()

	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Surname, actual.Surname)
	require.Equal(t, expected.AvatarURL, actual.AvatarURL)
	require.Equal(t, expected.Hint, actual.Hint)
	require.True(t, expected.RegisteredAt.Equal(actual.RegisteredAt))
}
