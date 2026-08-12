package activity

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func categoryActivityModelToEntity(model categoryActivityModel) entity.CategoryActivity {
	return entity.CategoryActivity{
		CategoryID:       model.CategoryID,
		CategoryTitle:    model.CategoryTitle,
		SubcategoryID:    nullableUUIDValue(model.SubcategoryID),
		SubcategoryTitle: nullableTextZeroValue(model.SubcategoryTitle),
		Views:            model.Views,
		Favorites:        model.Favorites,
		Purchases:        model.Purchases,
		Sales:            model.Sales,
	}
}

func nullableUUIDValue(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	id := uuid.UUID(value.Bytes)

	return &id
}

func nullableTextZeroValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}
