package oauth2

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"credctl/internal/formatter"
	"credctl/internal/provider"
	"credctl/internal/provider/oauth2/common"
)

// Flow types
const (
	FlowDevice            = "device"             // Device authorization flow
	FlowAuthCode          = "auth-code"          // Authorization code flow with PKCE
	FlowClientCredentials = "client-credentials" // Client credentials flow
)

// Provider implements a universal OAuth2/OIDC provider that supports multiple grant types
type Provider struct {
	// Discovery
	issuer string // If set, performs OIDC discovery and validates ID tokens

	// Core OAuth2 config
	clientID      string
	clientSecret  string
	scopes        []string
	tokenEndpoint string

	// Grant type detection (auto-detected from available endpoints)
	authEndpoint   string // If set → authorization_code flow
	deviceEndpoint string // If set → device flow
	redirectURI    string
	redirectPort   int

	// Flow options
	flow     string // Explicit flow selection (auto, device, auth-code, client-credentials)
	usePKCE  bool   // Use PKCE for authorization_code flow
	template string // Optional Go template for formatting output

	// Token cache
	tokens *common.TokenCache
}

func init() {
	// Register as "oauth2" - single universal provider
	provider.Register("oauth2", func() provider.Provider {
		return &Provider{}
	})
}

func (p *Provider) Type() string {
	return "oauth2"
}

func (p *Provider) Schema() provider.Schema {
	return provider.Schema{
		Fields: []provider.FieldDef{
			{
				Name:     provider.MetadataIssuer,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "OIDC issuer URL (enables auto-discovery and ID token validation)",
			},
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
				Required: false,
				Help:     "Token endpoint URL (auto-discovered if issuer is set)",
			},
			{
				Name:     provider.MetadataAuthEndpoint,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Authorization endpoint URL (enables authorization_code flow)",
			},
			{
				Name:     provider.MetadataDeviceEndpoint,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Device authorization endpoint URL (enables device flow)",
			},
			{
				Name:     provider.MetadataRedirectPort,
				Type:     provider.FieldTypeInt,
				Required: false,
				Default:  "8085",
				Help:     "Local port for OAuth callback server (for authorization_code flow)",
			},
			{
				Name:     provider.MetadataRedirectURI,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Custom redirect URI (overrides redirect_port)",
			},
			{
				Name:     "use_pkce",
				Type:     provider.FieldTypeBool,
				Required: false,
				Default:  "true",
				Help:     "Use PKCE extension for authorization_code flow (recommended for public clients)",
			},
			{
				Name:     "flow",
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "OAuth2 flow to use: device, auth-code, client-credentials",
			},
			{
				Name:     provider.MetadataTemplate,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Go template to format credentials (e.g., 'export TOKEN={{.access_token}}')",
			},
		},
	}
}

func (p *Provider) Init(config map[string]any) error {
	p.issuer = provider.GetStringOrDefault(config, provider.MetadataIssuer, "")
	p.clientID = provider.GetStringOrDefault(config, provider.MetadataClientID, "")
	p.clientSecret = provider.GetStringOrDefault(config, provider.MetadataClientSecret, "")
	p.tokenEndpoint = provider.GetStringOrDefault(config, provider.MetadataTokenEndpoint, "")
	p.authEndpoint = provider.GetStringOrDefault(config, provider.MetadataAuthEndpoint, "")
	p.deviceEndpoint = provider.GetStringOrDefault(config, provider.MetadataDeviceEndpoint, "")
	p.redirectPort = provider.GetIntOrDefault(config, provider.MetadataRedirectPort, 8085)
	p.redirectURI = provider.GetStringOrDefault(config, provider.MetadataRedirectURI, "")
	p.usePKCE = provider.GetBoolOrDefault(config, "use_pkce", true)
	p.flow = provider.GetStringOrDefault(config, "flow", "")
	p.template = provider.GetStringOrDefault(config, provider.MetadataTemplate, "")

	// Validate flow is provided
	if p.flow == "" {
		return fmt.Errorf("flow is required: must be one of: device, auth-code, client-credentials")
	}

	// Validate flow value
	switch p.flow {
	case FlowDevice, FlowAuthCode, FlowClientCredentials:
		// Valid flow
	default:
		return fmt.Errorf("invalid flow '%s': must be one of: device, auth-code, client-credentials", p.flow)
	}

	// Get scopes (default to "openid" if OIDC)
	if p.issuer != "" {
		p.scopes = common.GetScopesOrDefault(config)
	} else {
		p.scopes = common.GetScopes(config)
	}

	// Perform OIDC discovery if issuer is set
	if p.issuer != "" {
		doc, err := common.Discover(p.issuer)
		if err != nil {
			return fmt.Errorf("failed to discover OIDC endpoints: %w", err)
		}
		// Use discovered endpoints if not explicitly set
		if p.tokenEndpoint == "" {
			p.tokenEndpoint = doc.TokenEndpoint
		}

		// Only auto-configure endpoints based on explicit flow setting
		switch p.flow {
		case FlowAuthCode:
			// Auth code mode: only configure auth endpoint
			if p.authEndpoint == "" {
				p.authEndpoint = doc.AuthorizationEndpoint
			}
		case FlowDevice:
			// Device mode: only configure device endpoint
			if p.deviceEndpoint == "" && doc.DeviceEndpoint != "" {
				p.deviceEndpoint = doc.DeviceEndpoint
			}
		case FlowClientCredentials:
			// Client credentials: no interactive endpoints needed
			// Only token_endpoint is used
		}
	}

	// Validate we have at least one way to get tokens
	if p.tokenEndpoint == "" {
		return fmt.Errorf("token_endpoint is required (or provide issuer for auto-discovery)")
	}

	// Validate flow-specific requirements
	switch p.flow {
	case FlowDevice:
		if p.deviceEndpoint == "" {
			return fmt.Errorf("device flow requires device_endpoint (or issuer for auto-discovery)")
		}
	case FlowAuthCode:
		if p.authEndpoint == "" {
			return fmt.Errorf("auth-code flow requires auth_endpoint (or issuer for auto-discovery)")
		}
	case FlowClientCredentials:
		if p.clientSecret == "" {
			return fmt.Errorf("client-credentials flow requires client_secret")
		}
	}

	return nil
}

func (p *Provider) Get(ctx context.Context) ([]byte, error) {
	// Check if we have valid cached tokens
	if common.IsTokenValid(p.tokens) {
		return []byte(p.tokens.AccessToken), nil
	}

	// Try to refresh if we have a refresh token
	if p.tokens != nil && p.tokens.RefreshToken != "" {
		newTokens, err := common.RefreshAccessToken(p.tokenEndpoint, p.clientID, p.clientSecret, p.tokens.RefreshToken)
		if err == nil {
			p.tokens = newTokens
			return []byte(p.tokens.AccessToken), nil
		}
		// Refresh failed, continue to try other flows
	}

	// Handle flows based on explicit flow setting
	switch p.flow {
	case FlowClientCredentials:
		// Client credentials flow (non-interactive, machine-to-machine)
		tokens, err := common.GetClientCredentialsToken(p.tokenEndpoint, p.clientID, p.clientSecret, p.scopes)
		if err != nil {
			return nil, fmt.Errorf("client credentials grant failed: %w", err)
		}
		p.tokens = tokens
		return []byte(tokens.AccessToken), nil

	case FlowAuthCode:
		// Authorization Code Flow (PKCE) - automatic, opens browser
		if err := p.doAuthorizationCodeFlow(ctx); err != nil {
			return nil, fmt.Errorf("authorization code flow failed: %w", err)
		}
		return []byte(p.tokens.AccessToken), nil

	case FlowDevice:
		// Device Flow requires explicit login
		return nil, provider.ErrDeviceFlowRequiresLogin

	default:
		return nil, fmt.Errorf("unsupported flow: %s", p.flow)
	}
}

func (p *Provider) Login(ctx context.Context) error {
	// Login is for explicit interactive authentication
	// This is useful for:
	// - Device flow (requires user to see code and visit URL)
	// - Force re-authentication (invalidate current tokens)
	// - Pre-authenticate before using Get()

	var tokens *common.TokenCache
	var err error

	// Handle login based on explicit flow setting
	switch p.flow {
	case FlowDevice:
		tokens, err = common.AuthenticateDeviceFlow(ctx, p.deviceEndpoint, p.tokenEndpoint, p.clientID, p.clientSecret, p.scopes)

	case FlowAuthCode:
		if err := p.doAuthorizationCodeFlow(ctx); err != nil {
			return err
		}
		return nil

	case FlowClientCredentials:
		return fmt.Errorf("client-credentials flow does not support interactive login")

	default:
		return fmt.Errorf("unsupported flow: %s", p.flow)
	}

	if err != nil {
		return err
	}

	// Validate ID token if this is OIDC
	if p.issuer != "" && tokens.IDToken != "" {
		if err := p.validateIDToken(ctx, tokens.IDToken); err != nil {
			return fmt.Errorf("ID token validation failed: %w", err)
		}
	}

	p.tokens = tokens
	return nil
}

// doAuthorizationCodeFlow performs the authorization code flow with optional PKCE
func (p *Provider) doAuthorizationCodeFlow(ctx context.Context) error {
	code, codeVerifier, redirectURI, err := common.AuthenticateAuthCodeFlow(ctx, common.AuthCodeFlowParams{
		AuthEndpoint: p.authEndpoint,
		ClientID:     p.clientID,
		Scopes:       p.scopes,
		RedirectURI:  p.redirectURI,
		RedirectPort: p.redirectPort,
		UsePKCE:      p.usePKCE,
	})
	if err != nil {
		return err
	}

	tokens, err := common.ExchangeCodeForTokens(p.tokenEndpoint, p.clientID, p.clientSecret, code, redirectURI, codeVerifier)
	if err != nil {
		return err
	}

	// Validate ID token if this is OIDC
	if p.issuer != "" && tokens.IDToken != "" {
		if err := p.validateIDToken(ctx, tokens.IDToken); err != nil {
			return fmt.Errorf("ID token validation failed: %w", err)
		}
	}

	p.tokens = tokens
	return nil
}

func (p *Provider) validateIDToken(ctx context.Context, rawIDToken string) error {
	oidcProvider, err := common.NewOIDCProvider(ctx, p.issuer)
	if err != nil {
		return err
	}

	verifier := common.NewIDTokenVerifier(oidcProvider, p.clientID)
	_, err = common.VerifyIDToken(ctx, verifier, rawIDToken)
	return err
}

func (p *Provider) Metadata() map[string]any {
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
	if p.authEndpoint != "" {
		metadata[provider.MetadataAuthEndpoint] = p.authEndpoint
	}
	if p.deviceEndpoint != "" {
		metadata[provider.MetadataDeviceEndpoint] = p.deviceEndpoint
	}
	if p.redirectURI != "" {
		metadata[provider.MetadataRedirectURI] = p.redirectURI
	}
	if p.redirectPort != 0 {
		metadata[provider.MetadataRedirectPort] = p.redirectPort
	}
	if p.usePKCE {
		metadata["use_pkce"] = true
	}
	if p.flow != "" {
		metadata["flow"] = p.flow
	}
	if p.template != "" {
		metadata[provider.MetadataTemplate] = p.template
	}

	return metadata
}

func (p *Provider) SetTokens(accessToken, refreshToken string, expiresIn int) {
	p.tokens = &common.TokenCache{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(common.NormalizeExpiresIn(expiresIn)) * time.Second),
	}
}

func (p *Provider) GetTokens() (accessToken, refreshToken string, expiresIn int) {
	if p.tokens == nil {
		return "", "", 0
	}
	remaining := int(time.Until(p.tokens.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return p.tokens.AccessToken, p.tokens.RefreshToken, remaining
}

// GetCredentials returns the credentials in a structured format
// This implements the CredentialsProvider interface
func (p *Provider) GetCredentials(ctx context.Context) (*formatter.Credentials, error) {
	// Check if we have valid cached tokens
	if !common.IsTokenValid(p.tokens) {
		// Try to get fresh tokens using Get() logic
		_, err := p.Get(ctx)
		if err != nil {
			return nil, err
		}
	}

	if p.tokens == nil {
		return nil, fmt.Errorf("no tokens available")
	}

	// Build structured credentials with all available token fields
	fields := make(map[string]string)

	if p.tokens.AccessToken != "" {
		fields["access_token"] = p.tokens.AccessToken
	}
	if p.tokens.RefreshToken != "" {
		fields["refresh_token"] = p.tokens.RefreshToken
	}
	if p.tokens.IDToken != "" {
		fields["id_token"] = p.tokens.IDToken
	}
	if p.tokens.TokenType != "" {
		fields["token_type"] = p.tokens.TokenType
	}

	// Add expires_at as ISO8601 timestamp
	fields["expires_at"] = p.tokens.ExpiresAt.Format(time.RFC3339)

	// Add expires_in as seconds remaining
	remaining := int(time.Until(p.tokens.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	fields["expires_in"] = strconv.Itoa(remaining)

	return formatter.NewCredentials(fields), nil
}
