#!/bin/bash

# System packages installation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"
source "$SCRIPT_DIR/../utils/checks.sh"

echo_step "📦 Installing system packages..."

# Update package list
echo_info "Updating package list..."
sudo apt update -qq

# Core system packages
PACKAGES=(
    # Essential tools
    curl wget git unzip zip
    build-essential software-properties-common
    apt-transport-https ca-certificates gnupg lsb-release
    
    # System monitoring
    htop ncdu tree jq net-tools
    
    # Development tools
    vim nano zsh
    python3 python3-pip
    nodejs npm
    
    # Required for various operations
    bc  # For version comparisons
)

echo_info "Installing packages..."
for package in "${PACKAGES[@]}"; do
    if ! dpkg -l | grep -q "^ii  $package "; then
        echo_debug "Installing $package..."
        sudo apt install -y -qq "$package" > /dev/null 2>&1 || {
            echo_warning "Failed to install $package, continuing..."
        }
    else
        echo_debug "$package already installed"
    fi
done

# Clean up
sudo apt autoremove -y -qq > /dev/null 2>&1

echo_success "System packages installed successfully"