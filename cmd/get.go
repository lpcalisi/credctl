package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/client"
	"credctl/internal/formatter"
	"credctl/internal/protocol"

	"github.com/spf13/cobra"
)

var (
	raw bool
)

var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get credentials from a provider",
	Long: `Get credentials by executing the provider's command.

By default, uses the format configured in the provider.
Use --raw to always get the raw credential value.`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if name == "" {
			return fmt.Errorf("provider name cannot be empty")
		}

		// Send request to daemon (daemon only returns raw output)
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

		// Extract output from payload
		payloadBytes, err := json.Marshal(resp.Payload)
		if err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		var getRespPayload protocol.GetResponsePayload
		if err := json.Unmarshal(payloadBytes, &getRespPayload); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		rawOutput := getRespPayload.Output
		prov := getRespPayload.Provider

		// If --raw flag is set, always output raw
		if raw {
			fmt.Print(formatter.Raw(rawOutput))
			if len(rawOutput) > 0 && rawOutput[len(rawOutput)-1] != '\n' {
				fmt.Println()
			}
			return nil
		}

		// Use provider's configured format (default to raw if not set)
		format := "raw"
		if prov != nil && prov.Format != "" {
			format = prov.Format
		}

		switch format {
		case "raw":
			// Print raw output
			fmt.Print(formatter.Raw(rawOutput))
			if len(rawOutput) > 0 && rawOutput[len(rawOutput)-1] != '\n' {
				fmt.Println()
			}

		case "env":
			// Format as environment variables
			envVar := ""
			if prov != nil {
				envVar = prov.EnvVar
			}
			formatted, err := formatter.Env(rawOutput, envVar)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}
			fmt.Println(formatted)

		case "file":
			// Write to file
			if prov == nil {
				return fmt.Errorf("provider configuration not found")
			}

			writtenPath, err := formatter.File(rawOutput, prov)
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			fmt.Printf("Credential written to %s\n", writtenPath)

		default:
			return fmt.Errorf("unknown format: %s", format)
		}

		return nil
	},
}

func init() {
	getCmd.Flags().BoolVar(&raw, "raw", false, "Output raw credential value (ignores provider format)")
	rootCmd.AddCommand(getCmd)
}
