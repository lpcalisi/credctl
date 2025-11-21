package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StoredProvider represents a provider as stored in JSON
type StoredProvider struct {
	Name string         `json:"name"`
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// ProvidersDir returns the directory where providers are stored
func ProvidersDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".credctl", "providers"), nil
}

// Save persists a provider to disk
func Save(name string, prov Provider) error {
	dir, err := ProvidersDir()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create providers directory: %w", err)
	}

	// Create stored provider structure
	stored := StoredProvider{
		Name: name,
		Type: prov.Type(),
		Data: prov.Metadata(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provider: %w", err)
	}

	// Write to file
	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", name))
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write provider file: %w", err)
	}

	return nil
}

// Load reads a provider from disk
func Load(name string) (Provider, error) {
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

	// Try to unmarshal as new format first
	var stored StoredProvider
	if err := json.Unmarshal(data, &stored); err == nil && stored.Type != "" {
		// New format with type field
		return loadFromStored(name, stored)
	}

	// Fallback to old format (migrate automatically)
	return migrateOldFormat(name, data)
}

// loadFromStored creates a provider from the stored format
func loadFromStored(name string, stored StoredProvider) (Provider, error) {
	// Create provider instance
	prov, err := New(stored.Type)
	if err != nil {
		return nil, err
	}

	// Initialize with stored data
	if err := prov.Init(stored.Data); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	return prov, nil
}

// migrateOldFormat migrates an old provider format to new format
func migrateOldFormat(name string, data []byte) (Provider, error) {
	// Old format - unmarshal to generic map
	var oldData map[string]any
	if err := json.Unmarshal(data, &oldData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal old provider format: %w", err)
	}

	// Old format is always command type
	providerType := "command"

	// Create command provider
	prov, err := New(providerType)
	if err != nil {
		return nil, err
	}

	// Initialize with old data
	if err := prov.Init(oldData); err != nil {
		return nil, fmt.Errorf("failed to initialize migrated provider: %w", err)
	}

	// Save in new format automatically
	if err := Save(name, prov); err != nil {
		// Log warning but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to migrate provider %s to new format: %v\n", name, err)
	}

	return prov, nil
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
