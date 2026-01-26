package taskchain

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jiangxiaolong/agentflow-go/internal/database"
)

// ChainType 任务链类型
type ChainType string

const (
	ChainSequential ChainType = "sequential" // 串行：一个接一个
	ChainParallel   ChainType = "parallel"   // 并行：同时执行
	ChainTree       ChainType = "tree"       // 树形：支持分支
)

// ChainStatus 任务链状态
type ChainStatus string

const (
	ChainStatusPending   ChainStatus = "pending"
	ChainStatusRunning   ChainStatus = "running"
	ChainStatusCompleted ChainStatus = "completed"
	ChainStatusFailed    ChainStatus = "failed"
)

// TaskChain 任务链
type TaskChain struct {
	ID               string      `json:"id"`
	SessionUUID      string      `json:"session_uuid"`
	RootMessageUUID  string      `json:"root_message_uuid"`
	ChainType        ChainType   `json:"chain_type"`
	Status           ChainStatus `json:"status"`
	CreatedAt        time.Time   `json:"created_at"`
	StartedAt        *time.Time  `json:"started_at,omitempty"`
	CompletedAt      *time.Time  `json:"completed_at,omitempty"`
}

// TaskChainNode 任务链节点
type TaskChainNode struct {
	ID            int    `json:"id"`
	ChainID       string `json:"chain_id"`
	TaskID        int    `json:"task_id"`
	ParentNodeID  *int   `json:"parent_node_id,omitempty"`
	NodeOrder     int    `json:"node_order"`
	CreatedAt     time.Time `json:"created_at"`
}

// TaskChainManager 任务链管理器
type TaskChainManager struct {
	db *database.Database
}

// NewTaskChainManager 创建任务链管理器
func NewTaskChainManager(db *database.Database) *TaskChainManager {
	return &TaskChainManager{db: db}
}

// CreateSequentialChain 创建串行任务链
func (m *TaskChainManager) CreateSequentialChain(sessionUUID string, taskIDs []int) (*TaskChain, error) {
	chainID := uuid.New().String()
	rootMessageUUID := uuid.New().String()

	// 创建任务链
	chain := &TaskChain{
		ID:              chainID,
		SessionUUID:     sessionUUID,
		RootMessageUUID: rootMessageUUID,
		ChainType:       ChainSequential,
		Status:          ChainStatusPending,
	}

	if err := m.createChain(chain); err != nil {
		return nil, err
	}

	// 添加节点（按顺序）
	for i, taskID := range taskIDs {
		node := &TaskChainNode{
			ChainID:   chainID,
			TaskID:    taskID,
			NodeOrder: i,
		}

		if err := m.addNode(node); err != nil {
			return nil, fmt.Errorf("failed to add node %d: %w", i, err)
		}
	}

	return chain, nil
}

// CreateParallelChain 创建并行任务链
func (m *TaskChainManager) CreateParallelChain(sessionUUID string, taskIDs []int) (*TaskChain, error) {
	chainID := uuid.New().String()
	rootMessageUUID := uuid.New().String()

	// 创建任务链
	chain := &TaskChain{
		ID:              chainID,
		SessionUUID:     sessionUUID,
		RootMessageUUID: rootMessageUUID,
		ChainType:       ChainParallel,
		Status:          ChainStatusPending,
	}

	if err := m.createChain(chain); err != nil {
		return nil, err
	}

	// 添加节点（所有节点都是根节点，order 都是 0）
	for _, taskID := range taskIDs {
		node := &TaskChainNode{
			ChainID:   chainID,
			TaskID:    taskID,
			NodeOrder: 0, // 并行任务没有顺序
		}

		if err := m.addNode(node); err != nil {
			return nil, fmt.Errorf("failed to add parallel node: %w", err)
		}
	}

	return chain, nil
}

// CreateTreeChain 创建树形任务链
func (m *TaskChainManager) CreateTreeChain(sessionUUID string, tasks []TreeTaskDefinition) (*TaskChain, error) {
	chainID := uuid.New().String()
	rootMessageUUID := uuid.New().String()

	// 创建任务链
	chain := &TaskChain{
		ID:              chainID,
		SessionUUID:     sessionUUID,
		RootMessageUUID: rootMessageUUID,
		ChainType:       ChainTree,
		Status:          ChainStatusPending,
	}

	if err := m.createChain(chain); err != nil {
		return nil, err
	}

	// 添加节点（支持父子关系）
	for i, taskDef := range tasks {
		node := &TaskChainNode{
			ChainID:      chainID,
			TaskID:       taskDef.TaskID,
			ParentNodeID: taskDef.ParentNodeID,
			NodeOrder:    i,
		}

		if err := m.addNode(node); err != nil {
			return nil, fmt.Errorf("failed to add tree node %d: %w", i, err)
		}
	}

	return chain, nil
}

// TreeTaskDefinition 树形任务定义
type TreeTaskDefinition struct {
	TaskID       int
	ParentNodeID *int // 父节点 ID，null 表示根节点
}

// GetChain 获取任务链
func (m *TaskChainManager) GetChain(chainID string) (*TaskChain, error) {
	query := `
		SELECT id, session_uuid, root_message_uuid, chain_type, status,
		       created_at, started_at, completed_at
		FROM task_chains
		WHERE id = ?
	`

	row := m.db.GetDB().QueryRow(query, chainID)

	var chain TaskChain
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&chain.ID,
		&chain.SessionUUID,
		&chain.RootMessageUUID,
		&chain.ChainType,
		&chain.Status,
		&chain.CreatedAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chain not found: %s", chainID)
		}
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	if startedAt.Valid {
		chain.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		chain.CompletedAt = &completedAt.Time
	}

	return &chain, nil
}

// GetChainNodes 获取任务链的所有节点
func (m *TaskChainManager) GetChainNodes(chainID string) ([]TaskChainNode, error) {
	query := `
		SELECT id, chain_id, task_id, parent_node_id, node_order, created_at
		FROM task_chain_nodes
		WHERE chain_id = ?
		ORDER BY node_order ASC
	`

	rows, err := m.db.GetDB().Query(query, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain nodes: %w", err)
	}
	defer rows.Close()

	var nodes []TaskChainNode
	for rows.Next() {
		var node TaskChainNode
		err := rows.Scan(
			&node.ID,
			&node.ChainID,
			&node.TaskID,
			&node.ParentNodeID,
			&node.NodeOrder,
			&node.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// UpdateChainStatus 更新任务链状态
func (m *TaskChainManager) UpdateChainStatus(chainID string, status ChainStatus) error {
	query := `
		UPDATE task_chains
		SET status = ?,
		    started_at = CASE WHEN ? = 'running' THEN CURRENT_TIMESTAMP ELSE started_at END,
		    completed_at = CASE WHEN ? IN ('completed', 'failed') THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = ?
	`

	_, err := m.db.GetDB().Exec(query, status, status, status, chainID)
	if err != nil {
		return fmt.Errorf("failed to update chain status: %w", err)
	}

	return nil
}

// GetChainsBySession 获取 Session 的所有任务链
func (m *TaskChainManager) GetChainsBySession(sessionUUID string) ([]TaskChain, error) {
	query := `
		SELECT id, session_uuid, root_message_uuid, chain_type, status,
		       created_at, started_at, completed_at
		FROM task_chains
		WHERE session_uuid = ?
		ORDER BY created_at DESC
	`

	rows, err := m.db.GetDB().Query(query, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chains by session: %w", err)
	}
	defer rows.Close()

	var chains []TaskChain
	for rows.Next() {
		var chain TaskChain
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&chain.ID,
			&chain.SessionUUID,
			&chain.RootMessageUUID,
			&chain.ChainType,
			&chain.Status,
			&chain.CreatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		if startedAt.Valid {
			chain.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			chain.CompletedAt = &completedAt.Time
		}

		chains = append(chains, chain)
	}

	return chains, nil
}

// createChain 创建任务链（内部方法）
func (m *TaskChainManager) createChain(chain *TaskChain) error {
	query := `
		INSERT INTO task_chains (id, session_uuid, root_message_uuid, chain_type, status)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := m.db.GetDB().Exec(query,
		chain.ID,
		chain.SessionUUID,
		chain.RootMessageUUID,
		chain.ChainType,
		chain.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to create chain: %w", err)
	}

	return nil
}

// addNode 添加任务链节点（内部方法）
func (m *TaskChainManager) addNode(node *TaskChainNode) error {
	query := `
		INSERT INTO task_chain_nodes (chain_id, task_id, parent_node_id, node_order)
		VALUES (?, ?, ?, ?)
	`

	result, err := m.db.GetDB().Exec(query,
		node.ChainID,
		node.TaskID,
		node.ParentNodeID,
		node.NodeOrder,
	)

	if err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	node.ID = int(id)
	return nil
}
