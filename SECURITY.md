# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Grafana Gateway, please report it responsibly.

**DO NOT** open a public GitHub issue for security vulnerabilities.

### How to Report

1. Email: Send details to the repository owner via GitHub private vulnerability reporting
2. Go to the **Security** tab of this repository → **Report a vulnerability**
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- Acknowledgment within **48 hours**
- Status update within **7 days**
- We aim to release fixes within **30 days** of confirmed vulnerabilities

### Scope

The following are in scope:
- Authentication bypass
- Authorization policy bypass
- Rate limit bypass
- Header injection
- Information disclosure via audit logs or error messages
- Denial of service via resource exhaustion

### Out of Scope

- Vulnerabilities in Grafana itself (report to [Grafana Security](https://grafana.com/docs/grafana/latest/administration/security/))
- Issues requiring physical access
- Social engineering

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | ✅                 |
| < latest | ❌                |

## Security Best Practices for Operators

1. **Never expose Grafana directly** — always route through the gateway
2. **Use TLS** in front of the gateway
3. **Keep Grafana updated** — auth proxy mode has had security advisories (e.g., CVE-2022-35957)
4. **Rotate API keys** regularly
5. **Monitor audit logs** for suspicious activity
6. **Review authorization policies** periodically
