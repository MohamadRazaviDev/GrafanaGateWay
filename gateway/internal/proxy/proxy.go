package proxy

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// New creates a reverse proxy to the given Grafana URL.
// It handles both HTTP and WebSocket (Grafana Live) connections.
func New(grafanaURL string, logger *slog.Logger) (http.Handler, error) {
	target, err := url.Parse(grafanaURL)
	if err != nil {
		return nil, fmt.Errorf("parsing grafana url: %w", err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.SetXForwarded()
			r.Out.Host = target.Host
		},
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("proxy error",
				"method", r.Method,
				"path", r.URL.Path,
				"error", err,
			)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isWebSocketUpgrade(r) {
			handleWebSocket(w, r, target, logger)
			return
		}
		proxy.ServeHTTP(w, r)
	}), nil
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade.
func isWebSocketUpgrade(r *http.Request) bool {
	for _, v := range r.Header["Connection"] {
		if strings.EqualFold(strings.TrimSpace(v), "upgrade") {
			if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
				return true
			}
		}
	}
	return false
}

// handleWebSocket proxies WebSocket connections (for Grafana Live /api/live/).
func handleWebSocket(w http.ResponseWriter, r *http.Request, target *url.URL, logger *slog.Logger) {
	scheme := "ws"
	if target.Scheme == "https" {
		scheme = "wss"
	}

	backendURL := &url.URL{
		Scheme:   scheme,
		Host:     target.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	backendConn, err := dialer.Dial("tcp", target.Host)
	if err != nil {
		logger.Error("websocket dial failed", "target", backendURL.String(), "error", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		logger.Error("websocket hijack not supported")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		backendConn.Close()
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		logger.Error("websocket hijack failed", "error", err)
		backendConn.Close()
		return
	}

	// Forward the original request to backend
	reqURL := *r.URL
	reqURL.Host = target.Host
	reqURL.Scheme = scheme

	if err := r.Write(backendConn); err != nil {
		logger.Error("websocket write request failed", "error", err)
		clientConn.Close()
		backendConn.Close()
		return
	}

	logger.Info("websocket connection established",
		"path", r.URL.Path,
		"backend", backendURL.String(),
	)

	// Bidirectional copy
	errc := make(chan error, 2)
	go func() {
		_, err := copyConn(backendConn, clientConn)
		errc <- err
	}()
	go func() {
		_, err := copyConn(clientConn, backendConn)
		errc <- err
	}()

	// Wait for one direction to finish, then close both
	<-errc
	clientConn.Close()
	backendConn.Close()
}

func copyConn(dst, src net.Conn) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			total += int64(written)
			if writeErr != nil {
				return total, writeErr
			}
		}
		if readErr != nil {
			return total, readErr
		}
	}
}
