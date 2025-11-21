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

