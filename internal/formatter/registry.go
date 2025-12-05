package formatter

import "fmt"

// Formatter is the interface that all output formatters must implement
type Formatter interface {
	// Name returns the formatter name (e.g., "json", "text", "escaped")
	Name() string

	// Format formats the output bytes according to the formatter's rules
	Format(output []byte) ([]byte, error)
}

// FormatterFactory is a function that creates a new formatter instance
type FormatterFactory func() Formatter

var (
	formatters = make(map[string]FormatterFactory)
)

// Register registers a new formatter factory
func RegisterFormatter(name string, factory FormatterFactory) {
	formatters[name] = factory
}

// Get returns a formatter instance by name
func Get(name string) (Formatter, error) {
	factory, ok := formatters[name]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", name)
	}
	return factory(), nil
}

// List returns a list of registered formatter names
func List() []string {
	names := make([]string, 0, len(formatters))
	for name := range formatters {
		names = append(names, name)
	}
	return names
}
