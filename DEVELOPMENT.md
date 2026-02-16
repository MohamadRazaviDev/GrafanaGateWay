# Development Setup Guide

This guide will help you set up a complete development environment for the Grafana Gateway project locally.

## Prerequisites

You need to have these installed on your system:

- **Python 3.8+** - [Download](https://www.python.org/downloads/)
- **Node.js 16+** - [Download](https://nodejs.org/)
- **npm 7+** - Comes with Node.js
- **Go 1.22+** (optional, for gateway development) - [Download](https://golang.org/dl/)

## Quick Setup (Recommended)

Run the automated setup script:

```bash
./setup-dev-env.sh
```

This will:
✅ Create a Python virtual environment (`.venv`)  
✅ Install Python dev tools (ruff, black)  
✅ Install Node.js dependencies  
✅ Download Go modules  
✅ Generate environment scripts  

## Manual Setup

### 1. Create Python Virtual Environment

```bash
python3 -m venv .venv
source .venv/bin/activate  # On Windows: .venv\Scripts\activate
```

### 2. Install Python Tools

```bash
pip install --upgrade pip
pip install ruff black pylance
```

### 3. Install Node.js Dependencies

```bash
cd ui
npm install
cd ..
```

### 4. Download Go Dependencies

```bash
cd gateway
go mod download
go mod tidy
cd ..
```

### 5. Load Environment Variables

```bash
set -a && source .env.local && set +a
```

## Activate Environment

After initial setup, activate your dev environment anytime with:

```bash
source .activate-env.sh
```

Or use Make:

```bash
make setup
source .activate-env.sh
```

## Development Workflow

### Running Tests

```bash
make test           # All tests (Go + TypeScript)
make test-go        # Go tests only
make test-ui        # TypeScript tests only
```

### Linting

```bash
make lint           # All linters (Go + Python + etc)
make lint-go        # Go linting only  
make lint-ui        # TypeScript linting only
make lint-python    # Python tools only
```

### Formatting Code

```bash
make fmt            # Format all code
make fmt-go         # Go formatting
make fmt-ui         # TypeScript/UI formatting
make fmt-python     # Python formatting
```

### Pre-commit Checks

Run tests and linting together:

```bash
make check          # Lint + Test
```

### Cleaning Up

```bash
make clean          # Remove build artifacts and cache
```

## Project Structure

```
.
├── gateway/               # Go backend
│   ├── cmd/              # Entry points
│   ├── internal/         # Core packages
│   ├── go.mod            # Go dependencies
│   └── Dockerfile        # Container image
├── ui/                    # TypeScript/React frontend
│   ├── components/       # React components
│   ├── services/         # API services
│   ├── package.json      # npm dependencies
│   └── vite.config.ts    # Build config
├── deploy/               # Docker Compose & configs
├── helm/                 # Kubernetes Helm charts
├── .venv/                # Python virtual environment (created by setup)
├── .activate-env.sh      # Environment activation script (created by setup)
├── .env.local            # Local env variables (created by setup)
├── Makefile              # Build automation
└── setup-dev-env.sh      # Automated setup script
```

## Environment Variables

Key variables in `.env.local`:

```bash
# Gateway
GATEWAY_LOG_LEVEL=debug
GATEWAY_LISTEN_ADDR=:8080
GRAFANA_URL=http://localhost:3000
GATEWAY_AUTH_ENABLED=true
GATEWAY_RATE_LIMIT_ENABLED=true
```

## Building & Running

### Build Gateway Binary

```bash
cd gateway
go build -o grafana-gateway ./cmd/grafana-gateway
```

### Run Gateway with Example Config

```bash
./grafana-gateway -config config.example.yaml
```

### Build UI

```bash
cd ui
npm run build
```

## Virtual Environment Isolation

Everything is contained in `.venv/`:
- ✅ Doesn't conflict with other projects
- ✅ Easy to delete: `rm -rf .venv`
- ✅ Added to `.gitignore`
- ✅ All tools installed locally

To deactivate:

```bash
deactivate
```

## Troubleshooting

### Python not found
```bash
# Check Python installation
python3 --version
which python3
```

### Go not found
Go is optional. If you don't have it, you can still work on the UI/frontend.

### Virtual environment not activated
```bash
# Always run first:
source .activate-env.sh
```

### Dependencies failed to install
```bash
# Clear and reinstall
rm -rf .venv ui/node_modules gateway/go.sum
./setup-dev-env.sh
```

### Permission denied on setup script
```bash
chmod +x setup-dev-env.sh
./setup-dev-env.sh
```

## CI/CD Integration

The Makefile also works in CI/CD pipelines:

```bash
make install-tools  # Install tools
make check         # Run full validation
```

## Tips & Best Practices

1. **Always activate before working:**
   ```bash
   source .activate-env.sh
   ```

2. **Use Make for consistency:**
   ```bash
   make fmt   # Format before committing
   make check # Validate before pushing
   ```

3. **Keep virtual environment clean:**
   - Don't modify `.venv` manually
   - Recreate if issues: `rm -rf .venv && ./setup-dev-env.sh`

4. **Update dependencies:**
   ```bash
   cd ui && npm update
   cd ../gateway && go get -u ./...
   ```

## Support

For issues with setup:
1. Check this guide
2. Review CI configuration in `.github/workflows/ci.yml`
3. Open an issue with your error output

## Next Steps

- Run `make setup` to start
- Read [architecture.md](docs/architecture.md) to understand the system
- Check [SECURITY.md](SECURITY.md) for security guidelines
