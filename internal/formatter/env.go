package formatter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Env formats the output as shell export statements using heuristics
// 1. If output is valid JSON with keys: export each key
// 2. If output contains "export KEY=VALUE" lines: pass as-is
// 3. If output contains "KEY=VALUE" lines: add "export" prefix
// 4. If output is a raw token: use envVar parameter
func Env(output string, envVar string) (string, error) {
	output = strings.TrimSpace(output)

	// Heuristic 1: Try to parse as JSON
	if strings.HasPrefix(output, "{") {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(output), &data); err == nil {
			return jsonToExports(data), nil
		}
	}

	// Heuristic 2: Check if already has export statements
	exportRegex := regexp.MustCompile(`(?m)^export\s+[A-Z_][A-Z0-9_]*=`)
	if exportRegex.MatchString(output) {
		return output, nil
	}

	// Heuristic 3: Check for KEY=VALUE lines
	keyValueRegex := regexp.MustCompile(`(?m)^[A-Z_][A-Z0-9_]*=`)
	if keyValueRegex.MatchString(output) {
		return addExportPrefix(output), nil
	}

	// Heuristic 4: Raw token - requires envVar
	if envVar == "" {
		return "", fmt.Errorf("output appears to be a raw token, please specify --env-var")
	}

	// Validate envVar name
	if !regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`).MatchString(envVar) {
		return "", fmt.Errorf("invalid env var name: %s", envVar)
	}

	quoted, err := syntax.Quote(output, syntax.LangBash)
	if err != nil {
		return "", fmt.Errorf("failed to quote value: %w", err)
	}
	return fmt.Sprintf("export %s=%s", envVar, quoted), nil
}

// jsonToExports converts a JSON object to export statements
func jsonToExports(data map[string]interface{}) string {
	var lines []string
	for key, value := range data {
		// Convert value to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case nil:
			strValue = ""
		default:
			// For complex types, marshal back to JSON
			bytes, _ := json.Marshal(v)
			strValue = string(bytes)
		}
		quoted, err := syntax.Quote(strValue, syntax.LangBash)
		if err != nil {
			// Fallback to simple quoting if syntax.Quote fails
			quoted = fmt.Sprintf("'%s'", strings.ReplaceAll(strValue, "'", "'\"'\"'"))
		}
		lines = append(lines, fmt.Sprintf("export %s=%s", key, quoted))
	}
	return strings.Join(lines, "\n")
}

// addExportPrefix adds "export " to lines that look like KEY=VALUE
func addExportPrefix(output string) string {
	lines := strings.Split(output, "\n")
	var result []string
	keyValueRegex := regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)=(.*)$`)

	for _, line := range lines {
		if keyValueRegex.MatchString(line) {
			result = append(result, "export "+line)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
