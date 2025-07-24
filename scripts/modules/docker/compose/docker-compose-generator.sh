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

# Generate basic docker-compose configuration
generate_basic_compose() {
    # Use global config array - no parameters needed
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
    
    cat << EOF
version: '3.8'

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
    echo "$ports" | while IFS= read -r port; do
        echo "      - \"$port\""
    done
    
    # Add volumes
    echo "    volumes:"
    echo "      - ./src:/home/coder/project"
    echo "      - ./data:/home/coder/.local/share/code-server"
    echo "      - ./config:/home/coder/.config"
    echo "      - ./logs:/home/coder/logs"
    echo "      - \${HOME}/.ssh:/home/coder/.ssh:ro"
    echo "      - \${HOME}/codespaces/auth/tokens:/home/coder/.tokens:ro"
    
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
    echo "$labels" | while IFS= read -r label; do
        echo "      - \"$label\""
    done
    
    # Add command
    echo "    command: >"
    echo "      sh -c \""
    echo "        git config --global user.name '\$(cat /home/coder/.ssh/../codespaces/auth/git-config/name 2>/dev/null || echo '')' &&"
    echo "        git config --global user.email '\$(cat /home/coder/.ssh/../codespaces/auth/git-config/email 2>/dev/null || echo '')' &&"
    echo "        code-server --bind-addr 0.0.0.0:8080 --auth password"
    echo "      \""
    
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
    
    case "$language" in
        "node"|"nodejs")
            config[image]="${config[image]:-node:18-bullseye}"
            config[env_vars]="${config[env_vars]}\n      - NODE_ENV=development"
            config[volumes]="${config[volumes]}\n      - node_modules:/home/coder/project/node_modules"
            ;;
        "python")
            config[image]="${config[image]:-python:3.11-bullseye}"
            config[env_vars]="${config[env_vars]}\n      - PYTHONPATH=/home/coder/project"
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
    
    if [ ! -f "$compose_file" ]; then
        echo_error "Compose file not found: $compose_file"
        return 1
    fi
    
    # Basic validation with docker-compose
    if command -v docker-compose >/dev/null 2>&1; then
        if docker-compose -f "$compose_file" config >/dev/null 2>&1; then
            echo_success "Docker compose file is valid"
            return 0
        else
            echo_error "Docker compose file validation failed"
            return 1
        fi
    else
        echo_warning "docker-compose not found, skipping validation"
        return 0
    fi
}

# Export functions
export -f generate_basic_compose
export -f generate_language_compose
export -f generate_from_devcontainer
export -f validate_compose