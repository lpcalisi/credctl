package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/client"
	"credctl/internal/formatter"
	"credctl/internal/protocol"
	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

// Get returns the get command
func Get() *cobra.Command {
	var raw bool

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get credentials from a provider",
		Long: `Get credentials by executing the provider's command.

By default, uses the format configured in the provider.
Use --raw to always get the raw credential value.`,
		Args: cobra.ExactArgs(1),
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
			metadata := getRespPayload.Metadata

			// If --raw flag is set, always output raw
			if raw {
				fmt.Print(formatter.Raw(rawOutput))
				if len(rawOutput) > 0 && rawOutput[len(rawOutput)-1] != '\n' {
					fmt.Println()
				}
				return nil
			}

			// Use provider's configured format (default to raw if not set)
			format := provider.FormatRaw
			if metadata != nil {
				if formatVal, ok := metadata[provider.MetadataFormat].(string); ok {
					format = formatVal
				}
			}

			switch format {
			case provider.FormatRaw:
				// Print raw output
				fmt.Print(formatter.Raw(rawOutput))
				if len(rawOutput) > 0 && rawOutput[len(rawOutput)-1] != '\n' {
					fmt.Println()
				}

			case provider.FormatEnv:
				// Format as environment variables
				envVar := ""
				if metadata != nil {
					if envVarVal, ok := metadata[provider.MetadataEnvVar].(string); ok {
						envVar = envVarVal
					}
				}
				formatted, err := formatter.Env(rawOutput, envVar)
				if err != nil {
					return fmt.Errorf("failed to format output: %w", err)
				}
				fmt.Println(formatted)

			case provider.FormatFile:
				// Write to file
				if metadata == nil {
					return fmt.Errorf("provider configuration not found")
				}

				writtenPath, err := formatter.FileFromMetadata(rawOutput, metadata)
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

	cmd.Flags().BoolVar(&raw, "raw", false, "Output raw credential value (ignores provider format)")

	return cmd
}
