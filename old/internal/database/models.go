package database

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONField is a helper type for storing JSON in SQLite
type JSONField json.RawMessage

// Scan implements sql.Scanner interface
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	*j = JSONField(bytes)
	return nil
}

// Value implements driver.Valuer interface
func (j JSONField) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// String returns the JSON string
func (j JSONField) String() string {
	return string(json.RawMessage(j))
}

// Worker represents a worker in the system
type Worker struct {
	WorkerID     string     `json:"worker_id" db:"worker_id"`
	WorkerName   string     `json:"worker_name" db:"worker_name"`
	Platform     string     `json:"platform" db:"platform"`
	Capabilities JSONField  `json:"capabilities" db:"capabilities"` // JSON array
	Status       string     `json:"status" db:"status"`             // online, offline, busy, idle
	LastHeartbeat *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	RegisteredAt time.Time  `json:"registered_at" db:"registered_at"`
	Metadata     JSONField  `json:"metadata" db:"metadata"` // JSON object
}

// Task represents a task in the system
type Task struct {
	TaskID          string     `json:"task_id" db:"task_id"`
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description" db:"description"`
	AssignedTo      *string    `json:"assigned_to" db:"assigned_to"` // worker_id
	Status          string     `json:"status" db:"status"`           // pending, assigned, in_progress, completed, failed
	Priority        string     `json:"priority" db:"priority"`       // high, medium, low
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	StartedAt       *time.Time `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
	Progress        int        `json:"progress" db:"progress"` // 0-100
	Output          *string    `json:"output" db:"output"`
	Dependencies    JSONField  `json:"dependencies" db:"dependencies"`   // JSON array
	Tags            JSONField  `json:"tags" db:"tags"`                   // JSON array
	DeploymentMode  string     `json:"deployment_mode" db:"deployment_mode"` // cloud, standalone, both
	CreatedBy       *string    `json:"created_by" db:"created_by"`       // user ID
	ClaudeContext   *string    `json:"claude_context" db:"claude_context"` // Claude context for task execution
}

// TaskExecutionRecord represents a record of task execution
type TaskExecutionRecord struct {
	ID            int64      `json:"id" db:"id"`
	TaskID        string     `json:"task_id" db:"task_id"`
	WorkerID      string     `json:"worker_id" db:"worker_id"`
	Mode          string     `json:"mode" db:"mode"` // cloud, standalone
	ClaudeCallID  *string    `json:"claude_call_id" db:"claude_call_id"`
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	DurationMs    *int64     `json:"duration_ms" db:"duration_ms"`
	Status        string     `json:"status" db:"status"` // running, completed, failed
	InputPrompt   *string    `json:"input_prompt" db:"input_prompt"`
	Output        *string    `json:"output" db:"output"`
	Error         *string    `json:"error" db:"error"`
	TokensUsed    *int       `json:"tokens_used" db:"tokens_used"`
	Metadata      JSONField  `json:"metadata" db:"metadata"` // JSON object
}

// ProgressHistory represents a progress history event
type ProgressHistory struct {
	ID         int64     `json:"id" db:"id"`
	WorkerID   string    `json:"worker_id" db:"worker_id"`
	TaskID     string    `json:"task_id" db:"task_id"`
	EventType  string    `json:"event_type" db:"event_type"` // task_assigned, progress_update, task_completed, task_failed
	Message    *string   `json:"message" db:"message"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
}

// WorkerStatus represents the current status of a worker
type WorkerStatus struct {
	WorkerID      string       `json:"worker_id"`
	WorkerName    string       `json:"worker_name"`
	Platform      string       `json:"platform"`
	Capabilities []string     `json:"capabilities"`
	Status        string       `json:"status"`        // online, offline, busy, idle
	CurrentTask   *TaskSummary `json:"current_task"`  // nil if idle
	LastHeartbeat time.Time    `json:"last_heartbeat"`
	TasksCompleted int         `json:"tasks_completed"`
}

// TaskSummary represents a summary of a task
type TaskSummary struct {
	TaskID    string `json:"task_id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	Priority  string `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	PendingTasks      int `json:"pending_tasks"`
	InProgressTasks   int `json:"in_progress_tasks"`
	CompletedTasks    int `json:"completed_tasks"`
	FailedTasks       int `json:"failed_tasks"`
	OnlineWorkers     int `json:"online_workers"`
	IdleWorkers       int `json:"idle_workers"`
	BusyWorkers       int `json:"busy_workers"`
}

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	TaskID         string   `json:"task_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	Tags           []string `json:"tags"`
	Dependencies   []string `json:"dependencies"`
	DeploymentMode string   `json:"deployment_mode"` // cloud, standalone, both
	CreatedBy      string   `json:"created_by"`
	ClaudeContext  string   `json:"claude_context"`
}

// RegisterWorkerRequest represents a request to register a worker
type RegisterWorkerRequest struct {
	WorkerID     string                 `json:"worker_id"`
	WorkerName   string                 `json:"worker_name"`
	Platform     string                 `json:"platform"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// UpdateProgressRequest represents a request to update task progress
type UpdateProgressRequest struct {
	TaskID   string  `json:"task_id"`
	Progress int     `json:"progress"` // 0-100
	Message  *string `json:"message"`
}

// CompleteTaskRequest represents a request to complete a task
type CompleteTaskRequest struct {
	TaskID string  `json:"task_id"`
	Output *string `json:"output"`
	Error  *string `json:"error"`
}

// AssignTaskRequest represents a request to assign a task
type AssignTaskRequest struct {
	TaskID   string `json:"task_id"`
	WorkerID string `json:"worker_id"`
}
