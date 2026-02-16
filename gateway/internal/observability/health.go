package observability

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthChecker provides /healthz and /readyz endpoints.
type HealthChecker struct {
	ready      atomic.Bool
	grafanaURL string
}

// NewHealthChecker creates a new HealthChecker (starts not-ready).
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

// SetGrafanaURL enables upstream Grafana health checking in /readyz.
func (h *HealthChecker) SetGrafanaURL(url string) {
	h.grafanaURL = url
}

// SetReady marks the gateway as ready to serve traffic.
func (h *HealthChecker) SetReady(ready bool) {
	h.ready.Store(ready)
}

// IsReady returns whether the gateway is ready.
func (h *HealthChecker) IsReady() bool {
	return h.ready.Load()
}

type readyzResponse struct {
	Status  string `json:"status"`
	Grafana string `json:"grafana,omitempty"`
}

type healthResponse struct {
	Status string `json:"status"`
}

// HealthzHandler returns 200 if the process is alive.
func (h *HealthChecker) HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
	}
}

// ReadyzHandler returns 200 if the gateway is ready, 503 otherwise.
// If a Grafana URL is configured, it also probes Grafana's /api/health endpoint.
func (h *HealthChecker) ReadyzHandler() http.HandlerFunc {
	client := &http.Client{Timeout: 3 * time.Second}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !h.IsReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(readyzResponse{Status: "not ready"})
			return
		}

		// Check upstream Grafana if configured
		grafanaStatus := ""
		if h.grafanaURL != "" {
			resp, err := client.Get(h.grafanaURL + "/api/health")
			if err != nil || resp.StatusCode != http.StatusOK {
				grafanaStatus = "unreachable"
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(readyzResponse{
					Status:  "not ready",
					Grafana: grafanaStatus,
				})
				if resp != nil {
					_ = resp.Body.Close()
				}
				return
			}
			_ = resp.Body.Close()
			grafanaStatus = "ok"
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(readyzResponse{
			Status:  "ready",
			Grafana: grafanaStatus,
		})
	}
}
