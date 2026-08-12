package recap

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestRecapConversionRoundTrip(t *testing.T) {
	t.Parallel()

	want := entity.Recap{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Year:   2025,
		Archetype: entity.Archetype{
			UserArchetype: entity.ArchetypeCollector,
			Title:         "Коллекционер",
			Description:   "Сохраняет интересные находки",
			Reasons: []entity.ArchetypeReason{
				{
					Metric:      entity.MetricFavorites,
					Value:       "57",
					Explanation: "Много добавлений в избранное",
				},
			},
		},
		Slides:      []byte(`[{"type":"favorites"}]`),
		GeneratedAt: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC),
	}

	model, err := recapToModel(want)
	if err != nil {
		t.Fatalf("recapToModel() error = %v", err)
	}

	got, err := recapModelToEntity(model)
	if err != nil {
		t.Fatalf("recapModelToEntity() error = %v", err)
	}

	if got.ID != want.ID || got.UserID != want.UserID || got.Year != want.Year {
		t.Fatalf("round trip identifiers differ: got %+v, want %+v", got, want)
	}
	if got.Archetype.UserArchetype != want.Archetype.UserArchetype {
		t.Fatalf("round trip archetype differs: got %+v, want %+v", got, want)
	}
	if string(got.Slides) != string(want.Slides) {
		t.Fatalf("round trip slides = %s, want %s", got.Slides, want.Slides)
	}
}

func TestReasonsToEntityRejectsInvalidMetric(t *testing.T) {
	t.Parallel()

	_, err := reasonsToEntity([]archetypeReasonModel{
		{
			Metric:      "unknown",
			Value:       "1",
			Explanation: "Invalid metric",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "invalid metric") {
		t.Fatalf("reasonsToEntity() error = %v, want invalid metric", err)
	}
}

func TestRecapToModelRejectsInvalidJSONBInvariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		slides []byte
	}{
		{name: "invalid JSON", slides: []byte(`{`)},
		{name: "not an array", slides: []byte(`{"type":"intro"}`)},
		{name: "empty array", slides: []byte(`[]`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recap := validRecapForTest()
			recap.Slides = tt.slides

			if _, err := recapToModel(recap); err == nil {
				t.Fatal("recapToModel() error = nil, want validation error")
			}
		})
	}
}

func validRecapForTest() entity.Recap {
	return entity.Recap{
		UserID: uuid.New(),
		Year:   2025,
		Archetype: entity.Archetype{
			UserArchetype: entity.ArchetypeCollector,
			Title:         "Коллекционер",
			Description:   "Сохраняет интересные находки",
			Reasons: []entity.ArchetypeReason{
				{
					Metric:      entity.MetricFavorites,
					Value:       "57",
					Explanation: "Много добавлений в избранное",
				},
			},
		},
		Slides: []byte(`[{"type":"favorites"}]`),
	}
}
