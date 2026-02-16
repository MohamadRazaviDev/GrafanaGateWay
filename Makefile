.PHONY: help test lint fmt format clean check install-tools security

# ── Help ───────────────────────────────────────────────────────
help:
	@echo "Grafana Gateway — Development Targets"
	@echo ""
	@echo "  make check         Run all checks (format + lint + test) — use before PR"
	@echo "  make test          Run Go tests with race detector + coverage"
	@echo "  make lint          Run golangci-lint"
	@echo "  make fmt           Format all Go code (gofmt + goimports)"
	@echo "  make security      Run govulncheck + secrets scan"
	@echo "  make clean         Remove build artifacts"
	@echo "  make install-tools Install development tools"
	@echo "  make docker        Build Docker image"
	@echo ""

# ── Go Test ────────────────────────────────────────────────────
.PHONY: test
test:
	@echo "🧪 Running Go tests..."
	cd gateway && go test -race -count=1 -coverprofile=coverage.out ./...
	@echo ""
	@cd gateway && go tool cover -func=coverage.out | grep total
	@echo "✅ All tests passed!"

# ── Go Lint ────────────────────────────────────────────────────
.PHONY: lint
lint:
	@echo "🔍 Running go vet..."
	cd gateway && go vet ./...
	@echo "🔍 Running golangci-lint..."
	cd gateway && golangci-lint run ./...
	@echo "✅ Lint passed!"

# ── Go Format ──────────────────────────────────────────────────
.PHONY: fmt format
fmt format:
	@echo "📝 Formatting Go code..."
	gofmt -s -w gateway/
	@command -v goimports > /dev/null 2>&1 && goimports -w -local github.com/MohamadRazaviDev/GrafanaGateWay gateway/ || echo "⚠️  goimports not found (install: go install golang.org/x/tools/cmd/goimports@latest)"
	@echo "✅ Formatted!"

# ── Format Check (CI-style, no writes) ────────────────────────
.PHONY: fmt-check
fmt-check:
	@echo "🔍 Checking format..."
	@unformatted=$$(gofmt -l gateway/); \
	if [ -n "$$unformatted" ]; then \
		echo "❌ Unformatted files:"; \
		echo "$$unformatted"; \
		echo "Run: make fmt"; \
		exit 1; \
	fi
	@echo "✅ All files formatted!"

# ── Security ──────────────────────────────────────────────────
.PHONY: security
security:
	@echo "🔒 Running govulncheck..."
	@command -v govulncheck > /dev/null 2>&1 && (cd gateway && govulncheck ./...) || echo "⚠️  govulncheck not found (install: go install golang.org/x/vuln/cmd/govulncheck@latest)"
	@echo "🔒 Scanning for secrets..."
	@if grep -rnE 'AKIA[0-9A-Z]{16}|sk-[a-zA-Z0-9]{48}|ghp_[a-zA-Z0-9]{36}' \
		--include='*.go' --include='*.yaml' --include='*.yml' \
		gateway/ deploy/ docs/ 2>/dev/null; then \
		echo "❌ Potential secrets found!"; \
		exit 1; \
	fi
	@echo "✅ Security checks passed!"

# ── Pre-PR Check (runs everything) ────────────────────────────
.PHONY: check
check: fmt-check lint test security
	@echo ""
	@echo "🎉 All checks passed! Ready for PR."

# ── Docker ─────────────────────────────────────────────────────
.PHONY: docker
docker:
	@echo "🐳 Building Docker image..."
	docker build -t grafana-gateway:local gateway/
	@echo "✅ Docker image built: grafana-gateway:local"

# ── Clean ──────────────────────────────────────────────────────
.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	cd gateway && rm -f coverage.out grafana-gateway
	@echo "✅ Clean!"

# ── Install Tools ─────────────────────────────────────────────
.PHONY: install-tools
install-tools:
	@echo "📦 Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "✅ Tools installed!"
