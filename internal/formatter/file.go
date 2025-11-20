package formatter

import (
	"fmt"
	"os"
	"path/filepath"

	"credctl/internal/provider"
)

// File writes the output to the provider's configured file path
func File(output string, prov *provider.Provider) (string, error) {
	filePath := prov.FilePath
	if filePath == "" {
		return "", fmt.Errorf("no file path configured for provider (use --file-path when adding provider)")
	}

	// Expand home directory if needed
	if len(filePath) > 0 && filePath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = filepath.Join(homeDir, filePath[1:])
	}

	mode, err := prov.GetFileModeInt()
	if err != nil {
		return "", err
	}

	// Create parent directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(output), mode); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}
