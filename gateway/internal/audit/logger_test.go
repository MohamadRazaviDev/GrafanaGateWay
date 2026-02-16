package audit

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestLogEntry(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		encoder: json.NewEncoder(&buf),
		writer:  &buf,
		slog:    testLogger(),
	}

	logger.Log(Entry{
		Timestamp:  "2025-01-01T00:00:00Z",
		RequestID:  "req-123",
		User:       "alice",
		Team:       "sre",
		ClientIP:   "1.2.3.4",
		Method:     "GET",
		Path:       "/api/dashboards",
		StatusCode: 200,
		LatencyMs:  42,
	})

	var entry Entry
	if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
		t.Fatal(err)
	}
	if entry.User != "alice" {
		t.Errorf("expected alice, got %s", entry.User)
	}
	if entry.StatusCode != 200 {
		t.Errorf("expected 200, got %d", entry.StatusCode)
	}
}

func TestMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		encoder: json.NewEncoder(&buf),
		writer:  &buf,
		slog:    testLogger(),
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := logger.Middleware()(next)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "5.6.7.8:9999"
	req.Header.Set("X-Request-ID", "test-req-id")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var entry Entry
	if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
		t.Fatal(err)
	}
	if entry.Method != "GET" {
		t.Errorf("expected GET, got %s", entry.Method)
	}
	if entry.Path != "/api/test" {
		t.Errorf("expected /api/test, got %s", entry.Path)
	}
	if entry.RequestID != "test-req-id" {
		t.Errorf("expected test-req-id, got %s", entry.RequestID)
	}
	if entry.User != "_anonymous" {
		t.Errorf("expected _anonymous, got %s", entry.User)
	}
}

func TestResponseRecorder(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		encoder: json.NewEncoder(&buf),
		writer:  &buf,
		slog:    testLogger(),
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := logger.Middleware()(next)
	req := httptest.NewRequest("GET", "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var entry Entry
	_ = json.NewDecoder(&buf).Decode(&entry)
	if entry.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", entry.StatusCode)
	}
}
