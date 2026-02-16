package observability

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	hc := NewHealthChecker()
	handler := hc.HealthzHandler()

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp healthResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("expected ok, got %s", resp.Status)
	}
}

func TestReadyzNotReady(t *testing.T) {
	hc := NewHealthChecker()
	handler := hc.ReadyzHandler()

	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestReadyzReady(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetReady(true)
	handler := hc.ReadyzHandler()

	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp healthResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ready" {
		t.Errorf("expected ready, got %s", resp.Status)
	}
}
