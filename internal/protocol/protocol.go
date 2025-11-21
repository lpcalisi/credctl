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
}

// GetPayload is the payload for the "get" action
type GetPayload struct {
	Name string `json:"name"`
}

// DeletePayload is the payload for the "delete" action
type DeletePayload struct {
	Name string `json:"name"`
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
