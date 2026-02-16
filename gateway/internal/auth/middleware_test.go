package auth

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohamadRazaviDev/Grafana-Gateway/gateway/internal/config"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestMiddlewareValidKey(t *testing.T) {
	rawKey := "test-key-123"
	validator := NewAPIKeyValidator([]config.APIKey{
		{Hash: HashKey(rawKey), User: "alice", Team: "sre"},
	})

	var capturedUser string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = r.Header.Get("X-WEBAUTH-USER")
		id := GetIdentity(r.Context())
		if id == nil {
			t.Error("expected identity in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(validator, "X-WEBAUTH-USER", nil, testLogger())
	handler := mw(next)

	req := httptest.NewRequest("GET", "/api/dashboards", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedUser != "alice" {
		t.Errorf("expected alice, got %s", capturedUser)
	}
}

func TestMiddlewareMissingAuth(t *testing.T) {
	validator := NewAPIKeyValidator(nil)
	mw := Middleware(validator, "X-WEBAUTH-USER", nil, testLogger())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach next handler")
	}))

	req := httptest.NewRequest("GET", "/api/dashboards", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiddlewareInvalidKey(t *testing.T) {
	validator := NewAPIKeyValidator([]config.APIKey{
		{Hash: HashKey("correct-key"), User: "alice", Team: "sre"},
	})
	mw := Middleware(validator, "X-WEBAUTH-USER", nil, testLogger())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach next handler")
	}))

	req := httptest.NewRequest("GET", "/api/dashboards", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestMiddlewareSkipPaths(t *testing.T) {
	validator := NewAPIKeyValidator(nil) // no keys
	mw := Middleware(validator, "X-WEBAUTH-USER", []string{"/healthz", "/readyz", "/metrics"}, testLogger())

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("expected handler to be called for skip path")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMiddlewareInvalidFormat(t *testing.T) {
	validator := NewAPIKeyValidator(nil)
	mw := Middleware(validator, "X-WEBAUTH-USER", nil, testLogger())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach next handler")
	}))

	req := httptest.NewRequest("GET", "/api/dashboards", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
