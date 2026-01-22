package worker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ClaudeExecutor executes tasks using claudecli command
type ClaudeExecutor struct {
	cliPath string
	logger  *logrus.Logger
	enabled bool
}

// NewClaudeExecutor creates a new Claude CLI executor
func NewClaudeExecutor(logger *logrus.Logger) *ClaudeExecutor {
	executor := &ClaudeExecutor{
		logger: logger,
	}

	// Check if claudecli is available
	if path, err := exec.LookPath("claudecli"); err == nil {
		executor.enabled = true
		executor.cliPath = path
		logger.Infof("Claude CLI found: %s", path)
	} else {
		executor.enabled = false
		logger.Warn("Claude CLI not found in PATH")
	}

	return executor
}

// IsEnabled returns true if claudecli is available
func (e *ClaudeExecutor) IsEnabled() bool {
	return e.enabled
}

// ExecuteTask executes a task using claudecli
func (e *ClaudeExecutor) ExecuteTask(ctx context.Context, task Task) (string, error) {
	if !e.enabled {
		return "", fmt.Errorf("claudecli not available")
	}

	startTime := time.Now()
	e.logger.WithField("task_id", task.ID).Info("Executing task via Claude CLI")

	// Build claudecli command
	args := []string{
		"chat",
		"--prompt", task.Description,
		"--no-interactive",
	}

	// Create command
	cmd := exec.CommandContext(ctx, e.cliPath, args...)

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claudecli execution failed: %w: %s", err, string(output))
	}

	duration := time.Since(startTime)
	e.logger.WithFields(logrus.Fields{
		"task_id":  task.ID,
		"duration": duration,
	}).Info("Claude CLI execution completed")

	return string(output), nil
}

// ValidateCLI validates that claudecli is properly configured
func (e *ClaudeExecutor) ValidateCLI() error {
	if !e.enabled {
		return fmt.Errorf("claudecli not found in PATH")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.cliPath, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("claudecli validation failed: %w\nOutput: %s", err, string(output))
	}

	e.logger.Infof("claudecli version: %s", strings.TrimSpace(string(output)))
	return nil
}
