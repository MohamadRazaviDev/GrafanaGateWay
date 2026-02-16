package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRegistration(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	if m.RequestsTotal == nil {
		t.Error("expected RequestsTotal to be registered")
	}
	if m.RequestDuration == nil {
		t.Error("expected RequestDuration to be registered")
	}
	if m.AuthFailures == nil {
		t.Error("expected AuthFailures to be registered")
	}

	// Verify metrics can be gathered
	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	// promauto registers so they should be gatherable (some may have 0 samples)
	_ = families
}

func TestMetricsMiddleware(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := m.Middleware()(next)

	req := httptest.NewRequest("GET", "/api/dashboards/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range families {
		if f.GetName() == "gateway_requests_total" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected gateway_requests_total metric")
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/healthz", "/healthz"},
		{"/readyz", "/readyz"},
		{"/metrics", "/metrics"},
		{"/d/abc123", "/d/:uid"},
		{"/api/admin/users", "/api/admin/:action"},
		{"/api/dashboards/uid/abc", "/api/dashboards/:action"},
		{"/api/live/channel1", "/api/live/:channel"},
		{"/some/other/path", "/some/other/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePath(tt.input)
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
