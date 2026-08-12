package listing

import (
	"time"

	"github.com/google/uuid"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type listingPreviewModel struct {
	ID         uuid.UUID
	Title      string
	Price      int64
	CategoryID uuid.UUID
}

func (m *listingPreviewModel) Scan(row rowScanner) error {
	return row.Scan(
		&m.ID,
		&m.Title,
		&m.Price,
		&m.CategoryID,
	)
}

type favoriteListingPreviewModel struct {
	listingPreviewModel
	AddedAt time.Time
}

func (m *favoriteListingPreviewModel) Scan(row rowScanner) error {
	return row.Scan(
		&m.ID,
		&m.Title,
		&m.Price,
		&m.CategoryID,
		&m.AddedAt,
	)
}
