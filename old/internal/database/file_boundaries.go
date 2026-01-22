package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

)

// FileBoundary represents a file editing boundary for a worker
type FileBoundary struct {
	ID          int64     `json:"id"`
	TaskID      string    `json:"task_id"`
	WorkerID    string    `json:"worker_id"`
	FilePattern string    `json:"file_pattern"`      // e.g., "src/**/*.go"
	LineRange   string    `json:"line_range"`        // e.g., "1-100" or "1-50,60-100"
	AccessType  string    `json:"access_type"`       // "exclusive" | "shared" | "readonly"
	CreatedAt   time.Time `json:"created_at"`
}

// EditConflict represents an editing conflict
type EditConflict struct {
	ID          int64     `json:"id"`
	ConflictID  string    `json:"conflict_id"`
	TaskID      string    `json:"task_id"`
	WorkerID    string    `json:"worker_id"`
	FilePath    string    `json:"file_path"`
	LineNumber  int       `json:"line_number"`
	ConflictType string   `json:"conflict_type"`   // "overlap" | "concurrent" | "upstream_changed"
	Description string    `json:"description"`
	Severity    string    `json:"severity"`        // "low" | "medium" | "high" | "critical"
	Status      string    `json:"status"`         // "pending" | "resolving" | "resolved" | "escalated"
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy  string    `json:"resolved_by,omitempty"`
	Resolution  string    `json:"resolution,omitempty"`
}

// ConflictResolution represents a conflict resolution record
type ConflictResolution struct {
	ID                int64     `json:"id"`
	ConflictID        string    `json:"conflict_id"`
	ResolverAgentID   string    `json:"resolver_agent_id"`
	ResolutionAction  string    `json:"resolution_action"`  // "merge" | "override" | "skip" | "manual"
	OriginalContent   string    `json:"original_content"`
	ResolvedContent   string    `json:"resolved_content"`
	Explanation       string    `json:"explanation"`
	ResolvedAt        time.Time `json:"resolved_at"`
}

// PrivilegedAgentTask represents a task for high-privileged agent
type PrivilegedAgentTask struct {
	ID          int64              `json:"id"`
	TaskID      string            `json:"task_id"`
	Priority    int               `json:"priority"`
	TaskType    string            `json:"task_type"`     // "conflict_resolution" | "upstream_sync" | "manual_intervention"
	Description string            `json:"description"`
	ConflictIDs []string          `json:"conflict_ids"`
	Status      string            `json:"status"`
	AssignedTo  string            `json:"assigned_to"`
	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// FileEditLog represents a file edit operation log
type FileEditLog struct {
	ID         int64     `json:"id"`
	TaskID     string    `json:"task_id"`
	WorkerID   string    `json:"worker_id"`
	FilePath   string    `json:"file_path"`
	LineStart  int       `json:"line_start"`
	LineEnd    int       `json:"line_end"`
	Operation  string    `json:"operation"`  // "edit" | "delete" | "insert"
	OldContent string    `json:"old_content"`
	NewContent string    `json:"new_content"`
	Timestamp  time.Time `json:"timestamp"`
}

// === File Boundary Management ===

// CreateFileBoundariesTable creates the file boundaries table if not exists
func (db *Database) CreateFileBoundariesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS file_boundaries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		worker_id TEXT NOT NULL,
		file_pattern TEXT NOT NULL,
		line_range TEXT NOT NULL,
		access_type TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(task_id, worker_id, file_pattern)
	);`

	_, err := db.db.Exec(query)
	return err
}

// CreateEditConflictsTable creates the edit conflicts table
func (db *Database) CreateEditConflictsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS edit_conflicts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conflict_id TEXT NOT NULL UNIQUE,
		task_id TEXT NOT NULL,
		worker_id TEXT NOT NULL,
		file_path TEXT NOT NULL,
		line_number INTEGER NOT NULL,
		conflict_type TEXT NOT NULL,
		description TEXT,
		severity TEXT DEFAULT 'medium',
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		resolved_at TIMESTAMP,
		resolved_by TEXT,
		resolution TEXT
	);`

	_, err := db.db.Exec(query)
	return err
}

// AssignFileBoundaries assigns file boundaries to a worker
func (db *Database) AssignFileBoundaries(taskID, workerID string, boundaries []FileBoundary) error {
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO file_boundaries (task_id, worker_id, file_pattern, line_range, access_type)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(task_id, worker_id, file_pattern) DO UPDATE SET
			line_range = excluded.line_range,
			access_type = excluded.access_type
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, boundary := range boundaries {
		boundary.TaskID = taskID
		boundary.WorkerID = workerID

		_, err := stmt.Exec(boundary.TaskID, boundary.WorkerID, boundary.FilePattern,
			boundary.LineRange, boundary.AccessType)
		if err != nil {
			return fmt.Errorf("failed to insert boundary: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	db.logger.Infof("Assigned %d file boundaries to worker %s for task %s", len(boundaries), workerID, taskID)
	return nil
}

// GetFileBoundaries retrieves file boundaries for a worker
func (db *Database) GetFileBoundaries(taskID, workerID string) ([]FileBoundary, error) {
	query := `
		SELECT id, task_id, worker_id, file_pattern, line_range, access_type, created_at
		FROM file_boundaries
		WHERE task_id = ? AND worker_id = ?
		ORDER BY file_pattern
	`

	rows, err := db.db.Query(query, taskID, workerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file boundaries: %w", err)
	}
	defer rows.Close()

	var boundaries []FileBoundary
	for rows.Next() {
		var b FileBoundary
		err := rows.Scan(&b.ID, &b.TaskID, &b.WorkerID, &b.FilePattern,
			&b.LineRange, &b.AccessType, &b.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		boundaries = append(boundaries, b)
	}

	return boundaries, nil
}

// CheckFileOverlap checks if a file edit would overlap with another worker's boundary
func (db *Database) CheckFileOverlap(taskID, workerID, filePath string, lineNumber int) (bool, string, error) {
	query := `
		SELECT worker_id, file_pattern, line_range, access_type
		FROM file_boundaries
		WHERE task_id = ? AND worker_id != ?
	`

	rows, err := db.db.Query(query, taskID, workerID)
	if err != nil {
		return false, "", fmt.Errorf("failed to check overlap: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var otherWorker, pattern, lineRange, accessType string
		if err := rows.Scan(&otherWorker, &pattern, &lineRange, &accessType); err != nil {
			continue
		}

		// Simple pattern matching (can be improved with glob patterns)
		if matchesPattern(filePath, pattern) && isLineInRange(lineNumber, lineRange) {
			if accessType == "exclusive" {
				return true, otherWorker, nil
			}
		}
	}

	return false, "", nil
}

// matchesPattern checks if a file path matches a pattern
func matchesPattern(filePath, pattern string) bool {
	// Simplified pattern matching
	// TODO: Implement proper glob pattern matching
	return true
}

// isLineInRange checks if a line number is within a range
func isLineInRange(lineNumber int, lineRange string) bool {
	// Parse range like "1-100" or "1-50,60-100"
	// For now, return true as a placeholder
	return true
}

// === Conflict Management ===

// ReportConflict reports an editing conflict
func (db *Database) ReportConflict(conflict *EditConflict) error {
	conflict.Status = "pending"
	conflict.CreatedAt = time.Now()

	query := `
		INSERT INTO edit_conflicts
		(conflict_id, task_id, worker_id, file_path, line_number, conflict_type, description, severity, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.db.Exec(query,
		conflict.ConflictID, conflict.TaskID, conflict.WorkerID,
		conflict.FilePath, conflict.LineNumber, conflict.ConflictType,
		conflict.Description, conflict.Severity, conflict.Status)

	if err != nil {
		return fmt.Errorf("failed to report conflict: %w", err)
	}

	db.logger.Warnf("Conflict reported: %s - %s at %s:%d (severity: %s)",
		conflict.ConflictID, conflict.ConflictType, conflict.FilePath, conflict.LineNumber, conflict.Severity)

	return nil
}

// GetPendingConflicts retrieves all pending conflicts
func (db *Database) GetPendingConflicts() ([]EditConflict, error) {
	query := `
		SELECT id, conflict_id, task_id, worker_id, file_path, line_number,
		       conflict_type, description, severity, status, created_at
		FROM edit_conflicts
		WHERE status = 'pending'
		ORDER BY severity DESC, created_at ASC
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []EditConflict
	for rows.Next() {
		var c EditConflict
		err := rows.Scan(&c.ID, &c.ConflictID, &c.TaskID, &c.WorkerID,
			&c.FilePath, &c.LineNumber, &c.ConflictType, &c.Description,
			&c.Severity, &c.Status, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conflict: %w", err)
		}
		conflicts = append(conflicts, c)
	}

	return conflicts, nil
}

// ResolveConflict resolves a conflict
func (db *Database) ResolveConflict(conflictID, resolverID, resolution string, resolutionRecord *ConflictResolution) error {
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update conflict status
	now := time.Now()
	_, err = tx.Exec(`
		UPDATE edit_conflicts
		SET status = 'resolved', resolved_at = ?, resolved_by = ?, resolution = ?
		WHERE conflict_id = ?
	`, now, resolverID, resolution, conflictID)

	if err != nil {
		return fmt.Errorf("failed to update conflict: %w", err)
	}

	// Insert resolution record
	if resolutionRecord != nil {
		resolutionRecord.ResolvedAt = now
		_, err = tx.Exec(`
			INSERT INTO conflict_resolutions
			(conflict_id, resolver_agent_id, resolution_action, original_content, resolved_content, explanation, resolved_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, resolutionRecord.ConflictID, resolutionRecord.ResolverAgentID,
			resolutionRecord.ResolutionAction, resolutionRecord.OriginalContent,
			resolutionRecord.ResolvedContent, resolutionRecord.Explanation, resolutionRecord.ResolvedAt)

		if err != nil {
			return fmt.Errorf("failed to insert resolution: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	db.logger.Infof("Conflict %s resolved by %s: %s", conflictID, resolverID, resolution)
	return nil
}

// === Privileged Agent Tasks ===

// CreatePrivilegedAgentTask creates a task for privileged agent
func (db *Database) CreatePrivilegedAgentTask(task *PrivilegedAgentTask) error {
	task.Status = "pending"
	task.CreatedAt = time.Now()

	conflictIDsJSON, _ := json.Marshal(task.ConflictIDs)

	query := `
		INSERT INTO privileged_agent_tasks
		(task_id, priority, task_type, description, conflict_ids, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.db.Exec(query, task.TaskID, task.Priority, task.TaskType,
		task.Description, string(conflictIDsJSON), task.Status)

	if err != nil {
		return fmt.Errorf("failed to create privileged task: %w", err)
	}

	db.logger.Infof("Created privileged agent task: %s (type: %s, priority: %d)",
		task.TaskID, task.TaskType, task.Priority)

	return nil
}

// GetPrivilegedAgentTask retrieves a privileged agent task
func (db *Database) GetPrivilegedAgentTask(taskID string) (*PrivilegedAgentTask, error) {
	query := `
		SELECT id, task_id, priority, task_type, description, conflict_ids,
		       status, assigned_to, created_at, started_at, completed_at
		FROM privileged_agent_tasks
		WHERE task_id = ?
	`

	row := db.db.QueryRow(query, taskID)

	var task PrivilegedAgentTask
	var conflictIDsJSON string
	var startedAt, completedAt sql.NullTime

	err := row.Scan(&task.ID, &task.TaskID, &task.Priority, &task.TaskType,
		&task.Description, &conflictIDsJSON, &task.Status, &task.AssignedTo,
		&task.CreatedAt, &startedAt, &completedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to scan privileged task: %w", err)
	}

	json.Unmarshal([]byte(conflictIDsJSON), &task.ConflictIDs)

	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

// GetHighestPriorityPrivilegedTask gets the highest priority pending privileged task
func (db *Database) GetHighestPriorityPrivilegedTask() (*PrivilegedAgentTask, error) {
	query := `
		SELECT id, task_id, priority, task_type, description, conflict_ids,
		       status, assigned_to, created_at
		FROM privileged_agent_tasks
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT 1
	`

	row := db.db.QueryRow(query)

	var task PrivilegedAgentTask
	var conflictIDsJSON string

	err := row.Scan(&task.ID, &task.TaskID, &task.Priority, &task.TaskType,
		&task.Description, &conflictIDsJSON, &task.Status, &task.AssignedTo, &task.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No pending tasks
		}
		return nil, fmt.Errorf("failed to scan privileged task: %w", err)
	}

	json.Unmarshal([]byte(conflictIDsJSON), &task.ConflictIDs)

	return &task, nil
}

// UpdatePrivilegedTaskStatus updates the status of a privileged task
func (db *Database) UpdatePrivilegedTaskStatus(taskID, status string, assignedTo string) error {
	query := `
		UPDATE privileged_agent_tasks
		SET status = ?, assigned_to = ?
		WHERE task_id = ?
	`

	_, err := db.db.Exec(query, status, assignedTo, taskID)
	if err != nil {
		return fmt.Errorf("failed to update privileged task status: %w", err)
	}

	return nil
}

// === File Edit Logging ===

// LogFileEdit logs a file edit operation
func (db *Database) LogFileEdit(log *FileEditLog) error {
	log.Timestamp = time.Now()

	query := `
		INSERT INTO file_edit_log
		(task_id, worker_id, file_path, line_start, line_end, operation, old_content, new_content, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.db.Exec(query, log.TaskID, log.WorkerID, log.FilePath,
		log.LineStart, log.LineEnd, log.Operation, log.OldContent, log.NewContent, log.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to log file edit: %w", err)
	}

	return nil
}

// GetFileEditHistory retrieves edit history for a file
func (db *Database) GetFileEditHistory(filePath string, limit int) ([]FileEditLog, error) {
	query := `
		SELECT id, task_id, worker_id, file_path, line_start, line_end,
		       operation, old_content, new_content, timestamp
		FROM file_edit_log
		WHERE file_path = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.db.Query(query, filePath, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query edit history: %w", err)
	}
	defer rows.Close()

	var logs []FileEditLog
	for rows.Next() {
		var log FileEditLog
		err := rows.Scan(&log.ID, &log.TaskID, &log.WorkerID, &log.FilePath,
			&log.LineStart, &log.LineEnd, &log.Operation, &log.OldContent,
			log.NewContent, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}
