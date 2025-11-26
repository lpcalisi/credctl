package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"credctl/internal/provider"
)

// CommandProvider executes shell commands to retrieve credentials
type CommandProvider struct {
	command      string
	format       string
	envVar       string
	filePath     string
	fileMode     string
	loginCommand string
}

func init() {
	provider.Register("command", func() provider.Provider {
		return &CommandProvider{}
	})
}

func (p *CommandProvider) Type() string {
	return "command"
}

func (p *CommandProvider) Schema() provider.Schema {
	return provider.Schema{
		Fields: []provider.FieldDef{
			{
				Name:     provider.MetadataCommand,
				Type:     provider.FieldTypeString,
				Required: true,
				Help:     "Command to execute to retrieve credentials",
			},
			{
				Name:        provider.MetadataFormat,
				Type:        provider.FieldTypeString,
				Required:    false,
				Default:     provider.FormatRaw,
				Help:        "Output format",
				ValidValues: []string{provider.FormatRaw, provider.FormatEnv, provider.FormatFile},
			},
			{
				Name:     provider.MetadataEnvVar,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Environment variable name (for env format)",
			},
			{
				Name:     provider.MetadataFilePath,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "File path to write credentials to (for file format)",
			},
			{
				Name:     provider.MetadataFileMode,
				Type:     provider.FieldTypeString,
				Required: false,
				Default:  "0600",
				Help:     "File permissions in octal format (for file format)",
			},
			{
				Name:     provider.MetadataLoginCommand,
				Type:     provider.FieldTypeString,
				Required: false,
				Help:     "Login command to execute for interactive authentication",
			},
		},
	}
}

// Init initializes the provider with the given configuration
func (p *CommandProvider) Init(config map[string]any) error {
	p.command = config[provider.MetadataCommand].(string)
	p.format = provider.GetStringOrDefault(config, provider.MetadataFormat, provider.FormatRaw)
	p.envVar = provider.GetStringOrDefault(config, provider.MetadataEnvVar, "")
	p.filePath = provider.GetStringOrDefault(config, provider.MetadataFilePath, "")
	p.fileMode = provider.GetStringOrDefault(config, provider.MetadataFileMode, "0600")
	p.loginCommand = provider.GetStringOrDefault(config, provider.MetadataLoginCommand, "")
	return nil
}

// Get retrieves the credential by executing the configured command
func (p *CommandProvider) Get(ctx context.Context) ([]byte, error) {
	// Execute command with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "/bin/sh", "-c", p.command)

	// Capture stdout
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Trim trailing newlines
	result := strings.TrimRight(string(stdout), "\r\n")
	return []byte(result), nil
}

func (p *CommandProvider) Metadata() map[string]any {
	metadata := map[string]any{
		provider.MetadataCommand: p.command,
		provider.MetadataFormat:  p.format,
	}

	if p.envVar != "" {
		metadata[provider.MetadataEnvVar] = p.envVar
	}
	if p.filePath != "" {
		metadata[provider.MetadataFilePath] = p.filePath
	}
	if p.fileMode != "" && p.fileMode != "0600" {
		metadata[provider.MetadataFileMode] = p.fileMode
	}
	if p.loginCommand != "" {
		metadata[provider.MetadataLoginCommand] = p.loginCommand
	}

	return metadata
}

// Export helper functions that were previously in the Provider struct

func (p *CommandProvider) GetFormat() string {
	return p.format
}

func (p *CommandProvider) GetEnvVar() string {
	return p.envVar
}

func (p *CommandProvider) GetFilePath() string {
	return p.filePath
}

func (p *CommandProvider) GetFileMode() string {
	return p.fileMode
}

func (p *CommandProvider) GetLoginCommand() string {
	return p.loginCommand
}

// Login performs interactive authentication by executing the login command
// This implements the LoginProvider interface
func (p *CommandProvider) Login(ctx context.Context) error {
	if p.loginCommand == "" {
		return fmt.Errorf("no login command configured")
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", p.loginCommand)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("login command failed: %w", err)
	}

	return nil
}
