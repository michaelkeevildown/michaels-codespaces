#!/bin/bash

# Simple Component Selector Module
# Provides a basic text-based menu for component selection

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/registry.sh"

# Simple text-based selection
simple_select() {
    # Debug output
    [ "${DEBUG:-0}" -eq 1 ] && echo "DEBUG: Entering simple_select" >&2
    
    # Ensure components are registered
    register_components
    
    [ "${DEBUG:-0}" -eq 1 ] && echo "DEBUG: Components registered" >&2
    
    echo "Available components:"
    echo ""
    
    # Get all components
    local components=()
    local i=1
    local component_count=0
    
    while IFS= read -r component; do
        components+=("$component")
        local name=$(get_component_info "$component" "name")
        local desc=$(get_component_info "$component" "description")
        
        if [ -n "$name" ]; then
            printf "  %2d) %-20s - %s\n" "$i" "$name" "$desc"
            ((component_count++))
        fi
        ((i++))
    done < <(list_components)
    
    # Check if we found any components
    if [ $component_count -eq 0 ]; then
        echo "  (No components found - check installation)"
        echo ""
        echo "Returning to normal codespace creation..."
        return 1
    fi
    
    echo ""
    echo "Presets:"
    echo "   a) AI Development (GitHub CLI, Claude, Claude Flow)"
    echo "   f) Full Stack (All tools)"
    echo "   m) Minimal (GitHub CLI, Git tools)"
    echo "   d) DevOps (Docker, AWS, Terraform, K8s)"
    echo "   n) None (skip component installation)"
    echo ""
    
    read -p "Select components (comma-separated numbers), preset letter, or press Enter for AI Development: " selection
    
    # Default to AI Development preset if empty
    if [ -z "$selection" ]; then
        selection="a"
    fi
    
    # Handle selection
    case "$selection" in
        a|A)
            echo "github-cli claude claude-flow git-tools vscode-extensions"
            ;;
        f|F)
            echo "${components[@]}"
            ;;
        m|M)
            echo "github-cli git-tools"
            ;;
        d|D)
            echo "github-cli docker-in-docker aws-cli terraform k8s-tools git-tools"
            ;;
        n|N)
            echo ""
            ;;
        *)
            # Parse comma-separated numbers
            local selected=()
            IFS=',' read -ra SELECTIONS <<< "$selection"
            for sel in "${SELECTIONS[@]}"; do
                # Remove whitespace
                sel=$(echo "$sel" | xargs)
                # Check if it's a valid number
                if [[ "$sel" =~ ^[0-9]+$ ]] && [ "$sel" -ge 1 ] && [ "$sel" -le "${#components[@]}" ]; then
                    selected+=("${components[$((sel-1))]}")
                fi
            done
            echo "${selected[@]}"
            ;;
    esac
}

# Export function
export -f simple_select