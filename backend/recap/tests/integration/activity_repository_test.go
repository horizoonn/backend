//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	activityrepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/activity"
)

func TestActivityRepository_GetActivityTotals(t *testing.T) {
	testEnv.resetDatabase(t)
	fixture := insertActivityFixture(t)
	repository := activityrepo.New(testEnv.pool, operationTimeout)

	actual, err := repository.GetActivityTotals(
		testContext(t),
		fixture.userID,
		fixture.period,
	)

	require.NoError(t, err)
	require.Equal(t, entity.UserActivity{
		ActiveDays:         2,
		Views:              3,
		UniqueListingsSeen: 2,
		Favorites:          2,
		FavoritesActive:    1,
		Purchases:          1,
		Sales:              1,
		SalesAmount:        7_000,
		MessagesAsBuyer:    1,
		MessagesAsSeller:   1,
		CategoriesTouched:  2,
		ListingsCreated:    1,
	}, actual)
}

func TestActivityRepository_ListActivityByCategories(t *testing.T) {
	testEnv.resetDatabase(t)
	fixture := insertActivityFixture(t)
	repository := activityrepo.New(testEnv.pool, operationTimeout)

	actual, err := repository.ListActivityByCategories(
		testContext(t),
		fixture.userID,
		fixture.period,
	)

	require.NoError(t, err)
	require.Equal(t, []entity.CategoryActivity{
		{
			CategoryID:       fixture.categoryID,
			CategoryTitle:    "Авто",
			SubcategoryID:    &fixture.subcategoryID,
			SubcategoryTitle: "Автомобили",
			Views:            2,
			Favorites:        1,
			Purchases:        1,
		},
		{
			CategoryID:    fixture.secondCategoryID,
			CategoryTitle: "Дом",
			Views:         1,
			Favorites:     1,
			Sales:         1,
		},
	}, actual)
}
