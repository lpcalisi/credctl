package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteOutput writes output to a file with validation based on file extension
// Supported extensions:
// - .json: validates that content is valid JSON before writing
// - .txt: no validation
// - other extensions: no validation (same as .txt)
//
// Features:
// - Expands ~ to home directory
// - Creates parent directories if they don't exist
// - Default permissions 0600
func WriteOutput(output []byte, filePath string) error {
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

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".json" {
		if !json.Valid(output) {
			return fmt.Errorf("output is not valid JSON")
		}
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
