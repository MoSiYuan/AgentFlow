package database

import (
	"testing"

	"github.com/google/uuid"
)

// TestCreateClaudeMapping 测试创建 Claude 映射
func TestCreateClaudeMapping(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	// 创建一个测试任务
	taskID, err := db.CreateTask("Test Task", "Test Description", "default", "")
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// 创建 Claude 映射
	mapping := &ClaudeMapping{
		TaskID:       taskID,
		SessionUUID:  uuid.New().String(),
		MessageUUID:  uuid.New().String(),
		Slug:         "test-task-slug",
		Source:       "test",
	}

	err = db.CreateClaudeMapping(mapping)
	if err != nil {
		t.Fatalf("Failed to create claude mapping: %v", err)
	}

	// 验证映射已创建
	retrieved, err := db.GetClaudeMappingByTaskID(taskID)
	if err != nil {
		t.Fatalf("Failed to retrieve mapping: %v", err)
	}

	if retrieved.MessageUUID != mapping.MessageUUID {
		t.Errorf("MessageUUID mismatch: got %s, want %s", retrieved.MessageUUID, mapping.MessageUUID)
	}

	if retrieved.Slug != mapping.Slug {
		t.Errorf("Slug mismatch: got %s, want %s", retrieved.Slug, mapping.Slug)
	}

	if retrieved.Source != "test" {
		t.Errorf("Source mismatch: got %s, want test", retrieved.Source)
	}
}

// TestGetClaudeMappingByTaskID 测试通过 Task ID 获取映射
func TestGetClaudeMappingByTaskID(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	taskID, _ := db.CreateTask("Task", "Description", "default", "")

	mapping := &ClaudeMapping{
		TaskID:      taskID,
		SessionUUID: "test-session-123",
		MessageUUID: "test-message-456",
		Slug:        "test-slug",
		Source:      "test",
	}

	db.CreateClaudeMapping(mapping)

	retrieved, err := db.GetClaudeMappingByTaskID(taskID)
	if err != nil {
		t.Fatalf("Failed to retrieve mapping: %v", err)
	}

	if retrieved.TaskID != taskID {
		t.Errorf("TaskID mismatch: got %s, want %s", retrieved.TaskID, taskID)
	}
}

// TestGetClaudeMappingByMessageUUID 测试通过 Message UUID 获取映射
func TestGetClaudeMappingByMessageUUID(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	taskID, _ := db.CreateTask("Task", "Description", "default", "")

	messageUUID := uuid.New().String()
	mapping := &ClaudeMapping{
		TaskID:      taskID,
		SessionUUID: "test-session",
		MessageUUID: messageUUID,
		Slug:        "test-slug",
		Source:      "test",
	}

	db.CreateClaudeMapping(mapping)

	retrieved, err := db.GetClaudeMappingByMessageUUID(messageUUID)
	if err != nil {
		t.Fatalf("Failed to retrieve mapping by message UUID: %v", err)
	}

	if retrieved.MessageUUID != messageUUID {
		t.Errorf("MessageUUID mismatch")
	}
}

// TestGetTaskChainBySession 测试获取任务链
func TestGetTaskChainBySession(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	sessionUUID := uuid.New().String()

	// 创建多个任务形成链
	for i := 0; i < 3; i++ {
		taskID, _ := db.CreateTask(
			"Chain Task "+string(rune('A'+i)),
			"Description",
			"default",
			"",
		)

		mapping := &ClaudeMapping{
			TaskID:  taskID,
			SessionUUID: sessionUUID,
			MessageUUID: uuid.New().String(),
			Slug:     "chain-task-" + string(rune('A'+i)),
			Source:   "chain",
		}

		db.CreateClaudeMapping(mapping)
	}

	// 查询任务链
	chain, err := db.GetTaskChainBySession(sessionUUID)
	if err != nil {
		t.Fatalf("Failed to get task chain: %v", err)
	}

	if len(chain) != 3 {
		t.Errorf("Expected 3 tasks in chain, got %d", len(chain))
	}

	// 验证顺序（按 created_at ASC）
	for i := 0; i < len(chain); i++ {
		if chain[i].Source != "chain" {
			t.Errorf("Task %d source mismatch: got %s, want chain", i, chain[i].Source)
		}
	}
}

// TestGetClaudeMappingBySlug 测试通过 Slug 获取映射
func TestGetClaudeMappingBySlug(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	taskID, _ := db.CreateTask("Task", "Description", "default", "")

	slug := "friendly-slug-name"
	mapping := &ClaudeMapping{
		TaskID:      taskID,
		SessionUUID: "test-session",
		MessageUUID: uuid.New().String(),
		Slug:        slug,
		Source:      "test",
	}

	db.CreateClaudeMapping(mapping)

	retrieved, err := db.GetClaudeMappingBySlug(slug)
	if err != nil {
		t.Fatalf("Failed to get mapping by slug: %v", err)
	}

	if retrieved.Slug != slug {
		t.Errorf("Slug mismatch: got %s, want %s", retrieved.Slug, slug)
	}
}

// TestUpdateClaudeMappingSlug 测试更新 Slug
func TestUpdateClaudeMappingSlug(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	taskID, _ := db.CreateTask("Task", "Description", "default", "")

	mapping := &ClaudeMapping{
		TaskID:      taskID,
		SessionUUID: "test-session",
		MessageUUID: uuid.New().String(),
		Slug:        "old-slug",
		Source:      "test",
	}

	db.CreateClaudeMapping(mapping)

	// 更新 Slug
	newSlug := "new-slug"
	err := db.UpdateClaudeMappingSlug(taskID, newSlug)
	if err != nil {
		t.Fatalf("Failed to update slug: %v", err)
	}

	// 验证更新
	retrieved, _ := db.GetClaudeMappingByTaskID(taskID)
	if retrieved.Slug != newSlug {
		t.Errorf("Slug not updated: got %s, want %s", retrieved.Slug, newSlug)
	}
}

// TestDeleteClaudeMapping 测试删除映射
func TestDeleteClaudeMapping(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	taskID, _ := db.CreateTask("Task", "Description", "default", "")

	mapping := &ClaudeMapping{
		TaskID:      taskID,
		SessionUUID: "test-session",
		MessageUUID: uuid.New().String(),
		Slug:        "test-slug",
		Source:      "test",
	}

	db.CreateClaudeMapping(mapping)

	// 删除映射
	err := db.DeleteClaudeMapping(taskID)
	if err != nil {
		t.Fatalf("Failed to delete mapping: %v", err)
	}

	// 验证删除
	_, err = db.GetClaudeMappingByTaskID(taskID)
	if err == nil {
		t.Error("Expected error when getting deleted mapping, got nil")
	}
}

// TestListClaudeMappingsBySession 测试列出 Session 的所有映射
func TestListClaudeMappingsBySession(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	sessionUUID := uuid.New().String()

	// 创建 5 个映射
	for i := 0; i < 5; i++ {
		taskID, _ := db.CreateTask("Task", "Description", "default", "")

		mapping := &ClaudeMapping{
			TaskID:      taskID,
			SessionUUID: sessionUUID,
			MessageUUID: uuid.New().String(),
			Slug:        "test-slug",
			Source:      "test",
		}

		db.CreateClaudeMapping(mapping)
	}

	// 查询所有映射（limit 3, offset 0）
	mappings, err := db.ListClaudeMappingsBySession(sessionUUID, 3, 0)
	if err != nil {
		t.Fatalf("Failed to list mappings: %v", err)
	}

	if len(mappings) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(mappings))
	}

	// 查询下一页（offset 3）
	mappings2, err := db.ListClaudeMappingsBySession(sessionUUID, 3, 3)
	if err != nil {
		t.Fatalf("Failed to list mappings page 2: %v", err)
	}

	if len(mappings2) != 2 {
		t.Errorf("Expected 2 mappings on page 2, got %d", len(mappings2))
	}
}

// TestNonExistentMapping 测试查询不存在的映射
func TestNonExistentMapping(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	// 测试不存在的 Task ID
	_, err := db.GetClaudeMappingByTaskID("999999")
	if err == nil {
		t.Error("Expected error for non-existent task ID, got nil")
	}

	// 测试不存在的 Message UUID
	_, err = db.GetClaudeMappingByMessageUUID("non-existent-uuid")
	if err == nil {
		t.Error("Expected error for non-existent message UUID, got nil")
	}

	// 测试不存在的 Slug
	_, err = db.GetClaudeMappingBySlug("non-existent-slug")
	if err == nil {
		t.Error("Expected error for non-existent slug, got nil")
	}
}

// TestParentMessageUUID 测试父消息 UUID
func TestParentMessageUUID(t *testing.T) {
	db, _ := NewDatabase(":memory:")
	db.Init()

	// 创建父任务
	parentTaskID, _ := db.CreateTask("Parent Task", "Description", "default", "")

	parentMapping := &ClaudeMapping{
		TaskID:       parentTaskID,
		SessionUUID:  "test-session",
		MessageUUID:  uuid.New().String(),
		Slug:         "parent-slug",
		Source:       "test",
	}

	db.CreateClaudeMapping(parentMapping)

	// 创建子任务（带父 UUID）
	childTaskID, _ := db.CreateTask("Child Task", "Description", "default", "")

	childMapping := &ClaudeMapping{
		TaskID:             childTaskID,
		SessionUUID:        "test-session",
		MessageUUID:        uuid.New().String(),
		ParentMessageUUID:  &parentMapping.MessageUUID,
		Slug:               "child-slug",
		Source:             "test",
	}

	db.CreateClaudeMapping(childMapping)

	// 验证父子关系
	childRetrieved, _ := db.GetClaudeMappingByTaskID(childTaskID)
	if childRetrieved.ParentMessageUUID == nil {
		t.Error("Expected parent message UUID, got nil")
	} else if *childRetrieved.ParentMessageUUID != parentMapping.MessageUUID {
		t.Errorf("Parent message UUID mismatch: got %s, want %s",
			*childRetrieved.ParentMessageUUID, parentMapping.MessageUUID)
	}
}
