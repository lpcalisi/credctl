package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:          "export",
	Short:        "Export all providers to JSON",
	Long:         `Export all credential providers to JSON format for backup or distribution.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// List all providers
		names, err := provider.List()
		if err != nil {
			return fmt.Errorf("failed to list providers: %w", err)
		}

		// Load all providers
		var providers []*provider.Provider
		for _, name := range names {
			prov, err := provider.Load(name)
			if err != nil {
				return fmt.Errorf("failed to load provider %s: %w", name, err)
			}
			providers = append(providers, prov)
		}

		// Marshal to JSON with indentation
		data, err := json.MarshalIndent(providers, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal providers: %w", err)
		}

		// Output to stdout
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

