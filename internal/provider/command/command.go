package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"credctl/internal/formatter"
	"credctl/internal/provider"
)

// CommandProvider executes shell commands to retrieve credentials
type CommandProvider struct {
	command      string
	loginCommand string
	inputFormat  string
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
			{
				Name:        provider.MetadataInputFormat,
				Type:        provider.FieldTypeString,
				Required:    false,
				Default:     "raw",
				ValidValues: []string{"raw", "json", "env"},
				Help:        "Format of command output: raw (default), json, or env (KEY=VALUE)",
			},
		},
	}
}

// Init initializes the provider with the given configuration
func (p *CommandProvider) Init(config map[string]any) error {
	p.command = config[provider.MetadataCommand].(string)
	p.loginCommand = provider.GetStringOrDefault(config, provider.MetadataLoginCommand, "")
	p.inputFormat = provider.GetStringOrDefault(config, provider.MetadataInputFormat, "raw")
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

	if p.inputFormat != "" && p.inputFormat != "raw" {
		metadata[provider.MetadataInputFormat] = p.inputFormat
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

// GetCredentials returns the credentials in a structured format
// This implements the CredentialsProvider interface
func (p *CommandProvider) GetCredentials(ctx context.Context) (*formatter.Credentials, error) {
	// Get raw output from command
	output, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Parse based on input format
	var fields map[string]string
	switch p.inputFormat {
	case "json":
		fields, err = parseJSON(output)
	case "env":
		fields, err = parseEnv(output)
	case "raw", "":
		// For raw format, provide the entire output under "raw" field
		fields = map[string]string{
			"raw": string(output),
		}
	default:
		return nil, fmt.Errorf("unsupported input format: %s", p.inputFormat)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse %s output: %w", p.inputFormat, err)
	}

	return formatter.NewCredentials(fields), nil
}

// parseJSON parses JSON output into a flat map of string fields
func parseJSON(data []byte) (map[string]string, error) {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	fields := make(map[string]string)
	for key, value := range jsonData {
		// Convert all values to strings
		switch v := value.(type) {
		case string:
			fields[key] = v
		case float64:
			fields[key] = fmt.Sprintf("%g", v)
		case bool:
			fields[key] = fmt.Sprintf("%t", v)
		case nil:
			fields[key] = ""
		default:
			// For complex types (arrays, objects), marshal back to JSON string
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				fields[key] = fmt.Sprintf("%v", v)
			} else {
				fields[key] = string(jsonBytes)
			}
		}
	}

	return fields, nil
}

// parseEnv parses KEY=VALUE format into a map
func parseEnv(data []byte) (map[string]string, error) {
	fields := make(map[string]string)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		fields[key] = value
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no valid KEY=VALUE pairs found")
	}

	return fields, nil
}
