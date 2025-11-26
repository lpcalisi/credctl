package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"credctl/internal/protocol"
)

// ResolveSocketPath returns the Unix socket path
// Priority order:
// 1. CREDCTL_SOCK env var (if set)
// 2. Check if admin socket exists (~/.credctl/agent.sock) - assumes write access
// 3. Check if read-only socket exists (~/.credctl/agent-readonly.sock)
// 4. Error if no socket found
func ResolveSocketPath() (string, error) {
	// Check env var first
	if sockPath := os.Getenv("CREDCTL_SOCK"); sockPath != "" {
		return sockPath, nil
	}

	// Get home directory for default paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check if admin socket exists (assumes write access)
	adminSocketPath := filepath.Join(homeDir, ".credctl", "agent.sock")
	if _, err := os.Stat(adminSocketPath); err == nil {
		return adminSocketPath, nil
	}

	// Check if read-only socket exists
	readOnlySocketPath := filepath.Join(homeDir, ".credctl", "agent-readonly.sock")
	if _, err := os.Stat(readOnlySocketPath); err == nil {
		return readOnlySocketPath, nil
	}

	// No socket found
	return "", fmt.Errorf("no credctl socket found (is the daemon running?)")
}

// SendRequest sends a request to the daemon and returns the response
func SendRequest(req protocol.Request) (protocol.Response, error) {
	socketPath, err := ResolveSocketPath()
	if err != nil {
		return protocol.Response{}, err
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to connect to daemon: %w (is the daemon running?)", err)
	}
	defer func() { _ = conn.Close() }()

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := conn.Write(append(reqJSON, '\n')); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to send request: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return protocol.Response{}, fmt.Errorf("failed to read response: %w", err)
		}
		return protocol.Response{}, fmt.Errorf("no response received")
	}

	var resp protocol.Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp, nil
}
