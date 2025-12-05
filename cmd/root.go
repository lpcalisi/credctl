package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"sigs.k8s.io/release-utils/version"
)

// Root returns the root command for credctl
func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credctl",
		Short: "A tiny SSH-forwardable credential agent",
		Long: `credctl is a credential agent that allows you to register commands
that retrieve credentials and execute them on demand with different output formats.

The agent runs as a daemon and communicates via a Unix socket.`,
		Version:           "placeholder", // Required for Cobra/Fang to add -v flag
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(Add())
	cmd.AddCommand(Get())
	cmd.AddCommand(Delete())
	cmd.AddCommand(List())
	cmd.AddCommand(Daemon())
	cmd.AddCommand(Export())
	cmd.AddCommand(Import())
	cmd.AddCommand(Login())

	return cmd
}

// Execute runs the root command with Fang for beautiful output
func Execute() {
	rootCmd := Root()

	// Configure version template to use release-utils format with ASCII art
	info := version.GetVersionInfo()
	info.Name = rootCmd.Name()
	info.Description = rootCmd.Short
	rootCmd.SetVersionTemplate(info.String())

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
