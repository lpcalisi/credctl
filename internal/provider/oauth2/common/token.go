package common

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// RefreshAccessToken refreshes an OAuth2 access token using a refresh token
func RefreshAccessToken(tokenEndpoint, clientID, clientSecret, refreshToken string) (*TokenCache, error) {
	ctx := context.Background()

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenEndpoint,
		},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return OAuth2TokenToCache(newToken), nil
}

// ExchangeCodeForTokens exchanges an authorization code for tokens
func ExchangeCodeForTokens(tokenEndpoint, clientID, clientSecret, code, redirectURI, codeVerifier string) (*TokenCache, error) {
	ctx := context.Background()

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenEndpoint,
		},
		RedirectURL: redirectURI,
	}

	var opts []oauth2.AuthCodeOption
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	token, err := config.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return OAuth2TokenToCache(token), nil
}

// GetClientCredentialsToken obtains a token using the client credentials grant
func GetClientCredentialsToken(tokenEndpoint, clientID, clientSecret string, scopes []string) (*TokenCache, error) {
	ctx := context.Background()

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenEndpoint,
		Scopes:       scopes,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client credentials token: %w", err)
	}

	return OAuth2TokenToCache(token), nil
}
