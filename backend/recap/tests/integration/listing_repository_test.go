//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	listingrepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/listing"
)

func TestListingRepository_ListCategoryRecommendations(t *testing.T) {
	testEnv.resetDatabase(t)
	fixture := insertCategoryRecommendationsFixture(t)
	repository := listingrepo.New(testEnv.pool, operationTimeout)
	want := []entity.ListingPreview{
		listingPreview(fixture.preferredNewest),
		listingPreview(fixture.preferredOlder),
		listingPreview(fixture.nonPreferredNewest),
	}

	actual, err := repository.ListCategoryRecommendations(
		testContext(t),
		fixture.userID,
		fixture.categoryID,
		&fixture.preferredSubcategoryID,
		len(want),
	)

	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestListingRepository_ListOldestActiveFavorites(t *testing.T) {
	testEnv.resetDatabase(t)
	fixture := insertOldestActiveFavoritesFixture(t)
	repository := listingrepo.New(testEnv.pool, operationTimeout)
	want := []entity.FavoriteListingPreview{
		{
			ListingPreview: listingPreview(fixture.oldest),
			AddedAt:        fixture.period.From,
		},
		{
			ListingPreview: listingPreview(fixture.second),
			AddedAt:        fixture.period.From.Add(24 * time.Hour),
		},
	}

	actual, err := repository.ListOldestActiveFavorites(
		testContext(t),
		fixture.userID,
		fixture.period,
		len(want),
	)

	require.NoError(t, err)
	require.Len(t, actual, len(want))
	for index := range want {
		require.Equal(
			t,
			want[index].ListingPreview,
			actual[index].ListingPreview,
		)
		require.True(t, want[index].AddedAt.Equal(actual[index].AddedAt))
	}
}
