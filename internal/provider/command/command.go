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
	}

	if p.loginCommand != "" {
		metadata[provider.MetadataLoginCommand] = p.loginCommand
	}

	return metadata
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
