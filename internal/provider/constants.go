package provider

// Metadata field keys - constants to avoid hardcoding strings
const (
	// Common fields
	MetadataCommand      = "command"
	MetadataFormat       = "format"
	MetadataEnvVar       = "env_var"
	MetadataFilePath     = "file_path"
	MetadataFileMode     = "file_mode"
	MetadataLoginCommand = "login_command"
)

// Format values
const (
	FormatRaw  = "raw"
	FormatEnv  = "env"
	FormatFile = "file"
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

// OIDC provider types
const (
	ProviderOIDCPKCE   = "oidc-pkce"
	ProviderOIDCDevice = "oidc-device"
	ProviderOIDCClient = "oidc-client"
)

// OAuth2 provider types
const (
	ProviderOAuth2PKCE   = "oauth2-pkce"
	ProviderOAuth2Device = "oauth2-device"
	ProviderOAuth2Client = "oauth2-client"
)
