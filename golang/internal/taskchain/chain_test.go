package taskchain

import (
	"fmt"
	"testing"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
)

func TestCreateSequentialChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	manager := NewTaskChainManager(db)

	sessionUUID := "test-session-123"
	taskIDs := []int{1, 2, 3}

	chain, err := manager.CreateSequentialChain(sessionUUID, taskIDs)
	if err != nil {
		t.Fatalf("failed to create sequential chain: %v", err)
	}

	if chain == nil {
		t.Fatal("chain is nil")
	}

	if chain.ChainType != ChainSequential {
		t.Errorf("expected chain type %s, got %s", ChainSequential, chain.ChainType)
	}

	if chain.Status != ChainStatusPending {
		t.Errorf("expected status %s, got %s", ChainStatusPending, chain.Status)
	}

	// 验证节点
	nodes, err := manager.GetChainNodes(chain.ID)
	if err != nil {
		t.Fatalf("failed to get chain nodes: %v", err)
	}

	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(nodes))
	}

	// 验证顺序
	for i, node := range nodes {
		if node.NodeOrder != i {
			t.Errorf("node %d: expected order %d, got %d", i, i, node.NodeOrder)
		}
		if node.TaskID != taskIDs[i] {
			t.Errorf("node %d: expected task_id %d, got %d", i, taskIDs[i], node.TaskID)
		}
	}
}

func TestCreateParallelChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	manager := NewTaskChainManager(db)

	sessionUUID := "test-session-456"
	taskIDs := []int{4, 5, 6}

	chain, err := manager.CreateParallelChain(sessionUUID, taskIDs)
	if err != nil {
		t.Fatalf("failed to create parallel chain: %v", err)
	}

	if chain.ChainType != ChainParallel {
		t.Errorf("expected chain type %s, got %s", ChainParallel, chain.ChainType)
	}

	// 验证节点
	nodes, err := manager.GetChainNodes(chain.ID)
	if err != nil {
		t.Fatalf("failed to get chain nodes: %v", err)
	}

	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(nodes))
	}

	// 所有节点的 order 应该都是 0（并行）
	for _, node := range nodes {
		if node.NodeOrder != 0 {
			t.Errorf("parallel node: expected order 0, got %d", node.NodeOrder)
		}
	}
}

func TestCreateTreeChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	manager := NewTaskChainManager(db)

	sessionUUID := "test-session-789"

	// 创建一个简单的树结构：
	//     Task 1 (root)
	//       ├─ Task 2
	//       └─ Task 3
	tasks := []TreeTaskDefinition{
		{TaskID: 1, ParentNodeID: nil},    // 根节点
		{TaskID: 2, ParentNodeID: intPtr(1)}, // Task 1 的子节点
		{TaskID: 3, ParentNodeID: intPtr(1)}, // Task 1 的子节点
	}

	chain, err := manager.CreateTreeChain(sessionUUID, tasks)
	if err != nil {
		t.Fatalf("failed to create tree chain: %v", err)
	}

	if chain.ChainType != ChainTree {
		t.Errorf("expected chain type %s, got %s", ChainTree, chain.ChainType)
	}

	// 验证节点
	nodes, err := manager.GetChainNodes(chain.ID)
	if err != nil {
		t.Fatalf("failed to get chain nodes: %v", err)
	}

	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(nodes))
	}

	// 验证父子关系
	rootNode := findNodeByTaskID(nodes, 1)
	if rootNode == nil {
		t.Fatal("root node not found")
	}
	if rootNode.ParentNodeID != nil {
		t.Errorf("root node should not have parent, got %v", rootNode.ParentNodeID)
	}

	childNode2 := findNodeByTaskID(nodes, 2)
	if childNode2 == nil {
		t.Fatal("child node 2 not found")
	}
	if childNode2.ParentNodeID == nil || *childNode2.ParentNodeID != rootNode.ID {
		t.Errorf("node 2 should be child of root node")
	}
}

func TestGetChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	manager := NewTaskChainManager(db)

	sessionUUID := "test-session-get"
	taskIDs := []int{10, 20}

	createdChain, err := manager.CreateSequentialChain(sessionUUID, taskIDs)
	if err != nil {
		t.Fatalf("failed to create chain: %v", err)
	}

	// 获取创建的任务链
	retrievedChain, err := manager.GetChain(createdChain.ID)
	if err != nil {
		t.Fatalf("failed to get chain: %v", err)
	}

	if retrievedChain.ID != createdChain.ID {
		t.Errorf("expected id %s, got %s", createdChain.ID, retrievedChain.ID)
	}

	if retrievedChain.SessionUUID != sessionUUID {
		t.Errorf("expected session_uuid %s, got %s", sessionUUID, retrievedChain.SessionUUID)
	}
}

func TestUpdateChainStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	manager := NewTaskChainManager(db)

	chain, err := manager.CreateSequentialChain("test-session-status", []int{1})
	if err != nil {
		t.Fatalf("failed to create chain: %v", err)
	}

	// 更新为 running
	err = manager.UpdateChainStatus(chain.ID, ChainStatusRunning)
	if err != nil {
		t.Fatalf("failed to update chain status to running: %v", err)
	}

	updatedChain, err := manager.GetChain(chain.ID)
	if err != nil {
		t.Fatalf("failed to get updated chain: %v", err)
	}

	if updatedChain.Status != ChainStatusRunning {
		t.Errorf("expected status %s, got %s", ChainStatusRunning, updatedChain.Status)
	}

	if updatedChain.StartedAt == nil {
		t.Error("started_at should be set when status becomes running")
	}

	// 更新为 completed
	err = manager.UpdateChainStatus(chain.ID, ChainStatusCompleted)
	if err != nil {
		t.Fatalf("failed to update chain status to completed: %v", err)
	}

	updatedChain, err = manager.GetChain(chain.ID)
	if err != nil {
		t.Fatalf("failed to get updated chain: %v", err)
	}

	if updatedChain.Status != ChainStatusCompleted {
		t.Errorf("expected status %s, got %s", ChainStatusCompleted, updatedChain.Status)
	}

	if updatedChain.CompletedAt == nil {
		t.Error("completed_at should be set when status becomes completed")
	}
}

func TestGetChainsBySession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	manager := NewTaskChainManager(db)

	sessionUUID := "test-session-multi"

	// 创建多个任务链
	chain1, _ := manager.CreateSequentialChain(sessionUUID, []int{1, 2})
	chain2, _ := manager.CreateParallelChain(sessionUUID, []int{3, 4})

	chains, err := manager.GetChainsBySession(sessionUUID)
	if err != nil {
		t.Fatalf("failed to get chains by session: %v", err)
	}

	if len(chains) != 2 {
		t.Errorf("expected 2 chains, got %d", len(chains))
	}

	// 验证返回的链包含两个创建的链
	chainIDs := make(map[string]bool)
	for _, chain := range chains {
		chainIDs[chain.ID] = true
	}

	if !chainIDs[chain1.ID] {
		t.Errorf("chains should contain %s", chain1.ID)
	}
	if !chainIDs[chain2.ID] {
		t.Errorf("chains should contain %s", chain2.ID)
	}
}

func TestNonExistentChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	manager := NewTaskChainManager(db)

	_, err := manager.GetChain("non-existent-id")
	if err == nil {
		t.Error("expected error when getting non-existent chain")
	}
}

// 辅助函数

func setupTestDB(t *testing.T) *database.Database {
	db, err := database.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := db.Init(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// 创建测试任务
	for i := 1; i <= 10; i++ {
		taskID, _ := db.CreateTask(
			fmt.Sprintf("Test Task %d", i),
			fmt.Sprintf("Description %d", i),
			"default",
			"",
		)
		// 验证任务创建成功
		if taskID == "" {
			t.Fatalf("failed to create test task %d", i)
		}
	}

	return db
}

func intPtr(i int) *int {
	return &i
}

func findNodeByTaskID(nodes []TaskChainNode, taskID int) *TaskChainNode {
	for _, node := range nodes {
		if node.TaskID == taskID {
			return &node
		}
	}
	return nil
}
