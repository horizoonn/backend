package sharedrecap

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	sharedrecapmocks "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/sharedrecap/mocks"
)

type sharedRecapTestDependencies struct {
	recap       *sharedrecapmocks.MockRecapRepository
	profile     *sharedrecapmocks.MockProfileRepository
	sharedRecap *sharedrecapmocks.MockSharedRecapRepository
}

func newSharedRecapTestService(
	t *testing.T,
	tokenGenerator TokenGenerator,
) (*sharedRecapService, sharedRecapTestDependencies) {
	t.Helper()

	dependencies := sharedRecapTestDependencies{
		recap:       sharedrecapmocks.NewMockRecapRepository(t),
		profile:     sharedrecapmocks.NewMockProfileRepository(t),
		sharedRecap: sharedrecapmocks.NewMockSharedRecapRepository(t),
	}

	return NewSharedRecapService(
		dependencies.recap,
		dependencies.profile,
		dependencies.sharedRecap,
		tokenGenerator,
	), dependencies
}

func validSharedRecapToken() entity.SharedRecapToken {
	return entity.SharedRecapToken(strings.Repeat("a", entity.SharedRecapTokenLength))
}

func validRecapForSharing(recapID, profileID uuid.UUID) entity.Recap {
	return entity.Recap{
		ID:     recapID,
		UserID: profileID,
		Year:   maxSharedRecapYear,
		Archetype: entity.Archetype{
			UserArchetype: entity.ArchetypeExplorer,
		},
		Slides: json.RawMessage(`[{"type":"active_days","activeDays":42}]`),
	}
}

func validSharedRecap(recapID uuid.UUID) entity.SharedRecap {
	return entity.SharedRecap{
		Token:       validSharedRecapToken(),
		RecapID:     recapID,
		Year:        maxSharedRecapYear,
		DisplayName: "Анна",
		Archetype: entity.SharedArchetype{
			Name:        entity.ArchetypeExplorer,
			Title:       "Исследователь",
			Description: "Интерес к разным категориям и постоянный поиск новых находок.",
		},
		ActiveDays: 42,
		Badges:     []entity.SharedBadge{},
	}
}

func expectShareBuildInputs(
	ctx context.Context,
	dependencies sharedRecapTestDependencies,
	recap entity.Recap,
	profile entity.Profile,
) {
	dependencies.sharedRecap.EXPECT().
		GetByRecapID(ctx, recap.ID).
		Return(entity.SharedRecap{}, entity.ErrSharedRecapNotFound).
		Once()
	dependencies.recap.EXPECT().
		GetByID(ctx, recap.ID).
		Return(recap, nil).
		Once()
	dependencies.profile.EXPECT().
		GetByID(ctx, recap.UserID).
		Return(profile, nil).
		Once()
}
