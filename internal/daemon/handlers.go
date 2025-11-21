package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"credctl/internal/protocol"
	"credctl/internal/provider"
	_ "credctl/internal/provider/command" // Import to register providers
)

// Add processes an "add" request
func Add(state *State, payload interface{}, readOnly bool) protocol.Response {
	// Check permissions
	if readOnly {
		return protocol.Response{
			Status: "error",
			Error:  "permission denied: add operation not allowed on read-only socket",
		}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("invalid payload: %v", err),
		}
	}

	var addPayload protocol.AddPayload
	if err := json.Unmarshal(payloadBytes, &addPayload); err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("invalid payload: %v", err),
		}
	}

	// Validate payload
	if addPayload.Name == "" {
		return protocol.Response{
			Status: "error",
			Error:  "provider name cannot be empty",
		}
	}

	if addPayload.Type == "" {
		return protocol.Response{
			Status: "error",
			Error:  "provider type cannot be empty",
		}
	}

	// Create provider instance
	prov, err := provider.New(addPayload.Type)
	if err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("failed to create provider: %v", err),
		}
	}

	// Initialize provider with metadata
	if err := prov.Init(addPayload.Metadata); err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("failed to initialize provider: %v", err),
		}
	}

	// Add provider (saves to disk and memory)
	if err := state.Add(addPayload.Name, prov); err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("failed to add provider: %v", err),
		}
	}

	return protocol.Response{
		Status: "ok",
	}
}

// Get processes a "get" request
func Get(state *State, payload interface{}, readOnly bool) protocol.Response {
	// Get operation is allowed in both modes (no permission check needed)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("invalid payload: %v", err),
		}
	}

	var getPayload protocol.GetPayload
	if err := json.Unmarshal(payloadBytes, &getPayload); err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("invalid payload: %v", err),
		}
	}

	// Get provider
	prov, err := state.Get(getPayload.Name)
	if err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("provider not found: %s", getPayload.Name),
		}
	}

	// Execute provider Get with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	output, err := prov.Get(ctx)
	if err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("failed to get credential: %v", err),
		}
	}

	// Return output and provider metadata
	return protocol.Response{
		Status: "ok",
		Payload: protocol.GetResponsePayload{
			Output:   string(output),
			Metadata: prov.Metadata(),
		},
	}
}

// Delete processes a "delete" request
func Delete(state *State, payload interface{}, readOnly bool) protocol.Response {
	// Check permissions
	if readOnly {
		return protocol.Response{
			Status: "error",
			Error:  "permission denied: delete operation not allowed on read-only socket",
		}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("invalid payload: %v", err),
		}
	}

	var deletePayload protocol.DeletePayload
	if err := json.Unmarshal(payloadBytes, &deletePayload); err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("invalid payload: %v", err),
		}
	}

	if deletePayload.Name == "" {
		return protocol.Response{
			Status: "error",
			Error:  "provider name cannot be empty",
		}
	}

	// Delete provider (from disk and memory)
	if err := state.Delete(deletePayload.Name); err != nil {
		return protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("failed to delete provider: %v", err),
		}
	}

	return protocol.Response{
		Status: "ok",
	}
}
