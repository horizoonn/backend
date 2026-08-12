package sharedrecap

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestBuildSharedRecap(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	profileID := uuid.New()
	views := int64(125)
	recap := validRecapForSharing(recapID, profileID)
	recap.Slides = json.RawMessage(`[
		{"type":"active_days","activeDays":42},
		{"type":"views","views":125},
		{
			"type":"favorite_category",
			"category":{
				"id":"11111111-1111-1111-1111-111111111111",
				"title":"Транспорт"
			},
			"subcategory":{
				"id":"22222222-2222-2222-2222-222222222222",
				"title":"Автомобили"
			}
		},
		{"type":"purchases","badge":{"code":"buyer_gold"}},
		{"type":"sales","badge":{"code":"private_badge"}}
	]`)
	profile := entity.Profile{ID: profileID, Name: "Анна", Surname: "Смирнова", Hint: "private"}

	tests := []struct {
		name            string
		recap           entity.Recap
		profile         entity.Profile
		want            entity.SharedRecap
		wantErrContains string
	}{
		{
			name:    "builds allowlisted public snapshot",
			recap:   recap,
			profile: profile,
			want: entity.SharedRecap{
				RecapID:     recapID,
				Year:        maxSharedRecapYear,
				DisplayName: "Анна",
				Archetype: entity.SharedArchetype{
					Name:        entity.ArchetypeExplorer,
					Title:       "Исследователь",
					Description: "Интерес к разным категориям и постоянный поиск новых находок.",
				},
				ActiveDays: 42,
				Views:      &views,
				TopCategory: &entity.SharedCategory{
					CategoryTitle:    "Транспорт",
					SubcategoryTitle: "Автомобили",
				},
				Badges: []entity.SharedBadge{{
					Code:        "buyer_gold",
					Title:       "Знаток покупок",
					Description: "Умение находить подходящие вещи на Авито.",
					Level:       entity.BadgeLevelGold,
				}},
			},
		},
		{
			name:            "slides cannot be decoded",
			recap:           func() entity.Recap { value := recap; value.Slides = []byte(`{`); return value }(),
			profile:         profile,
			wantErrContains: "extract recap facts: decode recap slides",
		},
		{
			name: "archetype is not public",
			recap: func() entity.Recap {
				value := recap
				value.Archetype.UserArchetype = "private"
				return value
			}(),
			profile:         profile,
			wantErrContains: `build public archetype: unknown archetype "private"`,
		},
		{
			name:            "snapshot validation fails",
			recap:           recap,
			profile:         entity.Profile{ID: profileID, Name: "   "},
			wantErrContains: "display name is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := buildSharedRecap(tt.recap, tt.profile)

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidateSnapshot(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	negativeViews := int64(-1)
	valid := validSharedRecap(recapID)
	tests := []struct {
		name            string
		mutate          func(entity.SharedRecap) entity.SharedRecap
		wantErrContains string
	}{
		{
			name:   "valid",
			mutate: func(value entity.SharedRecap) entity.SharedRecap { return value },
		},
		{
			name: "year below range",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.Year = minSharedRecapYear - 1
				return value
			},
			wantErrContains: "year 2014 is out of range",
		},
		{
			name: "year above range",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.Year = maxSharedRecapYear + 1
				return value
			},
			wantErrContains: "year 2027 is out of range",
		},
		{
			name: "empty display name",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.DisplayName = "  "
				return value
			},
			wantErrContains: "display name is empty",
		},
		{
			name: "long display name",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.DisplayName = strings.Repeat("я", maxDisplayNameRunes+1)
				return value
			},
			wantErrContains: "display name exceeds maximum length 64",
		},
		{
			name: "negative active days",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.ActiveDays = -1
				return value
			},
			wantErrContains: "active days -1 is out of range",
		},
		{
			name: "too many active days",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.ActiveDays = maxActiveDays + 1
				return value
			},
			wantErrContains: "active days 367 is out of range",
		},
		{
			name: "negative views",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.Views = &negativeViews
				return value
			},
			wantErrContains: "views must not be negative",
		},
		{
			name: "empty category",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.TopCategory = &entity.SharedCategory{}
				return value
			},
			wantErrContains: "category title is empty",
		},
		{
			name: "long subcategory",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.TopCategory = &entity.SharedCategory{
					CategoryTitle:    "Транспорт",
					SubcategoryTitle: strings.Repeat("я", maxCategoryRunes+1),
				}
				return value
			},
			wantErrContains: "subcategory title exceeds maximum length 128",
		},
		{
			name: "long interest summary",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.InterestSummary = strings.Repeat("я", maxInterestRunes+1)
				return value
			},
			wantErrContains: "interest summary exceeds maximum length 512",
		},
		{
			name: "too many badges",
			mutate: func(value entity.SharedRecap) entity.SharedRecap {
				value.Badges = make([]entity.SharedBadge, maxSharedBadges+1)
				return value
			}, wantErrContains: "badges count 4 exceeds maximum 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateSnapshot(tt.mutate(valid))

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		optional        bool
		value           string
		maxRunes        int
		wantErrContains string
	}{
		{
			name:     "required valid unicode",
			value:    "Анна",
			maxRunes: 4,
		},
		{
			name:            "required blank",
			value:           " \t",
			maxRunes:        2,
			wantErrContains: "field is empty",
		},
		{
			name:            "required too long in runes",
			value:           "абв",
			maxRunes:        2,
			wantErrContains: "field exceeds maximum length 2",
		},
		{
			name:     "optional empty",
			optional: true,
			maxRunes: 2,
		},
		{
			name:            "optional blank is invalid",
			optional:        true,
			value:           "  ",
			maxRunes:        2,
			wantErrContains: "field is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var err error
			if tt.optional {
				err = validateOptionalText("field", tt.value, tt.maxRunes)
			} else {
				err = validateRequiredText("field", tt.value, tt.maxRunes)
			}

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
