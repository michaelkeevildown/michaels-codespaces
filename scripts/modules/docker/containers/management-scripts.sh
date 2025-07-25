#!/bin/bash

# Management Scripts Module
# Creates management scripts and aliases for codespaces

# Create README for a codespace
create_codespace_readme() {
    local codespace_dir="$1"
    local safe_name="$2"
    local repo_url="$3"
    local docker_image="$4"
    local language="$5"
    local ports="$6"
    local password="$7"
    
    local vs_code_port=$(echo "$ports" | cut -d',' -f1 | cut -d':' -f1)
    
    cat > "$codespace_dir/README.md" << EOF
# $safe_name Codespace

## Quick Start

\`\`\`bash
# Start codespace
mcs start $safe_name

# Access VS Code
open http://localhost:$vs_code_port
Password: $password

# Stop codespace
mcs stop $safe_name
\`\`\`

## Management Commands

- \`mcs info $safe_name\` - Show codespace information
- \`mcs logs $safe_name\` - View container logs
- \`mcs exec $safe_name\` - Enter container shell
- \`mcs restart $safe_name\` - Restart container
- \`mcs remove $safe_name\` - Remove codespace

## Configuration

- **Repository**: $repo_url
- **Docker Image**: $docker_image
- **Language**: ${language:-Not specified}
- **Created**: $(date +%Y-%m-%d)

## Ports

$(echo "$ports" | tr ',' '\n' | while IFS=':' read -r host container; do
    echo "- **$host** → Container port $container"
done)

## Files

- \`src/\` - Repository source code
- \`data/\` - VS Code server data
- \`config/\` - Configuration files
- \`logs/\` - Application logs
EOF
}

# Create shell aliases for a codespace
create_shell_aliases() {
    local codespace_dir="$1"
    local safe_name="$2"
    
    local alias_file="$codespace_dir/aliases.sh"
    cat > "$alias_file" << EOF
#!/bin/bash

# Aliases for $safe_name codespace

alias cd-$safe_name='cd $codespace_dir'
alias src-$safe_name='cd $codespace_dir/src'

# Legacy aliases for compatibility
alias start-$safe_name='mcs start $safe_name'
alias stop-$safe_name='mcs stop $safe_name'
alias logs-$safe_name='mcs logs $safe_name'
alias exec-$safe_name='mcs exec $safe_name'

echo "✅ Codespace aliases loaded for: $safe_name"
EOF
    
    chmod +x "$alias_file"
}

# Add aliases to shell configuration
register_aliases_in_shell() {
    local codespace_dir="$1"
    local safe_name="$2"
    local alias_file="$codespace_dir/aliases.sh"
    
    # Determine shell config file
    local shell_config=""
    if [ -n "$ZSH_VERSION" ] || [ -f ~/.zshrc ]; then
        shell_config="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ] || [ -f ~/.bashrc ]; then
        shell_config="$HOME/.bashrc"
    else
        echo_warning "Could not determine shell configuration file"
        return 1
    fi
    
    # Add to shell config if not already present
    if ! grep -q "# Codespace: $safe_name" "$shell_config" 2>/dev/null; then
        echo "" >> "$shell_config"
        echo "# Codespace: $safe_name" >> "$shell_config"
        echo "[ -f \"$alias_file\" ] && source \"$alias_file\"" >> "$shell_config"
        echo_success "Aliases added to $shell_config"
    else
        echo_debug "Aliases already registered in $shell_config"
    fi
}

# Remove aliases from shell configuration
unregister_aliases_from_shell() {
    local safe_name="$1"
    
    # Determine shell config file
    local shell_config=""
    if [ -n "$ZSH_VERSION" ] || [ -f ~/.zshrc ]; then
        shell_config="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ] || [ -f ~/.bashrc ]; then
        shell_config="$HOME/.bashrc"
    else
        return 0
    fi
    
    # Remove alias registration
    if [ -f "$shell_config" ]; then
        # Create temp file without the codespace aliases
        local temp_file=$(mktemp)
        awk -v name="$safe_name" '
            /^# Codespace: / && $0 ~ name {skip=2; next}
            skip > 0 {skip--; next}
            {print}
        ' "$shell_config" > "$temp_file"
        
        mv "$temp_file" "$shell_config"
        echo_debug "Removed aliases for $safe_name from $shell_config"
    fi
}

# Create VS Code workspace configuration
create_vscode_workspace() {
    local codespace_dir="$1"
    local safe_name="$2"
    local language="$3"
    
    local workspace_file="$codespace_dir/${safe_name}.code-workspace"
    
    # Create workspace configuration
    cat > "$workspace_file" << EOF
{
    "folders": [
        {
            "path": "/home/coder/project"
        }
    ],
    "settings": {
        "terminal.integrated.defaultProfile.linux": "bash",
        "terminal.integrated.cwd": "/home/coder/project",
        "workbench.startupEditor": "none",
        "explorer.openEditors.visible": 0,
        "files.autoSave": "afterDelay",
        "files.autoSaveDelay": 1000
    },
    "extensions": {
        "recommendations": [
EOF
    
    # Add language-specific recommendations
    case "$language" in
        "javascript"|"typescript"|"node"|"nodejs")
            cat >> "$workspace_file" << EOF
            "dbaeumer.vscode-eslint",
            "esbenp.prettier-vscode",
            "formulahendry.auto-rename-tag",
            "christian-kohler.npm-intellisense"
EOF
            ;;
        "python")
            cat >> "$workspace_file" << EOF
            "ms-python.python",
            "ms-python.vscode-pylance",
            "ms-python.black-formatter",
            "charliermarsh.ruff"
EOF
            ;;
        "go"|"golang")
            cat >> "$workspace_file" << EOF
            "golang.go"
EOF
            ;;
        "rust")
            cat >> "$workspace_file" << EOF
            "rust-lang.rust-analyzer",
            "tamasfe.even-better-toml"
EOF
            ;;
        "java")
            cat >> "$workspace_file" << EOF
            "vscjava.vscode-java-pack",
            "vscjava.vscode-maven"
EOF
            ;;
    esac
    
    cat >> "$workspace_file" << EOF
        ]
    }
}
EOF
    
    echo_debug "Created VS Code workspace file: $workspace_file"
}

# Create management scripts wrapper
create_management_scripts() {
    local codespace_dir="$1"
    local safe_name="$2"
    local repo_url="$3"
    local docker_image="$4"
    local language="$5"
    local ports="$6"
    local password="$7"
    
    echo_info "Creating management scripts..."
    
    # Create README
    create_codespace_readme "$codespace_dir" "$safe_name" "$repo_url" \
                          "$docker_image" "$language" "$ports" "$password"
    
    # Create shell aliases
    create_shell_aliases "$codespace_dir" "$safe_name"
    
    # Register aliases in shell
    register_aliases_in_shell "$codespace_dir" "$safe_name"
    
    # Create VS Code workspace configuration
    create_vscode_workspace "$codespace_dir" "$safe_name" "$language"
    
    echo_success "Management scripts created"
}

# Display success message after codespace creation
display_codespace_success() {
    local safe_name="$1"
    local ports="$2"
    local password="$3"
    
    local vs_code_port=$(echo "$ports" | cut -d',' -f1 | cut -d':' -f1)
    
    echo ""
    echo_success "🎉 Codespace created successfully!"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📦 Codespace: $safe_name"
    echo "🌐 VS Code URL: http://localhost:$vs_code_port"
    echo "🔑 Password: $password"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Quick commands:"
    echo "  mcs info $safe_name    - Show details"
    echo "  mcs logs $safe_name    - View logs"
    echo "  mcs exec $safe_name    - Enter container"
    echo "  mcs stop $safe_name    - Stop codespace"
    echo ""
    echo "To load aliases in current shell:"
    echo "  source ~/.zshrc"
    echo ""
}

# Export functions
export -f create_codespace_readme
export -f create_shell_aliases
export -f register_aliases_in_shell
export -f unregister_aliases_from_shell
export -f create_vscode_workspace
export -f create_management_scripts
export -f display_codespace_success