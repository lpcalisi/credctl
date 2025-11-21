package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "credctl",
	Short: "A tiny SSH-forwardable credential agent",
	Long: `credctl is a credential agent that allows you to register commands
that retrieve credentials and execute them on demand with different output formats.

The agent runs as a daemon and communicates via a Unix socket.`,
}

// Execute runs the root command with Fang for beautiful output
func Execute() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
