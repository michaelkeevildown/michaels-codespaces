#!/bin/bash

# Create monitoring and management tools

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

echo_step "📊 Creating monitoring tools..."

# System monitor script
cat > "$HOME/monitor-system.sh" << 'EOF'
#!/bin/bash

source "$HOME/codespaces/scripts/utils/colors.sh" 2>/dev/null || {
    # Fallback if colors not available
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
}

clear
echo "═══════════════════════════════════════════════════════════════"
echo "               Michael's Codespaces Monitor"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# System Info
echo "📊 SYSTEM INFORMATION"
echo "─────────────────────"
echo "🖥️  Hostname: $(hostname)"
echo "🏗️  Architecture: $(uname -m)"
echo "🐧 OS: $(lsb_release -d | cut -f2)"
echo "⏰ Uptime:$(uptime -p | sed 's/up//')"
echo ""

# Resources
echo "💻 SYSTEM RESOURCES"
echo "─────────────────────"
MEM_USAGE=$(free -h | grep '^Mem:' | awk '{print $3"/"$2}')
MEM_PERCENT=$(free | grep '^Mem:' | awk '{printf("%.1f", $3/$2 * 100)}')
echo "💾 Memory: $MEM_USAGE ($MEM_PERCENT%)"

CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1"%"}')
echo "🔥 CPU Load: $CPU_USAGE"

DISK_USAGE=$(df -h / | tail -1 | awk '{print $3"/"$2" ("$5" used)"}')
echo "💿 Disk: $DISK_USAGE"
echo ""

# Docker Status
echo "🐳 DOCKER STATUS"
echo "─────────────────────"
if systemctl is-active --quiet docker; then
    echo_success "Docker: Running"
    echo "📦 Version: $(docker --version | cut -d' ' -f3 | cut -d',' -f1)"
    
    RUNNING=$(docker ps -q | wc -l)
    TOTAL=$(docker ps -a -q | wc -l)
    IMAGES=$(docker images -q | wc -l)
    
    echo "🏃 Containers: $RUNNING running, $TOTAL total"
    echo "🖼️  Images: $IMAGES"
else
    echo_error "Docker: Not running"
fi
echo ""

# Codespaces
echo "🚀 CODESPACES"
echo "─────────────────────"
if [ -d ~/codespaces ]; then
    CODESPACE_COUNT=0
    RUNNING_COUNT=0
    
    for dir in ~/codespaces/*/; do
        basename=$(basename "$dir")
        if [[ ! "$basename" =~ ^(shared|auth|backups|scripts)$ ]] && [ -f "$dir/docker-compose.yml" ]; then
            ((CODESPACE_COUNT++))
            
            # Check if running
            if docker ps --format '{{.Names}}' | grep -q "$basename"; then
                ((RUNNING_COUNT++))
            fi
        fi
    done
    
    echo "📁 Total Codespaces: $CODESPACE_COUNT"
    echo "✅ Running: $RUNNING_COUNT"
    echo ""
    
    if [ $CODESPACE_COUNT -gt 0 ]; then
        echo "📋 CODESPACE LIST"
        echo "─────────────────────"
        
        for dir in ~/codespaces/*/; do
            basename=$(basename "$dir")
            if [[ ! "$basename" =~ ^(shared|auth|backups|scripts)$ ]] && [ -f "$dir/docker-compose.yml" ]; then
                if [ -f "$dir/.env" ]; then
                    VS_CODE_PORT=$(grep "VS_CODE_PORT=" "$dir/.env" | cut -d'=' -f2)
                    REPO_URL=$(grep "REPO_URL=" "$dir/.env" | cut -d'=' -f2)
                    
                    # Check container status
                    if docker ps --format '{{.Names}}' | grep -q "$basename"; then
                        STATUS="🟢 Running"
                        URL="http://localhost:$VS_CODE_PORT"
                    else
                        STATUS="⭕ Stopped"
                        URL="Port $VS_CODE_PORT"
                    fi
                    
                    echo ""
                    echo "📦 $basename"
                    echo "   Status: $STATUS"
                    echo "   VS Code: $URL"
                    echo "   Repo: $REPO_URL"
                fi
            fi
        done
    fi
else
    echo_warning "No codespaces directory found"
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "Quick commands:"
echo "  • list-codespaces    - List all codespaces"
echo "  • start-<name>       - Start a codespace"
echo "  • stop-<name>        - Stop a codespace"
echo "  • ~/setup-repo-codespace.sh <repo-url> - Create new codespace"
echo ""
EOF

chmod +x "$HOME/monitor-system.sh"

# List codespaces utility
mkdir -p "$HOME/codespaces/scripts/utils"
cat > "$HOME/codespaces/scripts/utils/list-codespaces.sh" << 'EOF'
#!/bin/bash

source "$HOME/codespaces/scripts/utils/colors.sh"

echo_info "📋 Available Codespaces"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

FOUND=0
for dir in ~/codespaces/*/; do
    basename=$(basename "$dir")
    if [[ ! "$basename" =~ ^(shared|auth|backups|scripts)$ ]] && [ -f "$dir/docker-compose.yml" ]; then
        FOUND=1
        
        if [ -f "$dir/.env" ]; then
            VS_CODE_PORT=$(grep "VS_CODE_PORT=" "$dir/.env" | cut -d'=' -f2)
            
            # Check if running
            if docker ps --format '{{.Names}}' | grep -q "$basename"; then
                echo_success "✅ $basename (running) - http://localhost:$VS_CODE_PORT"
            else
                echo_warning "⭕ $basename (stopped) - Port $VS_CODE_PORT"
            fi
        fi
    fi
done

if [ $FOUND -eq 0 ]; then
    echo_warning "No codespaces found"
    echo ""
    echo "Create one with:"
    echo "  ~/setup-repo-codespace.sh git@github.com:user/repo.git"
fi

echo ""
EOF

chmod +x "$HOME/codespaces/scripts/utils/list-codespaces.sh"

# Create manage-all script
mkdir -p "$HOME/codespaces/scripts/utils"
cat > "$HOME/codespaces/scripts/utils/manage-all.sh" << 'EOF'
#!/bin/bash

source "$HOME/codespaces/scripts/utils/colors.sh"

case "$1" in
    start-all)
        echo_step "🚀 Starting all codespaces..."
        for dir in ~/codespaces/*/; do
            if [ -f "$dir/docker-compose.yml" ]; then
                basename=$(basename "$dir")
                if [[ ! "$basename" =~ ^(shared|auth|backups|scripts)$ ]]; then
                    echo_info "Starting $basename..."
                    docker-compose -f "$dir/docker-compose.yml" up -d
                fi
            fi
        done
        echo_success "All codespaces started"
        ;;
        
    stop-all)
        echo_step "🛑 Stopping all codespaces..."
        for dir in ~/codespaces/*/; do
            if [ -f "$dir/docker-compose.yml" ]; then
                basename=$(basename "$dir")
                if [[ ! "$basename" =~ ^(shared|auth|backups|scripts)$ ]]; then
                    echo_info "Stopping $basename..."
                    docker-compose -f "$dir/docker-compose.yml" stop
                fi
            fi
        done
        echo_success "All codespaces stopped"
        ;;
        
    restart-all)
        $0 stop-all
        sleep 2
        $0 start-all
        ;;
        
    *)
        echo "Usage: $0 {start-all|stop-all|restart-all}"
        exit 1
        ;;
esac
EOF

chmod +x "$HOME/codespaces/scripts/utils/manage-all.sh"

echo_success "Monitoring tools created successfully"