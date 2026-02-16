package ratelimit

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohamadRazaviDev/Grafana-Gateway/gateway/internal/config"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestAllow(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:         true,
		RequestsPerSec:  10,
		Burst:           2,
		PerIP:           true,
		CleanupInterval: time.Hour,
	}
	limiter := New(cfg, testLogger())

	// First 2 requests should be allowed (burst)
	if !limiter.Allow("test-key") {
		t.Error("expected first request to be allowed")
	}
	if !limiter.Allow("test-key") {
		t.Error("expected second request to be allowed")
	}
}

func TestExceedBurst(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:         true,
		RequestsPerSec:  1,
		Burst:           1,
		PerIP:           true,
		CleanupInterval: time.Hour,
	}
	limiter := New(cfg, testLogger())

	limiter.Allow("burst-test")
	// Second immediate request should be denied (burst=1)
	if limiter.Allow("burst-test") {
		t.Error("expected request to be rate limited after exceeding burst")
	}
}

func TestDifferentKeysIndependent(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:         true,
		RequestsPerSec:  1,
		Burst:           1,
		PerIP:           true,
		CleanupInterval: time.Hour,
	}
	limiter := New(cfg, testLogger())

	limiter.Allow("key-a")
	// key-b should have its own bucket
	if !limiter.Allow("key-b") {
		t.Error("expected different key to have independent limit")
	}
}

func TestMiddleware429(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:         true,
		RequestsPerSec:  1,
		Burst:           1,
		PerIP:           true,
		CleanupInterval: time.Hour,
	}
	limiter := New(cfg, testLogger())

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := limiter.Middleware()(next)

	// First request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	// Second immediate request should be rate limited
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr.Code)
	}

	if rr.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header")
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		{"remote addr", "1.2.3.4:5678", "", "1.2.3.4"},
		{"xff", "10.0.0.1:5678", "9.8.7.6, 10.0.0.1", "9.8.7.6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			got := extractIP(req)
			if got != tt.want {
				t.Errorf("extractIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
