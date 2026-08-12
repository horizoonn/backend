package activity

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestCategoryActivityModelToEntity(t *testing.T) {
	t.Parallel()

	categoryID := uuid.New()
	subcategoryID := uuid.New()

	tests := []struct {
		name  string
		model categoryActivityModel
		want  entity.CategoryActivity
	}{
		{
			name: "with subcategory",
			model: categoryActivityModel{
				CategoryID:       categoryID,
				CategoryTitle:    "Электроника",
				SubcategoryID:    pgtype.UUID{Bytes: subcategoryID, Valid: true},
				SubcategoryTitle: pgtype.Text{String: "Телефоны", Valid: true},
				Views:            10,
				Favorites:        4,
				Purchases:        2,
				Sales:            1,
			},
			want: entity.CategoryActivity{
				CategoryID:       categoryID,
				CategoryTitle:    "Электроника",
				SubcategoryID:    &subcategoryID,
				SubcategoryTitle: "Телефоны",
				Views:            10,
				Favorites:        4,
				Purchases:        2,
				Sales:            1,
			},
		},
		{
			name: "without subcategory",
			model: categoryActivityModel{
				CategoryID:    categoryID,
				CategoryTitle: "Электроника",
			},
			want: entity.CategoryActivity{
				CategoryID:    categoryID,
				CategoryTitle: "Электроника",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := categoryActivityModelToEntity(tt.model)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("categoryActivityModelToEntity() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
