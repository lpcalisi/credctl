package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"credctl/internal/daemon"
	"credctl/internal/protocol"
)

// Add processes an "add" request
func Add(state *daemon.State, payload interface{}, readOnly bool) protocol.Response {
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

	prov := &addPayload.Provider

	// Validate provider
	if prov.Name == "" {
		return protocol.Response{
			Status: "error",
			Error:  "provider name cannot be empty",
		}
	}

	if prov.Command == "" {
		return protocol.Response{
			Status: "error",
			Error:  "provider command cannot be empty",
		}
	}

	// Add provider (saves to disk and memory)
	if err := state.Add(prov); err != nil {
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
func Get(state *daemon.State, payload interface{}, readOnly bool) protocol.Response {
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

	// Execute command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", prov.Command)

	// Capture stdout and stderr separately
	stdout, err := cmd.Output()
	if err != nil {
		// Command failed, capture stderr for context
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(exitErr.Stderr))
			if len(stderr) > 500 {
				stderr = stderr[:500] + "..."
			}
		}

		errorMsg := fmt.Sprintf("command failed: %v", err)
		if stderr != "" {
			errorMsg = fmt.Sprintf("command failed: %v: %s", err, stderr)
		}

		return protocol.Response{
			Status: "error",
			Error:  errorMsg,
		}
	}

	// Success: use only stdout, ignore any stderr
	// Return raw output and provider metadata - let the client format it
	rawOutput := strings.TrimRight(string(stdout), "\r\n")

	return protocol.Response{
		Status: "ok",
		Payload: protocol.GetResponsePayload{
			Output:   rawOutput,
			Provider: prov,
		},
	}
}

// Delete processes a "delete" request
func Delete(state *daemon.State, payload interface{}, readOnly bool) protocol.Response {
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
