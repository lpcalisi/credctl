package oidc

import (
	"context"
	"fmt"
	"time"

	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

// ClientProvider implements the Provider interface for OIDC Client Credentials Grant
type ClientProvider struct {
	issuer        string
	clientID      string
	clientSecret  string
	scopes        []string
	tokenEndpoint string
	tokens        *common.TokenCache
}

func init() {
	provider.Register(provider.ProviderOIDCClient, func() provider.Provider {
		return &ClientProvider{}
	})
}

// Type returns the provider type identifier
func (p *ClientProvider) Type() string {
	return provider.ProviderOIDCClient
}

// Schema returns the configuration schema for the Client Credentials provider
func (p *ClientProvider) Schema() provider.Schema {
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
				Required: true,
				Help:     "OAuth client secret",
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
				Name:     provider.MetadataTokenEndpoint,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Token endpoint URL (required if issuer not provided)",
			},
		},
	}
}

// Init initializes the provider with the given configuration
func (p *ClientProvider) Init(config map[string]any) error {
	p.issuer = provider.GetStringOrDefault(config, provider.MetadataIssuer, "")
	p.clientID = provider.GetStringOrDefault(config, provider.MetadataClientID, "")
	p.clientSecret = provider.GetStringOrDefault(config, provider.MetadataClientSecret, "")
	p.scopes = GetScopesOrDefault(config)
	p.tokenEndpoint = provider.GetStringOrDefault(config, provider.MetadataTokenEndpoint, "")

	// If issuer is provided, try to discover endpoints
	if p.issuer != "" && p.tokenEndpoint == "" {
		doc, err := Discover(p.issuer)
		if err != nil {
			return fmt.Errorf("failed to discover OIDC endpoints: %w", err)
		}
		p.tokenEndpoint = doc.TokenEndpoint
	}

	// Validate required endpoint
	if p.tokenEndpoint == "" {
		return fmt.Errorf("token_endpoint is required (or provide issuer for auto-discovery)")
	}

	return nil
}

// Get retrieves the credential (access token) from this provider
func (p *ClientProvider) Get(ctx context.Context) ([]byte, error) {
	if common.IsTokenValid(p.tokens) {
		return []byte(p.tokens.AccessToken), nil
	}

	tokens, err := common.GetClientCredentialsToken(p.tokenEndpoint, p.clientID, p.clientSecret, p.scopes)
	if err != nil {
		return nil, err
	}

	// Cache the token in memory
	p.tokens = tokens

	return []byte(tokens.AccessToken), nil
}

// Metadata returns provider metadata for serialization
func (p *ClientProvider) Metadata() map[string]any {
	metadata := map[string]any{
		provider.MetadataClientID:     p.clientID,
		provider.MetadataClientSecret: p.clientSecret,
	}

	if p.issuer != "" {
		metadata[provider.MetadataIssuer] = p.issuer
	}
	if len(p.scopes) > 0 {
		metadata[provider.MetadataScopes] = p.scopes
	}
	if p.tokenEndpoint != "" {
		metadata[provider.MetadataTokenEndpoint] = p.tokenEndpoint
	}

	return metadata
}

func (p *ClientProvider) SetTokens(accessToken, refreshToken string, expiresIn int) {
	p.tokens = &common.TokenCache{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(common.NormalizeExpiresIn(expiresIn)) * time.Second),
	}
}

// GetTokens returns the current cached tokens (implements TokenCacheProvider)
func (p *ClientProvider) GetTokens() (accessToken, refreshToken string, expiresIn int) {
	if p.tokens == nil {
		return "", "", 0
	}
	remaining := int(time.Until(p.tokens.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return p.tokens.AccessToken, p.tokens.RefreshToken, remaining
}
