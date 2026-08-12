package sharedrecap

import (
	"time"

	"github.com/google/uuid"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type sharedRecapModel struct {
	Token     string
	RecapID   uuid.UUID
	Snapshot  sharedRecapSnapshotModel
	CreatedAt time.Time
}

func (m *sharedRecapModel) Scan(row rowScanner) error {
	return row.Scan(
		&m.Token,
		&m.RecapID,
		&m.Snapshot,
		&m.CreatedAt,
	)
}

type sharedRecapSnapshotModel struct {
	Year            int32                `json:"year"`
	DisplayName     string               `json:"displayName"`
	Archetype       sharedArchetypeModel `json:"archetype"`
	ActiveDays      int32                `json:"activeDays"`
	Views           *int64               `json:"views,omitempty"`
	TopCategory     *sharedCategoryModel `json:"topCategory,omitempty"`
	InterestSummary string               `json:"interestSummary,omitempty"`
	Badges          []sharedBadgeModel   `json:"badges"`
}

type sharedArchetypeModel struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type sharedCategoryModel struct {
	CategoryTitle    string `json:"categoryTitle"`
	SubcategoryTitle string `json:"subcategoryTitle,omitempty"`
}

type sharedBadgeModel struct {
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Level       string  `json:"level"`
	IconURL     *string `json:"iconUrl,omitempty"`
}
