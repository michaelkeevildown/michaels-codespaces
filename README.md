# 🚀 Michael's Codespaces

> **Your own GitHub Codespaces, running on YOUR servers!** 
> 
> Transform any Ubuntu machine into a powerful development platform. Create isolated, browser-based VS Code environments for all your GitHub repositories. Work from anywhere, on any device, with just a web browser.

<p align="center">
  <img src="https://img.shields.io/badge/Ubuntu-20.04+-E95420?style=for-the-badge&logo=ubuntu&logoColor=white" alt="Ubuntu">
  <img src="https://img.shields.io/badge/Docker-20.10+-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/VS_Code-Browser-007ACC?style=for-the-badge&logo=visual-studio-code&logoColor=white" alt="VS Code">
</p>

---

## ✨ What's This All About?

Ever wished you could spin up a fresh development environment for each project without polluting your main system? Want to code from your iPad while your dev machine sits at home? Need to keep client projects completely isolated from each other?

**Michael's Codespaces** brings the magic of GitHub Codespaces to your own infrastructure:

```
Your Ubuntu Server + Our Magic = 
    ↓
💻 Browser-based VS Code for every repo
🐳 Complete isolation between projects  
🌍 Access from anywhere (beach coding, anyone?)
💾 Everything persists between sessions
🔒 Your code never leaves your servers
```

## 🎯 One-Line Install

Just like Homebrew, but for development environments:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/install.sh)"
```

That's it! In about 2 minutes, you'll have a fully functional codespace platform. ☕

## 🎮 How It Works

### 1️⃣ **Install** (2 minutes)
Run the command above. It sets up Docker, configures your system, and installs the `mcs` command.

### 2️⃣ **Authenticate** (30 seconds)
```bash
mcs doctor  # Check everything is working
~/codespaces/shared/scripts/setup-github-auth.sh
```
Add your SSH keys and GitHub token. Done once, works everywhere.

### 3️⃣ **Create Codespaces** (10 seconds per repo)
```bash
mcs create git@github.com:torvalds/linux.git
```

### 4️⃣ **Start Coding!**
```
🎉 Codespace created successfully!

📦 Codespace: torvalds-linux
🌐 VS Code URL: http://localhost:8080
🔑 Password: kT9mN3pQ2xR5wY7v
```

Open that URL and boom - you're coding in a fully isolated Linux kernel dev environment!

## 🎬 Real-World Example

Let's say you're working on three projects - watch the magic of auto-detection:

```bash
# Node.js project - auto-detects package.json
mcs create git@github.com:facebook/react.git
# ✅ Auto-configures: Node 20, npm/yarn ready, port 3000 forwarded

# Python project - detects requirements.txt or pyproject.toml  
mcs create git@github.com:python/cpython.git
# ✅ Auto-configures: Python 3.11, pip ready, virtual env setup

# Go project - finds go.mod
mcs create git@github.com:golang/go.git
# ✅ Auto-configures: Go 1.21, modules ready, proper GOPATH

# Multi-language project with .devcontainer
mcs create git@github.com:microsoft/vscode.git
# ✅ Uses .devcontainer.json config automatically
# ✅ Respects image, ports, environment vars from devcontainer

# Force specific language when auto-detection isn't enough
mcs create --language rust git@github.com:user/mixed-repo.git
# ✅ Forces Rust environment even if other files present
```

Each gets its own container, own ports, own dependencies. Switch between them instantly:

```bash
mcs list

NAME                          STATUS     URL
----                          ------     ---
clienta-legacy-app            running    http://localhost:8080
you-ai-startup                running    http://localhost:8081  
facebook-react                stopped    port 8082
```

## 🛠️ Daily Commands

Once installed, everything is managed through the `mcs` command:

```bash
# Core workflow
mcs create git@github.com:user/repo.git   # Create new codespace
mcs list (or mcs ls)                       # Show all codespaces
mcs start myproject                        # Fire it up
mcs stop myproject                         # Stop when done

# Enhanced creation with options
mcs create --language node git@github.com:user/app.git      # Auto-configure for Node.js
mcs create --image custom:latest git@github.com:user/repo   # Use custom Docker image
mcs create --ports "8090:8080,3001:3000" git@github.com:user/repo  # Custom port mapping
mcs create --name my-project git@github.com:user/repo       # Custom codespace name
mcs create --no-start git@github.com:user/repo              # Create but don't start
mcs create --force git@github.com:user/repo                 # Overwrite existing

# Development
mcs exec myproject                         # Jump into container shell
mcs exec myproject npm install             # Run commands directly
mcs logs myproject                         # Check what's happening

# Management  
mcs info myproject                         # Show details & resource usage
mcs restart myproject                      # Quick restart
mcs remove (or mcs rm) myproject           # Delete when done

# System health & cleanup
mcs status                                 # Full system overview
mcs doctor                                 # Comprehensive health check
mcs cleanup                                # Clean up stopped containers & unused resources
mcs destroy                                # Nuclear option - remove everything
mcs update                                 # Update to latest version
```

## 🌟 Cool Features

### 🏖️ **Code From Anywhere**
- **At home**: Direct access via `localhost`
- **At coffee shop**: Use Cloudflare tunnels (secure!)
- **On iPad/phone**: Just need a browser
- **From work**: SSH tunnel through corporate firewall

### 🎯 **Smart Isolation**
- Each project gets its own container
- No more "works on my machine" 
- Break things without fear
- Different Node/Python/Ruby versions? No problem!

### ⚡ **Developer Experience**
- Full VS Code in your browser
- All your extensions work
- Integrated terminal
- Git already configured
- Port forwarding just works
- **Smart language detection** - auto-configures Node, Python, Go, Rust, Java, Ruby, PHP, .NET
- **DevContainer support** - respects your `.devcontainer.json` settings
- **Flexible port mapping** - custom ports for any service
- **Environment file support** - load custom env vars from files

### 🔐 **Security First**
- Your code stays on YOUR server
- SSH key authentication
- Unique passwords per codespace
- Network isolation between projects
- No external dependencies

### 🧹 **System Management**
- **Health monitoring** with `mcs doctor` - checks Docker, ports, auth, resources
- **Smart cleanup** with `mcs cleanup` - removes stopped containers & unused images
- **Emergency reset** with `mcs destroy` - complete system reset when needed
- **Resource tracking** with `mcs status` - overview of all codespaces and usage
- **Port management** - automatic conflict detection and resolution

## 📊 System Requirements

**Minimum** (2-3 codespaces):
- Ubuntu 20.04+
- 4GB RAM
- 20GB storage
- Basic VPS ($5-10/month)

**Recommended** (10+ codespaces):
- Ubuntu 22.04
- 16GB RAM  
- 100GB SSD
- Dedicated server or beefy VPS

**Go Crazy** (50+ codespaces):
- Ubuntu 22.04
- 64GB+ RAM
- 500GB+ NVMe SSD
- Bare metal server

## 🚦 Quick Start Guide

```bash
# 1. Install (2 minutes)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/install.sh)"

# 2. Logout/login for Docker permissions
exit
ssh back-in

# 3. Check installation & setup GitHub auth
mcs doctor                                      # Verify everything is ready
~/codespaces/shared/scripts/setup-github-auth.sh   # Configure GitHub access

# 4. Create your first codespace
mcs create git@github.com:your/repo.git

# 5. List and manage
mcs list                                        # See your codespace
mcs start your-repo                             # Start if needed
# Open the URL shown in your browser!
```

### Development & Testing Installation

When testing new features or fixes, you can install from a specific branch:

```bash
# Install from a development branch
CODESPACE_BRANCH=fix-installation-directories /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/fix-installation-directories/install.sh)"

# Debug mode for troubleshooting
DEBUG=1 CODESPACE_BRANCH=my-feature /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/my-feature/install.sh)"

# Non-interactive installation (for scripts)
NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/install.sh)"
```

## 🎪 Advanced Usage

### DevContainer Integration

Michael's Codespaces automatically detects and uses `.devcontainer.json` configurations:

```bash
# Automatic devcontainer detection
mcs create git@github.com:microsoft/vscode.git
# ✅ Reads .devcontainer/devcontainer.json
# ✅ Uses specified Docker image
# ✅ Applies port mappings from config
# ✅ Sets up environment variables
# ✅ Runs postCreateCommand if specified

# Override devcontainer image if needed
mcs create --image custom:latest git@github.com:user/repo-with-devcontainer.git
```

**Example `.devcontainer.json` support:**
```json
{
  "image": "mcr.microsoft.com/devcontainers/python:3.11",
  "forwardPorts": [8080, 3000],
  "remoteEnv": {
    "PYTHONPATH": "/workspace"
  },
  "postCreateCommand": "pip install -r requirements.txt"
}
```
Michael's Codespaces will respect these settings automatically!

### Language-Specific Creation
```bash
# Force specific language environments
mcs create --language node git@github.com:user/frontend.git     # Node.js with npm/yarn
mcs create --language python git@github.com:user/ai-app.git    # Python with pip/poetry  
mcs create --language go git@github.com:user/service.git       # Go with modules
mcs create --language rust git@github.com:user/cli.git         # Rust with cargo
mcs create --language java git@github.com:user/spring.git      # Java with Maven
```

### Custom Docker Images & Ports
```bash
# Use your own Docker image
mcs create --image myregistry/custom:latest git@github.com:user/repo.git

# Custom port mapping for multiple services
mcs create --ports "8090:8080,3001:3000,5432:5432" git@github.com:user/fullstack.git

# Combine options for complex setups
mcs create --name my-project --language python --ports "8888:8080,5000:5000" --env-file .env.production git@github.com:user/ml-platform.git
```

### Environment Configuration
```bash
# Use environment files
echo "API_KEY=secret123" > ~/my-env-vars.txt
echo "DATABASE_URL=postgres://..." >> ~/my-env-vars.txt
mcs create --env-file ~/my-env-vars.txt git@github.com:user/app.git

# Create without auto-starting (for custom setup)
mcs create --no-start git@github.com:user/complex-setup.git
# ... do custom configuration ...
mcs start complex-setup
```

### Resource Limits & Custom Domains
```yaml
# Edit ~/codespaces/myproject/docker-compose.yml
services:
  myproject-dev:
    mem_limit: 2g
    cpus: 1.5
```

```nginx
# Nginx proxy for codespace.yourdomain.com
location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection upgrade;
}
```

## 🚨 Troubleshooting

### System Health Check
First, run the comprehensive health check:
```bash
mcs doctor
```
This checks Docker, ports, authentication, and system resources.

<details>
<summary><b>Docker permission denied?</b></summary>

You forgot to logout/login after install:
```bash
exit
ssh user@your-server
```
Or check if you're in the docker group:
```bash
groups | grep docker  # Should show 'docker'
```
</details>

<details>
<summary><b>Can't clone private repos?</b></summary>

Run the GitHub auth setup:
```bash
~/codespaces/shared/scripts/setup-github-auth.sh
ssh -T git@github.com  # Test it
```

Check if your SSH keys are loaded:
```bash
mcs doctor  # Will check GitHub authentication
```
</details>

<details>
<summary><b>Port already in use?</b></summary>

The system auto-finds free ports, but if there's a conflict:
```bash
# Check what's using the port
mcs status  # Shows port assignments
mcs info myproject  # Shows specific codespace ports

# Force new ports when creating
mcs create --ports "8090:8080,3001:3000" git@github.com:user/repo.git
```
</details>

<details>
<summary><b>Container won't start?</b></summary>

Check logs and container status:
```bash
mcs logs myproject        # View container logs
mcs doctor               # Check system health
docker ps -a             # See all containers

# Clean up and retry
mcs stop myproject
mcs start myproject

# Nuclear option
mcs remove myproject
mcs create --force git@github.com:user/repo.git
```
</details>

<details>
<summary><b>Running out of space/memory?</b></summary>

Clean up unused resources:
```bash
mcs cleanup              # Remove stopped containers & unused images
mcs status              # Check resource usage
mcs info myproject      # Check specific codespace resources

# See what's using space
docker system df
docker images --filter "dangling=true"
```
</details>

<details>
<summary><b>Language auto-detection wrong?</b></summary>

Force the correct language:
```bash
mcs create --language python git@github.com:user/repo.git
mcs create --image node:18 git@github.com:user/repo.git  # Specific image
```

Or check what was detected:
```bash
cat ~/codespaces/myproject/.env  # Shows detected settings
```
</details>

## 🤝 Contributing

Found a bug? Want a feature? PRs welcome!

```bash
mcs create git@github.com:michaelkeevildown/michaels-codespaces.git
# Now you're developing Michael's Codespaces... in Michael's Codespaces! 🤯
```

## 📜 License

MIT License - basically do whatever you want!

---

<p align="center">
  <b>Built with ❤️ by developers who got tired of "works on my machine"</b>
  <br>
  <a href="https://github.com/michaelkeevildown/michaels-codespaces/issues">Report Bug</a>
  ·
  <a href="https://github.com/michaelkeevildown/michaels-codespaces/issues">Request Feature</a>
  ·
  <a href="https://github.com/michaelkeevildown/michaels-codespaces">Star on GitHub</a>
</p>