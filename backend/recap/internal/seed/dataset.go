package seed

import (
	"time"

	"github.com/google/uuid"
)

type Dataset struct {
	Year          int
	Seed          uint64
	Users         []UserRow
	Categories    []CategoryRow
	Subcategories []SubcategoryRow
	Listings      []ListingRow
	Views         []ViewRow
	Favorites     []FavoriteRow
	Messages      []MessageRow
	Deals         []DealRow
}

type UserRow struct {
	ID           uuid.UUID
	Code         string
	Name         string
	Surname      string
	AvatarURL    *string
	Hint         *string
	RegisteredAt time.Time
}

type CategoryRow struct {
	ID    uuid.UUID
	Code  CategoryCode
	Title string
}

type SubcategoryRow struct {
	ID         uuid.UUID
	Code       string
	CategoryID uuid.UUID
	Title      string
}

type ListingStatus string

const (
	ListingActive ListingStatus = "active"
	ListingSold   ListingStatus = "sold"
	ListingClosed ListingStatus = "closed"
)

type ListingRow struct {
	ID            uuid.UUID
	SellerID      uuid.UUID
	Title         string
	Description   *string
	Price         int64
	CategoryID    uuid.UUID
	SubcategoryID *uuid.UUID
	Status        ListingStatus
	CreatedAt     time.Time
	ClosedAt      *time.Time
}

type ViewRow struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ListingID uuid.UUID
	ViewedAt  time.Time
}

type FavoriteRow struct {
	UserID    uuid.UUID
	ListingID uuid.UUID
	CreatedAt time.Time
}

type MessageRow struct {
	ID        uuid.UUID
	BuyerID   uuid.UUID
	SellerID  uuid.UUID
	ListingID uuid.UUID
	CreatedAt time.Time
}

type DealRow struct {
	ID          uuid.UUID
	ListingID   uuid.UUID
	BuyerID     uuid.UUID
	Price       int64
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type Summary struct {
	Users         int
	Categories    int
	Subcategories int
	Listings      int
	Views         int
	Favorites     int
	Messages      int
	Deals         int
}

func (d Dataset) Summary() Summary {
	return Summary{
		Users:         len(d.Users),
		Categories:    len(d.Categories),
		Subcategories: len(d.Subcategories),
		Listings:      len(d.Listings),
		Views:         len(d.Views),
		Favorites:     len(d.Favorites),
		Messages:      len(d.Messages),
		Deals:         len(d.Deals),
	}
}
