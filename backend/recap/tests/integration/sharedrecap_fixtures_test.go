//go:build integration

package integration

import (
	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func newSharedRecapFixture(recapID uuid.UUID) entity.SharedRecap {
	views := int64(125)
	iconURL := "https://example.test/badge.svg"

	return entity.SharedRecap{
		Token:       "abcdefghijklmnopqrstuv",
		RecapID:     recapID,
		Year:        2025,
		DisplayName: "Анна",
		Archetype: entity.SharedArchetype{
			Name:        entity.ArchetypeExplorer,
			Title:       "Исследователь года",
			Description: "Вы открывали новые категории.",
		},
		ActiveDays: 42,
		Views:      &views,
		TopCategory: &entity.SharedCategory{
			CategoryTitle:    "Транспорт",
			SubcategoryTitle: "Автомобили",
		},
		InterestSummary: "Главный интерес года — Транспорт.",
		Badges: []entity.SharedBadge{
			{
				Code:        "buyer_gold",
				Title:       "Знаток покупок",
				Description: "Умение находить подходящие вещи на Авито.",
				Level:       entity.BadgeLevelGold,
				IconURL:     &iconURL,
			},
		},
	}
}
