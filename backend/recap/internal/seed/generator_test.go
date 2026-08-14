package seed

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const generatorTestSeed uint64 = 42

func TestNewGenerator(t *testing.T) {
	t.Parallel()

	catalog := generatorTestCatalog()
	tests := []struct {
		name    string
		year    int
		seed    uint64
		wantErr string
	}{
		{
			name: "minimum supported year",
			year: minSeedYear,
			seed: generatorTestSeed,
		},
		{
			name: "maximum supported year",
			year: maxSeedYear,
			seed: generatorTestSeed,
		},
		{
			name:    "year below supported range",
			year:    minSeedYear - 1,
			seed:    generatorTestSeed,
			wantErr: "seed year must be in range 2015..2026",
		},
		{
			name:    "year above supported range",
			year:    maxSeedYear + 1,
			seed:    generatorTestSeed,
			wantErr: "seed year must be in range 2015..2026",
		},
		{
			name:    "zero seed",
			year:    maxSeedYear,
			seed:    0,
			wantErr: "seed must be non-zero",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			generator, err := NewGenerator(test.year, test.seed, catalog)

			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				require.Nil(t, generator)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.year, generator.year)
			assert.Equal(t, test.seed, generator.seed)
			assert.Equal(t, catalog, generator.catalog)
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		catalog         Catalog
		scenarios       func() []Scenario
		wantErrContains string
		check           func(*testing.T, Dataset)
	}{
		{
			name:      "generates deterministic linked dataset",
			catalog:   generatorTestCatalog(),
			scenarios: generatorTestScenarios,
			check: func(t *testing.T, dataset Dataset) {
				t.Helper()

				assert.Equal(t, maxSeedYear, dataset.Year)
				assert.Equal(t, generatorTestSeed, dataset.Seed)
				assert.Equal(t, Summary{
					Users:         2,
					Categories:    1,
					Subcategories: 1,
					Listings:      6,
					Views:         5,
					Favorites:     3,
					Messages:      3,
					Deals:         1,
				}, dataset.Summary())
				assertGeneratedFunnel(t, dataset)
			},
		},
		{
			name:      "generates all demo scenarios",
			catalog:   DefaultCatalog(),
			scenarios: DefaultScenarios,
			check: func(t *testing.T, dataset Dataset) {
				t.Helper()

				assertGeneratedDemoProfiles(t, dataset)
				assertDatasetReferences(t, dataset)
			},
		},
		{
			name:    "reports unknown category",
			catalog: generatorTestCatalog(),
			scenarios: func() []Scenario {
				scenarios := generatorTestScenarios()
				scenarios[1].PublishedListings.Categories = []CategoryWeight{{
					Category: CategoryCode("unknown"),
					Weight:   1,
				}}

				return scenarios
			},
			wantErrContains: `generate published listing for seller: unknown category "unknown"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			generator, err := NewGenerator(maxSeedYear, generatorTestSeed, test.catalog)
			require.NoError(t, err)

			scenarios := test.scenarios()
			got, err := generator.Generate(scenarios)

			if test.wantErrContains != "" {
				require.ErrorContains(t, err, test.wantErrContains)
				assert.Equal(t, Dataset{}, got)

				return
			}

			require.NoError(t, err)
			repeated, err := generator.Generate(scenarios)
			require.NoError(t, err)
			require.Equal(t, got, repeated)
			test.check(t, got)
		})
	}
}

func assertGeneratedDemoProfiles(t *testing.T, dataset Dataset) {
	t.Helper()

	assert.Equal(t, maxSeedYear, dataset.Year)
	assert.Equal(t, generatorTestSeed, dataset.Seed)
	require.Len(t, dataset.Users, len(DefaultScenarios()))

	profileCodes := make([]string, 0, len(dataset.Users))
	for _, user := range dataset.Users {
		profileCodes = append(profileCodes, user.Code)
	}

	assert.ElementsMatch(t, []string{
		"collector",
		"dealmaker",
		"negotiator",
		"explorer",
		"inactive",
	}, profileCodes)
	assert.NotEmpty(t, dataset.Listings)
	assert.NotEmpty(t, dataset.Views)
	assert.NotEmpty(t, dataset.Favorites)
	assert.NotEmpty(t, dataset.Messages)
	assert.NotEmpty(t, dataset.Deals)
}

func assertDatasetReferences(t *testing.T, dataset Dataset) {
	t.Helper()

	users := make(map[string]struct{}, len(dataset.Users))
	for _, user := range dataset.Users {
		users[user.ID.String()] = struct{}{}
	}

	categories := make(map[string]struct{}, len(dataset.Categories))
	for _, category := range dataset.Categories {
		categories[category.ID.String()] = struct{}{}
	}

	subcategories := make(map[string]struct{}, len(dataset.Subcategories))
	for _, subcategory := range dataset.Subcategories {
		assert.Contains(t, categories, subcategory.CategoryID.String())
		subcategories[subcategory.ID.String()] = struct{}{}
	}

	listings := make(map[string]struct{}, len(dataset.Listings))
	for _, listing := range dataset.Listings {
		assert.Contains(t, users, listing.SellerID.String())
		assert.Contains(t, categories, listing.CategoryID.String())
		require.NotNil(t, listing.SubcategoryID)
		assert.Contains(t, subcategories, listing.SubcategoryID.String())
		listings[listing.ID.String()] = struct{}{}
	}

	for _, view := range dataset.Views {
		assert.Contains(t, users, view.UserID.String())
		assert.Contains(t, listings, view.ListingID.String())
	}
	for _, favorite := range dataset.Favorites {
		assert.Contains(t, users, favorite.UserID.String())
		assert.Contains(t, listings, favorite.ListingID.String())
	}
	for _, message := range dataset.Messages {
		assert.Contains(t, users, message.BuyerID.String())
		assert.Contains(t, users, message.SellerID.String())
		assert.Contains(t, listings, message.ListingID.String())
	}
	for _, deal := range dataset.Deals {
		assert.Contains(t, users, deal.BuyerID.String())
		assert.Contains(t, listings, deal.ListingID.String())
	}
}

func assertGeneratedFunnel(t *testing.T, dataset Dataset) {
	t.Helper()

	users := make(map[string]UserRow, len(dataset.Users))
	for _, user := range dataset.Users {
		users[user.Code] = user
	}

	require.Contains(t, users, "buyer")
	require.Contains(t, users, "seller")
	assert.Equal(t, maxSeedYear-3, users["buyer"].RegisteredAt.Year())
	assert.Equal(t, maxSeedYear-5, users["seller"].RegisteredAt.Year())

	statuses := make(map[ListingStatus]int)
	listings := make(map[string]ListingRow, len(dataset.Listings))
	for _, listing := range dataset.Listings {
		statuses[listing.Status]++
		listings[listing.ID.String()] = listing
	}

	assert.Equal(t, 4, statuses[ListingActive])
	assert.Equal(t, 1, statuses[ListingClosed])
	assert.Equal(t, 1, statuses[ListingSold])

	require.Len(t, dataset.Deals, 1)
	deal := dataset.Deals[0]
	require.NotNil(t, deal.CompletedAt)
	assert.Equal(t, users["buyer"].ID, deal.BuyerID)
	assert.Equal(t, ListingSold, listings[deal.ListingID.String()].Status)

	for _, message := range dataset.Messages {
		assert.Equal(t, users["buyer"].ID, message.BuyerID)
		assert.Equal(t, users["seller"].ID, message.SellerID)
		assert.Contains(t, listings, message.ListingID.String())
	}
}

func generatorTestCatalog() Catalog {
	return Catalog{Categories: []CategorySpec{{
		Code:  CategoryHome,
		Title: "Для дома и дачи",
		Subcategories: []SubcategorySpec{{
			Code:  "furniture",
			Title: "Мебель",
			Products: []ProductTemplate{{
				Titles:      []string{"Стол", "Стул"},
				Description: "Хорошее состояние",
				MinPrice:    10_000,
				MaxPrice:    20_000,
			}},
		}},
	}}}
}

func generatorTestScenarios() []Scenario {
	return []Scenario{
		{
			Profile: ProfileSpec{
				Code:            "buyer",
				Name:            "Анна",
				Surname:         "Воронова",
				YearsOnPlatform: 3,
			},
			Funnels: []FunnelPlan{{
				SellerCode:      "seller",
				Category:        CategoryHome,
				Months:          []time.Month{time.January, time.June},
				Views:           CountRange{Min: 5, Max: 5},
				Favorites:       CountRange{Min: 3, Max: 3},
				ActiveFavorites: CountRange{Min: 1, Max: 1},
				Messages:        CountRange{Min: 3, Max: 3},
				Purchases:       CountRange{Min: 1, Max: 1},
			}},
		},
		{
			Profile: ProfileSpec{
				Code:            "seller",
				Name:            "Михаил",
				Surname:         "Орлов",
				YearsOnPlatform: 5,
			},
			PublishedListings: ListingPlan{
				Count:      CountRange{Min: 2, Max: 2},
				Categories: []CategoryWeight{{Category: CategoryHome, Weight: 1}},
				Months:     []time.Month{time.March},
			},
		},
	}
}
