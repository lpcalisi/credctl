package formatter

import (
	"encoding/json"
	"fmt"
)

// JSONFormatter validates that the output is valid JSON
type JSONFormatter struct{}

func init() {
	RegisterFormatter("json", func() Formatter {
		return &JSONFormatter{}
	})
}

func (f *JSONFormatter) Name() string {
	return "json"
}

func (f *JSONFormatter) Format(output []byte) ([]byte, error) {
	if !json.Valid(output) {
		return nil, fmt.Errorf("output is not valid JSON")
	}
	return output, nil
}
