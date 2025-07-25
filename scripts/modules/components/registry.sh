#!/bin/bash

# Component Registry Module
# Manages available components for container installation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALLERS_DIR="$SCRIPT_DIR/installers"
PRESETS_DIR="$SCRIPT_DIR/presets"

# Component metadata structure
declare -A COMPONENTS

# Register all available components
register_components() {
    # GitHub CLI
    COMPONENTS[github-cli]="GitHub CLI|Command-line interface for GitHub|github-cli.sh|"
    
    # Claude CLI
    COMPONENTS[claude]="Claude CLI|Anthropic's Claude AI assistant CLI|claude.sh|"
    
    # Claude Flow
    COMPONENTS[claude-flow]="Claude Flow|AI orchestration and workflow tool|claude-flow.sh|claude"
    
    # Docker in Docker
    COMPONENTS[docker-in-docker]="Docker in Docker|Run Docker inside containers|docker-in-docker.sh|"
    
    # AWS CLI
    COMPONENTS[aws-cli]="AWS CLI|Amazon Web Services command-line tools|aws-cli.sh|"
    
    # Terraform
    COMPONENTS[terraform]="Terraform|Infrastructure as Code tool|terraform.sh|"
    
    # Node.js tools
    COMPONENTS[node-tools]="Node.js Tools|npm, yarn, pnpm package managers|node-tools.sh|"
    
    # Python tools
    COMPONENTS[python-tools]="Python Tools|pip, poetry, virtualenv tools|python-tools.sh|"
    
    # Kubernetes tools
    COMPONENTS[k8s-tools]="Kubernetes Tools|kubectl, helm, k9s|k8s-tools.sh|"
    
    # Database clients
    COMPONENTS[db-clients]="Database Clients|PostgreSQL, MySQL, MongoDB clients|db-clients.sh|"
    
    # VS Code extensions
    COMPONENTS[vscode-extensions]="VS Code Extensions|Popular extensions pre-installed|vscode-extensions.sh|"
    
    # Git tools
    COMPONENTS[git-tools]="Git Tools|git-flow, git-lfs, hub|git-tools.sh|github-cli"
}

# Get component metadata
# Usage: get_component_info <component-id> <field>
# Fields: name, description, installer, dependencies
get_component_info() {
    local component="$1"
    local field="$2"
    
    if [ -z "${COMPONENTS[$component]}" ]; then
        return 1
    fi
    
    local info="${COMPONENTS[$component]}"
    case "$field" in
        name)
            echo "$info" | cut -d'|' -f1
            ;;
        description)
            echo "$info" | cut -d'|' -f2
            ;;
        installer)
            echo "$info" | cut -d'|' -f3
            ;;
        dependencies)
            echo "$info" | cut -d'|' -f4
            ;;
        *)
            return 1
            ;;
    esac
}

# List all available components
list_components() {
    register_components
    
    for component in "${!COMPONENTS[@]}"; do
        echo "$component"
    done | sort
}

# Get component display name
get_component_display() {
    local component="$1"
    local name=$(get_component_info "$component" "name")
    local desc=$(get_component_info "$component" "description")
    
    echo "$name|$desc"
}

# Check if component exists
component_exists() {
    local component="$1"
    register_components
    
    if [ -n "${COMPONENTS[$component]}" ]; then
        return 0
    else
        return 1
    fi
}

# Get component installer path
get_component_installer() {
    local component="$1"
    local installer=$(get_component_info "$component" "installer")
    
    if [ -n "$installer" ]; then
        echo "$INSTALLERS_DIR/$installer"
    fi
}

# Get component dependencies
get_component_dependencies() {
    local component="$1"
    local deps=$(get_component_info "$component" "dependencies")
    
    if [ -n "$deps" ]; then
        echo "$deps" | tr ',' ' '
    fi
}

# Resolve all dependencies for a component
resolve_dependencies() {
    local component="$1"
    local resolved=()
    local visited=()
    
    _resolve_deps_recursive "$component"
    
    # Return unique list
    printf '%s\n' "${resolved[@]}" | awk '!seen[$0]++'
}

# Recursive dependency resolver
_resolve_deps_recursive() {
    local component="$1"
    
    # Check if already visited (circular dependency check)
    for v in "${visited[@]}"; do
        if [ "$v" == "$component" ]; then
            return
        fi
    done
    
    visited+=("$component")
    
    # Get dependencies
    local deps=$(get_component_dependencies "$component")
    
    # Resolve dependencies first
    if [ -n "$deps" ]; then
        for dep in $deps; do
            _resolve_deps_recursive "$dep"
        done
    fi
    
    # Add component after dependencies
    resolved+=("$component")
}

# Validate component selection
validate_components() {
    local components=("$@")
    
    register_components
    
    for component in "${components[@]}"; do
        if ! component_exists "$component"; then
            echo "Unknown component: $component" >&2
            return 1
        fi
    done
    
    return 0
}

# Get installation order for components
get_install_order() {
    local components=("$@")
    local all_components=()
    
    # Resolve dependencies for each component
    for component in "${components[@]}"; do
        local deps=$(resolve_dependencies "$component")
        for dep in $deps; do
            all_components+=("$dep")
        done
    done
    
    # Return unique list in order
    printf '%s\n' "${all_components[@]}" | awk '!seen[$0]++'
}

# Load preset configuration
load_preset() {
    local preset="$1"
    local preset_file="$PRESETS_DIR/${preset}.preset"
    
    if [ ! -f "$preset_file" ]; then
        echo "Preset not found: $preset" >&2
        return 1
    fi
    
    # Source preset file to get PRESET_COMPONENTS array
    source "$preset_file"
    
    if [ -z "${PRESET_COMPONENTS+x}" ]; then
        echo "Invalid preset file: $preset" >&2
        return 1
    fi
    
    printf '%s\n' "${PRESET_COMPONENTS[@]}"
}

# List available presets
list_presets() {
    if [ -d "$PRESETS_DIR" ]; then
        for preset_file in "$PRESETS_DIR"/*.preset; do
            if [ -f "$preset_file" ]; then
                local preset_name=$(basename "$preset_file" .preset)
                
                # Source to get description
                source "$preset_file"
                echo "${preset_name}|${PRESET_DESCRIPTION:-No description}"
            fi
        done
    fi
}

# Initialize registry
register_components

# Export functions
export -f register_components
export -f get_component_info
export -f list_components
export -f get_component_display
export -f component_exists
export -f get_component_installer
export -f get_component_dependencies
export -f resolve_dependencies
export -f validate_components
export -f get_install_order
export -f load_preset
export -f list_presets