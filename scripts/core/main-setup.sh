#!/bin/bash

# Main setup orchestrator
# This script calls all the modular components in the correct order

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source utilities
source "$SCRIPT_DIR/../utils/colors.sh"
source "$SCRIPT_DIR/../utils/checks.sh"

echo_info "🚀 Starting Michael's Codespaces setup..."
echo ""

# Run setup modules in order
"$SCRIPT_DIR/01-system-packages.sh"
echo ""

"$SCRIPT_DIR/02-docker-setup.sh"
echo ""

"$SCRIPT_DIR/03-shell-setup.sh"
echo ""

"$SCRIPT_DIR/04-codespace-infrastructure.sh"
echo ""

"$SCRIPT_DIR/05-monitoring-tools.sh"
echo ""

"$SCRIPT_DIR/06-github-auth.sh"
echo ""

echo_success "✅ Main setup completed successfully!"