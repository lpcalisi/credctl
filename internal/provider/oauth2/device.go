package oauth2

import (
	"context"
	"time"

	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

type DeviceProvider struct {
	clientID       string
	clientSecret   string
	scopes         []string
	tokenEndpoint  string
	deviceEndpoint string
	tokens         *common.TokenCache
}

func init() {
	provider.Register(provider.ProviderOAuth2Device, func() provider.Provider {
		return &DeviceProvider{}
	})
}

func (p *DeviceProvider) Type() string {
	return provider.ProviderOAuth2Device
}

func (p *DeviceProvider) Schema() provider.Schema {
	return provider.Schema{
		Fields: []provider.FieldDef{
			{
				Name:     provider.MetadataClientID,
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "OAuth2 client ID",
			},
			{
				Name:     provider.MetadataClientSecret,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "OAuth2 client secret (required for confidential clients)",
				Hidden:   true,
			},
			{
				Name:     provider.MetadataScopes,
				Type:     provider.FieldTypeStringSlice,
				Required: false,
				Help:     "OAuth2 scopes to request",
			},
			{
				Name:     provider.MetadataTokenEndpoint,
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "Token endpoint URL",
			},
			{
				Name:     provider.MetadataDeviceEndpoint,
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "Device authorization endpoint URL",
			},
		},
	}
}

func (p *DeviceProvider) Init(config map[string]any) error {
	p.clientID = provider.GetStringOrDefault(config, provider.MetadataClientID, "")
	p.clientSecret = provider.GetStringOrDefault(config, provider.MetadataClientSecret, "")
	p.scopes = common.GetScopes(config)
	p.tokenEndpoint = provider.GetStringOrDefault(config, provider.MetadataTokenEndpoint, "")
	p.deviceEndpoint = provider.GetStringOrDefault(config, provider.MetadataDeviceEndpoint, "")

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
	}

	// No valid tokens and refresh failed - requires interactive authentication
	return nil, provider.ErrAuthenticationRequired
}

func (p *DeviceProvider) Login(ctx context.Context) error {
	tokens, err := p.authenticate(ctx)
	if err != nil {
		return err
	}
	p.tokens = tokens
	return nil
}

func (p *DeviceProvider) Metadata() map[string]any {
	metadata := map[string]any{
		provider.MetadataClientID:       p.clientID,
		provider.MetadataTokenEndpoint:  p.tokenEndpoint,
		provider.MetadataDeviceEndpoint: p.deviceEndpoint,
	}

	if p.clientSecret != "" {
		metadata[provider.MetadataClientSecret] = p.clientSecret
	}
	if len(p.scopes) > 0 {
		metadata[provider.MetadataScopes] = p.scopes
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
