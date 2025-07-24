#!/bin/bash

# Credentials Module
# Secure handling of passwords and tokens

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CREDENTIALS_DIR="$HOME/codespaces/.credentials"

# Initialize credentials directory
init_credentials_dir() {
    if [ ! -d "$CREDENTIALS_DIR" ]; then
        mkdir -p "$CREDENTIALS_DIR"
        chmod 700 "$CREDENTIALS_DIR"
    fi
}

# Generate secure password
generate_password() {
    local length="${1:-16}"
    local charset="${2:-A-Za-z0-9}"
    
    # Try different methods to generate password
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -base64 48 | tr -d "=+/" | cut -c1-$length
    elif command -v pwgen >/dev/null 2>&1; then
        pwgen -s $length 1
    elif [ -r /dev/urandom ]; then
        tr -dc "$charset" < /dev/urandom | head -c $length
    else
        # Fallback to less secure method
        date +%s%N | sha256sum | base64 | head -c $length
    fi
}

# Store credential securely
store_credential() {
    local name="$1"
    local value="$2"
    local codespace="${3:-global}"
    
    init_credentials_dir
    
    local cred_file="$CREDENTIALS_DIR/${codespace}.${name}"
    
    # Store with restricted permissions
    echo "$value" > "$cred_file"
    chmod 600 "$cred_file"
}

# Retrieve credential
get_credential() {
    local name="$1"
    local codespace="${2:-global}"
    
    local cred_file="$CREDENTIALS_DIR/${codespace}.${name}"
    
    if [ -f "$cred_file" ]; then
        cat "$cred_file"
    else
        echo ""
    fi
}

# Remove credential
remove_credential() {
    local name="$1"
    local codespace="${2:-global}"
    
    local cred_file="$CREDENTIALS_DIR/${codespace}.${name}"
    
    if [ -f "$cred_file" ]; then
        # Overwrite before deletion
        dd if=/dev/urandom of="$cred_file" bs=1024 count=1 2>/dev/null || true
        rm -f "$cred_file"
    fi
}

# Remove all credentials for a codespace
remove_codespace_credentials() {
    local codespace="$1"
    
    for cred_file in "$CREDENTIALS_DIR/${codespace}."*; do
        if [ -f "$cred_file" ]; then
            dd if=/dev/urandom of="$cred_file" bs=1024 count=1 2>/dev/null || true
            rm -f "$cred_file"
        fi
    done
}

# Encrypt credential (using openssl if available)
encrypt_credential() {
    local value="$1"
    local passphrase="${2:-$(get_master_passphrase)}"
    
    if command -v openssl >/dev/null 2>&1; then
        echo "$value" | openssl enc -aes-256-cbc -a -salt -pass pass:"$passphrase" 2>/dev/null
    else
        # No encryption available, encode only
        echo "$value" | base64
    fi
}

# Decrypt credential
decrypt_credential() {
    local encrypted="$1"
    local passphrase="${2:-$(get_master_passphrase)}"
    
    if command -v openssl >/dev/null 2>&1; then
        echo "$encrypted" | openssl enc -aes-256-cbc -d -a -pass pass:"$passphrase" 2>/dev/null || echo "$encrypted"
    else
        # No encryption available, decode only
        echo "$encrypted" | base64 -d 2>/dev/null || echo "$encrypted"
    fi
}

# Get or generate master passphrase
get_master_passphrase() {
    local master_file="$CREDENTIALS_DIR/.master"
    
    if [ ! -f "$master_file" ]; then
        init_credentials_dir
        generate_password 32 > "$master_file"
        chmod 600 "$master_file"
    fi
    
    cat "$master_file"
}

# Validate password strength
validate_password_strength() {
    local password="$1"
    local min_length="${2:-12}"
    
    # Check length
    if [ ${#password} -lt $min_length ]; then
        echo "Password too short (minimum $min_length characters)"
        return 1
    fi
    
    # Check complexity
    local has_lower=0
    local has_upper=0
    local has_digit=0
    local has_special=0
    
    if [[ "$password" =~ [a-z] ]]; then has_lower=1; fi
    if [[ "$password" =~ [A-Z] ]]; then has_upper=1; fi
    if [[ "$password" =~ [0-9] ]]; then has_digit=1; fi
    if [[ "$password" =~ [^a-zA-Z0-9] ]]; then has_special=1; fi
    
    local complexity=$((has_lower + has_upper + has_digit + has_special))
    
    if [ $complexity -lt 3 ]; then
        echo "Password not complex enough (needs at least 3 of: lowercase, uppercase, digits, special characters)"
        return 1
    fi
    
    return 0
}

# Export functions
export -f generate_password
export -f store_credential
export -f get_credential
export -f remove_credential
export -f remove_codespace_credentials
export -f encrypt_credential
export -f decrypt_credential
export -f validate_password_strength