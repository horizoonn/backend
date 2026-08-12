package recap

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func recapToModel(recap entity.Recap) (recapModel, error) {
	if !recap.Archetype.UserArchetype.Valid() {
		return recapModel{}, fmt.Errorf(
			"invalid archetype %q",
			recap.Archetype.UserArchetype,
		)
	}
	if strings.TrimSpace(recap.Archetype.Title) == "" {
		return recapModel{}, fmt.Errorf("archetype title is empty")
	}
	if strings.TrimSpace(recap.Archetype.Description) == "" {
		return recapModel{}, fmt.Errorf("archetype description is empty")
	}
	reasons, err := reasonsToModel(recap.Archetype.Reasons)
	if err != nil {
		return recapModel{}, fmt.Errorf("convert archetype reasons: %w", err)
	}
	if err := validateSlides(recap.Slides); err != nil {
		return recapModel{}, err
	}

	return recapModel{
		ID:                   recap.ID,
		UserID:               recap.UserID,
		Year:                 recap.Year,
		Archetype:            string(recap.Archetype.UserArchetype),
		ArchetypeTitle:       recap.Archetype.Title,
		ArchetypeDescription: recap.Archetype.Description,
		ArchetypeReasons:     reasons,
		Slides:               recap.Slides,
		GeneratedAt:          recap.GeneratedAt,
	}, nil
}

func recapModelToEntity(model recapModel) (entity.Recap, error) {
	archetypeName := entity.ArchetypeName(model.Archetype)
	if !archetypeName.Valid() {
		return entity.Recap{}, fmt.Errorf("invalid archetype %q", model.Archetype)
	}
	if strings.TrimSpace(model.ArchetypeTitle) == "" {
		return entity.Recap{}, fmt.Errorf("archetype title is empty")
	}
	if strings.TrimSpace(model.ArchetypeDescription) == "" {
		return entity.Recap{}, fmt.Errorf("archetype description is empty")
	}
	reasons, err := reasonsToEntity(model.ArchetypeReasons)
	if err != nil {
		return entity.Recap{}, fmt.Errorf("convert archetype reasons: %w", err)
	}
	if err := validateSlides(model.Slides); err != nil {
		return entity.Recap{}, err
	}

	return entity.Recap{
		ID:     model.ID,
		UserID: model.UserID,
		Year:   model.Year,
		Archetype: entity.Archetype{
			UserArchetype: archetypeName,
			Title:         model.ArchetypeTitle,
			Description:   model.ArchetypeDescription,
			Reasons:       reasons,
		},
		Slides:      model.Slides,
		GeneratedAt: model.GeneratedAt,
	}, nil
}

func reasonsToModel(reasons []entity.ArchetypeReason) ([]archetypeReasonModel, error) {
	if len(reasons) == 0 {
		return nil, fmt.Errorf("archetype reasons are empty")
	}

	result := make([]archetypeReasonModel, 0, len(reasons))
	for i, reason := range reasons {
		if !reason.Metric.Valid() {
			return nil, fmt.Errorf("reason %d has invalid metric %q", i, reason.Metric)
		}
		if strings.TrimSpace(reason.Value) == "" {
			return nil, fmt.Errorf("reason %d has empty value", i)
		}
		if strings.TrimSpace(reason.Explanation) == "" {
			return nil, fmt.Errorf("reason %d has empty explanation", i)
		}

		result = append(result, archetypeReasonModel{
			Metric:      string(reason.Metric),
			Value:       reason.Value,
			Explanation: reason.Explanation,
		})
	}

	return result, nil
}

func reasonsToEntity(reasons []archetypeReasonModel) ([]entity.ArchetypeReason, error) {
	if len(reasons) == 0 {
		return nil, fmt.Errorf("archetype reasons are empty")
	}

	result := make([]entity.ArchetypeReason, 0, len(reasons))
	for i, reason := range reasons {
		metric := entity.Metric(reason.Metric)
		if !metric.Valid() {
			return nil, fmt.Errorf("reason %d has invalid metric %q", i, reason.Metric)
		}
		if strings.TrimSpace(reason.Value) == "" {
			return nil, fmt.Errorf("reason %d has empty value", i)
		}
		if strings.TrimSpace(reason.Explanation) == "" {
			return nil, fmt.Errorf("reason %d has empty explanation", i)
		}

		result = append(result, entity.ArchetypeReason{
			Metric:      metric,
			Value:       reason.Value,
			Explanation: reason.Explanation,
		})
	}

	return result, nil
}

func validateSlides(slides json.RawMessage) error {
	var items []json.RawMessage
	if err := json.Unmarshal(slides, &items); err != nil {
		return fmt.Errorf("decode recap slides: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("recap slides are empty")
	}

	return nil
}
