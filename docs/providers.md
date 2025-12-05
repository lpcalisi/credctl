# Providers

A **provider** is a configured source that credctl uses to retrieve credentials. Each provider defines how to obtain a specific credential (token, secret, etc.) from a particular source.

## How Providers Work

1. **Configure** a provider with a unique name and type
2. **Request** credentials using `credctl get <name>`
3. The provider **executes** on your local machine
4. The **credential** is returned (and cached in memory)

## Available Provider Types

credctl supports the following provider types:

### 🔧 [Command Provider](command.md)
Execute local commands or scripts to get credentials.

**Use cases:** CLI tools (gh, aws, az), custom scripts, password managers

### 🔐 [OAuth2 Provider](oauth2.md)
Full OAuth2/OIDC client with multiple authentication flows.

**Use cases:** Google, Azure AD, Keycloak, Auth0, GitHub, any OAuth2-compatible service

### 🌐 [OAuth2 Proxy Provider](oauth2-proxy.md)
Simplified authentication via transparent OAuth2 proxies.

**Use cases:** Corporate proxies, simplified OAuth2 flows without client configuration

## Storage & Caching

- **Provider configurations**: Stored in `~/.credctl/providers/<name>.json`
- **Credentials**: Cached in memory only (not persisted to disk)
- **Execution**: Providers always run on your local machine (even when accessed remotely)