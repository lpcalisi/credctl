package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/client"
	"credctl/internal/protocol"
	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

func Login() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login <name>",
		Short: "Execute the login command for a provider",
		Long:  `Execute the interactive login command configured for a credential provider.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if name == "" {
				return fmt.Errorf("provider name cannot be empty")
			}

			// Try to get provider info from daemon first
			req := protocol.Request{
				Action: "describe",
				Payload: protocol.DescribePayload{
					Name: name,
				},
			}

			resp, err := client.SendRequest(req)
			var prov provider.Provider
			if err == nil && resp.Status == "ok" {
				// Successfully got provider info from daemon
				payloadBytes, err := json.Marshal(resp.Payload)
				if err != nil {
					return fmt.Errorf("failed to parse daemon response: %w", err)
				}

				var describeResp protocol.DescribeResponsePayload
				if err := json.Unmarshal(payloadBytes, &describeResp); err != nil {
					return fmt.Errorf("failed to parse daemon response: %w", err)
				}

				// Create provider instance from daemon info
				prov, err = provider.FromMetadata(describeResp.Type, describeResp.Metadata)
				if err != nil {
					return fmt.Errorf("failed to create provider from daemon info: %w", err)
				}
			} else {
				// Daemon approach failed, try loading from disk
				prov, err = provider.Load(name)
				if err != nil {
					return fmt.Errorf("failed to load provider: %w", err)
				}
			}

			// Check if provider supports login
			loginProvider, ok := prov.(provider.LoginProvider)
			if !ok {
				return fmt.Errorf("provider '%s' (type: %s) does not support login", name, prov.Type())
			}

			// Execute provider-specific login
			fmt.Printf("Running login for provider '%s'...\n", name)
			if err := loginProvider.Login(cmd.Context()); err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// If provider supports token caching, send tokens to daemon
			if tokenCacheProv, ok := prov.(provider.TokenCacheProvider); ok {
				accessToken, refreshToken, expiresIn := tokenCacheProv.GetTokens()
				if accessToken != "" {
					// Send tokens to daemon
					req := protocol.Request{
						Action: "set_tokens",
						Payload: protocol.SetTokensPayload{
							Name:         name,
							AccessToken:  accessToken,
							RefreshToken: refreshToken,
							ExpiresIn:    expiresIn,
						},
					}

					resp, err := client.SendRequest(req)
					if err != nil {
						// Log warning but don't fail - login was successful
						fmt.Printf("Warning: failed to sync tokens with daemon: %v\n", err)
					} else if resp.Status == "error" {
						fmt.Printf("Warning: daemon rejected tokens: %s\n", resp.Error)
					}
				}
			}

			fmt.Printf("Login successful for provider '%s'\n", name)
			return nil
		},
	}

	return cmd
}
