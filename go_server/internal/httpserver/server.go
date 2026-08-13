package httpserver

import (
	"fmt"
	"net/http"

	"dnf-launcher-go/internal/api"
	"dnf-launcher-go/internal/config"
	"dnf-launcher-go/internal/db"
	"dnf-launcher-go/internal/permissions"
	"dnf-launcher-go/internal/pvf"
)

func New(settings config.Settings) (*http.Server, error) {
	mux := http.NewServeMux()
	store, err := db.New(settings)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}
	if err := store.Ping(); err != nil {
		store.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	if err := permissions.EnsureTables(store.Tool()); err != nil {
		store.Close()
		return nil, fmt.Errorf("permissions table initialization failed: %w", err)
	}
	if err := pvf.EnsureTables(store.Tool()); err != nil {
		store.Close()
		return nil, fmt.Errorf("pvf table initialization failed: %w", err)
	}
	if err := pvf.ExpireStaleRefreshJobs(store); err != nil {
		store.Close()
		return nil, fmt.Errorf("pvf refresh job cleanup failed: %w", err)
	}
	api.RegisterRoutes(mux, settings, store)

	handler := withCORS(settings, mux)
	return &http.Server{
		Addr:    settings.ListenAddress(),
		Handler: handler,
	}, nil
}

func withCORS(settings config.Settings, next http.Handler) http.Handler {
	allowedOrigin := "*"
	if len(settings.CORSOrigins) > 0 && settings.CORSOrigins[0] != "" {
		allowedOrigin = settings.CORSOrigins[0]
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
