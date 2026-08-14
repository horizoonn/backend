package recap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const (
	avitoScheme = "https"
	avitoHost   = "www.avito.ru"
)

var avitoCategoryPaths = map[string]string{
	"Электроника":         "/all/bytovaya_elektronika",
	"Для дома и дачи":     "/all/dlya_doma_i_dachi",
	"Транспорт":           "/all/transport",
	"Хобби и отдых":       "/all/hobbi_i_otdyh",
	"Одежда и аксессуары": "/all/lichnye_veschi",
	"Недвижимость":        "/all/nedvizhimost",
}

type navigationSlide struct {
	Type            string       `json:"type"`
	Category        *categoryRef `json:"category,omitempty"`
	Recommendations []listingRef `json:"recommendations,omitempty"`
}

func (s *recapService) OpenSimilarRecommendation(
	ctx context.Context,
	recapID uuid.UUID,
	listingID uuid.UUID,
) (url.URL, error) {
	recap, err := s.Get(ctx, recapID)
	if err != nil {
		return url.URL{}, fmt.Errorf("open similar recommendation: %w", err)
	}

	slides, err := navigationSlides(recap.Slides)
	if err != nil {
		return url.URL{}, fmt.Errorf("open similar recommendation: %w", err)
	}

	for _, slide := range slides {
		if slide.Type != slideFavoriteCategory {
			continue
		}

		for _, recommendation := range slide.Recommendations {
			if recommendation.ID == listingID && strings.TrimSpace(recommendation.Title) != "" {
				return avitoSearchURL(recommendation.Title), nil
			}
		}
	}

	return url.URL{}, fmt.Errorf(
		"open similar recommendation %s: %w",
		listingID,
		entity.ErrRecommendationNotFound,
	)
}

func (s *recapService) OpenAction(
	ctx context.Context,
	recapID uuid.UUID,
	action string,
) (url.URL, error) {
	recap, err := s.Get(ctx, recapID)
	if err != nil {
		return url.URL{}, fmt.Errorf("open recap action: %w", err)
	}

	switch action {
	case ctaOpenFavorites:
		return avitoURL("/favorites"), nil
	case ctaCreateListing:
		return avitoURL("/additem"), nil
	case ctaOpenCategory:
		return favoriteCategoryURL(recap.Slides)
	default:
		return url.URL{}, fmt.Errorf(
			"open recap action %q: %w",
			action,
			entity.ErrUnsupportedRecapAction,
		)
	}
}

func (s *recapService) OpenHome(context.Context) url.URL {
	return avitoURL("/")
}

func favoriteCategoryURL(rawSlides json.RawMessage) (url.URL, error) {
	slides, err := navigationSlides(rawSlides)
	if err != nil {
		return url.URL{}, fmt.Errorf("resolve favorite category: %w", err)
	}

	for _, slide := range slides {
		if slide.Type == slideFavoriteCategory &&
			slide.Category != nil &&
			strings.TrimSpace(slide.Category.Title) != "" {
			return avitoCategoryURL(slide.Category.Title), nil
		}
	}

	return url.URL{}, fmt.Errorf(
		"resolve favorite category: %w",
		entity.ErrFavoriteCategoryNotFound,
	)
}

func avitoCategoryURL(title string) url.URL {
	title = strings.TrimSpace(title)
	if path, ok := avitoCategoryPaths[title]; ok {
		return avitoURL(path)
	}

	return avitoSearchURL(title)
}

func navigationSlides(raw json.RawMessage) ([]navigationSlide, error) {
	var slides []navigationSlide
	if err := json.Unmarshal(raw, &slides); err != nil {
		return nil, fmt.Errorf("decode recap snapshot: %w", err)
	}

	return slides, nil
}

func avitoSearchURL(query string) url.URL {
	destination := avitoURL("/all")
	values := destination.Query()
	values.Set("q", query)
	destination.RawQuery = values.Encode()

	return destination
}

func avitoURL(path string) url.URL {
	return url.URL{Scheme: avitoScheme, Host: avitoHost, Path: path}
}
