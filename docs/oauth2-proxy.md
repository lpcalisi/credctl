# OAuth2 Proxy Provider

The `oauth2-proxy` provider authenticates through corporate OAuth2 proxies that handle the OAuth2/OIDC flow transparently.

Unlike standard OAuth2, you don't need to configure `client_id` or `client_secret`. The proxy handles all OAuth2 complexity internally and returns tokens via a callback URL.

## Configuration

```bash
credctl add oauth2-proxy myproxy \
  --auth_url="https://auth.company.com/authenticate?callback_url=http://localhost:8085/callback"

credctl get myproxy  # Opens browser, authenticates, returns token
```

See [OAuth2 Provider](oauth2.md) for standard OAuth2 integration or [Providers Overview](providers.md) for all available provider types.
