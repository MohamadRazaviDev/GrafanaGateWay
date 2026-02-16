# Grafana Gateway — Architecture

## High-Level Overview

```
┌─────────┐       ┌──────────────────────────────────────┐       ┌──────────┐
│  Client  │──────▶│          Grafana Gateway (:8080)      │──────▶│  Grafana  │
│ (Browser │       │                                      │       │  (:3000)  │
│  / API)  │       │  ┌─────────┐  ┌────────┐  ┌───────┐ │       │           │
└─────────┘       │  │ Rate    │→ │ Auth   │→ │Policy │ │       │ (auth    │
                  │  │ Limit   │  │ Check  │  │Engine │ │       │  proxy   │
                  │  └─────────┘  └────────┘  └───────┘ │       │  mode)   │
                  │       │            │           │      │       └──────────┘
                  │  ┌─────────┐  ┌────────┐  ┌───────┐ │
                  │  │ Audit   │  │Metrics │  │Request│ │
                  │  │ Logger  │  │(Prom)  │  │  ID   │ │
                  │  └─────────┘  └────────┘  └───────┘ │
                  └──────────────────────────────────────┘
                           │              │
                     ┌─────▼──┐    ┌──────▼─────┐
                     │ Audit  │    │ Prometheus  │
                     │  Logs  │    │  (scrape)   │
                     │(JSON)  │    │             │
                     └────────┘    └─────────────┘
```

## Request Flow

1. **Client** sends HTTP request to Gateway on port 8080
2. **Request ID** — unique `X-Request-ID` injected if not present
3. **Metrics** — request counter and latency timer started
4. **Audit Logger** — begins recording request metadata
5. **Rate Limiter** — checks per-user and per-IP token buckets
   - If exceeded → `429 Too Many Requests`
6. **Auth Middleware** — validates `Authorization: Bearer <key>`
   - Computes SHA-256 of key, constant-time compares against stored hashes
   - If invalid → `401 Unauthorized` or `403 Forbidden`
   - If valid → sets `X-WEBAUTH-USER` header with mapped username
7. **Policy Engine** — evaluates allow/deny rules
   - Checks path patterns against user/team policies
   - Admin endpoints (`/api/admin/*`) blocked by default
   - If denied → `403 Forbidden`
8. **Reverse Proxy** — forwards request to Grafana
   - HTTP: standard `httputil.ReverseProxy`
   - WebSocket: TCP-level bidirectional copy for `/api/live/*`
9. **Grafana** receives request with `X-WEBAUTH-USER` header
   - Auth proxy mode: trusts the header, creates/maps user session
10. **Response** flows back through the middleware chain
    - Audit log entry written (JSON)
    - Metrics updated (status code, latency)

## WebSocket Support (Grafana Live)

Grafana Live uses WebSocket connections at `/api/live/`.
The gateway detects `Connection: Upgrade` + `Upgrade: websocket` headers
and establishes a bidirectional TCP tunnel between client and Grafana.

## Authentication Flow

```
Client                    Gateway                    Grafana
  │                         │                          │
  │ Authorization: Bearer X │                          │
  │────────────────────────▶│                          │
  │                         │ SHA-256(X) == stored?    │
  │                         │ (constant-time compare)  │
  │                         │                          │
  │                         │ X-WEBAUTH-USER: alice    │
  │                         │─────────────────────────▶│
  │                         │                          │
  │      200 OK             │      200 OK              │
  │◀────────────────────────│◀─────────────────────────│
```

## Configuration Hierarchy

1. **Defaults** (secure by default — auth on, rate limit on, audit on)
2. **YAML config file** (`--config path/to/config.yaml`)
3. **Environment variables** (highest priority, override everything)

## Deployment Topology

```
┌─────────────────────────────────────────┐
│           Docker Network                │
│                                         │
│  ┌─────────┐    ┌──────────┐           │
│  │ Gateway  │───▶│ Grafana  │           │
│  │  :8080   │    │  :3000   │           │
│  └────┬─────┘    └──────────┘           │
│       │                                  │
│  ┌────▼─────┐  (optional)              │
│  │Prometheus │                          │
│  │  :9090    │                          │
│  └───────────┘                          │
└─────────────────────────────────────────┘
       ▲
       │ :8080 exposed to host
   ┌───┴───┐
   │ Users │
   └───────┘
```

Only the gateway port is exposed. Grafana is internal to the Docker network,
ensuring all access goes through the gateway's auth + policy + rate limit chain.
