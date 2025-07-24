#!/bin/bash

# Docker installation and configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"
source "$SCRIPT_DIR/../utils/checks.sh"

echo_step "🐳 Setting up Docker..."

# Check if Docker is already installed
if command -v docker &> /dev/null; then
    echo_info "Docker is already installed ($(docker --version))"
    echo_info "Ensuring Docker is properly configured..."
else
    echo_info "Installing Docker..."
    
    # Remove any old Docker installations
    sudo apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true
    
    # Install Docker using official script
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sudo sh /tmp/get-docker.sh > /dev/null 2>&1
    rm /tmp/get-docker.sh
fi

# Add user to docker group
if ! groups $USER | grep -q docker; then
    echo_info "Adding $USER to docker group..."
    sudo usermod -aG docker $USER
    echo_warning "You'll need to logout and login again for Docker permissions"
fi

# Configure Docker daemon
echo_info "Configuring Docker daemon..."
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json > /dev/null << 'EOF'
{
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "10m",
        "max-file": "3"
    },
    "default-address-pools": [
        {
            "base": "172.20.0.0/16",
            "size": 24
        }
    ],
    "features": {
        "buildkit": true
    }
}
EOF

# Install docker-compose
if ! command -v docker-compose &> /dev/null; then
    echo_info "Installing docker-compose..."
    COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep -Po '"tag_name": "\K.*?(?=")')
    sudo curl -fsSL "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" \
        -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
else
    echo_info "docker-compose is already installed ($(docker-compose --version))"
fi

# Restart Docker to apply configuration
echo_info "Restarting Docker service..."
sudo systemctl restart docker
sudo systemctl enable docker

# Verify Docker is working
if sudo docker run --rm hello-world &> /dev/null; then
    echo_success "Docker installed and configured successfully"
else
    echo_error "Docker installation verification failed"
    exit 1
fi