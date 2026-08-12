package recap

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	recapmocks "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/recap/mocks"
)

func TestNewRecapService(t *testing.T) {
	t.Parallel()

	activityRepository := recapmocks.NewMockActivityRepository(t)
	recapRepository := recapmocks.NewMockRecapRepository(t)
	profileRepository := recapmocks.NewMockProfileRepository(t)
	listingRepository := recapmocks.NewMockListingRepository(t)

	service := NewRecapService(
		activityRepository,
		recapRepository,
		profileRepository,
		listingRepository,
	)

	require.Same(t, activityRepository, service.activityRepository)
	require.Same(t, recapRepository, service.recapRepository)
	require.Same(t, profileRepository, service.profileRepository)
	require.Same(t, listingRepository, service.listingRepository)
	require.Equal(t, DefaultCategoryWeights, service.categoryWeights)
}

func TestConvertYearToInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		year int
		want int32
	}{
		{name: "minimum supported year", year: minRecapYear, want: int32(minRecapYear)},
		{name: "maximum supported year", year: maxRecapYear, want: int32(maxRecapYear)},
		{name: "below supported range", year: minRecapYear - 1, want: 0},
		{name: "above supported range", year: maxRecapYear + 1, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, convertYearToInt32(test.year))
		})
	}
}

func TestRecapService_SeasonLeaders(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	categoryID := uuid.New()
	storageError := errors.New("season activity unavailable")
	windows := entity.Seasons(maxRecapYear)

	type args struct {
		profileID uuid.UUID
		year      int
	}

	tests := []struct {
		name            string
		args            args
		want            []seasonLeader
		wantErr         error
		wantErrContains string
		setupMocks      func(context.Context, *recapmocks.MockActivityRepository)
	}{
		{
			name: "aggregates split winter",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockActivityRepository) {
				repository.EXPECT().
					ListActivityByCategories(ctx, profileID, windows[0].Ranges[0]).
					Return([]entity.CategoryActivity{{
						CategoryID:    categoryID,
						CategoryTitle: "Электроника",
						Views:         1,
					}}, nil).
					Once()
				repository.EXPECT().
					ListActivityByCategories(ctx, profileID, windows[0].Ranges[1]).
					Return([]entity.CategoryActivity{{
						CategoryID:    categoryID,
						CategoryTitle: "Электроника",
						Views:         2,
					}}, nil).
					Once()
				for _, window := range windows[1:] {
					for _, period := range window.Ranges {
						repository.EXPECT().
							ListActivityByCategories(ctx, profileID, period).
							Return(nil, nil).
							Once()
					}
				}
			},
			want: []seasonLeader{{
				season: entity.SeasonWinter,
				category: entity.CategoryScore{
					CategoryID: categoryID,
					Title:      "Электроника",
					Score:      3,
				},
				share: 100,
			}},
		},
		{
			name: "repository error",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockActivityRepository) {
				repository.EXPECT().
					ListActivityByCategories(ctx, profileID, windows[0].Ranges[0]).
					Return(nil, storageError).
					Once()
			},
			wantErr:         storageError,
			wantErrContains: "list activity for winter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			repository := recapmocks.NewMockActivityRepository(t)
			tt.setupMocks(ctx, repository)
			service := &recapService{
				activityRepository: repository,
				categoryWeights:    DefaultCategoryWeights,
			}

			got, err := service.seasonLeaders(ctx, tt.args.profileID, tt.args.year)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
