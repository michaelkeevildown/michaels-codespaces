#!/bin/bash

# System checks and validations

check_root() {
    if [ "$EUID" -eq 0 ]; then
        echo_error "Don't run this script as root. Run as regular user."
        exit 1
    fi
}

check_ubuntu() {
    if ! command -v lsb_release &> /dev/null; then
        echo_error "This script requires Ubuntu. Please run on Ubuntu 20.04 or later."
        exit 1
    fi
    
    UBUNTU_VERSION=$(lsb_release -rs)
    if (( $(echo "$UBUNTU_VERSION < 20.04" | bc -l) )); then
        echo_error "Ubuntu 20.04 or later required. Found: $UBUNTU_VERSION"
        exit 1
    fi
}

check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo_error "Command '$1' not found. Please install it and try again."
        return 1
    fi
    return 0
}

check_docker() {
    if ! systemctl is-active --quiet docker; then
        echo_warning "Docker is not running. It will be started after installation."
        return 1
    fi
    return 0
}

check_port() {
    local port=$1
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        echo_warning "Port $port is already in use"
        return 1
    fi
    return 0
}