package observability

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// HealthChecker provides /healthz and /readyz endpoints.
type HealthChecker struct {
	ready atomic.Bool
}

// NewHealthChecker creates a new HealthChecker (starts not-ready).
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

// SetReady marks the gateway as ready to serve traffic.
func (h *HealthChecker) SetReady(ready bool) {
	h.ready.Store(ready)
}

// IsReady returns whether the gateway is ready.
func (h *HealthChecker) IsReady() bool {
	return h.ready.Load()
}

type healthResponse struct {
	Status string `json:"status"`
}

// HealthzHandler returns 200 if the process is alive.
func (h *HealthChecker) HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
	}
}

// ReadyzHandler returns 200 if the gateway is ready, 503 otherwise.
func (h *HealthChecker) ReadyzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if h.IsReady() {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(healthResponse{Status: "ready"})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(healthResponse{Status: "not ready"})
		}
	}
}
