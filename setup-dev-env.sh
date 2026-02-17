#!/bin/bash
set -e

echo "🚀 Setting up Grafana Gateway development environment..."
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# ── Prerequisites ───────────────────────────────────────────────────

echo "📋 Checking prerequisites..."

if ! command_exists go; then
    echo -e "${RED}❌ Go is required but not installed${NC}"
    echo "Install from: https://golang.org/dl/"
    exit 1
fi
echo -e "${GREEN}✅ Go found ($(go version | awk '{print $3}'))${NC}"

if ! command_exists node; then
    echo -e "${YELLOW}⚠️  Node.js not found (optional, for UI development)${NC}"
else
    echo -e "${GREEN}✅ Node.js found ($(node --version))${NC}"
fi

# ── Go Dependencies ────────────────────────────────────────────────

echo ""
echo "1️⃣  Downloading Go dependencies..."
cd "$PROJECT_ROOT/gateway"
go mod download
go mod tidy
echo -e "${GREEN}✅ Go dependencies downloaded${NC}"

# ── Go Tools ────────────────────────────────────────────────────────

echo ""
echo "2️⃣  Installing Go development tools..."
make -C "$PROJECT_ROOT" install-tools 2>/dev/null || {
    echo "   Installing golangci-lint, goimports, govulncheck..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 2>/dev/null
    go install golang.org/x/tools/cmd/goimports@latest 2>/dev/null
    go install golang.org/x/vuln/cmd/govulncheck@latest 2>/dev/null
}
echo -e "${GREEN}✅ Go tools installed${NC}"

# ── Summary ─────────────────────────────────────────────────────────

echo ""
echo "════════════════════════════════════════════════════════════"
echo -e "${GREEN}✅ Development environment setup complete!${NC}"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "📝 Quick Start:"
echo ""
echo "  Build & run:"
echo "    ${YELLOW}cd gateway && go build -o grafana-gateway ./cmd/grafana-gateway${NC}"
echo "    ${YELLOW}GRAFANA_URL=http://localhost:3000 GATEWAY_AUTH_ENABLED=false ./grafana-gateway${NC}"
echo ""
echo "  Make commands:"
echo "    ${YELLOW}make check${NC}    — Pre-PR gate (format + lint + test + security)"
echo "    ${YELLOW}make test${NC}     — Run tests with race detector"
echo "    ${YELLOW}make lint${NC}     — Run golangci-lint"
echo "    ${YELLOW}make fmt${NC}      — Format code (gofmt + goimports)"
echo "    ${YELLOW}make security${NC} — Vulnerability scan (govulncheck)"
echo ""
echo "  Docker Compose:"
echo "    ${YELLOW}cd deploy && docker compose up -d${NC}"
echo ""
