package recap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAwardBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int64
		want  *badge
	}{
		{
			name:  "below minimum threshold",
			value: 0,
			want:  nil,
		},
		{
			name:  "bronze threshold",
			value: 1,
			want: &badge{
				Code:        "buyer_bronze",
				Title:       "Первая покупка",
				Description: "Есть закрытые покупки за год",
				Level:       badgeBronze,
			},
		},
		{
			name:  "silver threshold",
			value: 5,
			want: &badge{
				Code:        "buyer_silver",
				Title:       "Уверенный покупатель",
				Description: "5 и больше покупок за год",
				Level:       badgeSilver,
			},
		},
		{
			name:  "highest matching threshold",
			value: 100,
			want: &badge{
				Code:        "buyer_gold",
				Title:       "Знаток покупок",
				Description: "10 и больше покупок за год",
				Level:       badgeGold,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, awardBadge(purchaseBadges, test.value))
		})
	}
}
