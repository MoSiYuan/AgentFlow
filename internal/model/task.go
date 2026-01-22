package model

import (
	"time"
)

// Task 任务模型
type Task struct {
	ID             string        `json:"id"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Context        string        `json:"context"`
	VerifyType     string        `json:"verify_type"`
	VerifyCommand  string        `json:"verify_command"`
	VerifyExpected string        `json:"verify_expected"`
	Priority       int           `json:"priority"`
	GroupName      string        `json:"group_name"`
	Status         string        `json:"status"` // pending, locked, running, completed, failed
	LockHolder     string        `json:"lock_holder"`
	LockTime       *time.Time    `json:"lock_time"`
	RetryCount     int           `json:"retry_count"`
	MaxRetries     int           `json:"max_retries"`
	Result         string        `json:"result"`
	Error          string        `json:"error"`
	CreatedAt      time.Time     `json:"created_at"`
	StartedAt      *time.Time    `json:"started_at"`
	CompletedAt    *time.Time    `json:"completed_at"`
}

// Worker Worker 模型
type Worker struct {
	ID            string            `json:"id"`
	GroupName     string            `json:"group_name"`
	Type          string            `json:"type"` // local, remote
	Status        string            `json:"status"` // idle, busy, offline
	Capabilities map[string]interface{} `json:"capabilities"`
	LastHeartbeat *time.Time        `json:"last_heartbeat"`
	Metadata      string            `json:"metadata"`
	CreatedAt     time.Time         `json:"created_at"`
}

// TaskLog 任务日志
type TaskLog struct {
	ID        int       `json:"id"`
	TaskID    string    `json:"task_id"`
	WorkerID  string    `json:"worker_id"`
	GroupName string    `json:"group_name"`
	Level     string    `json:"level"` // INFO, WARN, ERROR
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// TaskResult 任务结果
type TaskResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	Duration   int64  `json:"duration_ms,omitempty"`
	CommitHash string `json:"commit_hash,omitempty"`
}

// GroupInfo 工作组信息
type GroupInfo struct {
	GroupName     string `json:"group_name"`
	IdleWorkers   int    `json:"idle_workers"`
	BusyWorkers   int    `json:"busy_workers"`
	TotalWorkers  int    `json:"total_workers"`
	PendingTasks  int    `json:"pending_tasks"`
}

// SystemStats 系统统计
type SystemStats struct {
	PendingTasks  int `json:"pending_tasks"`
	RunningTasks  int `json:"running_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks   int `json:"failed_tasks"`
	TotalWorkers  int `json:"total_workers"`
	BusyWorkers   int `json:"busy_workers"`
}

// TaskRequest 创建任务请求
type TaskRequest struct {
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	GroupName     string `json:"group_name"`
	VerifyCommand string `json:"verify_command"`
	VerifyExpected string `json:"verify_expected"`
	Priority      int    `json:"priority"`
}

// RegisterWorkerRequest 注册 Worker 请求
type RegisterWorkerRequest struct {
	ID           string                 `json:"id" binding:"required"`
	GroupName    string                 `json:"group_name" binding:"required"`
	Type         string                 `json:"type"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// LockTaskRequest 锁定任务请求
type LockTaskRequest struct {
	WorkerID string `json:"worker_id" binding:"required"`
}
