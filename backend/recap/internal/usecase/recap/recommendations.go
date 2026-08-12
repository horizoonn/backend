package recap

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const oldestActiveFavoriteLimit = 1

const categoryRecommendationLimit = 3

type categoryRecapData struct {
	categories      []entity.CategoryScore
	recommendations []entity.ListingPreview
}

func (s *recapService) oldestActiveFavorite(
	ctx context.Context,
	profileID uuid.UUID,
	period entity.Period,
) (entity.FavoriteListingPreview, bool, error) {
	favorites, err := s.listingRepository.ListOldestActiveFavorites(
		ctx,
		profileID,
		period,
		oldestActiveFavoriteLimit,
	)
	if err != nil {
		return entity.FavoriteListingPreview{}, false, fmt.Errorf(
			"list oldest active favorites: %w",
			err,
		)
	}
	if len(favorites) == 0 {
		return entity.FavoriteListingPreview{}, false, nil
	}

	return favorites[0], true, nil
}

func (s *recapService) categoryRecommendations(
	ctx context.Context,
	profileID uuid.UUID,
	category entity.CategoryScore,
) ([]entity.ListingPreview, error) {
	var preferredSubcategoryID *uuid.UUID
	if category.Subcategory != nil {
		preferredSubcategoryID = &category.Subcategory.ID
	}

	recommendations, err := s.listingRepository.ListCategoryRecommendations(
		ctx,
		profileID,
		category.CategoryID,
		preferredSubcategoryID,
		categoryRecommendationLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("list category recommendations: %w", err)
	}

	return recommendations, nil
}

func (s *recapService) loadCategoryRecap(
	ctx context.Context,
	profileID uuid.UUID,
	period entity.Period,
) (categoryRecapData, error) {
	activity, err := s.activityRepository.ListActivityByCategories(ctx, profileID, period)
	if err != nil {
		return categoryRecapData{}, fmt.Errorf("list activity by categories: %w", err)
	}

	result := categoryRecapData{
		categories:      ScoreCategories(activity, s.categoryWeights),
		recommendations: make([]entity.ListingPreview, 0),
	}
	if len(result.categories) == 0 {
		return result, nil
	}

	result.recommendations, err = s.categoryRecommendations(ctx, profileID, result.categories[0])
	if err != nil {
		return categoryRecapData{}, err
	}

	return result, nil
}
