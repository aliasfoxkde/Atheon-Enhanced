package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// DaemonConfig holds daemon configuration
type DaemonConfig struct {
	SocketPath string
}

// ScanRequest represents a scan request
type ScanRequest struct {
	Path string `json:"path"`
}

// ScanResponse represents a scan response
type ScanResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Daemon is the Atheon daemon service
type Daemon struct {
	config DaemonConfig
	mu     sync.Mutex
}

// NewDaemon creates a new daemon instance
func NewDaemon(config DaemonConfig) *Daemon {
	return &Daemon{
		config: config,
	}
}

// Start begins the daemon listening loop
func (d *Daemon) Start(ctx context.Context) error {
	// Remove existing socket file
	_ = os.Remove(d.config.SocketPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(d.config.SocketPath), 0750); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Create Unix socket
	ln, err := net.Listen("unix", d.config.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	defer func() { _ = os.Remove(d.config.SocketPath) }()

	// Set socket permissions
	if err := os.Chmod(d.config.SocketPath, 0660); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	fmt.Printf("Atheon daemon listening on %s\n", d.config.SocketPath)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down daemon...")
		os.Exit(0)
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go d.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (d *Daemon) handleConnection(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	// Read request
	data, err := io.ReadAll(conn)
	if err != nil {
		d.sendError(conn, fmt.Sprintf("read error: %v", err))
		return
	}

	var req ScanRequest
	if err := json.Unmarshal(data, &req); err != nil {
		d.sendError(conn, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Validate path
	if req.Path == "" {
		d.sendError(conn, "path is required")
		return
	}

	// Process scan
	response := ScanResponse{
		Success: true,
		Message: fmt.Sprintf("Scan queued for %s", req.Path),
	}

	// Send response
	respData, err := json.Marshal(response)
	if err != nil {
		d.sendError(conn, fmt.Sprintf("marshal error: %v", err))
		return
	}

	if _, err := conn.Write(respData); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
	}
}

// sendError sends an error response
func (d *Daemon) sendError(conn net.Conn, msg string) {
	resp := ScanResponse{Success: false, Error: msg}
	data, _ := json.Marshal(resp)
	if _, err := conn.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "error response write failed: %v\n", err)
	}
}

// Status returns daemon status
func (d *Daemon) Status() map[string]interface{} {
	return map[string]interface{}{
		"status":    "running",
		"socket":    d.config.SocketPath,
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Atheon Daemon")
		fmt.Println("\nUsage: atheon-daemon [socket-path]")
		fmt.Println("\nDefault socket: /var/run/atheon/atheon.sock")
		fmt.Println("\nStarts a background service for scanning files.")
		fmt.Println("Uses Unix socket for IPC with clients.")
		os.Exit(0)
	}

	socketPath := "/var/run/atheon/atheon.sock"
	if len(os.Args) > 1 {
		socketPath = os.Args[1]
	}

	config := DaemonConfig{
		SocketPath: socketPath,
	}

	daemon := NewDaemon(config)
	ctx := context.Background()

	if err := daemon.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
		os.Exit(1)
	}
}
