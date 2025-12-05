package formatter

import (
	"fmt"
	"strconv"
)

// EscapedFormatter interprets escape sequences like \n, \t, \\, etc.
type EscapedFormatter struct{}

func init() {
	RegisterFormatter("escaped", func() Formatter {
		return &EscapedFormatter{}
	})
}

func (f *EscapedFormatter) Name() string {
	return "escaped"
}

func (f *EscapedFormatter) Format(output []byte) ([]byte, error) {
	unescaped, err := unescapeString(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to unescape output: %w", err)
	}
	return []byte(unescaped), nil
}

// unescapeString interprets escape sequences in a string
// Handles: \n, \t, \r, \\, and other standard Go escape sequences
func unescapeString(s string) (string, error) {
	// Wrap the string in quotes so strconv.Unquote can process it
	// We add quotes directly without using Quote() to avoid double-escaping
	quoted := `"` + s + `"`
	// Unquote will interpret the escape sequences
	unquoted, err := strconv.Unquote(quoted)
	if err != nil {
		return "", err
	}
	return unquoted, nil
}
