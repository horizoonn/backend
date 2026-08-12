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

func TestRecapService_OldestActiveFavorite(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	period := entity.YearPeriod(maxRecapYear)
	storageError := errors.New("favorites unavailable")
	favorite := entity.FavoriteListingPreview{
		ListingPreview: entity.ListingPreview{ID: uuid.New(), Title: "Смартфон"},
	}

	type args struct {
		profileID uuid.UUID
		period    entity.Period
	}

	tests := []struct {
		name       string
		args       args
		want       entity.FavoriteListingPreview
		wantFound  bool
		wantErr    error
		setupMocks func(context.Context, *recapmocks.MockListingRepository)
	}{
		{
			name: "repository error",
			args: args{profileID: profileID, period: period},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockListingRepository) {
				repository.EXPECT().
					ListOldestActiveFavorites(ctx, profileID, period, oldestActiveFavoriteLimit).
					Return(nil, storageError).
					Once()
			},
			wantErr: storageError,
		},
		{
			name: "no active favorites",
			args: args{profileID: profileID, period: period},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockListingRepository) {
				repository.EXPECT().
					ListOldestActiveFavorites(ctx, profileID, period, oldestActiveFavoriteLimit).
					Return(nil, nil).
					Once()
			},
		},
		{
			name: "returns oldest favorite",
			args: args{profileID: profileID, period: period},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockListingRepository) {
				repository.EXPECT().
					ListOldestActiveFavorites(ctx, profileID, period, oldestActiveFavoriteLimit).
					Return([]entity.FavoriteListingPreview{favorite}, nil).
					Once()
			},
			want:      favorite,
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			repository := recapmocks.NewMockListingRepository(t)
			tt.setupMocks(ctx, repository)
			service := &recapService{listingRepository: repository}

			got, found, err := service.oldestActiveFavorite(
				ctx,
				tt.args.profileID,
				tt.args.period,
			)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantFound, found)
		})
	}
}

func TestRecapService_CategoryRecommendations(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	categoryID := uuid.New()
	subcategoryID := uuid.New()
	storageError := errors.New("recommendations unavailable")
	recommendations := []entity.ListingPreview{{
		ID:         uuid.New(),
		Title:      "Смартфон",
		CategoryID: categoryID,
	}}

	type args struct {
		profileID uuid.UUID
		category  entity.CategoryScore
	}

	tests := []struct {
		name       string
		args       args
		want       []entity.ListingPreview
		wantErr    error
		setupMocks func(context.Context, *recapmocks.MockListingRepository)
	}{
		{
			name: "without preferred subcategory",
			args: args{
				profileID: profileID,
				category:  entity.CategoryScore{CategoryID: categoryID},
			},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockListingRepository) {
				repository.EXPECT().
					ListCategoryRecommendations(
						ctx,
						profileID,
						categoryID,
						(*uuid.UUID)(nil),
						categoryRecommendationLimit,
					).
					Return(recommendations, nil).
					Once()
			},
			want: recommendations,
		},
		{
			name: "with preferred subcategory",
			args: args{
				profileID: profileID,
				category: entity.CategoryScore{
					CategoryID:  categoryID,
					Subcategory: &entity.SubcategoryScore{ID: subcategoryID},
				},
			},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockListingRepository) {
				repository.EXPECT().
					ListCategoryRecommendations(
						ctx,
						profileID,
						categoryID,
						&subcategoryID,
						categoryRecommendationLimit,
					).
					Return(recommendations, nil).
					Once()
			},
			want: recommendations,
		},
		{
			name: "repository error",
			args: args{
				profileID: profileID,
				category:  entity.CategoryScore{CategoryID: categoryID},
			},
			setupMocks: func(ctx context.Context, repository *recapmocks.MockListingRepository) {
				repository.EXPECT().
					ListCategoryRecommendations(
						ctx,
						profileID,
						categoryID,
						(*uuid.UUID)(nil),
						categoryRecommendationLimit,
					).
					Return(nil, storageError).
					Once()
			},
			wantErr: storageError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			repository := recapmocks.NewMockListingRepository(t)
			tt.setupMocks(ctx, repository)
			service := &recapService{listingRepository: repository}

			got, err := service.categoryRecommendations(
				ctx,
				tt.args.profileID,
				tt.args.category,
			)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRecapService_LoadCategoryRecap(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	period := entity.YearPeriod(maxRecapYear)
	categoryID := uuid.New()
	activityError := errors.New("category activity unavailable")
	activities := []entity.CategoryActivity{{
		CategoryID:    categoryID,
		CategoryTitle: "Электроника",
		Views:         10,
	}}
	recommendations := []entity.ListingPreview{{
		ID:         uuid.New(),
		CategoryID: categoryID,
	}}

	type args struct {
		profileID uuid.UUID
		period    entity.Period
	}

	tests := []struct {
		name       string
		args       args
		want       categoryRecapData
		wantErr    error
		setupMocks func(
			context.Context,
			*recapmocks.MockActivityRepository,
			*recapmocks.MockListingRepository,
		)
	}{
		{
			name: "activity repository error",
			args: args{profileID: profileID, period: period},
			setupMocks: func(
				ctx context.Context,
				activityRepository *recapmocks.MockActivityRepository,
				_ *recapmocks.MockListingRepository,
			) {
				activityRepository.EXPECT().
					ListActivityByCategories(ctx, profileID, period).
					Return(nil, activityError).
					Once()
			},
			wantErr: activityError,
		},
		{
			name: "empty activity skips recommendations",
			args: args{profileID: profileID, period: period},
			setupMocks: func(
				ctx context.Context,
				activityRepository *recapmocks.MockActivityRepository,
				_ *recapmocks.MockListingRepository,
			) {
				activityRepository.EXPECT().
					ListActivityByCategories(ctx, profileID, period).
					Return(nil, nil).
					Once()
			},
			want: categoryRecapData{
				categories:      []entity.CategoryScore{},
				recommendations: []entity.ListingPreview{},
			},
		},
		{
			name: "scores activity and loads top category recommendations",
			args: args{profileID: profileID, period: period},
			setupMocks: func(
				ctx context.Context,
				activityRepository *recapmocks.MockActivityRepository,
				listingRepository *recapmocks.MockListingRepository,
			) {
				activityRepository.EXPECT().
					ListActivityByCategories(ctx, profileID, period).
					Return(activities, nil).
					Once()
				listingRepository.EXPECT().
					ListCategoryRecommendations(
						ctx,
						profileID,
						categoryID,
						(*uuid.UUID)(nil),
						categoryRecommendationLimit,
					).
					Return(recommendations, nil).
					Once()
			},
			want: categoryRecapData{
				categories:      ScoreCategories(activities, DefaultCategoryWeights),
				recommendations: recommendations,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			activityRepository := recapmocks.NewMockActivityRepository(t)
			listingRepository := recapmocks.NewMockListingRepository(t)
			tt.setupMocks(ctx, activityRepository, listingRepository)
			service := &recapService{
				activityRepository: activityRepository,
				listingRepository:  listingRepository,
				categoryWeights:    DefaultCategoryWeights,
			}

			got, err := service.loadCategoryRecap(
				ctx,
				tt.args.profileID,
				tt.args.period,
			)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
