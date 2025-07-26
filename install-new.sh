#!/bin/sh
# Minimal MCS installer - downloads binary and runs setup
# Usage: curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/install.sh | sh

set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

PLATFORM="${OS}-${ARCH}"
MCS_HOME="${MCS_HOME:-$HOME/.mcs}"
REPO="${MCS_REPO:-michaelkeevildown/michaels-codespaces}"
BRANCH="${MCS_BRANCH:-main}"

echo "Installing MCS for $PLATFORM..."

# Create directory
mkdir -p "$MCS_HOME/bin"

# Try to download pre-built binary first
BINARY_URL="https://github.com/$REPO/releases/latest/download/mcs-$PLATFORM"
if curl -fsSL "$BINARY_URL" -o "$MCS_HOME/bin/mcs" 2>/dev/null || \
   wget -q "$BINARY_URL" -O "$MCS_HOME/bin/mcs" 2>/dev/null; then
    echo "Downloaded pre-built binary"
else
    # Fallback: build from source if Go is available
    if command -v go >/dev/null 2>&1; then
        echo "Building from source..."
        TEMP_DIR=$(mktemp -d)
        trap 'rm -rf "$TEMP_DIR"' EXIT
        
        git clone -b "$BRANCH" "https://github.com/$REPO.git" "$TEMP_DIR" || exit 1
        cd "$TEMP_DIR/mcs-go"
        go build -o "$MCS_HOME/bin/mcs" cmd/mcs/main.go || exit 1
        echo "Built from source"
    else
        echo "No pre-built binary available and Go not installed" >&2
        echo "Install Go from https://go.dev or use a platform-specific installer" >&2
        exit 1
    fi
fi

# Make executable
chmod +x "$MCS_HOME/bin/mcs"

# Add to PATH for this session
export PATH="$MCS_HOME/bin:$PATH"

# Run setup
echo ""
exec "$MCS_HOME/bin/mcs" setup --bootstrap