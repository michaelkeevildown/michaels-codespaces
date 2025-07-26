#!/bin/sh
# One-liner MCS installer for testing branch
# Usage: curl -fsSL url | MCS_BRANCH=feat/mcs-go-status-command sh

set -e

# Configuration
MCS_HOME="${MCS_HOME:-$HOME/.mcs}"
BRANCH="${MCS_BRANCH:-feat/mcs-go-status-command}"
REPO="michaelkeevildown/michaels-codespaces"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
esac
PLATFORM="${OS}-${ARCH}"

echo "🚀 Installing MCS ($PLATFORM) from branch: $BRANCH"

# Quick method: build from source if Go is available
if command -v go >/dev/null 2>&1; then
    echo "📦 Building from source..."
    TEMP=$(mktemp -d)
    trap 'rm -rf "$TEMP"' EXIT
    
    git clone -q -b "$BRANCH" "https://github.com/$REPO.git" "$TEMP"
    cd "$TEMP/mcs-go"
    
    mkdir -p "$MCS_HOME/bin"
    go mod download
    go build -o "$MCS_HOME/bin/mcs" cmd/mcs/main.go
    
    echo "✅ Built successfully"
else
    echo "❌ Go not found. Installing Go first..."
    
    # Install Go on Linux
    if [ "$OS" = "linux" ]; then
        GO_VERSION="1.21.5"
        wget -q "https://go.dev/dl/go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
        sudo tar -C /usr/local -xzf "go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
        rm "go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
        export PATH=/usr/local/go/bin:$PATH
        
        # Now build
        TEMP=$(mktemp -d)
        trap 'rm -rf "$TEMP"' EXIT
        
        git clone -q -b "$BRANCH" "https://github.com/$REPO.git" "$TEMP"
        cd "$TEMP/mcs-go"
        
        mkdir -p "$MCS_HOME/bin"
        /usr/local/go/bin/go mod download
        /usr/local/go/bin/go build -o "$MCS_HOME/bin/mcs" cmd/mcs/main.go
    else
        echo "Please install Go from https://go.dev"
        exit 1
    fi
fi

# Make executable
chmod +x "$MCS_HOME/bin/mcs"

# Run setup
export PATH="$MCS_HOME/bin:$PATH"
echo ""
echo "🛠️  Running setup..."
exec "$MCS_HOME/bin/mcs" setup --bootstrap