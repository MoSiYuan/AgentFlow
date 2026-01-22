package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ClaudeServer manages Claude execution as a standalone HTTP server
type ClaudeServer struct {
	port   int
	logger *logrus.Logger
	server *http.Server
	mu     sync.Mutex
}

// NewClaudeServer creates a new Claude server
func NewClaudeServer(port int, logger *logrus.Logger) *ClaudeServer {
	cs := &ClaudeServer{
		port:   port,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/execute", cs.handleExecute)

	cs.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	return cs
}

// Start starts the Claude server
func (s *ClaudeServer) Start() error {
	s.logger.Infof("Starting Claude server on port %d", s.port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *ClaudeServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down Claude server...")
	return s.server.Shutdown(ctx)
}

// Request/Response structures
type ExecuteRequest struct {
	TaskID      string `json:"task_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	WorkerID    string `json:"worker_id"`
}

type ExecuteResponse struct {
	Output     string `json:"output"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	TokensUsed int64  `json:"tokens_used"`
}

// handleExecute handles task execution requests
func (s *ClaudeServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.WithError(err).Error("Failed to decode request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.logger.WithFields(logrus.Fields{
		"task_id":  req.TaskID,
		"title":    req.Title,
		"worker_id": req.WorkerID,
	}).Info("Executing task via Claude server")

	// Execute task
	output, err := s.executeTask(req.Description)

	resp := ExecuteResponse{
		Output:     output,
		Success:    err == nil,
		TokensUsed: 0, // TODO: Implement token counting
	}

	if err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// executeTask executes a task using Claude CLI
func (s *ClaudeServer) executeTask(description string) (string, error) {
	s.logger.Info("Executing task: ", description)

	// Check for claudecli wrapper
	claudecliPath := "/home/jiangxiaolong/bin/claudecli"
	// Fallback to PATH
	if _, err := exec.LookPath("claudecli"); err == nil {
		// found in PATH
	}

	// TODO: Implement actual Claude CLI execution
	// For now, return a simulated response
	return fmt.Sprintf("[Claude Server] Executed: %s", description), nil
}
