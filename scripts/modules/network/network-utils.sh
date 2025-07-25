#!/bin/bash

# Network Utilities Module
# Provides IP detection and validation functions

set -e

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
    echo_debug() { [ "${DEBUG:-0}" -eq 1 ] && echo "🔍 $1"; }
fi

# Source config manager
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/../storage/config-manager.sh" ]; then
    source "$SCRIPT_DIR/../storage/config-manager.sh"
fi

# Detect local/private IP address
detect_local_ip() {
    local ip=""
    
    # Try different methods to get local IP
    if command -v hostname >/dev/null 2>&1; then
        # Method 1: hostname command (most common)
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi
    
    if [ -z "$ip" ] && command -v ip >/dev/null 2>&1; then
        # Method 2: ip command
        ip=$(ip -4 addr show scope global | grep inet | awk '{print $2}' | cut -d/ -f1 | head -1)
    fi
    
    if [ -z "$ip" ] && command -v ifconfig >/dev/null 2>&1; then
        # Method 3: ifconfig (older systems)
        ip=$(ifconfig | grep -E 'inet[[:space:]]' | grep -v '127.0.0.1' | awk '{print $2}' | head -1)
    fi
    
    # Fallback to localhost if no IP found
    if [ -z "$ip" ]; then
        ip="localhost"
    fi
    
    echo "$ip"
}

# Detect public/external IP address
detect_public_ip() {
    local ip=""
    local timeout=5
    
    echo_debug "Detecting public IP address..."
    
    # Try multiple services for redundancy
    local services=(
        "https://ipinfo.io/ip"
        "https://api.ipify.org"
        "https://checkip.amazonaws.com"
        "https://ifconfig.me"
    )
    
    for service in "${services[@]}"; do
        ip=$(curl -s --max-time $timeout "$service" 2>/dev/null | tr -d '\n' | grep -E '^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$')
        if [ -n "$ip" ] && validate_ip_address "$ip"; then
            echo_debug "Got public IP from $service: $ip"
            echo "$ip"
            return 0
        fi
    done
    
    echo_warning "Could not detect public IP address"
    return 1
}

# Validate IP address format
validate_ip_address() {
    local ip="$1"
    
    if [ -z "$ip" ]; then
        return 1
    fi
    
    # Check for localhost or common local addresses
    if [ "$ip" = "localhost" ] || [ "$ip" = "127.0.0.1" ] || [ "$ip" = "::1" ]; then
        return 0
    fi
    
    # Validate IPv4 format
    if echo "$ip" | grep -qE '^([0-9]{1,3}\.){3}[0-9]{1,3}$'; then
        # Check each octet is <= 255
        local valid=true
        IFS='.' read -ra octets <<< "$ip"
        for octet in "${octets[@]}"; do
            if [ "$octet" -gt 255 ]; then
                valid=false
                break
            fi
        done
        
        if [ "$valid" = true ]; then
            return 0
        fi
    fi
    
    # Basic IPv6 validation
    if echo "$ip" | grep -qE '^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$'; then
        return 0
    fi
    
    return 1
}

# Get the appropriate host IP based on configuration
get_access_ip() {
    local mode=$(get_ip_mode 2>/dev/null || echo "localhost")
    local ip=""
    
    case "$mode" in
        "localhost")
            ip="localhost"
            ;;
        "auto")
            # Auto-detect local IP
            ip=$(detect_local_ip)
            ;;
        "public")
            # Try to detect public IP
            ip=$(detect_public_ip) || ip=$(detect_local_ip)
            ;;
        "manual")
            # Use configured IP
            ip=$(get_host_ip 2>/dev/null || echo "localhost")
            ;;
        *)
            ip="localhost"
            ;;
    esac
    
    echo "$ip"
}

# Get formatted URL with port
get_access_url() {
    local port="$1"
    local ip=$(get_access_ip)
    
    echo "http://${ip}:${port}"
}

# Update IP configuration interactively
configure_ip_interactive() {
    echo ""
    echo "IP Address Configuration"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Select IP address mode:"
    echo "  1) localhost (default)"
    echo "  2) Auto-detect local IP"
    echo "  3) Auto-detect public IP"
    echo "  4) Manual IP address"
    echo ""
    
    read -p "Choice [1-4]: " choice
    
    case "$choice" in
        1)
            set_ip_mode "localhost"
            set_host_ip "localhost"
            echo_success "Set to use localhost"
            ;;
        2)
            set_ip_mode "auto"
            local ip=$(detect_local_ip)
            set_host_ip "$ip"
            echo_success "Set to auto-detect local IP: $ip"
            ;;
        3)
            set_ip_mode "public"
            local ip=$(detect_public_ip)
            if [ -n "$ip" ]; then
                set_host_ip "$ip"
                echo_success "Set to use public IP: $ip"
            else
                echo_error "Could not detect public IP, falling back to local IP"
                ip=$(detect_local_ip)
                set_host_ip "$ip"
                set_ip_mode "auto"
            fi
            ;;
        4)
            read -p "Enter IP address or hostname: " custom_ip
            if validate_ip_address "$custom_ip"; then
                set_ip_mode "manual"
                set_host_ip "$custom_ip"
                echo_success "Set to use manual IP: $custom_ip"
            else
                echo_error "Invalid IP address format"
                return 1
            fi
            ;;
        *)
            echo_warning "Invalid choice, keeping current configuration"
            ;;
    esac
    
    echo ""
    show_config
}

# Update IP configuration with command line options
update_ip_config() {
    local option="$1"
    local value="$2"
    
    case "$option" in
        "--auto"|"-a")
            set_ip_mode "auto"
            local ip=$(detect_local_ip)
            set_host_ip "$ip"
            echo_success "Auto-detected local IP: $ip"
            ;;
        "--public"|"-p")
            set_ip_mode "public"
            local ip=$(detect_public_ip)
            if [ -n "$ip" ]; then
                set_host_ip "$ip"
                echo_success "Auto-detected public IP: $ip"
            else
                echo_error "Could not detect public IP"
                return 1
            fi
            ;;
        "--localhost"|"-l")
            set_ip_mode "localhost"
            set_host_ip "localhost"
            echo_success "Set to use localhost"
            ;;
        "--ip"|"-i")
            if [ -z "$value" ]; then
                echo_error "IP address required"
                return 1
            fi
            if validate_ip_address "$value"; then
                set_ip_mode "manual"
                set_host_ip "$value"
                echo_success "Set to use IP: $value"
            else
                echo_error "Invalid IP address: $value"
                return 1
            fi
            ;;
        *)
            echo_error "Unknown option: $option"
            return 1
            ;;
    esac
}

# Show detected IPs
show_detected_ips() {
    echo "Network Information:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    local local_ip=$(detect_local_ip)
    echo "Local IP: $local_ip"
    
    echo -n "Public IP: "
    local public_ip=$(detect_public_ip)
    if [ -n "$public_ip" ]; then
        echo "$public_ip"
    else
        echo "Could not detect"
    fi
    
    echo "Current configured IP: $(get_access_ip)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Export functions
export -f detect_local_ip
export -f detect_public_ip
export -f validate_ip_address
export -f get_access_ip
export -f get_access_url
export -f configure_ip_interactive
export -f update_ip_config
export -f show_detected_ips