package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	var noDaemon bool

	cmd := &cobra.Command{
		Use:   "get <name|type>",
		Short: "Get credentials from a provider",
		Long: `Get credentials from a provider using its configured format.

Daemon mode (default):
  credctl get <name>
  Retrieves credentials from a named provider registered with the daemon.

No-daemon mode (--no-daemon):
  credctl get <type> --no-daemon [provider-flags]
  Executes the provider inline without daemon. All config passed as flags.
  Ideal for CI environments.`,
		DisableFlagParsing: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Handle --help manually since DisableFlagParsing is true
			if containsFlag(args, "--help") || containsFlag(args, "-h") {
				_ = cmd.Help()
				os.Exit(0)
			}

			// Detect --no-daemon mode from raw args
			noDaemon = containsFlag(args, "--no-daemon")

			if noDaemon {
				return preRunNoDaemon(cmd, args)
			}

			// Standard daemon mode - re-enable normal parsing
			cmd.DisableFlagParsing = false
			return cmd.ParseFlags(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if noDaemon {
				return runNoDaemon(cmd, args, templateStr, outputPath, format)
			}

			// Daemon mode: get non-flag args after parsing
			name := cmd.Flags().Arg(0)

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
	cmd.Flags().BoolVar(&noDaemon, "no-daemon", false, "Run without daemon (CI mode, requires provider type and config flags)")

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

// containsFlag checks if args contains the specified flag
func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

// getProviderTypeFromArgs extracts the provider type (first non-flag argument)
func getProviderTypeFromArgs(args []string) string {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}

// preRunNoDaemon handles pre-run setup for no-daemon mode
func preRunNoDaemon(cmd *cobra.Command, args []string) error {
	// Get provider type from args (first non-flag argument)
	providerType := getProviderTypeFromArgs(args)
	if providerType == "" {
		return fmt.Errorf("no-daemon mode requires provider type: credctl get <type> --no-daemon [flags]\nAvailable types: %v", provider.ListTypes())
	}

	if !provider.IsRegistered(providerType) {
		return fmt.Errorf("unknown provider type '%s'\nAvailable types: %v", providerType, provider.ListTypes())
	}

	// Register provider-specific flags
	schema, err := provider.GetSchema(providerType)
	if err != nil {
		return err
	}
	provider.AddSchemaFlags(cmd, schema)

	// Re-enable flag parsing and parse
	cmd.DisableFlagParsing = false
	return cmd.ParseFlags(args)
}

// runNoDaemon executes provider inline without daemon
func runNoDaemon(cmd *cobra.Command, args []string, templateStr, outputPath, format string) error {
	providerType := getProviderTypeFromArgs(args)

	// Get provider schema and extract config from flags
	schema, err := provider.GetSchema(providerType)
	if err != nil {
		return err
	}

	config, err := provider.ExtractConfig(cmd, schema)
	if err != nil {
		return fmt.Errorf("failed to extract configuration: %w", err)
	}

	// Create and initialize provider
	prov, err := provider.New(providerType)
	if err != nil {
		return err
	}

	if err := prov.Init(config); err != nil {
		return fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Get credentials (always use Get() for raw output, like daemon does)
	ctx := cmd.Context()

	rawOutput, err := prov.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	// Get structured fields if provider supports it (for templates)
	var structuredFields map[string]string
	hasStructuredFields := false

	if credProv, ok := prov.(provider.CredentialsProvider); ok {
		creds, err := credProv.GetCredentials(ctx)
		if err == nil && creds != nil && creds.Fields != nil {
			structuredFields = creds.Fields
			hasStructuredFields = true
		}
	}

	// Determine effective format (flag > default)
	if format == "" {
		format = "text"
	}

	// Apply template if specified
	var finalOutput []byte

	if templateStr != "" {
		if !hasStructuredFields {
			return fmt.Errorf("template requested but provider does not support structured credentials")
		}

		creds := credentials.New(structuredFields)
		templatedOutput, err := credentials.ApplyTemplate(creds, templateStr)
		if err != nil {
			return fmt.Errorf("failed to apply template: %w", err)
		}
		finalOutput = templatedOutput
	} else {
		finalOutput = rawOutput
	}

	// Apply format
	fmtr, err := formatter.Get(format)
	if err != nil {
		available := formatter.List()
		return fmt.Errorf("unsupported format '%s', available formats: %v", format, available)
	}

	formattedOutput, err := fmtr.Format(finalOutput)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Handle output destination
	if outputPath != "" {
		if err := output.Write(formattedOutput, outputPath); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Printf("Credentials written to %s\n", outputPath)
		return nil
	}

	// Output to stdout
	fmt.Print(string(formattedOutput))
	if len(formattedOutput) > 0 && formattedOutput[len(formattedOutput)-1] != '\n' {
		fmt.Println()
	}

	return nil
}
