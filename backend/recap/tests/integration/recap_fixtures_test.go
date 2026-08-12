//go:build integration

package integration

import (
	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func newRecapFixture(profileID uuid.UUID) entity.Recap {
	return entity.Recap{
		UserID: profileID,
		Year:   2025,
		Archetype: entity.Archetype{
			UserArchetype: entity.ArchetypeExplorer,
			Title:         "Исследователь года",
			Description:   "Вы открывали новые категории.",
			Reasons: []entity.ArchetypeReason{
				{
					Metric:      entity.MetricCategories,
					Value:       "5",
					Explanation: "Вы интересовались пятью категориями.",
				},
			},
		},
		Slides: []byte(`[{"type":"active_days","activeDays":42}]`),
	}
}
