package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// Database wraps the SQLite database connection
type Database struct {
	db     *sql.DB
	logger *logrus.Logger
}

// New creates a new database instance
func New(dbPath string, logger *logrus.Logger) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	database := &Database{
		db:     db,
		logger: logger,
	}

	// Initialize schema
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.Infof("Database initialized: %s", dbPath)

	return database, nil
}

// initSchema creates the database schema
func (d *Database) initSchema() error {
	schema := `
	-- Workers table
	CREATE TABLE IF NOT EXISTS workers (
		worker_id TEXT PRIMARY KEY,
		worker_name TEXT NOT NULL,
		platform TEXT NOT NULL,
		capabilities TEXT,
		status TEXT DEFAULT 'offline',
		last_heartbeat TIMESTAMP,
		registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT
	);

	-- Tasks table
	CREATE TABLE IF NOT EXISTS tasks (
		task_id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		assigned_to TEXT,
		status TEXT DEFAULT 'pending',
		priority TEXT DEFAULT 'medium',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		started_at TIMESTAMP,
		completed_at TIMESTAMP,
		progress INTEGER DEFAULT 0,
		output TEXT,
		dependencies TEXT,
		tags TEXT,
		deployment_mode TEXT DEFAULT 'both',
		created_by TEXT,
		claude_context TEXT,
		FOREIGN KEY (assigned_to) REFERENCES workers(worker_id)
	);

	-- Task execution records table
	CREATE TABLE IF NOT EXISTS task_execution_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		worker_id TEXT NOT NULL,
		mode TEXT NOT NULL,
		claude_call_id TEXT,
		started_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		duration_ms INTEGER,
		status TEXT NOT NULL,
		input_prompt TEXT,
		output TEXT,
		error TEXT,
		tokens_used INTEGER,
		metadata TEXT,
		FOREIGN KEY (task_id) REFERENCES tasks(task_id)
	);

	-- Progress history table
	CREATE TABLE IF NOT EXISTS progress_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		worker_id TEXT NOT NULL,
		task_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		message TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (worker_id) REFERENCES workers(worker_id),
		FOREIGN KEY (task_id) REFERENCES tasks(task_id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_assigned ON tasks(assigned_to);
	CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
	CREATE INDEX IF NOT EXISTS idx_records_task ON task_execution_records(task_id);
	CREATE INDEX IF NOT EXISTS idx_records_worker ON task_execution_records(worker_id);
	CREATE INDEX IF NOT EXISTS idx_records_time ON task_execution_records(started_at);
	CREATE INDEX IF NOT EXISTS idx_progress_task ON progress_history(task_id);
	CREATE INDEX IF NOT EXISTS idx_progress_worker ON progress_history(worker_id);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return err
	}

	// Create Git-integrated tables
	if err := d.CreateGitTables(); err != nil {
		return fmt.Errorf("failed to create git tables: %w", err)
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// GetDB returns the underlying *sql.DB
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// Worker operations

// RegisterWorker registers a new worker
func (d *Database) RegisterWorker(req *RegisterWorkerRequest) error {
	capabilitiesJSON, _ := json.Marshal(req.Capabilities)
	metadataJSON, _ := json.Marshal(req.Metadata)

	query := `
	INSERT INTO workers (worker_id, worker_name, platform, capabilities, status, registered_at, metadata)
	VALUES (?, ?, ?, ?, 'online', CURRENT_TIMESTAMP, ?)
	ON CONFLICT(worker_id) DO UPDATE SET
		worker_name = excluded.worker_name,
		platform = excluded.platform,
		capabilities = excluded.capabilities,
		status = 'online',
		metadata = excluded.metadata
	`

	_, err := d.db.Exec(query, req.WorkerID, req.WorkerName, req.Platform, capabilitiesJSON, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	d.logger.Infof("Worker registered: %s (%s)", req.WorkerName, req.WorkerID)
	return nil
}

// UpdateWorkerHeartbeat updates worker heartbeat timestamp
func (d *Database) UpdateWorkerHeartbeat(workerID string) error {
	query := `UPDATE workers SET last_heartbeat = CURRENT_TIMESTAMP WHERE worker_id = ?`
	_, err := d.db.Exec(query, workerID)
	return err
}

// GetWorker retrieves a worker by ID
func (d *Database) GetWorker(workerID string) (*Worker, error) {
	query := `
	SELECT worker_id, worker_name, platform, capabilities, status, last_heartbeat, registered_at, metadata
	FROM workers WHERE worker_id = ?
	`

	row := d.db.QueryRow(query, workerID)

	var w Worker
	err := row.Scan(
		&w.WorkerID, &w.WorkerName, &w.Platform, &w.Capabilities,
		&w.Status, &w.LastHeartbeat, &w.RegisteredAt, &w.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &w, nil
}

// GetOnlineWorkers retrieves all online workers
func (d *Database) GetOnlineWorkers() ([]*Worker, error) {
	query := `
	SELECT worker_id, worker_name, platform, capabilities, status, last_heartbeat, registered_at, metadata
	FROM workers WHERE status = 'online' OR status = 'busy' OR status = 'idle'
	ORDER BY registered_at DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workers []*Worker
	for rows.Next() {
		var w Worker
		err := rows.Scan(
			&w.WorkerID, &w.WorkerName, &w.Platform, &w.Capabilities,
			&w.Status, &w.LastHeartbeat, &w.RegisteredAt, &w.Metadata,
		)
		if err != nil {
			return nil, err
		}
		workers = append(workers, &w)
	}

	return workers, nil
}

// MarkWorkerOffline marks a worker as offline
func (d *Database) MarkWorkerOffline(workerID string) error {
	query := `UPDATE workers SET status = 'offline' WHERE worker_id = ?`
	_, err := d.db.Exec(query, workerID)
	return err
}

// Task operations

// CreateTask creates a new task
func (d *Database) CreateTask(req *CreateTaskRequest) error {
	tagsJSON, _ := json.Marshal(req.Tags)
	dependenciesJSON, _ := json.Marshal(req.Dependencies)

	query := `
	INSERT INTO tasks (task_id, title, description, priority, tags, dependencies, deployment_mode, created_by, claude_context)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(
		query,
		req.TaskID, req.Title, req.Description, req.Priority,
		tagsJSON, dependenciesJSON, req.DeploymentMode, req.CreatedBy, req.ClaudeContext,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	d.logger.Infof("Task created: %s (%s)", req.TaskID, req.Title)
	return nil
}

// GetTask retrieves a task by ID
func (d *Database) GetTask(taskID string) (*Task, error) {
	query := `
	SELECT task_id, title, description, assigned_to, status, priority,
	       created_at, started_at, completed_at, progress, output,
	       dependencies, tags, deployment_mode, created_by, claude_context
	FROM tasks WHERE task_id = ?
	`

	row := d.db.QueryRow(query, taskID)

	var t Task
	err := row.Scan(
		&t.TaskID, &t.Title, &t.Description, &t.AssignedTo, &t.Status, &t.Priority,
		&t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.Progress, &t.Output,
		&t.Dependencies, &t.Tags, &t.DeploymentMode, &t.CreatedBy, &t.ClaudeContext,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// GetPendingTasks retrieves all pending tasks
func (d *Database) GetPendingTasks() ([]*Task, error) {
	query := `
	SELECT task_id, title, description, assigned_to, status, priority,
	       created_at, started_at, completed_at, progress, output,
	       dependencies, tags, deployment_mode, created_by, claude_context
	FROM tasks WHERE status = 'pending'
	ORDER BY
	       CASE priority
	           WHEN 'high' THEN 1
	           WHEN 'medium' THEN 2
	           WHEN 'low' THEN 3
	       END,
	       created_at ASC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.Scan(
			&t.TaskID, &t.Title, &t.Description, &t.AssignedTo, &t.Status, &t.Priority,
			&t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.Progress, &t.Output,
			&t.Dependencies, &t.Tags, &t.DeploymentMode, &t.CreatedBy, &t.ClaudeContext,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	return tasks, nil
}

// GetTasksByStatus retrieves tasks by status
func (d *Database) GetTasksByStatus(status string) ([]*Task, error) {
	query := `
	SELECT task_id, title, description, assigned_to, status, priority,
	       created_at, started_at, completed_at, progress, output,
	       dependencies, tags, deployment_mode, created_by, claude_context
	FROM tasks WHERE status = ?
	ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.Scan(
			&t.TaskID, &t.Title, &t.Description, &t.AssignedTo, &t.Status, &t.Priority,
			&t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.Progress, &t.Output,
			&t.Dependencies, &t.Tags, &t.DeploymentMode, &t.CreatedBy, &t.ClaudeContext,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	return tasks, nil
}

// AssignTask assigns a task to a worker
func (d *Database) AssignTask(taskID, workerID string) error {
	now := time.Now()
	query := `
	UPDATE tasks
	SET assigned_to = ?, status = 'assigned', started_at = ?
	WHERE task_id = ? AND status = 'pending'
	`

	result, err := d.db.Exec(query, workerID, now, taskID)
	if err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task not found or not pending")
	}

	// Record in progress history
	d.addProgressEvent(workerID, taskID, "task_assigned", nil)

	d.logger.Infof("Task assigned: %s -> %s", taskID, workerID)
	return nil
}

// UpdateTaskProgress updates task progress
func (d *Database) UpdateTaskProgress(taskID string, progress int, message *string) error {
	query := `UPDATE tasks SET progress = ? WHERE task_id = ?`
	_, err := d.db.Exec(query, progress, taskID)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	// Get assigned worker
	task, err := d.GetTask(taskID)
	if err != nil {
		return err
	}

	if task != nil && task.AssignedTo != nil {
		// Record in progress history
		d.addProgressEvent(*task.AssignedTo, taskID, "progress_update", message)
	}

	return nil
}

// CompleteTask marks a task as completed
func (d *Database) CompleteTask(taskID string, output, errMsg *string) error {
	now := time.Now()
	status := "completed"
	if errMsg != nil {
		status = "failed"
	}

	query := `
	UPDATE tasks
	SET status = ?, completed_at = ?, output = ?, progress = ?
	WHERE task_id = ?
	`

	progress := 100
	if errMsg != nil {
		progress = 0
	}

	result, err := d.db.Exec(query, status, now, output, progress, taskID)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	// Get assigned worker
	task, err := d.GetTask(taskID)
	if err != nil {
		return err
	}

	if task != nil && task.AssignedTo != nil {
		// Record in progress history
		eventType := "task_completed"
		if errMsg != nil {
			eventType = "task_failed"
		}
		d.addProgressEvent(*task.AssignedTo, taskID, eventType, errMsg)
	}

	d.logger.Infof("Task completed: %s (status: %s)", taskID, status)
	return nil
}

// GetSystemStatus retrieves overall system status
func (d *Database) GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{}

	// Count tasks by status
	row := d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'pending'`)
	err := row.Scan(&status.PendingTasks)
	if err != nil {
		return nil, err
	}

	row = d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'in_progress'`)
	err = row.Scan(&status.InProgressTasks)
	if err != nil {
		return nil, err
	}

	row = d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'completed'`)
	err = row.Scan(&status.CompletedTasks)
	if err != nil {
		return nil, err
	}

	row = d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'failed'`)
	err = row.Scan(&status.FailedTasks)
	if err != nil {
		return nil, err
	}

	// Count workers by status
	row = d.db.QueryRow(`SELECT COUNT(*) FROM workers WHERE status = 'online' OR status = 'busy' OR status = 'idle'`)
	err = row.Scan(&status.OnlineWorkers)
	if err != nil {
		return nil, err
	}

	row = d.db.QueryRow(`SELECT COUNT(*) FROM workers WHERE status = 'idle'`)
	err = row.Scan(&status.IdleWorkers)
	if err != nil {
		return nil, err
	}

	row = d.db.QueryRow(`SELECT COUNT(*) FROM workers WHERE status = 'busy'`)
	err = row.Scan(&status.BusyWorkers)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// addProgressEvent adds a progress history event
func (d *Database) addProgressEvent(workerID, taskID, eventType string, message *string) error {
	query := `
	INSERT INTO progress_history (worker_id, task_id, event_type, message)
	VALUES (?, ?, ?, ?)
	`
	_, err := d.db.Exec(query, workerID, taskID, eventType, message)
	return err
}

// CleanupOfflineWorkers marks workers as offline if they haven't sent heartbeat recently
func (d *Database) CleanupOfflineWorkers(timeout time.Duration) error {
	cutoff := time.Now().Add(-timeout)
	query := `UPDATE workers SET status = 'offline' WHERE status != 'offline' AND last_heartbeat < ?`

	result, err := d.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup offline workers: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		d.logger.Infof("Marked %d workers as offline", rows)
	}

	return nil
}
