#!/bin/bash

# Enhanced Codespace Creation Script
# Uses modular architecture for better maintainability

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source path configuration first
if [ -f "$SCRIPT_DIR/../utils/paths.sh" ]; then
    source "$SCRIPT_DIR/../utils/paths.sh"
else
    echo "ERROR: Path configuration not found" >&2
    exit 1
fi

# Verify paths are set up correctly
if ! verify_paths; then
    echo "ERROR: Path verification failed. Please run install-mcs.sh" >&2
    exit 1
fi

MODULES_DIR="$MCS_MODULES_DIR"

# Source all required modules
source "$MODULES_DIR/github/auth/github-auth.sh"
source "$MODULES_DIR/github/clone/github-clone.sh"
source "$MODULES_DIR/github/api/repo-validator.sh"
source "$MODULES_DIR/docker/compose/docker-compose-generator.sh"
source "$MODULES_DIR/docker/images/language-detector.sh"
source "$MODULES_DIR/docker/containers/container-manager.sh"
source "$MODULES_DIR/docker/containers/management-scripts.sh"
source "$MODULES_DIR/networking/port-manager.sh"
source "$MODULES_DIR/storage/env-manager.sh"
source "$MODULES_DIR/components/registry.sh"
source "$MODULES_DIR/components/ascii-selector.sh"
source "$MODULES_DIR/components/whiptail-selector.sh"
source "$MODULES_DIR/components/interactive-selector.sh"
source "$MODULES_DIR/components/simple-selector.sh"
source "$MODULES_DIR/components/manifest-generator.sh"

# Source utilities using path module
if ! source_utility "colors"; then
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
    echo_step() { echo "▶️  $1"; }
    echo_debug() { [ "${DEBUG:-0}" -eq 1 ] && echo "🔍 $1"; }
fi

# Source logging utilities
if [ -f "$MCS_UTILS_DIR/logging.sh" ]; then
    source "$MCS_UTILS_DIR/logging.sh"
    # Initialize logging
    init_logging
else
    # Fallback logging functions
    log_debug() { [ "${DEBUG:-0}" -eq 1 ] && echo "🔍 $1"; }
    log_info() { echo "ℹ️  $1"; }
    log_error() { echo "❌ $1"; }
    log_variables() { :; }
    log_command() { eval "$1"; }
    get_log_file() { echo "No log file available"; }
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
  --components <list> Comma-separated list of components to install
  --preset <name>   Use a component preset (ai-dev, full-stack, minimal)
  --interactive     Force interactive component selection (default)
  --no-interactive  Skip component selection
  --no-start        Don't start the container after creation
  --force           Overwrite existing codespace
  --shallow         Use shallow clone (depth=1) for faster cloning
  --depth <n>       Clone depth (1=shallow, 0=full history, default: auto-detect)
  --debug           Enable debug output
  --verbose         Enable verbose logging with real-time output

Examples:
  $0 git@github.com:facebook/react.git
  $0 https://github.com/nodejs/node.git --language node
  $0 git@github.com:user/repo.git --name my-project --ports "8090:8080"
  $0 https://github.com/homebrew/homebrew-core.git --shallow
  $0 git@github.com:torvalds/linux.git --depth 10
  $0 https://github.com/user/repo.git --no-interactive
  $0 git@github.com:user/repo.git --components github-cli,claude,claude-flow
  $0 https://github.com/user/repo.git --preset ai-dev

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
    FORCE_SHALLOW=false
    CLONE_DEPTH=""
    COMPONENTS=""
    PRESET=""
    INTERACTIVE=""
    
    # Check for help
    if [ $# -eq 0 ] || [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
        usage
        exit 0
    fi
    
    # Parse all arguments - repository URL can be anywhere
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
            --shallow)
                FORCE_SHALLOW=true
                shift
                ;;
            --depth)
                CLONE_DEPTH="$2"
                if ! [[ "$CLONE_DEPTH" =~ ^[0-9]+$ ]]; then
                    echo_error "Clone depth must be a number: $CLONE_DEPTH"
                    exit 1
                fi
                shift 2
                ;;
            --debug)
                export DEBUG=1
                shift
                ;;
            --verbose)
                export VERBOSE=1
                shift
                ;;
            --components)
                COMPONENTS="$2"
                shift 2
                ;;
            --preset)
                PRESET="$2"
                shift 2
                ;;
            --interactive)
                INTERACTIVE=true
                shift
                ;;
            --no-interactive)
                INTERACTIVE=false
                shift
                ;;
            git@*|https://github.com/*|http://github.com/*|https://gitlab.com/*|http://gitlab.com/*|*.git)
                # This looks like a repository URL
                if [ -n "$REPO_URL" ]; then
                    echo_error "Multiple repository URLs specified: '$REPO_URL' and '$1'"
                    exit 1
                fi
                REPO_URL="$1"
                shift
                ;;
            *)
                echo_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    # Validate that repository URL was provided
    if [ -z "$REPO_URL" ]; then
        echo_error "Repository URL is required"
        usage
        exit 1
    fi
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
        echo "Directory: $CODESPACE_DIR"
        echo ""
        
        # Check container status if container management is available
        if command -v get_container_status >/dev/null 2>&1; then
            local status=$(get_container_status "$CODESPACE_DIR")
            case "$status" in
                "running")
                    echo_info "Status: Container is currently running"
                    ;;
                "stopped")
                    echo_warning "Status: Container exists but is stopped"
                    ;;
                "not-found")
                    echo_warning "Status: Directory exists but no container configuration found"
                    ;;
            esac
            echo ""
        fi
        
        echo "Options:"
        echo "  • Use --force to remove and recreate: mcs create --force $REPO_URL"
        echo "  • Choose a different name: mcs create --name <new-name> $REPO_URL"
        echo "  • Remove manually: mcs remove $SAFE_NAME"
        echo "  • View existing codespaces: mcs list"
        exit 1
    fi
}

# Main creation process
create_codespace() {
    echo ""
    echo_step "🚀 Creating codespace: $SAFE_NAME"
    echo "Repository: $REPO_URL"
    echo ""
    
    log_info "Starting codespace creation: $SAFE_NAME"
    log_info "Repository URL: $REPO_URL"
    log_variables "Initial Config" SAFE_NAME REPO_URL CODESPACE_DIR DOCKER_IMAGE LANGUAGE CUSTOM_PORTS FORCE NO_START COMPONENTS PRESET INTERACTIVE
    
    # Handle component selection
    local selected_components=""
    
    # If no component-related flags were provided, default to interactive mode
    if [ -z "$COMPONENTS" ] && [ -z "$PRESET" ] && [ "$INTERACTIVE" != "false" ]; then
        INTERACTIVE=true
    fi
    
    if [ "$INTERACTIVE" == "true" ]; then
        echo ""
        echo_info "Component selection..."
        echo ""
        
        # EXTENSIVE DEBUG LOGGING
        echo_debug "=== START COMPONENT SELECTION DEBUG ==="
        echo_debug "Current directory: $(pwd)"
        echo_debug "SCRIPT_DIR: ${SCRIPT_DIR}"
        echo_debug "MCS_MODULES_DIR: ${MCS_MODULES_DIR}"
        
        # Debug: Check if functions are available
        echo_debug "Checking available component selection methods..."
        echo_debug "check_ascii_select available: $(type -t check_ascii_select || echo 'not found')"
        echo_debug "ascii_component_selection available: $(type -t ascii_component_selection || echo 'not found')"
        echo_debug "ascii_component_select available: $(type -t ascii_component_select || echo 'not found')"
        echo_debug "ascii_select available: $(type -t ascii_select || echo 'not found')"
        echo_debug "check_whiptail available: $(type -t check_whiptail || echo 'not found')"
        echo_debug "interactive_select available: $(type -t interactive_select || echo 'not found')"
        echo_debug "simple_select available: $(type -t simple_select || echo 'not found')"
        echo_debug "list_components available: $(type -t list_components || echo 'not found')"
        echo_debug "register_components available: $(type -t register_components || echo 'not found')"
        
        # Test component registry
        echo_debug "Testing component registry..."
        echo_debug "Calling list_components..."
        local test_components=$(list_components 2>&1)
        echo_debug "list_components output: '$test_components'"
        
        # Additional debug: test which selector will be used
        echo_debug "Testing selector availability..."
        if check_ascii_select 2>/dev/null; then
            echo_debug "ASCII selector check passed"
        else
            echo_debug "ASCII selector check failed: $(check_ascii_select 2>&1 || echo $?)"
        fi
        
        if check_whiptail 2>/dev/null; then
            echo_debug "Whiptail check passed"
        else
            echo_debug "Whiptail check failed"
        fi
        
        # Try ASCII selector first
        echo_debug "Attempting ASCII selector for component selection"
        echo_debug "Calling ascii_component_selection..."
        
        # Try ASCII selection with error handling
        local selection_output
        local selection_error
        if selection_output=$(ascii_component_selection 2>&1); then
            echo_debug "ASCII selection returned successfully"
            echo_debug "Selection output: '$selection_output'"
            selected_components="$selection_output"
            echo_debug "ASCII selection successful: $selected_components"
        else
            local exit_code=$?
            echo_debug "ASCII selector failed with exit code: $exit_code"
            echo_debug "Selection output/error: '$selection_output'"
            echo_info "Falling back to simple selection..."
            
            # Fall back to simple selection
            echo_debug "Calling simple_select..."
            if selected_components=$(simple_select 2>&1); then
                echo_debug "Simple select successful: '$selected_components'"
            else
                echo_debug "Simple select failed: $?"
                echo_warning "Component selection cancelled, continuing without components"
                selected_components=""
            fi
        fi
        
        echo_debug "Final selected_components: '$selected_components'"
        echo_debug "=== END COMPONENT SELECTION DEBUG ==="
    elif [ -n "$PRESET" ]; then
        echo_info "Loading preset: $PRESET"
        selected_components=$(load_preset "$PRESET") || {
            echo_error "Failed to load preset: $PRESET"
            exit 1
        }
    elif [ -n "$COMPONENTS" ]; then
        # Convert comma-separated to space-separated
        selected_components=$(echo "$COMPONENTS" | tr ',' ' ')
    fi
    
    # Validate selected components
    if [ -n "$selected_components" ]; then
        echo_info "Selected components: $selected_components"
        if ! validate_components $selected_components; then
            echo_error "Invalid components selected"
            exit 1
        fi
    fi
    
    # Clean up if forcing
    if [ -d "$CODESPACE_DIR" ] && [ "$FORCE" == "true" ]; then
        echo_warning "Removing existing codespace..."
        log_info "Force flag enabled, removing existing codespace"
        
        # Change to a safe directory first to avoid "current working directory" errors
        cd "$HOME" || cd /tmp
        
        if [ -f "$CODESPACE_DIR/docker-compose.yml" ]; then
            log_command "docker-compose -f \"$CODESPACE_DIR/docker-compose.yml\" down" "Stopping existing containers"
        fi
        unregister_codespace_ports "$SAFE_NAME"
        rm -rf "$CODESPACE_DIR"
        log_info "Existing codespace removed"
    fi
    
    # Create directory structure
    echo_info "Creating directory structure..."
    mkdir -p "$CODESPACE_DIR"/{src,data,config,logs}
    
    # Create component directories if components are selected
    if [ -n "$selected_components" ]; then
        mkdir -p "$CODESPACE_DIR"/{components,init}
    fi
    
    # Clone repository
    echo_info "Cloning repository..."
    if ! clone_with_retry "$REPO_URL" "$CODESPACE_DIR/src" 3 5 "" "$CLONE_DEPTH" "$FORCE_SHALLOW"; then
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
        LANGUAGE=$(detect_language "$CODESPACE_DIR/src")
        if [ -n "$LANGUAGE" ]; then
            DOCKER_IMAGE=$(get_language_image "$LANGUAGE")
        fi
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
        log_info "Auto-allocating ports for codespace: $SAFE_NAME"
        local allocated=$(allocate_codespace_ports "$SAFE_NAME" 2)
        log_debug "Port allocation result: '$allocated'"
        
        if [ -z "$allocated" ]; then
            log_error "Port allocation returned empty result"
            echo_error "Failed to allocate ports"
            rm -rf "$CODESPACE_DIR"
            exit 1
        fi
        
        local vs_code_port=$(echo "$allocated" | cut -d' ' -f1)
        local app_port=$(echo "$allocated" | cut -d' ' -f2)
        
        log_debug "Extracted ports - VS Code: '$vs_code_port', App: '$app_port'"
        
        # Validate that ports are valid numbers
        if [[ ! "$vs_code_port" =~ ^[0-9]+$ ]] || [ -z "$vs_code_port" ]; then
            log_error "Invalid VS Code port: '$vs_code_port'"
            echo_error "Failed to allocate valid VS Code port"
            vs_code_port="8080"
            log_info "Using fallback VS Code port: $vs_code_port"
        fi
        
        if [[ ! "$app_port" =~ ^[0-9]+$ ]] || [ -z "$app_port" ]; then
            log_error "Invalid app port: '$app_port'"
            echo_error "Failed to allocate valid app port"
            app_port="3000"
            log_info "Using fallback app port: $app_port"
        fi
        
        ports="${vs_code_port}:8080,${app_port}:3000"
        log_info "Final port configuration: $ports"
    fi
    
    echo_debug "Port configuration: $ports"
    
    # Generate password
    local password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-16 2>/dev/null || \
                    tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16)
    
    # Setup components if selected
    if [ -n "$selected_components" ]; then
        echo_info "Setting up components..."
        
        # Create component package
        create_init_package "$CODESPACE_DIR/components" $selected_components
        
        # Create manifest for container
        generate_manifest "$CODESPACE_DIR/init/components.manifest" $selected_components
        
        # Make component list available to docker-compose
        config[components]="$selected_components"
    fi
    
    # Create environment file
    create_env_file "$CODESPACE_DIR" "$SAFE_NAME" "$REPO_URL" "$DOCKER_IMAGE" "$LANGUAGE" "$ports" "$password" "$ENV_FILE"
    
    # Generate docker-compose.yml
    echo_info "Generating Docker configuration..."
    generate_docker_compose "$CODESPACE_DIR" "$ports" "$password" "$selected_components"
    
    # Create management scripts
    create_management_scripts "$CODESPACE_DIR" "$SAFE_NAME" "$REPO_URL" "$DOCKER_IMAGE" "$LANGUAGE" "$ports" "$password"
    
    # Start container if requested
    if [ "$NO_START" != "true" ]; then
        if ! start_container "$CODESPACE_DIR"; then
            echo_error "Failed to start codespace"
            exit 1
        fi
        
        # Verify persistence and repository accessibility
        if ! verify_container_persistence "$CODESPACE_DIR"; then
            echo_warning "Repository may not be properly accessible in VS Code"
            echo_info "Try restarting the container with: mcs restart $SAFE_NAME"
        fi
    fi
    
    # Display success message
    display_codespace_success "$SAFE_NAME" "$ports" "$password"
}

# Generate docker-compose.yml
generate_docker_compose() {
    local codespace_dir="$1"
    local ports="$2"
    local password="$3"
    local components="$4"
    
    # Prepare configuration
    declare -A config
    config[container_name]="${SAFE_NAME}-dev"
    config[image]="$DOCKER_IMAGE"
    config[password]="$password"
    config[ports]="$ports"
    config[env_vars]=""
    config[labels]="codespace.repo=$(basename "$REPO_URL" .git)\ncodespace.created=$(date +%Y-%m-%d)\ncodespace.url=$REPO_URL"
    config[language]="$LANGUAGE"
    
    # Generate based on components or language
    if [ -n "$components" ]; then
        generate_compose_with_components "$components" > "$codespace_dir/docker-compose.yml"
    elif [ -n "$LANGUAGE" ]; then
        generate_language_compose "$LANGUAGE" > "$codespace_dir/docker-compose.yml"
    else
        generate_basic_compose > "$codespace_dir/docker-compose.yml"
    fi
    
    # Validate the generated file
    validate_compose "$codespace_dir/docker-compose.yml"
}

# Main execution
main() {
    parse_arguments "$@"
    validate_inputs
    create_codespace
}

# Run main function
main "$@"