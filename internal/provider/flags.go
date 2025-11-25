package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// AddSchemaFlags adds flags to a cobra command based on the provider schema
// Panics if there's a flag collision between providers
func AddSchemaFlags(cmd *cobra.Command, schema Schema) {
	for _, field := range schema.Fields {
		flagName := field.Name

		// Check if flag already exists (collision detection)
		if cmd.Flags().Lookup(flagName) != nil {
			panic(fmt.Sprintf("flag collision detected: --%s is defined by multiple providers", flagName))
		}

		switch field.Type {
		case FieldTypeString:
			defaultVal := field.Default
			cmd.Flags().String(flagName, defaultVal, field.Help)

		case FieldTypeBool:
			defaultVal := field.Default == "true"
			cmd.Flags().Bool(flagName, defaultVal, field.Help)

		case FieldTypeInt:
			defaultVal := 0
			if field.Default != "" {
				if i, err := strconv.Atoi(field.Default); err == nil {
					defaultVal = i
				}
			}
			cmd.Flags().Int(flagName, defaultVal, field.Help)

		case FieldTypeStringSlice:
			var defaultVal []string
			if field.Default != "" {
				defaultVal = strings.Split(field.Default, ",")
			}
			cmd.Flags().StringSlice(flagName, defaultVal, field.Help)
		}

		// Mark as required if needed
		if field.Required {
			_ = cmd.MarkFlagRequired(flagName)
		}
	}
}

// ExtractConfig extracts configuration values from cobra command flags
func ExtractConfig(cmd *cobra.Command, schema Schema) (map[string]any, error) {
	config := make(map[string]any)

	for _, field := range schema.Fields {
		flagName := field.Name

		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			// Flag not defined, skip
			continue
		}

		// If flag wasn't changed and it's not required, use default
		if !flag.Changed && !field.Required {
			if field.Default != "" {
				switch field.Type {
				case FieldTypeString:
					config[field.Name] = field.Default
				case FieldTypeBool:
					config[field.Name] = field.Default == "true"
				case FieldTypeInt:
					if i, err := strconv.Atoi(field.Default); err == nil {
						config[field.Name] = i
					}
				case FieldTypeStringSlice:
					config[field.Name] = strings.Split(field.Default, ",")
				}
			}
			continue
		}

		// Extract value based on type
		var err error
		switch field.Type {
		case FieldTypeString:
			var val string
			val, err = cmd.Flags().GetString(flagName)
			if err == nil && val != "" {
				config[field.Name] = val
			}

		case FieldTypeBool:
			var val bool
			val, err = cmd.Flags().GetBool(flagName)
			if err == nil {
				config[field.Name] = val
			}

		case FieldTypeInt:
			var val int
			val, err = cmd.Flags().GetInt(flagName)
			if err == nil {
				config[field.Name] = val
			}

		case FieldTypeStringSlice:
			var val []string
			val, err = cmd.Flags().GetStringSlice(flagName)
			if err == nil && len(val) > 0 {
				config[field.Name] = val
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get flag %s: %w", flagName, err)
		}
	}

	// Validate the extracted config
	if err := ValidateConfig(config, schema); err != nil {
		return nil, err
	}

	return config, nil
}

// AddAllProviderFlags adds flags for all registered provider types
// Panics if there's a flag collision between providers
func AddAllProviderFlags(cmd *cobra.Command) {
	for _, provType := range ListTypes() {
		schema, err := GetSchema(provType)
		if err != nil {
			continue
		}
		// Add flags without prefix - will panic on collision
		AddSchemaFlags(cmd, schema)
	}
}
