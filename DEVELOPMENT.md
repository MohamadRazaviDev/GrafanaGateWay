# Development Setup Guide

## Prerequisites

| Tool | Version | Required | Install |
|------|---------|----------|---------|
| **Go** | 1.22+ | ✅ Yes | [golang.org/dl](https://golang.org/dl/) |
| **Docker** | 20+ | ✅ Yes | [docker.com](https://docs.docker.com/get-docker/) |
| **Node.js** | 16+ | ❌ Optional | [nodejs.org](https://nodejs.org/) |

## Quick Setup

```bash
# Automated (installs Go tools + downloads deps)
./setup-dev-env.sh

# Or manually
cd gateway
go mod download
cd ..
make install-tools
```

## Daily Workflow

```bash
# Run all pre-PR checks (format + lint + test + security)
make check

# Individual commands
make test       # go test -race ./...
make lint       # golangci-lint
make fmt        # gofmt + goimports
make security   # govulncheck
```

## Building & Running

### Build the Gateway

```bash
cd gateway
go build -o grafana-gateway ./cmd/grafana-gateway
```

### Run Locally (dev mode)

```bash
GRAFANA_URL=http://localhost:3000 GATEWAY_AUTH_ENABLED=false ./grafana-gateway
```

### Run with Docker Compose

```bash
cd deploy
cp .env.example .env          # edit .env with your API key hash
docker compose up -d           # gateway + Grafana + Prometheus
```

### Hash an API Key

```bash
cd gateway
go run ./cmd/grafana-gateway --hash-key "my-secret-key"
```

## Project Structure

```
GrafanaGateWay/
  gateway/                       # Go gateway service
    cmd/grafana-gateway/         # CLI entry point
    internal/
      config/                    # YAML + env config loading
      proxy/                     # Reverse proxy + WebSocket
      auth/                      # API key validation + middleware
      policy/                    # Authorization engine
      ratelimit/                 # Token bucket rate limiter
      audit/                     # Structured JSON audit logger
      observability/             # Health checks + Prometheus metrics
    Dockerfile                   # Multi-stage Docker build
    config.example.yaml          # Full config example
    .golangci.yml                # Linter configuration (v2)
  deploy/
    docker-compose.yml           # Grafana + Gateway + Prometheus
    grafana/grafana.ini          # Grafana auth proxy config
    .env.example                 # Environment variable template
    prometheus.yml               # Prometheus scrape config
  helm/                          # Kubernetes Helm chart
  docs/
    architecture.md              # Architecture diagrams
    threat-model.md              # Security threat model
  .github/workflows/ci.yml      # 5-stage CI pipeline
  Makefile                       # Build automation
```

## Environment Variables

Key variables for local development (set in `deploy/.env`):

```bash
GATEWAY_LOG_LEVEL=debug
GATEWAY_LISTEN_ADDR=:8080
GRAFANA_URL=http://localhost:3000
GATEWAY_AUTH_ENABLED=true
GATEWAY_RATE_LIMIT_ENABLED=true
GATEWAY_AUDIT_ENABLED=true
GATEWAY_METRICS_ENABLED=true
```

## CI/CD

The GitHub Actions pipeline runs automatically on pushes and PRs:

```bash
# Reproduce CI locally
make check          # format + lint + test + security
make docker         # build Docker image
```

## Troubleshooting

### Go not found
```bash
go version          # should print go1.22+
which go            # verify in PATH
```

### Dependencies failed
```bash
cd gateway
rm go.sum
go mod tidy
```

### Linter issues
```bash
make install-tools  # reinstall golangci-lint
make lint           # run linter
```
