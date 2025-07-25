#!/bin/bash

# Environment Manager Module
# Handles creation and management of environment files

# Create .env file for a codespace
create_env_file() {
    local codespace_dir="$1"
    local safe_name="$2"
    local repo_url="$3"
    local docker_image="$4"
    local language="$5"
    local ports="$6"
    local password="$7"
    local custom_env_file="$8"
    
    # Extract port numbers
    local vs_code_port=$(echo "$ports" | cut -d',' -f1 | cut -d':' -f1)
    local app_port=$(echo "$ports" | cut -d',' -f2 | cut -d':' -f1 2>/dev/null || echo "")
    
    # Get repository name
    local repo_name=$(basename "$repo_url" .git)
    
    # Get timezone
    local timezone=$(timedatectl show -p Timezone --value 2>/dev/null || echo "UTC")
    
    echo_info "Creating environment file..."
    
    cat > "$codespace_dir/.env" << EOF
# Codespace Configuration
CODESPACE_NAME=$safe_name
REPO_NAME=$repo_name
REPO_URL=$repo_url
CONTAINER_NAME=${safe_name}-dev
VS_CODE_PORT=$vs_code_port
APP_PORT=$app_port
PASSWORD=$password
DOCKER_IMAGE=$docker_image
LANGUAGE=$language
CREATED=$(date +%Y-%m-%d)

# User Configuration
USER=$USER
TZ=$timezone

# VS Code Configuration
CODE_SERVER_PASSWORD=$password
CODE_SERVER_AUTH=password
CODE_SERVER_BIND_ADDR=0.0.0.0:8080

# Development Environment
NODE_ENV=development
PYTHONUNBUFFERED=1
EOF

    # Add language-specific environment variables
    if [ -n "$language" ]; then
        echo "" >> "$codespace_dir/.env"
        echo "# Language-specific Configuration" >> "$codespace_dir/.env"
        
        case "$language" in
            node)
                echo "NPM_CONFIG_LOGLEVEL=warn" >> "$codespace_dir/.env"
                echo "NODE_OPTIONS=--max-old-space-size=4096" >> "$codespace_dir/.env"
                ;;
            python)
                echo "PYTHONDONTWRITEBYTECODE=1" >> "$codespace_dir/.env"
                echo "PIP_NO_CACHE_DIR=1" >> "$codespace_dir/.env"
                ;;
            go)
                echo "GO111MODULE=on" >> "$codespace_dir/.env"
                echo "GOPROXY=https://proxy.golang.org,direct" >> "$codespace_dir/.env"
                ;;
            rust)
                echo "RUST_LOG=info" >> "$codespace_dir/.env"
                echo "CARGO_HOME=/usr/local/cargo" >> "$codespace_dir/.env"
                ;;
            java)
                echo "MAVEN_OPTS=-Xmx1024m" >> "$codespace_dir/.env"
                echo "GRADLE_OPTS=-Xmx1024m" >> "$codespace_dir/.env"
                ;;
        esac
    fi

    # Append custom environment variables if provided
    if [ -n "$custom_env_file" ] && [ -f "$custom_env_file" ]; then
        echo "" >> "$codespace_dir/.env"
        echo "# Custom Environment Variables" >> "$codespace_dir/.env"
        cat "$custom_env_file" >> "$codespace_dir/.env"
        echo_info "Added custom environment variables from $custom_env_file"
    fi
    
    # Set proper permissions
    chmod 600 "$codespace_dir/.env"
    
    # Create .credentials file for easy recovery
    cat > "$codespace_dir/.credentials" << EOF
Michael's Codespaces Credentials
================================

Codespace: $safe_name
Repository: $repo_url
Created: $(date +%Y-%m-%d)

VS Code Access:
--------------
Port: $vs_code_port
Password: $password

To recover these credentials later, run:
  mcs recover $safe_name

To reset the password, run:
  mcs reset-password $safe_name
EOF
    
    # Set proper permissions for credentials file
    chmod 600 "$codespace_dir/.credentials"
    
    echo_success "Environment file created"
}

# Load environment variables from .env file
load_env_file() {
    local env_file="$1"
    
    if [ ! -f "$env_file" ]; then
        echo_error "Environment file not found: $env_file"
        return 1
    fi
    
    # Export variables from .env file
    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        [[ $key =~ ^#.*$ ]] || [[ -z "$key" ]] && continue
        
        # Remove leading/trailing whitespace
        key=$(echo "$key" | xargs)
        value=$(echo "$value" | xargs)
        
        # Export the variable
        export "$key=$value"
    done < "$env_file"
    
    echo_debug "Loaded environment from $env_file"
}

# Update a value in .env file
update_env_value() {
    local env_file="$1"
    local key="$2"
    local value="$3"
    
    if [ ! -f "$env_file" ]; then
        echo_error "Environment file not found: $env_file"
        return 1
    fi
    
    # Create temp file
    local temp_file=$(mktemp)
    
    # Update or add the key-value pair
    local found=false
    while IFS= read -r line; do
        if [[ "$line" =~ ^${key}= ]]; then
            echo "${key}=${value}" >> "$temp_file"
            found=true
        else
            echo "$line" >> "$temp_file"
        fi
    done < "$env_file"
    
    # Add key if not found
    if [ "$found" = false ]; then
        echo "${key}=${value}" >> "$temp_file"
    fi
    
    # Replace original file
    mv "$temp_file" "$env_file"
    chmod 600 "$env_file"
    
    echo_debug "Updated $key in $env_file"
}

# Get a value from .env file
get_env_value() {
    local env_file="$1"
    local key="$2"
    
    if [ ! -f "$env_file" ]; then
        return 1
    fi
    
    grep "^${key}=" "$env_file" 2>/dev/null | cut -d'=' -f2-
}

# Validate environment file
validate_env_file() {
    local env_file="$1"
    
    if [ ! -f "$env_file" ]; then
        echo_error "Environment file not found: $env_file"
        return 1
    fi
    
    # Check required fields
    local required_fields=(
        "CODESPACE_NAME"
        "REPO_URL"
        "CONTAINER_NAME"
        "VS_CODE_PORT"
        "PASSWORD"
        "DOCKER_IMAGE"
    )
    
    local missing_fields=()
    for field in "${required_fields[@]}"; do
        if ! grep -q "^${field}=" "$env_file"; then
            missing_fields+=("$field")
        fi
    done
    
    if [ ${#missing_fields[@]} -gt 0 ]; then
        echo_error "Missing required fields in environment file:"
        for field in "${missing_fields[@]}"; do
            echo "  - $field"
        done
        return 1
    fi
    
    echo_success "Environment file is valid"
    return 0
}

# Create backup of environment file
backup_env_file() {
    local env_file="$1"
    local backup_dir="${2:-$HOME/codespaces/backups}"
    
    if [ ! -f "$env_file" ]; then
        echo_error "Environment file not found: $env_file"
        return 1
    fi
    
    mkdir -p "$backup_dir"
    
    local codespace_name=$(basename "$(dirname "$env_file")")
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$backup_dir/${codespace_name}_env_${timestamp}.bak"
    
    cp "$env_file" "$backup_file"
    chmod 600 "$backup_file"
    
    echo_info "Environment backed up to: $backup_file"
}

# Export functions
export -f create_env_file
export -f load_env_file
export -f update_env_value
export -f get_env_value
export -f validate_env_file
export -f backup_env_file