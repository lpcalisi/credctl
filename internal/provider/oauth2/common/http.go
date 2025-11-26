package common

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"credctl/internal/provider"
)

func postFormJSON(endpoint string, data url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return io.ReadAll(resp.Body)
}

func IsTokenValid(tokens *TokenCache) bool {
	if tokens == nil || tokens.AccessToken == "" {
		return false
	}
	return time.Now().Add(30 * time.Second).Before(tokens.ExpiresAt)
}

func GetScopes(config map[string]any) []string {
	return provider.GetStringSliceOrDefault(config, provider.MetadataScopes, nil)
}
