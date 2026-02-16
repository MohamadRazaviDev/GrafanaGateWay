package ratelimit

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/MohamadRazaviDev/Grafana-Gateway/gateway/internal/auth"
	"github.com/MohamadRazaviDev/Grafana-Gateway/gateway/internal/config"
)

// Limiter provides per-user and per-IP rate limiting.
type Limiter struct {
	mu       sync.RWMutex
	limiters map[string]*entry
	rps      rate.Limit
	burst    int
	perUser  bool
	perIP    bool
	logger   *slog.Logger
}

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// New creates a new rate limiter from config.
func New(cfg config.RateLimitConfig, logger *slog.Logger) *Limiter {
	l := &Limiter{
		limiters: make(map[string]*entry),
		rps:      rate.Limit(cfg.RequestsPerSec),
		burst:    cfg.Burst,
		perUser:  cfg.PerUser,
		perIP:    cfg.PerIP,
		logger:   logger,
	}

	// Start cleanup goroutine to remove stale entries
	cleanupInterval := cfg.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = 5 * time.Minute
	}
	go l.cleanup(cleanupInterval)

	return l
}

// Allow checks if a request should be allowed.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	e, exists := l.limiters[key]
	if !exists {
		e = &entry{
			limiter:  rate.NewLimiter(l.rps, l.burst),
			lastSeen: time.Now(),
		}
		l.limiters[key] = e
	}
	e.lastSeen = time.Now()
	l.mu.Unlock()

	return e.limiter.Allow()
}

// Middleware creates an HTTP middleware for rate limiting.
func (l *Limiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keys := l.extractKeys(r)
			for _, key := range keys {
				if !l.Allow(key) {
					l.logger.Warn("rate limit exceeded",
						"key", key,
						"path", r.URL.Path,
						"remote_addr", r.RemoteAddr,
					)
					w.Header().Set("Retry-After", "1")
					http.Error(w, `{"error":"too_many_requests","message":"rate limit exceeded"}`, http.StatusTooManyRequests)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *Limiter) extractKeys(r *http.Request) []string {
	var keys []string
	if l.perUser {
		if identity := auth.GetIdentity(r.Context()); identity != nil {
			keys = append(keys, "user:"+identity.User)
		}
	}
	if l.perIP {
		ip := extractIP(r)
		if ip != "" {
			keys = append(keys, "ip:"+ip)
		}
	}
	// Fallback: at least rate limit by IP if no keys extracted
	if len(keys) == 0 {
		ip := extractIP(r)
		if ip != "" {
			keys = append(keys, "ip:"+ip)
		}
	}
	return keys
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (trusted proxy scenario)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (client IP)
		parts := splitFirst(xff, ",")
		ip := net.ParseIP(trimSpace(parts))
		if ip != nil {
			return ip.String()
		}
	}
	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			return s[:i]
		}
	}
	return s
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

func (l *Limiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		for key, e := range l.limiters {
			if time.Since(e.lastSeen) > 10*time.Minute {
				delete(l.limiters, key)
			}
		}
		l.mu.Unlock()
	}
}
