package database

import "time"

type Task struct {
	ID                 string     `json:"id"`
	ParentID           *int       `json:"parent_id,omitempty"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	GroupName          string     `json:"group_name"`
	CompletionCriteria string     `json:"completion_criteria"`
	Status             string     `json:"status"` // pending, running, completed, failed, blocked
	Priority           int        `json:"priority"`
	LockHolder         string     `json:"lock_holder"`
	LockTime           *time.Time `json:"lock_time"`
	Result             string     `json:"result"`
	Error              string     `json:"error"`
	WorkspaceDir       string     `json:"workspace_dir,omitempty"`
	Sandboxed          bool       `json:"sandboxed"`
	AllowNetwork       bool       `json:"allow_network"`
	MaxMemory          string     `json:"max_memory"`
	MaxCPU             int        `json:"max_cpu"`
	CreatedBy          string     `json:"created_by,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	StartedAt          *time.Time `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
}

type Worker struct {
	ID            string    `json:"id"`
	GroupName     string    `json:"group_name"`
	Type          string    `json:"type"`         // local, remote
	Capabilities  string    `json:"capabilities"`  // JSON string
	Status        string    `json:"status"`       // active, inactive
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
}

type TaskLog struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	WorkerID  string    `json:"worker_id"`
	Level     string    `json:"level"`   // info, warning, error
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupStat struct {
	GroupName      string `json:"group_name"`
	TotalTasks     int    `json:"total_tasks"`
	PendingTasks   int    `json:"pending_tasks"`
	RunningTasks   int    `json:"running_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
	FailedTasks    int    `json:"failed_tasks"`
}

// TaskNode 任务节点（用于层次结构）
type TaskNode struct {
	Task      *Task
	Children  []*TaskNode
	Depth     int
}
