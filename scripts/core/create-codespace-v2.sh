#!/bin/bash

# Enhanced Codespace Creation Script
# Uses modular architecture for better maintainability

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULES_DIR="$SCRIPT_DIR/../modules"

# Source all required modules
source "$MODULES_DIR/github/auth/github-auth.sh"
source "$MODULES_DIR/github/clone/github-clone.sh"
source "$MODULES_DIR/docker/compose/docker-compose-generator.sh"
source "$MODULES_DIR/networking/port-manager.sh"

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
    echo_step() { echo "▶️  $1"; }
    echo_debug() { [ "${DEBUG:-0}" -eq 1 ] && echo "🔍 $1"; }
fi

# Display usage
usage() {
    cat << EOF
Usage: $0 <repository-url> [options]

Create a new codespace from a GitHub repository.

Arguments:
  repository-url    Git repository URL (required)

Options:
  --name <name>     Custom codespace name (default: auto-generated)
  --image <image>   Docker image to use (default: codercom/code-server:latest)
  --language <lang> Language preset (node, python, go, rust, java)
  --ports <ports>   Port mappings (format: "8080:8080,3000:3000")
  --env-file <file> Environment variables file
  --no-start        Don't start the container after creation
  --force           Overwrite existing codespace
  --debug           Enable debug output

Examples:
  $0 git@github.com:facebook/react.git
  $0 https://github.com/nodejs/node.git --language node
  $0 git@github.com:user/repo.git --name my-project --ports "8090:8080"

EOF
}

# Parse command line arguments
parse_arguments() {
    REPO_URL=""
    CODESPACE_NAME=""
    DOCKER_IMAGE=""
    LANGUAGE=""
    CUSTOM_PORTS=""
    ENV_FILE=""
    NO_START=false
    FORCE=false
    
    # Check for help
    if [ $# -eq 0 ] || [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
        usage
        exit 0
    fi
    
    # First argument is repository URL
    REPO_URL="$1"
    shift
    
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            --name)
                CODESPACE_NAME="$2"
                shift 2
                ;;
            --image)
                DOCKER_IMAGE="$2"
                shift 2
                ;;
            --language|--lang)
                LANGUAGE="$2"
                shift 2
                ;;
            --ports)
                CUSTOM_PORTS="$2"
                shift 2
                ;;
            --env-file)
                ENV_FILE="$2"
                if [ ! -f "$ENV_FILE" ]; then
                    echo_error "Environment file not found: $ENV_FILE"
                    exit 1
                fi
                shift 2
                ;;
            --no-start)
                NO_START=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --debug)
                export DEBUG=1
                shift
                ;;
            *)
                echo_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Validate repository URL
validate_inputs() {
    # Validate URL format
    if ! validate_repo_url "$REPO_URL"; then
        echo_error "Invalid repository URL: $REPO_URL"
        exit 1
    fi
    
    # Check repository access
    echo_info "Checking repository access..."
    if ! validate_repo_access "$REPO_URL"; then
        echo_warning "Cannot verify repository access. It may be private or require authentication."
        echo_info "Make sure you have configured GitHub authentication if needed."
    fi
    
    # Generate codespace name if not provided
    if [ -z "$CODESPACE_NAME" ]; then
        local repo_name=$(basename "$REPO_URL" .git)
        local repo_owner=$(echo "$REPO_URL" | sed -E 's/.*[:/]([^/]+)\/[^/]+$/\1/')
        CODESPACE_NAME="${repo_owner}-${repo_name}"
    fi
    
    # Sanitize name
    SAFE_NAME=$(echo "$CODESPACE_NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g')
    CODESPACE_DIR="$HOME/codespaces/$SAFE_NAME"
    
    # Check if already exists
    if [ -d "$CODESPACE_DIR" ] && [ "$FORCE" != "true" ]; then
        echo_error "Codespace '$SAFE_NAME' already exists!"
        echo "Use --force to overwrite or choose a different name with --name"
        exit 1
    fi
}

# Main creation process
create_codespace() {
    echo ""
    echo_step "🚀 Creating codespace: $SAFE_NAME"
    echo "Repository: $REPO_URL"
    echo ""
    
    # Clean up if forcing
    if [ -d "$CODESPACE_DIR" ] && [ "$FORCE" == "true" ]; then
        echo_warning "Removing existing codespace..."
        if [ -f "$CODESPACE_DIR/docker-compose.yml" ]; then
            docker-compose -f "$CODESPACE_DIR/docker-compose.yml" down 2>/dev/null || true
        fi
        unregister_codespace_ports "$SAFE_NAME"
        rm -rf "$CODESPACE_DIR"
    fi
    
    # Create directory structure
    echo_info "Creating directory structure..."
    mkdir -p "$CODESPACE_DIR"/{src,data,config,logs}
    
    # Clone repository
    echo_info "Cloning repository..."
    if ! clone_with_retry "$REPO_URL" "$CODESPACE_DIR/src" 3 5; then
        echo_error "Failed to clone repository"
        rm -rf "$CODESPACE_DIR"
        exit 1
    fi
    
    # Check for .devcontainer
    local devcontainer_file=$(check_devcontainer "$CODESPACE_DIR/src")
    if [ -n "$devcontainer_file" ]; then
        echo_info "Found .devcontainer configuration: $devcontainer_file"
        
        # Parse image from devcontainer if not specified
        if [ -z "$DOCKER_IMAGE" ]; then
            local devcontainer_image=$(parse_devcontainer_image "$devcontainer_file")
            if [ -n "$devcontainer_image" ]; then
                DOCKER_IMAGE="$devcontainer_image"
                echo_info "Using image from devcontainer: $DOCKER_IMAGE"
            fi
        fi
    fi
    
    # Detect language if not specified
    if [ -z "$LANGUAGE" ] && [ -z "$DOCKER_IMAGE" ]; then
        detect_language "$CODESPACE_DIR/src"
    fi
    
    # Set default image if still not specified
    if [ -z "$DOCKER_IMAGE" ]; then
        DOCKER_IMAGE="codercom/code-server:latest"
    fi
    
    # Allocate ports
    echo_info "Allocating ports..."
    local ports
    if [ -n "$CUSTOM_PORTS" ]; then
        # Parse custom ports
        ports="$CUSTOM_PORTS"
        # Register each port
        IFS=',' read -ra PORT_PAIRS <<< "$CUSTOM_PORTS"
        local i=1
        for pair in "${PORT_PAIRS[@]}"; do
            local host_port=$(echo "$pair" | cut -d':' -f1)
            local service="vscode"
            [ $i -gt 1 ] && service="app$((i-1))"
            register_port "$host_port" "$SAFE_NAME" "$service"
            ((i++))
        done
    else
        # Auto-allocate ports
        local allocated=$(allocate_codespace_ports "$SAFE_NAME" 2)
        if [ -z "$allocated" ]; then
            echo_error "Failed to allocate ports"
            rm -rf "$CODESPACE_DIR"
            exit 1
        fi
        local vs_code_port=$(echo "$allocated" | cut -d' ' -f1)
        local app_port=$(echo "$allocated" | cut -d' ' -f2)
        ports="${vs_code_port}:8080,${app_port}:3000"
    fi
    
    echo_debug "Port configuration: $ports"
    
    # Generate password
    local password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-16 2>/dev/null || \
                    tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16)
    
    # Create environment file
    create_env_file "$CODESPACE_DIR" "$ports" "$password"
    
    # Generate docker-compose.yml
    echo_info "Generating Docker configuration..."
    generate_docker_compose "$CODESPACE_DIR" "$ports" "$password"
    
    # Create management scripts
    create_management_scripts "$CODESPACE_DIR" "$ports" "$password"
    
    # Start container if requested
    if [ "$NO_START" != "true" ]; then
        echo_info "Starting codespace..."
        if docker-compose -f "$CODESPACE_DIR/docker-compose.yml" up -d; then
            echo_success "Codespace started successfully!"
        else
            echo_error "Failed to start codespace"
            exit 1
        fi
    fi
    
    # Display success message
    display_success_message "$ports" "$password"
}

# Detect project language
detect_language() {
    local src_dir="$1"
    
    if [ -f "$src_dir/package.json" ]; then
        LANGUAGE="node"
        echo_info "Detected Node.js project"
    elif [ -f "$src_dir/requirements.txt" ] || [ -f "$src_dir/setup.py" ]; then
        LANGUAGE="python"
        echo_info "Detected Python project"
    elif [ -f "$src_dir/go.mod" ]; then
        LANGUAGE="go"
        echo_info "Detected Go project"
    elif [ -f "$src_dir/Cargo.toml" ]; then
        LANGUAGE="rust"
        echo_info "Detected Rust project"
    elif [ -f "$src_dir/pom.xml" ] || [ -f "$src_dir/build.gradle" ]; then
        LANGUAGE="java"
        echo_info "Detected Java project"
    fi
}

# Create environment file
create_env_file() {
    local codespace_dir="$1"
    local ports="$2"
    local password="$3"
    
    local vs_code_port=$(echo "$ports" | cut -d',' -f1 | cut -d':' -f1)
    local app_port=$(echo "$ports" | cut -d',' -f2 | cut -d':' -f1)
    
    cat > "$codespace_dir/.env" << EOF
# Codespace Configuration
CODESPACE_NAME=$SAFE_NAME
REPO_NAME=$(basename "$REPO_URL" .git)
REPO_URL=$REPO_URL
CONTAINER_NAME=${SAFE_NAME}-dev
VS_CODE_PORT=$vs_code_port
APP_PORT=$app_port
PASSWORD=$password
DOCKER_IMAGE=$DOCKER_IMAGE
LANGUAGE=$LANGUAGE
CREATED=$(date +%Y-%m-%d)

# User Configuration
USER=$USER
TZ=$(timedatectl show -p Timezone --value 2>/dev/null || echo "UTC")
EOF

    # Append custom environment variables
    if [ -n "$ENV_FILE" ]; then
        echo "" >> "$codespace_dir/.env"
        echo "# Custom Environment Variables" >> "$codespace_dir/.env"
        cat "$ENV_FILE" >> "$codespace_dir/.env"
    fi
}

# Generate docker-compose.yml
generate_docker_compose() {
    local codespace_dir="$1"
    local ports="$2"
    local password="$3"
    
    # Prepare configuration
    declare -A config
    config[container_name]="${SAFE_NAME}-dev"
    config[image]="$DOCKER_IMAGE"
    config[password]="$password"
    config[ports]="$ports"
    config[env_vars]=""
    config[labels]="codespace.repo=$(basename "$REPO_URL" .git)\ncodespace.created=$(date +%Y-%m-%d)\ncodespace.url=$REPO_URL"
    
    # Generate based on language if specified
    if [ -n "$LANGUAGE" ]; then
        generate_language_compose "$LANGUAGE" config > "$codespace_dir/docker-compose.yml"
    else
        generate_basic_compose config > "$codespace_dir/docker-compose.yml"
    fi
    
    # Validate the generated file
    validate_compose "$codespace_dir/docker-compose.yml"
}

# Create management scripts and aliases
create_management_scripts() {
    local codespace_dir="$1"
    local ports="$2"
    local password="$3"
    
    # Create README
    local vs_code_port=$(echo "$ports" | cut -d',' -f1 | cut -d':' -f1)
    
    cat > "$codespace_dir/README.md" << EOF
# $SAFE_NAME Codespace

## Quick Start

\`\`\`bash
# Start codespace
mcs start $SAFE_NAME

# Access VS Code
open http://localhost:$vs_code_port
Password: $password

# Stop codespace
mcs stop $SAFE_NAME
\`\`\`

## Management Commands

- \`mcs info $SAFE_NAME\` - Show codespace information
- \`mcs logs $SAFE_NAME\` - View container logs
- \`mcs exec $SAFE_NAME\` - Enter container shell
- \`mcs restart $SAFE_NAME\` - Restart container
- \`mcs remove $SAFE_NAME\` - Remove codespace

## Configuration

- **Repository**: $REPO_URL
- **Docker Image**: $DOCKER_IMAGE
- **Language**: ${LANGUAGE:-Not specified}
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

    # Add to shell aliases
    local alias_file="$codespace_dir/aliases.sh"
    cat > "$alias_file" << EOF
#!/bin/bash

# Aliases for $SAFE_NAME codespace

alias cd-$SAFE_NAME='cd $codespace_dir'
alias src-$SAFE_NAME='cd $codespace_dir/src'

# Legacy aliases for compatibility
alias start-$SAFE_NAME='mcs start $SAFE_NAME'
alias stop-$SAFE_NAME='mcs stop $SAFE_NAME'
alias logs-$SAFE_NAME='mcs logs $SAFE_NAME'
alias exec-$SAFE_NAME='mcs exec $SAFE_NAME'

echo "✅ Codespace aliases loaded for: $SAFE_NAME"
EOF

    # Add to .zshrc if not already present
    if ! grep -q "# Codespace: $SAFE_NAME" ~/.zshrc 2>/dev/null; then
        echo "" >> ~/.zshrc
        echo "# Codespace: $SAFE_NAME" >> ~/.zshrc
        echo "[ -f \"$alias_file\" ] && source \"$alias_file\"" >> ~/.zshrc
    fi
}

# Display success message
display_success_message() {
    local ports="$1"
    local password="$2"
    local vs_code_port=$(echo "$ports" | cut -d',' -f1 | cut -d':' -f1)
    
    echo ""
    echo_success "🎉 Codespace created successfully!"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📦 Codespace: $SAFE_NAME"
    echo "🌐 VS Code URL: http://localhost:$vs_code_port"
    echo "🔑 Password: $password"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Quick commands:"
    echo "  mcs info $SAFE_NAME    - Show details"
    echo "  mcs logs $SAFE_NAME    - View logs"
    echo "  mcs exec $SAFE_NAME    - Enter container"
    echo "  mcs stop $SAFE_NAME    - Stop codespace"
    echo ""
    echo "To load aliases in current shell:"
    echo "  source ~/.zshrc"
    echo ""
}

# Main execution
main() {
    parse_arguments "$@"
    validate_inputs
    create_codespace
}

# Run main function
main "$@"