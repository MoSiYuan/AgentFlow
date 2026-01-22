package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPExecutor executes tasks by calling the Claude HTTP server
type HTTPExecutor struct {
	serverURL string
	client   *http.Client
	logger   *logrus.Logger
	workerID string
}

// NewHTTPExecutor creates a new HTTP executor
func NewHTTPExecutor(serverURL string, workerID string, logger *logrus.Logger) *HTTPExecutor {
	// Claude HTTP server runs on port 8849 (Master port + 1)
	if serverURL == "" {
		serverURL = "http://localhost:8849"
	}

	return &HTTPExecutor{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 120 * time.Second, // 2 minutes for story generation
		},
		logger:   logger,
		workerID: workerID,
	}
}

// CheckServerAvailable checks if the Claude server is running
func (e *HTTPExecutor) CheckServerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.serverURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ExecuteTask executes a task via HTTP call to Claude server
func (e *HTTPExecutor) ExecuteTask(ctx context.Context, task Task) (string, error) {
	startTime := time.Now()

	e.logger.WithField("task_id", task.ID).Info("Executing task via HTTP")

	// Prepare request
	reqBody := map[string]interface{}{
		"task_id":     task.ID,
		"title":       task.Title,
		"description": e.buildPrompt(task),
		"worker_id":   e.workerID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP request
	resp, err := e.client.Post(e.serverURL+"/execute", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Parse response
	var respBody struct {
		Output     string `json:"output"`
		Success    bool   `json:"success"`
		Error      string `json:"error,omitempty"`
		TokensUsed int64  `json:"tokens_used"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	duration := time.Since(startTime)
	e.logger.WithFields(logrus.Fields{
		"task_id":     task.ID,
		"duration":    duration,
		"tokens_used": respBody.TokensUsed,
	}).Info("HTTP execution completed")

	if !respBody.Success {
		return "", fmt.Errorf("execution failed: %s", respBody.Error)
	}

	return respBody.Output, nil
}

// buildPrompt builds the prompt from task description
func (e *HTTPExecutor) buildPrompt(task Task) string {
	// Simple implementation - can be enhanced
	return task.Description
}
