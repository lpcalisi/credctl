package output

import (
	"fmt"
	"os"
	"path/filepath"
)

// Write writes output to a file
// Features:
// - Expands ~ to home directory
// - Creates parent directories if they don't exist
// - Default permissions 0600
func Write(output []byte, filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if len(filePath) > 0 && filePath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = filepath.Join(homeDir, filePath[1:])
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, output, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

