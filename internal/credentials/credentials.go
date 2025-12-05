package credentials

// Credentials represents credentials in structured format
// Allows providers to expose their credentials uniformly
// so they can be formatted with templates
type Credentials struct {
	Fields map[string]string
}

// New creates a new Credentials with the provided fields
func New(fields map[string]string) *Credentials {
	return &Credentials{
		Fields: fields,
	}
}

// Get retrieves the value of a field, returns empty string if it doesn't exist
func (c *Credentials) Get(key string) string {
	if c.Fields == nil {
		return ""
	}
	return c.Fields[key]
}

// Set sets the value of a field
func (c *Credentials) Set(key, value string) {
	if c.Fields == nil {
		c.Fields = make(map[string]string)
	}
	c.Fields[key] = value
}

// Has checks if a field exists
func (c *Credentials) Has(key string) bool {
	if c.Fields == nil {
		return false
	}
	_, exists := c.Fields[key]
	return exists
}
