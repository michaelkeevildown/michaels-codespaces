#!/bin/bash

# Container Management Module
# Handles Docker container lifecycle operations

# Start a codespace container
start_container() {
    local codespace_dir="$1"
    local compose_file="$codespace_dir/docker-compose.yml"
    
    if [ ! -f "$compose_file" ]; then
        echo_error "Docker compose file not found: $compose_file"
        return 1
    fi
    
    echo_info "Starting container..."
    if docker-compose -f "$compose_file" up -d; then
        echo_success "Container started successfully"
        wait_for_container_ready "$codespace_dir"
        return 0
    else
        echo_error "Failed to start container"
        return 1
    fi
}

# Stop a codespace container
stop_container() {
    local codespace_dir="$1"
    local compose_file="$codespace_dir/docker-compose.yml"
    
    if [ ! -f "$compose_file" ]; then
        echo_error "Docker compose file not found: $compose_file"
        return 1
    fi
    
    echo_info "Stopping container..."
    if docker-compose -f "$compose_file" stop; then
        echo_success "Container stopped successfully"
        return 0
    else
        echo_error "Failed to stop container"
        return 1
    fi
}

# Restart a codespace container
restart_container() {
    local codespace_dir="$1"
    
    echo_info "Restarting container..."
    if stop_container "$codespace_dir" && start_container "$codespace_dir"; then
        return 0
    else
        return 1
    fi
}

# Remove a codespace container and its resources
remove_container() {
    local codespace_dir="$1"
    local compose_file="$codespace_dir/docker-compose.yml"
    
    if [ ! -f "$compose_file" ]; then
        echo_warning "Docker compose file not found, skipping container removal"
        return 0
    fi
    
    echo_info "Removing container and resources..."
    if docker-compose -f "$compose_file" down -v --remove-orphans; then
        echo_success "Container removed successfully"
        return 0
    else
        echo_error "Failed to remove container"
        return 1
    fi
}

# Get container status
get_container_status() {
    local codespace_dir="$1"
    local compose_file="$codespace_dir/docker-compose.yml"
    
    if [ ! -f "$compose_file" ]; then
        echo "not-found"
        return
    fi
    
    # Get container name from docker-compose
    local container_name=$(docker-compose -f "$compose_file" ps -q 2>/dev/null)
    
    if [ -z "$container_name" ]; then
        echo "stopped"
        return
    fi
    
    # Check if container is running
    if docker ps -q --no-trunc | grep -q "^${container_name}$"; then
        echo "running"
    else
        echo "stopped"
    fi
}

# Execute command in container
exec_in_container() {
    local codespace_dir="$1"
    shift
    local command="$@"
    local compose_file="$codespace_dir/docker-compose.yml"
    
    if [ ! -f "$compose_file" ]; then
        echo_error "Docker compose file not found: $compose_file"
        return 1
    fi
    
    # Default to bash if no command specified
    if [ -z "$command" ]; then
        command="bash"
    fi
    
    docker-compose -f "$compose_file" exec -it dev $command
}

# View container logs
view_container_logs() {
    local codespace_dir="$1"
    local follow="${2:-false}"
    local compose_file="$codespace_dir/docker-compose.yml"
    
    if [ ! -f "$compose_file" ]; then
        echo_error "Docker compose file not found: $compose_file"
        return 1
    fi
    
    local log_args=""
    if [ "$follow" = "true" ]; then
        log_args="-f"
    else
        log_args="--tail=50"
    fi
    
    docker-compose -f "$compose_file" logs $log_args
}

# Wait for container to be ready
wait_for_container_ready() {
    local codespace_dir="$1"
    local max_attempts=30
    local attempt=0
    
    echo_info "Waiting for container to be ready..."
    
    # Get VS Code port from .env file
    local vs_code_port=$(grep "VS_CODE_PORT=" "$codespace_dir/.env" 2>/dev/null | cut -d'=' -f2)
    if [ -z "$vs_code_port" ]; then
        vs_code_port="8080"
    fi
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s -o /dev/null "http://localhost:$vs_code_port" 2>/dev/null; then
            echo_success "Container is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    echo ""
    echo_warning "Container may not be fully ready yet"
    return 0
}

# Get container resource usage
get_container_stats() {
    local codespace_dir="$1"
    local compose_file="$codespace_dir/docker-compose.yml"
    
    if [ ! -f "$compose_file" ]; then
        echo_error "Docker compose file not found: $compose_file"
        return 1
    fi
    
    local container_name=$(docker-compose -f "$compose_file" ps -q 2>/dev/null)
    
    if [ -z "$container_name" ]; then
        echo_error "Container not running"
        return 1
    fi
    
    docker stats --no-stream "$container_name"
}

# Clean up stopped containers and unused resources
cleanup_containers() {
    echo_info "Cleaning up Docker resources..."
    
    # Remove stopped containers
    local stopped_containers=$(docker ps -a -q -f status=exited)
    if [ -n "$stopped_containers" ]; then
        echo_info "Removing stopped containers..."
        docker rm $stopped_containers
    fi
    
    # Remove unused volumes
    echo_info "Removing unused volumes..."
    docker volume prune -f
    
    # Remove unused networks
    echo_info "Removing unused networks..."
    docker network prune -f
    
    echo_success "Cleanup completed"
}

# Export functions
export -f start_container
export -f stop_container
export -f restart_container
export -f remove_container
export -f get_container_status
export -f exec_in_container
export -f view_container_logs
export -f wait_for_container_ready
export -f get_container_stats
export -f cleanup_containers