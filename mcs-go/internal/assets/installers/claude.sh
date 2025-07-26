#!/bin/bash

# Claude Code Component Installer
# Installs Anthropic's Claude Code (claude-code) CLI tool

set -e

# Component metadata
metadata() {
    echo "name=Claude Code"
    echo "version=latest"
    echo "description=Anthropic's Claude AI coding assistant (claude-code)"
}

# Component dependencies
dependencies() {
    # Claude requires Node.js
    echo ""
}

# Installation function
install() {
    echo "Installing Claude Code..."
    
    # Check for Node.js
    if ! command -v node >/dev/null 2>&1; then
        echo "Node.js is required for Claude Code. Installing Node.js first..."
        install_nodejs
    fi
    
    # Install Claude Code globally via npm
    echo "Installing Claude Code via npm..."
    npm install -g claude-code@latest
    
    # Alternative: Install via npx if npm global install fails
    if ! command -v claude-code >/dev/null 2>&1; then
        echo "Creating claude-code wrapper for npx..."
        create_npx_wrapper
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

# Create npx wrapper
create_npx_wrapper() {
    echo "Creating claude-code wrapper..."
    
    # Create wrapper script
    sudo tee /usr/local/bin/claude-code > /dev/null << 'EOF'
#!/bin/bash
# Claude Code wrapper script
exec npx claude-code@latest "$@"
EOF
    
    # Make executable
    sudo chmod +x /usr/local/bin/claude-code
    
    echo "Created claude-code wrapper at /usr/local/bin/claude-code"
}

# Configuration function
configure() {
    echo "Configuring Claude Code..."
    
    # Check for API key in environment or file
    local api_key=""
    local api_key_file="/home/coder/.tokens/claude.key"
    
    if [ -n "$ANTHROPIC_API_KEY" ]; then
        api_key="$ANTHROPIC_API_KEY"
    elif [ -f "$api_key_file" ] && [ -s "$api_key_file" ]; then
        api_key=$(cat "$api_key_file")
    fi
    
    if [ -n "$api_key" ]; then
        # Configure Claude Code with API key
        export ANTHROPIC_API_KEY="$api_key"
        
        # Create Claude config directory
        mkdir -p ~/.config/claude-code
        
        # Save configuration
        cat > ~/.config/claude-code/config.json << EOF
{
  "api_key": "$api_key",
  "default_model": "claude-3-opus-20240229",
  "enable_caching": true
}
EOF
        
        # Also set in shell profile
        echo "export ANTHROPIC_API_KEY='$api_key'" >> ~/.bashrc
        echo "export ANTHROPIC_API_KEY='$api_key'" >> ~/.zshrc 2>/dev/null || true
        
        echo "Claude Code configured with API key"
    else
        echo "No Anthropic API key found"
        echo "Claude Code will work without an API key (using browser auth)"
        echo "To use API key auth, save your key to: $api_key_file"
        
        # Create placeholder for API key
        mkdir -p $(dirname "$api_key_file")
        touch "$api_key_file"
        chmod 600 "$api_key_file"
    fi
    
    # Set up Claude Code aliases
    cat >> ~/.bashrc << 'EOF'

# Claude Code aliases
alias claude='claude-code'
alias cc='claude-code'
alias ccd='claude-code --debug'
EOF
}

# Verification function
verify() {
    echo "Verifying Claude Code installation..."
    
    # Check if Claude Code is installed
    if ! command -v claude-code >/dev/null 2>&1; then
        # Check npm global installation
        if npm list -g claude-code >/dev/null 2>&1; then
            echo "Claude Code is installed via npm but not in PATH"
            echo "You may need to add npm global bin to PATH"
        else
            echo "Claude Code not found" >&2
            return 1
        fi
    else
        local version=$(claude-code --version 2>/dev/null || echo "unknown")
        echo "Claude Code installed: $version"
    fi
    
    # Check API key configuration
    if [ -n "$ANTHROPIC_API_KEY" ]; then
        echo "Anthropic API key is configured"
    else
        echo "Claude Code can work without API key (browser auth)"
    fi
    
    return 0
}

# Uninstall function
uninstall() {
    echo "Uninstalling Claude Code..."
    
    # Remove npm package
    if npm list -g claude-code >/dev/null 2>&1; then
        npm uninstall -g claude-code
    fi
    
    # Remove wrapper script
    if [ -f /usr/local/bin/claude-code ]; then
        sudo rm -f /usr/local/bin/claude-code
    fi
    
    # Remove configuration
    rm -rf ~/.config/claude-code
    
    # Remove from shell profiles
    sed -i '/ANTHROPIC_API_KEY/d' ~/.bashrc 2>/dev/null || true
    sed -i '/alias claude=/d' ~/.bashrc 2>/dev/null || true
    sed -i '/alias cc=/d' ~/.bashrc 2>/dev/null || true
    sed -i '/alias ccd=/d' ~/.bashrc 2>/dev/null || true
    
    echo "Claude Code uninstalled"
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