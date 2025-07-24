#!/bin/bash
# Michael's Codespaces Setup - Homebrew Style
# Simple, reliable, no fancy features - just like Homebrew

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

# Homebrew-style output functions (exactly like brew)
ohai() {
    echo "${tty_blue}==>${tty_bold} $*${tty_reset}"
}

# Setup modules
echo ""
ohai "Installing Michael's Codespaces"

# 1. System packages
ohai "Installing system packages..."
"$SCRIPT_DIR/01-system-packages.sh"

# 2. Docker
ohai "Installing Docker..."
"$SCRIPT_DIR/02-docker-setup.sh"

# 3. Shell
ohai "Configuring shell environment..."
"$SCRIPT_DIR/03-shell-setup.sh"

# 4. Infrastructure
ohai "Creating codespace infrastructure..."
"$SCRIPT_DIR/04-codespace-infrastructure.sh"

# 5. Monitoring
ohai "Installing monitoring tools..."
"$SCRIPT_DIR/05-monitoring-tools.sh"

# 6. GitHub
ohai "Configuring GitHub access..."
"$SCRIPT_DIR/06-github-auth.sh"

# Done
echo ""
ohai "Installation successful!"
echo ""
echo "Next steps:"
echo "  - Logout and login again for Docker permissions"
echo "  - Run 'mcs doctor' to verify installation"
echo "  - Create a codespace with 'mcs create <repo-url>'"
echo ""