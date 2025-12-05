# OAuth2 Provider

The `oauth2` provider is a unified OAuth2/OIDC client that supports multiple authentication flows. You must explicitly specify which flow to use with the required `--flow` parameter.

## Key Features

- **Explicit Flow Control**: `--flow` is **required** - you must explicitly specify which OAuth2 flow to use
- **Flexible Authentication**: Supports public clients (no secret) and confidential clients
- **OIDC Discovery**: Auto-discovers endpoints from issuer URL
- **Smart `Get()` Behavior**: Authorization Code flow runs automatically, Device flow requires explicit login
- **Token Management**: Automatic refresh token handling
- **PKCE Support**: Enabled by default for Authorization Code flow

## Supported Flows

### 1. Device Flow
**When**: For limited input devices (TV apps, CLI tools, IoT devices)  
**Behavior**: Requires explicit `credctl login` command

```bash
# Add provider with OIDC discovery
credctl add oauth2 google-device \
  --flow=device \
  --client_id=YOUR_CLIENT_ID \
  --client_secret=YOUR_CLIENT_SECRET \
  --issuer=https://accounts.google.com \
  --scopes=openid

# Must login explicitly (shows device code on screen)
credctl login google-device
# Output:
# To authenticate, visit: https://google.com/device
# And enter the code: ABC-DEF-GHI

# Then get token (uses cached token)
credctl get google-device
```

**Without OIDC Discovery (manual endpoints)**:
```bash
credctl add oauth2 google-device \
  --flow=device \
  --client_id=YOUR_CLIENT_ID \
  --client_secret=YOUR_CLIENT_SECRET \
  --device_endpoint=https://oauth2.googleapis.com/device/code \
  --token_endpoint=https://oauth2.googleapis.com/token \
  --scopes=openid
```

---

### 2. Authorization Code Flow (with PKCE)
**When**: For web applications and browser-based CLI tools  
**Behavior**: Runs automatically in `Get()`, opens browser

#### **Public Client (recommended - no client_secret needed)**:
```bash
# Add provider with OIDC discovery
credctl add oauth2 google-web \
  --flow=auth-code \
  --client_id=YOUR_CLIENT_ID \
  --issuer=https://accounts.google.com \
  --scopes=openid

# Get token (automatically opens browser)
credctl get google-web
# → Opens browser automatically
# → User authenticates
# → Token returned
```

#### **Confidential Client (with client_secret)**:
```bash
credctl add oauth2 google-confidential \
  --flow=auth-code \
  --client_id=YOUR_CLIENT_ID \
  --client_secret=YOUR_CLIENT_SECRET \
  --issuer=https://accounts.google.com \
  --scopes=openid
```

#### **Without OIDC Discovery**:
```bash
credctl add oauth2 github \
  --flow=auth-code \
  --client_id=YOUR_CLIENT_ID \
  --auth_endpoint=https://github.com/login/oauth/authorize \
  --token_endpoint=https://github.com/login/oauth/access_token \
  --scopes=repo,user
```

#### **Custom redirect port**:
```bash
credctl add oauth2 myapp \
  --flow=auth-code \
  --client_id=YOUR_CLIENT_ID \
  --issuer=https://accounts.google.com \
  --redirect_port=3000
```

#### **Disable PKCE (legacy servers)**:
```bash
credctl add oauth2 legacy \
  --flow=auth-code \
  --client_id=YOUR_CLIENT_ID \
  --client_secret=YOUR_CLIENT_SECRET \
  --auth_endpoint=https://legacy.example.com/oauth/authorize \
  --token_endpoint=https://legacy.example.com/oauth/token \
  --use_pkce=false
```

---

### 3. Client Credentials Flow
**When**: `client_secret` is provided with no interactive endpoints  
**Use case**: Machine-to-machine, service accounts  
**Behavior**: Runs automatically in `Get()`

```bash
# Add provider
credctl add oauth2 api-service \
  --client_id=YOUR_CLIENT_ID \
  --client_secret=YOUR_CLIENT_SECRET \
  --token_endpoint=https://api.example.com/oauth/token \
  --scopes=read,write \
  --flow=client-credentials

# Get token (automatic, non-interactive)
credctl get api-service
```

---

### 4. Refresh Token Flow
**When**: Valid refresh token exists  
**Behavior**: Automatic on token expiry

```bash
# Refresh happens automatically in Get() when:
# - Access token is expired
# - Refresh token is available

credctl get myapp
# → Detects expired token
# → Uses refresh token automatically
# → Returns new access token
```

## OIDC Discovery

When `issuer` is provided, the provider automatically discovers:
- `token_endpoint`
- `auth_endpoint`
- `device_endpoint` (if available)

Example:
```bash
credctl add oauth2 google \
  --client_id=YOUR_CLIENT_ID \
  --issuer=https://accounts.google.com
# Fetches: https://accounts.google.com/.well-known/openid-configuration
```

## Design

### Why Device Flow Requires Explicit Login

Device flow displays a code that the user must see and enter on another device. If this happened automatically in `Get()`:
- Code might get lost in logs
- Poor UX when triggered by background processes
- User might miss the authentication prompt

### Why Authorization Code Flow is Automatic

Authorization code flow opens a browser with immediate visual feedback:
- Browser window is obvious to the user
- No risk of missing the authentication prompt
- Perfect for credential helpers (git, docker, etc.)

## Common Patterns

### Flow Selection with OIDC Discovery

`--flow` is **required** and controls which flow to use. OIDC discovery will only configure the endpoints needed for your chosen flow:

```bash
# Authorization Code Flow (discovery configures auth_endpoint only)
credctl add oauth2 google \
  --client_id=YOUR_CLIENT_ID \
  --issuer=https://accounts.google.com \
  --flow=auth-code

# Device Flow (discovery configures device_endpoint only)
credctl add oauth2 google-device \
  --client_id=YOUR_CLIENT_ID \
  --client_secret=YOUR_CLIENT_SECRET \
  --issuer=https://accounts.google.com \
  --flow=device
```

### Pre-authenticate Before Use

```bash
# Add provider (flow is required)
credctl add oauth2 myapp \
  --client_id=... \
  --issuer=... \
  --flow=device

# Pre-authenticate (required for device flow)
credctl login myapp

# Use
credctl get myapp
```

## Token Storage

- Tokens are cached **in memory** by the daemon (not persisted to disk)
- Tokens persist across `credctl get` calls while daemon is running
- Refresh tokens are used automatically when access token expires
- Provider configuration is stored in `~/.credctl/providers/`