#!/bin/bash

# Component Manifest Generator Module
# Generates manifest files for container initialization

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/registry.sh"

# Generate manifest file
generate_manifest() {
    local output_file="$1"
    shift
    local components=("$@")
    
    # Validate components
    if ! validate_components "${components[@]}"; then
        echo "Invalid components specified" >&2
        return 1
    fi
    
    # Get installation order (with dependencies)
    local install_order=$(get_install_order "${components[@]}")
    
    # Create output directory
    local output_dir=$(dirname "$output_file")
    mkdir -p "$output_dir"
    
    # Write manifest
    echo "# Component manifest generated on $(date)" > "$output_file"
    echo "# Components will be installed in dependency order" >> "$output_file"
    echo "" >> "$output_file"
    
    for component in $install_order; do
        echo "$component" >> "$output_file"
    done
    
    echo "Generated manifest: $output_file"
    echo "Components to install: $(echo "$install_order" | wc -w)"
}

# Generate manifest with metadata
generate_detailed_manifest() {
    local output_file="$1"
    shift
    local components=("$@")
    
    # Validate components
    if ! validate_components "${components[@]}"; then
        echo "Invalid components specified" >&2
        return 1
    fi
    
    # Get installation order
    local install_order=$(get_install_order "${components[@]}")
    
    # Create output directory
    local output_dir=$(dirname "$output_file")
    mkdir -p "$output_dir"
    
    # Generate JSON manifest
    cat > "$output_file" << EOF
{
  "version": "1.0",
  "generated": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "components": [
EOF
    
    local first=true
    for component in $install_order; do
        local name=$(get_component_info "$component" "name")
        local description=$(get_component_info "$component" "description")
        local installer=$(get_component_info "$component" "installer")
        local deps=$(get_component_dependencies "$component")
        
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> "$output_file"
        fi
        
        cat >> "$output_file" << EOF
    {
      "id": "$component",
      "name": "$name",
      "description": "$description",
      "installer": "$installer",
      "dependencies": [$(echo "$deps" | sed 's/ /", "/g' | sed 's/^/"/;s/$/"/')]
    }
EOF
    done
    
    cat >> "$output_file" << EOF

  ],
  "metadata": {
    "total_components": $(echo "$install_order" | wc -w),
    "user_selected": ${#components[@]},
    "auto_dependencies": $(($(echo "$install_order" | wc -w) - ${#components[@]}))
  }
}
EOF
    
    echo "Generated detailed manifest: $output_file"
}

# Generate environment file for components
generate_env_file() {
    local output_file="$1"
    shift
    local components=("$@")
    
    # Create output directory
    local output_dir=$(dirname "$output_file")
    mkdir -p "$output_dir"
    
    # Generate environment variables
    cat > "$output_file" << EOF
# Component environment variables
# Generated on $(date)

# Component list
CODESPACE_COMPONENTS="$(echo "${components[@]}" | tr ' ' ',')"

# Component flags
EOF
    
    for component in "${components[@]}"; do
        local var_name=$(echo "$component" | tr '[:lower:]-' '[:upper:]_')
        echo "COMPONENT_${var_name}_ENABLED=true" >> "$output_file"
    done
    
    # Add component-specific environment variables
    echo "" >> "$output_file"
    echo "# Component-specific variables" >> "$output_file"
    
    for component in "${components[@]}"; do
        case "$component" in
            github-cli)
                echo "# GitHub CLI" >> "$output_file"
                echo "GH_PROMPT_ENABLED=true" >> "$output_file"
                ;;
            claude)
                echo "# Claude CLI" >> "$output_file"
                echo "CLAUDE_DEFAULT_MODEL=claude-3-opus-20240229" >> "$output_file"
                ;;
            claude-flow)
                echo "# Claude Flow" >> "$output_file"
                echo "CLAUDE_FLOW_PARALLEL_EXECUTION=true" >> "$output_file"
                ;;
            docker-in-docker)
                echo "# Docker in Docker" >> "$output_file"
                echo "DOCKER_BUILDKIT=1" >> "$output_file"
                ;;
        esac
    done
    
    echo "Generated environment file: $output_file"
}

# Copy component installers to a directory
copy_installers() {
    local target_dir="$1"
    shift
    local components=("$@")
    
    # Get installation order
    local install_order=$(get_install_order "${components[@]}")
    
    # Create target directory
    mkdir -p "$target_dir"
    
    # Copy installers
    local copied=0
    for component in $install_order; do
        local installer=$(get_component_installer "$component")
        
        if [ -f "$installer" ]; then
            cp "$installer" "$target_dir/"
            chmod +x "$target_dir/$(basename "$installer")"
            ((copied++))
        else
            echo "Warning: Installer not found for $component" >&2
        fi
    done
    
    echo "Copied $copied installer(s) to $target_dir"
}

# Create initialization package
create_init_package() {
    local package_dir="$1"
    shift
    local components=("$@")
    
    echo "Creating initialization package..."
    
    # Create package structure
    mkdir -p "$package_dir"/{installers,config}
    
    # Copy container init script
    local init_script="$SCRIPT_DIR/../../templates/container-init.sh"
    if [ -f "$init_script" ]; then
        cp "$init_script" "$package_dir/init.sh"
        chmod +x "$package_dir/init.sh"
    fi
    
    # Generate manifests
    generate_manifest "$package_dir/config/components.manifest" "${components[@]}"
    generate_detailed_manifest "$package_dir/config/components.json" "${components[@]}"
    generate_env_file "$package_dir/config/components.env" "${components[@]}"
    
    # Copy installers
    copy_installers "$package_dir/installers" "${components[@]}"
    
    # Create README
    cat > "$package_dir/README.md" << EOF
# Codespace Component Package

This package contains the components selected for installation in your codespace.

## Selected Components

$(for component in "${components[@]}"; do
    name=$(get_component_info "$component" "name")
    desc=$(get_component_info "$component" "description")
    echo "- **$name**: $desc"
done)

## Installation

This package will be automatically installed when the container starts.
To manually run the installation:

\`\`\`bash
bash /opt/codespace/init.sh
\`\`\`

## Files

- \`init.sh\`: Main initialization script
- \`config/components.manifest\`: Simple component list
- \`config/components.json\`: Detailed component metadata
- \`config/components.env\`: Environment variables
- \`installers/\`: Component installer scripts

Generated on: $(date)
EOF
    
    echo "Initialization package created: $package_dir"
}

# Export functions
export -f generate_manifest
export -f generate_detailed_manifest
export -f generate_env_file
export -f copy_installers
export -f create_init_package