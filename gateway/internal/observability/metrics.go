package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the gateway.
type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	AuthFailures     prometheus.Counter
	RateLimitHits    prometheus.Counter
	PolicyDenials    prometheus.Counter
	ActiveWebSockets prometheus.Gauge
}

// NewMetrics registers and returns all gateway metrics.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)

	return &Metrics{
		RequestsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of HTTP requests processed.",
		}, []string{"method", "path", "status"}),

		RequestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),

		AuthFailures: factory.NewCounter(prometheus.CounterOpts{
			Name: "gateway_auth_failures_total",
			Help: "Total number of authentication failures.",
		}),

		RateLimitHits: factory.NewCounter(prometheus.CounterOpts{
			Name: "gateway_rate_limit_hits_total",
			Help: "Total number of rate-limited requests.",
		}),

		PolicyDenials: factory.NewCounter(prometheus.CounterOpts{
			Name: "gateway_policy_denials_total",
			Help: "Total number of policy-denied requests.",
		}),

		ActiveWebSockets: factory.NewGauge(prometheus.GaugeOpts{
			Name: "gateway_active_websockets",
			Help: "Number of active WebSocket connections.",
		}),
	}
}

// Handler returns the Prometheus metrics HTTP handler.
func Handler() http.Handler {
	return promhttp.Handler()
}

// Middleware creates an HTTP middleware that records request metrics.
func (m *Metrics) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rec.statusCode)

			// Normalize path to avoid high cardinality
			normPath := normalizePath(r.URL.Path)

			m.RequestsTotal.WithLabelValues(r.Method, normPath, status).Inc()
			m.RequestDuration.WithLabelValues(r.Method, normPath).Observe(duration)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// normalizePath reduces path cardinality for metrics labels.
func normalizePath(p string) string {
	switch {
	case p == "/healthz" || p == "/readyz" || p == "/metrics":
		return p
	case len(p) >= 4 && p[:3] == "/d/":
		return "/d/:uid"
	case len(p) > 11 && p[:11] == "/api/admin/":
		return "/api/admin/:action"
	case len(p) > 16 && p[:16] == "/api/dashboards/":
		return "/api/dashboards/:action"
	case len(p) > 10 && p[:10] == "/api/live/":
		return "/api/live/:channel"
	default:
		// Catch-all for /api/ paths to prevent label cardinality explosion
		if len(p) > 5 && p[:5] == "/api/" {
			return "/api/:other"
		}
		return p
	}
}
