package credentials

import (
	"encoding/json"
	"fmt"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// Standard JWT claims that we extract
var standardClaims = []string{
	"exp", // Expiration time
	"iat", // Issued at
	"nbf", // Not before
	"sub", // Subject
	"iss", // Issuer
	"aud", // Audience
	"jti", // JWT ID
}

// EnrichWithJWTClaims inspects all credential fields and if any value
// looks like a JWT token, it parses the payload and adds the standard
// claims as additional fields with a suffix.
//
// For example, if field "token" contains a JWT with exp claim 1764978527,
// a new field "token_exp" will be added with value "1764978527".
//
// This allows templates to access JWT claims like:
//
//	{{.token_exp}} - expiration timestamp
//	{{.token_iat}} - issued at timestamp
//	{{.token_sub}} - subject
//	etc.
func (c *Credentials) EnrichWithJWTClaims() {
	if c.Fields == nil {
		return
	}

	// Collect new fields to add (don't modify map while iterating)
	newFields := make(map[string]string)

	for key, value := range c.Fields {
		claims, ok := parseJWTClaims(value)
		if !ok {
			continue
		}

		// Extract standard claims
		for _, claim := range standardClaims {
			if claimValue, exists := claims[claim]; exists {
				fieldName := key + "_" + claim
				newFields[fieldName] = formatClaimValue(claimValue)
			}
		}

		// Also add the full payload as JSON for custom claims access
		if payloadJSON, err := json.Marshal(claims); err == nil {
			newFields[fmt.Sprintf("%s_payload", key)] = string(payloadJSON)
		}
	}

	// Add all new fields
	for k, v := range newFields {
		c.Fields[k] = v
	}
}

// parseJWTClaims attempts to parse a string as a JWT and extract its claims.
// Returns the claims map and true if successful, nil and false otherwise.
// Note: This does NOT verify the JWT signature, only decodes the payload.
func parseJWTClaims(token string) (map[string]any, bool) {
	// Parse the JWT using go-jose (without signature verification)
	parsedToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{
		jose.RS256, jose.RS384, jose.RS512,
		jose.ES256, jose.ES384, jose.ES512,
		jose.PS256, jose.PS384, jose.PS512,
		jose.HS256, jose.HS384, jose.HS512,
		jose.EdDSA,
	})
	if err != nil {
		return nil, false
	}

	// Extract claims without verification
	var claims map[string]any
	if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, false
	}

	return claims, true
}

// formatClaimValue converts a claim value to a string representation
func formatClaimValue(v any) string {
	switch val := v.(type) {
	case float64:
		// Format numbers without decimal points for timestamps
		return fmt.Sprintf("%.0f", val)
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case []any:
		// For arrays (like aud can be), join with comma or marshal to JSON
		if len(val) == 1 {
			return formatClaimValue(val[0])
		}
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(jsonBytes)
	default:
		// For complex types, marshal to JSON
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(jsonBytes)
	}
}
