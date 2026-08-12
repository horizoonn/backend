package entity

import (
	"time"

	"github.com/google/uuid"
)

type ListingPreview struct {
	ID         uuid.UUID
	Title      string
	Price      int64
	CategoryID uuid.UUID
}

type FavoriteListingPreview struct {
	ListingPreview
	AddedAt time.Time
}
