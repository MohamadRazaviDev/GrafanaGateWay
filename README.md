# Grafana Gateway

A lightweight, cloud-native reverse proxy that sits in front of [Grafana](https://grafana.com/) and provides **authentication**, **authorization policies**, **rate limiting**, **audit logging**, and **Prometheus metrics** — all before traffic reaches Grafana.

```
┌────────┐     ┌─────────────────────┐     ┌──────────┐
│ Client │────▶│  Grafana Gateway    │────▶│  Grafana  │
│        │     │  :8080              │     │  :3000    │
└────────┘     │                     │     └──────────┘
               │ • Auth (API key)    │
               │ • Policy engine     │
               │ • Rate limiting     │
               │ • Audit logs (JSON) │
               │ • Prometheus /metrics│
               └─────────────────────┘
```

## Features

| Feature | Description |
|---------|-------------|
| **Reverse Proxy** | Forwards all HTTP + WebSocket traffic to Grafana (including Grafana Live `/api/live/`) |
| **Authentication** | API key validation with SHA-256 hashed keys + constant-time comparison |
| **Auth Proxy** | Forwards authenticated user as `X-WEBAUTH-USER` header (Grafana auth proxy mode) |
| **Authorization** | Path-based allow/deny policies per team/user; admin endpoints blocked by default |
| **Rate Limiting** | Token bucket per-user and per-IP with configurable RPS and burst |
| **Audit Logging** | Structured JSON logs for every request (user, path, status, latency, policy decision) |
| **Metrics** | Prometheus endpoint at `/metrics` with request counters, latency histograms, auth failures |
| **Health Checks** | `/healthz` (liveness) and `/readyz` (readiness) endpoints |
| **Request Tracing** | Auto-generated `X-Request-ID` header for distributed tracing |
| **Cloud Ready** | Docker multi-stage build, Docker Compose, Helm chart, GitHub Actions CI |

## Quick Start

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/)

### 1. Clone and configure

```bash
git clone https://github.com/MohamadRazaviDev/Grafana-Gateway.git
cd Grafana-Gateway
```

### 2. Generate an API key hash

```bash
# Using Docker
docker run --rm grafana-gateway --hash-key "my-secret-api-key"

# Or if you have Go installed
cd gateway && go run ./cmd/grafana-gateway --hash-key "my-secret-api-key"
```

### 3. Configure environment

```bash
cp deploy/.env.example deploy/.env
```

Edit `deploy/.env` and set:
- `GATEWAY_API_KEY_0_HASH` — the SHA-256 hash from step 2
- `GF_SECURITY_ADMIN_PASSWORD` — a strong Grafana admin password

### 4. Start everything

```bash
cd deploy
docker compose up -d
```

### 5. Test it

```bash
# Health check
curl http://localhost:8080/healthz

# Access Grafana through the gateway (with API key)
curl -H "Authorization: Bearer my-secret-api-key" http://localhost:8080/api/dashboards/home

# Metrics
curl http://localhost:8080/metrics
```

### Optional: Enable Prometheus monitoring

```bash
cd deploy
docker compose --profile monitoring up -d
# Prometheus UI at http://localhost:9090
```

## Configuration

All configuration is done via **environment variables** or a **YAML config file**. Environment variables always take precedence.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_LISTEN_ADDR` | `:8080` | Gateway listen address |
| `GRAFANA_URL` | `http://localhost:3000` | Grafana backend URL |
| `GRAFANA_AUTH_HEADER_NAME` | `X-WEBAUTH-USER` | Header name for auth proxy |
| `GATEWAY_AUTH_ENABLED` | `true` | Enable/disable authentication |
| `GATEWAY_API_KEY_N_HASH` | — | SHA-256 hash of API key N |
| `GATEWAY_API_KEY_N_USER` | — | Username mapped to API key N |
| `GATEWAY_API_KEY_N_TEAM` | — | Team mapped to API key N |
| `GATEWAY_API_KEY_N_ROLES` | — | Comma-separated roles for API key N |
| `GATEWAY_RATE_LIMIT_ENABLED` | `true` | Enable/disable rate limiting |
| `GATEWAY_RATE_LIMIT_RPS` | `50` | Requests per second |
| `GATEWAY_RATE_LIMIT_BURST` | `100` | Burst size |
| `GATEWAY_AUDIT_ENABLED` | `true` | Enable/disable audit logging |
| `GATEWAY_AUDIT_OUTPUT` | `stdout` | Audit log output (`stdout`, `stderr`, or file path) |
| `GATEWAY_METRICS_ENABLED` | `true` | Enable/disable Prometheus metrics |
| `GATEWAY_LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |

### YAML Config File

See [`gateway/config.example.yaml`](gateway/config.example.yaml) for a full example with policies.

```bash
grafana-gateway --config /path/to/config.yaml
```

### Example Policy (YAML)

```yaml
policies:
  - name: "sre-team"
    teams: ["sre"]
    allow_paths:
      - "/d/*"
      - "/api/dashboards/*"
    deny_paths:
      - "/api/admin/*"

  - name: "dev-readonly"
    teams: ["dev"]
    allow_paths:
      - "/d/*"
    allow_methods: ["GET"]
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed diagrams and request flow.

**Middleware chain** (outer → inner):
```
Request ID → Metrics → Audit → Rate Limit → Auth → Policy → Reverse Proxy → Grafana
```

## Project Structure

```
Grafana-Gateway/
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
  ui/                            # React admin dashboard (optional)
  deploy/
    docker-compose.yml           # Grafana + Gateway + Prometheus
    grafana/grafana.ini          # Grafana auth proxy config
    .env.example                 # Environment variable template
    prometheus.yml               # Prometheus scrape config
  helm/                          # Kubernetes Helm chart
  docs/
    architecture.md              # Architecture diagrams
    threat-model.md              # Security threat model
  .github/workflows/ci.yml      # CI pipeline
```

## Security

> ⚠️ **Important**: This gateway uses Grafana's [auth proxy](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/auth-proxy/) mode. This requires careful trust boundary management.

### Security Measures
- **API keys stored as SHA-256 hashes** — plaintext keys never touch config files
- **Constant-time comparison** — prevents timing attacks on key validation
- **Admin endpoints blocked by default** — `/api/admin/*` requires explicit policy
- **Deny-by-default policies** — deny takes precedence over allow
- **Rate limiting** — per-user and per-IP token buckets
- **Audit logging** — every request logged with user, path, status, latency
- **Non-root Docker container** — gateway runs as unprivileged user
- **Grafana not exposed** — only accessible through the gateway in Docker Compose

### Critical: Keep Grafana Updated
Grafana has published security advisories related to auth proxy configurations (e.g., [CVE-2022-35957](https://grafana.com/blog/2022/10/18/grafana-security-release-new-versions-with-security-fixes-for-cve-2022-39229-cve-2022-39201-and-cve-2022-31130/)). Always run the latest patched version.

See [docs/threat-model.md](docs/threat-model.md) for the full threat model.

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

## Development

For full setup instructions (virtual environment, tooling, make targets), see **[DEVELOPMENT.md](DEVELOPMENT.md)**.

Quick start:

```bash
# Automated setup (creates isolated .venv, installs all tools)
make setup
source .activate-env.sh

cd gateway

# Run tests
go test -race ./...

# Build
go build -o grafana-gateway ./cmd/grafana-gateway

# Run locally (auth disabled for development)
GRAFANA_URL=http://localhost:3000 GATEWAY_AUTH_ENABLED=false ./grafana-gateway

# Hash an API key
./grafana-gateway --hash-key "my-secret-key"
```

## License

MIT
