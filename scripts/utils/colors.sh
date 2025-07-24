#!/bin/bash

# Color definitions for consistent output

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Echo functions
echo_error() {
    echo -e "${RED}❌ $1${NC}" >&2
}

echo_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

echo_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

echo_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

echo_step() {
    echo -e "${PURPLE}▶️  $1${NC}"
}

echo_debug() {
    if [ "${DEBUG:-false}" = "true" ]; then
        echo -e "${CYAN}🔍 $1${NC}"
    fi
}