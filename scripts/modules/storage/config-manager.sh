#!/bin/bash

# Global Configuration Manager Module
# Manages system-wide settings for Michael's Codespaces

set -e

# Configuration directory and file
CONFIG_DIR="$HOME/.michaels-codespaces"
CONFIG_FILE="$CONFIG_DIR/config.json"

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
fi

# Initialize configuration directory and file
init_config() {
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        echo_debug "Created configuration directory: $CONFIG_DIR"
    fi
    
    if [ ! -f "$CONFIG_FILE" ]; then
        # Create default configuration
        cat > "$CONFIG_FILE" << EOF
{
  "host_ip": "localhost",
  "ip_mode": "localhost",
  "auto_detect_ip": false,
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "updated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
        chmod 600 "$CONFIG_FILE"
        echo_debug "Created default configuration file: $CONFIG_FILE"
    fi
}

# Get a configuration value
get_config_value() {
    local key="$1"
    local default_value="$2"
    
    init_config
    
    if command -v jq >/dev/null 2>&1; then
        local value=$(jq -r ".$key // null" "$CONFIG_FILE" 2>/dev/null)
        if [ "$value" != "null" ] && [ -n "$value" ]; then
            echo "$value"
        else
            echo "$default_value"
        fi
    else
        # Fallback to grep if jq is not available
        local value=$(grep "\"$key\":" "$CONFIG_FILE" 2>/dev/null | sed 's/.*: *"\(.*\)".*/\1/' | head -1)
        if [ -n "$value" ]; then
            echo "$value"
        else
            echo "$default_value"
        fi
    fi
}

# Set a configuration value
set_config_value() {
    local key="$1"
    local value="$2"
    
    init_config
    
    if command -v jq >/dev/null 2>&1; then
        # Use jq to update the configuration
        local temp_file=$(mktemp)
        jq ".$key = \"$value\" | .updated_at = \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"" "$CONFIG_FILE" > "$temp_file"
        mv "$temp_file" "$CONFIG_FILE"
        chmod 600 "$CONFIG_FILE"
    else
        # Fallback: recreate the file with updated value
        local host_ip=$(get_config_value "host_ip" "localhost")
        local ip_mode=$(get_config_value "ip_mode" "localhost")
        local auto_detect_ip=$(get_config_value "auto_detect_ip" "false")
        local created_at=$(get_config_value "created_at" "$(date -u +%Y-%m-%dT%H:%M:%SZ)")
        
        # Update the specific value
        case "$key" in
            "host_ip") host_ip="$value" ;;
            "ip_mode") ip_mode="$value" ;;
            "auto_detect_ip") auto_detect_ip="$value" ;;
        esac
        
        cat > "$CONFIG_FILE" << EOF
{
  "host_ip": "$host_ip",
  "ip_mode": "$ip_mode",
  "auto_detect_ip": $auto_detect_ip,
  "created_at": "$created_at",
  "updated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
        chmod 600 "$CONFIG_FILE"
    fi
    
    echo_debug "Set configuration: $key = $value"
}

# Get the configured host IP
get_host_ip() {
    get_config_value "host_ip" "localhost"
}

# Set the host IP
set_host_ip() {
    local ip="$1"
    set_config_value "host_ip" "$ip"
    echo_success "Host IP updated to: $ip"
}

# Get IP mode (localhost, auto, manual)
get_ip_mode() {
    get_config_value "ip_mode" "localhost"
}

# Set IP mode
set_ip_mode() {
    local mode="$1"
    set_config_value "ip_mode" "$mode"
}

# Check if auto-detect is enabled
is_auto_detect_enabled() {
    local auto_detect=$(get_config_value "auto_detect_ip" "false")
    [ "$auto_detect" = "true" ]
}

# Enable/disable auto-detect
set_auto_detect() {
    local enabled="$1"
    set_config_value "auto_detect_ip" "$enabled"
}

# Show current configuration
show_config() {
    init_config
    
    echo "Current Configuration:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Host IP: $(get_host_ip)"
    echo "IP Mode: $(get_ip_mode)"
    echo "Auto-detect: $(get_config_value "auto_detect_ip" "false")"
    echo "Config file: $CONFIG_FILE"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Export functions
export -f init_config
export -f get_config_value
export -f set_config_value
export -f get_host_ip
export -f set_host_ip
export -f get_ip_mode
export -f set_ip_mode
export -f is_auto_detect_enabled
export -f set_auto_detect
export -f show_config