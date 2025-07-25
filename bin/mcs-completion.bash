#!/bin/bash

# Bash completion for mcs (Michael's Codespaces)

_mcs_completion() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Main commands
    commands="create list start stop restart remove logs exec info status update autoupdate update-ip recover reset-password cleanup destroy doctor help"
    
    # Aliases
    aliases="new ls up down rm delete log run show monitor upgrade ip clean uninstall check"
    
    case "${COMP_CWORD}" in
        1)
            # Complete commands
            COMPREPLY=( $(compgen -W "${commands} ${aliases}" -- ${cur}) )
            ;;
        2)
            # Complete based on the command
            case "${prev}" in
                start|stop|restart|remove|rm|delete|logs|log|exec|run|info|show|up|down|recover|reset-password)
                    # Get list of codespace names
                    local codespaces=""
                    if [ -d "$HOME/codespaces" ]; then
                        for dir in "$HOME/codespaces"/*/; do
                            if [ -f "$dir/docker-compose.yml" ]; then
                                local name=$(basename "$dir")
                                if [[ ! "$name" =~ ^(shared|auth|backups|scripts)$ ]]; then
                                    codespaces="${codespaces} ${name}"
                                fi
                            fi
                        done
                    fi
                    COMPREPLY=( $(compgen -W "${codespaces}" -- ${cur}) )
                    ;;
                autoupdate)
                    # Autoupdate subcommands
                    COMPREPLY=( $(compgen -W "status on off enable disable interval help" -- ${cur}) )
                    ;;
            esac
            ;;
    esac
    
    return 0
}

complete -F _mcs_completion mcs