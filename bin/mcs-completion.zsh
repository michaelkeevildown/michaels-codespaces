#compdef mcs

# Zsh completion for mcs (Michael's Codespaces)

_mcs() {
    local -a commands
    commands=(
        'create:Create a new codespace from a GitHub repository'
        'list:List all codespaces and their status'
        'start:Start a codespace'
        'stop:Stop a running codespace'
        'restart:Restart a codespace'
        'remove:Remove a codespace permanently'
        'logs:View logs for a codespace'
        'exec:Execute a command in a codespace'
        'info:Show detailed information about a codespace'
        'status:Show system and codespaces status'
        'update:Update Michael'"'"'s Codespaces to latest version'
        'doctor:Check system health and configuration'
        'help:Show help message'
    )
    
    local -a aliases
    aliases=(
        'new:Alias for create'
        'ls:Alias for list'
        'up:Alias for start'
        'down:Alias for stop'
        'rm:Alias for remove'
        'delete:Alias for remove'
        'log:Alias for logs'
        'run:Alias for exec'
        'show:Alias for info'
        'monitor:Alias for status'
        'upgrade:Alias for update'
        'check:Alias for doctor'
    )
    
    _arguments '1: :->command' '2: :->args'
    
    case $state in
        command)
            _describe 'command' commands
            _describe 'alias' aliases
            ;;
        args)
            case $words[2] in
                start|stop|restart|remove|rm|delete|logs|log|exec|run|info|show|up|down)
                    # Get codespace names
                    local -a codespaces
                    if [ -d "$HOME/codespaces" ]; then
                        for dir in "$HOME/codespaces"/*/; do
                            if [ -f "$dir/docker-compose.yml" ]; then
                                local name=$(basename "$dir")
                                if [[ ! "$name" =~ ^(shared|auth|backups|scripts)$ ]]; then
                                    codespaces+=($name)
                                fi
                            fi
                        done
                    fi
                    _describe 'codespace' codespaces
                    ;;
                create|new)
                    _message 'repository URL'
                    ;;
            esac
            ;;
    esac
}

_mcs "$@"