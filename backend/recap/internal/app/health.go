package app

import (
	"context"
	"net/http"
	"time"
)

type healthChecker interface {
	Ping(ctx context.Context) error
}

func registerHealthEndpoints(
	mux *http.ServeMux,
	checker healthChecker,
	readinessTimeout time.Duration,
) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if err := checker.Ping(ctx); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)

			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
}
