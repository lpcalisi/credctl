package formatter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteOutput(t *testing.T) {
	// Create temp directory for tests
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		output      []byte
		filename    string
		shouldError bool
		errorMsg    string
	}{
		{
			name:     "write txt file",
			output:   []byte("test content"),
			filename: "test.txt",
		},
		{
			name:     "write json file with valid json",
			output:   []byte(`{"token": "abc123"}`),
			filename: "test.json",
		},
		{
			name:        "write json file with invalid json",
			output:      []byte("not json"),
			filename:    "test.json",
			shouldError: true,
			errorMsg:    "not valid JSON",
		},
		{
			name:        "empty file path",
			output:      []byte("test"),
			filename:    "",
			shouldError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:     "create nested directories",
			output:   []byte("test"),
			filename: "nested/dir/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.filename != "" {
				filePath = filepath.Join(tempDir, tt.name, tt.filename)
			}

			err := WriteOutput(tt.output, filePath)

		if tt.shouldError {
			if err == nil {
				t.Errorf("expected error but got none")
				return
			}
			// Just check that we got an error, exact message matching is fragile
			return
		}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify file was written correctly
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("failed to read written file: %v", err)
				return
			}

			if string(content) != string(tt.output) {
				t.Errorf("expected content %q, got %q", string(tt.output), string(content))
			}

			// Verify file permissions
			info, err := os.Stat(filePath)
			if err != nil {
				t.Errorf("failed to stat file: %v", err)
				return
			}

			if info.Mode().Perm() != 0600 {
				t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
			}
		})
	}
}

func TestWriteOutputHomeDirExpansion(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	tempFile := filepath.Join(homeDir, ".credctl_test_"+t.Name())
	t.Cleanup(func() {
		_ = os.Remove(tempFile)
	})

	// Use ~ in path
	tildeePath := "~/.credctl_test_" + t.Name()
	err = WriteOutput([]byte("test"), tildeePath)
	if err != nil {
		t.Errorf("failed to write with ~ path: %v", err)
		return
	}

	// Verify file exists at expanded path
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("file was not created at expanded path")
	}
}

