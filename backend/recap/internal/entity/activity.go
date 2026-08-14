package entity

import (
	"time"

	"github.com/google/uuid"
)

// Period is a half-open time range [From, To).
type Period struct {
	From time.Time
	To   time.Time
}

func (p Period) Valid() bool {
	return !p.From.IsZero() && !p.To.IsZero() && p.From.Before(p.To)
}

// YearPeriod returns the UTC calendar year as a period.
func YearPeriod(year int) Period {
	from := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

	return Period{From: from, To: from.AddDate(1, 0, 0)}
}

// Season is a part of the year used to show how interests changed.
type Season string

const (
	SeasonWinter Season = "winter"
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonAutumn Season = "autumn"
)

// SeasonWindow is a season inside a single calendar year. Winter carries two
// ranges because January-February and December belong to the same season but
// sit on the opposite ends of the year.
type SeasonWindow struct {
	Season Season
	Ranges []Period
}

// Seasons splits a calendar year into meteorological seasons.
func Seasons(year int) []SeasonWindow {
	monthStart := func(month time.Month) time.Time {
		return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	}

	return []SeasonWindow{
		{
			Season: SeasonWinter,
			Ranges: []Period{
				{From: monthStart(time.January), To: monthStart(time.March)},
				{From: monthStart(time.December), To: monthStart(time.December).AddDate(0, 1, 0)},
			},
		},
		{
			Season: SeasonSpring,
			Ranges: []Period{{From: monthStart(time.March), To: monthStart(time.June)}},
		},
		{
			Season: SeasonSummer,
			Ranges: []Period{{From: monthStart(time.June), To: monthStart(time.September)}},
		},
		{
			Season: SeasonAutumn,
			Ranges: []Period{{From: monthStart(time.September), To: monthStart(time.December)}},
		},
	}
}

// CategoryActivity is what a user did inside one category/subcategory pair
// during a period. Subcategory fields are empty for listings without one.
type CategoryActivity struct {
	CategoryID       uuid.UUID
	CategoryTitle    string
	SubcategoryID    *uuid.UUID
	SubcategoryTitle string
	Views            int64
	Favorites        int64
	Purchases        int64
	Sales            int64
}

type DayActivity struct {
	Date    time.Time
	Actions int64
}

// CategoryScore is a category ranked by the weighted activity score.
type CategoryScore struct {
	CategoryID  uuid.UUID
	Title       string
	Score       float64
	Subcategory *SubcategoryScore
}

// SubcategoryScore is the strongest subcategory inside a scored category.
type SubcategoryScore struct {
	ID    uuid.UUID
	Title string
	Score float64
}

// UserActivity holds the counters a recap is built from. Amounts are in kopecks
// and never leave the service as exact numbers - the API exposes ranges only.
type UserActivity struct {
	ActiveDays         int64
	Views              int64
	UniqueListingsSeen int64
	Favorites          int64
	FavoritesActive    int64
	Purchases          int64
	Sales              int64
	SalesAmount        int64
	MessagesAsBuyer    int64
	MessagesAsSeller   int64
	// CategoriesTouched is how many distinct categories the user met during the
	// year. It separates a broad explorer from someone digging in one category.
	CategoriesTouched int64
	// ListingsCreated counts published listings, not closed deals: a user who
	// posted twenty listings and sold two still behaves like a seller.
	ListingsCreated int64
}

// Messages returns the total number of messages in both roles.
func (a UserActivity) Messages() int64 {
	return a.MessagesAsBuyer + a.MessagesAsSeller
}

func (a UserActivity) TotalActions() int64 {
	return a.Views +
		a.Favorites +
		a.Purchases +
		a.Sales +
		a.Messages() +
		a.ListingsCreated
}
