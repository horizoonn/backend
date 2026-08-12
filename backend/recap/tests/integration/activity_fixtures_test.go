//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

type activityFixture struct {
	userID           uuid.UUID
	categoryID       uuid.UUID
	subcategoryID    uuid.UUID
	secondCategoryID uuid.UUID
	period           entity.Period
}

func insertActivityFixture(t *testing.T) activityFixture {
	t.Helper()

	period := entity.YearPeriod(2025)
	secondDay := period.From.Add(24 * time.Hour)
	beforePeriod := period.From.AddDate(0, -1, 0)

	userID := uuid.New()
	otherUserID := uuid.New()
	categoryID := uuid.New()
	subcategoryID := uuid.New()
	secondCategoryID := uuid.New()
	outsideCategoryID := uuid.New()
	activeListingID := uuid.New()
	inactiveListingID := uuid.New()
	purchaseListingID := uuid.New()
	ownListingID := uuid.New()
	outOfPeriodListingID := uuid.New()

	insertProfile(t, profileFixture{
		ID:           userID,
		Name:         "Анна",
		Surname:      "Иванова",
		RegisteredAt: beforePeriod,
	})
	insertProfile(t, profileFixture{
		ID:           otherUserID,
		Name:         "Иван",
		Surname:      "Петров",
		RegisteredAt: beforePeriod,
	})

	mustExec(
		t,
		"INSERT INTO recap.categories (id, title) VALUES ($1, $2)",
		categoryID,
		"Авто",
	)
	mustExec(
		t,
		"INSERT INTO recap.categories (id, title) VALUES ($1, $2)",
		secondCategoryID,
		"Дом",
	)
	mustExec(
		t,
		"INSERT INTO recap.categories (id, title) VALUES ($1, $2)",
		outsideCategoryID,
		"Хобби",
	)
	mustExec(
		t,
		`INSERT INTO recap.subcategories (id, category_id, title)
		 VALUES ($1, $2, $3)`,
		subcategoryID,
		categoryID,
		"Автомобили",
	)

	insertListing(
		t,
		activeListingID,
		otherUserID,
		categoryID,
		&subcategoryID,
		"active",
		beforePeriod,
	)
	insertListing(
		t,
		inactiveListingID,
		otherUserID,
		secondCategoryID,
		nil,
		"sold",
		beforePeriod,
	)
	insertListing(
		t,
		purchaseListingID,
		otherUserID,
		categoryID,
		&subcategoryID,
		"sold",
		beforePeriod,
	)
	insertListing(
		t,
		ownListingID,
		userID,
		secondCategoryID,
		nil,
		"sold",
		period.From,
	)
	insertListing(
		t,
		outOfPeriodListingID,
		otherUserID,
		outsideCategoryID,
		nil,
		"active",
		beforePeriod,
	)

	insertView(t, userID, activeListingID, period.From)
	insertView(t, userID, activeListingID, period.From.Add(time.Hour))
	insertView(t, userID, inactiveListingID, secondDay)
	insertView(t, userID, outOfPeriodListingID, period.To)

	mustExec(
		t,
		`INSERT INTO recap.favorites (user_id, listing_id, created_at)
		 VALUES ($1, $2, $3), ($1, $4, $5)`,
		userID,
		activeListingID,
		period.From.Add(2*time.Hour),
		inactiveListingID,
		secondDay.Add(time.Hour),
	)

	insertDeal(
		t,
		purchaseListingID,
		userID,
		5_000,
		secondDay,
	)
	insertDeal(
		t,
		ownListingID,
		otherUserID,
		7_000,
		secondDay.Add(2*time.Hour),
	)

	insertMessage(
		t,
		userID,
		otherUserID,
		activeListingID,
		period.From.Add(3*time.Hour),
	)
	insertMessage(
		t,
		otherUserID,
		userID,
		ownListingID,
		secondDay.Add(3*time.Hour),
	)
	insertMessage(
		t,
		userID,
		otherUserID,
		outOfPeriodListingID,
		period.To,
	)

	return activityFixture{
		userID:           userID,
		categoryID:       categoryID,
		subcategoryID:    subcategoryID,
		secondCategoryID: secondCategoryID,
		period:           period,
	}
}

func insertListing(
	t *testing.T,
	id uuid.UUID,
	sellerID uuid.UUID,
	categoryID uuid.UUID,
	subcategoryID *uuid.UUID,
	status string,
	createdAt time.Time,
) {
	t.Helper()

	mustExec(
		t,
		`INSERT INTO recap.listings (
			id,
			seller_id,
			title,
			price,
			category_id,
			subcategory_id,
			status,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id,
		sellerID,
		"Тестовое объявление",
		10_000,
		categoryID,
		subcategoryID,
		status,
		createdAt,
	)
}

func insertView(
	t *testing.T,
	userID uuid.UUID,
	listingID uuid.UUID,
	viewedAt time.Time,
) {
	t.Helper()

	mustExec(
		t,
		`INSERT INTO recap.views (id, user_id, listing_id, viewed_at)
		 VALUES ($1, $2, $3, $4)`,
		uuid.New(),
		userID,
		listingID,
		viewedAt,
	)
}

func insertDeal(
	t *testing.T,
	listingID uuid.UUID,
	buyerID uuid.UUID,
	price int64,
	completedAt time.Time,
) {
	t.Helper()

	mustExec(
		t,
		`INSERT INTO recap.deals (
			id,
			listing_id,
			buyer_id,
			price,
			created_at,
			completed_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(),
		listingID,
		buyerID,
		price,
		completedAt.Add(-time.Hour),
		completedAt,
	)
}

func insertMessage(
	t *testing.T,
	buyerID uuid.UUID,
	sellerID uuid.UUID,
	listingID uuid.UUID,
	createdAt time.Time,
) {
	t.Helper()

	mustExec(
		t,
		`INSERT INTO recap.messages (
			id,
			buyer_id,
			seller_id,
			listing_id,
			created_at
		) VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		buyerID,
		sellerID,
		listingID,
		createdAt,
	)
}
