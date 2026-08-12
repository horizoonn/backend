package sharedrecap

import (
	"fmt"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func sharedRecapToModel(sharedRecap entity.SharedRecap) (sharedRecapModel, error) {
	archetype, err := archetypeToModel(sharedRecap.Archetype)
	if err != nil {
		return sharedRecapModel{}, fmt.Errorf("convert archetype: %w", err)
	}
	badges, err := badgesToModel(sharedRecap.Badges)
	if err != nil {
		return sharedRecapModel{}, fmt.Errorf("convert badges: %w", err)
	}

	return sharedRecapModel{
		Token:   string(sharedRecap.Token),
		RecapID: sharedRecap.RecapID,
		Snapshot: sharedRecapSnapshotModel{
			Year:            sharedRecap.Year,
			DisplayName:     sharedRecap.DisplayName,
			Archetype:       archetype,
			ActiveDays:      sharedRecap.ActiveDays,
			Views:           sharedRecap.Views,
			TopCategory:     categoryToModel(sharedRecap.TopCategory),
			InterestSummary: sharedRecap.InterestSummary,
			Badges:          badges,
		},
		CreatedAt: sharedRecap.CreatedAt,
	}, nil
}

func sharedRecapModelToEntity(model sharedRecapModel) (entity.SharedRecap, error) {
	archetype, err := archetypeToEntity(model.Snapshot.Archetype)
	if err != nil {
		return entity.SharedRecap{}, fmt.Errorf("convert archetype: %w", err)
	}
	badges, err := badgesToEntity(model.Snapshot.Badges)
	if err != nil {
		return entity.SharedRecap{}, fmt.Errorf("convert badges: %w", err)
	}

	return entity.SharedRecap{
		Token:           entity.SharedRecapToken(model.Token),
		RecapID:         model.RecapID,
		Year:            model.Snapshot.Year,
		DisplayName:     model.Snapshot.DisplayName,
		Archetype:       archetype,
		ActiveDays:      model.Snapshot.ActiveDays,
		Views:           model.Snapshot.Views,
		TopCategory:     categoryToEntity(model.Snapshot.TopCategory),
		InterestSummary: model.Snapshot.InterestSummary,
		Badges:          badges,
		CreatedAt:       model.CreatedAt,
	}, nil
}

func archetypeToModel(archetype entity.SharedArchetype) (sharedArchetypeModel, error) {
	if !archetype.Name.Valid() {
		return sharedArchetypeModel{}, fmt.Errorf("invalid archetype %q", archetype.Name)
	}

	return sharedArchetypeModel{
		Code:        string(archetype.Name),
		Title:       archetype.Title,
		Description: archetype.Description,
	}, nil
}

func archetypeToEntity(archetype sharedArchetypeModel) (entity.SharedArchetype, error) {
	name := entity.ArchetypeName(archetype.Code)
	if !name.Valid() {
		return entity.SharedArchetype{}, fmt.Errorf("invalid archetype %q", archetype.Code)
	}

	return entity.SharedArchetype{
		Name:        name,
		Title:       archetype.Title,
		Description: archetype.Description,
	}, nil
}

func categoryToModel(category *entity.SharedCategory) *sharedCategoryModel {
	if category == nil {
		return nil
	}

	return &sharedCategoryModel{
		CategoryTitle:    category.CategoryTitle,
		SubcategoryTitle: category.SubcategoryTitle,
	}
}

func categoryToEntity(category *sharedCategoryModel) *entity.SharedCategory {
	if category == nil {
		return nil
	}

	return &entity.SharedCategory{
		CategoryTitle:    category.CategoryTitle,
		SubcategoryTitle: category.SubcategoryTitle,
	}
}

func badgesToModel(badges []entity.SharedBadge) ([]sharedBadgeModel, error) {
	result := make([]sharedBadgeModel, 0, len(badges))
	for i, badge := range badges {
		if !badge.Level.Valid() {
			return nil, fmt.Errorf("badge %d has invalid level %q", i, badge.Level)
		}

		result = append(result, sharedBadgeModel{
			Code:        badge.Code,
			Title:       badge.Title,
			Description: badge.Description,
			Level:       string(badge.Level),
			IconURL:     badge.IconURL,
		})
	}

	return result, nil
}

func badgesToEntity(badges []sharedBadgeModel) ([]entity.SharedBadge, error) {
	result := make([]entity.SharedBadge, 0, len(badges))
	for i, badge := range badges {
		level := entity.BadgeLevel(badge.Level)
		if !level.Valid() {
			return nil, fmt.Errorf("badge %d has invalid level %q", i, badge.Level)
		}

		result = append(result, entity.SharedBadge{
			Code:        badge.Code,
			Title:       badge.Title,
			Description: badge.Description,
			Level:       level,
			IconURL:     badge.IconURL,
		})
	}

	return result, nil
}
