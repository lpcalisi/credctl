package common

import (
	"context"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"golang.org/x/oauth2"
)

// AuthenticateDeviceFlow performs OAuth2 device authorization flow
func AuthenticateDeviceFlow(ctx context.Context, deviceEndpoint, tokenEndpoint, clientID, clientSecret string, scopes []string) (*TokenCache, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:       deviceEndpoint, // Device endpoint goes here
			TokenURL:      tokenEndpoint,
			DeviceAuthURL: deviceEndpoint,
		},
		Scopes: scopes,
	}

	// Request device authorization
	deviceAuth, err := config.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to request device authorization: %w", err)
	}

	// Display user instructions with nice formatting
	displayDeviceAuthInstructions(deviceAuth)

	// Poll for token
	token, err := config.DeviceAccessToken(ctx, deviceAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to get device token: %w", err)
	}

	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	fmt.Fprintln(os.Stderr, successStyle.Render("Authentication successful!"))

	return OAuth2TokenToCache(token), nil
}

// displayDeviceAuthInstructions shows formatted device auth instructions to the user
func displayDeviceAuthInstructions(deviceAuth *oauth2.DeviceAuthResponse) {
	boldStyle := lipgloss.NewStyle().Bold(true)
	codeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, boldStyle.Render("To authenticate, visit:"))

	verificationURL := deviceAuth.VerificationURIComplete
	if verificationURL == "" {
		verificationURL = deviceAuth.VerificationURI
	}

	if verificationURL != "" {
		hyperlink := formatHyperlink(verificationURL, verificationURL)
		fmt.Fprintf(os.Stderr, "  %s\n", hyperlink)
		if deviceAuth.VerificationURIComplete == "" && deviceAuth.UserCode != "" {
			fmt.Fprintf(os.Stderr, "\nAnd enter the code: %s\n", codeStyle.Render(deviceAuth.UserCode))
		}
	} else {
		fmt.Fprintln(os.Stderr, "  (URL not provided by server)")
		if deviceAuth.UserCode != "" {
			fmt.Fprintf(os.Stderr, "\nUser code: %s\n", codeStyle.Render(deviceAuth.UserCode))
		}
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, boldStyle.Render("Waiting for authentication..."))
}
