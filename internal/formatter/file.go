package formatter

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"credctl/internal/provider"
)

// FileFromMetadata writes the output to a file using metadata configuration
func FileFromMetadata(output string, metadata map[string]any) (string, error) {
	filePath, ok := metadata[provider.MetadataFilePath].(string)
	if !ok || filePath == "" {
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

	// Get file mode (default to 0600)
	fileMode := os.FileMode(0600)
	if fileModeStr, ok := metadata[provider.MetadataFileMode].(string); ok && fileModeStr != "" {
		mode, err := strconv.ParseUint(fileModeStr, 8, 32)
		if err != nil {
			return "", fmt.Errorf("invalid file mode: %s", fileModeStr)
		}
		fileMode = os.FileMode(mode)
	}

	// Create parent directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(output), fileMode); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}
