package command

import (
	"context"
	"testing"

	"credctl/internal/provider"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    map[string]string
		shouldError bool
	}{
		{
			name:  "simple json object",
			input: `{"token": "abc123", "expires_in": 3600}`,
			expected: map[string]string{
				"token":      "abc123",
				"expires_in": "3600",
			},
		},
		{
			name:  "json with different types",
			input: `{"access_token": "token123", "expires_in": 3600, "active": true, "region": "us-east-1"}`,
			expected: map[string]string{
				"access_token": "token123",
				"expires_in":   "3600",
				"active":       "true",
				"region":       "us-east-1",
			},
		},
		{
			name:  "json with null value",
			input: `{"token": "abc", "refresh_token": null}`,
			expected: map[string]string{
				"token":         "abc",
				"refresh_token": "",
			},
		},
		{
			name:  "json with nested object",
			input: `{"token": "abc", "metadata": {"region": "us-west"}}`,
			expected: map[string]string{
				"token":    "abc",
				"metadata": `{"region":"us-west"}`,
			},
		},
		{
			name:  "json with array",
			input: `{"token": "abc", "scopes": ["read", "write"]}`,
			expected: map[string]string{
				"token":  "abc",
				"scopes": `["read","write"]`,
			},
		},
		{
			name:        "invalid json",
			input:       `not valid json`,
			shouldError: true,
		},
		{
			name:        "empty json",
			input:       `{}`,
			expected:    map[string]string{},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSON([]byte(tt.input))

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("field %s: expected %q, got %q", key, expectedValue, result[key])
				}
			}
		})
	}
}

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    map[string]string
		shouldError bool
	}{
		{
			name: "simple key-value pairs",
			input: `AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
AWS_SESSION_TOKEN=token123`,
			expected: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_SESSION_TOKEN":     "token123",
			},
		},
		{
			name: "with quotes",
			input: `TOKEN="abc123"
API_KEY='xyz789'`,
			expected: map[string]string{
				"TOKEN":   "abc123",
				"API_KEY": "xyz789",
			},
		},
		{
			name: "with comments and empty lines",
			input: `# This is a comment
TOKEN=abc123

# Another comment
API_KEY=xyz789`,
			expected: map[string]string{
				"TOKEN":   "abc123",
				"API_KEY": "xyz789",
			},
		},
		{
			name:  "with spaces around equals",
			input: `TOKEN = abc123`,
			expected: map[string]string{
				"TOKEN": "abc123",
			},
		},
		{
			name:  "value with equals sign",
			input: `TOKEN=abc=123=xyz`,
			expected: map[string]string{
				"TOKEN": "abc=123=xyz",
			},
		},
		{
			name:        "no valid pairs",
			input:       "just some text without equals",
			shouldError: true,
		},
		{
			name:        "empty input",
			input:       "",
			shouldError: true,
		},
		{
			name:        "only comments",
			input:       "# comment\n# another comment",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEnv([]byte(tt.input))

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("field %s: expected %q, got %q", key, expectedValue, result[key])
				}
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		inputFormat  string
		expectedKeys []string
		shouldError  bool
	}{
		{
			name:         "json format",
			command:      `echo '{"token":"abc123","expires_in":3600}'`,
			inputFormat:  "json",
			expectedKeys: []string{"token", "expires_in"},
		},
		{
			name:         "env format",
			command:      `printf "TOKEN=abc123\nAPI_KEY=xyz789"`,
			inputFormat:  "env",
			expectedKeys: []string{"TOKEN", "API_KEY"},
		},
		{
			name:         "raw format with plain text",
			command:      `echo "just plain text"`,
			inputFormat:  "raw",
			expectedKeys: []string{"raw"},
		},
		{
			name:        "invalid json",
			command:     `echo "not json"`,
			inputFormat: "json",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &CommandProvider{
				command:     tt.command,
				inputFormat: tt.inputFormat,
			}

			ctx := context.Background()
			creds, err := p.GetCredentials(ctx)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if creds == nil {
				t.Errorf("expected credentials but got nil")
				return
			}

			for _, key := range tt.expectedKeys {
				if !creds.Has(key) {
					t.Errorf("expected key %q to exist", key)
				}
			}
		})
	}
}

func TestGetCredentials_RawContent(t *testing.T) {
	p := &CommandProvider{
		command:     `printf "my-secret-token-12345"`,
		inputFormat: "raw",
	}

	ctx := context.Background()
	creds, err := p.GetCredentials(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if creds == nil {
		t.Errorf("expected credentials but got nil")
		return
	}

	if !creds.Has("raw") {
		t.Errorf("expected 'raw' field to exist")
		return
	}

	expected := "my-secret-token-12345"
	actual := creds.Get("raw")
	if actual != expected {
		t.Errorf("expected raw content %q, got %q", expected, actual)
	}
}

func TestCommandProvider_Init(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expected    CommandProvider
		shouldError bool
	}{
		{
			name: "basic configuration",
			config: map[string]any{
				provider.MetadataCommand: "echo test",
			},
			expected: CommandProvider{
				command:     "echo test",
				inputFormat: "raw",
			},
		},
		{
			name: "with input format",
			config: map[string]any{
				provider.MetadataCommand:     "cat token.json",
				provider.MetadataInputFormat: "json",
			},
			expected: CommandProvider{
				command:     "cat token.json",
				inputFormat: "json",
			},
		},
		{
			name: "with login command",
			config: map[string]any{
				provider.MetadataCommand:      "get-token",
				provider.MetadataLoginCommand: "login",
				provider.MetadataInputFormat:  "env",
			},
			expected: CommandProvider{
				command:      "get-token",
				loginCommand: "login",
				inputFormat:  "env",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &CommandProvider{}
			err := p.Init(tt.config)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if p.command != tt.expected.command {
				t.Errorf("command: expected %q, got %q", tt.expected.command, p.command)
			}

			if p.loginCommand != tt.expected.loginCommand {
				t.Errorf("loginCommand: expected %q, got %q", tt.expected.loginCommand, p.loginCommand)
			}

			if p.inputFormat != tt.expected.inputFormat {
				t.Errorf("inputFormat: expected %q, got %q", tt.expected.inputFormat, p.inputFormat)
			}
		})
	}
}
