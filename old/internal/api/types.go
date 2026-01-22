package api

import "github.com/jiangxiaolong/cpds-go/internal/database"

// Response is a standard API response wrapper
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// WorkerListResponse is the response for GET /api/workers
type WorkerListResponse struct {
	Workers []*database.WorkerStatus `json:"workers"`
}

// TaskListResponse is the response for GET /api/tasks/*
type TaskListResponse struct {
	Tasks []*database.Task `json:"tasks"`
	Total int              `json:"total"`
}

// TaskPendingResponse is the response for GET /api/tasks/pending
type TaskPendingResponse struct {
	Tasks []*database.Task `json:"tasks"`
}

// RegisterWorkerResponse is the response for POST /api/workers/register
type RegisterWorkerResponse struct {
	WorkerID string `json:"worker_id"`
	Message  string `json:"message"`
}

// AssignTaskResponse is the response for POST /api/tasks/assign
type AssignTaskResponse struct {
	TaskID   string `json:"task_id"`
	WorkerID string `json:"worker_id"`
	Message  string `json:"message"`
}

// StatusResponse is the response for GET /api/status
type StatusResponse struct {
	*database.SystemStatus
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// WorkerTasksResponse is the response for GET /api/workers/tasks
type WorkerTasksResponse struct {
	WorkerID   string             `json:"worker_id"`
	CurrentTask *database.Task     `json:"current_task,omitempty"`
	Completed  []*database.Task    `json:"completed,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err string) *Response {
	return &Response{
		Success: false,
		Error:   err,
	}
}

// NewMessageResponse creates a message response
func NewMessageResponse(message string) *Response {
	return &Response{
		Success: true,
		Message: message,
	}
}

// === Git-Integrated Task Types ===

// CreateGitTaskRequest is the request for POST /api/tasks/create-with-git
type CreateGitTaskRequest struct {
	TaskID        string                       `json:"task_id" binding:"required"`
	Title         string                       `json:"title" binding:"required"`
	Description   string                       `json:"description"`
	Priority      string                       `json:"priority"`
	FileBoundaries []GitFileBoundaryRequest    `json:"file_boundaries" binding:"required"`
	DeploymentMode string                      `json:"deployment_mode"`
}

// GitFileBoundaryRequest represents a file boundary in the request
type GitFileBoundaryRequest struct {
	FilePath  string `json:"file_path" binding:"required"`  // e.g., "src/controllers/user.go"
	LineStart int    `json:"line_start" binding:"required"`
	LineEnd   int    `json:"line_end" binding:"required"`
	LockType  string `json:"lock_type"`  // "read" | "write" (default: "write")
}

// CreateGitTaskResponse is the response for POST /api/tasks/create-with-git
type CreateGitTaskResponse struct {
	TaskID    string   `json:"task_id"`
	GitBranch string   `json:"git_branch"`
	Files     []string `json:"files_locked"`
	Message   string   `json:"message"`
}

// GetGitTaskResponse is the response for GET /api/tasks/:task_id/git
type GetGitTaskResponse struct {
	*database.GitTask
	ConflictedFiles []string `json:"conflicted_files,omitempty"`
}

// ConflictListResponse is the response for GET /api/conflicts
type ConflictListResponse struct {
	Conflicts []database.GitConflict `json:"conflicts"`
	Total     int                    `json:"total"`
}
