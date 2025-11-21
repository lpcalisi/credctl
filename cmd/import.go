package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"credctl/internal/client"
	"credctl/internal/protocol"
	"credctl/internal/provider"

	"github.com/spf13/cobra"
)

var (
	overwrite bool
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import providers from JSON file or stdin",
	Long:  `Import credential providers from a JSON file or stdin. By default, skips existing providers.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var data []byte
		var err error

		// Read from file or stdin
		if len(args) == 0 || args[0] == "-" {
			// Read from stdin
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
		} else {
			// Read from file
			filePath := args[0]
			data, err = os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
		}

		// Parse JSON
		var providers []*provider.Provider
		if err := json.Unmarshal(data, &providers); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		if len(providers) == 0 {
			fmt.Println("No providers found in file")
			return nil
		}

		// Import each provider via daemon
		imported := 0
		skipped := 0
		failed := 0

		for _, prov := range providers {
			// Check if provider already exists (unless overwrite flag is set)
			if !overwrite {
				_, err := provider.Load(prov.Name)
				if err == nil {
					fmt.Printf("Skipping '%s' (already exists, use --overwrite to replace)\n", prov.Name)
					skipped++
					continue
				}
			}

			// Send add request to daemon
			req := protocol.Request{
				Action: "add",
				Payload: protocol.AddPayload{
					Provider: *prov,
				},
			}

			resp, err := client.SendRequest(req)
			if err != nil {
				fmt.Printf("Failed to import '%s': %v\n", prov.Name, err)
				failed++
				continue
			}

			if resp.Status == "error" {
				fmt.Printf("Failed to import '%s': %s\n", prov.Name, resp.Error)
				failed++
				continue
			}

			fmt.Printf("Imported '%s'\n", prov.Name)
			imported++
		}

		// Summary
		fmt.Printf("\nImport complete: %d imported, %d skipped, %d failed\n", imported, skipped, failed)

		if failed > 0 {
			return fmt.Errorf("some providers failed to import")
		}

		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing providers")
	rootCmd.AddCommand(importCmd)
}
