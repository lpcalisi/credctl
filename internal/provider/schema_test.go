package provider

import (
	"testing"
)

func TestGetStringOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]any
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "key exists with string value",
			config:       map[string]any{"foo": "bar"},
			key:          "foo",
			defaultValue: "default",
			want:         "bar",
		},
		{
			name:         "key does not exist",
			config:       map[string]any{"other": "value"},
			key:          "foo",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "key exists but not a string",
			config:       map[string]any{"foo": 123},
			key:          "foo",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "empty config",
			config:       map[string]any{},
			key:          "foo",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "nil value",
			config:       map[string]any{"foo": nil},
			key:          "foo",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "empty string value",
			config:       map[string]any{"foo": ""},
			key:          "foo",
			defaultValue: "default",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringOrDefault(tt.config, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetStringOrDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetBoolOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]any
		key          string
		defaultValue bool
		want         bool
	}{
		{
			name:         "key exists with true",
			config:       map[string]any{"enabled": true},
			key:          "enabled",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "key exists with false",
			config:       map[string]any{"enabled": false},
			key:          "enabled",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "key does not exist, default true",
			config:       map[string]any{},
			key:          "enabled",
			defaultValue: true,
			want:         true,
		},
		{
			name:         "key does not exist, default false",
			config:       map[string]any{},
			key:          "enabled",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "key exists but not a bool (string)",
			config:       map[string]any{"enabled": "true"},
			key:          "enabled",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "key exists but not a bool (int)",
			config:       map[string]any{"enabled": 1},
			key:          "enabled",
			defaultValue: false,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBoolOrDefault(tt.config, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetBoolOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]any
		key          string
		defaultValue int
		want         int
	}{
		{
			name:         "key exists with int value",
			config:       map[string]any{"count": 42},
			key:          "count",
			defaultValue: 0,
			want:         42,
		},
		{
			name:         "key exists with float64 (JSON unmarshal)",
			config:       map[string]any{"count": float64(42)},
			key:          "count",
			defaultValue: 0,
			want:         42,
		},
		{
			name:         "key does not exist",
			config:       map[string]any{},
			key:          "count",
			defaultValue: 10,
			want:         10,
		},
		{
			name:         "key exists but not numeric (string)",
			config:       map[string]any{"count": "42"},
			key:          "count",
			defaultValue: 10,
			want:         10,
		},
		{
			name:         "zero value",
			config:       map[string]any{"count": 0},
			key:          "count",
			defaultValue: 10,
			want:         0,
		},
		{
			name:         "negative value",
			config:       map[string]any{"count": -5},
			key:          "count",
			defaultValue: 0,
			want:         -5,
		},
		{
			name:         "float64 with decimal truncated",
			config:       map[string]any{"count": float64(42.9)},
			key:          "count",
			defaultValue: 0,
			want:         42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIntOrDefault(tt.config, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetIntOrDefault() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetStringSliceOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]any
		key          string
		defaultValue []string
		want         []string
	}{
		{
			name:         "key exists with []string",
			config:       map[string]any{"tags": []string{"a", "b", "c"}},
			key:          "tags",
			defaultValue: nil,
			want:         []string{"a", "b", "c"},
		},
		{
			name:         "key exists with []interface{} (JSON unmarshal)",
			config:       map[string]any{"tags": []interface{}{"a", "b", "c"}},
			key:          "tags",
			defaultValue: nil,
			want:         []string{"a", "b", "c"},
		},
		{
			name:         "key does not exist",
			config:       map[string]any{},
			key:          "tags",
			defaultValue: []string{"default"},
			want:         []string{"default"},
		},
		{
			name:         "empty slice",
			config:       map[string]any{"tags": []string{}},
			key:          "tags",
			defaultValue: []string{"default"},
			want:         []string{},
		},
		{
			name:         "[]interface{} with mixed types (non-strings ignored)",
			config:       map[string]any{"tags": []interface{}{"a", 123, "b"}},
			key:          "tags",
			defaultValue: nil,
			want:         []string{"a", "b"},
		},
		{
			name:         "key exists but wrong type (string)",
			config:       map[string]any{"tags": "not-a-slice"},
			key:          "tags",
			defaultValue: []string{"default"},
			want:         []string{"default"},
		},
		{
			name:         "nil default",
			config:       map[string]any{},
			key:          "tags",
			defaultValue: nil,
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringSliceOrDefault(tt.config, tt.key, tt.defaultValue)

			// Handle nil comparison
			if tt.want == nil && got == nil {
				return
			}
			if tt.want == nil || got == nil {
				t.Errorf("GetStringSliceOrDefault() = %v, want %v", got, tt.want)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetStringSliceOrDefault() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetStringSliceOrDefault()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		schema  Schema
		wantErr bool
		errMsg  string
	}{
		{
			name:   "all required fields present",
			config: map[string]any{"command": "echo hello", "format": "raw"},
			schema: Schema{
				Fields: []FieldDef{
					{Name: "command", Type: FieldTypeString, Required: true},
					{Name: "format", Type: FieldTypeString, Required: true},
				},
			},
			wantErr: false,
		},
		{
			name:   "missing required field",
			config: map[string]any{"format": "raw"},
			schema: Schema{
				Fields: []FieldDef{
					{Name: "command", Type: FieldTypeString, Required: true},
					{Name: "format", Type: FieldTypeString, Required: true},
				},
			},
			wantErr: true,
			errMsg:  "required field 'command' is missing",
		},
		{
			name:   "optional field missing is ok",
			config: map[string]any{"command": "echo hello"},
			schema: Schema{
				Fields: []FieldDef{
					{Name: "command", Type: FieldTypeString, Required: true},
					{Name: "format", Type: FieldTypeString, Required: false},
				},
			},
			wantErr: false,
		},
		{
			name:   "valid enum value",
			config: map[string]any{"format": "env"},
			schema: Schema{
				Fields: []FieldDef{
					{Name: "format", Type: FieldTypeString, ValidValues: []string{"raw", "env", "file"}},
				},
			},
			wantErr: false,
		},
		{
			name:   "invalid enum value",
			config: map[string]any{"format": "invalid"},
			schema: Schema{
				Fields: []FieldDef{
					{Name: "format", Type: FieldTypeString, ValidValues: []string{"raw", "env", "file"}},
				},
			},
			wantErr: true,
			errMsg:  "field 'format' must be one of:",
		},
		{
			name:   "enum field not present (allowed if not required)",
			config: map[string]any{},
			schema: Schema{
				Fields: []FieldDef{
					{Name: "format", Type: FieldTypeString, ValidValues: []string{"raw", "env"}},
				},
			},
			wantErr: false,
		},
		{
			name:   "enum value not a string",
			config: map[string]any{"format": 123},
			schema: Schema{
				Fields: []FieldDef{
					{Name: "format", Type: FieldTypeString, ValidValues: []string{"raw", "env"}},
				},
			},
			wantErr: true,
			errMsg:  "field 'format' must be a string",
		},
		{
			name:    "empty schema",
			config:  map[string]any{"anything": "goes"},
			schema:  Schema{Fields: []FieldDef{}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config, tt.schema)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateConfig() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

