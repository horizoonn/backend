package recap

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestRecapService_OpenAction(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	storageError := errors.New("recap lookup failed")

	tests := []struct {
		name          string
		action        string
		slides        json.RawMessage
		repositoryErr error
		wantURL       string
		wantErr       error
		wantErrText   string
	}{
		{
			name:    "electronics category",
			action:  ctaOpenCategory,
			slides:  categoryNavigationSlides(t, "Электроника"),
			wantURL: "https://www.avito.ru/all/bytovaya_elektronika",
		},
		{
			name:    "home category",
			action:  ctaOpenCategory,
			slides:  categoryNavigationSlides(t, "Для дома и дачи"),
			wantURL: "https://www.avito.ru/all/dlya_doma_i_dachi",
		},
		{
			name:    "transport category",
			action:  ctaOpenCategory,
			slides:  categoryNavigationSlides(t, "Транспорт"),
			wantURL: "https://www.avito.ru/all/transport",
		},
		{
			name:    "hobby category",
			action:  ctaOpenCategory,
			slides:  categoryNavigationSlides(t, "Хобби и отдых"),
			wantURL: "https://www.avito.ru/all/hobbi_i_otdyh",
		},
		{
			name:    "fashion category",
			action:  ctaOpenCategory,
			slides:  categoryNavigationSlides(t, "Одежда и аксессуары"),
			wantURL: "https://www.avito.ru/all/lichnye_veschi",
		},
		{
			name:    "real estate category",
			action:  ctaOpenCategory,
			slides:  categoryNavigationSlides(t, "Недвижимость"),
			wantURL: "https://www.avito.ru/all/nedvizhimost",
		},
		{
			name:    "unknown category uses search",
			action:  ctaOpenCategory,
			slides:  categoryNavigationSlides(t, "Работа и услуги"),
			wantURL: "https://www.avito.ru/all?q=%D0%A0%D0%B0%D0%B1%D0%BE%D1%82%D0%B0+%D0%B8+%D1%83%D1%81%D0%BB%D1%83%D0%B3%D0%B8",
		},
		{
			name:    "favorites",
			action:  ctaOpenFavorites,
			slides:  json.RawMessage(`[]`),
			wantURL: "https://www.avito.ru/favorites",
		},
		{
			name:    "create listing",
			action:  ctaCreateListing,
			slides:  json.RawMessage(`[]`),
			wantURL: "https://www.avito.ru/additem",
		},
		{
			name:    "unsupported action",
			action:  "redirect_anywhere",
			slides:  json.RawMessage(`[]`),
			wantErr: entity.ErrUnsupportedRecapAction,
		},
		{
			name:          "recap lookup fails",
			action:        ctaOpenFavorites,
			repositoryErr: storageError,
			wantErr:       storageError,
		},
		{
			name:        "invalid snapshot",
			action:      ctaOpenCategory,
			slides:      json.RawMessage(`{`),
			wantErrText: "decode recap snapshot",
		},
		{
			name:    "favorite category is missing",
			action:  ctaOpenCategory,
			slides:  json.RawMessage(`[{"type":"views"}]`),
			wantErr: entity.ErrFavoriteCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			service, dependencies := newRecapTestService(t)
			dependencies.recap.EXPECT().
				GetByID(ctx, recapID).
				Return(entity.Recap{ID: recapID, Slides: tt.slides}, tt.repositoryErr).
				Once()

			destination, err := service.OpenAction(ctx, recapID, tt.action)

			assertNavigationResult(t, destination.String(), err, tt.wantURL, tt.wantErr, tt.wantErrText)
		})
	}
}

func TestRecapService_OpenSimilarRecommendation(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	listingID := uuid.New()
	storageError := errors.New("recap lookup failed")

	tests := []struct {
		name          string
		listingID     uuid.UUID
		slides        json.RawMessage
		repositoryErr error
		wantURL       string
		wantErr       error
		wantErrText   string
	}{
		{
			name:      "recommendation title becomes a search query",
			listingID: listingID,
			slides:    recommendationNavigationSlides(t, listingID, "Журнальный столик"),
			wantURL:   "https://www.avito.ru/all?q=%D0%96%D1%83%D1%80%D0%BD%D0%B0%D0%BB%D1%8C%D0%BD%D1%8B%D0%B9+%D1%81%D1%82%D0%BE%D0%BB%D0%B8%D0%BA",
		},
		{
			name:      "special characters are encoded",
			listingID: listingID,
			slides:    recommendationNavigationSlides(t, listingID, "Стол & стулья"),
			wantURL:   "https://www.avito.ru/all?q=%D0%A1%D1%82%D0%BE%D0%BB+%26+%D1%81%D1%82%D1%83%D0%BB%D1%8C%D1%8F",
		},
		{
			name:      "unknown recommendation",
			listingID: uuid.New(),
			slides:    recommendationNavigationSlides(t, listingID, "Журнальный столик"),
			wantErr:   entity.ErrRecommendationNotFound,
		},
		{
			name:      "recommendation from another recap is unavailable",
			listingID: listingID,
			slides:    recommendationNavigationSlides(t, uuid.New(), "Журнальный столик"),
			wantErr:   entity.ErrRecommendationNotFound,
		},
		{
			name:      "blank recommendation title",
			listingID: listingID,
			slides:    recommendationNavigationSlides(t, listingID, "   "),
			wantErr:   entity.ErrRecommendationNotFound,
		},
		{
			name:          "recap lookup fails",
			listingID:     listingID,
			repositoryErr: storageError,
			wantErr:       storageError,
		},
		{
			name:        "invalid snapshot",
			listingID:   listingID,
			slides:      json.RawMessage(`{`),
			wantErrText: "decode recap snapshot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			service, dependencies := newRecapTestService(t)
			dependencies.recap.EXPECT().
				GetByID(ctx, recapID).
				Return(entity.Recap{ID: recapID, Slides: tt.slides}, tt.repositoryErr).
				Once()

			destination, err := service.OpenSimilarRecommendation(ctx, recapID, tt.listingID)

			assertNavigationResult(t, destination.String(), err, tt.wantURL, tt.wantErr, tt.wantErrText)
		})
	}
}

func TestRecapService_OpenHome(t *testing.T) {
	t.Parallel()

	service, _ := newRecapTestService(t)
	destination := service.OpenHome(t.Context())

	require.Equal(t, "https://www.avito.ru/", destination.String())
}

func categoryNavigationSlides(t *testing.T, title string) json.RawMessage {
	t.Helper()

	return marshalNavigationSlides(t, []navigationSlide{{
		Type:     slideFavoriteCategory,
		Category: &categoryRef{ID: uuid.New(), Title: title},
	}})
}

func recommendationNavigationSlides(
	t *testing.T,
	listingID uuid.UUID,
	title string,
) json.RawMessage {
	t.Helper()

	return marshalNavigationSlides(t, []navigationSlide{{
		Type: slideFavoriteCategory,
		Recommendations: []listingRef{{
			ID:    listingID,
			Title: title,
		}},
	}})
}

func marshalNavigationSlides(t *testing.T, slides []navigationSlide) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(slides)
	require.NoError(t, err)

	return raw
}

func assertNavigationResult(
	t *testing.T,
	gotURL string,
	err error,
	wantURL string,
	wantErr error,
	wantErrText string,
) {
	t.Helper()

	switch {
	case wantErr != nil:
		require.ErrorIs(t, err, wantErr)
	case wantErrText != "":
		require.ErrorContains(t, err, wantErrText)
	default:
		require.NoError(t, err)
	}
	require.Equal(t, wantURL, gotURL)
}
