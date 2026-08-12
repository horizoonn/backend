package sharedrecap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestPublicArchetype(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		archetype       entity.ArchetypeName
		want            entity.SharedArchetype
		wantErrContains string
	}{
		{
			name:      "collector",
			archetype: entity.ArchetypeCollector,
			want: entity.SharedArchetype{
				Name:        entity.ArchetypeCollector,
				Title:       "Коллекционер",
				Description: "Умение замечать интересные предложения и сохранять находки на будущее.",
			},
		},
		{
			name:      "dealmaker",
			archetype: entity.ArchetypeDealmaker,
			want: entity.SharedArchetype{
				Name:        entity.ArchetypeDealmaker,
				Title:       "Делец",
				Description: "Активное участие в сделках и умение находить новые возможности.",
			},
		},
		{
			name:      "negotiator",
			archetype: entity.ArchetypeNegotiator,
			want: entity.SharedArchetype{
				Name:        entity.ArchetypeNegotiator,
				Title:       "Переговорщик",
				Description: "Готовность обсуждать детали и находить подходящие условия.",
			},
		},
		{
			name:      "explorer",
			archetype: entity.ArchetypeExplorer,
			want: entity.SharedArchetype{
				Name:        entity.ArchetypeExplorer,
				Title:       "Исследователь",
				Description: "Интерес к разным категориям и постоянный поиск новых находок.",
			},
		},
		{
			name:            "unknown",
			archetype:       "unknown",
			wantErrContains: `unknown archetype "unknown"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := publicArchetype(tt.archetype)

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPublicBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		code  string
		want  entity.SharedBadge
		known bool
	}{
		{
			name:  "buyer bronze",
			code:  "buyer_bronze",
			known: true,
			want: entity.SharedBadge{
				Code:        "buyer_bronze",
				Title:       "Первая покупка",
				Description: "Первый шаг в поиске и покупке подходящих вещей.",
				Level:       entity.BadgeLevelBronze,
			},
		},
		{
			name:  "buyer silver",
			code:  "buyer_silver",
			known: true,
			want: entity.SharedBadge{
				Code:        "buyer_silver",
				Title:       "Уверенный покупатель",
				Description: "Внимательный выбор вещей и уверенность в покупательском сценарии.",
				Level:       entity.BadgeLevelSilver,
			},
		},
		{
			name:  "buyer gold",
			code:  "buyer_gold",
			known: true,
			want: entity.SharedBadge{
				Code:        "buyer_gold",
				Title:       "Знаток покупок",
				Description: "Умение находить подходящие вещи на Авито.",
				Level:       entity.BadgeLevelGold,
			},
		},
		{
			name:  "seller bronze",
			code:  "seller_bronze",
			known: true,
			want: entity.SharedBadge{
				Code:        "seller_bronze",
				Title:       "Первая продажа",
				Description: "Первый шаг в передаче вещей новым владельцам.",
				Level:       entity.BadgeLevelBronze,
			},
		},
		{
			name:  "seller silver",
			code:  "seller_silver",
			known: true,
			want: entity.SharedBadge{
				Code:        "seller_silver",
				Title:       "Опытный продавец",
				Description: "Уверенный подход к размещению и продаже вещей.",
				Level:       entity.BadgeLevelSilver,
			},
		},
		{
			name:  "seller gold",
			code:  "seller_gold",
			known: true,
			want: entity.SharedBadge{
				Code:        "seller_gold",
				Title:       "Мастер продаж",
				Description: "Умение находить новых владельцев для разных вещей.",
				Level:       entity.BadgeLevelGold,
			},
		},
		{name: "unknown", code: "private_badge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, known := publicBadge(tt.code)

			require.Equal(t, tt.known, known)
			require.Equal(t, tt.want, got)
		})
	}
}
