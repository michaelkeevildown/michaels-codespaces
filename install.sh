#!/bin/bash

# Michael's Codespaces Installer
# This script is designed to be run directly via curl, like Homebrew
# Usage: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/install.sh)"

set -e

# Configuration
CODESPACE_HOME="${CODESPACE_HOME:-$HOME/.mcs}"
CODESPACE_REPO="https://github.com/michaelkeevildown/michaels-codespaces.git"
CODESPACE_SCRIPTS="$HOME/codespaces"
CODESPACE_BRANCH="${CODESPACE_BRANCH:-main}"

# Colors for output (Homebrew style)
if [[ -t 1 ]]; then
    tty_red=$(printf '\033[31m')
    tty_green=$(printf '\033[32m')
    tty_yellow=$(printf '\033[33m')
    tty_blue=$(printf '\033[34m')
    tty_bold=$(printf '\033[1m')
    tty_reset=$(printf '\033[0m')
else
    tty_red=''
    tty_green=''
    tty_yellow=''
    tty_blue=''
    tty_bold=''
    tty_reset=''
fi

# Export for child scripts
export tty_red tty_green tty_yellow tty_blue tty_bold tty_reset

# Helper functions
info() {
    printf "${tty_blue}==>${tty_bold} %s${tty_reset}\n" "$1"
}

success() {
    printf "${tty_green}==>${tty_bold} %s${tty_reset}\n" "$1"
}

warning() {
    printf "${tty_yellow}Warning${tty_reset}: %s\n" "$1"
}

error() {
    printf "${tty_red}Error${tty_reset}: %s\n" "$1" >&2
}

# Fail fast with a concise message
abort() {
    error "$1"
    exit 1
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main installation
main() {
    # Ensure we're in a valid directory
    cd "$HOME" 2>/dev/null || cd /tmp || abort "Cannot find a valid working directory"
    
    # Banner with box drawing
    printf "\n${tty_bold}┌────────────────────────────────────┐${tty_reset}\n"
    printf "${tty_bold}│  Michael's Codespaces Installer    │${tty_reset}\n"
    printf "${tty_bold}└────────────────────────────────────┘${tty_reset}\n\n"

    # Checks with spinner
    printf "${tty_blue}==>${tty_reset} ${tty_bold}Checking system requirements${tty_reset}\n"
    
    if [[ "$EUID" -eq 0 ]]; then
        abort "Don't run this installer as root."
    fi
    
    if ! command_exists lsb_release; then
        abort "This installer requires Ubuntu. Please run on Ubuntu 20.04 or later."
    fi
    
    UBUNTU_VERSION=$(lsb_release -rs)
    if (( $(echo "$UBUNTU_VERSION < 20.04" | bc -l) )); then
        abort "Ubuntu 20.04 or later required. Found: $UBUNTU_VERSION"
    fi
    
    # Check for required commands
    for cmd in git curl sudo; do
        if ! command_exists "$cmd"; then
            abort "Required command not found: $cmd"
        fi
    done
    
    success "System requirements OK"
    printf "  Platform: Ubuntu %s\n" "$UBUNTU_VERSION"
    printf "  User: %s\n" "$USER"
    printf "  Home: %s\n" "$HOME"
    printf "\n"
    
    # Check if already installed
    if [[ -d "$CODESPACE_HOME" ]]; then
        warning "Michael's Codespaces is already installed at $CODESPACE_HOME"
        read -p "Reinstall? This will update to the latest version [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            abort "Installation cancelled."
        fi
        info "Removing existing installation..."
        rm -rf "$CODESPACE_HOME"
    fi
    
    # Clone repository
    info "Installing Michael's Codespaces to $CODESPACE_HOME..."
    if [[ "$CODESPACE_BRANCH" != "main" ]]; then
        info "Using branch: $CODESPACE_BRANCH"
        git clone --depth=1 --branch "$CODESPACE_BRANCH" "$CODESPACE_REPO" "$CODESPACE_HOME" || abort "Failed to clone repository"
    else
        git clone --depth=1 "$CODESPACE_REPO" "$CODESPACE_HOME" || abort "Failed to clone repository"
    fi
    
    # Run setup from the cloned repository
    info "Running setup scripts..."
    cd "$CODESPACE_HOME"
    
    # Make scripts executable
    find scripts -name "*.sh" -exec chmod +x {} \;
    chmod +x bin/mcs
    
    # Run main setup - Homebrew style
    if [ -f ./scripts/core/main-setup-homebrew.sh ]; then
        ./scripts/core/main-setup-homebrew.sh || abort "Setup failed"
    elif [ -f ./scripts/core/main-setup.sh ]; then
        ./scripts/core/main-setup.sh || abort "Setup failed"
    else
        abort "No setup script found"
    fi
    
    # Install key scripts to user's home
    info "Installing codespace tools..."
    
    # Add to PATH (for zsh)
    if [[ -f "$HOME/.zshrc" ]] && ! grep -q ".mcs/bin" "$HOME/.zshrc"; then
        echo '' >> "$HOME/.zshrc"
        echo "# Michael's Codespaces" >> "$HOME/.zshrc"
        echo "export PATH=\"\$HOME/.mcs/bin:\$PATH\"" >> "$HOME/.zshrc"
    fi
    
    # Add to PATH (for bash)
    if [[ -f "$HOME/.bashrc" ]] && ! grep -q ".mcs/bin" "$HOME/.bashrc"; then
        echo '' >> "$HOME/.bashrc"
        echo "# Michael's Codespaces" >> "$HOME/.bashrc"
        echo "export PATH=\"\$HOME/.mcs/bin:\$PATH\"" >> "$HOME/.bashrc"
    fi
    
    # Create bin directory and install mcs command
    mkdir -p "$CODESPACE_HOME/bin"
    
    # Install mcs command
    if [ -f "$CODESPACE_HOME/bin/mcs" ]; then
        info "Installing mcs command..."
        sudo ln -sf "$CODESPACE_HOME/bin/mcs" /usr/local/bin/mcs
        
        # Install completions
        if [ -d /etc/bash_completion.d ] && [ -f "$CODESPACE_HOME/bin/mcs-completion.bash" ]; then
            sudo cp "$CODESPACE_HOME/bin/mcs-completion.bash" /etc/bash_completion.d/mcs
        fi
        
        if [ -d /usr/share/zsh/site-functions ] && [ -f "$CODESPACE_HOME/bin/mcs-completion.zsh" ]; then
            sudo cp "$CODESPACE_HOME/bin/mcs-completion.zsh" /usr/share/zsh/site-functions/_mcs
        fi
        
        success "mcs command installed with completions"
    fi
    
    # Configure IP address settings
    info "Configuring IP address settings..."
    if [ -f "$CODESPACE_HOME/scripts/modules/network/network-utils.sh" ] && \
       [ -f "$CODESPACE_HOME/scripts/modules/storage/config-manager.sh" ]; then
        source "$CODESPACE_HOME/scripts/modules/network/network-utils.sh"
        source "$CODESPACE_HOME/scripts/modules/storage/config-manager.sh"
        
        # Initialize config
        init_config
        
        # Detect and show available IPs
        printf "\n"
        info "Detected network configuration:"
        local local_ip=$(detect_local_ip)
        printf "  Local IP: %s\n" "$local_ip"
        
        # Ask user for preference
        printf "\n"
        printf "How would you like to access your codespaces?\n"
        printf "  1) localhost (default - local access only)\n"
        printf "  2) %s (LAN access)\n" "$local_ip"
        printf "  3) Auto-detect public IP (internet access)\n"
        printf "  4) Enter custom IP/hostname\n"
        printf "\n"
        
        if [ -z "$NONINTERACTIVE" ]; then
            read -p "Choice [1-4]: " ip_choice
            
            case "$ip_choice" in
                2)
                    set_ip_mode "auto"
                    set_host_ip "$local_ip"
                    success "Configured to use local IP: $local_ip"
                    ;;
                3)
                    info "Detecting public IP..."
                    local public_ip=$(detect_public_ip)
                    if [ -n "$public_ip" ]; then
                        set_ip_mode "public"
                        set_host_ip "$public_ip"
                        success "Configured to use public IP: $public_ip"
                    else
                        warning "Could not detect public IP, using local IP"
                        set_ip_mode "auto"
                        set_host_ip "$local_ip"
                    fi
                    ;;
                4)
                    read -p "Enter IP address or hostname: " custom_ip
                    if validate_ip_address "$custom_ip"; then
                        set_ip_mode "manual"
                        set_host_ip "$custom_ip"
                        success "Configured to use: $custom_ip"
                    else
                        warning "Invalid IP/hostname, using localhost"
                        set_ip_mode "localhost"
                        set_host_ip "localhost"
                    fi
                    ;;
                *)
                    set_ip_mode "localhost"
                    set_host_ip "localhost"
                    info "Using default: localhost"
                    ;;
            esac
        else
            # Non-interactive mode - use localhost
            set_ip_mode "localhost"
            set_host_ip "localhost"
            info "Non-interactive mode: using localhost"
        fi
        
        printf "\n"
        info "You can change this later with: mcs update-ip"
    fi
    
    # Success message
    printf "\n"
    success "🎉 Michael's Codespaces installed successfully!"
    printf "\n"
    printf "Installation location: %s\n" "$CODESPACE_HOME"
    printf "Codespaces directory: %s\n" "$CODESPACE_SCRIPTS"
    printf "\n"
    printf "%s\n" "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "\n"
    printf "${tty_yellow}Next steps:${tty_reset}\n"
    printf "\n"
    printf "1. ${tty_bold}Logout and login again${tty_reset} for Docker permissions:\n"
    printf "   $ exit\n"
    printf "   $ ssh %s@%s\n" "$USER" "$(hostname -I | awk '{print $1}')"
    printf "\n"
    printf "2. ${tty_bold}Create your first codespace${tty_reset}:\n"
    printf "   $ mcs create git@github.com:user/repo.git\n"
    printf "\n"
    printf "For help: ${tty_bold}mcs help${tty_reset}\n"
    printf "Check health: ${tty_bold}mcs doctor${tty_reset}\n"
    printf "\n"
    printf "%s\n" "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "\n"
}

# Run main installation
main "$@"