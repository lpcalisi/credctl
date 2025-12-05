package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"credctl/internal/client"
	"credctl/internal/protocol"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

// List returns the list command
func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured providers",
		Long:  `List all configured providers with their types.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Send request to daemon
			req := protocol.Request{
				Action:  "list",
				Payload: nil, // No payload needed for list
			}

			resp, err := client.SendRequest(req)
			if err != nil {
				return err
			}

			if resp.Status == "error" {
				return fmt.Errorf("error: %s", resp.Error)
			}

			// Extract providers from payload
			payloadBytes, err := json.Marshal(resp.Payload)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			var listResp protocol.ListResponsePayload
			if err := json.Unmarshal(payloadBytes, &listResp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Sort providers by name for consistent output
			sort.Slice(listResp.Providers, func(i, j int) bool {
				return listResp.Providers[i].Name < listResp.Providers[j].Name
			})

			// Display providers with styled output
			if len(listResp.Providers) == 0 {
				noProvidersStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Italic(true)
				fmt.Println(noProvidersStyle.Render("No providers configured."))
				return nil
			}

			// Create styles
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				MarginBottom(1)

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

			nameStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

			typeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("141"))

			borderStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

			// Calculate column widths
			nameColWidth := len("NAME")
			typeColWidth := len("TYPE")
			for _, prov := range listResp.Providers {
				if len(prov.Name) > nameColWidth {
					nameColWidth = len(prov.Name)
				}
				if len(prov.Type) > typeColWidth {
					typeColWidth = len(prov.Type)
				}
			}

			// Add some padding
			nameColWidth += 2
			typeColWidth += 2

			// Print title
			fmt.Println(titleStyle.Render("Configured Providers"))

			// Print table header
			nameHeader := headerStyle.Render(fmt.Sprintf("%-*s", nameColWidth, "NAME"))
			typeHeader := headerStyle.Render(fmt.Sprintf("%-*s", typeColWidth, "TYPE"))
			separator := borderStyle.Render("│")
			fmt.Printf("%s %s %s\n", nameHeader, separator, typeHeader)

			// Print separator line
			nameLine := strings.Repeat("─", nameColWidth)
			typeLine := strings.Repeat("─", typeColWidth)
			crossSep := borderStyle.Render("┼")
			fmt.Printf("%s %s %s\n", borderStyle.Render(nameLine), crossSep, borderStyle.Render(typeLine))

			// Print providers
			for _, prov := range listResp.Providers {
				name := nameStyle.Render(fmt.Sprintf("%-*s", nameColWidth, prov.Name))
				provType := typeStyle.Render(fmt.Sprintf("%-*s", typeColWidth, prov.Type))
				separator := borderStyle.Render("│")
				fmt.Printf("%s %s %s\n", name, separator, provType)
			}

			// Print footer with count
			footerStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true).
				MarginTop(1)

			countMsg := fmt.Sprintf("Total: %d provider(s)", len(listResp.Providers))
			fmt.Println(footerStyle.Render(countMsg))

			return nil
		},
	}

	return cmd
}
