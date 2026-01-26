package sync

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"bytes"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
	"github.com/sirupsen/logrus"
)

func setupTestSync(t *testing.T) (*TaskSynchronizer, *database.Database, string) {
	// Create temporary directory for Claude files
	tempDir := t.TempDir()

	// Create in-memory database
	db, err := database.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := db.Init(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// Create test task
	taskID, err := db.CreateTask("Test Task", "Description", "default", "")
	if err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	if taskID == "" {
		t.Fatal("task ID is empty")
	}

	// Create Claude mapping
	mapping := &database.ClaudeMapping{
		TaskID:      taskID,
		SessionUUID: "test-session-123",
		MessageUUID: "test-message-456",
		Slug:        "test-task",
		Source:      "test",
	}

	err = db.CreateClaudeMapping(mapping)
	if err != nil {
		t.Fatalf("failed to create Claude mapping: %v", err)
	}

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	synchronizer := NewTaskSynchronizer(db, tempDir, logger)

	return synchronizer, db, tempDir
}

func TestSyncToClaude(t *testing.T) {
	sync, db, tempDir := setupTestSync(t)
	defer db.Close()

	ctx := context.Background()

	// Sync task to Claude
	err := sync.SyncTaskStatus(ctx, "1", SyncToClaude)
	if err != nil {
		t.Fatalf("failed to sync task to Claude: %v", err)
	}

	// Verify session file was created
	sessionFile := filepath.Join(tempDir, "test-session-123.jsonl")
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Errorf("session file was not created: %s", sessionFile)
	}

	// Verify file content
	content, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("failed to read session file: %v", err)
	}

	if len(content) == 0 {
		t.Error("session file is empty")
	}
}

func TestSyncFromClaude(t *testing.T) {
	sync, db, tempDir := setupTestSync(t)
	defer db.Close()

	ctx := context.Background()

	// First, manually create a session file with status update
	sessionFile := filepath.Join(tempDir, "test-session-123.jsonl")
	update := ClaudeStatusUpdate{
		Type:      "agentflow_status_update",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: ClaudeStatusUpdateData{
			TaskID:      "1",
			MessageUUID: "test-message-456",
			Status:      "completed",
		},
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("failed to marshal update: %v", err)
	}

	if err := sync.appendToFile(sessionFile, data); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	// Update task status to pending first
	if err := db.UpdateTaskStatus("1", "pending"); err != nil {
		t.Fatalf("failed to update task status: %v", err)
	}

	// Sync from Claude
	err = sync.SyncTaskStatus(ctx, "1", SyncFromClaude)
	if err != nil {
		t.Fatalf("failed to sync task from Claude: %v", err)
	}

	// Verify task status was updated
	task, err := db.GetTask("1")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if task.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", task.Status)
	}
}

func TestSyncBidirectional(t *testing.T) {
	sync, db, tempDir := setupTestSync(t)
	defer db.Close()

	ctx := context.Background()

	// Create session file with status update
	sessionFile := filepath.Join(tempDir, "test-session-123.jsonl")
	update := ClaudeStatusUpdate{
		Type:      "agentflow_status_update",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: ClaudeStatusUpdateData{
			TaskID:      "1",
			MessageUUID: "test-message-456",
			Status:      "in_progress",
		},
	}

	data, _ := json.Marshal(update)
	sync.appendToFile(sessionFile, data)

	// Sync bidirectionally
	err := sync.SyncTaskStatus(ctx, "1", SyncBidirectional)
	if err != nil {
		t.Fatalf("failed to sync task bidirectionally: %v", err)
	}

	// Verify session file has new entry from AgentFlow
	content, _ := os.ReadFile(sessionFile)
	lines := len(bytes.Split(content, []byte{'\n'}))
	if lines < 2 {
		t.Errorf("expected at least 2 lines in session file, got %d", lines)
	}
}

func TestSyncTaskChain(t *testing.T) {
	sync, db, tempDir := setupTestSync(t)
	defer db.Close()

	ctx := context.Background()

	// Create additional tasks
	taskID2, _ := db.CreateTask("Task 2", "Description", "default", "")
	taskID3, _ := db.CreateTask("Task 3", "Description", "default", "")

	// Create task IDs for syncing
	taskIDs := []string{"1", taskID2, taskID3}

	// Sync individual tasks
	for _, taskID := range taskIDs {
		err := sync.SyncTaskStatus(ctx, taskID, SyncToClaude)
		if err != nil {
			t.Errorf("failed to sync task %s: %v", taskID, err)
		}
	}

	// Verify session files were created
	sessionFile := filepath.Join(tempDir, "test-session-123.jsonl")
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Error("session file was not created")
	}
}

func TestGetSyncStatus(t *testing.T) {
	sync, db, _ := setupTestSync(t)
	defer db.Close()

	status := sync.GetSyncStatus()

	if status == nil {
		t.Fatal("status is nil")
	}

	if autoSync, ok := status["auto_sync"].(bool); !ok || !autoSync {
		t.Error("expected auto_sync to be true")
	}

	if claudeDir, ok := status["claude_dir"].(string); !ok || claudeDir == "" {
		t.Error("expected claude_dir to be set")
	}
}

func TestValidateClaudeSession(t *testing.T) {
	sync, db, _ := setupTestSync(t)
	defer db.Close()

	ctx := context.Background()

	// Sync to create session file
	err := sync.SyncTaskStatus(ctx, "1", SyncToClaude)
	if err != nil {
		t.Fatalf("failed to sync task: %v", err)
	}

	// Validate existing session
	err = sync.ValidateClaudeSession("test-session-123")
	if err != nil {
		t.Errorf("failed to validate existing session: %v", err)
	}

	// Validate non-existent session
	err = sync.ValidateClaudeSession("non-existent-session")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestSetAutoSync(t *testing.T) {
	sync, db, _ := setupTestSync(t)
	defer db.Close()

	// Disable auto-sync
	sync.SetAutoSync(false)

	status := sync.GetSyncStatus()
	if autoSync, ok := status["auto_sync"].(bool); ok && autoSync {
		t.Error("expected auto_sync to be false")
	}

	// Re-enable auto-sync
	sync.SetAutoSync(true)

	status = sync.GetSyncStatus()
	if autoSync, ok := status["auto_sync"].(bool); !ok || !autoSync {
		t.Error("expected auto_sync to be true")
	}
}

func TestSetSyncInterval(t *testing.T) {
	sync, db, _ := setupTestSync(t)
	defer db.Close()

	// Set custom interval
	sync.SetSyncInterval(10 * time.Second)

	status := sync.GetSyncStatus()
	if interval, ok := status["sync_interval"].(string); !ok || interval != "10s" {
		t.Errorf("expected sync_interval '10s', got %v", interval)
	}
}

func TestReadStatusUpdates(t *testing.T) {
	sync, db, tempDir := setupTestSync(t)
	defer db.Close()

	// Create session file with multiple updates
	sessionFile := filepath.Join(tempDir, "test-session-123.jsonl")

	updates := []ClaudeStatusUpdate{
		{
			Type:      "agentflow_status_update",
			Timestamp: time.Now().Format(time.RFC3339),
			Data: ClaudeStatusUpdateData{
				TaskID:      "1",
				MessageUUID: "test-message-456",
				Status:      "pending",
			},
		},
		{
			Type:      "other_type",
			Timestamp: time.Now().Format(time.RFC3339),
			Data: ClaudeStatusUpdateData{
				TaskID:      "1",
				MessageUUID: "test-message-456",
				Status:      "in_progress",
			},
		},
		{
			Type:      "agentflow_status_update",
			Timestamp: time.Now().Format(time.RFC3339),
			Data: ClaudeStatusUpdateData{
				TaskID:      "1",
				MessageUUID: "test-message-456",
				Status:      "completed",
			},
		},
	}

	for _, update := range updates {
		data, _ := json.Marshal(update)
		if err := sync.appendToFile(sessionFile, data); err != nil {
			t.Fatalf("failed to write update: %v", err)
		}
	}

	// Read status updates
	readUpdates, err := sync.readStatusUpdates(sessionFile)
	if err != nil {
		t.Fatalf("failed to read status updates: %v", err)
	}

	// Should only return agentflow_status_update types
	expectedCount := 2
	if len(readUpdates) != expectedCount {
		t.Errorf("expected %d updates, got %d", expectedCount, len(readUpdates))
	}
}

func TestUpdateTaskStatusInAgentFlow(t *testing.T) {
	sync, db, _ := setupTestSync(t)
	defer db.Close()

	// Test valid status update
	err := sync.updateTaskStatusInAgentFlow("1", "completed")
	if err != nil {
		t.Errorf("failed to update task status: %v", err)
	}

	task, _ := db.GetTask("1")
	if task.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", task.Status)
	}

	// Test invalid status
	err = sync.updateTaskStatusInAgentFlow("1", "invalid_status")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetClaudeSessionFile(t *testing.T) {
	sync, db, _ := setupTestSync(t)
	defer db.Close()

	sessionFile := sync.getClaudeSessionFile("test-session-abc")

	expected := "test-session-abc.jsonl"
	if filepath.Base(sessionFile) != expected {
		t.Errorf("expected filename '%s', got '%s'", expected, filepath.Base(sessionFile))
	}
}

// Helper functions
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
