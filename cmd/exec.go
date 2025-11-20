package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

// executeInteractive runs a command interactively with inherited stdin/stdout/stderr
func executeInteractive(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
