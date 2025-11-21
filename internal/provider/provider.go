package provider

import "context"

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
