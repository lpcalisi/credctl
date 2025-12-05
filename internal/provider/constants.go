package provider

// Metadata field keys - constants to avoid hardcoding strings
const (
	// Common fields
	MetadataCommand      = "command"
	MetadataLoginCommand = "login_command"
	MetadataTemplate     = "template" // Template Go para formatear output
)

// OIDC metadata field keys
const (
	MetadataIssuer         = "issuer"
	MetadataClientID       = "client_id"
	MetadataClientSecret   = "client_secret"
	MetadataScopes         = "scopes"
	MetadataAuthEndpoint   = "auth_endpoint"
	MetadataTokenEndpoint  = "token_endpoint"
	MetadataDeviceEndpoint = "device_endpoint"
	MetadataRedirectPort   = "redirect_port"
	MetadataRedirectURI    = "redirect_uri"
)
