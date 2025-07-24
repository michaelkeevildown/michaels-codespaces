# Michael's Codespaces - Architecture Documentation

## Overview

Michael's Codespaces uses a modular, enterprise-grade architecture designed for extensibility, maintainability, and scalability. The system is organized into distinct modules that handle specific aspects of codespace management.

## Directory Structure

```
.
├── bin/
│   └── mcs                          # Main CLI command
├── scripts/
│   ├── core/                        # Core functionality scripts
│   │   ├── create-codespace.sh      # Main codespace creation script
│   │   └── ...                      # Other core scripts
│   ├── modules/                     # Modular components
│   │   ├── github/                  # GitHub integration
│   │   │   ├── auth/               # Authentication handling
│   │   │   ├── api/                # API interactions
│   │   │   ├── clone/              # Repository cloning
│   │   │   └── webhooks/           # Webhook support
│   │   ├── docker/                  # Docker management
│   │   │   ├── compose/            # Docker Compose generation
│   │   │   ├── images/             # Image management
│   │   │   ├── containers/         # Container lifecycle
│   │   │   └── networks/           # Network configuration
│   │   ├── networking/              # Network and port management
│   │   ├── security/                # Security and credentials
│   │   ├── storage/                 # Data persistence
│   │   └── monitoring/              # Logging and monitoring
│   └── utils/                       # Utility functions
└── docs/                            # Documentation

```

## Module Descriptions

### GitHub Module (`modules/github/`)

Handles all GitHub-related operations:

- **auth/**: GitHub authentication (SSH keys, PATs)
- **clone/**: Repository cloning with retry logic
- **api/**: GitHub API interactions (future)
- **webhooks/**: Webhook processing (future)

#### Key Functions:
- `check_github_token()` - Validate GitHub token
- `convert_to_auth_url()` - Convert URLs for authenticated access
- `clone_with_retry()` - Clone repositories with retry logic
- `check_devcontainer()` - Detect .devcontainer configuration

### Docker Module (`modules/docker/`)

Manages Docker containers and configurations:

- **compose/**: Docker Compose file generation
- **images/**: Image management and caching
- **containers/**: Container lifecycle management
- **networks/**: Network isolation and configuration

#### Key Functions:
- `generate_basic_compose()` - Generate standard docker-compose.yml
- `generate_language_compose()` - Language-specific configurations
- `validate_compose()` - Validate generated configurations

### Networking Module (`modules/networking/`)

Handles port allocation and network management:

- Port registry to track allocations
- Automatic port discovery
- Port conflict resolution

#### Key Functions:
- `find_available_port()` - Discover available ports
- `allocate_codespace_ports()` - Allocate ports for a codespace
- `register_port()` - Track port usage
- `cleanup_stale_ports()` - Remove unused port registrations

### Security Module (`modules/security/`)

Manages credentials and security:

- Secure password generation
- Credential storage with encryption
- Token management

#### Key Functions:
- `generate_password()` - Create secure passwords
- `store_credential()` - Securely store credentials
- `encrypt_credential()` - Encrypt sensitive data

### Monitoring Module (`modules/monitoring/`)

Provides logging and monitoring:

- Structured logging
- Log rotation
- Operation tracking

#### Key Functions:
- `log_operation()` - Log codespace operations
- `log_error()` - Log errors with context

## Core Scripts

### create-codespace.sh

The modular creation script that:
1. Validates repository URLs
2. Supports custom Docker images
3. Detects project languages
4. Handles .devcontainer.json files
5. Manages port allocation
6. Creates comprehensive documentation

### mcs Command

The main CLI interface providing:
- Homebrew-style command structure
- Comprehensive help system
- Health checks (doctor command)
- Cleanup and uninstall options

## Data Flow

1. **User runs**: `mcs create <repo-url> [options]`
2. **CLI validates** input and calls create-codespace.sh
3. **Creation script**:
   - Validates repository access (GitHub module)
   - Clones repository (GitHub module)
   - Detects language/framework
   - Allocates ports (Networking module)
   - Generates credentials (Security module)
   - Creates docker-compose.yml (Docker module)
   - Starts container
   - Logs operation (Monitoring module)

## Configuration Files

### .env File
Stores codespace configuration:
```
CODESPACE_NAME=example-project
REPO_URL=git@github.com:user/repo.git
VS_CODE_PORT=8080
PASSWORD=secure-password
DOCKER_IMAGE=custom-image:latest
```

### docker-compose.yml
Generated based on:
- Project language detection
- .devcontainer.json (if present)
- User-specified options
- Security best practices

## Security Considerations

1. **Credentials**: Stored with 600 permissions
2. **Tokens**: Never logged or exposed
3. **Passwords**: Generated using cryptographically secure methods
4. **Network**: Isolated Docker networks per codespace
5. **Volumes**: Read-only mounts for sensitive data

## Extensibility

The modular architecture allows easy extension:

1. **New Languages**: Add to `generate_language_compose()`
2. **New Git Providers**: Extend GitHub module
3. **Custom Images**: Supported via --image flag
4. **Environment Variables**: Via --env-file option
5. **Port Mappings**: Custom via --ports flag

## Future Enhancements

1. **Kubernetes Support**: Deploy to k8s clusters
2. **Multi-container**: Support complex applications
3. **Remote Development**: Cloud-based codespaces
4. **Team Features**: Shared codespaces, permissions
5. **UI Dashboard**: Web-based management interface