package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"credctl/internal/provider"

	"github.com/coreos/go-oidc/v3/oidc"
)

// IsTokenValid checks if a token is still valid (with 30 second buffer)
func IsTokenValid(tokens *TokenCache) bool {
	if tokens == nil || tokens.AccessToken == "" {
		return false
	}
	return time.Now().Add(30 * time.Second).Before(tokens.ExpiresAt)
}

// GetScopes extracts scopes from config
func GetScopes(config map[string]any) []string {
	return provider.GetStringSliceOrDefault(config, provider.MetadataScopes, nil)
}

// GetScopesOrDefault returns scopes from config or defaults to ["openid"] for OIDC
func GetScopesOrDefault(config map[string]any) []string {
	scopes := GetScopes(config)
	if len(scopes) == 0 {
		return []string{"openid"}
	}
	return scopes
}

// DiscoveryDocument represents an OIDC discovery document
type DiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	DeviceEndpoint        string `json:"device_authorization_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

// Discover fetches the OIDC discovery document from an issuer
func Discover(issuer string) (*DiscoveryDocument, error) {
	wellKnownURL := fmt.Sprintf("%s/.well-known/openid-configuration", strings.TrimSuffix(issuer, "/"))

	resp, err := http.Get(wellKnownURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned status %d", resp.StatusCode)
	}

	var doc DiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to parse discovery document: %w", err)
	}

	return &doc, nil
}

// NewOIDCProvider creates an OIDC provider for the given issuer
func NewOIDCProvider(ctx context.Context, issuer string) (*oidc.Provider, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	return provider, nil
}

// NewIDTokenVerifier creates an ID token verifier with the given configuration
func NewIDTokenVerifier(provider *oidc.Provider, clientID string) *oidc.IDTokenVerifier {
	return provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})
}

// VerifyIDToken verifies an ID token and returns the verified token
func VerifyIDToken(ctx context.Context, verifier *oidc.IDTokenVerifier, rawIDToken string) (*oidc.IDToken, error) {
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}
	return idToken, nil
}

// ExtractClaims extracts claims from an ID token into the provided struct
func ExtractClaims(idToken *oidc.IDToken, claims interface{}) error {
	if err := idToken.Claims(claims); err != nil {
		return fmt.Errorf("failed to extract claims: %w", err)
	}
	return nil
}

// StandardClaims represents common OIDC claims
type StandardClaims struct {
	Subject       string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	Picture       string   `json:"picture"`
	Groups        []string `json:"groups"`
}
