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
    echo "⚠️  Docker is not installed. Installing Docker..."
    
    # Install Docker on Ubuntu
    sudo apt-get update -qq
    sudo apt-get install -y -qq ca-certificates curl gnupg lsb-release
    
    # Add Docker's official GPG key
    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    
    # Set up the repository
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # Install Docker Engine
    sudo apt-get update -qq
    sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    # Add current user to docker group
    sudo usermod -aG docker $USER
    
    # Start Docker
    sudo systemctl start docker
    sudo systemctl enable docker
    
    echo "✅ Docker installed successfully"
    echo "⚠️  Note: You may need to log out and back in for docker group permissions to take effect"
    
    # For this session, use sudo with docker commands
    alias docker='sudo docker'
fi
echo "✅ Docker found"

# Check Git
if ! command -v git &> /dev/null; then
    echo "⚠️  Git is not installed. Installing Git..."
    sudo apt-get install -y -qq git
    echo "✅ Git installed"
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

# Set branch for the installer
export MCS_BRANCH="feat/mcs-go-status-command"

# If GITHUB_TOKEN is set, pass it through
if [ -n "${GITHUB_TOKEN:-}" ]; then
    echo "✅ GitHub token detected, will use for authentication"
fi

# Clone the repository
if ! git clone -b "$MCS_BRANCH" https://github.com/michaelkeevildown/michaels-codespaces.git mcs-test 2>/dev/null; then
    echo "❌ Failed to clone repository"
    echo "⚠️  If this is a private repository, please provide a GitHub token:"
    echo ""
    echo "  export GITHUB_TOKEN='your-github-token'"
    echo "  curl -fsSL <install-script-url> | bash"
    echo ""
    exit 1
fi

# Navigate to mcs-go directory
cd mcs-test/mcs-go

# Run the install script (it will use the GITHUB_TOKEN and MCS_BRANCH env vars)
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