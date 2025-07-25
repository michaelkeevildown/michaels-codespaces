#!/bin/bash

# Install whiptail if not present
# This provides Ubuntu installer-style menus for component selection

set -e

echo "Checking for whiptail..."

if command -v whiptail >/dev/null 2>&1; then
    echo "✅ Whiptail is already installed"
    exit 0
fi

if command -v dialog >/dev/null 2>&1; then
    echo "✅ Dialog is installed (can be used as alternative)"
    exit 0
fi

echo "ℹ️  Whiptail not found. Installing..."

# Detect package manager and install
if command -v apt-get >/dev/null 2>&1; then
    # Debian/Ubuntu
    echo "Installing whiptail via apt..."
    sudo apt-get update -qq
    sudo apt-get install -y whiptail
elif command -v yum >/dev/null 2>&1; then
    # RHEL/CentOS/Fedora
    echo "Installing newt (whiptail) via yum..."
    sudo yum install -y newt
elif command -v dnf >/dev/null 2>&1; then
    # Newer Fedora
    echo "Installing newt (whiptail) via dnf..."
    sudo dnf install -y newt
elif command -v brew >/dev/null 2>&1; then
    # macOS with Homebrew
    echo "Installing newt (whiptail) via brew..."
    brew install newt
else
    echo "⚠️  Unable to install whiptail automatically"
    echo ""
    echo "Please install whiptail manually:"
    echo "  Ubuntu/Debian: sudo apt-get install whiptail"
    echo "  RHEL/CentOS:   sudo yum install newt"
    echo "  Fedora:        sudo dnf install newt"
    echo "  macOS:         brew install newt"
    echo ""
    echo "Component selection will use text-based menus instead."
    exit 1
fi

echo "✅ Whiptail installed successfully"