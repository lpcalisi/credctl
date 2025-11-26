package oidc

import (
	"context"
	"fmt"
	"time"

	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

// PKCEProvider implements OIDC Authorization Code with PKCE flow
type PKCEProvider struct {
	issuer        string
	clientID      string
	clientSecret  string
	scopes        []string
	authEndpoint  string
	tokenEndpoint string
	redirectPort  int
	redirectURI   string // Custom redirect URI (if set, overrides redirectPort)
	tokens        *common.TokenCache
}

func init() {
	provider.Register(provider.ProviderOIDCPKCE, func() provider.Provider {
		return &PKCEProvider{}
	})
}

func (p *PKCEProvider) Type() string {
	return provider.ProviderOIDCPKCE
}

func (p *PKCEProvider) Schema() provider.Schema {
	return provider.Schema{
		Fields: []provider.FieldDef{
			{
				Name:     provider.MetadataIssuer,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "OIDC issuer URL for auto-discovery of endpoints",
			},
			{
				Name:     provider.MetadataClientID,
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "OAuth client ID",
			},
			{
				Name:     provider.MetadataClientSecret,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "OAuth client secret (optional for public clients)",
				Hidden:   true,
			},
			{
				Name:     provider.MetadataScopes,
				Type:     provider.FieldTypeStringSlice,
				Required: false,
				Default:  "openid",
				Help:     "OAuth scopes to request",
			},
			{
				Name:     provider.MetadataAuthEndpoint,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Authorization endpoint URL (required if issuer not provided)",
			},
			{
				Name:     provider.MetadataTokenEndpoint,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Token endpoint URL (required if issuer not provided)",
			},
			{
				Name:     provider.MetadataRedirectPort,
				Type:     provider.FieldTypeInt,
				Required: false,
				Default:  "8085",
				Help:     "Local port for OAuth callback server (ignored if redirect_uri is set)",
			},
			{
				Name:     provider.MetadataRedirectURI,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Custom redirect URI (overrides redirect_port if set)",
			},
		},
	}
}

// Init initializes the provider with the given configuration
func (p *PKCEProvider) Init(config map[string]any) error {
	p.issuer = provider.GetStringOrDefault(config, provider.MetadataIssuer, "")
	p.clientID = provider.GetStringOrDefault(config, provider.MetadataClientID, "")
	p.clientSecret = provider.GetStringOrDefault(config, provider.MetadataClientSecret, "")
	p.scopes = GetScopesOrDefault(config)
	p.authEndpoint = provider.GetStringOrDefault(config, provider.MetadataAuthEndpoint, "")
	p.tokenEndpoint = provider.GetStringOrDefault(config, provider.MetadataTokenEndpoint, "")
	p.redirectPort = provider.GetIntOrDefault(config, provider.MetadataRedirectPort, 8085)
	p.redirectURI = provider.GetStringOrDefault(config, provider.MetadataRedirectURI, "")

	// If issuer is provided, try to discover endpoints
	if p.issuer != "" && (p.authEndpoint == "" || p.tokenEndpoint == "") {
		doc, err := Discover(p.issuer)
		if err != nil {
			return fmt.Errorf("failed to discover OIDC endpoints: %w", err)
		}
		if p.authEndpoint == "" {
			p.authEndpoint = doc.AuthorizationEndpoint
		}
		if p.tokenEndpoint == "" {
			p.tokenEndpoint = doc.TokenEndpoint
		}
	}

	// Validate required endpoints
	if p.authEndpoint == "" {
		return fmt.Errorf("auth_endpoint is required (or provide issuer for auto-discovery)")
	}
	if p.tokenEndpoint == "" {
		return fmt.Errorf("token_endpoint is required (or provide issuer for auto-discovery)")
	}

	return nil
}

func (p *PKCEProvider) Get(ctx context.Context) ([]byte, error) {
	if common.IsTokenValid(p.tokens) {
		return []byte(p.tokens.AccessToken), nil
	}

	if p.tokens != nil && p.tokens.RefreshToken != "" {
		newTokens, err := common.RefreshAccessToken(p.tokenEndpoint, p.clientID, p.clientSecret, p.tokens.RefreshToken)
		if err == nil {
			p.tokens = newTokens
			return []byte(p.tokens.AccessToken), nil
		}
		// Refresh failed, need to re-authenticate
	}

	// No valid tokens and refresh failed - requires interactive authentication
	return nil, provider.ErrAuthenticationRequired
}

func (p *PKCEProvider) Metadata() map[string]any {
	metadata := map[string]any{
		provider.MetadataClientID:     p.clientID,
		provider.MetadataRedirectPort: p.redirectPort,
	}

	if p.issuer != "" {
		metadata[provider.MetadataIssuer] = p.issuer
	}
	if p.clientSecret != "" {
		metadata[provider.MetadataClientSecret] = p.clientSecret
	}
	if len(p.scopes) > 0 {
		metadata[provider.MetadataScopes] = p.scopes
	}
	if p.authEndpoint != "" {
		metadata[provider.MetadataAuthEndpoint] = p.authEndpoint
	}
	if p.tokenEndpoint != "" {
		metadata[provider.MetadataTokenEndpoint] = p.tokenEndpoint
	}
	if p.redirectURI != "" {
		metadata[provider.MetadataRedirectURI] = p.redirectURI
	}

	return metadata
}

func (p *PKCEProvider) SetTokens(accessToken, refreshToken string, expiresIn int) {
	p.tokens = &common.TokenCache{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(common.NormalizeExpiresIn(expiresIn)) * time.Second),
	}
}

func (p *PKCEProvider) GetTokens() (accessToken, refreshToken string, expiresIn int) {
	if p.tokens == nil {
		return "", "", 0
	}
	remaining := int(time.Until(p.tokens.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return p.tokens.AccessToken, p.tokens.RefreshToken, remaining
}
