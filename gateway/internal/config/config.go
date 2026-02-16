package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all gateway configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Grafana  GrafanaConfig  `yaml:"grafana"`
	Auth     AuthConfig     `yaml:"auth"`
	Policies []Policy       `yaml:"policies"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Audit    AuditConfig    `yaml:"audit"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

type ServerConfig struct {
	ListenAddr      string        `yaml:"listen_addr"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type GrafanaConfig struct {
	URL             string `yaml:"url"`
	AuthHeaderName  string `yaml:"auth_header_name"`
	AuthHeaderUser  string `yaml:"auth_header_user"`
}

type AuthConfig struct {
	Enabled     bool      `yaml:"enabled"`
	APIKeys     []APIKey  `yaml:"api_keys"`
	SkipPaths   []string  `yaml:"skip_paths"`
}

type APIKey struct {
	// SHA-256 hash of the actual key (never store plaintext)
	Hash     string   `yaml:"hash"`
	User     string   `yaml:"user"`
	Team     string   `yaml:"team"`
	Roles    []string `yaml:"roles"`
}

type Policy struct {
	Name          string   `yaml:"name"`
	Users         []string `yaml:"users"`
	Teams         []string `yaml:"teams"`
	AllowPaths    []string `yaml:"allow_paths"`
	DenyPaths     []string `yaml:"deny_paths"`
	AllowMethods  []string `yaml:"allow_methods"`
}

type RateLimitConfig struct {
	Enabled       bool    `yaml:"enabled"`
	RequestsPerSec float64 `yaml:"requests_per_sec"`
	Burst         int     `yaml:"burst"`
	PerUser       bool    `yaml:"per_user"`
	PerIP         bool    `yaml:"per_ip"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

type AuditConfig struct {
	Enabled bool   `yaml:"enabled"`
	Output  string `yaml:"output"` // "stdout" or file path
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// DefaultConfig returns a secure-by-default configuration.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddr:      ":8080",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 15 * time.Second,
		},
		Grafana: GrafanaConfig{
			URL:            "http://localhost:3000",
			AuthHeaderName: "X-WEBAUTH-USER",
			AuthHeaderUser: "",
		},
		Auth: AuthConfig{
			Enabled:   true,
			SkipPaths: []string{"/healthz", "/readyz", "/metrics"},
		},
		RateLimit: RateLimitConfig{
			Enabled:         true,
			RequestsPerSec:  50,
			Burst:           100,
			PerUser:         true,
			PerIP:           true,
			CleanupInterval: 5 * time.Minute,
		},
		Audit: AuditConfig{
			Enabled: true,
			Output:  "stdout",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}

// Load reads config from a YAML file, then applies env var overrides.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	applyEnvOverrides(cfg)

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides to config.
// Env vars take precedence over YAML values.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("GATEWAY_LISTEN_ADDR"); v != "" {
		cfg.Server.ListenAddr = v
	}
	if v := os.Getenv("GATEWAY_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ReadTimeout = d
		}
	}
	if v := os.Getenv("GATEWAY_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.WriteTimeout = d
		}
	}
	if v := os.Getenv("GRAFANA_URL"); v != "" {
		cfg.Grafana.URL = v
	}
	if v := os.Getenv("GRAFANA_AUTH_HEADER_NAME"); v != "" {
		cfg.Grafana.AuthHeaderName = v
	}
	if v := os.Getenv("GATEWAY_AUTH_ENABLED"); v != "" {
		cfg.Auth.Enabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("GATEWAY_RATE_LIMIT_ENABLED"); v != "" {
		cfg.RateLimit.Enabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("GATEWAY_RATE_LIMIT_RPS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.RateLimit.RequestsPerSec = f
		}
	}
	if v := os.Getenv("GATEWAY_RATE_LIMIT_BURST"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.RateLimit.Burst = i
		}
	}
	if v := os.Getenv("GATEWAY_AUDIT_ENABLED"); v != "" {
		cfg.Audit.Enabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("GATEWAY_AUDIT_OUTPUT"); v != "" {
		cfg.Audit.Output = v
	}
	if v := os.Getenv("GATEWAY_METRICS_ENABLED"); v != "" {
		cfg.Metrics.Enabled = strings.EqualFold(v, "true") || v == "1"
	}

	// API keys from env: GATEWAY_API_KEY_0_HASH, GATEWAY_API_KEY_0_USER, etc.
	for i := 0; ; i++ {
		prefix := fmt.Sprintf("GATEWAY_API_KEY_%d_", i)
		hash := os.Getenv(prefix + "HASH")
		if hash == "" {
			break
		}
		key := APIKey{
			Hash: hash,
			User: os.Getenv(prefix + "USER"),
			Team: os.Getenv(prefix + "TEAM"),
		}
		if roles := os.Getenv(prefix + "ROLES"); roles != "" {
			key.Roles = strings.Split(roles, ",")
		}
		cfg.Auth.APIKeys = append(cfg.Auth.APIKeys, key)
	}
}

func validate(cfg *Config) error {
	if cfg.Grafana.URL == "" {
		return fmt.Errorf("grafana.url is required")
	}
	if cfg.Server.ListenAddr == "" {
		return fmt.Errorf("server.listen_addr is required")
	}
	if cfg.Auth.Enabled && len(cfg.Auth.APIKeys) == 0 {
		return fmt.Errorf("auth is enabled but no API keys configured; set GATEWAY_API_KEY_0_HASH or add keys to config")
	}
	if cfg.RateLimit.Enabled && cfg.RateLimit.RequestsPerSec <= 0 {
		return fmt.Errorf("rate_limit.requests_per_sec must be positive when rate limiting is enabled")
	}
	return nil
}
