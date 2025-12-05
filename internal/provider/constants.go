package provider

// Metadata field keys - constants to avoid hardcoding strings
const (
	// Common fields
	MetadataCommand      = "command"
	MetadataLoginCommand = "login_command"
	MetadataTemplate     = "template"     // Go template for output formatting
	MetadataInputFormat  = "input_format" // Format of command output (raw, json, env, yaml)
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
