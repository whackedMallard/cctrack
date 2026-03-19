package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/ksred/cctrack/internal/config"
	"github.com/ksred/cctrack/internal/hub"
	"github.com/ksred/cctrack/internal/store"
)

// newTestAPI creates an API with a temp DB and returns a cleanup function.
func newTestAPI(t *testing.T, port int) (*API, func()) {
	t.Helper()

	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	h := hub.New()
	h.Start()

	cfg := &config.Config{Port: port}
	a := New(s, h, cfg)

	cleanup := func() {
		h.Stop()
		s.Close()
	}
	return a, cleanup
}

func TestWebSocket_AllowsSameOrigin(t *testing.T) {
	// Create the test server first to discover the actual port,
	// then set cfg.Port to match so OriginPatterns align.
	a, cleanup := newTestAPI(t, 0)
	defer cleanup()

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Extract actual port and update config so OriginPatterns match
	actualPort := ts.Listener.Addr().(*net.TCPAddr).Port
	a.cfg.Port = actualPort

	wsURL := "ws" + ts.URL[len("http"):] + "/api/v1/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	origin := fmt.Sprintf("http://127.0.0.1:%d", actualPort)
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{origin}},
	})
	if err != nil {
		t.Fatalf("WebSocket dial with same origin failed: %v", err)
	}
	conn.Close(websocket.StatusNormalClosure, "done")
}

func TestWebSocket_AllowsLocalhostOrigin(t *testing.T) {
	a, cleanup := newTestAPI(t, 0)
	defer cleanup()

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	actualPort := ts.Listener.Addr().(*net.TCPAddr).Port
	a.cfg.Port = actualPort

	wsURL := "ws" + ts.URL[len("http"):] + "/api/v1/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	origin := fmt.Sprintf("http://localhost:%d", actualPort)
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{origin}},
	})
	if err != nil {
		t.Fatalf("WebSocket dial with localhost origin failed: %v", err)
	}
	conn.Close(websocket.StatusNormalClosure, "done")
}

func TestWebSocket_RejectsCrossOrigin(t *testing.T) {
	// Use a fixed port for OriginPatterns — the cross-origin "http://evil.com"
	// won't match regardless of port.
	a, cleanup := newTestAPI(t, 9999)
	defer cleanup()

	mux := http.NewServeMux()
	a.RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + ts.URL[len("http"):] + "/api/v1/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"http://evil.com"}},
	})
	if err == nil {
		t.Fatal("WebSocket dial with cross-origin should have been rejected")
	}
}
