package recap

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const (
	slideIntro            = "intro"
	slideActiveDays       = "active_days"
	slideViews            = "views"
	slideFavorites        = "favorites"
	slideFavoriteCategory = "favorite_category"
	slidePurchases        = "purchases"
	slideSales            = "sales"
	slideMessages         = "messages"
	slideInterests        = "interests"
	slideArchetype        = "archetype"
	slideFinal            = "final"
)

// Stat tiles of the final screen, the StatTile.code enum of the contract.
const (
	statActiveDays = "active_days"
	statViews      = "views"
	statFavorites  = "favorites"
	statMessages   = "messages"
	statSeasons    = "seasons"
)

const (
	ctaOpenCategory  = "open_category"
	ctaOpenFavorites = "open_favorites"
	ctaCreateListing = "create_listing"
	ctaShareRecap    = "share_recap"
)

const currencyRUB = "RUB"

type (
	slideBase struct {
		Type     string `json:"type"`
		Title    string `json:"title"`
		Subtitle string `json:"subtitle,omitempty"`
		Cta      *cta   `json:"cta,omitempty"`
	}

	cta struct {
		Action     string     `json:"action"`
		Title      string     `json:"title"`
		CategoryID *uuid.UUID `json:"categoryId,omitempty"`
	}

	categoryRef struct {
		ID    uuid.UUID `json:"id"`
		Title string    `json:"title"`
	}

	listingRef struct {
		ID         uuid.UUID  `json:"id"`
		Title      string     `json:"title"`
		Price      int64      `json:"price"`
		CategoryID uuid.UUID  `json:"categoryId"`
		AddedAt    *time.Time `json:"addedAt,omitempty"`
	}

	amountRange struct {
		Min      int64  `json:"min"`
		Max      *int64 `json:"max"`
		Currency string `json:"currency"`
		Label    string `json:"label"`
	}

	archetypeBody struct {
		Code        string               `json:"code"`
		Title       string               `json:"title"`
		Description string               `json:"description"`
		Reasons     []archetypeReasonRef `json:"reasons"`
	}

	archetypeReasonRef struct {
		Metric      string `json:"metric"`
		Value       string `json:"value"`
		Explanation string `json:"explanation"`
	}

	periodInterest struct {
		Period      string       `json:"period"`
		Category    categoryRef  `json:"category"`
		Subcategory *categoryRef `json:"subcategory,omitempty"`
		Weight      int32        `json:"weight"`
	}

	introSlide struct {
		slideBase
		Year int32 `json:"year"`
	}

	activeDaysSlide struct {
		slideBase
		ActiveDays int32 `json:"activeDays"`
	}

	viewsSlide struct {
		slideBase
		Views int64 `json:"views"`
	}

	favoritesSlide struct {
		slideBase
		Favorites      int64       `json:"favorites"`
		OldestFavorite *listingRef `json:"oldestFavorite,omitempty"`
		StillAvailable int32       `json:"stillAvailable"`
	}

	categorySlide struct {
		slideBase
		Category        categoryRef  `json:"category"`
		Subcategory     *categoryRef `json:"subcategory,omitempty"`
		Share           int32        `json:"share"`
		Recommendations []listingRef `json:"recommendations,omitempty"`
	}

	purchasesSlide struct {
		slideBase
		Purchases int32  `json:"purchases"`
		Badge     *badge `json:"badge,omitempty"`
	}

	salesSlide struct {
		slideBase
		Sales       int32        `json:"sales"`
		AmountRange *amountRange `json:"amountRange,omitempty"`
		Badge       *badge       `json:"badge,omitempty"`
	}

	messagesSlide struct {
		slideBase
		Messages int64 `json:"messages"`
	}

	interestsSlide struct {
		slideBase
		Periods      []periodInterest `json:"periods"`
		ShiftSummary string           `json:"shiftSummary,omitempty"`
	}

	archetypeSlide struct {
		slideBase
		Archetype archetypeBody `json:"archetype"`
	}

	statTile struct {
		Code  string `json:"code"`
		Value int64  `json:"value"`
		Label string `json:"label"`
	}

	finalSlide struct {
		slideBase
		Stats   []statTile `json:"stats"`
		Actions []cta      `json:"actions"`
	}
)

type seasonLeader struct {
	season   entity.Season
	category entity.CategoryScore
	share    int32
}

type slideInput struct {
	year                    int32
	activity                entity.UserActivity
	categories              []entity.CategoryScore
	seasons                 []seasonLeader
	archetype               entity.Archetype
	oldestFavorite          *entity.FavoriteListingPreview
	categoryRecommendations []entity.ListingPreview
}

// slideBuilder makes one screen. The second value is false when the screen has
// no data behind it: an empty "0 sales" card adds nothing but a swipe.
type slideBuilder func(input slideInput) (any, bool)

var storyboard = []slideBuilder{
	buildIntroSlide,
	buildActiveDaysSlide,
	buildViewsSlide,
	buildFavoritesSlide,
	buildCategorySlide,
	buildPurchasesSlide,
	buildSalesSlide,
	buildMessagesSlide,
	buildInterestsSlide,
	buildArchetypeSlide,
	buildFinalSlide,
}

func buildSlides(input slideInput) (json.RawMessage, error) {
	slides := make([]any, 0, len(storyboard))

	for _, build := range storyboard {
		if slide, ok := build(input); ok {
			slides = append(slides, slide)
		}
	}

	raw, err := json.Marshal(slides)
	if err != nil {
		return nil, fmt.Errorf("marshal slides: %w", err)
	}

	return raw, nil
}

func buildIntroSlide(input slideInput) (any, bool) {
	return introSlide{
		slideBase: slideBase{
			Type:     slideIntro,
			Title:    fmt.Sprintf("Вcпомни каким был %d год вместе с Авито", input.year),
			Subtitle: "Мы собрали для вас самые важные моменты вашей активности на площадке за этот год.",
		},
		Year: input.year,
	}, true
}

func buildActiveDaysSlide(input slideInput) (any, bool) {
	return activeDaysSlide{
		slideBase: slideBase{
			Type:     slideActiveDays,
			Title:    titleActiveDays,
			Subtitle: activeDaysHeadline(input.activity.ActiveDays),
		},
		ActiveDays: toInt32(input.activity.ActiveDays),
	}, true
}

func buildViewsSlide(input slideInput) (any, bool) {
	activity := input.activity
	if activity.Views == 0 {
		return nil, false
	}

	return viewsSlide{
		slideBase: slideBase{
			Type:     slideViews,
			Title:    titleViews,
			Subtitle: viewsHeadline(activity.Views),
		},
		Views: activity.Views,
	}, true
}

func buildFavoritesSlide(input slideInput) (any, bool) {
	activity := input.activity
	if activity.Favorites == 0 {
		return nil, false
	}

	return favoritesSlide{
		slideBase: slideBase{
			Type:     slideFavorites,
			Title:    titleFavorites,
			Subtitle: favoritesHeadline(activity.Favorites),
			Cta: &cta{
				Action: ctaOpenFavorites,
				Title:  "Посмотреть избранное",
			},
		},
		Favorites:      activity.Favorites,
		OldestFavorite: favoriteListingRefOf(input.oldestFavorite),
		StillAvailable: toInt32(activity.FavoritesActive),
	}, true
}

func buildPurchasesSlide(input slideInput) (any, bool) {
	activity := input.activity
	if activity.Purchases == 0 {
		return nil, false
	}

	return purchasesSlide{
		slideBase: slideBase{
			Type:     slidePurchases,
			Title:    titlePurchases,
			Subtitle: purchasesHeadline(activity.Purchases),
		},
		Purchases: toInt32(activity.Purchases),
		Badge:     awardBadge(purchaseBadges, activity.Purchases),
	}, true
}

func buildSalesSlide(input slideInput) (any, bool) {
	activity := input.activity
	if activity.Sales == 0 {
		return nil, false
	}

	return salesSlide{
		slideBase: slideBase{
			Type:     slideSales,
			Title:    titleSales,
			Subtitle: salesHeadline(activity.Sales),
		},
		Sales:       toInt32(activity.Sales),
		AmountRange: salesAmountRange(activity.SalesAmount),
		Badge:       awardBadge(saleBadges, activity.Sales),
	}, true
}

func buildMessagesSlide(input slideInput) (any, bool) {
	total := input.activity.Messages()
	if total == 0 {
		return nil, false
	}

	return messagesSlide{
		slideBase: slideBase{
			Type:     slideMessages,
			Title:    titleMessages,
			Subtitle: messagesHeadline(total),
		},
		Messages: total,
	}, true
}

func buildCategorySlide(input slideInput) (any, bool) {
	if len(input.categories) == 0 {
		return nil, false
	}

	favorite := input.categories[0]
	share := CategoryShare(input.categories, favorite)

	return categorySlide{
		slideBase: slideBase{
			Type:     slideFavoriteCategory,
			Title:    titleFavoriteCategory,
			Subtitle: categoryHeadline(favorite, share),
			Cta: &cta{
				Action:     ctaOpenCategory,
				Title:      "Смотреть новые объявления",
				CategoryID: &favorite.CategoryID,
			},
		},
		Category:        categoryRef{ID: favorite.CategoryID, Title: favorite.Title},
		Subcategory:     subcategoryRefOf(favorite),
		Share:           share,
		Recommendations: listingRefsOf(input.categoryRecommendations),
	}, true
}

func buildInterestsSlide(input slideInput) (any, bool) {
	if len(input.seasons) == 0 {
		return nil, false
	}

	periods := make([]periodInterest, 0, len(input.seasons))

	for _, leader := range input.seasons {
		periods = append(periods, periodInterest{
			Period:      string(leader.season),
			Category:    categoryRef{ID: leader.category.CategoryID, Title: leader.category.Title},
			Subcategory: subcategoryRefOf(leader.category),
			Weight:      leader.share,
		})
	}

	summary := interestsSummary(input.seasons)

	return interestsSlide{
		slideBase: slideBase{
			Type:     slideInterests,
			Title:    titleInterests,
			Subtitle: summary,
		},
		Periods:      periods,
		ShiftSummary: summary,
	}, true
}

func buildArchetypeSlide(input slideInput) (any, bool) {
	archetype := input.archetype
	reasons := make([]archetypeReasonRef, 0, len(archetype.Reasons))

	for _, reason := range archetype.Reasons {
		reasons = append(reasons, archetypeReasonRef{
			Metric:      string(reason.Metric),
			Value:       reason.Value,
			Explanation: reason.Explanation,
		})
	}

	return archetypeSlide{
		slideBase: slideBase{
			Type:     slideArchetype,
			Title:    titleArchetype,
			Subtitle: archetype.Title,
		},
		Archetype: archetypeBody{
			Code:        string(archetype.UserArchetype),
			Title:       archetype.Title,
			Description: archetype.Description,
			Reasons:     reasons,
		},
	}, true
}

func buildFinalSlide(input slideInput) (any, bool) {
	actions := []cta{{Action: ctaShareRecap, Title: "Поделиться итогами"}}

	if len(input.categories) > 0 {
		favorite := input.categories[0]
		actions = append(actions, cta{
			Action:     ctaOpenCategory,
			Title:      "Вернуться в " + favorite.Title,
			CategoryID: &favorite.CategoryID,
		})
	}

	if input.activity.FavoritesActive > 0 {
		actions = append(actions, cta{Action: ctaOpenFavorites, Title: "Открыть избранное"})
	}

	return finalSlide{
		slideBase: slideBase{
			Type:     slideFinal,
			Title:    titleFinal,
			Subtitle: "Поделитесь итогами и вернитесь к тому, что отложили",
		},
		Stats:   finalStats(input),
		Actions: actions,
	}, true
}

func finalStats(input slideInput) []statTile {
	activity := input.activity

	candidates := []statTile{
		{Code: statActiveDays, Value: activity.ActiveDays, Label: activeDaysStatLabel(activity.ActiveDays)},
		{Code: statViews, Value: activity.Views, Label: viewsStatLabel(activity.Views)},
		{Code: statFavorites, Value: activity.Favorites, Label: favoritesStatLabel(activity.Favorites)},
		{Code: statMessages, Value: activity.Messages(), Label: messagesStatLabel(activity.Messages())},
		{Code: statSeasons, Value: int64(len(input.seasons)), Label: seasonsStatLabel(int64(len(input.seasons)))},
	}

	stats := make([]statTile, 0, len(candidates))

	for _, tile := range candidates {
		if tile.Value > 0 {
			stats = append(stats, tile)
		}
	}

	return stats
}

func subcategoryRefOf(category entity.CategoryScore) *categoryRef {
	if category.Subcategory == nil {
		return nil
	}

	return &categoryRef{ID: category.Subcategory.ID, Title: category.Subcategory.Title}
}

func favoriteListingRefOf(listing *entity.FavoriteListingPreview) *listingRef {
	if listing == nil {
		return nil
	}

	return &listingRef{
		ID:         listing.ID,
		Title:      listing.Title,
		Price:      listing.Price,
		CategoryID: listing.CategoryID,
		AddedAt:    &listing.AddedAt,
	}
}

func listingRefsOf(listings []entity.ListingPreview) []listingRef {
	result := make([]listingRef, 0, len(listings))
	for _, listing := range listings {
		result = append(result, listingRef{
			ID:         listing.ID,
			Title:      listing.Title,
			Price:      listing.Price,
			CategoryID: listing.CategoryID,
		})
	}

	return result
}

func interestsSummary(seasons []seasonLeader) string {
	if len(seasons) == 0 {
		return ""
	}

	sameCategory := true
	for _, leader := range seasons[1:] {
		if leader.category.CategoryID != seasons[0].category.CategoryID {
			sameCategory = false

			break
		}
	}

	if sameCategory {
		return "Весь год вас держала одна тема: " + seasons[0].category.Title
	}

	return "Ваши интересы менялись вместе с сезонами"
}

func salesAmountRange(amountKopecks int64) *amountRange {
	if amountKopecks <= 0 {
		return nil
	}

	bounds := []int64{0, 5_000, 15_000, 50_000, 100_000, 300_000, 1_000_000}
	roubles := amountKopecks / 100

	for index := len(bounds) - 1; index >= 0; index-- {
		if roubles < bounds[index] {
			continue
		}

		if index == len(bounds)-1 {
			return &amountRange{
				Min:      bounds[index],
				Currency: currencyRUB,
				Label:    "более " + formatRoubles(bounds[index]),
			}
		}

		upper := bounds[index+1]

		return &amountRange{
			Min:      bounds[index],
			Max:      &upper,
			Currency: currencyRUB,
			Label:    formatRoubles(bounds[index]) + "–" + formatRoubles(upper),
		}
	}

	return nil
}

func formatRoubles(value int64) string {
	digits := fmt.Sprintf("%d", value)

	var grouped strings.Builder
	for index, digit := range digits {
		if index > 0 && (len(digits)-index)%3 == 0 {
			grouped.WriteRune(' ')
		}

		grouped.WriteRune(digit)
	}

	return grouped.String() + " ₽"
}

func toInt32(value int64) int32 {
	if value < 0 {
		return 0
	}

	if value > math.MaxInt32 {
		return math.MaxInt32
	}

	return int32(value)
}
