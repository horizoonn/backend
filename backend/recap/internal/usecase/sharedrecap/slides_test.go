package sharedrecap

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestExtractRecapFacts(t *testing.T) {
	t.Parallel()

	activeDays := int32(42)
	views := int64(125)
	tests := []struct {
		name            string
		slides          json.RawMessage
		want            recapFacts
		wantErrContains string
	}{
		{
			name: "extracts allowlisted facts and ignores unknown data",
			slides: json.RawMessage(`[
				{"type":"active_days","activeDays":42,"private":"ignored"},
				{"type":"views","views":125},
				{"type":"favorite_category","category":{"id":"11111111-1111-1111-1111-111111111111","title":"Транспорт"}},
				{"type":"purchases","badge":{"code":"buyer_bronze"}},
				{"type":"sales","badge":{"code":"buyer_bronze"}},
				{"type":"sales","badge":{"code":"private_badge"}},
				{"type":"private_slide","messages":["secret"]},
				{"type":"interests","periods":[
					{"period":"winter","category":{"id":"11111111-1111-1111-1111-111111111111","title":"Транспорт"}},
					{"period":"spring","category":{"id":"11111111-1111-1111-1111-111111111111","title":"Транспорт"}}
				]}
			]`),
			want: recapFacts{
				activeDays:      &activeDays,
				views:           &views,
				topCategory:     &entity.SharedCategory{CategoryTitle: "Транспорт"},
				interestSummary: "Главный интерес года — Транспорт.",
				badges: []entity.SharedBadge{{
					Code:        "buyer_bronze",
					Title:       "Первая покупка",
					Description: "Первый шаг в поиске и покупке подходящих вещей.",
					Level:       entity.BadgeLevelBronze,
				}},
				badgeCodes: map[string]struct{}{"buyer_bronze": {}},
			},
		},
		{
			name:            "invalid slides json",
			slides:          []byte(`{`),
			wantErrContains: "decode recap slides",
		},
		{
			name:            "invalid slide json",
			slides:          []byte(`[1]`),
			wantErrContains: "decode recap slide 0: decode slide type",
		},
		{
			name:            "active days missing",
			slides:          []byte(`[]`),
			wantErrContains: "active days slide is missing",
		},
		{
			name:            "activeDays field missing",
			slides:          []byte(`[{"type":"active_days"}]`),
			wantErrContains: "activeDays is missing",
		},
		{
			name:            "views field missing",
			slides:          []byte(`[{"type":"active_days","activeDays":1},{"type":"views"}]`),
			wantErrContains: "views is missing",
		},
		{
			name: "badge code missing",
			slides: []byte(
				`[{"type":"active_days","activeDays":1},{"type":"purchases","badge":{}}]`,
			),
			wantErrContains: "badge code is missing",
		},
		{
			name: "malformed interests",
			slides: []byte(
				`[{"type":"active_days","activeDays":1},{"type":"interests","periods":"bad"}]`,
			),
			wantErrContains: "decode interests slide",
		},
		{
			name: "category ids are not exposed",
			slides: []byte(`[
				{"type":"active_days","activeDays":1},
				{
					"type":"favorite_category",
					"category":{
						"id":"11111111-1111-1111-1111-111111111111",
						"title":"Транспорт"
					}
				}
			]`),
			want: recapFacts{
				activeDays: func() *int32 {
					value := int32(1)
					return &value
				}(),
				topCategory: &entity.SharedCategory{
					CategoryTitle: "Транспорт",
				},
				badges:     []entity.SharedBadge{},
				badgeCodes: map[string]struct{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := extractRecapFacts(tt.slides)

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestApplyStoredSlide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		rawSlide        json.RawMessage
		assertFacts     func(*testing.T, recapFacts)
		wantErrContains string
	}{
		{
			name:     "active days",
			rawSlide: []byte(`{"type":"active_days","activeDays":7}`),
			assertFacts: func(t *testing.T, facts recapFacts) {
				require.Equal(t, int32(7), *facts.activeDays)
			},
		},
		{
			name:     "views",
			rawSlide: []byte(`{"type":"views","views":9}`),
			assertFacts: func(t *testing.T, facts recapFacts) {
				require.Equal(t, int64(9), *facts.views)
			},
		},
		{
			name: "category with subcategory",
			rawSlide: []byte(
				`{"type":"favorite_category","category":{"title":"Транспорт"},` +
					`"subcategory":{"title":"Авто"}}`,
			),
			assertFacts: func(t *testing.T, facts recapFacts) {
				require.Equal(t, &entity.SharedCategory{
					CategoryTitle:    "Транспорт",
					SubcategoryTitle: "Авто",
				}, facts.topCategory)
			},
		},
		{
			name:     "badge without achievement",
			rawSlide: []byte(`{"type":"purchases"}`),
			assertFacts: func(t *testing.T, facts recapFacts) {
				require.Empty(t, facts.badges)
			},
		},
		{
			name: "interests",
			rawSlide: []byte(`{
				"type":"interests",
				"periods":[{
					"period":"summer",
					"category":{
						"id":"11111111-1111-1111-1111-111111111111",
						"title":"Дача"
					}
				}]
			}`),
			assertFacts: func(t *testing.T, facts recapFacts) {
				require.Equal(t, "Главный интерес года — Дача.", facts.interestSummary)
			},
		},
		{
			name:     "unknown type is ignored",
			rawSlide: []byte(`{"type":"private"}`),
			assertFacts: func(t *testing.T, facts recapFacts) {
				require.Equal(t, recapFacts{
					badges:     []entity.SharedBadge{},
					badgeCodes: map[string]struct{}{},
				}, facts)
			},
		},
		{
			name:            "malformed envelope",
			rawSlide:        []byte(`{`),
			wantErrContains: "decode slide type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			facts := recapFacts{badges: []entity.SharedBadge{}, badgeCodes: map[string]struct{}{}}

			err := applyStoredSlide(&facts, tt.rawSlide)

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
				tt.assertFacts(t, facts)
			}
		})
	}
}

func TestBuildInterestSummary(t *testing.T) {
	t.Parallel()

	transportID := uuid.New()
	homeID := uuid.New()
	period := func(season entity.Season, id uuid.UUID, title string) storedInterestPeriod {
		return storedInterestPeriod{Period: season, Category: storedCategoryRef{ID: id, Title: title}}
	}
	tests := []struct {
		name    string
		periods []storedInterestPeriod
		want    string
	}{
		{name: "empty"},
		{
			name: "invalid periods are skipped",
			periods: []storedInterestPeriod{
				period("monsoon", transportID, "Транспорт"),
				period(entity.SeasonWinter, homeID, "  "),
			},
		},
		{
			name: "same category",
			periods: []storedInterestPeriod{
				period(entity.SeasonWinter, transportID, "Транспорт"),
				period(entity.SeasonSpring, transportID, "Транспорт"),
			},
			want: "Главный интерес года — Транспорт.",
		},
		{
			name: "seasonal story",
			periods: []storedInterestPeriod{
				period(entity.SeasonWinter, transportID, "Транспорт"),
				period(entity.SeasonSpring, homeID, "Дом"),
			},
			want: "Зимой — Транспорт, весной — Дом.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, buildInterestSummary(tt.periods))
		})
	}
}

func TestSlideTextHelpers(t *testing.T) {
	t.Parallel()

	t.Run("capitalize", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "", capitalize(""))
		require.Equal(t, "Зимой", capitalize("зимой"))
	})
	t.Run("truncate unicode", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "коротко", truncateText("коротко", 10))
		require.Equal(t, "абв…", truncateText("абвгде", 4))
		require.Equal(t, "аб…", truncateText("аб  вг", 4))
	})
	t.Run("same category", func(t *testing.T) {
		t.Parallel()
		id := uuid.New()
		require.True(t, sameCategory([]storedInterestPeriod{
			{Category: storedCategoryRef{ID: id}},
			{Category: storedCategoryRef{ID: id}},
		}))
		require.False(t, sameCategory([]storedInterestPeriod{
			{Category: storedCategoryRef{ID: id}},
			{Category: storedCategoryRef{ID: uuid.New()}},
		}))
	})
}

func TestPublicSeasonLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		season entity.Season
		want   string
		known  bool
	}{
		{season: entity.SeasonWinter, want: "зимой", known: true},
		{season: entity.SeasonSpring, want: "весной", known: true},
		{season: entity.SeasonSummer, want: "летом", known: true},
		{season: entity.SeasonAutumn, want: "осенью", known: true},
		{season: "monsoon"},
	}

	for _, tt := range tests {
		t.Run(string(tt.season), func(t *testing.T) {
			t.Parallel()
			got, known := publicSeasonLabel(tt.season)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.known, known)
		})
	}
}

func TestApplyInterestsSlideTruncatesSummary(t *testing.T) {
	t.Parallel()

	longTitle := strings.Repeat("я", maxInterestRunes+100)
	rawSlide, err := json.Marshal(map[string]any{
		"type": "interests",
		"periods": []map[string]any{{
			"period":   entity.SeasonWinter,
			"category": map[string]any{"id": uuid.New(), "title": longTitle},
		}},
	})
	require.NoError(t, err)
	facts := recapFacts{}

	err = applyInterestsSlide(&facts, rawSlide)

	require.NoError(t, err)
	require.Len(t, []rune(facts.interestSummary), maxInterestRunes)
	require.True(t, strings.HasSuffix(facts.interestSummary, "…"))
}
