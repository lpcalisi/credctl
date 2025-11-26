package oauth2

import (
	"context"
	"time"

	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

type ClientProvider struct {
	clientID      string
	clientSecret  string
	scopes        []string
	tokenEndpoint string
	tokens        *common.TokenCache
}

func init() {
	provider.Register(provider.ProviderOAuth2Client, func() provider.Provider {
		return &ClientProvider{}
	})
}

func (p *ClientProvider) Type() string {
	return provider.ProviderOAuth2Client
}

func (p *ClientProvider) Schema() provider.Schema {
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
				Required: true,
				Help:     "OAuth2 client secret",
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
		},
	}
}

func (p *ClientProvider) Init(config map[string]any) error {
	p.clientID = provider.GetStringOrDefault(config, provider.MetadataClientID, "")
	p.clientSecret = provider.GetStringOrDefault(config, provider.MetadataClientSecret, "")
	p.scopes = common.GetScopes(config)
	p.tokenEndpoint = provider.GetStringOrDefault(config, provider.MetadataTokenEndpoint, "")

	return nil
}

func (p *ClientProvider) Get(ctx context.Context) ([]byte, error) {
	if common.IsTokenValid(p.tokens) {
		return []byte(p.tokens.AccessToken), nil
	}

	tokens, err := common.GetClientCredentialsToken(p.tokenEndpoint, p.clientID, p.clientSecret, p.scopes)
	if err != nil {
		return nil, err
	}

	p.tokens = tokens
	return []byte(tokens.AccessToken), nil
}

func (p *ClientProvider) Metadata() map[string]any {
	metadata := map[string]any{
		provider.MetadataClientID:      p.clientID,
		provider.MetadataClientSecret:  p.clientSecret,
		provider.MetadataTokenEndpoint: p.tokenEndpoint,
	}

	if len(p.scopes) > 0 {
		metadata[provider.MetadataScopes] = p.scopes
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
