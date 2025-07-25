#!/bin/bash

# Path Configuration Module
# Centralizes all path management for Michael's Codespaces

# Determine the root directory of the installation
get_codespace_home() {
    if [ -n "$CODESPACE_HOME" ]; then
        echo "$CODESPACE_HOME"
    else
        # Try to find it relative to this script
        local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        local potential_home="$(cd "$script_dir/../.." && pwd)"
        
        # Verify this is actually the codespace home
        if [ -f "$potential_home/bin/mcs" ]; then
            echo "$potential_home"
        else
            # Fall back to default location
            echo "$HOME/.mcs"
        fi
    fi
}

# Export the base paths
export CODESPACE_HOME="${CODESPACE_HOME:-$(get_codespace_home)}"
export CODESPACES_DIR="${CODESPACES_DIR:-$HOME/codespaces}"

# Define all standard paths
export MCS_BIN_DIR="$CODESPACE_HOME/bin"
export MCS_SCRIPTS_DIR="$CODESPACE_HOME/scripts"
export MCS_CORE_DIR="$MCS_SCRIPTS_DIR/core"
export MCS_MODULES_DIR="$MCS_SCRIPTS_DIR/modules"
export MCS_UTILS_DIR="$MCS_SCRIPTS_DIR/utils"
export MCS_TEMPLATES_DIR="$MCS_SCRIPTS_DIR/templates"
export MCS_COMPONENTS_DIR="$MCS_MODULES_DIR/components"
export MCS_INSTALLERS_DIR="$MCS_COMPONENTS_DIR/installers"
export MCS_PRESETS_DIR="$MCS_COMPONENTS_DIR/presets"

# Codespaces directories
export MCS_AUTH_DIR="$CODESPACES_DIR/auth"
export MCS_TOKENS_DIR="$MCS_AUTH_DIR/tokens"
export MCS_SHARED_DIR="$CODESPACES_DIR/shared"
export MCS_BACKUPS_DIR="$CODESPACES_DIR/backups"

# Verify critical paths exist
verify_paths() {
    local missing=0
    
    if [ ! -d "$CODESPACE_HOME" ]; then
        echo "ERROR: CODESPACE_HOME not found: $CODESPACE_HOME" >&2
        ((missing++))
    fi
    
    if [ ! -f "$MCS_BIN_DIR/mcs" ]; then
        echo "ERROR: mcs command not found: $MCS_BIN_DIR/mcs" >&2
        ((missing++))
    fi
    
    if [ ! -d "$MCS_SCRIPTS_DIR" ]; then
        echo "ERROR: Scripts directory not found: $MCS_SCRIPTS_DIR" >&2
        ((missing++))
    fi
    
    return $missing
}

# Create required directories
create_required_dirs() {
    mkdir -p "$CODESPACES_DIR"
    mkdir -p "$MCS_AUTH_DIR"
    mkdir -p "$MCS_TOKENS_DIR"
    mkdir -p "$MCS_SHARED_DIR"
    mkdir -p "$MCS_BACKUPS_DIR"
}

# Get a specific script path
get_script_path() {
    local script_name="$1"
    local script_path
    
    case "$script_name" in
        create-codespace)
            script_path="$MCS_CORE_DIR/create-codespace.sh"
            ;;
        colors)
            script_path="$MCS_UTILS_DIR/colors.sh"
            ;;
        logging)
            script_path="$MCS_UTILS_DIR/logging.sh"
            ;;
        registry)
            script_path="$MCS_COMPONENTS_DIR/registry.sh"
            ;;
        interactive-selector)
            script_path="$MCS_COMPONENTS_DIR/interactive-selector.sh"
            ;;
        manifest-generator)
            script_path="$MCS_COMPONENTS_DIR/manifest-generator.sh"
            ;;
        *)
            echo "Unknown script: $script_name" >&2
            return 1
            ;;
    esac
    
    if [ -f "$script_path" ]; then
        echo "$script_path"
    else
        echo "Script not found: $script_path" >&2
        return 1
    fi
}

# Source a utility script with fallback
source_utility() {
    local util_name="$1"
    local util_path=$(get_script_path "$util_name")
    
    if [ -f "$util_path" ]; then
        source "$util_path"
        return 0
    else
        return 1
    fi
}

# Export functions
export -f get_codespace_home
export -f verify_paths
export -f create_required_dirs
export -f get_script_path
export -f source_utility