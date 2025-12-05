package credentials

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// createTestJWT creates a simple JWT for testing (unsigned)
// Note: This is NOT a valid signed JWT, just the structure for parsing tests
func createTestJWT(claims map[string]any) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Fake signature
	signature := base64.RawURLEncoding.EncodeToString([]byte("test-signature"))

	return header + "." + payload + "." + signature
}

func TestEnrichWithJWTClaims(t *testing.T) {
	tests := []struct {
		name           string
		fields         map[string]string
		expectedFields map[string]string
	}{
		{
			name: "JWT with exp and iat claims",
			fields: map[string]string{
				"token": createTestJWT(map[string]any{
					"exp": 1764978527,
					"iat": 1733442527,
					"sub": "user@example.com",
				}),
			},
			expectedFields: map[string]string{
				"token_exp": "1764978527",
				"token_iat": "1733442527",
				"token_sub": "user@example.com",
			},
		},
		{
			name: "JWT with all standard claims",
			fields: map[string]string{
				"token": createTestJWT(map[string]any{
					"exp": 1764978527,
					"iat": 1733442527,
					"nbf": 1733442527,
					"sub": "user123",
					"iss": "https://auth.example.com",
					"aud": "my-app",
					"jti": "unique-id-123",
				}),
			},
			expectedFields: map[string]string{
				"token_exp": "1764978527",
				"token_iat": "1733442527",
				"token_nbf": "1733442527",
				"token_sub": "user123",
				"token_iss": "https://auth.example.com",
				"token_aud": "my-app",
				"token_jti": "unique-id-123",
			},
		},
		{
			name: "multiple tokens with different field names",
			fields: map[string]string{
				"token": createTestJWT(map[string]any{
					"exp": 1764978527,
					"sub": "user1",
				}),
				"access_token": createTestJWT(map[string]any{
					"exp": 1764979999,
					"sub": "user2",
				}),
			},
			expectedFields: map[string]string{
				"token_exp":        "1764978527",
				"token_sub":        "user1",
				"access_token_exp": "1764979999",
				"access_token_sub": "user2",
			},
		},
		{
			name: "non-JWT token is ignored",
			fields: map[string]string{
				"token": "not-a-jwt-token",
			},
			expectedFields: map[string]string{},
		},
		{
			name: "mixed JWT and non-JWT fields",
			fields: map[string]string{
				"token":    createTestJWT(map[string]any{"exp": 1764978527}),
				"api_key":  "simple-api-key",
				"username": "testuser",
			},
			expectedFields: map[string]string{
				"token_exp": "1764978527",
			},
		},
		{
			name: "JWT with array audience",
			fields: map[string]string{
				"token": createTestJWT(map[string]any{
					"exp": 1764978527,
					"aud": []string{"app1", "app2"},
				}),
			},
			expectedFields: map[string]string{
				"token_exp": "1764978527",
				"token_aud": `["app1","app2"]`,
			},
		},
		{
			name:           "nil fields",
			fields:         nil,
			expectedFields: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &Credentials{Fields: tt.fields}
			creds.EnrichWithJWTClaims()

			// Check that all expected fields were added
			for key, expectedValue := range tt.expectedFields {
				if gotValue, exists := creds.Fields[key]; !exists {
					t.Errorf("expected field %q to exist", key)
				} else if gotValue != expectedValue {
					t.Errorf("field %q: expected %q, got %q", key, expectedValue, gotValue)
				}
			}

			// Check that payload field was added for JWT tokens
			if tt.fields != nil {
				for key, value := range tt.fields {
					if _, isJWT := parseJWTClaims(value); isJWT {
						payloadKey := key + "_payload"
						if _, exists := creds.Fields[payloadKey]; !exists {
							t.Errorf("expected payload field %q to exist", payloadKey)
						}
					}
				}
			}
		})
	}
}

func TestParseJWTClaims(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantOk   bool
		wantExp  float64
		wantSub  string
	}{
		{
			name:    "valid JWT",
			token:   createTestJWT(map[string]any{"exp": 1764978527, "sub": "test"}),
			wantOk:  true,
			wantExp: 1764978527,
			wantSub: "test",
		},
		{
			name:   "not a JWT - no dots",
			token:  "justasimpletoken",
			wantOk: false,
		},
		{
			name:   "not a JWT - only one dot",
			token:  "part1.part2",
			wantOk: false,
		},
		{
			name:   "not a JWT - four parts",
			token:  "part1.part2.part3.part4",
			wantOk: false,
		},
		{
			name:   "invalid base64 in payload",
			token:  "eyJhbGciOiJIUzI1NiJ9.!!!invalid!!!.signature",
			wantOk: false,
		},
		{
			name:   "valid base64 but not JSON",
			token:  "eyJhbGciOiJIUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte("not json")) + ".sig",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, ok := parseJWTClaims(tt.token)

			if ok != tt.wantOk {
				t.Errorf("parseJWTClaims() ok = %v, want %v", ok, tt.wantOk)
				return
			}

			if !ok {
				return
			}

			if exp, exists := claims["exp"].(float64); exists && exp != tt.wantExp {
				t.Errorf("exp = %v, want %v", exp, tt.wantExp)
			}

			if sub, exists := claims["sub"].(string); exists && sub != tt.wantSub {
				t.Errorf("sub = %v, want %v", sub, tt.wantSub)
			}
		})
	}
}

func TestFormatClaimValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"float64 timestamp", float64(1764978527), "1764978527"},
		{"string", "hello", "hello"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"single element array", []any{"single"}, "single"},
		{"multi element array", []any{"a", "b"}, `["a","b"]`},
		{"nested object", map[string]any{"key": "value"}, `{"key":"value"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatClaimValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatClaimValue(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestApplyTemplateWithJWT(t *testing.T) {
	// Test that ApplyTemplate automatically enriches JWT claims
	jwt := createTestJWT(map[string]any{
		"exp": 1764978527,
		"iat": 1733442527,
		"sub": "user@example.com",
	})

	creds := New(map[string]string{
		"token": jwt,
	})

	// Template that uses JWT claims
	template := `{"tiger-token": "{{.token}}", "tiger-token-expiration": {{.token_exp}}}`

	result, err := ApplyTemplate(creds, template)
	if err != nil {
		t.Fatalf("ApplyTemplate failed: %v", err)
	}

	expected := `{"tiger-token": "` + jwt + `", "tiger-token-expiration": 1764978527}`
	if string(result) != expected {
		t.Errorf("got:\n%s\nwant:\n%s", string(result), expected)
	}
}

