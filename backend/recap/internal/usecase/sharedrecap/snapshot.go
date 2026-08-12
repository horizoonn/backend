package sharedrecap

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const (
	minSharedRecapYear  = 2015
	maxSharedRecapYear  = 2026
	maxActiveDays       = 366
	maxDisplayNameRunes = 64
	maxCategoryRunes    = 128
	maxInterestRunes    = 512
	maxSharedBadges     = 3
)

func buildSharedRecap(
	recap entity.Recap,
	profile entity.Profile,
) (entity.SharedRecap, error) {
	facts, err := extractRecapFacts(recap.Slides)
	if err != nil {
		return entity.SharedRecap{}, fmt.Errorf("extract recap facts: %w", err)
	}

	archetype, err := publicArchetype(recap.Archetype.UserArchetype)
	if err != nil {
		return entity.SharedRecap{}, fmt.Errorf("build public archetype: %w", err)
	}

	sharedRecap := entity.SharedRecap{
		RecapID:         recap.ID,
		Year:            recap.Year,
		DisplayName:     profile.Name,
		Archetype:       archetype,
		ActiveDays:      *facts.activeDays,
		Views:           facts.views,
		TopCategory:     facts.topCategory,
		InterestSummary: facts.interestSummary,
		Badges:          facts.badges,
	}
	if err := validateSnapshot(sharedRecap); err != nil {
		return entity.SharedRecap{}, err
	}

	return sharedRecap, nil
}

func validateSnapshot(sharedRecap entity.SharedRecap) error {
	if err := validateSnapshotMetrics(sharedRecap); err != nil {
		return err
	}
	if err := validateCategory(sharedRecap.TopCategory); err != nil {
		return err
	}
	if err := validateOptionalText(
		"interest summary",
		sharedRecap.InterestSummary,
		maxInterestRunes,
	); err != nil {
		return err
	}
	if len(sharedRecap.Badges) > maxSharedBadges {
		return fmt.Errorf(
			"badges count %d exceeds maximum %d",
			len(sharedRecap.Badges),
			maxSharedBadges,
		)
	}

	return nil
}

func validateSnapshotMetrics(sharedRecap entity.SharedRecap) error {
	if sharedRecap.Year < minSharedRecapYear || sharedRecap.Year > maxSharedRecapYear {
		return fmt.Errorf("year %d is out of range", sharedRecap.Year)
	}
	if err := validateRequiredText(
		"display name",
		sharedRecap.DisplayName,
		maxDisplayNameRunes,
	); err != nil {
		return err
	}
	if sharedRecap.ActiveDays < 0 || sharedRecap.ActiveDays > maxActiveDays {
		return fmt.Errorf("active days %d is out of range", sharedRecap.ActiveDays)
	}
	if sharedRecap.Views != nil && *sharedRecap.Views < 0 {
		return fmt.Errorf("views must not be negative")
	}

	return nil
}

func validateCategory(category *entity.SharedCategory) error {
	if category == nil {
		return nil
	}
	if err := validateRequiredText(
		"category title",
		category.CategoryTitle,
		maxCategoryRunes,
	); err != nil {
		return err
	}

	return validateOptionalText(
		"subcategory title",
		category.SubcategoryTitle,
		maxCategoryRunes,
	)
}

func validateRequiredText(field, value string, maxRunes int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is empty", field)
	}
	if utf8.RuneCountInString(value) > maxRunes {
		return fmt.Errorf("%s exceeds maximum length %d", field, maxRunes)
	}

	return nil
}

func validateOptionalText(field, value string, maxRunes int) error {
	if value == "" {
		return nil
	}

	return validateRequiredText(field, value, maxRunes)
}
