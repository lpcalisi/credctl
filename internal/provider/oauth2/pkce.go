package oauth2

import (
	"context"
	"time"

	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

type PKCEProvider struct {
	clientID      string
	clientSecret  string
	scopes        []string
	authEndpoint  string
	tokenEndpoint string
	redirectPort  int
	redirectURI   string
	tokens        *common.TokenCache
}

func init() {
	provider.Register(provider.ProviderOAuth2PKCE, func() provider.Provider {
		return &PKCEProvider{}
	})
}

func (p *PKCEProvider) Type() string {
	return provider.ProviderOAuth2PKCE
}

func (p *PKCEProvider) Schema() provider.Schema {
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
				Help:     "OAuth2 client secret (optional for public clients)",
				Hidden:   true,
			},
			{
				Name:     provider.MetadataScopes,
				Type:     provider.FieldTypeStringSlice,
				Required: false,
				Help:     "OAuth2 scopes to request",
			},
			{
				Name:     provider.MetadataAuthEndpoint,
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "Authorization endpoint URL",
			},
			{
				Name:     provider.MetadataTokenEndpoint,
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "Token endpoint URL",
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

func (p *PKCEProvider) Init(config map[string]any) error {
	p.clientID = provider.GetStringOrDefault(config, provider.MetadataClientID, "")
	p.clientSecret = provider.GetStringOrDefault(config, provider.MetadataClientSecret, "")
	p.scopes = common.GetScopes(config)
	p.authEndpoint = provider.GetStringOrDefault(config, provider.MetadataAuthEndpoint, "")
	p.tokenEndpoint = provider.GetStringOrDefault(config, provider.MetadataTokenEndpoint, "")
	p.redirectPort = provider.GetIntOrDefault(config, provider.MetadataRedirectPort, 8085)
	p.redirectURI = provider.GetStringOrDefault(config, provider.MetadataRedirectURI, "")

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
	}

	// No valid tokens and refresh failed - requires interactive authentication
	return nil, provider.ErrAuthenticationRequired
}

func (p *PKCEProvider) Metadata() map[string]any {
	metadata := map[string]any{
		provider.MetadataClientID:     p.clientID,
		provider.MetadataRedirectPort: p.redirectPort,
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
