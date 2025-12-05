package credentials

import (
	"testing"
)

func TestApplyTemplate(t *testing.T) {
	tests := []struct {
		name        string
		creds       *Credentials
		template    string
		expected    string
		shouldError bool
	}{
		{
			name: "simple export template",
			creds: New(map[string]string{
				"token": "abc123",
			}),
			template: "export TOKEN={{.token}}",
			expected: "export TOKEN=abc123",
		},
		{
			name: "json template",
			creds: New(map[string]string{
				"token": "xyz789",
			}),
			template: `{"jwt": "{{.token}}"}`,
			expected: `{"jwt": "xyz789"}`,
		},
		{
			name: "multiple fields",
			creds: New(map[string]string{
				"token":        "tok123",
				"access_token": "acc456",
			}),
			template: "export TOKEN={{.token}}\nexport ACCESS={{.access_token}}",
			expected: "export TOKEN=tok123\nexport ACCESS=acc456",
		},
		{
			name:        "nil credentials",
			creds:       nil,
			template:    "export TOKEN={{.token}}",
			shouldError: true,
		},
		{
			name: "empty template",
			creds: New(map[string]string{
				"token": "abc123",
			}),
			template:    "",
			shouldError: true,
		},
		{
			name: "invalid template syntax",
			creds: New(map[string]string{
				"token": "abc123",
			}),
			template:    "export TOKEN={{.token",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyTemplate(tt.creds, tt.template)

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

			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestCredentials(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		fields := map[string]string{"token": "abc"}
		creds := New(fields)

		if creds.Fields["token"] != "abc" {
			t.Errorf("expected token=abc, got %s", creds.Fields["token"])
		}
	})

	t.Run("Get", func(t *testing.T) {
		creds := New(map[string]string{"token": "abc"})

		if creds.Get("token") != "abc" {
			t.Errorf("expected abc, got %s", creds.Get("token"))
		}

		if creds.Get("nonexistent") != "" {
			t.Errorf("expected empty string for nonexistent key")
		}
	})

	t.Run("Set", func(t *testing.T) {
		creds := New(nil)
		creds.Set("token", "xyz")

		if creds.Get("token") != "xyz" {
			t.Errorf("expected xyz, got %s", creds.Get("token"))
		}
	})

	t.Run("Has", func(t *testing.T) {
		creds := New(map[string]string{"token": "abc"})

		if !creds.Has("token") {
			t.Errorf("expected Has(token) to be true")
		}

		if creds.Has("nonexistent") {
			t.Errorf("expected Has(nonexistent) to be false")
		}
	})
}
