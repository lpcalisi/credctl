package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"credctl/internal/protocol"

	"github.com/sevlyar/go-daemon"
)

func handleConn(conn net.Conn, state *State, readOnly bool) {
	defer func() { _ = conn.Close() }()

	socketType := "admin"
	if readOnly {
		socketType = "readonly"
	}

	log.Printf("[%s] new connection", socketType)

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("[%s] error reading from connection: %v", socketType, err)
		} else {
			log.Printf("[%s] connection closed by client (no data)", socketType)
		}
		return
	}

	rawRequest := scanner.Bytes()
	log.Printf("[%s] received %d bytes", socketType, len(rawRequest))

	var req protocol.Request
	if err := json.Unmarshal(rawRequest, &req); err != nil {
		log.Printf("[%s] error parsing request: %v", socketType, err)
		resp := protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("invalid JSON: %v", err),
		}
		respJSON, _ := json.Marshal(resp)
		_, _ = conn.Write(append(respJSON, '\n'))
		return
	}

	log.Printf("[%s] processing action=%s", socketType, req.Action)

	var resp protocol.Response
	switch req.Action {
	case "add":
		resp = Add(state, req.Payload, readOnly)
	case "get":
		resp = Get(state, req.Payload, readOnly)
	case "delete":
		resp = Delete(state, req.Payload, readOnly)
	case "set_tokens":
		resp = SetTokens(state, req.Payload, readOnly)
	default:
		resp = protocol.Response{
			Status: "error",
			Error:  fmt.Sprintf("unknown action: %s", req.Action),
		}
	}

	log.Printf("[%s] action=%s status=%s", socketType, req.Action, resp.Status)

	respJSON, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[%s] error marshaling response: %v", socketType, err)
		return
	}

	log.Printf("[%s] sending %d bytes response", socketType, len(respJSON))

	if _, err := conn.Write(append(respJSON, '\n')); err != nil {
		log.Printf("[%s] error writing response: %v", socketType, err)
	} else {
		log.Printf("[%s] response sent successfully", socketType)
	}
}

// DaemonInfo contains information about the running daemon
type DaemonInfo struct {
	AdminSocket    string
	ReadOnlySocket string
	PID            int
	LogFile        string
}

// Run starts the daemon and returns daemon info for the parent process
func Run() (*DaemonInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Define socket paths
	adminSocketPath := filepath.Join(homeDir, ".credctl", "agent.sock")
	readOnlySocketPath := filepath.Join(homeDir, ".credctl", "agent-readonly.sock")
	pidFile := filepath.Join(homeDir, ".credctl", "credctl.pid")
	logFile := filepath.Join(homeDir, ".credctl", "daemon.log")

	// Create directory for sockets and PID file
	dir := filepath.Join(homeDir, ".credctl")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Setup daemon context
	cntxt := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0644,
		LogFileName: logFile,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}

	// Daemonize
	d, err := cntxt.Reborn()
	if err != nil {
		return nil, fmt.Errorf("failed to daemonize: %w", err)
	}
	if d != nil {
		// Parent process - return daemon info
		return &DaemonInfo{
			AdminSocket:    adminSocketPath,
			ReadOnlySocket: readOnlySocketPath,
			PID:            d.Pid,
			LogFile:        logFile,
		}, nil
	}
	defer func() { _ = cntxt.Release() }()

	// Child process continues here
	log.Printf("daemon started with PID %d", os.Getpid())

	// Check if admin socket exists and if daemon is running
	if _, err := os.Stat(adminSocketPath); err == nil {
		// Socket file exists, try to connect
		testConn, err := net.Dial("unix", adminSocketPath)
		if err == nil {
			_ = testConn.Close()
			log.Printf("daemon already running")
			return nil, nil
		}
		// Socket exists but can't connect, remove it
		_ = os.Remove(adminSocketPath)
	}

	// Remove read-only socket if it exists
	_ = os.Remove(readOnlySocketPath)

	// Create admin listener
	adminListener, err := net.Listen("unix", adminSocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin listener: %w", err)
	}
	defer func() { _ = adminListener.Close() }()

	// Set admin socket permissions (only owner)
	if err := os.Chmod(adminSocketPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set admin socket permissions: %w", err)
	}

	// Create read-only listener
	readOnlyListener, err := net.Listen("unix", readOnlySocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create read-only listener: %w", err)
	}
	defer func() { _ = readOnlyListener.Close() }()

	// Set read-only socket permissions (owner and group can access)
	if err := os.Chmod(readOnlySocketPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set read-only socket permissions: %w", err)
	}

	// Load state from disk
	state, err := NewState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	log.Printf("listening on admin socket: %s", adminSocketPath)
	log.Printf("listening on read-only socket: %s", readOnlySocketPath)

	// Setup signal handler for cleanup
	cleanup := func() {
		_ = os.Remove(adminSocketPath)
		_ = os.Remove(readOnlySocketPath)
	}
	daemon.SetSigHandler(termHandler(cleanup), syscall.SIGTERM)
	daemon.SetSigHandler(termHandler(cleanup), syscall.SIGINT)

	// Accept connections from admin socket
	go func() {
		for {
			conn, err := adminListener.Accept()
			if err != nil {
				log.Printf("error accepting admin connection: %v", err)
				continue
			}
			go handleConn(conn, state, false) // false = not read-only
		}
	}()

	// Accept connections from read-only socket (main goroutine)
	for {
		conn, err := readOnlyListener.Accept()
		if err != nil {
			log.Printf("error accepting read-only connection: %v", err)
			continue
		}
		go handleConn(conn, state, true) // true = read-only
	}

	// This line is never reached in normal operation
	// return nil, nil
}

// termHandler returns a signal handler that cleans up and exits
func termHandler(cleanup func()) daemon.SignalHandlerFunc {
	return func(sig os.Signal) error {
		log.Printf("received signal %v, shutting down", sig)
		cleanup()
		return daemon.ErrStop
	}
}
