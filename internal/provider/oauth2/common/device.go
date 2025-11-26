package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

func requestDeviceAuthorization(deviceEndpoint, clientID string, scopes []string) (*DeviceAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	body, err := postFormJSON(deviceEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("failed to request device authorization: %w", err)
	}

	var deviceAuth DeviceAuthResponse
	if err := json.Unmarshal(body, &deviceAuth); err != nil {
		return nil, fmt.Errorf("failed to parse device authorization response: %w", err)
	}

	if deviceAuth.Error != "" {
		return nil, fmt.Errorf("device authorization failed: %s - %s", deviceAuth.Error, deviceAuth.ErrorDescription)
	}

	return &deviceAuth, nil
}

func pollForDeviceTokenResponse(ctx context.Context, tokenEndpoint, clientID, clientSecret, deviceCode string, deviceAuth *DeviceAuthResponse) (*TokenResponse, error) {
	interval := time.Duration(deviceAuth.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(time.Duration(deviceAuth.ExpiresIn) * time.Second)

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)
	data.Set("client_id", clientID)
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("device authorization timed out")
		}

		body, err := postFormJSON(tokenEndpoint, data)
		if err != nil {
			return nil, fmt.Errorf("failed to poll token endpoint: %w", err)
		}

		var tokenResp TokenResponse
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			return nil, fmt.Errorf("failed to parse token response: %w", err)
		}

		switch tokenResp.Error {
		case "":
			return &tokenResp, nil

		case "authorization_pending":
			time.Sleep(interval)
			continue

		case "slow_down":
			interval += 5 * time.Second
			time.Sleep(interval)
			continue

		case "expired_token":
			return nil, fmt.Errorf("device code expired, please try again")

		case "access_denied":
			return nil, fmt.Errorf("access denied by user")

		default:
			return nil, fmt.Errorf("token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
		}
	}
}

func AuthenticateDeviceFlow(ctx context.Context, deviceEndpoint, tokenEndpoint, clientID, clientSecret string, scopes []string) (*TokenResponse, error) {
	deviceAuth, err := requestDeviceAuthorization(deviceEndpoint, clientID, scopes)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "To authenticate, visit:")

	verificationURL := deviceAuth.VerificationURIComplete
	if verificationURL == "" {
		verificationURL = deviceAuth.VerificationURI
	}
	if verificationURL == "" {
		verificationURL = deviceAuth.VerificationURL
	}

	if verificationURL != "" {
		hyperlink := formatHyperlink(verificationURL, verificationURL)
		fmt.Fprintf(os.Stderr, "  %s\n", hyperlink)
		if deviceAuth.VerificationURIComplete == "" && deviceAuth.UserCode != "" {
			fmt.Fprintf(os.Stderr, "\nAnd enter the code: %s\n", deviceAuth.UserCode)
		}
	} else {
		fmt.Fprintln(os.Stderr, "  (URL not provided by server)")
		if deviceAuth.UserCode != "" {
			fmt.Fprintf(os.Stderr, "\nUser code: %s\n", deviceAuth.UserCode)
		}
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Waiting for authentication...")

	tokenResp, err := pollForDeviceTokenResponse(ctx, tokenEndpoint, clientID, clientSecret, deviceAuth.DeviceCode, deviceAuth)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(os.Stderr, "Authentication successful!")

	return tokenResp, nil
}
