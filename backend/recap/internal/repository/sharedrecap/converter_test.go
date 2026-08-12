package sharedrecap

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestSharedRecapConversionRoundTrip(t *testing.T) {
	t.Parallel()

	want := validSharedRecapForTest()
	model, err := sharedRecapToModel(want)
	if err != nil {
		t.Fatalf("sharedRecapToModel() error = %v", err)
	}

	got, err := sharedRecapModelToEntity(model)
	if err != nil {
		t.Fatalf("sharedRecapModelToEntity() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip differs:\ngot  %#v\nwant %#v", got, want)
	}
}

func TestSharedRecapSnapshotContainsOnlyPublicFields(t *testing.T) {
	t.Parallel()

	model, err := sharedRecapToModel(validSharedRecapForTest())
	if err != nil {
		t.Fatalf("sharedRecapToModel() error = %v", err)
	}

	payload, err := json.Marshal(model.Snapshot)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	for _, forbidden := range []string{
		"token",
		"recapId",
		"surname",
		"favorites",
		"purchases",
		"sales",
		"messages",
		"reasons",
	} {
		if strings.Contains(string(payload), `"`+forbidden+`"`) {
			t.Fatalf("snapshot contains forbidden field %q: %s", forbidden, payload)
		}
	}
}

func TestSharedRecapToModelRejectsInvalidEnums(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*entity.SharedRecap)
		want   string
	}{
		{
			name: "invalid archetype",
			mutate: func(sharedRecap *entity.SharedRecap) {
				sharedRecap.Archetype.Name = "unknown"
			},
			want: "invalid archetype",
		},
		{
			name: "invalid badge level",
			mutate: func(sharedRecap *entity.SharedRecap) {
				sharedRecap.Badges[0].Level = "platinum"
			},
			want: "invalid level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sharedRecap := validSharedRecapForTest()
			tt.mutate(&sharedRecap)

			_, err := sharedRecapToModel(sharedRecap)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("sharedRecapToModel() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestSharedRecapModelToEntityRejectsInvalidEnums(t *testing.T) {
	t.Parallel()

	validModel, err := sharedRecapToModel(validSharedRecapForTest())
	if err != nil {
		t.Fatalf("sharedRecapToModel() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*sharedRecapModel)
		want   string
	}{
		{
			name: "invalid archetype",
			mutate: func(model *sharedRecapModel) {
				model.Snapshot.Archetype.Code = "unknown"
			},
			want: "invalid archetype",
		},
		{
			name: "invalid badge level",
			mutate: func(model *sharedRecapModel) {
				model.Snapshot.Badges[0].Level = "platinum"
			},
			want: "invalid level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := validModel
			model.Snapshot.Badges = append([]sharedBadgeModel(nil), validModel.Snapshot.Badges...)
			tt.mutate(&model)

			_, err := sharedRecapModelToEntity(model)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("sharedRecapModelToEntity() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func validSharedRecapForTest() entity.SharedRecap {
	views := int64(1248)
	iconURL := "https://example.com/badges/explorer.svg"

	return entity.SharedRecap{
		Token:       "h8T5P0Bm7rWqLk6uDe2Ftg",
		RecapID:     uuid.New(),
		Year:        2026,
		DisplayName: "Пётр",
		Archetype: entity.SharedArchetype{
			Name:        entity.ArchetypeExplorer,
			Title:       "Исследователь",
			Description: "Интерес к разным категориям и постоянный поиск новых находок.",
		},
		ActiveDays: 243,
		Views:      &views,
		TopCategory: &entity.SharedCategory{
			CategoryTitle:    "Электроника",
			SubcategoryTitle: "Смартфоны",
		},
		InterestSummary: "Зимой чаще привлекала техника, а летом — велосипеды.",
		Badges: []entity.SharedBadge{
			{
				Code:        "explorer_gold",
				Title:       "Любознательность",
				Description: "Интерес к разным категориям и новым находкам.",
				Level:       entity.BadgeLevelGold,
				IconURL:     &iconURL,
			},
		},
		CreatedAt: time.Date(2026, time.January, 9, 8, 35, 0, 0, time.UTC),
	}
}
