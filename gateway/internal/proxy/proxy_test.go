package proxy

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReverseProxy(t *testing.T) {
	// Create a mock Grafana backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Backend", "grafana")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("grafana response"))
	}))
	defer backend.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	proxy, err := New(backend.URL, logger)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/dashboards", nil)
	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "grafana response" {
		t.Errorf("expected 'grafana response', got %q", rr.Body.String())
	}
}

func TestReverseProxyForwardsHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("X-WEBAUTH-USER")
		_, _ = w.Write([]byte("user:" + user))
	}))
	defer backend.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	proxy, err := New(backend.URL, logger)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-WEBAUTH-USER", "alice")
	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Body.String() != "user:alice" {
		t.Errorf("expected 'user:alice', got %q", rr.Body.String())
	}
}

func TestIsWebSocketUpgrade(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    bool
	}{
		{"websocket", map[string]string{"Connection": "Upgrade", "Upgrade": "websocket"}, true},
		{"no upgrade", map[string]string{}, false},
		{"only connection", map[string]string{"Connection": "Upgrade"}, false},
		{"case insensitive", map[string]string{"Connection": "upgrade", "Upgrade": "WebSocket"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/live/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if got := isWebSocketUpgrade(req); got != tt.want {
				t.Errorf("isWebSocketUpgrade() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyBadBackend(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	proxy, err := New("http://127.0.0.1:1", logger)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
}

func TestHopByHopHeaderStripping(t *testing.T) {
	// Verify that hop-by-hop headers are NOT forwarded to the backend.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// These hop-by-hop headers should be stripped by the reverse proxy
		if v := r.Header.Get("Keep-Alive"); v != "" {
			t.Errorf("Keep-Alive header should be stripped, got %q", v)
		}
		if v := r.Header.Get("Proxy-Authenticate"); v != "" {
			t.Errorf("Proxy-Authenticate header should be stripped, got %q", v)
		}
		if v := r.Header.Get("Proxy-Authorization"); v != "" {
			t.Errorf("Proxy-Authorization header should be stripped, got %q", v)
		}
		// Note: Go's ReverseProxy preserves "TE: trailers" per RFC 7230 §4.3
		// but strips other TE values. This is correct behavior.
		if v := r.Header.Get("Trailer"); v != "" {
			t.Errorf("Trailer header should be stripped, got %q", v)
		}
		// X-Custom-Header should be preserved (not hop-by-hop)
		if v := r.Header.Get("X-Custom-Header"); v != "keep-me" {
			t.Errorf("expected X-Custom-Header=keep-me, got %q", v)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	proxy, err := New(backend.URL, logger)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/dashboards", nil)
	req.Header.Set("Keep-Alive", "timeout=5")
	req.Header.Set("Proxy-Authenticate", "Basic")
	req.Header.Set("Proxy-Authorization", "Basic dXNlcjpwYXNz")
	req.Header.Set("TE", "trailers")
	req.Header.Set("Trailer", "X-Checksum")
	req.Header.Set("X-Custom-Header", "keep-me")

	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
