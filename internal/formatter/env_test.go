package formatter

import (
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func TestEnv(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		envVar  string
		want    string
		wantErr bool
		errMsg  string
	}{
		// JSON parsing tests
		{
			name:    "JSON object with string values",
			output:  `{"AWS_ACCESS_KEY_ID": "AKIAIOSFODNN7EXAMPLE", "AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI"}`,
			envVar:  "",
			wantErr: false,
		},
		{
			name:    "JSON object with null value",
			output:  `{"KEY": null}`,
			envVar:  "",
			want:    "export KEY=''",
			wantErr: false,
		},
		{
			name:    "JSON object with numeric value",
			output:  `{"PORT": 8080}`,
			envVar:  "",
			want:    "export PORT=8080",
			wantErr: false,
		},
		{
			name:    "JSON object with nested object (marshaled back)",
			output:  `{"CONFIG": {"nested": "value"}}`,
			envVar:  "",
			wantErr: false,
		},

		// Already has export statements
		{
			name:    "output already has export statements",
			output:  "export AWS_TOKEN=abc123\nexport OTHER_VAR=value",
			envVar:  "",
			want:    "export AWS_TOKEN=abc123\nexport OTHER_VAR=value",
			wantErr: false,
		},
		{
			name:    "single export statement",
			output:  "export MY_VAR=myvalue",
			envVar:  "",
			want:    "export MY_VAR=myvalue",
			wantErr: false,
		},

		// KEY=VALUE format (without export)
		{
			name:    "KEY=VALUE without export",
			output:  "AWS_TOKEN=abc123",
			envVar:  "",
			want:    "export AWS_TOKEN=abc123",
			wantErr: false,
		},
		{
			name:    "multiple KEY=VALUE lines",
			output:  "KEY1=value1\nKEY2=value2",
			envVar:  "",
			want:    "export KEY1=value1\nexport KEY2=value2",
			wantErr: false,
		},
		{
			name:    "KEY=VALUE with underscore in name",
			output:  "MY_SECRET_KEY=supersecret",
			envVar:  "",
			want:    "export MY_SECRET_KEY=supersecret",
			wantErr: false,
		},

		// Raw token tests
		{
			name:    "raw token with envVar specified",
			output:  "my-secret-token-12345",
			envVar:  "AUTH_TOKEN",
			want:    "export AUTH_TOKEN=my-secret-token-12345",
			wantErr: false,
		},
		{
			name:    "raw token without envVar should error",
			output:  "my-secret-token-12345",
			envVar:  "",
			wantErr: true,
			errMsg:  "output appears to be a raw token",
		},
		{
			name:    "raw token with special characters needs quoting",
			output:  "token with spaces",
			envVar:  "MY_TOKEN",
			want:    "export MY_TOKEN='token with spaces'",
			wantErr: false,
		},
		{
			name:    "raw token with single quotes",
			output:  "it's a token",
			envVar:  "MY_TOKEN",
			want:    `export MY_TOKEN="it's a token"`,
			wantErr: false,
		},

		// Invalid envVar name
		{
			name:    "invalid envVar name (lowercase)",
			output:  "token",
			envVar:  "mytoken",
			wantErr: true,
			errMsg:  "invalid env var name",
		},
		{
			name:    "invalid envVar name (starts with number)",
			output:  "token",
			envVar:  "1TOKEN",
			wantErr: true,
			errMsg:  "invalid env var name",
		},
		{
			name:    "valid envVar name with underscore prefix",
			output:  "token",
			envVar:  "_MY_VAR",
			want:    "export _MY_VAR=token",
			wantErr: false,
		},

		// Edge cases
		{
			name:    "empty output becomes raw token",
			output:  "",
			envVar:  "EMPTY",
			want:    "export EMPTY=''",
			wantErr: false,
		},
		{
			name:    "whitespace trimmed",
			output:  "  export MY_VAR=value  \n",
			envVar:  "",
			want:    "export MY_VAR=value",
			wantErr: false,
		},
		{
			name:    "token with newline character",
			output:  "multi\nline",
			envVar:  "TOKEN",
			want:    "export TOKEN=$'multi\\nline'",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Env(tt.output, tt.envVar)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Env() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Env() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Env() unexpected error: %v", err)
				return
			}

			// For JSON tests, we can't predict exact output order, so just check it contains exports
			if strings.HasPrefix(tt.output, "{") && tt.want == "" {
				if !strings.Contains(got, "export ") {
					t.Errorf("Env() JSON output should contain 'export', got: %q", got)
				}
				return
			}

			if tt.want != "" && got != tt.want {
				t.Errorf("Env() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJsonToExports(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		want map[string]bool // Check that these exports exist
	}{
		{
			name: "string value",
			data: map[string]interface{}{"KEY": "value"},
			want: map[string]bool{"export KEY=value": true},
		},
		{
			name: "nil value",
			data: map[string]interface{}{"KEY": nil},
			want: map[string]bool{"export KEY=": true},
		},
		{
			name: "numeric value",
			data: map[string]interface{}{"PORT": 8080},
			want: map[string]bool{"export PORT=8080": true},
		},
		{
			name: "boolean value",
			data: map[string]interface{}{"ENABLED": true},
			want: map[string]bool{"export ENABLED=true": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jsonToExports(tt.data)
			for expected := range tt.want {
				if !strings.Contains(got, expected) {
					t.Errorf("jsonToExports() = %q, want to contain %q", got, expected)
				}
			}
		})
	}
}

func TestAddExportPrefix(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "single KEY=VALUE",
			output: "MY_VAR=value",
			want:   "export MY_VAR=value",
		},
		{
			name:   "multiple KEY=VALUE",
			output: "VAR1=a\nVAR2=b",
			want:   "export VAR1=a\nexport VAR2=b",
		},
		{
			name:   "mixed lines",
			output: "VAR1=a\nsome comment\nVAR2=b",
			want:   "export VAR1=a\nsome comment\nexport VAR2=b",
		},
		{
			name:   "already has export (not matched)",
			output: "export VAR=value",
			want:   "export VAR=value", // Doesn't double-add
		},
		{
			name:   "lowercase var not matched",
			output: "lowercase=value",
			want:   "lowercase=value",
		},
		{
			name:   "var starting with underscore",
			output: "_PRIVATE=secret",
			want:   "export _PRIVATE=secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addExportPrefix(tt.output)
			if got != tt.want {
				t.Errorf("addExportPrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "simple string no quoting needed",
			s:    "simple",
			want: "simple",
		},
		{
			name: "string with space needs quoting",
			s:    "hello world",
			want: "'hello world'",
		},
		{
			name: "string with single quote",
			s:    "it's",
			want: `"it's"`,
		},
		{
			name: "string with dollar sign",
			s:    "$HOME",
			want: "'$HOME'",
		},
		{
			name: "string with backtick",
			s:    "`command`",
			want: "'`command`'",
		},
		{
			name: "string with double quote",
			s:    `say "hello"`,
			want: `'say "hello"'`,
		},
		{
			name: "empty string",
			s:    "",
			want: "''",
		},
		{
			name: "string with newline",
			s:    "line1\nline2",
			want: "$'line1\\nline2'",
		},
		{
			name: "string with tab",
			s:    "col1\tcol2",
			want: "$'col1\\tcol2'",
		},
		{
			name: "alphanumeric with dash and underscore",
			s:    "my-token_123",
			want: "my-token_123", // No quoting needed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := syntax.Quote(tt.s, syntax.LangBash)
			if err != nil {
				t.Errorf("syntax.Quote(%q) error: %v", tt.s, err)
				return
			}
			if got != tt.want {
				t.Errorf("syntax.Quote(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}
