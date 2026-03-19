package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ksred/cctrack/internal/config"
	"github.com/ksred/cctrack/internal/hub"
	"github.com/ksred/cctrack/internal/store"
)

func newTestAPI(t *testing.T) (*API, *http.ServeMux, func()) {
	t.Helper()

	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	h := hub.New()
	h.Start()

	cfg := &config.Config{Port: 7432}
	a := New(s, h, cfg)

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	cleanup := func() {
		h.Stop()
		s.Close()
	}
	return a, mux, cleanup
}

func TestPostSettings_RejectsOversizedBody(t *testing.T) {
	_, mux, cleanup := newTestAPI(t)
	defer cleanup()

	// Create a body larger than 1 MB
	huge := `{"monthly_budget_usd":` + strings.Repeat("1", 1<<20) + `}`
	req := httptest.NewRequest("POST", "/api/v1/settings", strings.NewReader(huge))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// LimitReader truncates the body, causing invalid JSON → 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized body, got %d", rec.Code)
	}
}

func TestPostSettings_AcceptsNormalBody(t *testing.T) {
	_, mux, cleanup := newTestAPI(t)
	defer cleanup()

	body := strings.NewReader(`{"monthly_budget_usd": 50.0}`)
	req := httptest.NewRequest("POST", "/api/v1/settings", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for normal body, got %d; body: %s", rec.Code, rec.Body.String())
	}
}
