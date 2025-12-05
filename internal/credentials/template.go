package credentials

import (
	"bytes"
	"fmt"
	"text/template"
)

// ApplyTemplate applies a Go template to structured credentials
// The template has access to all credential fields using the
// {{.field_name}} syntax
//
// If any credential field contains a JWT token, its standard claims
// are automatically extracted and made available with a suffix:
//   - {{.token_exp}} - expiration timestamp
//   - {{.token_iat}} - issued at timestamp
//   - {{.token_sub}} - subject
//   - {{.token_iss}} - issuer
//   - {{.token_payload}} - full JWT payload as JSON
//
// Example:
//
//	template: "export TOKEN={{.token}}"
//	creds: &Credentials{Fields: map[string]string{"token": "abc123"}}
//	result: "export TOKEN=abc123"
func ApplyTemplate(creds *Credentials, tmplStr string) ([]byte, error) {
	if creds == nil {
		return nil, fmt.Errorf("credentials cannot be nil")
	}

	if tmplStr == "" {
		return nil, fmt.Errorf("template string cannot be empty")
	}

	// Enrich credentials with JWT claims if any tokens are JWTs
	creds.EnrichWithJWTClaims()

	// Parse the template
	tmpl, err := template.New("output").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Execute the template with credential fields
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, creds.Fields); err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	return buf.Bytes(), nil
}
