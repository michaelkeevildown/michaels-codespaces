#!/bin/bash

# Port Manager Module
# Handles port allocation and management for codespaces

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORT_REGISTRY="$HOME/codespaces/.port-registry"

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
fi

# Initialize port registry
init_port_registry() {
    if [ ! -f "$PORT_REGISTRY" ]; then
        echo "# Port Registry for Codespaces" > "$PORT_REGISTRY"
        echo "# Format: PORT|CODESPACE_NAME|SERVICE|ALLOCATED_DATE" >> "$PORT_REGISTRY"
    fi
}

# Check if port is in use (system-wide)
is_port_in_use() {
    local port="$1"
    
    # Check with netstat
    if command -v netstat >/dev/null 2>&1; then
        if netstat -tuln 2>/dev/null | grep -q ":$port "; then
            return 0
        fi
    fi
    
    # Check with lsof
    if command -v lsof >/dev/null 2>&1; then
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            return 0
        fi
    fi
    
    # Check with ss
    if command -v ss >/dev/null 2>&1; then
        if ss -tuln | grep -q ":$port "; then
            return 0
        fi
    fi
    
    return 1
}

# Check if port is registered
is_port_registered() {
    local port="$1"
    
    init_port_registry
    
    if grep -q "^$port|" "$PORT_REGISTRY" 2>/dev/null; then
        return 0
    fi
    
    return 1
}

# Register a port
register_port() {
    local port="$1"
    local codespace_name="$2"
    local service="${3:-unknown}"
    
    init_port_registry
    
    # Remove any existing registration for this port
    unregister_port "$port"
    
    # Add new registration
    echo "$port|$codespace_name|$service|$(date +%Y-%m-%d)" >> "$PORT_REGISTRY"
    
    echo_debug "Registered port $port for $codespace_name ($service)"
}

# Unregister a port
unregister_port() {
    local port="$1"
    
    init_port_registry
    
    # Remove port from registry
    if [ -f "$PORT_REGISTRY" ]; then
        grep -v "^$port|" "$PORT_REGISTRY" > "$PORT_REGISTRY.tmp" || true
        mv "$PORT_REGISTRY.tmp" "$PORT_REGISTRY"
    fi
}

# Unregister all ports for a codespace
unregister_codespace_ports() {
    local codespace_name="$1"
    
    init_port_registry
    
    if [ -f "$PORT_REGISTRY" ]; then
        grep -v "|$codespace_name|" "$PORT_REGISTRY" > "$PORT_REGISTRY.tmp" || true
        mv "$PORT_REGISTRY.tmp" "$PORT_REGISTRY"
    fi
}

# Find available port in range
find_available_port() {
    local start_port="${1:-8080}"
    local end_port="${2:-$((start_port + 100))}"
    local exclude_ports="${3:-}"
    
    local port=$start_port
    
    while [ $port -le $end_port ]; do
        # Check if port is excluded
        if [ -n "$exclude_ports" ] && echo "$exclude_ports" | grep -q "\\b$port\\b"; then
            ((port++))
            continue
        fi
        
        # Check if port is in use or registered
        if ! is_port_in_use $port && ! is_port_registered $port; then
            echo $port
            return 0
        fi
        
        ((port++))
    done
    
    return 1
}

# Allocate ports for a codespace
allocate_codespace_ports() {
    local codespace_name="$1"
    local num_ports="${2:-2}"
    local base_port="${3:-8080}"
    
    local allocated_ports=""
    local port_offset=0
    
    # Allocate VS Code port
    local vs_code_port=$(find_available_port $base_port $((base_port + 100)))
    if [ -z "$vs_code_port" ]; then
        echo_error "Failed to find available VS Code port"
        return 1
    fi
    register_port "$vs_code_port" "$codespace_name" "vscode"
    allocated_ports="$vs_code_port"
    
    # Allocate additional ports
    if [ $num_ports -gt 1 ]; then
        local app_base=$((base_port < 7000 ? 7680 : base_port + 1000))
        
        for i in $(seq 2 $num_ports); do
            local app_port=$(find_available_port $app_base $((app_base + 100)) "$allocated_ports")
            if [ -z "$app_port" ]; then
                echo_error "Failed to find available port for service $i"
                # Rollback allocations
                for port in $allocated_ports; do
                    unregister_port "$port"
                done
                return 1
            fi
            register_port "$app_port" "$codespace_name" "app$((i-1))"
            allocated_ports="$allocated_ports $app_port"
        done
    fi
    
    echo "$allocated_ports"
}

# List all registered ports
list_registered_ports() {
    init_port_registry
    
    if [ ! -s "$PORT_REGISTRY" ] || [ $(grep -v "^#" "$PORT_REGISTRY" | wc -l) -eq 0 ]; then
        echo_warning "No ports registered"
        return
    fi
    
    echo_info "Registered Ports:"
    echo "PORT  | CODESPACE                    | SERVICE | DATE"
    echo "------|------------------------------|---------|------------"
    
    grep -v "^#" "$PORT_REGISTRY" | while IFS='|' read -r port codespace service date; do
        printf "%-5s | %-28s | %-7s | %s\n" "$port" "$codespace" "$service" "$date"
    done
}

# Clean up stale port registrations
cleanup_stale_ports() {
    init_port_registry
    
    local temp_file=$(mktemp)
    echo "# Port Registry for Codespaces" > "$temp_file"
    echo "# Format: PORT|CODESPACE_NAME|SERVICE|ALLOCATED_DATE" >> "$temp_file"
    
    local cleaned=0
    
    grep -v "^#" "$PORT_REGISTRY" | while IFS='|' read -r port codespace service date; do
        # Check if codespace still exists
        if [ -d "$HOME/codespaces/$codespace" ]; then
            echo "$port|$codespace|$service|$date" >> "$temp_file"
        else
            echo_debug "Removing stale port registration: $port for $codespace"
            ((cleaned++))
        fi
    done
    
    mv "$temp_file" "$PORT_REGISTRY"
    
    if [ $cleaned -gt 0 ]; then
        echo_success "Cleaned up $cleaned stale port registrations"
    fi
}

# Export functions
export -f is_port_in_use
export -f is_port_registered
export -f register_port
export -f unregister_port
export -f unregister_codespace_ports
export -f find_available_port
export -f allocate_codespace_ports
export -f list_registered_ports
export -f cleanup_stale_ports