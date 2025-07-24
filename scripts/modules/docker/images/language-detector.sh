#!/bin/bash

# Language Detection Module
# Detects project language and recommends appropriate Docker images

# Detect project language from source directory
detect_language() {
    local src_dir="$1"
    local detected_language=""
    
    # Node.js detection
    if [ -f "$src_dir/package.json" ]; then
        detected_language="node"
        echo_info "Detected Node.js project"
    # Python detection
    elif [ -f "$src_dir/requirements.txt" ] || [ -f "$src_dir/setup.py" ] || [ -f "$src_dir/Pipfile" ] || [ -f "$src_dir/pyproject.toml" ]; then
        detected_language="python"
        echo_info "Detected Python project"
    # Go detection
    elif [ -f "$src_dir/go.mod" ] || [ -f "$src_dir/go.sum" ]; then
        detected_language="go"
        echo_info "Detected Go project"
    # Rust detection
    elif [ -f "$src_dir/Cargo.toml" ] || [ -f "$src_dir/Cargo.lock" ]; then
        detected_language="rust"
        echo_info "Detected Rust project"
    # Java detection
    elif [ -f "$src_dir/pom.xml" ] || [ -f "$src_dir/build.gradle" ] || [ -f "$src_dir/build.gradle.kts" ]; then
        detected_language="java"
        echo_info "Detected Java project"
    # Ruby detection
    elif [ -f "$src_dir/Gemfile" ] || [ -f "$src_dir/Rakefile" ]; then
        detected_language="ruby"
        echo_info "Detected Ruby project"
    # PHP detection
    elif [ -f "$src_dir/composer.json" ] || [ -f "$src_dir/composer.lock" ]; then
        detected_language="php"
        echo_info "Detected PHP project"
    # .NET detection
    elif find "$src_dir" -maxdepth 2 -name "*.csproj" -o -name "*.fsproj" -o -name "*.vbproj" | grep -q .; then
        detected_language="dotnet"
        echo_info "Detected .NET project"
    fi
    
    echo "$detected_language"
}

# Get recommended Docker image for a language
get_language_image() {
    local language="$1"
    local image=""
    
    case "$language" in
        node)
            image="node:20-bullseye"
            ;;
        python)
            image="python:3.11-bullseye"
            ;;
        go)
            image="golang:1.21-bullseye"
            ;;
        rust)
            image="rust:1.75-bullseye"
            ;;
        java)
            image="openjdk:17-bullseye"
            ;;
        ruby)
            image="ruby:3.2-bullseye"
            ;;
        php)
            image="php:8.2-cli-bullseye"
            ;;
        dotnet)
            image="mcr.microsoft.com/dotnet/sdk:8.0"
            ;;
        *)
            # Default to code-server base image
            image="codercom/code-server:latest"
            ;;
    esac
    
    echo "$image"
}

# Parse devcontainer.json for image information
parse_devcontainer_image() {
    local devcontainer_file="$1"
    local image=""
    
    if [ -f "$devcontainer_file" ]; then
        # Try to extract image field from JSON
        # Using simple grep/sed as jq might not be available
        image=$(grep -E '"image":\s*"[^"]+"' "$devcontainer_file" | \
                sed -E 's/.*"image":\s*"([^"]+)".*/\1/' | \
                head -n1)
        
        # If no image, try dockerFile
        if [ -z "$image" ]; then
            local dockerfile=$(grep -E '"dockerFile":\s*"[^"]+"' "$devcontainer_file" | \
                             sed -E 's/.*"dockerFile":\s*"([^"]+)".*/\1/' | \
                             head -n1)
            if [ -n "$dockerfile" ]; then
                echo_debug "Devcontainer uses Dockerfile: $dockerfile"
            fi
        fi
    fi
    
    echo "$image"
}

# Check for devcontainer configuration
check_devcontainer() {
    local src_dir="$1"
    local devcontainer_file=""
    
    # Check common devcontainer locations
    if [ -f "$src_dir/.devcontainer/devcontainer.json" ]; then
        devcontainer_file="$src_dir/.devcontainer/devcontainer.json"
    elif [ -f "$src_dir/.devcontainer.json" ]; then
        devcontainer_file="$src_dir/.devcontainer.json"
    fi
    
    echo "$devcontainer_file"
}

# Get language-specific environment variables
get_language_env_vars() {
    local language="$1"
    local env_vars=""
    
    case "$language" in
        node)
            env_vars="NODE_ENV=development"
            ;;
        python)
            env_vars="PYTHONUNBUFFERED=1"
            ;;
        go)
            env_vars="GO111MODULE=on CGO_ENABLED=0"
            ;;
        rust)
            env_vars="RUST_BACKTRACE=1"
            ;;
        java)
            env_vars="JAVA_OPTS=-Xmx512m"
            ;;
        ruby)
            env_vars="BUNDLE_PATH=/usr/local/bundle"
            ;;
        php)
            env_vars="COMPOSER_ALLOW_SUPERUSER=1"
            ;;
        dotnet)
            env_vars="DOTNET_CLI_TELEMETRY_OPTOUT=1"
            ;;
    esac
    
    echo "$env_vars"
}

# Export functions
export -f detect_language
export -f get_language_image
export -f parse_devcontainer_image
export -f check_devcontainer
export -f get_language_env_vars