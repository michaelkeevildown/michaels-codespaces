#!/bin/bash

# Michael's Codespaces Installation Script
# Properly sets up paths and symlinks for MCS

set -e

# Colors
if [[ -t 1 ]]; then
    tty_red=$(printf '\033[31m')
    tty_green=$(printf '\033[32m')
    tty_yellow=$(printf '\033[33m')
    tty_blue=$(printf '\033[34m')
    tty_bold=$(printf '\033[1m')
    tty_reset=$(printf '\033[0m')
else
    tty_red='' tty_green='' tty_yellow='' tty_blue=''
    tty_bold='' tty_reset=''
fi

# Helper functions
info() {
    printf "${tty_blue}==>${tty_reset} ${tty_bold}%s${tty_reset}\n" "$1"
}

success() {
    printf "${tty_green}✓${tty_reset} %s\n" "$1"
}

warning() {
    printf "${tty_yellow}⚠${tty_reset}  %s\n" "$1"
}

error() {
    printf "${tty_red}✗${tty_reset} %s\n" "$1" >&2
}

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$HOME/.mcs"
CODESPACES_DIR="$HOME/codespaces"

info "Installing Michael's Codespaces..."

# Check if running from the codebase directory
if [ ! -f "$SCRIPT_DIR/bin/mcs" ]; then
    error "This script must be run from the Michael's Codespaces directory"
    exit 1
fi

# Create codespaces directory if it doesn't exist
if [ ! -d "$CODESPACES_DIR" ]; then
    info "Creating codespaces directory..."
    mkdir -p "$CODESPACES_DIR"
    success "Created $CODESPACES_DIR"
fi

# Handle installation directory
if [ -d "$INSTALL_DIR" ] && [ ! -L "$INSTALL_DIR" ]; then
    warning "Installation directory already exists: $INSTALL_DIR"
    read -p "Do you want to replace it? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        info "Backing up existing installation..."
        mv "$INSTALL_DIR" "${INSTALL_DIR}.backup.$(date +%Y%m%d_%H%M%S)"
        success "Backed up existing installation"
    else
        error "Installation cancelled"
        exit 1
    fi
fi

# Create symlink to installation directory
if [ "$SCRIPT_DIR" != "$INSTALL_DIR" ]; then
    info "Creating installation symlink..."
    ln -sfn "$SCRIPT_DIR" "$INSTALL_DIR"
    success "Created symlink: $INSTALL_DIR -> $SCRIPT_DIR"
else
    info "Already installed in the correct location"
fi

# Ensure bin directory is in PATH
info "Setting up PATH..."
BIN_PATH="$INSTALL_DIR/bin"

# Add to .zshrc if using zsh
if [ -f ~/.zshrc ]; then
    if ! grep -q "$BIN_PATH" ~/.zshrc; then
        echo "" >> ~/.zshrc
        echo "# Michael's Codespaces" >> ~/.zshrc
        echo "export PATH=\"$BIN_PATH:\$PATH\"" >> ~/.zshrc
        success "Added MCS to PATH in ~/.zshrc"
    else
        info "MCS already in PATH in ~/.zshrc"
    fi
fi

# Add to .bashrc if using bash
if [ -f ~/.bashrc ]; then
    if ! grep -q "$BIN_PATH" ~/.bashrc; then
        echo "" >> ~/.bashrc
        echo "# Michael's Codespaces" >> ~/.bashrc
        echo "export PATH=\"$BIN_PATH:\$PATH\"" >> ~/.bashrc
        success "Added MCS to PATH in ~/.bashrc"
    else
        info "MCS already in PATH in ~/.bashrc"
    fi
fi

# Make all scripts executable
info "Making scripts executable..."
chmod +x "$SCRIPT_DIR/bin/mcs"
chmod +x "$SCRIPT_DIR/scripts/core/"*.sh
chmod +x "$SCRIPT_DIR/scripts/modules/components/"*.sh
chmod +x "$SCRIPT_DIR/scripts/modules/components/installers/"*.sh
chmod +x "$SCRIPT_DIR/scripts/templates/"*.sh
find "$SCRIPT_DIR/scripts" -name "*.sh" -type f -exec chmod +x {} \;
success "Scripts are now executable"

# Create necessary directories
info "Creating required directories..."
mkdir -p "$CODESPACES_DIR/auth/tokens"
mkdir -p "$CODESPACES_DIR/shared"
mkdir -p "$CODESPACES_DIR/backups"
success "Created required directories"

# Set up environment variable
info "Setting up environment variables..."
echo "export CODESPACE_HOME=\"$INSTALL_DIR\"" > "$INSTALL_DIR/.env"
success "Created environment configuration"

# Verify installation
info "Verifying installation..."
if [ -L "$INSTALL_DIR" ] && [ -d "$INSTALL_DIR" ]; then
    success "Installation directory verified"
else
    error "Installation directory verification failed"
    exit 1
fi

if [ -x "$BIN_PATH/mcs" ]; then
    success "MCS command is executable"
else
    error "MCS command is not executable"
    exit 1
fi

# Final instructions
echo ""
success "Michael's Codespaces installed successfully!"
echo ""
echo "Installation Details:"
echo "  • Installation directory: $INSTALL_DIR"
echo "  • Codespaces directory: $CODESPACES_DIR"
echo "  • Command location: $BIN_PATH/mcs"
echo ""
echo "Next steps:"
echo "  1. Reload your shell: source ~/.zshrc (or source ~/.bashrc)"
echo "  2. Verify installation: mcs doctor"
echo "  3. Create your first codespace: mcs create <repo-url>"
echo ""

# Offer to reload shell
read -p "Would you like to reload your shell now? [Y/n] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    if [ -n "$ZSH_VERSION" ]; then
        source ~/.zshrc
        success "Shell reloaded"
    elif [ -n "$BASH_VERSION" ]; then
        source ~/.bashrc
        success "Shell reloaded"
    fi
    
    # Test the command
    if command -v mcs >/dev/null 2>&1; then
        success "MCS command is available!"
        echo ""
        echo "You can now run: mcs help"
    fi
fi