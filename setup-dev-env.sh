#!/bin/bash

set -e

echo "🚀 Setting up Grafana Gateway development environment..."
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VENV_DIR="$PROJECT_ROOT/.venv"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check Python
echo "📋 Checking prerequisites..."
if ! command_exists python3; then
    echo -e "${RED}❌ python3 is required but not installed${NC}"
    echo "Install from: https://www.python.org/downloads/"
    exit 1
fi
echo -e "${GREEN}✅ Python 3 found${NC}"

# Check Node.js
if ! command_exists node; then
    echo -e "${RED}❌ Node.js is required but not installed${NC}"
    echo "Install from: https://nodejs.org/"
    exit 1
fi
echo -e "${GREEN}✅ Node.js found ($(node --version))${NC}"

# Check npm
if ! command_exists npm; then
    echo -e "${RED}❌ npm is required but not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ npm found ($(npm --version))${NC}"

# Check Go
if ! command_exists go; then
    echo -e "${YELLOW}⚠️  Go is not installed in your PATH${NC}"
    echo "Install from: https://golang.org/dl/"
    echo "Or set GO_ROOT environment variable"
    echo ""
    read -p "Continue without Go? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✅ Go found ($(go version))${NC}"
fi

echo ""
echo "📦 Setting up project..."
echo ""

# Create Python virtual environment
echo "1️⃣  Creating Python virtual environment..."
if [ -d "$VENV_DIR" ]; then
    echo "   Virtual environment already exists at $VENV_DIR"
else
    python3 -m venv "$VENV_DIR"
    echo -e "${GREEN}✅ Virtual environment created${NC}"
fi

# Activate virtual environment
source "$VENV_DIR/bin/activate"
echo -e "${GREEN}✅ Virtual environment activated${NC}"

# Upgrade pip
echo ""
echo "2️⃣  Upgrading pip..."
pip install --upgrade pip setuptools wheel > /dev/null 2>&1
echo -e "${GREEN}✅ pip upgraded${NC}"

# Install Python development tools
echo ""
echo "3️⃣  Installing Python development tools (ruff, black)..."
pip install ruff black > /dev/null 2>&1
echo -e "${GREEN}✅ Python tools installed${NC}"

# Install Node.js dependencies for UI
echo ""
echo "4️⃣  Installing Node.js dependencies..."
cd "$PROJECT_ROOT/ui"
npm install > /dev/null 2>&1
echo -e "${GREEN}✅ npm dependencies installed${NC}"

# Go modules
if command_exists go; then
    echo ""
    echo "5️⃣  Downloading Go dependencies..."
    cd "$PROJECT_ROOT/gateway"
    go mod download
    go mod tidy
    echo -e "${GREEN}✅ Go dependencies downloaded${NC}"
fi

# Create activate script
echo ""
echo "6️⃣  Creating environment activation script..."
cat > "$PROJECT_ROOT/.activate-env.sh" << 'EOF'
#!/bin/bash
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$PROJECT_ROOT/.venv/bin/activate"
export PYTHONPATH="$PROJECT_ROOT:$PYTHONPATH"
export PATH="$PROJECT_ROOT/gateway/bin:$PATH"
echo "✅ Development environment activated!"
echo "   Python: $(python --version 2>&1)"
echo "   pip: $(pip --version)"
echo "   node: $(node --version)"
if command -v go >/dev/null 2>&1; then
    echo "   go: $(go version | awk '{print $3}')"
fi
EOF
chmod +x "$PROJECT_ROOT/.activate-env.sh"
echo -e "${GREEN}✅ Activation script created${NC}"

# Create .env.local file
echo ""
echo "7️⃣  Creating .env.local file..."
cat > "$PROJECT_ROOT/.env.local" << 'EOF'
# Grafana Gateway Development Environment
# Source: source .activate-env.sh

# Python
PYTHONPATH=.

# Go
GO111MODULE=on

# Node
NODE_ENV=development

# Gateway
GATEWAY_LOG_LEVEL=debug
GATEWAY_LISTEN_ADDR=:8080
GRAFANA_URL=http://localhost:3000
GATEWAY_AUTH_ENABLED=true
GATEWAY_RATE_LIMIT_ENABLED=true
GATEWAY_AUDIT_ENABLED=true
GATEWAY_METRICS_ENABLED=true
EOF
echo -e "${GREEN}✅ .env.local created${NC}"

# Summary
echo ""
echo "════════════════════════════════════════════════════════════"
echo -e "${GREEN}✅ Development environment setup complete!${NC}"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "📝 Quick Start:"
echo ""
echo "1. Activate the environment:"
echo "   ${YELLOW}source .activate-env.sh${NC}"
echo ""
echo "2. Load environment variables:"
echo "   ${YELLOW}set -a && source .env.local && set +a${NC}"
echo ""
echo "3. Build and run the gateway:"
echo "   ${YELLOW}cd gateway && go build -o grafana-gateway ./cmd/grafana-gateway${NC}"
echo "   ${YELLOW}./grafana-gateway -config config.example.yaml${NC}"
echo ""
echo "4. Available make commands:"
echo "   ${YELLOW}make test${NC}   - Run all tests"
echo "   ${YELLOW}make lint${NC}   - Run all linters"
echo "   ${YELLOW}make fmt${NC}    - Format code"
echo "   ${YELLOW}make check${NC}  - Pre-commit checks (lint + test)"
echo ""
echo "5. Virtual environment location:"
echo "   ${YELLOW}$VENV_DIR${NC}"
echo ""
echo "📦 Installed components:"
echo "   • Python virtual environment"
echo "   • Python dev tools: ruff, black"
echo "   • Node.js dependencies (ui/node_modules)"
if command_exists go; then
    echo "   • Go modules (gateway/go.mod)"
fi
echo ""
echo "🔧 To deactivate environment later:"
echo "   ${YELLOW}deactivate${NC}"
echo ""
