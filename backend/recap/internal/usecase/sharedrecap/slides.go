package sharedrecap

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const (
	slideActiveDays       = "active_days"
	slideViews            = "views"
	slideFavoriteCategory = "favorite_category"
	slidePurchases        = "purchases"
	slideSales            = "sales"
	slideInterests        = "interests"
)

type (
	storedSlideEnvelope struct {
		Type string `json:"type"`
	}

	storedActiveDaysSlide struct {
		ActiveDays *int32 `json:"activeDays"`
	}

	storedViewsSlide struct {
		Views *int64 `json:"views"`
	}

	storedCategoryRef struct {
		ID    uuid.UUID `json:"id"`
		Title string    `json:"title"`
	}

	storedCategorySlide struct {
		Category    storedCategoryRef  `json:"category"`
		Subcategory *storedCategoryRef `json:"subcategory,omitempty"`
	}

	storedBadgeRef struct {
		Code string `json:"code"`
	}

	storedBadgeSlide struct {
		Badge *storedBadgeRef `json:"badge,omitempty"`
	}

	storedInterestPeriod struct {
		Period   entity.Season     `json:"period"`
		Category storedCategoryRef `json:"category"`
	}

	storedInterestsSlide struct {
		Periods []storedInterestPeriod `json:"periods"`
	}

	recapFacts struct {
		activeDays      *int32
		views           *int64
		topCategory     *entity.SharedCategory
		interestSummary string
		badges          []entity.SharedBadge
		badgeCodes      map[string]struct{}
	}
)

func extractRecapFacts(slides json.RawMessage) (recapFacts, error) {
	var storedSlides []json.RawMessage
	if err := json.Unmarshal(slides, &storedSlides); err != nil {
		return recapFacts{}, fmt.Errorf("decode recap slides: %w", err)
	}

	facts := recapFacts{
		badges:     make([]entity.SharedBadge, 0),
		badgeCodes: make(map[string]struct{}),
	}
	for index, rawSlide := range storedSlides {
		if err := applyStoredSlide(&facts, rawSlide); err != nil {
			return recapFacts{}, fmt.Errorf("decode recap slide %d: %w", index, err)
		}
	}

	if facts.activeDays == nil {
		return recapFacts{}, fmt.Errorf("active days slide is missing")
	}

	return facts, nil
}

func applyStoredSlide(facts *recapFacts, rawSlide json.RawMessage) error {
	var envelope storedSlideEnvelope
	if err := json.Unmarshal(rawSlide, &envelope); err != nil {
		return fmt.Errorf("decode slide type: %w", err)
	}

	switch envelope.Type {
	case slideActiveDays:
		return applyActiveDaysSlide(facts, rawSlide)
	case slideViews:
		return applyViewsSlide(facts, rawSlide)
	case slideFavoriteCategory:
		return applyCategorySlide(facts, rawSlide)
	case slidePurchases, slideSales:
		return applyBadgeSlide(facts, rawSlide)
	case slideInterests:
		return applyInterestsSlide(facts, rawSlide)
	default:
		return nil
	}
}

func applyActiveDaysSlide(facts *recapFacts, rawSlide json.RawMessage) error {
	var slide storedActiveDaysSlide
	if err := json.Unmarshal(rawSlide, &slide); err != nil {
		return fmt.Errorf("decode active days slide: %w", err)
	}
	if slide.ActiveDays == nil {
		return fmt.Errorf("activeDays is missing")
	}

	facts.activeDays = slide.ActiveDays

	return nil
}

func applyViewsSlide(facts *recapFacts, rawSlide json.RawMessage) error {
	var slide storedViewsSlide
	if err := json.Unmarshal(rawSlide, &slide); err != nil {
		return fmt.Errorf("decode views slide: %w", err)
	}
	if slide.Views == nil {
		return fmt.Errorf("views is missing")
	}

	facts.views = slide.Views

	return nil
}

func applyCategorySlide(facts *recapFacts, rawSlide json.RawMessage) error {
	var slide storedCategorySlide
	if err := json.Unmarshal(rawSlide, &slide); err != nil {
		return fmt.Errorf("decode favorite category slide: %w", err)
	}

	category := &entity.SharedCategory{CategoryTitle: slide.Category.Title}
	if slide.Subcategory != nil {
		category.SubcategoryTitle = slide.Subcategory.Title
	}
	facts.topCategory = category

	return nil
}

func applyBadgeSlide(facts *recapFacts, rawSlide json.RawMessage) error {
	var slide storedBadgeSlide
	if err := json.Unmarshal(rawSlide, &slide); err != nil {
		return fmt.Errorf("decode badge slide: %w", err)
	}
	if slide.Badge == nil {
		return nil
	}
	if slide.Badge.Code == "" {
		return fmt.Errorf("badge code is missing")
	}
	if _, exists := facts.badgeCodes[slide.Badge.Code]; exists {
		return nil
	}

	badge, known := publicBadge(slide.Badge.Code)
	if !known {
		return nil
	}

	facts.badges = append(facts.badges, badge)
	facts.badgeCodes[badge.Code] = struct{}{}

	return nil
}

func applyInterestsSlide(facts *recapFacts, rawSlide json.RawMessage) error {
	var slide storedInterestsSlide
	if err := json.Unmarshal(rawSlide, &slide); err != nil {
		return fmt.Errorf("decode interests slide: %w", err)
	}

	facts.interestSummary = truncateText(
		buildInterestSummary(slide.Periods),
		maxInterestRunes,
	)

	return nil
}

func buildInterestSummary(periods []storedInterestPeriod) string {
	validPeriods := make([]storedInterestPeriod, 0, len(periods))
	for _, period := range periods {
		if strings.TrimSpace(period.Category.Title) == "" {
			continue
		}
		if _, known := publicSeasonLabel(period.Period); !known {
			continue
		}

		validPeriods = append(validPeriods, period)
	}
	if len(validPeriods) == 0 {
		return ""
	}

	if sameCategory(validPeriods) {
		return "Главный интерес года — " + validPeriods[0].Category.Title + "."
	}

	parts := make([]string, 0, len(validPeriods))
	for index, period := range validPeriods {
		label, _ := publicSeasonLabel(period.Period)
		if index == 0 {
			label = capitalize(label)
		}
		parts = append(parts, label+" — "+period.Category.Title)
	}

	return strings.Join(parts, ", ") + "."
}

func capitalize(text string) string {
	if text == "" {
		return text
	}

	first, size := utf8.DecodeRuneInString(text)

	return string(unicode.ToUpper(first)) + text[size:]
}

func truncateText(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}

	return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
}

func sameCategory(periods []storedInterestPeriod) bool {
	for _, period := range periods[1:] {
		if period.Category.ID != periods[0].Category.ID {
			return false
		}
	}

	return true
}

func publicSeasonLabel(season entity.Season) (string, bool) {
	switch season {
	case entity.SeasonWinter:
		return "зимой", true
	case entity.SeasonSpring:
		return "весной", true
	case entity.SeasonSummer:
		return "летом", true
	case entity.SeasonAutumn:
		return "осенью", true
	default:
		return "", false
	}
}
