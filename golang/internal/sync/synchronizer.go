package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
	"github.com/sirupsen/logrus"
)

// SyncDirection represents the synchronization direction
type SyncDirection string

const (
	SyncToClaude      SyncDirection = "to_claude"       // AgentFlow → Claude
	SyncFromClaude    SyncDirection = "from_claude"     // Claude → AgentFlow
	SyncBidirectional SyncDirection = "bidirectional" // 双向同步
)

// ClaudeSessionFile represents the Claude session file structure
type ClaudeSessionFile struct {
	Path      string
	SessionUUID string
}

// ClaudeStatusUpdate represents a status update in Claude session file
type ClaudeStatusUpdate struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Data      ClaudeStatusUpdateData `json:"data"`
}

// ClaudeStatusUpdateData contains the status update details
type ClaudeStatusUpdateData struct {
	TaskID      string `json:"task_id"`
	MessageUUID string `json:"message_uuid"`
	Status      string `json:"status"`
	Result      string `json:"result,omitempty"`
	Error       string `json:"error,omitempty"`
}

// TaskSynchronizer handles synchronization between AgentFlow and Claude
type TaskSynchronizer struct {
	db            *database.Database
	claudeDir     string
	logger        *logrus.Logger
	autoSync      bool
	syncInterval  time.Duration
}

// NewTaskSynchronizer creates a new task synchronizer
func NewTaskSynchronizer(db *database.Database, claudeDir string, logger *logrus.Logger) *TaskSynchronizer {
	if logger == nil {
		logger = logrus.New()
		logger.SetOutput(os.Stdout)
		logger.SetLevel(logrus.InfoLevel)
	}

	return &TaskSynchronizer{
		db:           db,
		claudeDir:    claudeDir,
		logger:       logger,
		autoSync:     true,
		syncInterval: 5 * time.Second,
	}
}

// SetAutoSync enables or disables automatic synchronization
func (ts *TaskSynchronizer) SetAutoSync(enabled bool) {
	ts.autoSync = enabled
}

// SetSyncInterval sets the synchronization interval
func (ts *TaskSynchronizer) SetSyncInterval(interval time.Duration) {
	ts.syncInterval = interval
}

// SyncTaskStatus synchronizes task status between AgentFlow and Claude
func (ts *TaskSynchronizer) SyncTaskStatus(
	ctx context.Context,
	taskID string,
	direction SyncDirection,
) error {
	// Get task from database
	task, err := ts.db.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get Claude mapping
	mapping, err := ts.db.GetClaudeMappingByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get Claude mapping: %w", err)
	}

	if mapping == nil {
		ts.logger.Warnf("No Claude mapping found for task %s", taskID)
		return nil
	}

	switch direction {
	case SyncToClaude:
		return ts.syncToClaude(ctx, task, mapping)
	case SyncFromClaude:
		return ts.syncFromClaude(ctx, mapping)
	case SyncBidirectional:
		// First sync to Claude, then from Claude
		if err := ts.syncToClaude(ctx, task, mapping); err != nil {
			return err
		}
		return ts.syncFromClaude(ctx, mapping)
	default:
		return fmt.Errorf("invalid sync direction: %s", direction)
	}
}

// syncToClaude writes AgentFlow task status to Claude session file
func (ts *TaskSynchronizer) syncToClaude(
	ctx context.Context,
	task *database.Task,
	mapping *database.ClaudeMapping,
) error {
	// Build session file path
	sessionFile := ts.getClaudeSessionFile(mapping.SessionUUID)

	// Create status update
	update := ClaudeStatusUpdate{
		Type:      "agentflow_status_update",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: ClaudeStatusUpdateData{
			TaskID:      mapping.TaskID,
			MessageUUID: mapping.MessageUUID,
			Status:      task.Status,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal status update: %w", err)
	}

	// Append to session file (JSON Lines format)
	if err := ts.appendToFile(sessionFile, data); err != nil {
		return fmt.Errorf("failed to write to session file: %w", err)
	}

	ts.logger.Infof("Synced task %s status to Claude: %s", task.ID, task.Status)
	return nil
}

// syncFromClaude reads status updates from Claude session file
func (ts *TaskSynchronizer) syncFromClaude(
	ctx context.Context,
	mapping *database.ClaudeMapping,
) error {
	// Build session file path
	sessionFile := ts.getClaudeSessionFile(mapping.SessionUUID)

	// Check if file exists
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		ts.logger.Warnf("Claude session file not found: %s", sessionFile)
		return nil
	}

	// Read and parse status updates
	updates, err := ts.readStatusUpdates(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to read status updates: %w", err)
	}

	// Process updates for this task
	for _, update := range updates {
		if update.Data.MessageUUID == mapping.MessageUUID {
			// Update task status in AgentFlow
			if err := ts.updateTaskStatusInAgentFlow(mapping.TaskID, update.Data.Status); err != nil {
				ts.logger.Errorf("Failed to update task status: %v", err)
				continue
			}
			ts.logger.Infof("Synced task %s status from Claude: %s", mapping.TaskID, update.Data.Status)
		}
	}

	return nil
}

// getClaudeSessionFile builds the path to Claude session file
func (ts *TaskSynchronizer) getClaudeSessionFile(sessionUUID string) string {
	// Claude stores sessions as: {session_uuid}.jsonl
	return filepath.Join(ts.claudeDir, sessionUUID+".jsonl")
}

// appendToFile appends JSON data to a file (JSON Lines format)
func (ts *TaskSynchronizer) appendToFile(filePath string, data []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Write data with newline
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// readStatusUpdates reads status updates from Claude session file
func (ts *TaskSynchronizer) readStatusUpdates(filePath string) ([]ClaudeStatusUpdate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var updates []ClaudeStatusUpdate
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var update ClaudeStatusUpdate
		if err := decoder.Decode(&update); err != nil {
			ts.logger.Warnf("Failed to decode status update: %v", err)
			continue
		}

		// Only process agentflow_status_update types
		if update.Type == "agentflow_status_update" {
			updates = append(updates, update)
		}
	}

	return updates, nil
}

// updateTaskStatusInAgentFlow updates task status in AgentFlow database
func (ts *TaskSynchronizer) updateTaskStatusInAgentFlow(taskID string, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending":    true,
		"in_progress": true,
		"completed":  true,
		"failed":     true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Update task status
	err := ts.db.UpdateTaskStatus(taskID, status)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// SyncTaskChain synchronizes entire task chain status
// Note: This is a placeholder implementation. Full chain sync requires taskchain package integration.
func (ts *TaskSynchronizer) SyncTaskChain(
	ctx context.Context,
	chainID string,
	direction SyncDirection,
) error {
	// TODO: Implement full chain sync using taskchain package
	ts.logger.Infof("Chain sync requested for %s (not yet implemented)", chainID)
	return nil
}

// StartAutoSync starts automatic synchronization in the background
func (ts *TaskSynchronizer) StartAutoSync(ctx context.Context) error {
	if !ts.autoSync {
		ts.logger.Info("Auto-sync is disabled")
		return nil
	}

	ts.logger.Info("Starting auto-sync")

	ticker := time.NewTicker(ts.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ts.logger.Info("Auto-sync stopped")
			return ctx.Err()
		case <-ticker.C:
			// Sync all active tasks
			if err := ts.syncAllActiveTasks(ctx); err != nil {
				ts.logger.Errorf("Auto-sync failed: %v", err)
			}
		}
	}
}

// syncAllActiveTasks synchronizes all active tasks
func (ts *TaskSynchronizer) syncAllActiveTasks(ctx context.Context) error {
	// TODO: Implement active task syncing
	// This requires database layer to provide query methods
	ts.logger.Debug("syncAllActiveTasks called (not yet implemented)")
	return nil
}

// GetSyncStatus returns synchronization status
func (ts *TaskSynchronizer) GetSyncStatus() map[string]interface{} {
	return map[string]interface{}{
		"auto_sync":      ts.autoSync,
		"sync_interval":  ts.syncInterval.String(),
		"claude_dir":     ts.claudeDir,
	}
}

// ValidateClaudeSession checks if Claude session file is accessible
func (ts *TaskSynchronizer) ValidateClaudeSession(sessionUUID string) error {
	sessionFile := ts.getClaudeSessionFile(sessionUUID)

	// Check if file exists
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("session file does not exist: %s", sessionFile)
	}

	// Check if file is readable
	file, err := os.Open(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	return nil
}
