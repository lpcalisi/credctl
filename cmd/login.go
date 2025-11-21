package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/client"
	"credctl/internal/protocol"
	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login <name>",
	Short: "Execute the login command for a provider",
	Long:  `Execute the interactive login command configured for a credential provider.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if name == "" {
			return fmt.Errorf("provider name cannot be empty")
		}

		// Get provider from daemon
		req := protocol.Request{
			Action: "get",
			Payload: protocol.GetPayload{
				Name: name,
			},
		}

		resp, err := client.SendRequest(req)
		if err != nil {
			return err
		}

		if resp.Status == "error" {
			return fmt.Errorf("error: %s", resp.Error)
		}

		// Extract metadata and type from response
		payloadBytes, err := json.Marshal(resp.Payload)
		if err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		var getRespPayload protocol.GetResponsePayload
		if err := json.Unmarshal(payloadBytes, &getRespPayload); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		metadata := getRespPayload.Metadata
		if metadata == nil {
			return fmt.Errorf("provider data not found")
		}

		// Get provider type from metadata (should be included in GetResponsePayload)
		// For now, we need to load the provider to check if it supports login
		// We'll load it from disk directly
		prov, err := provider.Load(name)
		if err != nil {
			return fmt.Errorf("failed to load provider: %w", err)
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

		fmt.Printf("Login successful for provider '%s'\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
