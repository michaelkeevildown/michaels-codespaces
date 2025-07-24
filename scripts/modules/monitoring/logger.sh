#!/bin/bash

# Logger Module
# Provides structured logging for codespace operations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${LOG_DIR:-$HOME/codespaces/.logs}"
LOG_LEVEL="${LOG_LEVEL:-INFO}"

# Create log directory
mkdir -p "$LOG_DIR"

# Log levels
declare -A LOG_LEVELS=(
    [DEBUG]=0
    [INFO]=1
    [WARNING]=2
    [ERROR]=3
    [CRITICAL]=4
)

# Get current log level value
get_log_level_value() {
    echo "${LOG_LEVELS[$LOG_LEVEL]:-1}"
}

# Log message
log_message() {
    local level="$1"
    local message="$2"
    local component="${3:-general}"
    
    # Check if we should log this level
    local level_value="${LOG_LEVELS[$level]:-1}"
    local current_level_value=$(get_log_level_value)
    
    if [ $level_value -lt $current_level_value ]; then
        return
    fi
    
    # Format timestamp
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Log to file
    local log_file="$LOG_DIR/codespaces.log"
    echo "[$timestamp] [$level] [$component] $message" >> "$log_file"
    
    # Also log to component-specific file
    local component_log="$LOG_DIR/${component}.log"
    echo "[$timestamp] [$level] $message" >> "$component_log"
    
    # Rotate logs if too large (>10MB)
    for file in "$log_file" "$component_log"; do
        if [ -f "$file" ] && [ $(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null) -gt 10485760 ]; then
            mv "$file" "${file}.old"
            touch "$file"
        fi
    done
}

# Logging functions
log_debug() {
    log_message "DEBUG" "$1" "${2:-general}"
}

log_info() {
    log_message "INFO" "$1" "${2:-general}"
}

log_warning() {
    log_message "WARNING" "$1" "${2:-general}"
}

log_error() {
    log_message "ERROR" "$1" "${2:-general}"
}

log_critical() {
    log_message "CRITICAL" "$1" "${2:-general}"
}

# Log codespace operation
log_operation() {
    local operation="$1"
    local codespace="$2"
    local status="$3"
    local details="${4:-}"
    
    local message="Operation: $operation | Codespace: $codespace | Status: $status"
    [ -n "$details" ] && message="$message | Details: $details"
    
    case "$status" in
        "started"|"starting")
            log_info "$message" "operations"
            ;;
        "completed"|"success")
            log_info "$message" "operations"
            ;;
        "failed"|"error")
            log_error "$message" "operations"
            ;;
        *)
            log_debug "$message" "operations"
            ;;
    esac
}

# Export functions
export -f log_message
export -f log_debug
export -f log_info
export -f log_warning
export -f log_error
export -f log_critical
export -f log_operation