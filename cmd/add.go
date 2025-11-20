package cmd

import (
	"fmt"

	"credctl/internal/client"
	"credctl/internal/protocol"
	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

var (
	format   string
	envVar   string
	filePath string
	fileMode string
	login    string
	runLogin bool
)

var addCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Add a credential provider",
	Long:  `Add a credential provider that executes a command to retrieve credentials.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		command := args[1]

		if name == "" {
			return fmt.Errorf("provider name cannot be empty")
		}

		if command == "" {
			return fmt.Errorf("provider command cannot be empty")
		}

		// Execute login command if requested
		if runLogin && login != "" {
			fmt.Printf("Running login command: %s\n", login)
			if err := executeInteractive(login); err != nil {
				return fmt.Errorf("login command failed: %w", err)
			}
			fmt.Println("Login successful")
		}

		// Create provider
		prov := provider.Provider{
			Name:         name,
			Command:      command,
			Format:       format,
			EnvVar:       envVar,
			FilePath:     filePath,
			FileMode:     fileMode,
			LoginCommand: login,
		}

		// Send request to daemon
		req := protocol.Request{
			Action: "add",
			Payload: protocol.AddPayload{
				Provider: prov,
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
	addCmd.Flags().StringVar(&format, "format", "raw", "Output format: raw, env, file")
	addCmd.Flags().StringVar(&envVar, "env-var", "", "Environment variable name (for env format)")
	addCmd.Flags().StringVar(&filePath, "file-path", "", "File path to write credentials to (for file format)")
	addCmd.Flags().StringVar(&fileMode, "file-mode", "0600", "File permissions in octal format (for file format)")
	addCmd.Flags().StringVar(&login, "login", "", "Login command to execute for interactive authentication")
	addCmd.Flags().BoolVar(&runLogin, "run-login", false, "Execute the login command before adding the provider")
	rootCmd.AddCommand(addCmd)
}
