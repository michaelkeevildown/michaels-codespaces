#!/bin/bash

# Docker Compose Generator Module
# Creates docker-compose.yml files for codespaces

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
fi

# Source logging utilities
if [ -f "$SCRIPT_DIR/../../utils/logging.sh" ]; then
    source "$SCRIPT_DIR/../../utils/logging.sh"
else
    # Fallback logging functions
    log_debug() { [ "${DEBUG:-0}" -eq 1 ] && echo "🔍 $1"; }
    log_info() { echo "ℹ️  $1"; }
    log_error() { echo "❌ $1"; }
    log_file_content() { :; }
    log_variables() { :; }
fi

# Generate basic docker-compose configuration
generate_basic_compose() {
    # Use global config array - no parameters needed
    log_info "Generating basic docker-compose configuration"
    
    # Extract configuration
    local container_name="${config[container_name]}"
    local image="${config[image]:-codercom/code-server:latest}"
    local password="${config[password]}"
    local ports="${config[ports]}"
    local env_vars="${config[env_vars]:-}"
    local volumes="${config[volumes]:-}"
    local networks="${config[networks]:-}"
    local labels="${config[labels]:-}"
    local healthcheck="${config[healthcheck]:-true}"
    local components="${config[components]:-}"
    local init_script="${config[init_script]:-}"
    
    # Extract codespace name from container name (remove -dev suffix)
    local codespace_name="${container_name%-dev}"
    
    # Log configuration values
    log_variables "Docker Compose Config" container_name image password ports env_vars volumes networks labels healthcheck components
    
    cat << EOF
services:
  ${container_name}:
    image: ${image}
    container_name: ${container_name}
    restart: unless-stopped
    environment:
      - PASSWORD=${password}
      - TZ=\${TZ:-UTC}
      - DOCKER_USER=\${USER}
EOF

    # Add custom environment variables
    if [ -n "$env_vars" ]; then
        echo "$env_vars" | while IFS= read -r env_var; do
            echo "      - $env_var"
        done
    fi
    
    # Add ports
    echo "    ports:"
    if [ -n "$ports" ]; then
        # Split comma-separated ports and add each one
        IFS=',' read -ra PORT_ARRAY <<< "$ports"
        for port in "${PORT_ARRAY[@]}"; do
            # Trim whitespace
            port=$(echo "$port" | xargs)
            if [ -n "$port" ]; then
                echo "      - \"$port\""
                log_debug "Added port mapping: $port"
            fi
        done
    else
        log_warning "No ports specified in configuration"
        echo "      - \"8080:8080\""
    fi
    
    # Add volumes
    echo "    volumes:"
    echo "      - ./src:/home/coder/${codespace_name}"
    echo "      - ./data:/home/coder/.local/share/code-server"
    echo "      - ./config:/home/coder/.config"
    echo "      - ./logs:/home/coder/logs"
    echo "      - \${HOME}/.ssh:/home/coder/.ssh:ro"
    echo "      - \${HOME}/codespaces/auth/tokens:/home/coder/.tokens:ro"
    
    # Add component-related volumes if components are enabled
    if [ -n "$components" ]; then
        echo "      - ./components:/opt/codespace/components:ro"
        echo "      - ./init:/home/coder/.codespace-init"
    fi
    
    if [ -n "$volumes" ]; then
        echo "$volumes" | while IFS= read -r volume; do
            echo "      - $volume"
        done
    fi
    
    # Add networks
    echo "    networks:"
    echo "      - codespace-network"
    if [ -n "$networks" ]; then
        echo "$networks" | while IFS= read -r network; do
            echo "      - $network"
        done
    fi
    
    # Add labels
    echo "    labels:"
    if [ -n "$labels" ]; then
        # Convert \n to actual newlines and process each label
        echo -e "$labels" | while IFS= read -r label; do
            if [ -n "$label" ]; then
                echo "      - \"$label\""
                log_debug "Added label: $label"
            fi
        done
    else
        log_debug "No labels specified"
    fi
    
    # Add command or entrypoint based on components
    if [ -n "$components" ] && [ -n "$init_script" ]; then
        # Use custom entrypoint that runs init script then starts code-server
        echo "    entrypoint: [\"/bin/bash\", \"-c\"]"
        echo "    command: [\"bash /opt/codespace/components/init.sh && exec code-server --bind-addr 0.0.0.0:8080 --auth password /home/coder/${codespace_name}\"]"
    else
        # Default code-server command
        echo "    command: [\"--bind-addr\", \"0.0.0.0:8080\", \"--auth\", \"password\", \"/home/coder/${codespace_name}\"]"
    fi
    
    # Add healthcheck if enabled
    if [ "$healthcheck" == "true" ]; then
        cat << EOF
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
EOF
    fi
    
    # Add networks section
    cat << EOF

networks:
  codespace-network:
    name: ${container_name%-dev}-network
    driver: bridge
EOF
}

# Generate compose file for specific languages/frameworks
generate_language_compose() {
    local language="$1"
    
    # Extract codespace name from container name (remove -dev suffix)
    local container_name="${config[container_name]}"
    local codespace_name="${container_name%-dev}"
    
    case "$language" in
        "node"|"nodejs")
            config[image]="${config[image]:-node:18-bullseye}"
            config[env_vars]="${config[env_vars]}\n      - NODE_ENV=development"
            config[volumes]="${config[volumes]}\n      - node_modules:/home/coder/${codespace_name}/node_modules"
            ;;
        "python")
            config[image]="${config[image]:-python:3.11-bullseye}"
            config[env_vars]="${config[env_vars]}\n      - PYTHONPATH=/home/coder/${codespace_name}"
            config[volumes]="${config[volumes]}\n      - pip_cache:/home/coder/.cache/pip"
            ;;
        "go"|"golang")
            config[image]="${config[image]:-golang:1.21-bullseye}"
            config[env_vars]="${config[env_vars]}\n      - GOPATH=/home/coder/go"
            config[volumes]="${config[volumes]}\n      - go_modules:/go/pkg/mod"
            ;;
        "rust")
            config[image]="${config[image]:-rust:latest}"
            config[env_vars]="${config[env_vars]}\n      - CARGO_HOME=/home/coder/.cargo"
            config[volumes]="${config[volumes]}\n      - cargo_cache:/home/coder/.cargo"
            ;;
        "java")
            config[image]="${config[image]:-openjdk:17-bullseye}"
            config[env_vars]="${config[env_vars]}\n      - JAVA_HOME=/usr/local/openjdk-17"
            config[volumes]="${config[volumes]}\n      - maven_cache:/home/coder/.m2"
            ;;
    esac
    
    generate_basic_compose
    
    # Add volume definitions if needed
    case "$language" in
        "node"|"nodejs")
            echo -e "\nvolumes:\n  node_modules:"
            ;;
        "python")
            echo -e "\nvolumes:\n  pip_cache:"
            ;;
        "go"|"golang")
            echo -e "\nvolumes:\n  go_modules:"
            ;;
        "rust")
            echo -e "\nvolumes:\n  cargo_cache:"
            ;;
        "java")
            echo -e "\nvolumes:\n  maven_cache:"
            ;;
    esac
}

# Generate compose from .devcontainer.json
generate_from_devcontainer() {
    local devcontainer_file="$1"
    
    if [ ! -f "$devcontainer_file" ]; then
        echo_error "devcontainer file not found: $devcontainer_file"
        return 1
    fi
    
    # Parse image from devcontainer
    local image=$(parse_devcontainer_image "$devcontainer_file")
    if [ -n "$image" ]; then
        config[image]="$image"
    fi
    
    # Parse other settings (simplified - in production use jq)
    # Add more parsing as needed
    
    generate_basic_compose
}

# Validate compose file
validate_compose() {
    local compose_file="$1"
    
    log_info "Validating docker-compose file: $compose_file"
    
    if [ ! -f "$compose_file" ]; then
        echo_error "Compose file not found: $compose_file"
        log_error "Compose file not found: $compose_file"
        return 1
    fi
    
    # Log the generated compose file content
    log_file_content "$compose_file" "Generated docker-compose.yml"
    
    # Basic validation with docker-compose
    if command -v docker-compose >/dev/null 2>&1; then
        log_info "Running docker-compose validation"
        
        # Capture both stdout and stderr
        local validation_output
        local validation_exit_code
        
        if validation_output=$(docker-compose -f "$compose_file" config 2>&1); then
            validation_exit_code=0
            log_info "Docker compose validation successful"
            log_debug "Validation output: $validation_output"
            echo_success "Docker compose file is valid"
            return 0
        else
            validation_exit_code=$?
            log_error "Docker compose validation failed (exit code: $validation_exit_code)"
            log_error "Validation error output: $validation_output"
            echo_error "Docker compose file validation failed"
            echo_error "Check log file for details: $(get_log_file)"
            return 1
        fi
    else
        log_warning "docker-compose command not found, skipping validation"
        echo_warning "docker-compose not found, skipping validation"
        return 0
    fi
}

# Generate compose with component support
generate_compose_with_components() {
    local components="$1"
    
    # Add components to config
    config[components]="$components"
    config[init_script]="/opt/codespace/components/init.sh"
    
    # Add component environment variables
    if [ -n "$components" ]; then
        config[env_vars]="${config[env_vars]}\n      - CODESPACE_COMPONENTS=$components"
        config[env_vars]="${config[env_vars]}\n      - CODESPACE_INIT_LOG=/home/coder/.codespace-init/init.log"
    fi
    
    # Generate based on language or basic
    if [ -n "${config[language]}" ]; then
        generate_language_compose "${config[language]}"
    else
        generate_basic_compose
    fi
}

# Export functions
export -f generate_basic_compose
export -f generate_language_compose
export -f generate_from_devcontainer
export -f generate_compose_with_components
export -f validate_compose