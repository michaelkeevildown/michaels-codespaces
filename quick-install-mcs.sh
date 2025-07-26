#!/bin/bash

# Quick MCS Go installer for testing
# Works with public repos, supports private repos with GITHUB_TOKEN

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Quick MCS Go Installer${NC}"
echo "=========================="
echo ""

# Configuration
BRANCH="${MCS_BRANCH:-feat/mcs-go-status-command}"
REPO_URL="https://github.com/michaelkeevildown/michaels-codespaces.git"
MCS_HOME="$HOME/.mcs"

echo "Repository: $REPO_URL"
echo "Branch: $BRANCH"
echo ""

# Install Git if needed
if ! command -v git &> /dev/null; then
    echo -e "${YELLOW}Installing Git...${NC}"
    sudo apt-get update -qq
    sudo apt-get install -y -qq git
fi

# Install Go if needed
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Installing Go 1.21...${NC}"
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    rm go1.21.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
fi

# Clone repository
echo -e "${BLUE}Cloning MCS repository...${NC}"
if [ -d "$MCS_HOME" ]; then
    echo "Backing up existing installation..."
    mv "$MCS_HOME" "${MCS_HOME}.backup.$(date +%Y%m%d_%H%M%S)"
fi

# Try cloning
if [ -n "${GITHUB_TOKEN:-}" ]; then
    echo "Using GitHub token for authentication..."
    git clone -b "$BRANCH" "https://token:${GITHUB_TOKEN}@github.com/michaelkeevildown/michaels-codespaces.git" "$MCS_HOME" 2>/dev/null || {
        echo -e "${RED}Failed to clone with token${NC}"
        exit 1
    }
    # Remove token from URL
    cd "$MCS_HOME"
    git remote set-url origin "$REPO_URL"
else
    # Try public clone
    git clone -b "$BRANCH" "$REPO_URL" "$MCS_HOME" || {
        echo -e "${RED}Failed to clone repository${NC}"
        echo ""
        echo "If this is a private repository, run:"
        echo "  export GITHUB_TOKEN='your-token'"
        echo "  curl -fsSL <this-script-url> | bash"
        exit 1
    }
fi

# Build MCS
echo -e "${BLUE}Building MCS...${NC}"
cd "$MCS_HOME/mcs-go"
go mod download
mkdir -p "$MCS_HOME/bin"
go build -o "$MCS_HOME/bin/mcs" cmd/mcs/main.go

# Add to PATH
export PATH="$MCS_HOME/bin:$PATH"

# Test
echo ""
if "$MCS_HOME/bin/mcs" version; then
    echo -e "${GREEN}✅ MCS installed successfully!${NC}"
else
    echo -e "${RED}❌ Installation failed${NC}"
    exit 1
fi

echo ""
echo "Add to PATH:"
echo '  echo "export PATH=\$HOME/.mcs/bin:\$PATH" >> ~/.bashrc'
echo '  source ~/.bashrc'
echo ""
echo "Test with: mcs status"