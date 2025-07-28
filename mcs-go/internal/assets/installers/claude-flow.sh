#!/bin/bash

# Claude Flow Component Installer
# Installs Claude Flow AI orchestration tool

set -e

# Component metadata
metadata() {
    echo "name=Claude Flow"
    echo "version=alpha"
    echo "description=AI orchestration and workflow automation tool (initializes via npx)"
}

# Component dependencies
dependencies() {
    # Claude Flow may benefit from Claude CLI
    echo "claude"
}

# Installation function
install() {
    echo "Installing Claude Flow..."
    
    # Check for Node.js (required for npm)
    if ! command -v node >/dev/null 2>&1; then
        echo "Node.js is required for Claude Flow. Installing Node.js first..."
        install_nodejs
    fi
    
    # Set up npm prefix for local installation
    export NPM_PREFIX="$HOME/.npm-global"
    mkdir -p "$NPM_PREFIX"
    npm config set prefix "$NPM_PREFIX"
    
    # Initialize Claude Flow using npx
    echo "Initializing Claude Flow via npx..."
    cd "$HOME" || exit 1
    npx claude-flow@alpha init --force
    
    # Create wrapper script for claude-flow command
    if mkdir -p "$HOME/.local/bin" 2>/dev/null; then
        cat > "$HOME/.local/bin/claude-flow" << 'EOF'
#!/bin/bash
# Claude Flow wrapper script
exec npx claude-flow@alpha "$@"
EOF
        chmod +x "$HOME/.local/bin/claude-flow"
        echo "Created claude-flow wrapper at $HOME/.local/bin/claude-flow"
    else
        echo "Warning: Could not create ~/.local/bin"
        create_npx_wrapper
    fi
    
    # Update PATH if needed
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
        export PATH="$HOME/.local/bin:$PATH"
    fi
    
    # Create Claude Flow directories
    mkdir -p /home/coder/.claude-flow/{workflows,templates,data}
}

# Install Node.js if not present
install_nodejs() {
    if command -v apt-get >/dev/null 2>&1; then
        # Debian/Ubuntu
        echo "Error: Node.js is required but not installed."
        echo "Please ask your administrator to install Node.js."
        return 1
    elif command -v yum >/dev/null 2>&1; then
        # RHEL/CentOS
        echo "Error: Node.js is required but not installed."
        echo "Please ask your administrator to install Node.js."
        return 1
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
        mkdir -p "$HOME/.local"
        cp -r /tmp/node-${node_version}-linux-${arch}/* "$HOME/.local/"
        rm -rf /tmp/node-${node_version}-linux-${arch}
        export PATH="$HOME/.local/bin:$PATH"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    fi
}

# Create npx wrapper
create_npx_wrapper() {
    echo "Creating alternative claude-flow wrapper..."
    
    # Try alternative locations
    local wrapper_dir=""
    if [ -w "$HOME/bin" ]; then
        wrapper_dir="$HOME/bin"
    elif [ -w "$HOME/.npm-global/bin" ]; then
        wrapper_dir="$HOME/.npm-global/bin"
    else
        echo "Warning: No writable directory found for wrapper script"
        echo "You can use 'npx claude-flow@alpha' directly"
        return 1
    fi
    
    mkdir -p "$wrapper_dir"
    cat > "$wrapper_dir/claude-flow" << 'EOF'
#!/bin/bash
# Claude Flow wrapper script
exec npx claude-flow@alpha "$@"
EOF
    chmod +x "$wrapper_dir/claude-flow"
    echo "Created claude-flow wrapper at $wrapper_dir/claude-flow"
}

# Configuration function
configure() {
    echo "Configuring Claude Flow environment..."
    
    # Check for API keys
    local anthropic_key=""
    local anthropic_key_file="$HOME/.tokens/claude.key"
    
    if [ -n "$ANTHROPIC_API_KEY" ]; then
        anthropic_key="$ANTHROPIC_API_KEY"
    elif [ -f "$anthropic_key_file" ] && [ -s "$anthropic_key_file" ]; then
        anthropic_key=$(cat "$anthropic_key_file")
    fi
    
    # Set up environment variables if API key found
    if [ -n "$anthropic_key" ]; then
        echo "export ANTHROPIC_API_KEY='$anthropic_key'" >> ~/.bashrc
        export ANTHROPIC_API_KEY="$anthropic_key"
        echo "Anthropic API key configured"
    else
        echo "No Anthropic API key found"
        echo "Claude Flow will work without an API key (using browser auth)"
    fi
    
    # Set up Claude Flow aliases
    cat >> ~/.bashrc << 'EOF'

# Claude Flow aliases
alias cf='npx claude-flow@alpha'
alias cfw='npx claude-flow@alpha workflow'
alias cfa='npx claude-flow@alpha agent'
alias cfs='npx claude-flow@alpha swarm'
EOF
    
    # Source bashrc to make aliases available immediately
    source ~/.bashrc
    
    echo "Claude Flow environment configured successfully"
}


# Verification function
verify() {
    echo "Verifying Claude Flow installation..."
    
    # Check if Claude Flow wrapper exists and works
    if command -v claude-flow >/dev/null 2>&1; then
        echo "Claude Flow wrapper found"
        # Test if it works
        if claude-flow --version >/dev/null 2>&1; then
            local version=$(claude-flow --version 2>/dev/null || echo "unknown")
            echo "Claude Flow is working: $version"
        else
            echo "Claude Flow wrapper exists but command failed"
            echo "Try running: npx claude-flow@alpha --version"
        fi
    else
        echo "Claude Flow wrapper not found in PATH"
        echo "You can use: npx claude-flow@alpha"
    fi
    
    # Check if Claude Flow is initialized
    if [ -f "$HOME/claude-flow.config.json" ] || [ -f "$HOME/.claude/settings.json" ]; then
        echo "Claude Flow is initialized"
    else
        echo "Claude Flow not initialized"
        echo "Run: npx claude-flow@alpha init --force"
    fi
    
    # Check API key configuration
    if [ -f "$config_file" ]; then
        if grep -q "YOUR_ANTHROPIC_API_KEY" "$config_file"; then
            echo "Warning: Anthropic API key not configured"
            echo "Update $config_file with your API key"
        else
            echo "Anthropic API key is configured"
        fi
    fi
    
    # Test npx command
    echo "Testing Claude Flow via npx..."
    if npx claude-flow@alpha --version >/dev/null 2>&1; then
        echo "Claude Flow is accessible via npx"
    else
        echo "Error: Could not run Claude Flow via npx"
    fi
    
    return 0
}

# Uninstall function
uninstall() {
    echo "Uninstalling Claude Flow..."
    
    # Remove wrapper scripts
    rm -f "$HOME/.local/bin/claude-flow"
    rm -f "$HOME/bin/claude-flow"
    rm -f "$HOME/.npm-global/bin/claude-flow"
    
    # Remove configuration and data
    rm -rf "$HOME/.claude-flow"
    rm -rf "$HOME/.claude"
    rm -rf "$HOME/.swarm"
    rm -rf "$HOME/.hive-mind"
    rm -f "$HOME/claude-flow.config.json"
    rm -f "$HOME/.mcp.json"
    
    # Remove from shell profiles
    sed -i '/CLAUDE_FLOW/d' ~/.bashrc 2>/dev/null || true
    sed -i '/alias cf=/d' ~/.bashrc 2>/dev/null || true
    sed -i '/alias cfw=/d' ~/.bashrc 2>/dev/null || true
    sed -i '/alias cfa=/d' ~/.bashrc 2>/dev/null || true
    sed -i '/alias cfs=/d' ~/.bashrc 2>/dev/null || true
    
    # Remove completion
    sudo rm -f /etc/bash_completion.d/claude-flow-completion.bash
    
    echo "Claude Flow uninstalled"
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
            echo ""
            echo "✅ Claude Flow initialized successfully!"
            echo "You can now use 'claude-flow' or 'npx claude-flow@alpha' commands."
            echo ""
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