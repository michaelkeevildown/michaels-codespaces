#!/bin/bash

# Complete MCS Go installer for Ubuntu with all dependencies
# Usage: curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/feat/mcs-go-status-command/install-mcs-ubuntu.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
echo_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

echo_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

echo_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

echo_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo ""
echo -e "${BLUE}🚀 Michael's Codespaces (MCS) - Ubuntu Installer${NC}"
echo "================================================"
echo ""

# Detect Ubuntu version
if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo_info "Detected: $NAME $VERSION"
else
    echo_error "Cannot detect OS version. This script is designed for Ubuntu."
    exit 1
fi

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
   echo_warning "Running as root. MCS will be installed system-wide."
else
   echo_info "Running as user: $USER"
fi

# Update package list
echo ""
echo_info "Updating package list..."
sudo apt-get update -qq

# Install basic dependencies
echo_info "Installing basic dependencies..."
sudo apt-get install -y -qq \
    curl \
    wget \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    build-essential

# Install Git if not present
if ! command -v git &> /dev/null; then
    echo_info "Installing Git..."
    sudo apt-get install -y -qq git
    echo_success "Git installed"
else
    echo_success "Git already installed: $(git --version)"
fi

# Install Docker if not present
if ! command -v docker &> /dev/null; then
    echo ""
    echo_info "Installing Docker..."
    
    # Remove old versions
    sudo apt-get remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true
    
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
    
    # Start and enable Docker
    sudo systemctl start docker
    sudo systemctl enable docker
    
    # Add current user to docker group (if not root)
    if [ "$EUID" -ne 0 ]; then
        sudo usermod -aG docker $USER
        echo_warning "Added $USER to docker group. You may need to log out and back in for this to take effect."
    fi
    
    echo_success "Docker installed successfully"
else
    echo_success "Docker already installed: $(docker --version)"
fi

# Test Docker
if sudo docker run --rm hello-world &>/dev/null; then
    echo_success "Docker is working correctly"
else
    echo_error "Docker test failed. Please check Docker installation."
    exit 1
fi

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo ""
    echo_info "Installing Go 1.21..."
    
    # Detect architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            GO_ARCH="amd64"
            ;;
        aarch64)
            GO_ARCH="arm64"
            ;;
        *)
            echo_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    # Download and install Go
    GO_VERSION="1.21.5"
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    rm "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    
    echo_success "Go ${GO_VERSION} installed"
else
    echo_success "Go already installed: $(go version)"
fi

# Clone MCS repository from feature branch
echo ""
echo_info "Downloading MCS from feature branch..."
MCS_TEMP="/tmp/mcs-install-$$"
rm -rf "$MCS_TEMP"

# Set branch for the installer
export MCS_BRANCH="feat/mcs-go-status-command"

# If GITHUB_TOKEN is set, pass it through
if [ -n "${GITHUB_TOKEN:-}" ]; then
    echo_info "GitHub token detected, will use for authentication"
fi

# Clone the repository
if ! git clone -b "$MCS_BRANCH" https://github.com/michaelkeevildown/michaels-codespaces.git "$MCS_TEMP" 2>/dev/null; then
    echo_error "Failed to clone repository"
    echo_warning "If this is a private repository, please provide a GitHub token:"
    echo ""
    echo "  export GITHUB_TOKEN='your-github-token'"
    echo "  curl -fsSL <install-script-url> | bash"
    echo ""
    exit 1
fi

# Navigate to mcs-go directory
cd "$MCS_TEMP/mcs-go"

# Run the MCS installer (it will use the GITHUB_TOKEN and MCS_BRANCH env vars)
echo ""
echo_info "Installing MCS..."
bash install.sh

# Add to PATH for current session
export PATH="$HOME/.mcs/bin:$PATH"

# Verify installation
echo ""
echo_info "Verifying installation..."
if command -v mcs &> /dev/null; then
    echo_success "MCS installed successfully!"
    echo ""
    mcs version
else
    echo_error "MCS installation failed"
    exit 1
fi

# Clean up
rm -rf "$MCS_TEMP"

# Test status command
echo ""
echo_info "Testing status command..."
echo ""
mcs status

# Add PATH instruction
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo_success "Installation complete!"
echo ""
echo "📝 Add MCS to your PATH permanently:"
echo ""
echo '    echo "export PATH=\$HOME/.mcs/bin:\$PATH" >> ~/.bashrc'
echo '    source ~/.bashrc'
echo ""
echo "🐳 If you were added to the docker group, log out and back in."
echo ""
echo "🎯 Quick commands to try:"
echo "    mcs version      - Show version"
echo "    mcs status       - Show system status"
echo "    mcs doctor       - Check system health"
echo "    mcs create test  - Create a test codespace"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"