#!/bin/bash

# Test script for installing MCS Go from feature branch
# Usage: curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/feat/mcs-go-status-command/test-mcs-go-branch.sh | bash

set -e

echo "🚀 Testing MCS Go installation from feat/mcs-go-status-command branch"
echo ""

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi
echo "✅ Docker found"

# Check Git
if ! command -v git &> /dev/null; then
    echo "❌ Git is not installed. Please install Git first."
    exit 1
fi
echo "✅ Git found"

# Check Go (for building from source)
if ! command -v go &> /dev/null; then
    echo "⚠️  Go is not installed. Installing Go 1.21..."
    
    # Install Go on Ubuntu
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    rm go1.21.5.linux-amd64.tar.gz
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    
    echo "✅ Go 1.21 installed"
fi

go_version=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $go_version"

# Clone repository from feature branch
echo ""
echo "📥 Cloning MCS from feature branch..."
cd /tmp
rm -rf mcs-test
git clone -b feat/mcs-go-status-command https://github.com/michaelkeevildown/michaels-codespaces.git mcs-test

# Navigate to mcs-go directory
cd mcs-test/mcs-go

# Run the install script
echo ""
echo "🔨 Running MCS Go installer..."
bash install.sh

# Test the installation
echo ""
echo "🧪 Testing MCS installation..."
echo ""

# Add to PATH for this session
export PATH="$HOME/.mcs/bin:$PATH"

# Show version
if mcs version; then
    echo "✅ MCS installed successfully!"
else
    echo "❌ MCS installation failed"
    exit 1
fi

echo ""
echo "📊 Testing status command..."
mcs status

echo ""
echo "🎯 Quick test commands:"
echo "  mcs version      - Show version"
echo "  mcs status       - Show system status"
echo "  mcs doctor       - Check system health"
echo "  mcs list         - List codespaces"
echo "  mcs create test  - Create a test codespace"
echo ""
echo "📝 Add to PATH permanently:"
echo '  echo "export PATH=\$HOME/.mcs/bin:\$PATH" >> ~/.bashrc'
echo '  source ~/.bashrc'
echo ""