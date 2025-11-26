package protocol

// Request represents a request to the daemon
type Request struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

// AddPayload is the payload for the "add" action
type AddPayload struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata"`
	Force    bool           `json:"force,omitempty"`
}

// GetPayload is the payload for the "get" action
type GetPayload struct {
	Name string `json:"name"`
}

// DeletePayload is the payload for the "delete" action
type DeletePayload struct {
	Name string `json:"name"`
}

// SetTokensPayload is the payload for the "set_tokens" action
type SetTokensPayload struct {
	Name         string `json:"name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"` // seconds until expiration
}

// Response represents a response from the daemon
type Response struct {
	Status  string      `json:"status"`
	Error   string      `json:"error,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// GetResponsePayload is the payload of response for "get"
type GetResponsePayload struct {
	Output   string         `json:"output"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// DescribePayload is the payload for the "describe" action
type DescribePayload struct {
	Name string `json:"name"`
}

// DescribeResponsePayload is the payload of response for "describe"
type DescribeResponsePayload struct {
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata"`
}
