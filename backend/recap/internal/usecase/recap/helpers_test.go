package recap

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	recapmocks "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/recap/mocks"
)

type recapTestDependencies struct {
	activity *recapmocks.MockActivityRepository
	recap    *recapmocks.MockRecapRepository
	profile  *recapmocks.MockProfileRepository
	listing  *recapmocks.MockListingRepository
}

func newRecapTestService(t *testing.T) (*recapService, recapTestDependencies) {
	t.Helper()

	dependencies := recapTestDependencies{
		activity: recapmocks.NewMockActivityRepository(t),
		recap:    recapmocks.NewMockRecapRepository(t),
		profile:  recapmocks.NewMockProfileRepository(t),
		listing:  recapmocks.NewMockListingRepository(t),
	}

	return NewRecapService(
		dependencies.activity,
		dependencies.recap,
		dependencies.profile,
		dependencies.listing,
	), dependencies
}

func expectMissingRecap(
	ctx context.Context,
	dependencies recapTestDependencies,
	profileID uuid.UUID,
) {
	dependencies.profile.EXPECT().
		GetByID(ctx, profileID).
		Return(entity.Profile{ID: profileID}, nil).
		Once()
	dependencies.recap.EXPECT().
		GetByProfileAndYear(ctx, profileID, int32(maxRecapYear)).
		Return(entity.Recap{}, entity.ErrRecapNotFound).
		Once()
}

func expectSeasonActivity(
	ctx context.Context,
	repository *recapmocks.MockActivityRepository,
	profileID uuid.UUID,
	year int,
	activities []entity.CategoryActivity,
) {
	for _, season := range entity.Seasons(year) {
		for _, period := range season.Ranges {
			repository.EXPECT().
				ListActivityByCategories(ctx, profileID, period).
				Return(activities, nil).
				Once()
		}
	}
}
