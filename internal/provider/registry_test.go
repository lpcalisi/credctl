package provider

import (
	"context"
	"testing"
)

// MockProvider is a simple test provider
type MockProvider struct {
	providerType string
	schema       Schema
	metadata     map[string]any
}

func (m *MockProvider) Type() string {
	return m.providerType
}

func (m *MockProvider) Schema() Schema {
	return m.schema
}

func (m *MockProvider) Init(config map[string]any) error {
	return nil
}

func (m *MockProvider) Get(ctx context.Context) ([]byte, error) {
	return []byte("mock-output"), nil
}

func (m *MockProvider) Metadata() map[string]any {
	return m.metadata
}

// newMockFactory creates a factory function for a mock provider
func newMockFactory(providerType string) ProviderFactory {
	return func() Provider {
		return &MockProvider{
			providerType: providerType,
			schema:       Schema{},
			metadata:     make(map[string]any),
		}
	}
}

func TestRegisterAndNew(t *testing.T) {
	// Register a test provider
	Register("test-provider", newMockFactory("test-provider"))

	// Try to create a new instance
	prov, err := New("test-provider")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if prov == nil {
		t.Fatal("New() returned nil provider")
	}

	if prov.Type() != "test-provider" {
		t.Errorf("Provider.Type() = %q, want %q", prov.Type(), "test-provider")
	}
}

func TestNewUnknownProvider(t *testing.T) {
	_, err := New("unknown-provider-xyz")
	if err == nil {
		t.Error("New() expected error for unknown provider, got nil")
	}

	if err.Error() != "unknown provider type: unknown-provider-xyz" {
		t.Errorf("New() error = %q, want %q", err.Error(), "unknown provider type: unknown-provider-xyz")
	}
}

func TestListTypes(t *testing.T) {
	// Register some providers for testing
	Register("alpha-provider", newMockFactory("alpha-provider"))
	Register("beta-provider", newMockFactory("beta-provider"))

	types := ListTypes()

	// Should be sorted
	if len(types) < 2 {
		t.Fatalf("ListTypes() returned %d types, expected at least 2", len(types))
	}

	// Check that it's sorted (alpha before beta)
	foundAlpha := false
	foundBeta := false
	alphaIdx := -1
	betaIdx := -1

	for i, typ := range types {
		if typ == "alpha-provider" {
			foundAlpha = true
			alphaIdx = i
		}
		if typ == "beta-provider" {
			foundBeta = true
			betaIdx = i
		}
	}

	if !foundAlpha {
		t.Error("ListTypes() missing 'alpha-provider'")
	}
	if !foundBeta {
		t.Error("ListTypes() missing 'beta-provider'")
	}
	if foundAlpha && foundBeta && alphaIdx > betaIdx {
		t.Error("ListTypes() should return sorted list, 'alpha-provider' should come before 'beta-provider'")
	}
}

func TestIsRegistered(t *testing.T) {
	Register("registered-provider", newMockFactory("registered-provider"))

	tests := []struct {
		name         string
		providerType string
		want         bool
	}{
		{
			name:         "registered provider",
			providerType: "registered-provider",
			want:         true,
		},
		{
			name:         "unregistered provider",
			providerType: "not-registered-xyz",
			want:         false,
		},
		{
			name:         "empty string",
			providerType: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRegistered(tt.providerType)
			if got != tt.want {
				t.Errorf("IsRegistered(%q) = %v, want %v", tt.providerType, got, tt.want)
			}
		})
	}
}

func TestGetSchema(t *testing.T) {
	// Register a provider with a specific schema
	testSchema := Schema{
		Fields: []FieldDef{
			{Name: "field1", Type: FieldTypeString, Required: true},
			{Name: "field2", Type: FieldTypeBool, Required: false},
		},
	}

	Register("schema-test-provider", func() Provider {
		return &MockProvider{
			providerType: "schema-test-provider",
			schema:       testSchema,
			metadata:     make(map[string]any),
		}
	})

	schema, err := GetSchema("schema-test-provider")
	if err != nil {
		t.Fatalf("GetSchema() unexpected error: %v", err)
	}

	if len(schema.Fields) != 2 {
		t.Errorf("GetSchema() returned %d fields, want 2", len(schema.Fields))
	}

	if schema.Fields[0].Name != "field1" {
		t.Errorf("GetSchema() first field name = %q, want %q", schema.Fields[0].Name, "field1")
	}
}

func TestGetSchemaUnknownProvider(t *testing.T) {
	_, err := GetSchema("unknown-schema-provider-xyz")
	if err == nil {
		t.Error("GetSchema() expected error for unknown provider, got nil")
	}
}

func TestRegisterOverwrite(t *testing.T) {
	// Register initial provider
	Register("overwrite-test", func() Provider {
		return &MockProvider{
			providerType: "version1",
			schema:       Schema{},
			metadata:     make(map[string]any),
		}
	})

	// Overwrite with new factory
	Register("overwrite-test", func() Provider {
		return &MockProvider{
			providerType: "version2",
			schema:       Schema{},
			metadata:     make(map[string]any),
		}
	})

	// Should get the new version
	prov, err := New("overwrite-test")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if prov.Type() != "version2" {
		t.Errorf("After overwrite, Provider.Type() = %q, want %q", prov.Type(), "version2")
	}
}

func TestNewCreatesNewInstances(t *testing.T) {
	callCount := 0
	Register("instance-test", func() Provider {
		callCount++
		return &MockProvider{
			providerType: "instance-test",
			schema:       Schema{},
			metadata:     make(map[string]any),
		}
	})

	// Reset counter
	callCount = 0

	// Create multiple instances
	_, _ = New("instance-test")
	_, _ = New("instance-test")
	_, _ = New("instance-test")

	if callCount != 3 {
		t.Errorf("Factory was called %d times, want 3 (new instance each time)", callCount)
	}
}

