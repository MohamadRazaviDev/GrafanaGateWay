package observability

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzAlwaysOK(t *testing.T) {
	hc := NewHealthChecker()
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	hc.HealthzHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp healthResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("expected ok, got %s", resp.Status)
	}
}

func TestReadyzNotReady(t *testing.T) {
	hc := NewHealthChecker()
	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	hc.ReadyzHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestReadyzReady(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetReady(true)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	hc.ReadyzHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp readyzResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ready" {
		t.Errorf("expected ready, got %s", resp.Status)
	}
}

func TestReadyzWithGrafanaUp(t *testing.T) {
	// Mock a healthy Grafana backend
	grafana := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"database":"ok"}`))
	}))
	defer grafana.Close()

	hc := NewHealthChecker()
	hc.SetReady(true)
	hc.SetGrafanaURL(grafana.URL)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	hc.ReadyzHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp readyzResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Grafana != "ok" {
		t.Errorf("expected grafana=ok, got %s", resp.Grafana)
	}
}

func TestReadyzWithGrafanaDown(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetReady(true)
	hc.SetGrafanaURL("http://127.0.0.1:1") // unreachable

	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	hc.ReadyzHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}

	var resp readyzResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Grafana != "unreachable" {
		t.Errorf("expected grafana=unreachable, got %s", resp.Grafana)
	}
}
