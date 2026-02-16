package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/audit"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/auth"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/config"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/observability"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/policy"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/proxy"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/ratelimit"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	var (
		configPath  string
		showVersion bool
		hashKey     string
	)

	flag.StringVar(&configPath, "config", "", "Path to config file (YAML)")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.StringVar(&hashKey, "hash-key", "", "Hash an API key and print the SHA-256 hash, then exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("grafana-gateway %s (commit: %s, built: %s)\n", version, commit, buildDate)
		os.Exit(0)
	}

	if hashKey != "" {
		fmt.Println(auth.HashKey(hashKey))
		os.Exit(0)
	}

	// Initialize structured logger
	logLevel := slog.LevelInfo
	if os.Getenv("GATEWAY_LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("starting grafana-gateway",
		"version", version,
		"listen_addr", cfg.Server.ListenAddr,
		"grafana_url", cfg.Grafana.URL,
	)

	// Health checker
	health := observability.NewHealthChecker()
	health.SetGrafanaURL(cfg.Grafana.URL)

	// Prometheus metrics
	reg := prometheus.DefaultRegisterer
	metrics := observability.NewMetrics(reg)

	// Create reverse proxy
	proxyHandler, err := proxy.New(cfg.Grafana.URL, logger)
	if err != nil {
		logger.Error("failed to create proxy", "error", err)
		os.Exit(1)
	}

	// Build middleware chain (innermost first, outermost last)
	handler := proxyHandler

	// Policy engine (if policies configured)
	policyEngine := policy.NewEngine(cfg.Policies, logger)
	handler = policyEngine.Middleware()(handler)

	// Auth middleware
	if cfg.Auth.Enabled {
		validator := auth.NewAPIKeyValidator(cfg.Auth.APIKeys)
		handler = auth.Middleware(validator, cfg.Grafana.AuthHeaderName, cfg.Auth.SkipPaths, logger)(handler)
	}

	// Rate limiting
	if cfg.RateLimit.Enabled {
		limiter := ratelimit.New(cfg.RateLimit, logger)
		handler = limiter.Middleware()(handler)
	}

	// Audit logging
	if cfg.Audit.Enabled {
		auditLogger, err := audit.New(cfg.Audit.Output, logger)
		if err != nil {
			logger.Error("failed to create audit logger", "error", err)
			os.Exit(1)
		}
		handler = auditLogger.Middleware()(handler)
	}

	// Metrics middleware (outermost — captures everything)
	handler = metrics.Middleware()(handler)

	// Request ID injection
	handler = requestIDMiddleware(handler)

	// Build mux
	mux := http.NewServeMux()
	mux.Handle("/healthz", health.HealthzHandler())
	mux.Handle("/readyz", health.ReadyzHandler())
	if cfg.Metrics.Enabled {
		mux.Handle(cfg.Metrics.Path, observability.Handler())
	}
	mux.Handle("/", handler)

	server := &http.Server{
		Addr:         cfg.Server.ListenAddr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		health.SetReady(true)
		logger.Info("gateway is ready", "addr", cfg.Server.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("shutting down gateway...")
	health.SetReady(false)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("gateway stopped")
}

// requestIDMiddleware injects a unique X-Request-ID header if not present.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-ID") == "" {
			id := generateRequestID()
			r.Header.Set("X-Request-ID", id)
		}
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r)
	})
}

func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
