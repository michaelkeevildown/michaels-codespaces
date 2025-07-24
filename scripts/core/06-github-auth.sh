#!/bin/bash

# GitHub authentication setup - Homebrew-style
# Minimal, automatic, non-interactive where possible

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

echo_step "🔐 Configuring GitHub access..."

# Add GitHub's SSH key fingerprint automatically (no prompt)
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Add GitHub to known hosts silently
if ! grep -q "github.com" ~/.ssh/known_hosts 2>/dev/null; then
    echo_debug "Adding GitHub SSH fingerprint..."
    ssh-keyscan -t ed25519 github.com >> ~/.ssh/known_hosts 2>/dev/null
fi

# Check for existing SSH key
SSH_KEY_PATH="$HOME/.ssh/id_ed25519"
if [ ! -f "$SSH_KEY_PATH" ]; then
    echo_debug "Generating SSH key..."
    # Use hostname and username for the key comment
    ssh-keygen -t ed25519 -C "${USER}@$(hostname)" -f "$SSH_KEY_PATH" -N "" -q
fi

# Set up basic git config if not already done
if ! git config --global user.name &>/dev/null; then
    echo_debug "Setting default git config..."
    git config --global user.name "${USER}"
    git config --global user.email "${USER}@$(hostname)"
    git config --global init.defaultBranch main
    git config --global pull.rebase false
fi

# Create directories for auth
mkdir -p ~/codespaces/auth/git-config
mkdir -p ~/codespaces/auth/tokens
mkdir -p ~/codespaces/auth/ssh

# Copy git config for containers
cp ~/.gitconfig ~/codespaces/auth/git-config/.gitconfig 2>/dev/null || true

# Create a marker file to indicate GitHub setup is needed
touch ~/codespaces/.github-setup-pending

echo_success "GitHub access configured"
echo_info "SSH key generated at: ~/.ssh/id_ed25519.pub"