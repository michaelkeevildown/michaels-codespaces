#!/bin/bash

# Repository Validator Module
# Validates repository URLs and fetches repository metadata

# Validate repository URL format
validate_repo_url() {
    local url="$1"
    
    # Check for common Git URL patterns
    if [[ "$url" =~ ^git@.*:.*\.git$ ]] || 
       [[ "$url" =~ ^https?://.*\.git$ ]] || 
       [[ "$url" =~ ^git://.*\.git$ ]] ||
       [[ "$url" =~ ^ssh://.*\.git$ ]] ||
       [[ "$url" =~ ^https://github\.com/[^/]+/[^/]+$ ]] ||
       [[ "$url" =~ ^https://gitlab\.com/[^/]+/[^/]+$ ]] ||
       [[ "$url" =~ ^https://bitbucket\.org/[^/]+/[^/]+$ ]]; then
        return 0
    else
        return 1
    fi
}

# Extract repository information from URL
parse_repo_url() {
    local url="$1"
    local info_type="${2:-all}"  # owner, name, provider, or all
    
    local provider=""
    local owner=""
    local name=""
    
    # Remove .git suffix if present
    url="${url%.git}"
    
    # Detect provider and extract owner/name
    if [[ "$url" =~ github\.com ]]; then
        provider="github"
        if [[ "$url" =~ ^git@github\.com:(.+)/(.+)$ ]]; then
            owner="${BASH_REMATCH[1]}"
            name="${BASH_REMATCH[2]}"
        elif [[ "$url" =~ ^https://github\.com/([^/]+)/([^/]+) ]]; then
            owner="${BASH_REMATCH[1]}"
            name="${BASH_REMATCH[2]}"
        fi
    elif [[ "$url" =~ gitlab\.com ]]; then
        provider="gitlab"
        if [[ "$url" =~ ^git@gitlab\.com:(.+)/(.+)$ ]]; then
            owner="${BASH_REMATCH[1]}"
            name="${BASH_REMATCH[2]}"
        elif [[ "$url" =~ ^https://gitlab\.com/([^/]+)/([^/]+) ]]; then
            owner="${BASH_REMATCH[1]}"
            name="${BASH_REMATCH[2]}"
        fi
    elif [[ "$url" =~ bitbucket\.org ]]; then
        provider="bitbucket"
        if [[ "$url" =~ ^git@bitbucket\.org:(.+)/(.+)$ ]]; then
            owner="${BASH_REMATCH[1]}"
            name="${BASH_REMATCH[2]}"
        elif [[ "$url" =~ ^https://bitbucket\.org/([^/]+)/([^/]+) ]]; then
            owner="${BASH_REMATCH[1]}"
            name="${BASH_REMATCH[2]}"
        fi
    fi
    
    # Return requested information
    case "$info_type" in
        owner)
            echo "$owner"
            ;;
        name)
            echo "$name"
            ;;
        provider)
            echo "$provider"
            ;;
        all)
            echo "provider=$provider"
            echo "owner=$owner"
            echo "name=$name"
            ;;
    esac
}

# Check if repository is accessible (requires authentication setup)
check_repo_access() {
    local url="$1"
    
    # Try to access the repository
    if git ls-remote "$url" HEAD &>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Get repository metadata via API
get_repo_metadata() {
    local url="$1"
    local token="${GITHUB_TOKEN:-}"
    
    # Parse repository information
    local provider=$(parse_repo_url "$url" "provider")
    local owner=$(parse_repo_url "$url" "owner")
    local name=$(parse_repo_url "$url" "name")
    
    if [ "$provider" != "github" ]; then
        echo_warning "Metadata fetching only supported for GitHub repositories"
        return 1
    fi
    
    if [ -z "$token" ]; then
        echo_warning "No GitHub token found. Some metadata may be unavailable."
    fi
    
    # Fetch repository data from GitHub API
    local api_url="https://api.github.com/repos/${owner}/${name}"
    local curl_opts="-s"
    
    if [ -n "$token" ]; then
        curl_opts="$curl_opts -H 'Authorization: token ${token}'"
    fi
    
    local response=$(curl $curl_opts "$api_url" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        # Parse basic information
        local description=$(echo "$response" | grep -o '"description":"[^"]*' | cut -d'"' -f4)
        local default_branch=$(echo "$response" | grep -o '"default_branch":"[^"]*' | cut -d'"' -f4)
        local language=$(echo "$response" | grep -o '"language":"[^"]*' | cut -d'"' -f4)
        local stars=$(echo "$response" | grep -o '"stargazers_count":[0-9]*' | cut -d':' -f2)
        local is_private=$(echo "$response" | grep -o '"private":[^,]*' | cut -d':' -f2)
        
        echo "Repository: ${owner}/${name}"
        [ -n "$description" ] && echo "Description: $description"
        [ -n "$default_branch" ] && echo "Default Branch: $default_branch"
        [ -n "$language" ] && echo "Primary Language: $language"
        [ -n "$stars" ] && echo "Stars: $stars"
        [ -n "$is_private" ] && echo "Private: $is_private"
        
        return 0
    else
        echo_error "Failed to fetch repository metadata"
        return 1
    fi
}

# Convert repository URL to authenticated format
convert_to_authenticated_url() {
    local url="$1"
    local token="${2:-${GITHUB_TOKEN:-}}"
    
    # Only convert GitHub HTTPS URLs
    if [[ "$url" =~ ^https://github\.com/ ]] && [ -n "$token" ]; then
        # Insert token into URL
        echo "$url" | sed "s|https://|https://${token}@|"
    else
        # Return original URL
        echo "$url"
    fi
}

# Normalize repository URL (remove trailing slashes, .git, etc.)
normalize_repo_url() {
    local url="$1"
    
    # Remove trailing slashes
    url="${url%/}"
    
    # Add .git suffix if missing and it's a git URL
    if [[ "$url" =~ ^git@ ]] && [[ ! "$url" =~ \.git$ ]]; then
        url="${url}.git"
    fi
    
    echo "$url"
}

# Check API rate limits
check_github_rate_limit() {
    local token="${GITHUB_TOKEN:-}"
    
    if [ -z "$token" ]; then
        echo_warning "No GitHub token configured. API rate limits apply."
        return 1
    fi
    
    local response=$(curl -s -H "Authorization: token ${token}" \
                    https://api.github.com/rate_limit 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        local remaining=$(echo "$response" | grep -o '"remaining":[0-9]*' | head -1 | cut -d':' -f2)
        local limit=$(echo "$response" | grep -o '"limit":[0-9]*' | head -1 | cut -d':' -f2)
        
        if [ -n "$remaining" ] && [ -n "$limit" ]; then
            echo "GitHub API Rate Limit: ${remaining}/${limit} requests remaining"
            
            if [ "$remaining" -lt 10 ]; then
                echo_warning "Low API rate limit remaining!"
            fi
        fi
    fi
}

# Export functions
export -f validate_repo_url
export -f parse_repo_url
export -f check_repo_access
export -f get_repo_metadata
export -f convert_to_authenticated_url
export -f normalize_repo_url
export -f check_github_rate_limit