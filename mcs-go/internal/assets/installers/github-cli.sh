#!/bin/bash

# GitHub CLI Component Installer
# Installs and configures GitHub CLI in containers

set -e

# Component metadata
metadata() {
    echo "name=GitHub CLI"
    echo "version=latest"
    echo "description=GitHub command-line interface with authentication"
}

# Component dependencies
dependencies() {
    # No dependencies
    echo ""
}

# Installation function
install() {
    echo "Installing GitHub CLI..."
    
    # Detect OS and architecture
    local os="linux"
    local arch=$(uname -m)
    
    case "$arch" in
        x86_64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            echo "Unsupported architecture: $arch" >&2
            return 1
            ;;
    esac
    
    # Install based on available package manager
    if command -v apt-get >/dev/null 2>&1; then
        install_debian
    elif command -v yum >/dev/null 2>&1; then
        install_rhel
    elif command -v apk >/dev/null 2>&1; then
        install_alpine
    else
        install_binary
    fi
}

# Install on Debian/Ubuntu
install_debian() {
    echo "Installing GitHub CLI via apt..."
    
    # Install dependencies
    apt-get update -qq
    apt-get install -y -qq curl gpg sudo
    
    # Add GitHub CLI repository
    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
    
    # Install gh
    apt-get update -qq
    apt-get install -y -qq gh
}

# Install on RHEL/CentOS/Fedora
install_rhel() {
    echo "Installing GitHub CLI via yum..."
    
    # Add repository
    yum-config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo
    
    # Install gh
    yum install -y gh
}

# Install on Alpine
install_alpine() {
    echo "Installing GitHub CLI on Alpine..."
    
    # GitHub CLI is not in Alpine repos, use binary installation
    install_binary
}

# Binary installation (fallback)
install_binary() {
    echo "Installing GitHub CLI from binary..."
    
    local arch=$(uname -m)
    case "$arch" in
        x86_64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
    esac
    
    # Get latest version
    local version=$(curl -s https://api.github.com/repos/cli/cli/releases/latest | grep -Po '"tag_name": "v\K[^"]*')
    
    if [ -z "$version" ]; then
        echo "Failed to get latest GitHub CLI version" >&2
        return 1
    fi
    
    # Download and install
    local url="https://github.com/cli/cli/releases/download/v${version}/gh_${version}_linux_${arch}.tar.gz"
    
    curl -fsSL "$url" | tar -xz -C /tmp
    sudo mv "/tmp/gh_${version}_linux_${arch}/bin/gh" /usr/local/bin/
    sudo chmod +x /usr/local/bin/gh
    
    # Cleanup
    rm -rf "/tmp/gh_${version}_linux_${arch}"
}

# Configuration function
configure() {
    echo "Configuring GitHub CLI..."
    
    # Check if token is available
    local token_file="/home/coder/.tokens/github.token"
    
    if [ -f "$token_file" ] && [ -s "$token_file" ]; then
        local token=$(cat "$token_file")
        
        # Configure gh with token
        echo "$token" | gh auth login --with-token
        
        # Verify authentication
        if gh auth status >/dev/null 2>&1; then
            echo "GitHub CLI authenticated successfully"
            
            # Get username for git config
            local username=$(gh api user --jq .login 2>/dev/null || echo "")
            
            if [ -n "$username" ]; then
                # Configure git
                git config --global user.name "$username"
                git config --global user.email "${username}@users.noreply.github.com"
                git config --global init.defaultBranch main
                
                echo "Git configured for user: $username"
            fi
        else
            echo "GitHub CLI authentication failed" >&2
            return 1
        fi
    else
        echo "GitHub token not found at $token_file"
        echo "GitHub CLI installed but not authenticated"
        echo "To authenticate manually, run: gh auth login"
    fi
    
    # Set up git credential helper to use gh
    git config --global credential.helper "!gh auth git-credential"
    
    # Configure gh settings
    gh config set git_protocol https
    gh config set prompt enabled
}

# Verification function
verify() {
    echo "Verifying GitHub CLI installation..."
    
    # Check if gh is installed
    if ! command -v gh >/dev/null 2>&1; then
        echo "GitHub CLI not found in PATH" >&2
        return 1
    fi
    
    # Check version
    local version=$(gh --version | head -n1)
    echo "GitHub CLI installed: $version"
    
    # Check authentication status
    if gh auth status >/dev/null 2>&1; then
        echo "GitHub CLI is authenticated"
        gh auth status
    else
        echo "GitHub CLI is not authenticated"
    fi
    
    # Check git configuration
    local git_user=$(git config --global user.name 2>/dev/null || echo "")
    local git_email=$(git config --global user.email 2>/dev/null || echo "")
    
    if [ -n "$git_user" ] && [ -n "$git_email" ]; then
        echo "Git is configured: $git_user <$git_email>"
    else
        echo "Git user configuration is incomplete"
    fi
    
    return 0
}

# Uninstall function
uninstall() {
    echo "Uninstalling GitHub CLI..."
    
    # Remove based on installation method
    if command -v apt-get >/dev/null 2>&1 && dpkg -l gh >/dev/null 2>&1; then
        apt-get remove -y gh
    elif command -v yum >/dev/null 2>&1 && rpm -q gh >/dev/null 2>&1; then
        yum remove -y gh
    elif [ -f /usr/local/bin/gh ]; then
        sudo rm -f /usr/local/bin/gh
    fi
    
    # Remove auth config
    rm -rf ~/.config/gh
    
    echo "GitHub CLI uninstalled"
}

# Main function
main() {
    local action="${1:-install}"
    
    case "$action" in
        metadata)
            metadata
            ;;
        dependencies)
            dependencies
            ;;
        install)
            install
            ;;
        configure)
            configure
            ;;
        verify)
            verify
            ;;
        uninstall)
            uninstall
            ;;
        *)
            echo "Usage: $0 {metadata|dependencies|install|configure|verify|uninstall}" >&2
            return 1
            ;;
    esac
}

# Run main if executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi