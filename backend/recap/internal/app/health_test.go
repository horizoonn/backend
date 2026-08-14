package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type healthCheckerStub struct {
	err error
}

func (s healthCheckerStub) Ping(context.Context) error {
	return s.err
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	registerHealthEndpoints(mux, nil, time.Second)

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ok\n", response.Body.String())
}

func TestReadinessEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "database is ready",
			wantStatus: http.StatusOK,
			wantBody:   "ok\n",
		},
		{
			name:       "database is unavailable",
			pingErr:    errors.New("database unavailable"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "not ready\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			registerHealthEndpoints(mux, healthCheckerStub{err: tt.pingErr}, time.Second)

			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil)
			response := httptest.NewRecorder()

			mux.ServeHTTP(response, request)

			require.Equal(t, tt.wantStatus, response.Code)
			require.Equal(t, tt.wantBody, response.Body.String())
		})
	}
}
