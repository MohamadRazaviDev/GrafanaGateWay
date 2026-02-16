package audit

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MohamadRazaviDev/Grafana-Gateway/gateway/internal/auth"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp      string `json:"timestamp"`
	RequestID      string `json:"request_id"`
	User           string `json:"user"`
	Team           string `json:"team"`
	ClientIP       string `json:"client_ip"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	StatusCode     int    `json:"status_code"`
	LatencyMs      int64  `json:"latency_ms"`
	PolicyDecision string `json:"policy_decision"`
	UserAgent      string `json:"user_agent"`
}

// Logger writes structured audit log entries.
type Logger struct {
	encoder *json.Encoder
	mu      sync.Mutex
	writer  io.Writer
	slog    *slog.Logger
}

// New creates a new audit logger.
// output can be "stdout", "stderr", or a file path.
func New(output string, logger *slog.Logger) (*Logger, error) {
	var writer io.Writer

	switch output {
	case "stdout", "":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, err
		}
		writer = f
	}

	return &Logger{
		encoder: json.NewEncoder(writer),
		writer:  writer,
		slog:    logger,
	}, nil
}

// Log writes an audit entry.
func (l *Logger) Log(entry Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.encoder.Encode(entry); err != nil {
		l.slog.Error("failed to write audit log", "error", err)
	}
}

// responseRecorder captures the status code from the response.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// Middleware creates an HTTP middleware that logs all requests.
func (l *Logger) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			latency := time.Since(start).Milliseconds()

			user := "_anonymous"
			team := ""
			if identity := auth.GetIdentity(r.Context()); identity != nil {
				user = identity.User
				team = identity.Team
			}

			entry := Entry{
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				RequestID:  r.Header.Get("X-Request-ID"),
				User:       user,
				Team:       team,
				ClientIP:   r.RemoteAddr,
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rec.statusCode,
				LatencyMs:  latency,
				UserAgent:  r.Header.Get("User-Agent"),
			}

			l.Log(entry)
		})
	}
}
