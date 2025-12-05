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
	var templateStr string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get credentials from a provider",
		Long:  `Get credentials from a provider using its configured format.`,
		Args:  cobra.ExactArgs(1),
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
				// Handle errors based on error type (structured error handling)
				switch resp.ErrorType {
				case protocol.ErrorTypeAuthRequired:
					return fmt.Errorf("authentication required for provider '%s'\n\nRun: credctl login %s", name, name)
				case protocol.ErrorTypeDeviceFlowRequired:
					// Device flow error already has a descriptive message
					return fmt.Errorf("%s", resp.Error)
				default:
					return fmt.Errorf("error: %s", resp.Error)
				}
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
			structuredFields := getRespPayload.StructuredFields
			hasStructuredFields := getRespPayload.HasStructuredFields

			// Determine the final output
			var finalOutput []byte

			// Check if template is requested (from flag or metadata)
			effectiveTemplate := templateStr
			if effectiveTemplate == "" && metadata != nil {
				if tmpl, ok := metadata[provider.MetadataTemplate].(string); ok {
					effectiveTemplate = tmpl
				}
			}

			// If template is requested, apply it
			if effectiveTemplate != "" {
				if !hasStructuredFields {
					return fmt.Errorf("template requested but provider does not support structured credentials")
				}

				creds := formatter.NewCredentials(structuredFields)
				templatedOutput, err := formatter.ApplyTemplate(creds, effectiveTemplate)
				if err != nil {
					return fmt.Errorf("failed to apply template: %w", err)
				}
				finalOutput = templatedOutput
			} else {
				// No template, use raw output
				finalOutput = []byte(rawOutput)
			}

			// Handle output destination
			if outputPath != "" {
				// Write to file
				if err := formatter.WriteOutput(finalOutput, outputPath); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				fmt.Printf("Credentials written to %s\n", outputPath)
				return nil
			}

			// Output to stdout
			fmt.Print(string(finalOutput))
			if len(finalOutput) > 0 && finalOutput[len(finalOutput)-1] != '\n' {
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&templateStr, "template", "", "Go template to format credentials (e.g., 'export TOKEN={{.token}}')")
	cmd.Flags().StringVar(&outputPath, "output", "", "Write output to file (validates JSON if extension is .json)")

	return cmd
}
