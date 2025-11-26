package common

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func RefreshAccessToken(tokenEndpoint, clientID, clientSecret, refreshToken string) (*TokenCache, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", clientID)
	data.Set("refresh_token", refreshToken)
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}

	body, err := postFormJSON(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token refresh failed: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	cache := &TokenCache{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(NormalizeExpiresIn(tokenResp.ExpiresIn)) * time.Second),
	}

	if cache.RefreshToken == "" {
		cache.RefreshToken = refreshToken
	}

	return cache, nil
}

func ExchangeCodeForTokens(tokenEndpoint, clientID, clientSecret, code, redirectURI, codeVerifier string) (*TokenCache, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	body, err := postFormJSON(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token exchange failed: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &TokenCache{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(NormalizeExpiresIn(tokenResp.ExpiresIn)) * time.Second),
	}, nil
}

func GetClientCredentialsToken(tokenEndpoint, clientID, clientSecret string, scopes []string) (*TokenCache, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	body, err := postFormJSON(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("failed to get client credentials token: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("client credentials grant failed: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &TokenCache{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresAt:   time.Now().Add(time.Duration(NormalizeExpiresIn(tokenResp.ExpiresIn)) * time.Second),
	}, nil
}
