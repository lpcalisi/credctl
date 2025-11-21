package provider

import (
	"fmt"
	"sort"
	"sync"
)

// ProviderFactory is a function that creates a new Provider instance
type ProviderFactory func() Provider

var (
	registry = make(map[string]ProviderFactory)
	mu       sync.RWMutex
)

// Register registers a provider type with its factory function
func Register(providerType string, factory ProviderFactory) {
	mu.Lock()
	defer mu.Unlock()
	registry[providerType] = factory
}

// New creates a new provider instance of the specified type
func New(providerType string) (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()

	factory, exists := registry[providerType]
	if !exists {
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}

	return factory(), nil
}

// ListTypes returns a sorted list of all registered provider types
func ListTypes() []string {
	mu.RLock()
	defer mu.RUnlock()

	types := make([]string, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// GetSchema returns the schema for a specific provider type
func GetSchema(providerType string) (Schema, error) {
	prov, err := New(providerType)
	if err != nil {
		return Schema{}, err
	}
	return prov.Schema(), nil
}

// IsRegistered checks if a provider type is registered
func IsRegistered(providerType string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, exists := registry[providerType]
	return exists
}
