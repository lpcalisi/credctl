<div align="center">
  <img src="logo.png" alt="credctl logo" width="200"/>
  
  A credential helper daemon that runs on your local machine and provides secure credential access to remote environments via SSH socket forwarding.
</div>

---

## Why credctl?

- 🔒 **Credentials stay on your machine** - Never copy secrets to remote servers
- 🚀 **Works from anywhere** - SSH into servers, Docker containers, dev environments
- 🔑 **Multiple providers** - Command execution, OAuth2/OIDC flows
- 🎯 **Simple** - One daemon, forward the socket, done

## Installation

```bash
make install
```

Or manually:
```bash
go build -o credctl
sudo mv credctl /usr/local/bin/
```

## Quick Start

**1. Start daemon on your local machine:**
```bash
eval $(credctl daemon start)
```

**2. Add a credential provider:**
```bash
# GitHub CLI example
credctl add command gh --command "gh auth token"

# Google OAuth2 example
credctl add oauth2 google \
  --flow=auth-code \
  --client_id=YOUR_CLIENT_ID \
  --issuer=https://accounts.google.com
```

**3. Get credentials:**
```bash
credctl get gh
```

That's it on your local machine! ✅

## Remote Access

Now the magic: use credentials from **anywhere** without copying them.

### SSH Socket Forwarding

**On local machine, build for remote:**
```bash
make build-linux-amd64
scp credctl_linux_amd64 user@server:/tmp/credctl
```

**SSH with socket forward:**
```bash
ssh -R /tmp/credctl.sock:$HOME/.credctl/agent-readonly.sock user@server
```

**On remote server:**
```bash
export CREDCTL_SOCK=/tmp/credctl.sock
/tmp/credctl get gh    # Gets credential from YOUR machine!
```

### Docker Volume Mount (Linux only)

```bash
docker run -v $HOME/.credctl:/root/.credctl:ro \
  -e CREDCTL_SOCK=/root/.credctl/agent-readonly.sock \
  myimage

# Inside container
credctl get gh    # Works!
```

## Providers

### Command Provider
Execute any command to get credentials:
```bash
credctl add command gh --command "gh auth token"
credctl add command aws --command "aws sts get-session-token --output text --query 'Credentials.SessionToken'"
```

### OAuth2 Provider
Full OAuth2/OIDC support with multiple flows:
```bash
# Authorization Code (browser-based)
credctl add oauth2 google --flow=auth-code --client_id=... --issuer=...

# Device Flow (TV/CLI)
credctl add oauth2 google-tv --flow=device --client_id=... --client_secret=... --issuer=...

# Client Credentials (service-to-service)
credctl add oauth2 api --flow=client-credentials --client_id=... --client_secret=... --token_endpoint=...
```

📖 See [docs/oauth2.md](docs/oauth2.md) for detailed OAuth2 documentation.

## Common Commands

```bash
credctl get <name>              # Get credential
credctl login <name>            # Interactive login (OAuth2/OIDC)
credctl export > backup.json    # Export all providers
credctl import backup.json      # Import providers
credctl delete <name>           # Delete provider
tail -f $CREDCTL_LOGS           # View logs
```

## Use Cases

- 🐳 **Docker containers** - Mount socket, no secrets in images
- 🔧 **Remote development** - SSH to server, access local credentials
- 🌐 **CI/CD** - Forward socket to build agents
- 📱 **Multiple environments** - One daemon, many remotes
- 🔄 **Credential rotation** - Update once, works everywhere

---

**License:** MIT