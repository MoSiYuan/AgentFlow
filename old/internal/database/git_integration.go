package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jiangxiaolong/cpds-go/internal/git"
)

// GitFileBoundary represents a file boundary for Git integration
type GitFileBoundary struct {
	FilePath string `json:"file_path"`  // e.g., "src/controllers/user.go"
}

// GitTask represents a task with Git integration
type GitTask struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	WorkerID      string          `json:"worker_id"`
	GitBranch     string          `json:"git_branch"`       // Git分支名，如 "worker-1/task-001"
	FileBoundaries []GitFileBoundary `json:"file_boundaries"` // 文件边界定义
	Locks         []GitLock       `json:"locks"`            // Git文件锁
	Status        string          `json:"status"`           // pending | assigned | in_progress | completed
	CreatedAt     time.Time       `json:"created_at"`
	AssignedAt     *time.Time     `json:"assigned_at,omitempty"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
}

// GitLock represents a Git file lock record
type GitLock struct {
	ID         int64      `json:"id"`
	TaskID     string     `json:"task_id"`
	WorkerID   string     `json:"worker_id"`
	FilePath   string     `json:"file_path"`
	LockType   string     `json:"lock_type"`   // "read" | "write"
	AcquiredAt time.Time  `json:"acquired_at"`
	ReleasedAt *time.Time `json:"released_at,omitempty"`
	Status     string     `json:"status"`       // "active" | "released"
}

// GitConflict represents a Git merge conflict
type GitConflict struct {
	ID            string     `json:"id"`
	TaskID        string     `json:"task_id"`
	WorkerID      string     `json:"worker_id"`
	ConflictType  string     `json:"conflict_type"` // "git_merge" | "file_locked" | "boundary_overlap"`
	FilePaths     []string   `json:"file_paths"`
	Description   string     `json:"description"`
	Severity      string     `json:"severity"`      // "low" | "medium" | "high" | "critical"`
	Status        string     `json:"status"`        // "pending" | "resolving" | "resolved"
	CreatedAt     time.Time  `json:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy    string     `json:"resolved_by,omitempty"`
	Resolution    string     `json:"resolution,omitempty"`
}

// === Git Integrated Task Management ===

// CreateGitTask creates a Git task record in the database
func (db *Database) CreateGitTask(task *GitTask) error {
	boundariesJSON, _ := json.Marshal(task.FileBoundaries)

	query := `
		INSERT INTO git_tasks (id, title, description, git_branch, file_boundaries, status)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			git_branch = excluded.git_branch,
			file_boundaries = excluded.file_boundaries,
			status = excluded.status
	`

	_, err := db.db.Exec(query, task.ID, task.Title, task.Description,
		task.GitBranch, string(boundariesJSON), task.Status)

	if err != nil {
		return fmt.Errorf("failed to create git task: %w", err)
	}

	db.logger.Infof("Git task created: %s (branch: %s)", task.ID, task.GitBranch)
	return nil
}

// CreateGitTables creates Git-integrated tables
func (db *Database) CreateGitTables() error {
	schema := `
	-- Git tasks table
	CREATE TABLE IF NOT EXISTS git_tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		worker_id TEXT,
		git_branch TEXT NOT NULL UNIQUE,
		file_boundaries TEXT,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		assigned_at TIMESTAMP,
		completed_at TIMESTAMP
	);

	-- Git locks table
	CREATE TABLE IF NOT EXISTS git_locks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		worker_id TEXT NOT NULL,
		file_path TEXT NOT NULL,
		lock_type TEXT NOT NULL,
		acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		released_at TIMESTAMP,
		status TEXT DEFAULT 'active',
		FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
	);

	-- Git conflicts table
	CREATE TABLE IF NOT EXISTS git_conflicts (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		worker_id TEXT NOT NULL,
		conflict_type TEXT NOT NULL,
		file_paths TEXT,
		description TEXT,
		severity TEXT DEFAULT 'medium',
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		resolved_at TIMESTAMP,
		resolved_by TEXT,
		resolution TEXT,
		FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_git_tasks_status ON git_tasks(status);
	CREATE INDEX IF NOT EXISTS idx_git_tasks_worker ON git_tasks(worker_id);
	CREATE INDEX IF NOT EXISTS idx_git_locks_file ON git_locks(file_path);
	CREATE INDEX IF NOT EXISTS idx_git_locks_status ON git_locks(status);
	CREATE INDEX IF NOT EXISTS idx_git_conflicts_status ON git_conflicts(status);
	`

	_, err := db.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create Git tables: %w", err)
	}

	db.logger.Info("Git-integrated tables created successfully")
	return nil
}

// AssignTaskWithGit assigns a task with Git integration
func (db *Database) AssignTaskWithGit(task *GitTask, gitClient *git.GitClient) error {
	// 1. Create Git branch
	branchName := task.GitBranch
	if err := gitClient.CreateBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	// 2. Lock files in Git
	for _, boundary := range task.FileBoundaries {
		if err := gitClient.LockFile(boundary.FilePath, task.WorkerID); err != nil {
			// Rollback: delete branch and unlock all files
			gitClient.DeleteBranch(branchName, true)
			return fmt.Errorf("failed to lock file %s: %w", boundary.FilePath, err)
		}
	}

	// 3. Store in SQLite
	task.Status = "assigned"
	now := time.Now()
	task.AssignedAt = &now

	boundariesJSON, _ := json.Marshal(task.FileBoundaries)

	query := `
		INSERT INTO git_tasks (id, title, description, worker_id, git_branch, file_boundaries, status, assigned_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			worker_id = excluded.worker_id,
			git_branch = excluded.git_branch,
			file_boundaries = excluded.file_boundaries,
			status = excluded.status,
			assigned_at = excluded.assigned_at
	`

	_, err := db.db.Exec(query, task.ID, task.Title, task.Description,
		task.WorkerID, task.GitBranch, string(boundariesJSON), task.Status, task.AssignedAt)

	if err != nil {
		return fmt.Errorf("failed to store git task: %w", err)
	}

	db.logger.Infof("Task assigned with Git: %s -> branch %s, %d files locked",
		task.ID, branchName, len(task.FileBoundaries))

	return nil
}

// GetGitTask retrieves a Git task
func (db *Database) GetGitTask(taskID string) (*GitTask, error) {
	query := `
		SELECT id, title, description, worker_id, git_branch, file_boundaries, status, created_at, assigned_at, completed_at
		FROM git_tasks
		WHERE id = ?
	`

	row := db.db.QueryRow(query, taskID)

	var task GitTask
	var workerID sql.NullString
	var boundariesJSON sql.NullString
	var assignedAt, completedAt sql.NullTime

	err := row.Scan(&task.ID, &task.Title, &task.Description, &workerID,
		&task.GitBranch, &boundariesJSON, &task.Status, &task.CreatedAt,
		&assignedAt, &completedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Task not found
		}
		return nil, fmt.Errorf("failed to scan git task: %w", err)
	}

	if workerID.Valid {
		task.WorkerID = workerID.String
	}

	if assignedAt.Valid {
		task.AssignedAt = &assignedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	// Parse file boundaries
	if boundariesJSON.Valid {
		json.Unmarshal([]byte(boundariesJSON.String), &task.FileBoundaries)
	}

	// Load locks
	locks, err := db.GetGitLocksForTask(taskID)
	if err == nil {
		task.Locks = locks
	}

	return &task, nil
}

// UpdateGitTaskStatus updates task status
func (db *Database) UpdateGitTaskStatus(taskID, status string) error {
	now := time.Now()
	var completedAt sql.NullTime

	if status == "completed" {
		completedAt.Valid = true
		completedAt.Time = now
	}

	query := `
		UPDATE git_tasks
		SET status = ?, completed_at = ?
		WHERE id = ?
	`

	_, err := db.db.Exec(query, status, completedAt, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	db.logger.Infof("Git task %s status updated to %s", taskID, status)
	return nil
}

// GetGitLocksForTask retrieves all locks for a task
func (db *Database) GetGitLocksForTask(taskID string) ([]GitLock, error) {
	query := `
		SELECT id, task_id, worker_id, file_path, lock_type, acquired_at, released_at, status
		FROM git_locks
		WHERE task_id = ? AND status = 'active'
		ORDER BY acquired_at ASC
	`

	rows, err := db.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query git locks: %w", err)
	}
	defer rows.Close()

	var locks []GitLock
	for rows.Next() {
		var lock GitLock
		var releasedAt sql.NullTime

		err := rows.Scan(&lock.ID, &lock.TaskID, &lock.WorkerID, &lock.FilePath,
			&lock.LockType, &lock.AcquiredAt, &releasedAt, &lock.Status)

		if err != nil {
			return nil, fmt.Errorf("failed to scan lock: %w", err)
		}

		if releasedAt.Valid {
			lock.ReleasedAt = &releasedAt.Time
		}

		locks = append(locks, lock)
	}

	return locks, nil
}

// ReleaseAllTaskLocks releases all locks for a task
func (db *Database) ReleaseAllTaskLocks(taskID string, gitClient *git.GitClient) error {
	locks, err := db.GetGitLocksForTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get locks: %w", err)
	}

	released := 0
	for _, lock := range locks {
		if err := gitClient.UnlockFile(lock.FilePath); err == nil {
			// Update database
			now := time.Now()
			_, err := db.db.Exec(`
				UPDATE git_locks
				SET status = 'released', released_at = ?
				WHERE id = ?
			`, now, lock.ID)

			if err == nil {
				released++
			}
		}
	}

	db.logger.Infof("Released %d/%d locks for task %s", released, len(locks), taskID)
	return nil
}

// ReportGitConflict reports a Git merge conflict
func (db *Database) ReportGitConflict(conflict *GitConflict) error {
	conflict.Status = "pending"
	conflict.CreatedAt = time.Now()

	filePathsJSON, _ := json.Marshal(conflict.FilePaths)

	query := `
		INSERT INTO git_conflicts
		(id, task_id, worker_id, conflict_type, file_paths, description, severity, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status = excluded.status,
			severity = excluded.severity
	`

	_, err := db.db.Exec(query, conflict.ID, conflict.TaskID, conflict.WorkerID,
		conflict.ConflictType, string(filePathsJSON), conflict.Description,
		conflict.Severity, conflict.Status, conflict.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to report conflict: %w", err)
	}

	db.logger.Warnf("Git conflict reported: %s - %s (severity: %s)",
		conflict.ID, conflict.ConflictType, conflict.Severity)

	// Create privileged agent task if critical or high severity
	if conflict.Severity == "high" || conflict.Severity == "critical" {
		privTask := &PrivilegedAgentTask{
			TaskID:     fmt.Sprintf("CONFLICT-%s", conflict.ID),
			Priority:   100,
			TaskType:   "conflict_resolution",
			Description: fmt.Sprintf("Resolve %s conflict for task %s", conflict.ConflictType, conflict.TaskID),
			ConflictIDs: []string{conflict.ID},
			Status:     "pending",
		}

		if err := db.CreatePrivilegedAgentTask(privTask); err != nil {
			db.logger.Errorf("Failed to create privileged task: %v", err)
		}
	}

	return nil
}

// ResolveGitConflict resolves a Git conflict
func (db *Database) ResolveGitConflict(conflictID, resolverID, resolution string, resolutionRecord *ConflictResolution) error {
	now := time.Now()

	// Update conflict status
	_, err := db.db.Exec(`
		UPDATE git_conflicts
		SET status = 'resolved', resolved_at = ?, resolved_by = ?, resolution = ?
		WHERE id = ?
	`, now, resolverID, resolution, conflictID)

	if err != nil {
		return fmt.Errorf("failed to update conflict: %w", err)
	}

	// Insert resolution record if provided
	if resolutionRecord != nil {
		resolutionRecord.ResolvedAt = now

		query := `
			INSERT INTO conflict_resolutions
			(conflict_id, resolver_agent_id, resolution_action, original_content, resolved_content, explanation, resolved_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`

		_, err = db.db.Exec(query, resolutionRecord.ConflictID, resolutionRecord.ResolverAgentID,
			resolutionRecord.ResolutionAction, resolutionRecord.OriginalContent,
			resolutionRecord.ResolvedContent, resolutionRecord.Explanation, resolutionRecord.ResolvedAt)

		if err != nil {
			return fmt.Errorf("failed to insert resolution: %w", err)
		}
	}

	db.logger.Infof("Git conflict %s resolved by %s: %s", conflictID, resolverID, resolution)
	return nil
}

// GetPendingGitConflicts retrieves all pending Git conflicts
func (db *Database) GetPendingGitConflicts() ([]GitConflict, error) {
	query := `
		SELECT id, task_id, worker_id, conflict_type, file_paths, description, severity, status, created_at
		FROM git_conflicts
		WHERE status = 'pending'
		ORDER BY severity DESC, created_at ASC
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []GitConflict
	for rows.Next() {
		var c GitConflict
		var filePathsJSON sql.NullString

		err := rows.Scan(&c.ID, &c.TaskID, &c.WorkerID, &c.ConflictType, &filePathsJSON,
			&c.Description, &c.Severity, &c.Status, &c.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan conflict: %w", err)
		}

		if filePathsJSON.Valid {
			json.Unmarshal([]byte(filePathsJSON.String), &c.FilePaths)
		}

		conflicts = append(conflicts, c)
	}

	return conflicts, nil
}

// GetWorkerActiveGitTasks retrieves all active Git tasks for a worker
func (db *Database) GetWorkerActiveGitTasks(workerID string) ([]GitTask, error) {
	query := `
		SELECT id, title, description, worker_id, git_branch, file_boundaries, status, created_at, assigned_at, completed_at
		FROM git_tasks
		WHERE worker_id = ? AND status IN ('assigned', 'in_progress')
		ORDER BY assigned_at DESC
	`

	rows, err := db.db.Query(query, workerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query worker tasks: %w", err)
	}
	defer rows.Close()

	var tasks []GitTask
	for rows.Next() {
		var task GitTask
		var boundariesJSON sql.NullString
		var assignedAt, completedAt sql.NullTime

		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.WorkerID,
			&task.GitBranch, &boundariesJSON, &task.Status, &task.CreatedAt,
			&assignedAt, &completedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if assignedAt.Valid {
			task.AssignedAt = &assignedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		// Parse file boundaries
		if boundariesJSON.Valid {
			json.Unmarshal([]byte(boundariesJSON.String), &task.FileBoundaries)
		}

		// Load locks
		locks, err := db.GetGitLocksForTask(task.ID)
		if err == nil {
			task.Locks = locks
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// CleanupCompletedGitTasks cleans up completed Git tasks
func (db *Database) CleanupCompletedGitTasks(gitClient *git.GitClient) error {
	// Get all completed tasks older than 1 hour
	query := `
		SELECT id, git_branch, worker_id
		FROM git_tasks
		WHERE status = 'completed' AND completed_at < datetime('now', '-1 hour')
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query completed tasks: %w", err)
	}
	defer rows.Close()

	cleaned := 0
	for rows.Next() {
		var taskID, branchName, workerID string
		if err := rows.Scan(&taskID, &branchName, &workerID); err != nil {
			continue
		}

		// Delete branch
		if err := gitClient.DeleteBranch(branchName, false); err != nil {
			db.logger.Warnf("Failed to delete branch %s: %v", branchName, err)
			continue
		}

		// Delete from database
		_, err = db.db.Exec("DELETE FROM git_tasks WHERE id = ?", taskID)
		if err != nil {
			db.logger.Warnf("Failed to delete task %s: %v", taskID, err)
			continue
		}

		cleaned++
		db.logger.Infof("Cleaned up completed task: %s (branch %s)", taskID, branchName)
	}

	if cleaned > 0 {
		db.logger.Infof("Cleaned up %d completed Git tasks", cleaned)
	}

	return nil
}
