package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.ListenAddr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.Server.ListenAddr)
	}
	if cfg.Grafana.URL != "http://localhost:3000" {
		t.Errorf("expected http://localhost:3000, got %s", cfg.Grafana.URL)
	}
	if !cfg.Auth.Enabled {
		t.Error("expected auth enabled by default")
	}
	if !cfg.RateLimit.Enabled {
		t.Error("expected rate limit enabled by default")
	}
	if !cfg.Audit.Enabled {
		t.Error("expected audit enabled by default")
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("GRAFANA_URL", "http://grafana:3000")
	t.Setenv("GATEWAY_LISTEN_ADDR", ":9090")
	t.Setenv("GATEWAY_AUTH_ENABLED", "false")
	t.Setenv("GATEWAY_RATE_LIMIT_RPS", "100")
	t.Setenv("GATEWAY_RATE_LIMIT_BURST", "200")

	cfg := DefaultConfig()
	applyEnvOverrides(cfg)

	if cfg.Grafana.URL != "http://grafana:3000" {
		t.Errorf("expected http://grafana:3000, got %s", cfg.Grafana.URL)
	}
	if cfg.Server.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.Server.ListenAddr)
	}
	if cfg.Auth.Enabled {
		t.Error("expected auth disabled via env")
	}
	if cfg.RateLimit.RequestsPerSec != 100 {
		t.Errorf("expected 100 rps, got %f", cfg.RateLimit.RequestsPerSec)
	}
}

func TestEnvAPIKeys(t *testing.T) {
	t.Setenv("GATEWAY_API_KEY_0_HASH", "abc123hash")
	t.Setenv("GATEWAY_API_KEY_0_USER", "alice")
	t.Setenv("GATEWAY_API_KEY_0_TEAM", "sre")
	t.Setenv("GATEWAY_API_KEY_0_ROLES", "viewer,editor")
	t.Setenv("GATEWAY_API_KEY_1_HASH", "def456hash")
	t.Setenv("GATEWAY_API_KEY_1_USER", "bob")
	t.Setenv("GATEWAY_API_KEY_1_TEAM", "dev")

	cfg := DefaultConfig()
	cfg.Auth.APIKeys = nil
	applyEnvOverrides(cfg)

	if len(cfg.Auth.APIKeys) != 2 {
		t.Fatalf("expected 2 api keys, got %d", len(cfg.Auth.APIKeys))
	}
	if cfg.Auth.APIKeys[0].User != "alice" {
		t.Errorf("expected alice, got %s", cfg.Auth.APIKeys[0].User)
	}
	if len(cfg.Auth.APIKeys[0].Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(cfg.Auth.APIKeys[0].Roles))
	}
}

func TestLoadFromYAML(t *testing.T) {
	yamlContent := `
server:
  listen_addr: ":9999"
grafana:
  url: "http://test-grafana:3000"
auth:
  enabled: true
  api_keys:
    - hash: "testhash"
      user: "testuser"
      team: "testteam"
rate_limit:
  enabled: true
  requests_per_sec: 10
  burst: 20
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.ListenAddr != ":9999" {
		t.Errorf("expected :9999, got %s", cfg.Server.ListenAddr)
	}
	if cfg.Grafana.URL != "http://test-grafana:3000" {
		t.Errorf("expected http://test-grafana:3000, got %s", cfg.Grafana.URL)
	}
}

func TestValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Auth.Enabled = true
	cfg.Auth.APIKeys = nil
	if err := validate(cfg); err == nil {
		t.Error("expected validation error for auth enabled with no keys")
	}

	cfg.Auth.Enabled = false
	cfg.Grafana.URL = ""
	if err := validate(cfg); err == nil {
		t.Error("expected validation error for empty grafana url")
	}
}
