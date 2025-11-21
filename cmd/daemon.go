package cmd

import (
	"fmt"

	"credctl/internal/daemon"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the credctl daemon",
	Long: `Start the credctl daemon in the background.
The daemon listens on a Unix socket and handles credential requests.

To configure your shell, run:
  eval $(credctl daemon)`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := daemon.Run()
		if err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Print ssh-agent style output
		fmt.Printf("CREDCTL_SOCK=%s; export CREDCTL_SOCK;\n", info.AdminSocket)
		fmt.Printf("CREDCTL_PID=%d; export CREDCTL_PID;\n", info.PID)
		fmt.Printf("echo Agent pid %d;\n", info.PID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
