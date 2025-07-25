#!/bin/bash

# Claude CLI Component Installer
# Installs Anthropic's Claude CLI tool

set -e

# Component metadata
metadata() {
    echo "name=Claude CLI"
    echo "version=latest"
    echo "description=Anthropic's Claude AI assistant command-line interface"
}

# Component dependencies
dependencies() {
    # Claude requires Node.js
    echo ""
}

# Installation function
install() {
    echo "Installing Claude CLI..."
    
    # Check for Node.js
    if ! command -v node >/dev/null 2>&1; then
        echo "Node.js is required for Claude CLI. Installing Node.js first..."
        install_nodejs
    fi
    
    # Install Claude CLI globally via npm
    echo "Installing Claude CLI via npm..."
    npm install -g @anthropic-ai/claude-cli
    
    # Alternative: Install via direct download if npm fails
    if ! command -v claude >/dev/null 2>&1; then
        echo "npm installation failed, trying direct download..."
        install_direct
    fi
}

# Install Node.js if not present
install_nodejs() {
    if command -v apt-get >/dev/null 2>&1; then
        # Debian/Ubuntu
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        sudo apt-get install -y nodejs
    elif command -v yum >/dev/null 2>&1; then
        # RHEL/CentOS
        curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
        sudo yum install -y nodejs
    else
        # Manual installation
        echo "Installing Node.js manually..."
        local node_version="v20.10.0"
        local arch=$(uname -m)
        
        case "$arch" in
            x86_64)
                arch="x64"
                ;;
            aarch64|arm64)
                arch="arm64"
                ;;
        esac
        
        local url="https://nodejs.org/dist/${node_version}/node-${node_version}-linux-${arch}.tar.xz"
        
        curl -fsSL "$url" | tar -xJ -C /tmp
        sudo cp -r /tmp/node-${node_version}-linux-${arch}/* /usr/local/
        rm -rf /tmp/node-${node_version}-linux-${arch}
    fi
}

# Direct installation method
install_direct() {
    echo "Installing Claude CLI directly..."
    
    # Create directory for Claude
    sudo mkdir -p /opt/claude-cli
    
    # Download latest Claude CLI
    # Note: Update this URL when Claude CLI releases are available
    local claude_url="https://github.com/anthropics/claude-cli/releases/latest/download/claude-linux-x64"
    
    if curl -fsSL -o /tmp/claude "$claude_url"; then
        sudo mv /tmp/claude /opt/claude-cli/claude
        sudo chmod +x /opt/claude-cli/claude
        sudo ln -sf /opt/claude-cli/claude /usr/local/bin/claude
    else
        echo "Direct download failed. Claude CLI may not be publicly available yet." >&2
        echo "Please check https://www.anthropic.com for installation instructions." >&2
        return 1
    fi
}

# Configuration function
configure() {
    echo "Configuring Claude CLI..."
    
    # Check for API key in environment or file
    local api_key=""
    local api_key_file="/home/coder/.tokens/claude.key"
    
    if [ -n "$ANTHROPIC_API_KEY" ]; then
        api_key="$ANTHROPIC_API_KEY"
    elif [ -f "$api_key_file" ] && [ -s "$api_key_file" ]; then
        api_key=$(cat "$api_key_file")
    fi
    
    if [ -n "$api_key" ]; then
        # Configure Claude with API key
        export ANTHROPIC_API_KEY="$api_key"
        
        # Create Claude config directory
        mkdir -p ~/.config/claude
        
        # Save configuration
        cat > ~/.config/claude/config.json << EOF
{
  "api_key": "$api_key",
  "default_model": "claude-3-opus-20240229",
  "max_tokens": 4096,
  "temperature": 0.7
}
EOF
        
        # Also set in shell profile
        echo "export ANTHROPIC_API_KEY='$api_key'" >> ~/.bashrc
        echo "export ANTHROPIC_API_KEY='$api_key'" >> ~/.zshrc 2>/dev/null || true
        
        echo "Claude CLI configured with API key"
    else
        echo "No Anthropic API key found"
        echo "To configure Claude CLI, set ANTHROPIC_API_KEY environment variable"
        echo "or save your API key to: $api_key_file"
        
        # Create placeholder for API key
        mkdir -p $(dirname "$api_key_file")
        touch "$api_key_file"
        chmod 600 "$api_key_file"
    fi
    
    # Set up Claude aliases
    cat >> ~/.bashrc << 'EOF'

# Claude CLI aliases
alias claude-opus='claude --model claude-3-opus-20240229'
alias claude-sonnet='claude --model claude-3-sonnet-20240229'
alias claude-haiku='claude --model claude-3-haiku-20240307'
EOF
}

# Verification function
verify() {
    echo "Verifying Claude CLI installation..."
    
    # Check if Claude is installed
    if ! command -v claude >/dev/null 2>&1; then
        # Check npm global installation
        if npm list -g @anthropic-ai/claude-cli >/dev/null 2>&1; then
            echo "Claude CLI is installed via npm but not in PATH"
            echo "You may need to add npm global bin to PATH"
        else
            echo "Claude CLI not found" >&2
            return 1
        fi
    else
        local version=$(claude --version 2>/dev/null || echo "unknown")
        echo "Claude CLI installed: $version"
    fi
    
    # Check API key configuration
    if [ -n "$ANTHROPIC_API_KEY" ]; then
        echo "Anthropic API key is configured"
    else
        echo "Anthropic API key is not configured"
        echo "Set ANTHROPIC_API_KEY to use Claude CLI"
    fi
    
    return 0
}

# Uninstall function
uninstall() {
    echo "Uninstalling Claude CLI..."
    
    # Remove npm package
    if npm list -g @anthropic-ai/claude-cli >/dev/null 2>&1; then
        npm uninstall -g @anthropic-ai/claude-cli
    fi
    
    # Remove direct installation
    if [ -L /usr/local/bin/claude ]; then
        sudo rm -f /usr/local/bin/claude
    fi
    
    if [ -d /opt/claude-cli ]; then
        sudo rm -rf /opt/claude-cli
    fi
    
    # Remove configuration
    rm -rf ~/.config/claude
    
    # Remove from shell profiles
    sed -i '/ANTHROPIC_API_KEY/d' ~/.bashrc 2>/dev/null || true
    sed -i '/claude-opus/d' ~/.bashrc 2>/dev/null || true
    sed -i '/claude-sonnet/d' ~/.bashrc 2>/dev/null || true
    sed -i '/claude-haiku/d' ~/.bashrc 2>/dev/null || true
    
    echo "Claude CLI uninstalled"
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