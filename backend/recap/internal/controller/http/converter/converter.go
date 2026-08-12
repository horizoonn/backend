package converter

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func ConvertEntityProfileToAPIProfile(profile entity.Profile) recapapi.Profile {
	result := recapapi.Profile{
		ID:      recapapi.UUID(profile.ID),
		Name:    profile.Name,
		Surname: profile.Surname,
	}

	if profile.Hint != "" {
		result.Hint = recapapi.NewOptString(profile.Hint)
	}

	return result
}

func ConvertEntityRecapToAPIRecap(recap entity.Recap) (recapapi.Recap, error) {
	slides, err := ConvertRawMessageToAPISlides(recap.Slides)
	if err != nil {
		return recapapi.Recap{}, err
	}

	return recapapi.Recap{
		ID:          recapapi.UUID(recap.ID),
		ProfileId:   recapapi.UUID(recap.UserID),
		Year:        ConvertIntToAPIYear(recap.Year),
		Status:      recapapi.RecapStatusReady,
		Archetype:   ConvertEntityArchetypeToAPIArchetype(recap.Archetype),
		Slides:      slides,
		GeneratedAt: recap.GeneratedAt,
	}, nil
}

func ConvertRawMessageToAPISlides(raw json.RawMessage) ([]recapapi.Slide, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var slides []recapapi.Slide
	if err := json.Unmarshal(raw, &slides); err != nil {
		return nil, fmt.Errorf("decode slides: %w", err)
	}

	return slides, nil
}

func ConvertEntityArchetypeToAPIArchetype(archetype entity.Archetype) recapapi.Archetype {
	reasons := make([]recapapi.ArchetypeReason, 0, len(archetype.Reasons))

	for _, reason := range archetype.Reasons {
		reasons = append(reasons, recapapi.ArchetypeReason{
			Metric:      recapapi.MetricCode(reason.Metric),
			Value:       reason.Value,
			Explanation: reason.Explanation,
		})
	}

	return recapapi.Archetype{
		Code:        recapapi.ArchetypeCode(archetype.UserArchetype),
		Title:       archetype.Title,
		Description: archetype.Description,
		Reasons:     reasons,
	}
}

func ConvertIntToAPIYear(year int32) recapapi.Year {
	return recapapi.Year(year)
}

func ConvertEntitySharedRecapToAPISharedRecap(
	sharedRecap entity.SharedRecap,
) (recapapi.SharedRecap, error) {
	badges, err := convertEntitySharedBadgesToAPI(sharedRecap.Badges)
	if err != nil {
		return recapapi.SharedRecap{}, err
	}

	result := recapapi.SharedRecap{
		Year:        ConvertIntToAPIYear(sharedRecap.Year),
		DisplayName: sharedRecap.DisplayName,
		Archetype: recapapi.SharedArchetype{
			Code:        recapapi.SharedArchetypeCode(sharedRecap.Archetype.Name),
			Title:       sharedRecap.Archetype.Title,
			Description: sharedRecap.Archetype.Description,
		},
		ActiveDays: sharedRecap.ActiveDays,
		Badges:     badges,
	}

	if sharedRecap.Views != nil {
		result.Views = recapapi.NewOptInt64(*sharedRecap.Views)
	}
	if sharedRecap.TopCategory != nil {
		category := recapapi.SharedCategory{
			CategoryTitle: sharedRecap.TopCategory.CategoryTitle,
		}
		if sharedRecap.TopCategory.SubcategoryTitle != "" {
			category.SubcategoryTitle = recapapi.NewOptString(
				sharedRecap.TopCategory.SubcategoryTitle,
			)
		}
		result.TopCategory = recapapi.NewOptSharedCategory(category)
	}
	if sharedRecap.InterestSummary != "" {
		result.InterestSummary = recapapi.NewOptString(sharedRecap.InterestSummary)
	}

	return result, nil
}

func convertEntitySharedBadgesToAPI(
	badges []entity.SharedBadge,
) ([]recapapi.SharedBadge, error) {
	result := make([]recapapi.SharedBadge, 0, len(badges))
	for i, badge := range badges {
		converted := recapapi.SharedBadge{
			Code:        badge.Code,
			Title:       badge.Title,
			Description: badge.Description,
			Level:       recapapi.SharedBadgeLevel(badge.Level),
		}
		if badge.IconURL != nil {
			iconURL, err := url.Parse(*badge.IconURL)
			if err != nil {
				return nil, fmt.Errorf("parse badge %d icon url: %w", i, err)
			}
			converted.IconUrl = recapapi.NewOptURI(*iconURL)
		}

		result = append(result, converted)
	}

	return result, nil
}
