#!/bin/bash

# Shell setup with Zsh and Oh My Zsh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

echo_step "🐚 Setting up Zsh..."

# Check if Oh My Zsh is already installed
if [ -d "$HOME/.oh-my-zsh" ]; then
    echo_info "Oh My Zsh is already installed"
else
    echo_info "Installing Oh My Zsh..."
    sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
fi

# Backup existing .zshrc if it exists
if [ -f ~/.zshrc ] && ! grep -q "CODESPACE BASE ALIASES" ~/.zshrc; then
    cp ~/.zshrc ~/.zshrc.backup
    echo_info "Backed up existing .zshrc to .zshrc.backup"
fi

# Add codespace aliases to .zshrc
echo_info "Adding codespace aliases..."
cat >> ~/.zshrc << 'EOF'

# ============================================================================
# CODESPACE BASE ALIASES
# ============================================================================

# System shortcuts
alias ll='ls -la'
alias la='ls -A'
alias l='ls -CF'
alias ..='cd ..'
alias ...='cd ../..'
alias cls='clear'

# Docker shortcuts
alias dps='docker ps'
alias dpa='docker ps -a' 
alias di='docker images'
alias dprune='docker system prune -f'
alias dlogs='docker logs -f'
alias dexec='docker exec -it'

# Git shortcuts
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git pull'
alias gco='git checkout'
alias gb='git branch'

# System monitoring
alias ports='netstat -tuln'
alias meminfo='free -h'
alias diskinfo='df -h'
alias cpuinfo='lscpu'

# Codespace management
alias codespaces='cd ~/codespaces && ls -la'
alias list-codespaces='~/codespaces/scripts/utils/list-codespaces.sh'
alias monitor='~/monitor-system.sh'

# Quick navigation
alias cds='cd ~/codespaces'
alias shared='cd ~/codespaces/shared'

# Helpful functions
# Quick way to enter a running container
denter() {
    docker exec -it "$1" /bin/bash || docker exec -it "$1" /bin/sh
}

# Show container resource usage
dstats() {
    docker stats --no-stream
}

# Remove all stopped containers
dclean() {
    docker container prune -f
}

EOF

# Set Zsh as default shell if it isn't already
if [ "$SHELL" != "/usr/bin/zsh" ] && [ "$SHELL" != "/bin/zsh" ]; then
    echo_info "Setting Zsh as default shell..."
    sudo chsh -s $(which zsh) $USER
    echo_warning "Shell changed to Zsh. This will take effect on next login."
fi

echo_success "Zsh setup completed"