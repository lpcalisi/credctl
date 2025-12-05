package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/client"
	"credctl/internal/credentials"
	"credctl/internal/formatter"
	"credctl/internal/output"
	"credctl/internal/protocol"
	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

// Get returns the get command
func Get() *cobra.Command {
	var templateStr string
	var outputPath string
	var format string

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get credentials from a provider",
		Long:  `Get credentials from a provider using its configured format.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0] // Use only the first argument, ignore additional ones

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

			// Determine effective values (flag > metadata > default)
			effectiveFormat := getEffective(format, metadata, provider.MetadataFormat, "text")
			effectiveOutput := getEffective(outputPath, metadata, provider.MetadataOutput, "")
			effectiveTemplate := getEffective(templateStr, metadata, provider.MetadataTemplate, "")

			// Apply template if specified
			var finalOutput []byte

			// If template is requested, apply it
			if effectiveTemplate != "" {
				if !hasStructuredFields {
					return fmt.Errorf("template requested but provider does not support structured credentials")
				}

				creds := credentials.New(structuredFields)
				templatedOutput, err := credentials.ApplyTemplate(creds, effectiveTemplate)
				if err != nil {
					return fmt.Errorf("failed to apply template: %w", err)
				}
				finalOutput = templatedOutput
			} else {
				// No template, use raw output
				finalOutput = []byte(rawOutput)
			}

			// Apply format
			fmtr, err := formatter.Get(effectiveFormat)
			if err != nil {
				// Show available formats in error
				available := formatter.List()
				return fmt.Errorf("unsupported format '%s', available formats: %v", effectiveFormat, available)
			}

			formattedOutput, err := fmtr.Format(finalOutput)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			// Handle output destination
			if effectiveOutput != "" {
				// Write to file
				if err := output.Write(formattedOutput, effectiveOutput); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				fmt.Printf("Credentials written to %s\n", effectiveOutput)
				return nil
			}

			// Output to stdout
			fmt.Print(string(formattedOutput))
			if len(formattedOutput) > 0 && formattedOutput[len(formattedOutput)-1] != '\n' {
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&templateStr, "template", "", "Go template to format credentials (e.g., 'export TOKEN={{.token}}')")
	cmd.Flags().StringVar(&outputPath, "output", "", "Write output to file")
	cmd.Flags().StringVar(&format, "format", "", "Output format: json, text, escaped (default: text, or provider's default)")

	return cmd
}

// getEffective returns the effective value for a configuration option
// Priority: flag value > metadata value > default value
func getEffective(flagValue string, metadata map[string]any, metadataKey string, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if metadata != nil {
		if val, ok := metadata[metadataKey].(string); ok && val != "" {
			return val
		}
	}
	return defaultValue
}
