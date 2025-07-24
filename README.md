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

Let's say you're working on three projects:

```bash
# Client project with ancient dependencies
mcs create git@github.com:clientA/legacy-app.git
# ✅ Runs Node 8, PHP 5.6, MySQL 5.5 - totally isolated!

# Your modern side project  
mcs create git@github.com:you/ai-startup.git
# ✅ Runs Node 20, Python 3.11, PostgreSQL 15 - no conflicts!

# Contributing to open source
mcs create git@github.com:facebook/react.git
# ✅ Perfect dev environment without touching your system!
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
mcs list                                   # Show all codespaces
mcs start myproject                        # Fire it up
mcs stop myproject                         # Stop when done

# Development
mcs exec myproject                         # Jump into container shell
mcs exec myproject npm install             # Run commands directly
mcs logs myproject                         # Check what's happening

# Management  
mcs info myproject                         # Show details & resource usage
mcs restart myproject                      # Quick restart
mcs remove myproject                       # Delete when done

# System
mcs status                                 # Full system overview
mcs doctor                                 # Check health
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

### 🔐 **Security First**
- Your code stays on YOUR server
- SSH key authentication
- Unique passwords per codespace
- Network isolation between projects
- No external dependencies

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

## 🎪 Advanced Tricks

### Custom Domains
```nginx
# Nginx proxy for codespace.yourdomain.com
location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection upgrade;
}
```

### Resource Limits
```yaml
# Edit ~/codespaces/myproject/docker-compose.yml
services:
  myproject-dev:
    mem_limit: 2g
    cpus: 1.5
```

### Custom Images
```dockerfile
# Use different base image for specific needs
FROM codercom/code-server:latest-cuda  # GPU support!
```

## 🚨 Troubleshooting

<details>
<summary><b>Docker permission denied?</b></summary>

You forgot to logout/login after install:
```bash
exit
ssh user@your-server
```
</details>

<details>
<summary><b>Can't clone private repos?</b></summary>

Run the GitHub auth setup:
```bash
~/codespaces/shared/scripts/setup-github-auth.sh
ssh -T git@github.com  # Test it
```
</details>

<details>
<summary><b>Port already in use?</b></summary>

The installer auto-finds free ports, but you can change them:
```bash
cd ~/codespaces/myproject
vim .env  # Change VS_CODE_PORT
docker-compose up -d
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