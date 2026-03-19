package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksred/cctrack/internal/config"
	"github.com/ksred/cctrack/internal/hub"
	"github.com/ksred/cctrack/internal/store"
)

// newTestAPI creates an API with a temp DB and returns a cleanup function.
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

// --- P1-1: validateLogDir tests ---

func TestValidateLogDir_ValidAbsolutePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	// Use a directory that definitely exists under home
	err := validateLogDir(home)
	if err != nil {
		t.Errorf("home dir should be valid: %v", err)
	}
}

func TestValidateLogDir_ValidTildePath(t *testing.T) {
	// ~/ expands to home, which exists
	err := validateLogDir("~/")
	if err != nil {
		t.Errorf("~/ should be valid: %v", err)
	}
}

func TestValidateLogDir_RejectsOutsideHome(t *testing.T) {
	err := validateLogDir("/etc")
	if err == nil {
		t.Error("expected error for /etc, got nil")
	}
	if !strings.Contains(err.Error(), "within your home directory") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateLogDir_RejectsRelativePath(t *testing.T) {
	err := validateLogDir("../../../etc")
	if err == nil {
		t.Error("expected error for relative path, got nil")
	}
	if !strings.Contains(err.Error(), "absolute path") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateLogDir_RejectsTraversal(t *testing.T) {
	home, _ := os.UserHomeDir()
	// Path that starts in home but traverses out
	traversal := filepath.Join(home, "..", "..", "etc")
	err := validateLogDir(traversal)
	if err == nil {
		t.Error("expected error for path traversal, got nil")
	}
}

func TestValidateLogDir_RejectsNonexistent(t *testing.T) {
	home, _ := os.UserHomeDir()
	err := validateLogDir(filepath.Join(home, "definitely-does-not-exist-xyz"))
	if err == nil {
		t.Error("expected error for nonexistent path, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateLogDir_RejectsFile(t *testing.T) {
	// Create a temp file (not a directory) under home
	home, _ := os.UserHomeDir()
	tmpFile := filepath.Join(home, ".cctrack-test-file-tmp")
	if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	err := validateLogDir(tmpFile)
	if err == nil {
		t.Error("expected error for file path, got nil")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// --- P1-1: POST /api/v1/settings integration test ---

func TestPostSettings_RejectsInvalidLogDir(t *testing.T) {
	_, mux, cleanup := newTestAPI(t)
	defer cleanup()

	body := strings.NewReader(`{"log_dir": "/etc"}`)
	req := httptest.NewRequest("POST", "/api/v1/settings", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for /etc log_dir, got %d", rec.Code)
	}
}

func TestPostSettings_AcceptsValidLogDir(t *testing.T) {
	_, mux, cleanup := newTestAPI(t)
	defer cleanup()

	home, _ := os.UserHomeDir()
	body := strings.NewReader(`{"log_dir": "` + home + `"}`)
	req := httptest.NewRequest("POST", "/api/v1/settings", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid log_dir, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

// --- P1-3: Error sanitisation tests ---

func TestInternalError_HidesDetails(t *testing.T) {
	rec := httptest.NewRecorder()
	internalError(rec, "test-context", os.ErrPermission)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != "internal server error" {
		t.Errorf("expected generic error message, got %q", body)
	}

	// Ensure the real error is NOT in the response body
	if strings.Contains(body, "permission") {
		t.Error("response body leaks the real error")
	}
}

func TestSummaryEndpoint_ReturnsGenericErrorOnFailure(t *testing.T) {
	// Create an API backed by an already-closed store to force errors
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	h := hub.New()
	h.Start()
	defer h.Stop()

	cfg := &config.Config{Port: 7432}
	a := New(s, h, cfg)
	mux := http.NewServeMux()
	a.RegisterRoutes(mux)

	// Close the store to force DB errors
	s.Close()

	req := httptest.NewRequest("GET", "/api/v1/summary", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != "internal server error" {
		t.Errorf("expected generic error, got %q", body)
	}
}
