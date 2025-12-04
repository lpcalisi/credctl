package common

import (
	"time"

	"golang.org/x/oauth2"
)

const DefaultTokenExpiry = 365 * 24 * 60 * 60

type TokenCache struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	IDToken      string // For OIDC flows
}

func NormalizeExpiresIn(expiresIn int) int {
	if expiresIn <= 0 {
		return DefaultTokenExpiry
	}
	return expiresIn
}

// OAuth2TokenToCache converts an oauth2.Token to a TokenCache
func OAuth2TokenToCache(token *oauth2.Token) *TokenCache {
	cache := &TokenCache{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
	}

	// Extract ID token if present (for OIDC flows)
	if idToken, ok := token.Extra("id_token").(string); ok {
		cache.IDToken = idToken
	}

	// Normalize expiry if not set
	if cache.ExpiresAt.IsZero() {
		cache.ExpiresAt = time.Now().Add(time.Duration(DefaultTokenExpiry) * time.Second)
	}

	return cache
}

// ToOAuth2Token converts a TokenCache to an oauth2.Token
func (tc *TokenCache) ToOAuth2Token() *oauth2.Token {
	token := &oauth2.Token{
		AccessToken:  tc.AccessToken,
		RefreshToken: tc.RefreshToken,
		TokenType:    tc.TokenType,
		Expiry:       tc.ExpiresAt,
	}

	if tc.IDToken != "" {
		token = token.WithExtra(map[string]interface{}{
			"id_token": tc.IDToken,
		})
	}

	return token
}
