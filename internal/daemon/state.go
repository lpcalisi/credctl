package daemon

import (
	"fmt"
	"sync"

	"credctl/internal/provider"
)

// State represents the daemon's in-memory state
type State struct {
	providers map[string]provider.Provider
	mu        sync.RWMutex
}

// NewState creates a new daemon state and loads providers from disk
func NewState() (*State, error) {
	s := &State{
		providers: make(map[string]provider.Provider),
	}

	// Load all providers from disk
	if err := s.LoadAll(); err != nil {
		return nil, err
	}

	return s, nil
}

// LoadAll loads all providers from disk into memory
func (s *State) LoadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	names, err := provider.List()
	if err != nil {
		return err
	}

	for _, name := range names {
		prov, err := provider.Load(name)
		if err != nil {
			// Log error but continue loading other providers
			continue
		}
		s.providers[name] = prov
	}

	return nil
}

// Add adds a provider to memory and persists it to disk
func (s *State) Add(name string, prov provider.Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	exists, err := provider.Exists(name)
	if err != nil {
		return fmt.Errorf("failed to check provider existence: %w", err)
	}
	if exists {
		return fmt.Errorf("provider '%s' already exists", name)
	}

	// Save to disk first
	if err := provider.Save(name, prov); err != nil {
		return err
	}

	// Then update memory
	s.providers[name] = prov
	return nil
}

// Get retrieves a provider from memory
func (s *State) Get(name string) (provider.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prov, exists := s.providers[name]
	if !exists {
		// Try loading from disk as fallback
		diskProv, err := provider.Load(name)
		if err != nil {
			return nil, err
		}
		return diskProv, nil
	}

	return prov, nil
}

// Delete removes a provider from memory and disk
func (s *State) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete from disk first
	if err := provider.Delete(name); err != nil {
		return err
	}

	// Then remove from memory
	delete(s.providers, name)
	return nil
}
