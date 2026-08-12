package sharedrecap

import (
	"fmt"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func publicArchetype(name entity.ArchetypeName) (entity.SharedArchetype, error) {
	switch name {
	case entity.ArchetypeCollector:
		return entity.SharedArchetype{
			Name:        name,
			Title:       "Коллекционер",
			Description: "Умение замечать интересные предложения и сохранять находки на будущее.",
		}, nil
	case entity.ArchetypeDealmaker:
		return entity.SharedArchetype{
			Name:        name,
			Title:       "Делец",
			Description: "Активное участие в сделках и умение находить новые возможности.",
		}, nil
	case entity.ArchetypeNegotiator:
		return entity.SharedArchetype{
			Name:        name,
			Title:       "Переговорщик",
			Description: "Готовность обсуждать детали и находить подходящие условия.",
		}, nil
	case entity.ArchetypeExplorer:
		return entity.SharedArchetype{
			Name:        name,
			Title:       "Исследователь",
			Description: "Интерес к разным категориям и постоянный поиск новых находок.",
		}, nil
	default:
		return entity.SharedArchetype{}, fmt.Errorf("unknown archetype %q", name)
	}
}

func publicBadge(code string) (entity.SharedBadge, bool) {
	switch code {
	case "buyer_bronze":
		return entity.SharedBadge{
			Code:        code,
			Title:       "Первая покупка",
			Description: "Первый шаг в поиске и покупке подходящих вещей.",
			Level:       entity.BadgeLevelBronze,
		}, true
	case "buyer_silver":
		return entity.SharedBadge{
			Code:        code,
			Title:       "Уверенный покупатель",
			Description: "Внимательный выбор вещей и уверенность в покупательском сценарии.",
			Level:       entity.BadgeLevelSilver,
		}, true
	case "buyer_gold":
		return entity.SharedBadge{
			Code:        code,
			Title:       "Знаток покупок",
			Description: "Умение находить подходящие вещи на Авито.",
			Level:       entity.BadgeLevelGold,
		}, true
	case "seller_bronze":
		return entity.SharedBadge{
			Code:        code,
			Title:       "Первая продажа",
			Description: "Первый шаг в передаче вещей новым владельцам.",
			Level:       entity.BadgeLevelBronze,
		}, true
	case "seller_silver":
		return entity.SharedBadge{
			Code:        code,
			Title:       "Опытный продавец",
			Description: "Уверенный подход к размещению и продаже вещей.",
			Level:       entity.BadgeLevelSilver,
		}, true
	case "seller_gold":
		return entity.SharedBadge{
			Code:        code,
			Title:       "Мастер продаж",
			Description: "Умение находить новых владельцев для разных вещей.",
			Level:       entity.BadgeLevelGold,
		}, true
	default:
		return entity.SharedBadge{}, false
	}
}
