package query

import (
	"fmt"
	"testing"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
)

func setupTestQuery(t *testing.T) *UnifiedQuery {
	db, err := database.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := db.Init(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// Create test tasks
	for i := 1; i <= 5; i++ {
		taskID, err := db.CreateTask(
			fmt.Sprintf("Task %d", i),
			fmt.Sprintf("Description %d", i),
			"default",
			"",
		)
		if err != nil {
			t.Fatalf("failed to create task %d: %v", i, err)
		}

		if taskID == "" {
			t.Fatalf("task ID %d is empty", i)
		}

		// Create Claude mappings for some tasks
		if i <= 3 {
			mapping := &database.ClaudeMapping{
				TaskID:      taskID,
				SessionUUID: "test-session-123",
				MessageUUID: fmt.Sprintf("msg-%03d", i),
				Slug:        fmt.Sprintf("task-%d", i),
				Source:      "test",
			}

			err = db.CreateClaudeMapping(mapping)
			if err != nil {
				t.Fatalf("failed to create mapping for task %d: %v", i, err)
			}
		}
	}

	return NewUnifiedQuery(db)
}

func TestQueryTasks(t *testing.T) {
	uq := setupTestQuery(t)

	// Query all tasks
	query := &TaskQuery{
		Limit: intPtr(10),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}

	if len(result.Tasks) != 5 {
		t.Errorf("expected 5 tasks, got %d", len(result.Tasks))
	}

	if result.TotalCount != 5 {
		t.Errorf("expected total count 5, got %d", result.TotalCount)
	}
}

func TestQueryTasksByStatus(t *testing.T) {
	uq := setupTestQuery(t)

	// Query tasks by status
	status := "pending"
	query := &TaskQuery{
		Status: &status,
		Limit:  intPtr(10),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks by status: %v", err)
	}

	// All tasks should be pending initially
	if len(result.Tasks) == 0 {
		t.Error("expected at least one pending task")
	}

	for _, task := range result.Tasks {
		if task.Status != "pending" {
			t.Errorf("expected status 'pending', got '%s'", task.Status)
		}
	}
}

func TestQueryTasksBySessionUUID(t *testing.T) {
	uq := setupTestQuery(t)

	// Query tasks by Claude session
	sessionUUID := "test-session-123"
	query := &TaskQuery{
		SessionUUID: &sessionUUID,
		Limit:       intPtr(10),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks by session: %v", err)
	}

	if len(result.Tasks) != 3 {
		t.Errorf("expected 3 tasks with Claude mappings, got %d", len(result.Tasks))
	}

	// Verify all tasks have Claude info
	for _, task := range result.Tasks {
		if task.Claude == nil {
			t.Errorf("task %s should have Claude mapping", task.ID)
		}
		if task.Claude.SessionUUID != sessionUUID {
			t.Errorf("expected session UUID %s, got %s", sessionUUID, task.Claude.SessionUUID)
		}
	}
}

func TestQueryTasksByMessageUUID(t *testing.T) {
	uq := setupTestQuery(t)

	// Query task by Claude message UUID
	messageUUID := "msg-001"
	query := &TaskQuery{
		MessageUUID: &messageUUID,
		Limit:       intPtr(10),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks by message UUID: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result.Tasks))
	}

	if result.Tasks[0].Claude == nil {
		t.Error("expected task to have Claude mapping")
	}

	if result.Tasks[0].Claude.MessageUUID != messageUUID {
		t.Errorf("expected message UUID %s, got %s", messageUUID, result.Tasks[0].Claude.MessageUUID)
	}
}

func TestQueryTasksBySlug(t *testing.T) {
	uq := setupTestQuery(t)

	// Query task by slug
	slug := "task-1"
	query := &TaskQuery{
		Slug:  &slug,
		Limit: intPtr(10),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks by slug: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result.Tasks))
	}

	if result.Tasks[0].Claude == nil {
		t.Error("expected task to have Claude mapping")
	}

	if result.Tasks[0].Claude.Slug != slug {
		t.Errorf("expected slug %s, got %s", slug, result.Tasks[0].Claude.Slug)
	}
}

func TestGetTaskByClaudeMessageUUID(t *testing.T) {
	uq := setupTestQuery(t)

	// Get task by message UUID
	task, err := uq.GetTaskByClaudeMessageUUID("msg-002")
	if err != nil {
		t.Fatalf("failed to get task by message UUID: %v", err)
	}

	if task == nil {
		t.Fatal("expected task to be returned")
	}

	if task.Claude == nil {
		t.Error("expected task to have Claude mapping")
	}

	if task.Claude.MessageUUID != "msg-002" {
		t.Errorf("expected message UUID 'msg-002', got '%s'", task.Claude.MessageUUID)
	}
}

func TestGetTaskByClaudeMessageUUIDNotFound(t *testing.T) {
	uq := setupTestQuery(t)

	// Get non-existent task
	_, err := uq.GetTaskByClaudeMessageUUID("non-existent")
	if err == nil {
		t.Error("expected error for non-existent message UUID")
	}
}

func TestGetTasksByClaudeSession(t *testing.T) {
	uq := setupTestQuery(t)

	// Get tasks by session
	result, err := uq.GetTasksByClaudeSession("test-session-123", 10, 0)
	if err != nil {
		t.Fatalf("failed to get tasks by session: %v", err)
	}

	if len(result.Tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(result.Tasks))
	}

	if result.TotalCount != 3 {
		t.Errorf("expected total count 3, got %d", result.TotalCount)
	}
}

func TestGetTaskBySlug(t *testing.T) {
	uq := setupTestQuery(t)

	// Get task by slug
	task, err := uq.GetTaskBySlug("task-2")
	if err != nil {
		t.Fatalf("failed to get task by slug: %v", err)
	}

	if task == nil {
		t.Fatal("expected task to be returned")
	}

	if task.Claude.Slug != "task-2" {
		t.Errorf("expected slug 'task-2', got '%s'", task.Claude.Slug)
	}
}

func TestGetTaskBySlugNotFound(t *testing.T) {
	uq := setupTestQuery(t)

	// Get non-existent task
	_, err := uq.GetTaskBySlug("non-existent-slug")
	if err == nil {
		t.Error("expected error for non-existent slug")
	}
}

func TestGetTasksWithClaudeInfo(t *testing.T) {
	uq := setupTestQuery(t)

	// Get tasks with Claude info
	result, err := uq.GetTasksWithClaudeInfo(10, 0)
	if err != nil {
		t.Fatalf("failed to get tasks with Claude info: %v", err)
	}

	// Should only return tasks with Claude mappings
	if len(result.Tasks) != 3 {
		t.Errorf("expected 3 tasks with Claude info, got %d", len(result.Tasks))
	}

	for _, task := range result.Tasks {
		if task.Claude == nil {
			t.Errorf("task %s should have Claude mapping", task.ID)
		}
	}
}

func TestQueryTasksWithPagination(t *testing.T) {
	uq := setupTestQuery(t)

	// Get first page
	limit := 2
	offset := 0
	query := &TaskQuery{
		Limit:  &limit,
		Offset: &offset,
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}

	if len(result.Tasks) != 2 {
		t.Errorf("expected 2 tasks on first page, got %d", len(result.Tasks))
	}

	// Get second page
	offset = 2
	query.Offset = &offset

	result2, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}

	if len(result2.Tasks) != 2 {
		t.Errorf("expected 2 tasks on second page, got %d", len(result2.Tasks))
	}

	// Verify tasks are different
	if result.Tasks[0].ID == result2.Tasks[0].ID {
		t.Error("expected different tasks on different pages")
	}
}

func TestQueryTasksWithSorting(t *testing.T) {
	uq := setupTestQuery(t)

	// Query tasks sorted by title ascending
	orderBy := "title"
	orderDesc := false
	query := &TaskQuery{
		OrderBy:   &orderBy,
		OrderDesc: &orderDesc,
		Limit:     intPtr(10),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}

	if len(result.Tasks) < 2 {
		t.Fatal("expected at least 2 tasks")
	}

	// Verify ascending order
	if result.Tasks[0].Title > result.Tasks[1].Title {
		t.Error("expected tasks to be sorted by title ascending")
	}
}

func TestQueryTasksWithMultipleFilters(t *testing.T) {
	uq := setupTestQuery(t)

	// Query with multiple filters
	sessionUUID := "test-session-123"
	group := "default"
	query := &TaskQuery{
		SessionUUID: &sessionUUID,
		Group:       &group,
		Limit:       intPtr(10),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}

	// Verify all filters are applied
	for _, task := range result.Tasks {
		if task.Claude != nil && task.Claude.SessionUUID != sessionUUID {
			t.Errorf("expected session UUID %s, got %s", sessionUUID, task.Claude.SessionUUID)
		}
		if task.GroupName != group {
			t.Errorf("expected group %s, got %s", group, task.GroupName)
		}
	}
}

