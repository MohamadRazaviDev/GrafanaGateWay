# Grafana Gateway — Threat Model

## Overview

This document identifies key security risks in the Grafana Gateway architecture
and the mitigations in place. It follows a simplified STRIDE approach.

---

## Trust Boundary

```
UNTRUSTED                    TRUST BOUNDARY                 TRUSTED
─────────────────────────────────┬──────────────────────────────────
  Internet / Internal Network    │    Docker Network
  ┌────────┐                     │    ┌─────────┐    ┌─────────┐
  │ Client │─────── :8080 ──────▶│───▶│ Gateway │───▶│ Grafana │
  └────────┘                     │    └─────────┘    └─────────┘
─────────────────────────────────┴──────────────────────────────────
```

**Key principle**: Grafana trusts the `X-WEBAUTH-USER` header from the gateway.
If an attacker can bypass the gateway and reach Grafana directly, they can
impersonate any user by setting this header.

---

## Threats & Mitigations

### T1: Auth Proxy Header Injection

**Risk**: HIGH — If Grafana is exposed directly, attackers can set
`X-WEBAUTH-USER` and impersonate any user.

**Mitigations**:
- Grafana is **not exposed** to the host in Docker Compose (only `expose: 3000`)
- `grafana.ini` sets `whitelist = gateway` to only accept proxy headers from the gateway container
- **Keep Grafana updated** — Grafana has published security advisories related to auth proxy (e.g., CVE-2022-35957). Always run the latest patch version.

### T2: API Key Compromise

**Risk**: MEDIUM — Stolen API keys allow full access under that key's permissions.

**Mitigations**:
- Keys stored as **SHA-256 hashes** (never plaintext)
- Constant-time comparison prevents timing attacks
- Rate limiting reduces blast radius of compromised keys
- Audit logs enable detection of unusual activity
- Keys can be rotated by updating config (no code changes)

### T3: Rate Limit Bypass

**Risk**: LOW — Attacker uses multiple IPs or forged headers to bypass rate limits.

**Mitigations**:
- Rate limiting is applied **per-user AND per-IP** (dual bucket)
- `X-Forwarded-For` is only trusted from the first hop
- Global rate limits can be added as a backstop

### T4: Unauthorized Dashboard Access

**Risk**: MEDIUM — Users access dashboards they shouldn't see.

**Mitigations**:
- Policy engine enforces **path-based allow/deny** per team/user
- **Deny takes precedence** over allow
- Admin endpoints (`/api/admin/*`) are **blocked by default**
- Policies loaded from config file — no runtime modification via API

### T5: Audit Log Tampering

**Risk**: LOW — Attacker modifies audit logs to cover tracks.

**Mitigations**:
- Logs written to stdout (captured by Docker/container runtime)
- In production, ship logs to external system (ELK, Loki, CloudWatch)
- File-based output uses append-only mode (0600 permissions)

### T6: Denial of Service

**Risk**: MEDIUM — Resource exhaustion via high request volume.

**Mitigations**:
- Per-user and per-IP rate limiting with configurable thresholds
- HTTP server timeouts (read, write, idle)
- Graceful shutdown prevents connection drops
- Docker healthcheck enables container restart on failure

### T7: Secrets in Source Control

**Risk**: HIGH — Accidental commit of API keys or passwords.

**Mitigations**:
- `.env` files are in `.gitignore`
- Config example files use placeholder values only
- API keys stored as hashes (even if config is leaked, keys are not reversible)
- `--hash-key` CLI tool for generating hashes locally

### T8: TLS / Man-in-the-Middle

**Risk**: MEDIUM — Traffic between client and gateway intercepted.

**Mitigations**:
- **Recommendation**: Deploy TLS termination in front of gateway (nginx, cloud LB, or direct TLS)
- Internal Docker network traffic is isolated
- TLS 1.2 minimum enforced on proxy-to-Grafana connections if HTTPS

---

## Recommendations for Production

1. **Always use TLS** in front of the gateway (terminate at load balancer or add TLS to gateway)
2. **Keep Grafana updated** — auth proxy has had CVEs; patch promptly
3. **Rotate API keys** periodically and audit key usage
4. **Ship audit logs** to an external, immutable log store
5. **Network segmentation** — ensure Grafana is only reachable from the gateway
6. **Monitor metrics** — set alerts on `gateway_auth_failures_total` and `gateway_rate_limit_hits_total`
7. **Review policies** regularly — ensure they follow least-privilege principle
