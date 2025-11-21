package cmd

import (
	"fmt"

	"credctl/internal/client"
	"credctl/internal/protocol"
	"credctl/internal/provider"
	_ "credctl/internal/provider/command" // Import to register provider

	"github.com/spf13/cobra"
)

var (
	runLogin bool
)

var addCmd = &cobra.Command{
	Use:   "add <type> <name>",
	Short: "Add a credential provider",
	Long: `Add a credential provider.

Examples:
  credctl add command github --command "gh auth token" --format env
  credctl add command aws --command "aws sts get-session-token" --format raw
  
Available provider types: ` + fmt.Sprintf("%v", provider.ListTypes()),
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerType := args[0]
		name := args[1]

		if providerType == "" {
			return fmt.Errorf("provider type cannot be empty")
		}

		if name == "" {
			return fmt.Errorf("provider name cannot be empty")
		}

		// Get schema for the provider type
		schema, err := provider.GetSchema(providerType)
		if err != nil {
			return fmt.Errorf("unknown provider type '%s': %w\nAvailable types: %v", providerType, err, provider.ListTypes())
		}

		// Extract config from flags
		config, err := provider.ExtractConfig(cmd, schema)
		if err != nil {
			return fmt.Errorf("failed to extract configuration: %w", err)
		}

		// Create provider instance
		prov, err := provider.New(providerType)
		if err != nil {
			return err
		}

		// Initialize provider with config
		if err := prov.Init(config); err != nil {
			return fmt.Errorf("failed to initialize provider: %w", err)
		}

		// Execute login command if requested
		if runLogin {
			loginProvider, ok := prov.(provider.LoginProvider)
			if !ok {
				return fmt.Errorf("provider type '%s' does not support login", providerType)
			}

			fmt.Printf("Running login for provider '%s'...\n", name)
			if err := loginProvider.Login(cmd.Context()); err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
			fmt.Println("Login successful")
		}

		// Send request to daemon with provider name and metadata
		req := protocol.Request{
			Action: "add",
			Payload: protocol.AddPayload{
				Name:     name,
				Type:     prov.Type(),
				Metadata: prov.Metadata(),
			},
		}

		resp, err := client.SendRequest(req)
		if err != nil {
			return err
		}

		if resp.Status == "error" {
			return fmt.Errorf("error: %s", resp.Error)
		}

		fmt.Printf("Provider '%s' added successfully\n", name)
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&runLogin, "run-login", false, "Execute the login command before adding the provider")

	// Add flags for all registered provider types
	provider.AddAllProviderFlags(addCmd)

	rootCmd.AddCommand(addCmd)
}
