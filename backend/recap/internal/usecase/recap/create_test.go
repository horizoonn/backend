package recap

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestRecapService_Create(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	existingRecapID := uuid.New()
	categoryID := uuid.New()
	profileError := errors.New("profile storage unavailable")
	recapLookupError := errors.New("recap storage unavailable")
	activityError := errors.New("activity storage unavailable")
	oldestFavoriteError := errors.New("favorites unavailable")
	categoryActivityError := errors.New("category activity unavailable")
	recommendationError := errors.New("recommendations unavailable")
	seasonActivityError := errors.New("season activity unavailable")
	categoryCountError := errors.New("category count unavailable")
	createError := errors.New("recap create failed")

	type args struct {
		profileID     uuid.UUID
		year          int
		cancelContext bool
	}

	tests := []struct {
		name            string
		args            args
		want            entity.RecapCreation
		wantErr         error
		wantErrContains string
		setupMocks      func(context.Context, recapTestDependencies)
	}{
		{
			name: "canceled context",
			args: args{
				profileID:     profileID,
				year:          maxRecapYear,
				cancelContext: true,
			},
			wantErr: context.Canceled,
		},
		{
			name: "missing profile id",
			args: args{
				profileID: uuid.Nil,
				year:      maxRecapYear,
			},
			wantErr: entity.ErrProfileIDRequired,
		},
		{
			name: "year below supported range",
			args: args{
				profileID: profileID,
				year:      minRecapYear - 1,
			},
			wantErrContains: "year 2014 is out of range",
		},
		{
			name: "year above supported range",
			args: args{
				profileID: profileID,
				year:      maxRecapYear + 1,
			},
			wantErrContains: "year 2027 is out of range",
		},
		{
			name: "profile lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				dependencies.profile.EXPECT().
					GetByID(ctx, profileID).
					Return(entity.Profile{}, profileError).
					Once()
			},
			wantErr: profileError,
		},
		{
			name: "existing recap lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				dependencies.profile.EXPECT().
					GetByID(ctx, profileID).
					Return(entity.Profile{ID: profileID}, nil).
					Once()
				dependencies.recap.EXPECT().
					GetByProfileAndYear(ctx, profileID, int32(maxRecapYear)).
					Return(entity.Recap{}, recapLookupError).
					Once()
			},
			wantErr: recapLookupError,
		},
		{
			name: "returns existing recap without rebuilding",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				dependencies.profile.EXPECT().
					GetByID(ctx, profileID).
					Return(entity.Profile{ID: profileID}, nil).
					Once()
				dependencies.recap.EXPECT().
					GetByProfileAndYear(ctx, profileID, int32(maxRecapYear)).
					Return(entity.Recap{ID: existingRecapID}, nil).
					Once()
			},
			want: entity.RecapCreation{ID: existingRecapID, Created: false},
		},
		{
			name: "activity totals lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{}, activityError).
					Once()
			},
			wantErr: activityError,
		},
		{
			name: "not enough activity",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{Views: minActionsForRecap - 1}, nil).
					Once()
			},
			wantErr:         entity.ErrNotEnoughActivity,
			wantErrContains: "build recap for 2026",
		},
		{
			name: "oldest active favorite lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{
						Views:           minActionsForRecap,
						FavoritesActive: 1,
					}, nil).
					Once()
				dependencies.listing.EXPECT().
					ListOldestActiveFavorites(
						ctx,
						profileID,
						entity.YearPeriod(maxRecapYear),
						oldestActiveFavoriteLimit,
					).
					Return(nil, oldestFavoriteError).
					Once()
			},
			wantErr: oldestFavoriteError,
		},
		{
			name: "category activity lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{Views: minActionsForRecap}, nil).
					Once()
				dependencies.activity.EXPECT().
					ListActivityByCategories(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(nil, categoryActivityError).
					Once()
			},
			wantErr: categoryActivityError,
		},
		{
			name: "category recommendations lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{Views: minActionsForRecap}, nil).
					Once()
				activities := []entity.CategoryActivity{{
					CategoryID:    categoryID,
					CategoryTitle: "Транспорт",
					Views:         minActionsForRecap,
				}}
				dependencies.activity.EXPECT().
					ListActivityByCategories(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(activities, nil).
					Once()
				dependencies.listing.EXPECT().
					ListCategoryRecommendations(
						ctx,
						profileID,
						categoryID,
						(*uuid.UUID)(nil),
						categoryRecommendationLimit,
					).
					Return(nil, recommendationError).
					Once()
			},
			wantErr: recommendationError,
		},
		{
			name: "season activity lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{Views: minActionsForRecap}, nil).
					Once()
				dependencies.activity.EXPECT().
					ListActivityByCategories(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(nil, nil).
					Once()
				firstSeasonPeriod := entity.Seasons(maxRecapYear)[0].Ranges[0]
				dependencies.activity.EXPECT().
					ListActivityByCategories(ctx, profileID, firstSeasonPeriod).
					Return(nil, seasonActivityError).
					Once()
			},
			wantErr: seasonActivityError,
		},
		{
			name: "category count lookup fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{Views: minActionsForRecap}, nil).
					Once()
				dependencies.activity.EXPECT().
					ListActivityByCategories(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(nil, nil).
					Once()
				expectSeasonActivity(ctx, dependencies.activity, profileID, maxRecapYear, nil)
				dependencies.activity.EXPECT().
					CountCategories(ctx).
					Return(0, categoryCountError).
					Once()
			},
			wantErr: categoryCountError,
		},
		{
			name: "recap creation fails",
			args: args{profileID: profileID, year: maxRecapYear},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				expectMissingRecap(ctx, dependencies, profileID)
				dependencies.activity.EXPECT().
					GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(entity.UserActivity{Views: minActionsForRecap}, nil).
					Once()
				dependencies.activity.EXPECT().
					ListActivityByCategories(ctx, profileID, entity.YearPeriod(maxRecapYear)).
					Return(nil, nil).
					Once()
				expectSeasonActivity(ctx, dependencies.activity, profileID, maxRecapYear, nil)
				dependencies.activity.EXPECT().
					CountCategories(ctx).
					Return(1, nil).
					Once()
				dependencies.recap.EXPECT().
					Create(ctx, mock.Anything).
					Return(entity.RecapCreation{}, createError).
					Once()
			},
			wantErr: createError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			if tt.args.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			service, dependencies := newRecapTestService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(ctx, dependencies)
			}

			got, err := service.Create(ctx, tt.args.profileID, tt.args.year)

			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				require.Error(t, err)
			default:
				require.NoError(t, err)
			}
			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRecapService_Create_PersistsGeneratedRecap(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	profileID := uuid.New()
	categoryID := uuid.New()
	subcategoryID := uuid.New()
	listingID := uuid.New()
	creation := entity.RecapCreation{ID: uuid.New(), Created: true}
	activity := entity.UserActivity{
		ActiveDays:         40,
		Views:              50,
		UniqueListingsSeen: 30,
		Favorites:          8,
		FavoritesActive:    2,
		Purchases:          1,
		MessagesAsBuyer:    3,
		CategoriesTouched:  4,
	}
	categoryActivity := entity.CategoryActivity{
		CategoryID:       categoryID,
		CategoryTitle:    "Электроника",
		SubcategoryID:    &subcategoryID,
		SubcategoryTitle: "Телефоны",
		Views:            50,
		Favorites:        8,
		Purchases:        1,
	}
	oldestFavorite := entity.FavoriteListingPreview{
		ListingPreview: entity.ListingPreview{
			ID:         listingID,
			Title:      "Смартфон",
			Price:      25_000_00,
			CategoryID: categoryID,
		},
	}
	recommendations := []entity.ListingPreview{{
		ID:         uuid.New(),
		Title:      "Другой смартфон",
		Price:      30_000_00,
		CategoryID: categoryID,
	}}
	totalCategories := int64(12)
	service, dependencies := newRecapTestService(t)

	expectMissingRecap(ctx, dependencies, profileID)
	dependencies.activity.EXPECT().
		GetActivityTotals(ctx, profileID, entity.YearPeriod(maxRecapYear)).
		Return(activity, nil).
		Once()
	dependencies.listing.EXPECT().
		ListOldestActiveFavorites(
			ctx,
			profileID,
			entity.YearPeriod(maxRecapYear),
			oldestActiveFavoriteLimit,
		).
		Return([]entity.FavoriteListingPreview{oldestFavorite}, nil).
		Once()
	dependencies.activity.EXPECT().
		ListActivityByCategories(ctx, profileID, entity.YearPeriod(maxRecapYear)).
		Return([]entity.CategoryActivity{categoryActivity}, nil).
		Once()
	dependencies.listing.EXPECT().
		ListCategoryRecommendations(
			ctx,
			profileID,
			categoryID,
			&subcategoryID,
			categoryRecommendationLimit,
		).
		Return(recommendations, nil).
		Once()
	expectSeasonActivity(
		ctx,
		dependencies.activity,
		profileID,
		maxRecapYear,
		[]entity.CategoryActivity{categoryActivity},
	)
	dependencies.activity.EXPECT().
		CountCategories(ctx).
		Return(totalCategories, nil).
		Once()

	var persisted entity.Recap
	dependencies.recap.EXPECT().
		Create(ctx, mock.Anything).
		Run(func(_ context.Context, recap entity.Recap) {
			persisted = recap
		}).
		Return(creation, nil).
		Once()

	got, err := service.Create(ctx, profileID, maxRecapYear)

	require.NoError(t, err)
	require.Equal(t, creation, got)
	require.Equal(t, profileID, persisted.UserID)
	require.Equal(t, int32(maxRecapYear), persisted.Year)
	require.Equal(t, DetectArchetype(activity, totalCategories), persisted.Archetype)
	require.True(t, json.Valid(persisted.Slides))

	var slides []struct {
		Type string `json:"type"`
	}
	require.NoError(t, json.Unmarshal(persisted.Slides, &slides))

	types := make([]string, 0, len(slides))
	for _, slide := range slides {
		types = append(types, slide.Type)
	}
	require.Contains(t, types, slideIntro)
	require.Contains(t, types, slideFavorites)
	require.Contains(t, types, slideFavoriteCategory)
	require.Contains(t, types, slideInterests)
	require.Contains(t, types, slideArchetype)
	require.Contains(t, types, slideFinal)
}
