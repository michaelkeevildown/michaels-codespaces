#!/bin/bash

# Setup script for creating individual repository codespaces
# Usage: ./setup-repo-codespace.sh git@github.com:user/repo.git

set -e

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    # Fallback definitions
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
    echo_step() { echo "▶️  $1"; }
fi

# Check arguments
if [ $# -eq 0 ]; then
    echo_error "Usage: $0 <git-repository-url> [options]"
    echo ""
    echo "Examples:"
    echo "  $0 git@github.com:user/repo.git"
    echo "  $0 https://github.com/user/repo.git"
    echo "  $0 https://github.com/user/repo.git --image node:18"
    echo "  $0 git@github.com:user/repo.git --ports 8090:8080,3001:3000"
    echo ""
    echo "Options:"
    echo "  --image <image>    Use custom Docker image (default: codercom/code-server:latest)"
    echo "  --ports <mapping>  Custom port mappings (format: host:container,host:container)"
    echo "  --env <file>       Path to environment file to load"
    exit 1
fi

REPO_URL="$1"
shift

# Parse optional arguments
CUSTOM_IMAGE=""
CUSTOM_PORTS=""
ENV_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --image)
            CUSTOM_IMAGE="$2"
            shift 2
            ;;
        --ports)
            CUSTOM_PORTS="$2"
            shift 2
            ;;
        --env)
            ENV_FILE="$2"
            shift 2
            ;;
        *)
            echo_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validate repository URL
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

if ! validate_repo_url "$REPO_URL"; then
    echo_error "Invalid repository URL format: $REPO_URL"
    echo "Expected formats:"
    echo "  - git@github.com:user/repo.git"
    echo "  - https://github.com/user/repo.git"
    echo "  - https://github.com/user/repo (GitHub/GitLab/Bitbucket)"
    exit 1
fi

# Extract repository name from URL
REPO_NAME=$(basename "$REPO_URL" .git)
REPO_OWNER=$(echo "$REPO_URL" | sed -E 's/.*[:/]([^/]+)\/[^/]+$/\1/')
FULL_REPO_NAME="${REPO_OWNER}-${REPO_NAME}"

# Create safe name for directories and containers
SAFE_REPO_NAME=$(echo "$FULL_REPO_NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g')

echo ""
echo_step "🚀 Creating codespace for: $REPO_NAME"
echo "Repository: $REPO_URL"
echo "Codespace name: $SAFE_REPO_NAME"
echo ""

# Check if codespace already exists
CODESPACE_DIR="$HOME/codespaces/$SAFE_REPO_NAME"
if [ -d "$CODESPACE_DIR" ]; then
    echo_error "Codespace '$SAFE_REPO_NAME' already exists!"
    echo "Directory: $CODESPACE_DIR"
    echo ""
    echo "To remove it, run:"
    echo "  rm -rf $CODESPACE_DIR"
    exit 1
fi

# Enhanced port management
find_available_port() {
    local start_port=$1
    local max_port=$((start_port + 100))
    local port=$start_port
    
    while [ $port -le $max_port ]; do
        if ! netstat -tuln 2>/dev/null | grep -q ":$port " && 
           ! lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo $port
            return 0
        fi
        ((port++))
    done
    
    echo_error "No available ports found in range $start_port-$max_port"
    return 1
}

# Port allocation with custom ports support
if [ -n "$CUSTOM_PORTS" ]; then
    echo_info "Using custom port mappings: $CUSTOM_PORTS"
    # Parse custom ports later in docker-compose generation
    VS_CODE_PORT=$(echo "$CUSTOM_PORTS" | cut -d',' -f1 | cut -d':' -f1)
    APP_PORT=$(echo "$CUSTOM_PORTS" | cut -d',' -f2 | cut -d':' -f1 || echo "7680")
else
    echo_info "Finding available ports..."
    VS_CODE_PORT=$(find_available_port 8080)
    if [ $? -ne 0 ]; then
        echo_error "Failed to find available VS Code port"
        exit 1
    fi
    
    APP_PORT=$(find_available_port 7680)
    if [ $? -ne 0 ]; then
        echo_error "Failed to find available app port"
        exit 1
    fi
fi

echo "  VS Code Port: $VS_CODE_PORT"
echo "  App Port: $APP_PORT"

# Create codespace directory structure with error handling
echo_info "Creating codespace structure..."
if ! mkdir -p "$CODESPACE_DIR"/{src,data,config,logs}; then
    echo_error "Failed to create codespace directory structure"
    exit 1
fi

# Check for .devcontainer.json in the repository (will check after clone)
DEVCONTAINER_IMAGE=""

# Generate secure password with fallback
if command -v openssl >/dev/null 2>&1; then
    PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-16)
else
    # Fallback to /dev/urandom
    PASSWORD=$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16)
fi

# Use custom image if provided, otherwise default
DOCKER_IMAGE=${CUSTOM_IMAGE:-"codercom/code-server:latest"}

# Create environment file with enhanced configuration
echo_info "Creating configuration..."
cat > "$CODESPACE_DIR/.env" << EOF
# Codespace Configuration
REPO_NAME=$REPO_NAME
REPO_URL=$REPO_URL
CONTAINER_NAME=$SAFE_REPO_NAME-dev
VS_CODE_PORT=$VS_CODE_PORT
APP_PORT=$APP_PORT
PASSWORD=$PASSWORD
DOCKER_IMAGE=$DOCKER_IMAGE
CREATED=$(date +%Y-%m-%d)

# User Configuration
USER=$USER
TZ=$(timedatectl show -p Timezone --value 2>/dev/null || echo "UTC")
EOF

# Append custom environment variables if env file provided
if [ -n "$ENV_FILE" ] && [ -f "$ENV_FILE" ]; then
    echo "" >> "$CODESPACE_DIR/.env"
    echo "# Custom Environment Variables" >> "$CODESPACE_DIR/.env"
    cat "$ENV_FILE" >> "$CODESPACE_DIR/.env"
fi

# Create docker-compose.yml with enhanced configuration
echo_info "Creating Docker configuration..."

# Generate port mappings
if [ -n "$CUSTOM_PORTS" ]; then
    PORT_MAPPINGS=""
    IFS=',' read -ra PORTS <<< "$CUSTOM_PORTS"
    for port in "${PORTS[@]}"; do
        PORT_MAPPINGS="$PORT_MAPPINGS\n      - \"$port\""
    done
else
    PORT_MAPPINGS="\n      - \"${VS_CODE_PORT}:8080\"\n      - \"${APP_PORT}:3000\""
fi

cat > "$CODESPACE_DIR/docker-compose.yml" << EOF
version: '3.8'

services:
  ${SAFE_REPO_NAME}-dev:
    image: ${DOCKER_IMAGE}
    container_name: ${SAFE_REPO_NAME}-dev
    restart: unless-stopped
    environment:
      - PASSWORD=${PASSWORD}
      - TZ=\${TZ:-UTC}
      - DOCKER_USER=\${USER}
      - GIT_AUTHOR_NAME=\$(cat ~/codespaces/auth/git-config/name 2>/dev/null || echo '')
      - GIT_AUTHOR_EMAIL=\$(cat ~/codespaces/auth/git-config/email 2>/dev/null || echo '')
    ports:${PORT_MAPPINGS}
    volumes:
      - ./src:/home/coder/project
      - ./data:/home/coder/.local/share/code-server
      - ./config:/home/coder/.config
      - ./logs:/home/coder/logs
      - \${HOME}/.ssh:/home/coder/.ssh:ro
      - \${HOME}/codespaces/auth/tokens:/home/coder/.tokens:ro
    networks:
      - codespace-network
    labels:
      - "codespace.repo=${REPO_NAME}"
      - "codespace.created=$(date +%Y-%m-%d)"
      - "codespace.image=${DOCKER_IMAGE}"
    command: >
      sh -c "
        git config --global user.name '\$(cat /home/coder/.ssh/../codespaces/auth/git-config/name 2>/dev/null || echo '')' &&
        git config --global user.email '\$(cat /home/coder/.ssh/../codespaces/auth/git-config/email 2>/dev/null || echo '')' &&
        code-server --bind-addr 0.0.0.0:8080 --auth password
      "
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  codespace-network:
    name: ${SAFE_REPO_NAME}-network
    driver: bridge
EOF

# Clone repository
echo_info "Cloning repository..."

# Check for GitHub token
TOKEN_FILE="$HOME/codespaces/auth/tokens/github.token"
if [ -f "$TOKEN_FILE" ] && [ -s "$TOKEN_FILE" ]; then
    GITHUB_TOKEN=$(cat "$TOKEN_FILE")
    
    # Convert SSH URL to HTTPS with token if needed
    if [[ "$REPO_URL" =~ ^git@github\.com:(.+)$ ]]; then
        REPO_PATH="${BASH_REMATCH[1]}"
        CLONE_URL="https://${GITHUB_TOKEN}@github.com/${REPO_PATH}"
        echo_debug "Using token authentication for GitHub"
    elif [[ "$REPO_URL" =~ ^https://github\.com/(.+)$ ]]; then
        REPO_PATH="${BASH_REMATCH[1]}"
        CLONE_URL="https://${GITHUB_TOKEN}@github.com/${REPO_PATH}"
        echo_debug "Using token authentication for GitHub"
    else
        # Non-GitHub URL, use as-is
        CLONE_URL="$REPO_URL"
    fi
else
    # No token, try clone as-is (will work for public repos or if SSH is set up)
    CLONE_URL="$REPO_URL"
    echo_debug "No GitHub token found, using default authentication"
fi

# Clone with progress
if ! git clone "$CLONE_URL" "$CODESPACE_DIR/src" 2>&1 | while IFS= read -r line; do
    if [[ "$line" =~ "Receiving objects" ]] || [[ "$line" =~ "Resolving deltas" ]]; then
        echo_status "$line"
    fi
done; then
    clear_status
    echo_error "Failed to clone repository. Please check:"
    echo "  - Repository URL is correct"
    echo "  - You have access to the repository"
    echo "  - GitHub token is set (see: ~/codespaces/auth/tokens/README.md)"
    rm -rf "$CODESPACE_DIR"
    exit 1
fi

clear_status

# Remove token from git config in the cloned repo for security
if [ -d "$CODESPACE_DIR/src/.git" ]; then
    cd "$CODESPACE_DIR/src"
    # Set the remote URL without the token
    git remote set-url origin "$REPO_URL"
    cd - > /dev/null
fi

# Create README for the codespace
echo_info "Creating documentation..."
cat > "$CODESPACE_DIR/README.md" << EOF
# $REPO_NAME Development Environment

## Quick Start

\`\`\`bash
# Start development environment
start-$SAFE_REPO_NAME

# Access VS Code
open http://localhost:$VS_CODE_PORT

# View logs
logs-$SAFE_REPO_NAME

# Stop environment
stop-$SAFE_REPO_NAME
\`\`\`

## Access Credentials

- **VS Code URL**: http://localhost:$VS_CODE_PORT
- **Password**: $PASSWORD

## Container Details

- **Container**: ${SAFE_REPO_NAME}-dev
- **VS Code Port**: $VS_CODE_PORT
- **App Port**: $APP_PORT (mapped to container port 3000)
- **Created**: $(date +%Y-%m-%d)

## Useful Commands

- \`cd-$SAFE_REPO_NAME\` - Navigate to project directory
- \`exec-$SAFE_REPO_NAME\` - Enter container shell
- \`rebuild-$SAFE_REPO_NAME\` - Rebuild container
- \`remove-$SAFE_REPO_NAME\` - Remove this codespace

## Repository

$REPO_URL
EOF

# Create management scripts
echo_info "Creating management commands..."

# Create aliases file
cat > "$CODESPACE_DIR/aliases.sh" << EOF
#!/bin/bash

# Aliases for $REPO_NAME codespace

alias cd-$SAFE_REPO_NAME='cd $CODESPACE_DIR'
alias src-$SAFE_REPO_NAME='cd $CODESPACE_DIR/src'
alias start-$SAFE_REPO_NAME='docker-compose -f $CODESPACE_DIR/docker-compose.yml up -d'
alias stop-$SAFE_REPO_NAME='docker-compose -f $CODESPACE_DIR/docker-compose.yml stop'
alias logs-$SAFE_REPO_NAME='docker-compose -f $CODESPACE_DIR/docker-compose.yml logs -f'
alias exec-$SAFE_REPO_NAME='docker exec -it ${SAFE_REPO_NAME}-dev /bin/bash'
alias rebuild-$SAFE_REPO_NAME='docker-compose -f $CODESPACE_DIR/docker-compose.yml up -d --build'
alias remove-$SAFE_REPO_NAME='docker-compose -f $CODESPACE_DIR/docker-compose.yml down && rm -rf $CODESPACE_DIR'

echo "✅ Codespace aliases loaded for: $REPO_NAME"
EOF

# Add aliases to .zshrc
if ! grep -q "# Codespace: $SAFE_REPO_NAME" ~/.zshrc; then
    echo "" >> ~/.zshrc
    echo "# Codespace: $SAFE_REPO_NAME" >> ~/.zshrc
    echo "[ -f \"$CODESPACE_DIR/aliases.sh\" ] && source \"$CODESPACE_DIR/aliases.sh\"" >> ~/.zshrc
fi

# Start the codespace
echo_info "Starting codespace..."
docker-compose -f "$CODESPACE_DIR/docker-compose.yml" up -d

# Wait for container to be ready
echo_info "Waiting for VS Code to be ready..."
sleep 5

# Check if container is running
if docker ps | grep -q "${SAFE_REPO_NAME}-dev"; then
    echo ""
    echo_success "🎉 Codespace created successfully!"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📦 Codespace: $SAFE_REPO_NAME"
    echo "🌐 VS Code URL: http://localhost:$VS_CODE_PORT"
    echo "🔑 Password: $PASSWORD"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Useful commands:"
    echo "  start-$SAFE_REPO_NAME    - Start codespace"
    echo "  stop-$SAFE_REPO_NAME     - Stop codespace"
    echo "  logs-$SAFE_REPO_NAME     - View logs"
    echo "  exec-$SAFE_REPO_NAME     - Enter container"
    echo "  cd-$SAFE_REPO_NAME       - Go to codespace directory"
    echo ""
    echo "To load aliases in current shell:"
    echo "  source ~/.zshrc"
    echo ""
else
    echo_error "Container failed to start. Check logs with:"
    echo "  docker-compose -f $CODESPACE_DIR/docker-compose.yml logs"
fi