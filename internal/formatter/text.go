package formatter

// TextFormatter performs no transformation on the output
type TextFormatter struct{}

func init() {
	RegisterFormatter("text", func() Formatter {
		return &TextFormatter{}
	})
}

func (f *TextFormatter) Name() string {
	return "text"
}

func (f *TextFormatter) Format(output []byte) ([]byte, error) {
	// No transformation needed for text format
	return output, nil
}
