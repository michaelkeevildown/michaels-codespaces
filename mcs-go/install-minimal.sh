#!/bin/bash

# Minimal MCS Bootstrap Installer
# Just downloads and installs the MCS binary, then hands off to Go code

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}MCS Bootstrap Installer${NC}"
echo "======================="
echo ""

# Configuration
MCS_HOME="${MCS_HOME:-$HOME/.mcs}"
BRANCH="${MCS_BRANCH:-main}"
REPO="michaelkeevildown/michaels-codespaces"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) PLATFORM="${OS}-${ARCH}" ;;
    *) echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1 ;;
esac

echo "Platform: $PLATFORM"

# Create bin directory
mkdir -p "$MCS_HOME/bin"

# Download MCS binary
BINARY_URL="https://github.com/$REPO/releases/download/$BRANCH/mcs-$PLATFORM"
echo -e "${BLUE}Downloading MCS...${NC}"

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$BINARY_URL" -o "$MCS_HOME/bin/mcs" || {
        # Fallback: try to download from raw branch
        echo "Release not found, trying branch artifacts..."
        curl -fsSL "https://raw.githubusercontent.com/$REPO/$BRANCH/mcs-go/dist/mcs-$PLATFORM" -o "$MCS_HOME/bin/mcs" || {
            echo -e "${RED}Failed to download MCS binary${NC}"
            echo ""
            echo "You can build from source instead:"
            echo "  git clone https://github.com/$REPO.git"
            echo "  cd michaels-codespaces/mcs-go"
            echo "  go build -o $MCS_HOME/bin/mcs cmd/mcs/main.go"
            exit 1
        }
    }
elif command -v wget >/dev/null 2>&1; then
    wget -q "$BINARY_URL" -O "$MCS_HOME/bin/mcs" || {
        echo -e "${RED}Failed to download MCS binary${NC}"
        exit 1
    }
else
    echo -e "${RED}Neither curl nor wget found. Please install one.${NC}"
    exit 1
fi

# Make executable
chmod +x "$MCS_HOME/bin/mcs"

# Test binary
if ! "$MCS_HOME/bin/mcs" version >/dev/null 2>&1; then
    echo -e "${RED}Downloaded binary doesn't work${NC}"
    echo "This might be due to:"
    echo "  - Wrong architecture/OS"
    echo "  - Missing dependencies"
    echo "  - Corrupted download"
    exit 1
fi

echo -e "${GREEN}✅ MCS binary installed${NC}"

# Add to PATH if needed
if [[ ":$PATH:" != *":$MCS_HOME/bin:"* ]]; then
    echo ""
    echo "Add to PATH by running:"
    echo "  export PATH=\"$MCS_HOME/bin:\$PATH\""
    echo ""
    echo "Or add to your shell config:"
    echo "  echo 'export PATH=\"$MCS_HOME/bin:\$PATH\"' >> ~/.bashrc"
fi

# Hand off to MCS for the rest
echo ""
echo -e "${BLUE}Running MCS setup...${NC}"
echo ""

# Export PATH for this session
export PATH="$MCS_HOME/bin:$PATH"

# Run MCS setup
exec "$MCS_HOME/bin/mcs" setup --bootstrap