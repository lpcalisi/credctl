package oauth2proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

// Provider implements a provider for OAuth2 transparent proxies
// These proxies handle the OAuth flow internally and return tokens in the callback URL
type Provider struct {
	authURL      string // Full URL of the proxy (including callback_url parameter)
	tokenField   string // Which token to return: "token", "access_token", or "both"
	redirectPort int    // Local port for callback server

	tokens *common.TokenCache // Cached tokens
}

func init() {
	provider.Register("oauth2-proxy", func() provider.Provider {
		return &Provider{}
	})
}

func (p *Provider) Type() string {
	return "oauth2-proxy"
}

func (p *Provider) Schema() provider.Schema {
	return provider.Schema{
		Fields: []provider.FieldDef{
			{
				Name:     "auth_url",
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "Full URL of the proxy authentication endpoint (including callback_url parameter)",
			},
			{
				Name:     "token_field",
				Type:     provider.FieldTypeString,
				Required: false,
				Default:  "token",
				Help:     "Token to return: token, access_token, or both (as JSON)",
			},
			{
				Name:     provider.MetadataRedirectPort,
				Type:     provider.FieldTypeInt,
				Required: false,
				Default:  "8085",
				Help:     "Local port for OAuth callback server",
			},
		},
	}
}

func (p *Provider) Init(config map[string]any) error {
	p.authURL = provider.GetStringOrDefault(config, "auth_url", "")
	p.tokenField = provider.GetStringOrDefault(config, "token_field", "token")
	p.redirectPort = provider.GetIntOrDefault(config, provider.MetadataRedirectPort, 8085)

	// Validate required fields
	if p.authURL == "" {
		return fmt.Errorf("auth_url is required")
	}

	// Validate token_field
	switch p.tokenField {
	case "token", "access_token", "both":
		// Valid values
	default:
		return fmt.Errorf("invalid token_field '%s': must be one of: token, access_token, both", p.tokenField)
	}

	return nil
}

func (p *Provider) Get(ctx context.Context) ([]byte, error) {
	// Check if we have valid cached tokens
	if common.IsTokenValid(p.tokens) {
		return p.formatToken()
	}

	// No valid token, perform authentication flow
	if err := p.doProxyAuthFlow(ctx); err != nil {
		return nil, err
	}

	return p.formatToken()
}

func (p *Provider) Login(ctx context.Context) error {
	// Force re-authentication by clearing cache
	p.tokens = nil
	return p.doProxyAuthFlow(ctx)
}

func (p *Provider) Metadata() map[string]any {
	metadata := map[string]any{
		"auth_url":    p.authURL,
		"token_field": p.tokenField,
	}

	if p.redirectPort != 0 {
		metadata[provider.MetadataRedirectPort] = p.redirectPort
	}

	return metadata
}

// doProxyAuthFlow performs the proxy authentication flow
// It reuses the callback server infrastructure from oauth2/common
func (p *Provider) doProxyAuthFlow(ctx context.Context) error {
	// Open browser with the authentication URL (already includes callback_url)
	if err := common.OpenBrowser(p.authURL); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	// Start callback server and wait for the redirect
	result, err := common.StartCallbackServer(ctx, p.redirectPort, "/callback")
	if err != nil {
		return err
	}

	// Extract tokens from query parameters
	token := result.Params.Get("token")
	accessToken := result.Params.Get("access_token")

	// Check if we got at least one token
	if token == "" && accessToken == "" {
		return fmt.Errorf("no tokens received in callback (expected 'token' or 'access_token' parameters)")
	}

	// Cache the tokens
	// Since the proxy doesn't provide expires_in, we set a reasonable default (1 hour)
	p.tokens = &common.TokenCache{
		AccessToken: accessToken,
		IDToken:     token, // Store token in IDToken field
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	return nil
}

// formatToken returns the token according to the token_field configuration
func (p *Provider) formatToken() ([]byte, error) {
	if p.tokens == nil {
		return nil, fmt.Errorf("no tokens available")
	}

	switch p.tokenField {
	case "token":
		if p.tokens.IDToken == "" {
			return nil, fmt.Errorf("token not available")
		}
		return []byte(p.tokens.IDToken), nil

	case "access_token":
		if p.tokens.AccessToken == "" {
			return nil, fmt.Errorf("access_token not available")
		}
		return []byte(p.tokens.AccessToken), nil

	case "both":
		result := map[string]string{
			"token":        p.tokens.IDToken,
			"access_token": p.tokens.AccessToken,
		}
		data, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tokens: %w", err)
		}
		return data, nil

	default:
		return nil, fmt.Errorf("invalid token_field: %s", p.tokenField)
	}
}

// SetTokens sets the cached tokens (used by daemon for persistence)
func (p *Provider) SetTokens(accessToken, refreshToken string, expiresIn int) {
	p.tokens = &common.TokenCache{
		AccessToken: accessToken,
		IDToken:     refreshToken, // We use RefreshToken field to store the main token
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
}

// GetTokens returns the cached tokens (used by daemon for persistence)
func (p *Provider) GetTokens() (accessToken, refreshToken string, expiresIn int) {
	if p.tokens == nil {
		return "", "", 0
	}
	remaining := int(time.Until(p.tokens.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return p.tokens.AccessToken, p.tokens.IDToken, remaining
}
