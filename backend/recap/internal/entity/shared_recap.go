package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	SharedRecapTokenLength            = 22
	BadgeLevelBronze       BadgeLevel = "bronze"
	BadgeLevelSilver       BadgeLevel = "silver"
	BadgeLevelGold         BadgeLevel = "gold"
)

type (
	SharedRecapToken string
	BadgeLevel       string
)

type SharedRecap struct {
	Token           SharedRecapToken
	RecapID         uuid.UUID
	Year            int32
	DisplayName     string
	Archetype       SharedArchetype
	ActiveDays      int32
	Views           *int64
	TopCategory     *SharedCategory
	InterestSummary string
	Badges          []SharedBadge
	CreatedAt       time.Time
}

type SharedArchetype struct {
	Name        ArchetypeName
	Title       string
	Description string
}

type SharedCategory struct {
	CategoryTitle    string
	SubcategoryTitle string
}

type SharedBadge struct {
	Code        string
	Title       string
	Description string
	Level       BadgeLevel
	IconURL     *string
}

func (l BadgeLevel) Valid() bool {
	switch l {
	case BadgeLevelBronze, BadgeLevelSilver, BadgeLevelGold:
		return true
	default:
		return false
	}
}

type SharedRecapCreation struct {
	Token     SharedRecapToken
	CreatedAt time.Time
	Created   bool
}
