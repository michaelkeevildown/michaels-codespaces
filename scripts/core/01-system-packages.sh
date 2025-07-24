#!/bin/bash

# System packages installation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"
source "$SCRIPT_DIR/../utils/checks.sh"

echo_step "📦 Installing system packages..."

# Update package list
echo_info "Updating package list..."
echo_status "Downloading package information from Ubuntu repositories..."
sudo apt update -qq
clear_status

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

# Count packages to install
packages_to_install=0
for package in "${PACKAGES[@]}"; do
    if ! dpkg -l | grep -q "^ii  $package "; then
        ((packages_to_install++))
    fi
done

if [ $packages_to_install -eq 0 ]; then
    echo_success "All packages already installed"
else
    echo_info "Installing $packages_to_install packages..."
    
    # Install with progress
    installed=0
    for package in "${PACKAGES[@]}"; do
        if ! dpkg -l | grep -q "^ii  $package "; then
            progress_bar $installed $packages_to_install 40
            printf " Installing %s" "$package"
            
            if sudo apt install -y -qq "$package" > /dev/null 2>&1; then
                ((installed++))
            else
                echo_warning "Failed to install $package, continuing..."
            fi
        fi
    done
    
    # Complete progress bar
    progress_bar $packages_to_install $packages_to_install 40
    printf " Done!          \n"
fi

# Clean up
sudo apt autoremove -y -qq > /dev/null 2>&1

echo_success "System packages installed successfully"