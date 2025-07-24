# Michael's Codespaces Scripts

This directory contains the modular script architecture for Michael's Codespaces. The system is designed with enterprise-grade modularity, reusability, and maintainability in mind.

## Directory Structure

```
scripts/
├── core/                           # Core functionality scripts
│   └── create-codespace.sh        # Main codespace creation orchestrator
├── modules/                        # Reusable module components
│   ├── docker/                    # Docker-related modules
│   │   ├── compose/              # Docker Compose generation
│   │   │   └── docker-compose-generator.sh
│   │   ├── containers/           # Container lifecycle management
│   │   │   ├── container-manager.sh
│   │   │   └── management-scripts.sh
│   │   └── images/               # Image detection and management
│   │       └── language-detector.sh
│   ├── github/                    # GitHub integration modules
│   │   ├── api/                  # API interactions
│   │   │   └── repo-validator.sh
│   │   ├── auth/                 # Authentication handling
│   │   │   └── github-auth.sh
│   │   └── clone/                # Repository cloning
│   │       └── github-clone.sh
│   ├── networking/                # Network and port management
│   │   └── port-manager.sh
│   ├── security/                  # Security and credentials
│   │   └── credentials.sh
│   ├── storage/                   # Data persistence
│   │   └── env-manager.sh
│   └── monitoring/                # Logging and monitoring
│       └── logger.sh
└── utils/                         # Utility functions
    └── colors.sh                  # Terminal color helpers
```

## Core Scripts

### create-codespace.sh
The main orchestrator that coordinates all modules to create a codespace. It:
- Validates repository URLs
- Clones repositories
- Detects project languages
- Allocates ports
- Generates Docker configurations
- Creates management scripts
- Starts containers

Usage:
```bash
create-codespace.sh <repository-url> [options]
```

## Module Descriptions

### Docker Modules

#### docker/compose/docker-compose-generator.sh
Generates Docker Compose configurations based on:
- Project language
- Custom settings
- Security best practices

Key functions:
- `generate_basic_compose()` - Basic VS Code configuration
- `generate_language_compose()` - Language-specific setups
- `validate_compose()` - Validates generated files

#### docker/images/language-detector.sh
Detects project languages and recommends Docker images.

Key functions:
- `detect_language()` - Analyzes project files
- `get_language_image()` - Returns appropriate Docker image
- `parse_devcontainer_image()` - Extracts image from devcontainer.json
- `check_devcontainer()` - Finds devcontainer configuration

#### docker/containers/container-manager.sh
Manages Docker container lifecycle.

Key functions:
- `start_container()` - Starts a codespace container
- `stop_container()` - Stops a container gracefully
- `restart_container()` - Restarts a container
- `remove_container()` - Removes container and resources
- `get_container_status()` - Returns container state
- `exec_in_container()` - Executes commands in container
- `view_container_logs()` - Shows container logs
- `wait_for_container_ready()` - Waits for container startup

#### docker/containers/management-scripts.sh
Creates management scripts and documentation.

Key functions:
- `create_management_scripts()` - Main wrapper function
- `create_codespace_readme()` - Generates README.md
- `create_shell_aliases()` - Creates shell shortcuts
- `register_aliases_in_shell()` - Adds to .zshrc/.bashrc
- `display_codespace_success()` - Shows success message

### GitHub Modules

#### github/api/repo-validator.sh
Validates and parses repository information.

Key functions:
- `validate_repo_url()` - Checks URL format
- `parse_repo_url()` - Extracts owner/name/provider
- `check_repo_access()` - Verifies repository access
- `get_repo_metadata()` - Fetches repo information
- `convert_to_authenticated_url()` - Adds token to URL
- `check_github_rate_limit()` - Monitors API limits

#### github/auth/github-auth.sh
Handles GitHub authentication.

Key functions:
- `check_github_token()` - Validates tokens
- `setup_github_auth()` - Configures authentication
- `validate_repo_access()` - Checks access permissions

#### github/clone/github-clone.sh
Clones repositories with retry logic.

Key functions:
- `clone_with_retry()` - Clones with automatic retries
- `convert_to_auth_url()` - Converts URLs for auth

### Storage Modules

#### storage/env-manager.sh
Manages environment configuration files.

Key functions:
- `create_env_file()` - Creates .env files
- `load_env_file()` - Loads environment variables
- `update_env_value()` - Updates specific values
- `get_env_value()` - Retrieves values
- `validate_env_file()` - Validates configuration
- `backup_env_file()` - Creates backups

### Networking Modules

#### networking/port-manager.sh
Manages port allocation and tracking.

Key functions:
- `find_available_port()` - Finds free ports
- `allocate_codespace_ports()` - Allocates port sets
- `register_port()` - Records port usage
- `unregister_port()` - Releases ports
- `cleanup_stale_ports()` - Removes unused entries

### Security Modules

#### security/credentials.sh
Handles password and credential management.

Key functions:
- `generate_password()` - Creates secure passwords
- `store_credential()` - Securely stores credentials
- `encrypt_credential()` - Encrypts sensitive data

### Monitoring Modules

#### monitoring/logger.sh
Provides structured logging capabilities.

Key functions:
- `log_operation()` - Logs codespace operations
- `log_error()` - Logs errors with context

## Usage Examples

### Creating a Codespace
```bash
# Basic usage
./scripts/core/create-codespace.sh git@github.com:user/repo.git

# With custom name
./scripts/core/create-codespace.sh https://github.com/user/repo.git --name my-project

# With specific language
./scripts/core/create-codespace.sh git@github.com:user/repo.git --language node

# With custom ports
./scripts/core/create-codespace.sh git@github.com:user/repo.git --ports "8090:8080,3001:3000"
```

### Using Modules Directly
```bash
# Source a module
source scripts/modules/docker/images/language-detector.sh

# Detect language
language=$(detect_language "/path/to/project")
echo "Detected: $language"

# Get recommended image
image=$(get_language_image "$language")
echo "Recommended image: $image"
```

## Adding New Features

### Adding a New Language
1. Update `detect_language()` in `language-detector.sh`
2. Add image mapping in `get_language_image()`
3. Add language-specific compose generation in `docker-compose-generator.sh`
4. Add environment variables in `get_language_env_vars()`

### Adding a New Module
1. Create module file in appropriate directory
2. Add functions with clear names
3. Export functions at the end
4. Source in scripts that need it
5. Document in this README

## Best Practices

1. **Function Naming**: Use descriptive names that indicate the module
2. **Error Handling**: Always check return codes
3. **Logging**: Use echo_* functions for consistent output
4. **Parameters**: Validate all input parameters
5. **Documentation**: Comment complex logic
6. **Exports**: Only export public functions

## Environment Variables

Key environment variables used:
- `CODESPACE_HOME` - Base directory for codespaces
- `GITHUB_TOKEN` - GitHub authentication token
- `DEBUG` - Enable debug output (0/1)

## Security Considerations

- Credentials stored with 600 permissions
- Tokens never logged or displayed
- Passwords generated cryptographically
- Network isolation between codespaces

## Contributing

When modifying scripts:
1. Test changes thoroughly
2. Update relevant documentation
3. Maintain backward compatibility
4. Follow existing code style
5. Add error handling

## Troubleshooting

Common issues:
- **Port conflicts**: Check port registry
- **Authentication failures**: Verify GitHub token
- **Module not found**: Check sourcing paths
- **Permission denied**: Check file permissions