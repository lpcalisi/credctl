package cmd

import (
	"fmt"

	"credctl/internal/client"
	"credctl/internal/protocol"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a credential provider",
	Long: `Delete a credential provider by name.
This removes the provider configuration from disk.`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if name == "" {
			return fmt.Errorf("provider name cannot be empty")
		}

		// Send request to daemon
		req := protocol.Request{
			Action: "delete",
			Payload: protocol.DeletePayload{
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

		fmt.Printf("Provider '%s' deleted successfully\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
