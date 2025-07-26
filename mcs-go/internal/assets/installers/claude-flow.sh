#!/bin/bash

# Claude Flow Component Installer
# Installs Claude Flow AI orchestration tool

set -e

# Component metadata
metadata() {
    echo "name=Claude Flow"
    echo "version=latest"
    echo "description=AI orchestration and workflow automation tool"
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
    
    # Install Claude Flow via npm (using alpha version)
    echo "Installing Claude Flow via npm..."
    npm install -g claude-flow@alpha
    
    # Alternative: Create npx wrapper if npm global install fails
    if ! command -v claude-flow >/dev/null 2>&1; then
        echo "Creating claude-flow wrapper for npx..."
        create_npx_wrapper
    fi
    
    # Create Claude Flow directories
    mkdir -p /home/coder/.claude-flow/{workflows,templates,data}
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
    echo "Creating claude-flow wrapper..."
    
    # Create wrapper script
    sudo tee /usr/local/bin/claude-flow > /dev/null << 'EOF'
#!/bin/bash
# Claude Flow wrapper script
exec npx claude-flow@alpha "$@"
EOF
    
    # Make executable
    sudo chmod +x /usr/local/bin/claude-flow
    
    echo "Created claude-flow wrapper at /usr/local/bin/claude-flow"
}

# Configuration function
configure() {
    echo "Configuring Claude Flow..."
    
    # Create configuration directory
    local config_dir="/home/coder/.claude-flow"
    mkdir -p "$config_dir"
    
    # Check for API keys
    local anthropic_key=""
    local anthropic_key_file="/home/coder/.tokens/claude.key"
    
    if [ -n "$ANTHROPIC_API_KEY" ]; then
        anthropic_key="$ANTHROPIC_API_KEY"
    elif [ -f "$anthropic_key_file" ] && [ -s "$anthropic_key_file" ]; then
        anthropic_key=$(cat "$anthropic_key_file")
    fi
    
    # Create Claude Flow configuration
    cat > "$config_dir/config.json" << EOF
{
  "version": "1.0",
  "api": {
    "anthropic": {
      "key": "${anthropic_key:-YOUR_ANTHROPIC_API_KEY}",
      "model": "claude-3-opus-20240229"
    }
  },
  "workflows": {
    "directory": "$config_dir/workflows",
    "autoSave": true
  },
  "orchestration": {
    "maxAgents": 10,
    "defaultTopology": "mesh",
    "parallelExecution": true
  },
  "memory": {
    "enabled": true,
    "directory": "$config_dir/data",
    "ttl": 86400
  },
  "performance": {
    "enableCaching": true,
    "enableMetrics": true
  }
}
EOF
    
    # Create sample workflows
    create_sample_workflows "$config_dir/workflows"
    
    # Set up environment variables
    cat >> ~/.bashrc << EOF

# Claude Flow configuration
export CLAUDE_FLOW_HOME="$config_dir"
export CLAUDE_FLOW_CONFIG="$config_dir/config.json"

# Claude Flow aliases
alias cf='claude-flow'
alias cfw='claude-flow workflow'
alias cfa='claude-flow agent'
alias cfs='claude-flow swarm'
EOF
    
    # Create shell completion
    if command -v claude-flow >/dev/null 2>&1; then
        claude-flow completion bash > /tmp/claude-flow-completion.bash 2>/dev/null || true
        if [ -f /tmp/claude-flow-completion.bash ]; then
            sudo mv /tmp/claude-flow-completion.bash /etc/bash_completion.d/
        fi
    fi
    
    echo "Claude Flow configured successfully"
}

# Create sample workflows
create_sample_workflows() {
    local workflows_dir="$1"
    mkdir -p "$workflows_dir"
    
    # Sample workflow: Code Review
    cat > "$workflows_dir/code-review.json" << 'EOF'
{
  "name": "Code Review Workflow",
  "description": "Automated code review with multiple agents",
  "agents": [
    {
      "name": "analyzer",
      "type": "code-analyzer",
      "tasks": ["syntax", "style", "complexity"]
    },
    {
      "name": "security",
      "type": "security-scanner",
      "tasks": ["vulnerabilities", "secrets", "dependencies"]
    },
    {
      "name": "optimizer",
      "type": "performance-optimizer",
      "tasks": ["bottlenecks", "memory", "algorithms"]
    }
  ],
  "topology": "hierarchical",
  "output": "markdown"
}
EOF
    
    # Sample workflow: Project Setup
    cat > "$workflows_dir/project-setup.json" << 'EOF'
{
  "name": "Project Setup Workflow",
  "description": "Initialize new project with best practices",
  "agents": [
    {
      "name": "architect",
      "type": "project-architect",
      "tasks": ["structure", "dependencies", "configuration"]
    },
    {
      "name": "documenter",
      "type": "documentation-writer",
      "tasks": ["readme", "contributing", "license"]
    },
    {
      "name": "tester",
      "type": "test-generator",
      "tasks": ["unit-tests", "integration-tests", "ci-cd"]
    }
  ],
  "topology": "mesh",
  "interactive": true
}
EOF
}

# Verification function
verify() {
    echo "Verifying Claude Flow installation..."
    
    # Check if Claude Flow is installed
    if ! command -v claude-flow >/dev/null 2>&1; then
        # Check npm global installation
        if npm list -g claude-flow >/dev/null 2>&1; then
            echo "Claude Flow is installed via npm but not in PATH"
            echo "You may need to add npm global bin to PATH"
        else
            echo "Claude Flow not found" >&2
            return 1
        fi
    else
        local version=$(claude-flow --version 2>/dev/null || echo "unknown")
        echo "Claude Flow installed: $version"
    fi
    
    # Check configuration
    local config_file="/home/coder/.claude-flow/config.json"
    if [ -f "$config_file" ]; then
        echo "Claude Flow configuration found"
    else
        echo "Claude Flow configuration not found"
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
    
    # Test Claude Flow
    if command -v claude-flow >/dev/null 2>&1; then
        echo "Testing Claude Flow..."
        claude-flow --help >/dev/null 2>&1 && echo "Claude Flow is working correctly"
    fi
    
    return 0
}

# Uninstall function
uninstall() {
    echo "Uninstalling Claude Flow..."
    
    # Remove npm package
    if npm list -g claude-flow >/dev/null 2>&1; then
        npm uninstall -g claude-flow
    fi
    
    # Remove configuration and data
    rm -rf /home/coder/.claude-flow
    
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