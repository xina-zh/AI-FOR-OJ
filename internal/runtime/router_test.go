package runtime

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/handler"
	"ai-for-oj/internal/service"
)

type fakeDB struct {
	err error
}

func (f fakeDB) PingContext(context.Context) error {
	return f.err
}

func TestHealthRoute(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Name: "ai-for-oj",
			Env:  "test",
		},
	}

	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestHealthRouteReturns503WhenDatabaseIsDown(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Name: "ai-for-oj",
			Env:  "test",
		},
	}

	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{err: errors.New("db down")})
	healthHandler := handler.NewHealthHandler(healthService)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, resp.Code)
	}
}
