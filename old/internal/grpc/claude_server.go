package grpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// ClaudeServer manages Claude execution as a simple HTTP server
type ClaudeServer struct {
	port   int
	logger *logrus.Logger
	mu     sync.Mutex
}

// NewClaudeServer creates a new Claude server
func NewClaudeServer(port int, logger *logrus.Logger) *ClaudeServer {
	return &ClaudeServer{
		port:   port,
		logger: logger,
	}
}

// Start starts the HTTP server
func (s *ClaudeServer) Start() error {
	s.logger.Infof("Starting Claude server on port %d", s.port)

	http.HandleFunc("/execute", s.handleExecute)

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

// Request/Response structures for JSON
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

func (s *ClaudeServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received execution request")

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Errorf("Failed to decode request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.logger.Infof("Executing task %s for worker %s", req.TaskID, req.WorkerID)

	// Call claude CLI
	output, err := s.callClaude(req.Description)
	if err != nil {
		s.logger.Errorf("Claude execution failed: %v", err)
		json.NewEncoder(w).Encode(ExecuteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Send response
	json.NewEncoder(w).Encode(ExecuteResponse{
		Output:     output,
		Success:    true,
		TokensUsed: estimateTokens(output),
	})

	s.logger.Infof("Task %s completed", req.TaskID)
}

func (s *ClaudeServer) callClaude(prompt string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Calling claude CLI...")

	// Check if claude CLI is available
	_, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude CLI not found: %w", err)
	}

	// Prepare command
	cmd := exec.Command("claude", "-p")
	cmd.Stdin = strings.NewReader(prompt)

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude execution failed: %w\nOutput: %s", err, string(output))
	}

	result := string(output)
	s.logger.Infof("Claude response received (%d bytes)", len(result))

	return result, nil
}

func estimateTokens(text string) int64 {
	return int64(len(text) / 2)
}
