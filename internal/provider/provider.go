package provider

import (
	"context"
	"errors"
)

var (
	// ErrAuthenticationRequired is returned by Get() when interactive authentication is required
	ErrAuthenticationRequired = errors.New("authentication required")
)

// Provider is the interface that all credential providers must implement
type Provider interface {
	// Type returns the provider type identifier (e.g., "command", "oidc")
	Type() string

	// Schema returns the configuration schema for this provider
	Schema() Schema

	// Init initializes the provider with the given configuration
	Init(config map[string]any) error

	// Get retrieves the credential from this provider
	Get(ctx context.Context) ([]byte, error)

	// Metadata returns provider metadata for serialization (export/import)
	Metadata() map[string]any
}

// LoginProvider is an optional interface for providers that support interactive login
type LoginProvider interface {
	Provider

	// Login performs interactive authentication for this provider
	Login(ctx context.Context) error
}

// TokenCacheProvider is an optional interface for providers that cache tokens in memory
type TokenCacheProvider interface {
	Provider

	// SetTokens sets the cached tokens for this provider
	SetTokens(accessToken, refreshToken string, expiresIn int)

	// GetTokens returns the current cached tokens (for sending to daemon after login)
	GetTokens() (accessToken, refreshToken string, expiresIn int)
}
