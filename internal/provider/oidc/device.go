package oidc

import (
	"context"
	"fmt"
	"time"

	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

// DeviceProvider implements the Provider and LoginProvider interfaces for OIDC Device Authorization Grant
type DeviceProvider struct {
	issuer         string
	clientID       string
	clientSecret   string
	scopes         []string
	tokenEndpoint  string
	deviceEndpoint string
	tokens         *common.TokenCache
}

func init() {
	provider.Register(provider.ProviderOIDCDevice, func() provider.Provider {
		return &DeviceProvider{}
	})
}

// Type returns the provider type identifier
func (p *DeviceProvider) Type() string {
	return provider.ProviderOIDCDevice
}

// Schema returns the configuration schema for the Device provider
func (p *DeviceProvider) Schema() provider.Schema {
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
				Help:     "OAuth client secret (required for confidential clients)",
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
			{
				Name:     provider.MetadataDeviceEndpoint,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Device authorization endpoint URL (required if issuer not provided)",
			},
		},
	}
}

func (p *DeviceProvider) Init(config map[string]any) error {
	p.issuer = provider.GetStringOrDefault(config, provider.MetadataIssuer, "")
	p.clientID = provider.GetStringOrDefault(config, provider.MetadataClientID, "")
	p.clientSecret = provider.GetStringOrDefault(config, provider.MetadataClientSecret, "")
	p.scopes = GetScopesOrDefault(config)
	p.tokenEndpoint = provider.GetStringOrDefault(config, provider.MetadataTokenEndpoint, "")
	p.deviceEndpoint = provider.GetStringOrDefault(config, provider.MetadataDeviceEndpoint, "")

	// If issuer is provided, try to discover endpoints
	if p.issuer != "" && (p.tokenEndpoint == "" || p.deviceEndpoint == "") {
		doc, err := Discover(p.issuer)
		if err != nil {
			return fmt.Errorf("failed to discover OIDC endpoints: %w", err)
		}
		if p.tokenEndpoint == "" {
			p.tokenEndpoint = doc.TokenEndpoint
		}
		if p.deviceEndpoint == "" {
			p.deviceEndpoint = doc.DeviceEndpoint
		}
	}

	// Validate required endpoints
	if p.tokenEndpoint == "" {
		return fmt.Errorf("token_endpoint is required (or provide issuer for auto-discovery)")
	}
	if p.deviceEndpoint == "" {
		return fmt.Errorf("device_endpoint is required (or provide issuer for auto-discovery)")
	}

	return nil
}

func (p *DeviceProvider) Get(ctx context.Context) ([]byte, error) {
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

// Login performs interactive authentication for this provider
// This implements the LoginProvider interface
func (p *DeviceProvider) Login(ctx context.Context) error {
	tokens, err := p.authenticate(ctx)
	if err != nil {
		return err
	}
	p.tokens = tokens
	return nil
}

// Metadata returns provider metadata for serialization
func (p *DeviceProvider) Metadata() map[string]any {
	metadata := map[string]any{
		provider.MetadataClientID: p.clientID,
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
	if p.tokenEndpoint != "" {
		metadata[provider.MetadataTokenEndpoint] = p.tokenEndpoint
	}
	if p.deviceEndpoint != "" {
		metadata[provider.MetadataDeviceEndpoint] = p.deviceEndpoint
	}

	return metadata
}

func (p *DeviceProvider) SetTokens(accessToken, refreshToken string, expiresIn int) {
	p.tokens = &common.TokenCache{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(common.NormalizeExpiresIn(expiresIn)) * time.Second),
	}
}

// GetTokens returns the current cached tokens (implements TokenCacheProvider)
func (p *DeviceProvider) GetTokens() (accessToken, refreshToken string, expiresIn int) {
	if p.tokens == nil {
		return "", "", 0
	}
	remaining := int(time.Until(p.tokens.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return p.tokens.AccessToken, p.tokens.RefreshToken, remaining
}

func (p *DeviceProvider) authenticate(ctx context.Context) (*common.TokenCache, error) {
	tokenResp, err := common.AuthenticateDeviceFlow(ctx, p.deviceEndpoint, p.tokenEndpoint, p.clientID, p.clientSecret, p.scopes)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(common.NormalizeExpiresIn(tokenResp.ExpiresIn)) * time.Second)
	return &common.TokenCache{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    expiresAt,
	}, nil
}
