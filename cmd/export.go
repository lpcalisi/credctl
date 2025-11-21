package cmd

import (
	"encoding/json"
	"fmt"

	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

// ExportedProvider represents a provider for export
type ExportedProvider struct {
	Name string         `json:"name"`
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all providers to JSON",
	Long:  `Export all credential providers to JSON format for backup or distribution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// List all providers
		names, err := provider.List()
		if err != nil {
			return fmt.Errorf("failed to list providers: %w", err)
		}

		// Load all providers
		var exported []ExportedProvider
		for _, name := range names {
			prov, err := provider.Load(name)
			if err != nil {
				return fmt.Errorf("failed to load provider %s: %w", name, err)
			}

			exported = append(exported, ExportedProvider{
				Name: name,
				Type: prov.Type(),
				Data: prov.Metadata(),
			})
		}

		// Marshal to JSON with indentation
		data, err := json.MarshalIndent(exported, "", "  ")
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
