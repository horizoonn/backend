package recap_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	recapcontroller "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/controller/http/recap"
	recapmocks "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/controller/http/recap/mocks"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const testYear = 2026

type errorResponse struct {
	Code      recapapi.ErrorCode `json:"code"`
	Message   string             `json:"message"`
	RequestID string             `json:"requestId"`
}

func TestCreateRecapHTTP(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	recapID := uuid.New()
	storageErr := errors.New("storage unavailable")

	tests := []struct {
		name        string
		creation    entity.RecapCreation
		serviceErr  error
		wantStatus  int
		wantCode    recapapi.ErrorCode
		wantMessage string
	}{
		{
			name:       "creates recap",
			creation:   entity.RecapCreation{ID: recapID, Created: true},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns existing recap",
			creation:   entity.RecapCreation{ID: recapID},
			wantStatus: http.StatusOK,
		},
		{
			name:        "profile is not found",
			serviceErr:  entity.ErrProfileNotFound,
			wantStatus:  http.StatusNotFound,
			wantCode:    recapapi.ErrorCodeProfileNotFound,
			wantMessage: entity.ErrProfileNotFound.Error(),
		},
		{
			name:        "activity is insufficient",
			serviceErr:  entity.ErrNotEnoughActivity,
			wantStatus:  http.StatusConflict,
			wantCode:    recapapi.ErrorCodeNotEnoughActivity,
			wantMessage: entity.ErrNotEnoughActivity.Error(),
		},
		{
			name:        "service fails",
			serviceErr:  storageErr,
			wantStatus:  http.StatusInternalServerError,
			wantCode:    recapapi.ErrorCodeInternalError,
			wantMessage: "internal error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, recapService, _, _ := newHTTPServer(t)
			recapService.EXPECT().
				Create(mock.Anything, profileID, testYear).
				Return(test.creation, test.serviceErr).
				Once()

			request := newJSONRequest(
				t,
				http.MethodPost,
				"/recaps",
				`{"profileId":"`+profileID.String()+`","year":2026}`,
			)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			require.Equal(t, test.wantStatus, response.Code)
			if test.wantCode == "" {
				var body recapapi.CreateRecapResponse
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				assert.Equal(t, recapapi.UUID(recapID), body.ID)

				return
			}

			assertErrorResponse(t, response, test.wantCode, test.wantMessage)
		})
	}
}

func TestListProfilesHTTP(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	avatarURL := "https://cdn.example.com/avatar.png"
	registeredAt := time.Date(2020, time.March, 14, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		profiles   []entity.Profile
		serviceErr error
		wantStatus int
		wantCode   recapapi.ErrorCode
	}{
		{
			name: "returns profiles",
			profiles: []entity.Profile{{
				ID:           profileID,
				Name:         "Анна",
				Surname:      "Воронова",
				AvatarURL:    &avatarURL,
				Hint:         "Сохраняет находки",
				RegisteredAt: registeredAt,
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "service fails",
			serviceErr: errors.New("profiles unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   recapapi.ErrorCodeInternalError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, _, profileService, _ := newHTTPServer(t)
			profileService.EXPECT().
				List(mock.Anything).
				Return(test.profiles, test.serviceErr).
				Once()

			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/profiles", nil)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			require.Equal(t, test.wantStatus, response.Code)
			if test.wantCode != "" {
				assertErrorResponse(t, response, test.wantCode, "internal error")

				return
			}

			var body struct {
				Items []recapapi.Profile `json:"items"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			require.Len(t, body.Items, 1)
			assert.Equal(t, recapapi.UUID(profileID), body.Items[0].ID)
			assert.Equal(t, "Анна", body.Items[0].Name)
			assert.Equal(t, "Сохраняет находки", body.Items[0].Hint.Value)
			assert.Equal(t, avatarURL, body.Items[0].AvatarUrl.Value.String())
			assert.Equal(t, registeredAt, body.Items[0].RegisteredAt.Value)
		})
	}
}

func TestGetRecapHTTP(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	generatedAt := time.Date(testYear, time.August, 13, 12, 0, 0, 0, time.UTC)
	stored := entity.Recap{
		ID:     recapID,
		UserID: uuid.New(),
		Year:   testYear,
		Archetype: entity.Archetype{
			UserArchetype: entity.ArchetypeExplorer,
			Title:         "Исследователь",
			Description:   "Изучает разные категории",
		},
		Slides:      json.RawMessage(`[{"type":"intro","title":"Ваш год","year":2026}]`),
		GeneratedAt: generatedAt,
	}

	tests := []struct {
		name       string
		recap      entity.Recap
		serviceErr error
		wantStatus int
		wantCode   recapapi.ErrorCode
	}{
		{
			name:       "returns recap",
			recap:      stored,
			wantStatus: http.StatusOK,
		},
		{
			name:       "recap is not found",
			serviceErr: entity.ErrRecapNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   recapapi.ErrorCodeRecapNotFound,
		},
		{
			name: "stored slides are invalid",
			recap: entity.Recap{
				ID:        recapID,
				UserID:    uuid.New(),
				Year:      testYear,
				Archetype: stored.Archetype,
				Slides:    json.RawMessage(`{`),
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   recapapi.ErrorCodeInternalError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, recapService, _, _ := newHTTPServer(t)
			recapService.EXPECT().
				Get(mock.Anything, recapID).
				Return(test.recap, test.serviceErr).
				Once()

			request := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				"/recaps/"+recapID.String(),
				nil,
			)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			require.Equal(t, test.wantStatus, response.Code)
			if test.wantCode != "" {
				message := "internal error"
				if test.wantCode == recapapi.ErrorCodeRecapNotFound {
					message = entity.ErrRecapNotFound.Error()
				}
				assertErrorResponse(t, response, test.wantCode, message)

				return
			}

			var body recapapi.Recap
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			assert.Equal(t, recapapi.UUID(recapID), body.ID)
			assert.Equal(t, recapapi.RecapStatusReady, body.Status)
			assert.Equal(t, recapapi.ArchetypeCodeExplorer, body.Archetype.Code)
			require.Len(t, body.Slides, 1)
			assert.Equal(t, generatedAt, body.GeneratedAt)
		})
	}
}

func newHTTPServer(
	t *testing.T,
) (
	*recapapi.Server,
	*recapmocks.MockRecapService,
	*recapmocks.MockProfileService,
	*recapmocks.MockSharedRecapService,
) {
	t.Helper()

	recapService := recapmocks.NewMockRecapService(t)
	profileService := recapmocks.NewMockProfileService(t)
	sharedService := recapmocks.NewMockSharedRecapService(t)
	publicBaseURL := url.URL{Scheme: "https", Host: "recap50.ru"}
	handler := recapcontroller.NewRecapServer(
		zap.NewNop(),
		recapService,
		profileService,
		sharedService,
		publicBaseURL,
	)

	server, err := recapapi.NewServer(handler)
	require.NoError(t, err)

	return server, recapService, profileService, sharedService
}

func newJSONRequest(t *testing.T, method, target, body string) *http.Request {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	return request
}

func assertErrorResponse(
	t *testing.T,
	response *httptest.ResponseRecorder,
	wantCode recapapi.ErrorCode,
	wantMessage string,
) {
	t.Helper()

	var body errorResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, wantCode, body.Code)
	assert.Equal(t, wantMessage, body.Message)
	assert.NotEmpty(t, body.RequestID)
	_, err := uuid.Parse(body.RequestID)
	assert.NoError(t, err)
}
