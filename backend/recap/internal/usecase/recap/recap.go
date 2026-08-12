// Package recap implements recap business scenarios.
package recap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const (
	minActionsForRecap = 5

	minRecapYear = 2015
	maxRecapYear = 2026
)

type (
	activityRepository interface {
		GetActivityTotals(
			ctx context.Context,
			userID uuid.UUID,
			period entity.Period,
		) (entity.UserActivity, error)

		ListActivityByCategories(
			ctx context.Context,
			userID uuid.UUID,
			period entity.Period,
		) ([]entity.CategoryActivity, error)

		CountCategories(ctx context.Context) (int64, error)
	}

	recapRepository interface {
		Create(ctx context.Context, recap entity.Recap) (entity.RecapCreation, error)

		GetByID(ctx context.Context, id uuid.UUID) (entity.Recap, error)

		GetByProfileAndYear(
			ctx context.Context,
			profileID uuid.UUID,
			year int32,
		) (entity.Recap, error)
	}

	profileRepository interface {
		GetByID(ctx context.Context, id uuid.UUID) (entity.Profile, error)
	}

	listingRepository interface {
		ListOldestActiveFavorites(
			ctx context.Context,
			userID uuid.UUID,
			period entity.Period,
			limit int,
		) ([]entity.FavoriteListingPreview, error)

		ListCategoryRecommendations(
			ctx context.Context,
			userID uuid.UUID,
			categoryID uuid.UUID,
			preferredSubcategoryID *uuid.UUID,
			limit int,
		) ([]entity.ListingPreview, error)
	}
)

type recapService struct {
	activityRepository activityRepository
	recapRepository    recapRepository
	profileRepository  profileRepository
	listingRepository  listingRepository
	categoryWeights    CategoryWeights
}

func NewRecapService(
	activityRepository activityRepository,
	recapRepository recapRepository,
	profileRepository profileRepository,
	listingRepository listingRepository,
) *recapService {
	return &recapService{
		activityRepository: activityRepository,
		recapRepository:    recapRepository,
		profileRepository:  profileRepository,
		listingRepository:  listingRepository,
		categoryWeights:    DefaultCategoryWeights,
	}
}

func (s *recapService) Create(
	ctx context.Context,
	profileID uuid.UUID,
	year int,
) (entity.RecapCreation, error) {
	if err := ctx.Err(); err != nil {
		return entity.RecapCreation{}, fmt.Errorf("create recap: %w", err)
	}
	if profileID == uuid.Nil {
		return entity.RecapCreation{}, fmt.Errorf("create recap: %w", entity.ErrProfileIDRequired)
	}
	if year < minRecapYear || year > maxRecapYear {
		return entity.RecapCreation{}, fmt.Errorf("create recap: year %d is out of range", year)
	}

	if _, err := s.profileRepository.GetByID(ctx, profileID); err != nil {
		return entity.RecapCreation{}, fmt.Errorf("create recap: %w", err)
	}

	existing, err := s.recapRepository.GetByProfileAndYear(ctx, profileID, convertYearToInt32(year))
	if err == nil {
		return entity.RecapCreation{ID: existing.ID, Created: false}, nil
	}

	if !errors.Is(err, entity.ErrRecapNotFound) {
		return entity.RecapCreation{}, fmt.Errorf("create recap: %w", err)
	}

	recap, err := s.buildRecap(ctx, profileID, year)
	if err != nil {
		return entity.RecapCreation{}, err
	}

	creation, err := s.recapRepository.Create(ctx, recap)
	if err != nil {
		return entity.RecapCreation{}, fmt.Errorf("create recap: %w", err)
	}

	return creation, nil
}

func (s *recapService) Get(ctx context.Context, recapID uuid.UUID) (entity.Recap, error) {
	if err := ctx.Err(); err != nil {
		return entity.Recap{}, fmt.Errorf("get recap: %w", err)
	}
	if recapID == uuid.Nil {
		return entity.Recap{}, fmt.Errorf("get recap: %w", entity.ErrRecapIDRequired)
	}

	recap, err := s.recapRepository.GetByID(ctx, recapID)
	if err != nil {
		return entity.Recap{}, fmt.Errorf("get recap: %w", err)
	}

	return recap, nil
}

// buildRecap collects the metrics, applies the rules and assembles the slides.
func (s *recapService) buildRecap(
	ctx context.Context,
	profileID uuid.UUID,
	year int,
) (entity.Recap, error) {
	period := entity.YearPeriod(year)

	activity, err := s.activityRepository.GetActivityTotals(ctx, profileID, period)
	if err != nil {
		return entity.Recap{}, fmt.Errorf("get activity totals: %w", err)
	}

	if activity.TotalActions() < minActionsForRecap {
		return entity.Recap{}, fmt.Errorf(
			"build recap for %d: %w",
			year,
			entity.ErrNotEnoughActivity,
		)
	}

	var oldestFavorite *entity.FavoriteListingPreview
	if activity.FavoritesActive > 0 {
		favorite, found, favoriteErr := s.oldestActiveFavorite(ctx, profileID, period)
		if favoriteErr != nil {
			return entity.Recap{}, favoriteErr
		}
		if found {
			oldestFavorite = &favorite
		}
	}

	categoryData, err := s.loadCategoryRecap(ctx, profileID, period)
	if err != nil {
		return entity.Recap{}, err
	}

	seasons, err := s.seasonLeaders(ctx, profileID, year)
	if err != nil {
		return entity.Recap{}, err
	}

	totalCategories, err := s.activityRepository.CountCategories(ctx)
	if err != nil {
		return entity.Recap{}, fmt.Errorf("count categories: %w", err)
	}

	archetype := DetectArchetype(activity, totalCategories)

	slides, err := buildSlides(slideInput{
		year:                    convertYearToInt32(year),
		activity:                activity,
		categories:              categoryData.categories,
		seasons:                 seasons,
		archetype:               archetype,
		oldestFavorite:          oldestFavorite,
		categoryRecommendations: categoryData.recommendations,
	})
	if err != nil {
		return entity.Recap{}, fmt.Errorf("build slides: %w", err)
	}

	return entity.Recap{
		UserID:    profileID,
		Year:      convertYearToInt32(year),
		Archetype: archetype,
		Slides:    slides,
	}, nil
}

func convertYearToInt32(year int) int32 {
	if year < minRecapYear || year > maxRecapYear {
		return 0
	}

	return int32(year)
}

// seasonLeaders shows how interests moved during the year. Winter is queried
// twice because January-February and December belong to the same season.
func (s *recapService) seasonLeaders(
	ctx context.Context,
	profileID uuid.UUID,
	year int,
) ([]seasonLeader, error) {
	windows := entity.Seasons(year)
	leaders := make([]seasonLeader, 0, len(windows))

	for _, window := range windows {
		activities := make([]entity.CategoryActivity, 0)

		for _, period := range window.Ranges {
			chunk, err := s.activityRepository.ListActivityByCategories(ctx, profileID, period)
			if err != nil {
				return nil, fmt.Errorf("list activity for %s: %w", window.Season, err)
			}

			activities = append(activities, chunk...)
		}

		scores := ScoreCategories(activities, s.categoryWeights)
		if len(scores) == 0 {
			continue
		}

		leaders = append(leaders, seasonLeader{
			season:   window.Season,
			category: scores[0],
			share:    CategoryShare(scores, scores[0]),
		})
	}

	return leaders, nil
}
