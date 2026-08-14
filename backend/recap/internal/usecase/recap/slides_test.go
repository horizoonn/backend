package recap

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestFinalStatsCollectTheNumbersOfTheYear(t *testing.T) {
	t.Parallel()

	input := slideInput{
		activity: entity.UserActivity{
			ActiveDays:       128,
			Views:            127,
			Favorites:        21,
			MessagesAsBuyer:  100,
			MessagesAsSeller: 28,
		},
		seasons: []seasonLeader{
			{season: entity.SeasonWinter},
			{season: entity.SeasonSpring},
			{season: entity.SeasonSummer},
			{season: entity.SeasonAutumn},
		},
	}

	stats := finalStats(input)

	expected := []struct {
		code  string
		value int64
	}{
		{statActiveDays, 128},
		{statViews, 127},
		{statFavorites, 21},
		{statMessages, 128},
		{statSeasons, 4},
	}

	require.Len(t, stats, len(expected))

	for i, want := range expected {
		assert.Equalf(t, want.code, stats[i].Code, "unexpected code at position %d", i)
		assert.Equalf(t, want.value, stats[i].Value, "unexpected value for %s", want.code)
		assert.NotEmptyf(t, stats[i].Label, "tile %s has no label", want.code)
	}
}

func TestFinalStatsSkipEmptyCounters(t *testing.T) {
	t.Parallel()

	stats := finalStats(slideInput{
		activity: entity.UserActivity{ActiveDays: 12, Views: 40},
	})

	require.Len(t, stats, 2)

	for _, tile := range stats {
		assert.NotZerof(t, tile.Value, "tile %s has nothing to show", tile.Code)
	}
}

func TestBuildFinalSlideActions(t *testing.T) {
	t.Parallel()

	categoryID := uuid.New()
	category := entity.CategoryScore{
		CategoryID: categoryID,
		Title:      "Для дома и дачи",
	}

	tests := []struct {
		name        string
		archetype   entity.ArchetypeName
		categories  []entity.CategoryScore
		favorites   int64
		wantActions []cta
	}{
		{
			name:      "dealmaker can create a listing",
			archetype: entity.ArchetypeDealmaker,
			wantActions: []cta{
				{Action: ctaShareRecap, Title: "Поделиться итогами"},
				{Action: ctaCreateListing, Title: "Разместить объявление"},
			},
		},
		{
			name:      "collector has no create listing action",
			archetype: entity.ArchetypeCollector,
			wantActions: []cta{
				{Action: ctaShareRecap, Title: "Поделиться итогами"},
			},
		},
		{
			name:       "dealmaker keeps every available action in order",
			archetype:  entity.ArchetypeDealmaker,
			categories: []entity.CategoryScore{category},
			favorites:  1,
			wantActions: []cta{
				{Action: ctaShareRecap, Title: "Поделиться итогами"},
				{Action: ctaCreateListing, Title: "Разместить объявление"},
				{
					Action:     ctaOpenCategory,
					Title:      "Вернуться в Для дома и дачи",
					CategoryID: &categoryID,
				},
				{Action: ctaOpenFavorites, Title: "Открыть избранное"},
			},
		},
		{
			name:       "collector keeps product actions without create listing",
			archetype:  entity.ArchetypeCollector,
			categories: []entity.CategoryScore{category},
			favorites:  1,
			wantActions: []cta{
				{Action: ctaShareRecap, Title: "Поделиться итогами"},
				{
					Action:     ctaOpenCategory,
					Title:      "Вернуться в Для дома и дачи",
					CategoryID: &categoryID,
				},
				{Action: ctaOpenFavorites, Title: "Открыть избранное"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			slide, ok := buildFinalSlide(slideInput{
				activity:   entity.UserActivity{FavoritesActive: tt.favorites},
				categories: tt.categories,
				archetype:  entity.Archetype{UserArchetype: tt.archetype},
			})

			require.True(t, ok)
			final, ok := slide.(finalSlide)
			require.True(t, ok)
			require.Equal(t, tt.wantActions, final.Actions)
		})
	}
}
