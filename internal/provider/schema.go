package provider

import "fmt"

// FieldType represents the type of a configuration field
type FieldType string

const (
	FieldTypeString      FieldType = "string"
	FieldTypeBool        FieldType = "bool"
	FieldTypeInt         FieldType = "int"
	FieldTypeStringSlice FieldType = "[]string"
)

// FieldDef defines a configuration field for a provider
type FieldDef struct {
	Name        string
	Type        FieldType
	Required    bool
	Default     string
	Help        string
	Hidden      bool     // For secrets/sensitive values
	ValidValues []string // For enum-like fields
}

// Schema defines the configuration schema for a provider
type Schema struct {
	Fields []FieldDef
}

// Helper functions for extracting typed values from config maps

func GetStringOrDefault(config map[string]any, key, defaultValue string) string {
	if val, ok := config[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func GetBoolOrDefault(config map[string]any, key string, defaultValue bool) bool {
	if val, ok := config[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func GetIntOrDefault(config map[string]any, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		// Handle float64 from JSON unmarshaling
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}

func GetStringSliceOrDefault(config map[string]any, key string, defaultValue []string) []string {
	if val, ok := config[key]; ok {
		if slice, ok := val.([]string); ok {
			return slice
		}
		// Handle []interface{} from JSON unmarshaling
		if iSlice, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(iSlice))
			for _, item := range iSlice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return defaultValue
}

// ValidateConfig validates a config map against a schema
func ValidateConfig(config map[string]any, schema Schema) error {
	// Check required fields
	for _, field := range schema.Fields {
		if field.Required {
			if _, ok := config[field.Name]; !ok {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}

		// Validate enum values
		if len(field.ValidValues) > 0 {
			if val, ok := config[field.Name]; ok {
				strVal, ok := val.(string)
				if !ok {
					return fmt.Errorf("field '%s' must be a string", field.Name)
				}
				valid := false
				for _, validVal := range field.ValidValues {
					if strVal == validVal {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("field '%s' must be one of: %v", field.Name, field.ValidValues)
				}
			}
		}
	}

	return nil
}
