package recap_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestShareRecapHTTP(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	createdAt := time.Date(testYear, time.August, 13, 12, 0, 0, 0, time.UTC)
	token := entity.SharedRecapToken("abcdefghijklmnopqrstuv")

	tests := []struct {
		name       string
		creation   entity.SharedRecapCreation
		serviceErr error
		wantStatus int
		wantCode   recapapi.ErrorCode
	}{
		{
			name: "creates public recap",
			creation: entity.SharedRecapCreation{
				Token:     token,
				CreatedAt: createdAt,
				Created:   true,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "returns existing public recap",
			creation: entity.SharedRecapCreation{
				Token:     token,
				CreatedAt: createdAt,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "recap is not found",
			serviceErr: entity.ErrRecapNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   recapapi.ErrorCodeRecapNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, _, _, sharedService := newHTTPServer(t)
			sharedService.EXPECT().
				Share(mock.Anything, recapID).
				Return(test.creation, test.serviceErr).
				Once()

			request := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				"/recaps/"+recapID.String()+"/share",
				nil,
			)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			require.Equal(t, test.wantStatus, response.Code)
			if test.wantCode != "" {
				assertErrorResponse(t, response, test.wantCode, entity.ErrRecapNotFound.Error())

				return
			}

			var body recapapi.SharedRecapLink
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			assert.Equal(t, recapapi.SharedRecapToken(token), body.Token)
			assert.Equal(t, "https://recap50.ru/shared-recaps/"+string(token), body.URL.String())
			assert.Equal(t, createdAt, body.CreatedAt)
		})
	}
}

func TestGetSharedRecapHTTP(t *testing.T) {
	t.Parallel()

	token := entity.SharedRecapToken("abcdefghijklmnopqrstuv")
	stored := entity.SharedRecap{
		Token:       token,
		Year:        testYear,
		DisplayName: "Анна",
		Archetype: entity.SharedArchetype{
			Name:        entity.ArchetypeCollector,
			Title:       "Коллекционер",
			Description: "Сохраняет находки",
		},
		ActiveDays: 120,
		Badges:     []entity.SharedBadge{},
		CreatedAt:  time.Date(testYear, time.August, 13, 12, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name       string
		shared     entity.SharedRecap
		serviceErr error
		wantStatus int
		wantCode   recapapi.ErrorCode
	}{
		{
			name:       "returns public recap",
			shared:     stored,
			wantStatus: http.StatusOK,
		},
		{
			name:       "public recap is not found",
			serviceErr: entity.ErrSharedRecapNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   recapapi.ErrorCodeSharedRecapNotFound,
		},
		{
			name:       "service fails",
			serviceErr: errors.New("shared recap unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   recapapi.ErrorCodeInternalError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, _, _, sharedService := newHTTPServer(t)
			sharedService.EXPECT().
				Get(mock.Anything, token).
				Return(test.shared, test.serviceErr).
				Once()

			request := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				"/shared-recaps/"+string(token),
				nil,
			)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			require.Equal(t, test.wantStatus, response.Code)
			if test.wantCode != "" {
				message := "internal error"
				if test.wantCode == recapapi.ErrorCodeSharedRecapNotFound {
					message = entity.ErrSharedRecapNotFound.Error()
				}
				assertErrorResponse(t, response, test.wantCode, message)

				return
			}

			var body recapapi.SharedRecap
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			assert.Equal(t, "Анна", body.DisplayName)
			assert.Equal(t, recapapi.SharedArchetypeCodeCollector, body.Archetype.Code)
			assert.Equal(t, int32(120), body.ActiveDays)
		})
	}
}
