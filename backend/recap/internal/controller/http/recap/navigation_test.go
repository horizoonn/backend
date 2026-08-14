package recap_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestOpenRecapActionHTTP(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	destination := url.URL{Scheme: "https", Host: "www.avito.ru", Path: "/favorites"}
	tests := []struct {
		name        string
		serviceErr  error
		wantStatus  int
		wantCode    recapapi.ErrorCode
		wantMessage string
	}{
		{
			name:       "redirects to service destination",
			wantStatus: http.StatusFound,
		},
		{
			name:        "recap is not found",
			serviceErr:  entity.ErrRecapNotFound,
			wantStatus:  http.StatusNotFound,
			wantCode:    recapapi.ErrorCodeRecapNotFound,
			wantMessage: "recap action not found",
		},
		{
			name:        "action is unsupported",
			serviceErr:  entity.ErrUnsupportedRecapAction,
			wantStatus:  http.StatusBadRequest,
			wantCode:    recapapi.ErrorCodeBadRequest,
			wantMessage: entity.ErrUnsupportedRecapAction.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, recapService, _, _ := newHTTPServer(t)
			recapService.EXPECT().
				OpenAction(mock.Anything, recapID, "open_favorites").
				Return(destination, test.serviceErr).
				Once()

			request := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				"/recaps/"+recapID.String()+"/actions/open_favorites",
				nil,
			)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			require.Equal(t, test.wantStatus, response.Code)
			if test.wantCode != "" {
				assert.Empty(t, response.Header().Get("Location"))
				assertErrorResponse(t, response, test.wantCode, test.wantMessage)

				return
			}

			assert.Equal(t, destination.String(), response.Header().Get("Location"))
		})
	}
}
