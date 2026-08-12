package activity

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type categoryActivityModel struct {
	CategoryID       uuid.UUID
	CategoryTitle    string
	SubcategoryID    pgtype.UUID
	SubcategoryTitle pgtype.Text
	Views            int64
	Favorites        int64
	Purchases        int64
	Sales            int64
}

func (m *categoryActivityModel) Scan(row rowScanner) error {
	return row.Scan(
		&m.CategoryID,
		&m.CategoryTitle,
		&m.SubcategoryID,
		&m.SubcategoryTitle,
		&m.Views,
		&m.Favorites,
		&m.Purchases,
		&m.Sales,
	)
}
