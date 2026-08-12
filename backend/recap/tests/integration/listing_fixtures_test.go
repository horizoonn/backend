//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

type listingFixture struct {
	ID            uuid.UUID
	SellerID      uuid.UUID
	Title         string
	Price         int64
	CategoryID    uuid.UUID
	SubcategoryID *uuid.UUID
	Status        string
	CreatedAt     time.Time
}

type categoryRecommendationsFixture struct {
	userID                 uuid.UUID
	categoryID             uuid.UUID
	preferredSubcategoryID uuid.UUID
	preferredNewest        listingFixture
	preferredOlder         listingFixture
	nonPreferredNewest     listingFixture
}

type oldestActiveFavoritesFixture struct {
	userID uuid.UUID
	period entity.Period
	oldest listingFixture
	second listingFixture
}

func insertCategoryRecommendationsFixture(
	t *testing.T,
) categoryRecommendationsFixture {
	t.Helper()

	userID, sellerID := insertListingUsers(t)
	categoryID := uuid.New()
	otherCategoryID := uuid.New()
	preferredSubcategoryID := uuid.New()
	otherSubcategoryID := uuid.New()
	createdAt := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	insertListingCategory(t, categoryID, "Электроника")
	insertListingCategory(t, otherCategoryID, "Дом")
	insertListingSubcategory(t, preferredSubcategoryID, categoryID, "Смартфоны")
	insertListingSubcategory(t, otherSubcategoryID, categoryID, "Ноутбуки")

	preferredNewest := listingFixture{
		ID:            uuid.MustParse("00000000-0000-0000-0000-000000000101"),
		SellerID:      sellerID,
		Title:         "Новый смартфон",
		Price:         90_000,
		CategoryID:    categoryID,
		SubcategoryID: &preferredSubcategoryID,
		Status:        "active",
		CreatedAt:     createdAt.Add(48 * time.Hour),
	}
	preferredOlder := listingFixture{
		ID:            uuid.MustParse("00000000-0000-0000-0000-000000000102"),
		SellerID:      sellerID,
		Title:         "Смартфон",
		Price:         50_000,
		CategoryID:    categoryID,
		SubcategoryID: &preferredSubcategoryID,
		Status:        "active",
		CreatedAt:     createdAt.Add(24 * time.Hour),
	}
	nonPreferredNewest := listingFixture{
		ID:            uuid.MustParse("00000000-0000-0000-0000-000000000103"),
		SellerID:      sellerID,
		Title:         "Ноутбук",
		Price:         120_000,
		CategoryID:    categoryID,
		SubcategoryID: &otherSubcategoryID,
		Status:        "active",
		CreatedAt:     createdAt.Add(72 * time.Hour),
	}
	nonPreferredOlder := listingFixture{
		ID:            uuid.MustParse("00000000-0000-0000-0000-000000000104"),
		SellerID:      sellerID,
		Title:         "Старый ноутбук",
		Price:         40_000,
		CategoryID:    categoryID,
		SubcategoryID: &otherSubcategoryID,
		Status:        "active",
		CreatedAt:     createdAt,
	}

	for _, fixture := range []listingFixture{
		preferredNewest,
		preferredOlder,
		nonPreferredNewest,
		nonPreferredOlder,
		{
			ID:            uuid.New(),
			SellerID:      userID,
			Title:         "Собственное объявление",
			Price:         1,
			CategoryID:    categoryID,
			SubcategoryID: &preferredSubcategoryID,
			Status:        "active",
			CreatedAt:     createdAt.Add(96 * time.Hour),
		},
		{
			ID:            uuid.New(),
			SellerID:      sellerID,
			Title:         "Неактивное объявление",
			Price:         1,
			CategoryID:    categoryID,
			SubcategoryID: &preferredSubcategoryID,
			Status:        "closed",
			CreatedAt:     createdAt.Add(96 * time.Hour),
		},
		{
			ID:         uuid.New(),
			SellerID:   sellerID,
			Title:      "Другая категория",
			Price:      1,
			CategoryID: otherCategoryID,
			Status:     "active",
			CreatedAt:  createdAt.Add(96 * time.Hour),
		},
	} {
		insertListingRecord(t, fixture)
	}

	favoritedID := uuid.New()
	insertListingRecord(t, listingFixture{
		ID:            favoritedID,
		SellerID:      sellerID,
		Title:         "Уже в избранном",
		Price:         1,
		CategoryID:    categoryID,
		SubcategoryID: &preferredSubcategoryID,
		Status:        "active",
		CreatedAt:     createdAt.Add(96 * time.Hour),
	})
	insertFavorite(t, userID, favoritedID, createdAt)

	soldID := uuid.New()
	insertListingRecord(t, listingFixture{
		ID:            soldID,
		SellerID:      sellerID,
		Title:         "Уже продано",
		Price:         1,
		CategoryID:    categoryID,
		SubcategoryID: &preferredSubcategoryID,
		Status:        "active",
		CreatedAt:     createdAt.Add(96 * time.Hour),
	})
	insertCompletedDeal(t, soldID, userID, createdAt.Add(97*time.Hour))

	return categoryRecommendationsFixture{
		userID:                 userID,
		categoryID:             categoryID,
		preferredSubcategoryID: preferredSubcategoryID,
		preferredNewest:        preferredNewest,
		preferredOlder:         preferredOlder,
		nonPreferredNewest:     nonPreferredNewest,
	}
}

func insertOldestActiveFavoritesFixture(
	t *testing.T,
) oldestActiveFavoritesFixture {
	t.Helper()

	userID, sellerID := insertListingUsers(t)
	categoryID := uuid.New()
	period := entity.YearPeriod(2025)
	insertListingCategory(t, categoryID, "Электроника")

	oldest := listingFixture{
		ID:         uuid.MustParse("00000000-0000-0000-0000-000000000201"),
		SellerID:   sellerID,
		Title:      "Первое избранное",
		Price:      10_000,
		CategoryID: categoryID,
		Status:     "active",
		CreatedAt:  period.From.AddDate(0, -1, 0),
	}
	second := listingFixture{
		ID:         uuid.MustParse("00000000-0000-0000-0000-000000000202"),
		SellerID:   sellerID,
		Title:      "Второе избранное",
		Price:      20_000,
		CategoryID: categoryID,
		Status:     "active",
		CreatedAt:  period.From.AddDate(0, -1, 0),
	}
	third := listingFixture{
		ID:         uuid.MustParse("00000000-0000-0000-0000-000000000203"),
		SellerID:   sellerID,
		Title:      "Третье избранное",
		Price:      30_000,
		CategoryID: categoryID,
		Status:     "active",
		CreatedAt:  period.From.AddDate(0, -1, 0),
	}

	for _, fixture := range []listingFixture{oldest, second, third} {
		insertListingRecord(t, fixture)
	}
	insertFavorite(t, userID, oldest.ID, period.From)
	insertFavorite(t, userID, second.ID, period.From.Add(24*time.Hour))
	insertFavorite(t, userID, third.ID, period.From.Add(48*time.Hour))

	insertFilteredFavorites(t, userID, sellerID, categoryID, period)

	return oldestActiveFavoritesFixture{
		userID: userID,
		period: period,
		oldest: oldest,
		second: second,
	}
}

func insertFilteredFavorites(
	t *testing.T,
	userID uuid.UUID,
	sellerID uuid.UUID,
	categoryID uuid.UUID,
	period entity.Period,
) {
	t.Helper()

	tests := []struct {
		listing listingFixture
		addedAt time.Time
		deal    bool
	}{
		{
			listing: listingFixture{
				ID: uuid.New(), SellerID: sellerID, Title: "До периода", Price: 1,
				CategoryID: categoryID, Status: "active", CreatedAt: period.From.AddDate(0, -1, 0),
			},
			addedAt: period.From.Add(-time.Nanosecond),
		},
		{
			listing: listingFixture{
				ID: uuid.New(), SellerID: sellerID, Title: "После периода", Price: 1,
				CategoryID: categoryID, Status: "active", CreatedAt: period.From.AddDate(0, -1, 0),
			},
			addedAt: period.To,
		},
		{
			listing: listingFixture{
				ID: uuid.New(), SellerID: sellerID, Title: "Неактивное", Price: 1,
				CategoryID: categoryID, Status: "closed", CreatedAt: period.From.AddDate(0, -1, 0),
			},
			addedAt: period.From,
		},
		{
			listing: listingFixture{
				ID: uuid.New(), SellerID: userID, Title: "Собственное", Price: 1,
				CategoryID: categoryID, Status: "active", CreatedAt: period.From.AddDate(0, -1, 0),
			},
			addedAt: period.From,
		},
		{
			listing: listingFixture{
				ID: uuid.New(), SellerID: sellerID, Title: "Купленное", Price: 1,
				CategoryID: categoryID, Status: "active", CreatedAt: period.From.AddDate(0, -1, 0),
			},
			addedAt: period.From,
			deal:    true,
		},
	}

	for _, tt := range tests {
		insertListingRecord(t, tt.listing)
		insertFavorite(t, userID, tt.listing.ID, tt.addedAt)
		if tt.deal {
			insertCompletedDeal(t, tt.listing.ID, userID, period.From.Add(time.Hour))
		}
	}
}

func insertListingUsers(t *testing.T) (uuid.UUID, uuid.UUID) {
	t.Helper()

	registeredAt := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	userID := uuid.New()
	sellerID := uuid.New()
	insertProfile(t, profileFixture{
		ID: userID, Name: "Анна", Surname: "Иванова", RegisteredAt: registeredAt,
	})
	insertProfile(t, profileFixture{
		ID: sellerID, Name: "Иван", Surname: "Петров", RegisteredAt: registeredAt,
	})

	return userID, sellerID
}

func insertListingCategory(t *testing.T, id uuid.UUID, title string) {
	t.Helper()

	mustExec(t, "INSERT INTO recap.categories (id, title) VALUES ($1, $2)", id, title)
}

func insertListingSubcategory(
	t *testing.T,
	id uuid.UUID,
	categoryID uuid.UUID,
	title string,
) {
	t.Helper()

	mustExec(
		t,
		`INSERT INTO recap.subcategories (id, category_id, title)
		 VALUES ($1, $2, $3)`,
		id,
		categoryID,
		title,
	)
}

func insertListingRecord(t *testing.T, fixture listingFixture) {
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
		fixture.ID,
		fixture.SellerID,
		fixture.Title,
		fixture.Price,
		fixture.CategoryID,
		fixture.SubcategoryID,
		fixture.Status,
		fixture.CreatedAt,
	)
}

func insertFavorite(
	t *testing.T,
	userID uuid.UUID,
	listingID uuid.UUID,
	addedAt time.Time,
) {
	t.Helper()

	mustExec(
		t,
		`INSERT INTO recap.favorites (user_id, listing_id, created_at)
		 VALUES ($1, $2, $3)`,
		userID,
		listingID,
		addedAt,
	)
}

func insertCompletedDeal(
	t *testing.T,
	listingID uuid.UUID,
	buyerID uuid.UUID,
	completedAt time.Time,
) {
	t.Helper()

	mustExec(
		t,
		`INSERT INTO recap.deals (
			id, listing_id, buyer_id, price, created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(),
		listingID,
		buyerID,
		1,
		completedAt.Add(-time.Hour),
		completedAt,
	)
}

func listingPreview(fixture listingFixture) entity.ListingPreview {
	return entity.ListingPreview{
		ID:         fixture.ID,
		Title:      fixture.Title,
		Price:      fixture.Price,
		CategoryID: fixture.CategoryID,
	}
}
