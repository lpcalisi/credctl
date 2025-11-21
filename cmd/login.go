package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/client"
	"credctl/internal/protocol"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:          "login <name>",
	Short:        "Execute the login command for a provider",
	Long:         `Execute the interactive login command configured for a credential provider.`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
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

		// Extract provider from response
		payloadBytes, err := json.Marshal(resp.Payload)
		if err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		var getRespPayload protocol.GetResponsePayload
		if err := json.Unmarshal(payloadBytes, &getRespPayload); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		prov := getRespPayload.Provider
		if prov == nil {
			return fmt.Errorf("provider data not found")
		}

		// Check if login command is configured
		if prov.LoginCommand == "" {
			return fmt.Errorf("provider '%s' does not have a login command configured\nUse 'credctl add' with --login flag to configure one", name)
		}

		// Execute login command interactively
		fmt.Printf("Running login command for '%s': %s\n", name, prov.LoginCommand)
		if err := executeInteractive(prov.LoginCommand); err != nil {
			return fmt.Errorf("login command failed: %w", err)
		}

		fmt.Printf("Login successful for provider '%s'\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
