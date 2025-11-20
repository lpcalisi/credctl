package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Provider represents a credential provider configuration
type Provider struct {
	Name         string `json:"name"`
	Command      string `json:"command"`
	Format       string `json:"format,omitempty"`        // Output format: raw, env, file
	EnvVar       string `json:"env_var,omitempty"`       // For env format
	FilePath     string `json:"file_path,omitempty"`     // For file format
	FileMode     string `json:"file_mode,omitempty"`     // For file format
	LoginCommand string `json:"login_command,omitempty"` // Command to run for interactive login
}

// ProvidersDir returns the directory where providers are stored
func ProvidersDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".credctl", "providers"), nil
}

// Save persists the provider to disk
func (p *Provider) Save() error {
	dir, err := ProvidersDir()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create providers directory: %w", err)
	}

	// Marshal provider to JSON
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provider: %w", err)
	}

	// Write to file
	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", p.Name))
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write provider file: %w", err)
	}

	return nil
}

// Load reads a provider from disk
func Load(name string) (*Provider, error) {
	dir, err := ProvidersDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(dir, name+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("provider not found: %s", name)
		}
		return nil, fmt.Errorf("failed to read provider file: %w", err)
	}

	var p Provider
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider: %w", err)
	}

	return &p, nil
}

// Delete removes a provider from disk
func Delete(name string) error {
	dir, err := ProvidersDir()
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, name+".json")
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("provider not found: %s", name)
		}
		return fmt.Errorf("failed to delete provider file: %w", err)
	}

	return nil
}

// List returns all provider names
func List() ([]string, error) {
	dir, err := ProvidersDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read providers directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}
	return names, nil
}

// GetFileModeInt converts the FileMode string to os.FileMode
func (p *Provider) GetFileModeInt() (os.FileMode, error) {
	if p.FileMode == "" {
		return 0600, nil // Default mode
	}

	// Parse octal string
	mode, err := strconv.ParseUint(p.FileMode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode: %s", p.FileMode)
	}

	return os.FileMode(mode), nil
}
