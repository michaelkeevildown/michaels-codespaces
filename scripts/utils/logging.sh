#!/bin/bash

# Logging Utilities for Michael's Codespaces
# Provides comprehensive logging for debugging and troubleshooting

# Configuration
LOGS_DIR="${HOME}/codespaces/logs"
LOG_FILE="${LOGS_DIR}/mcs-$(date +%Y%m%d-%H%M%S).log"
VERBOSE=${VERBOSE:-0}

# Ensure logs directory exists
init_logging() {
    mkdir -p "$LOGS_DIR"
    
    # Create log file with header
    cat > "$LOG_FILE" << EOF
=== Michael's Codespaces Log ===
Started: $(date)
Command: $0 $@
PID: $$
Working Dir: $(pwd)
User: $(whoami)
System: $(uname -a)
=====================================

EOF
    
    # Clean up old logs (keep last 10)
    find "$LOGS_DIR" -name "mcs-*.log" -type f | sort | head -n -10 | xargs rm -f 2>/dev/null || true
    
    log_info "Logging initialized: $LOG_FILE"
}

# Core logging function
write_log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
    
    # Also output to console if verbose or if it's an error
    if [ "$VERBOSE" -eq 1 ] || [ "$level" = "ERROR" ]; then
        echo "[$level] $message" >&2
    fi
}

# Logging functions
log_debug() {
    write_log "DEBUG" "$1"
}

log_info() {
    write_log "INFO" "$1"
}

log_warning() {
    write_log "WARN" "$1"
}

log_error() {
    write_log "ERROR" "$1"
}

# Log command execution with output capture
log_command() {
    local cmd="$1"
    local description="${2:-$cmd}"
    
    log_info "Executing: $description"
    log_debug "Command: $cmd"
    
    # Execute command and capture both stdout and stderr
    local output
    local exit_code
    
    if output=$(eval "$cmd" 2>&1); then
        exit_code=0
        log_debug "Command succeeded: $description"
        if [ -n "$output" ]; then
            log_debug "Output: $output"
        fi
    else
        exit_code=$?
        log_error "Command failed ($exit_code): $description"
        if [ -n "$output" ]; then
            log_error "Error output: $output"
        fi
    fi
    
    echo "$output"
    return $exit_code
}

# Log file contents
log_file_content() {
    local file="$1"
    local description="${2:-$file}"
    
    log_info "Logging file content: $description"
    if [ -f "$file" ]; then
        echo "=== Content of $file ===" >> "$LOG_FILE"
        cat "$file" >> "$LOG_FILE"
        echo "=== End of $file ===" >> "$LOG_FILE"
        log_debug "File content logged: $file ($(wc -l < "$file") lines)"
    else
        log_error "File not found for logging: $file"
    fi
}

# Log environment variables
log_environment() {
    local prefix="${1:-}"
    
    log_info "Logging environment variables (prefix: $prefix)"
    echo "=== Environment Variables ===" >> "$LOG_FILE"
    if [ -n "$prefix" ]; then
        env | grep "^$prefix" | sort >> "$LOG_FILE"
    else
        env | sort >> "$LOG_FILE"
    fi
    echo "=== End Environment Variables ===" >> "$LOG_FILE"
}

# Log variable state
log_variables() {
    local title="$1"
    shift
    
    log_info "Logging variables: $title"
    echo "=== Variables: $title ===" >> "$LOG_FILE"
    for var in "$@"; do
        echo "$var = ${!var}" >> "$LOG_FILE"
    done
    echo "=== End Variables ===" >> "$LOG_FILE"
}

# Get current log file path
get_log_file() {
    echo "$LOG_FILE"
}

# Export functions for use in other scripts
export -f init_logging
export -f write_log
export -f log_debug
export -f log_info
export -f log_warning
export -f log_error
export -f log_command
export -f log_file_content
export -f log_environment
export -f log_variables
export -f get_log_file