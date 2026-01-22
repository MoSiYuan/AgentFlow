package database

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Init() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("读取 schema 失败: %w", err)
	}

	_, err = d.db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("执行 schema 失败: %w", err)
	}

	return d.configure()
}

func (d *Database) configure() error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000",  // 64MB
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 30000000000",
		"PRAGMA page_size = 4096",
	}

	for _, pragma := range pragmas {
		if _, err := d.db.Exec(pragma); err != nil {
			return fmt.Errorf("配置 pragma 失败: %s: %w", pragma, err)
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// Task operations
func (d *Database) CreateTask(title, description, groupName, completionCriteria string) (string, error) {
	query := `
		INSERT INTO tasks (title, description, group_name, completion_criteria)
		VALUES (?, ?, ?, ?)
	`
	result, err := d.db.Exec(query, title, description, groupName, completionCriteria)
	if err != nil {
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("获取任务 ID 失败: %w", err)
	}

	return fmt.Sprintf("%d", id), nil
}

func (d *Database) GetTask(id string) (*Task, error) {
	query := `
		SELECT id, title, description, group_name, status,
		       lock_holder, lock_time, result, error,
		       created_at, started_at, completed_at
		FROM tasks
		WHERE id = ?
	`
	row := d.db.QueryRow(query, id)

	var task Task
	var lockTime, startedAt, completedAt sql.NullTime

	err := row.Scan(
		&task.ID, &task.Title, &task.Description, &task.GroupName, &task.Status,
		&task.LockHolder, &lockTime, &task.Result, &task.Error,
		&task.CreatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	if lockTime.Valid {
		task.LockTime = &lockTime.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

func (d *Database) ListTasks(status, groupName string) ([]Task, error) {
	query := `
		SELECT id, title, description, group_name, status,
		       lock_holder, lock_time, created_at, started_at, completed_at
		FROM tasks
		WHERE 1=1
	`
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if groupName != "" {
		query += " AND group_name = ?"
		args = append(args, groupName)
	}

	query += " ORDER BY created_at DESC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("列出任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var lockTime, startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.GroupName, &task.Status,
			&task.LockHolder, &lockTime, &task.CreatedAt, &startedAt, &completedAt,
		)
		if err != nil {
			return nil, err
		}

		if lockTime.Valid {
			task.LockTime = &lockTime.Time
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (d *Database) LockTask(taskID, workerID string) (bool, error) {
	query := `
		UPDATE tasks
		SET status = 'running',
		    lock_holder = ?,
		    lock_time = ?,
		    started_at = ?
		WHERE id = ?
		  AND status = 'pending'
		  AND (lock_holder IS NULL OR lock_holder = ? OR lock_time < datetime('now', '-5 minutes'))
	`
	result, err := d.db.Exec(query, workerID, time.Now().UTC(), time.Now().UTC(), taskID, workerID)
	if err != nil {
		return false, fmt.Errorf("锁定任务失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("获取影响行数失败: %w", err)
	}

	return affected > 0, nil
}

func (d *Database) UnlockTask(taskID, workerID string) error {
	query := `
		UPDATE tasks
		SET lock_holder = NULL,
		    lock_time = NULL
		WHERE id = ? AND lock_holder = ?
	`
	_, err := d.db.Exec(query, taskID, workerID)
	if err != nil {
		return fmt.Errorf("解锁任务失败: %w", err)
	}

	return nil
}

func (d *Database) CompleteTask(taskID, workerID, result string) error {
	query := `
		UPDATE tasks
		SET status = 'completed',
		    lock_holder = NULL,
		    lock_time = NULL,
		    result = ?,
		    completed_at = ?
		WHERE id = ? AND lock_holder = ?
	`
	_, err := d.db.Exec(query, result, time.Now().UTC(), taskID, workerID)
	if err != nil {
		return fmt.Errorf("完成任务失败: %w", err)
	}

	return nil
}

func (d *Database) FailTask(taskID, workerID, errorMsg string) error {
	query := `
		UPDATE tasks
		SET status = 'failed',
		    lock_holder = NULL,
		    lock_time = NULL,
		    error = ?,
		    completed_at = ?
		WHERE id = ? AND lock_holder = ?
	`
	_, err := d.db.Exec(query, errorMsg, time.Now().UTC(), taskID, workerID)
	if err != nil {
		return fmt.Errorf("标记任务失败: %w", err)
	}

	return nil
}

// Worker operations
func (d *Database) RegisterWorker(workerID, groupName, workerType string, capabilities string) error {
	query := `
		INSERT OR REPLACE INTO workers (id, group_name, type, capabilities, status, last_heartbeat)
		VALUES (?, ?, ?, ?, 'active', ?)
	`
	_, err := d.db.Exec(query, workerID, groupName, workerType, capabilities, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("注册 worker 失败: %w", err)
	}

	return nil
}

func (d *Database) UpdateWorkerHeartbeat(workerID string) error {
	query := `
		UPDATE workers
		SET last_heartbeat = ?
		WHERE id = ?
	`
	_, err := d.db.Exec(query, time.Now().UTC(), workerID)
	if err != nil {
		return fmt.Errorf("更新心跳失败: %w", err)
	}

	return nil
}

func (d *Database) ListWorkers(groupName string) ([]Worker, error) {
	query := `
		SELECT id, group_name, type, capabilities, status, last_heartbeat, created_at
		FROM workers
		WHERE 1=1
	`
	args := []interface{}{}

	if groupName != "" {
		query += " AND group_name = ?"
		args = append(args, groupName)
	}

	query += " ORDER BY created_at DESC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("列出 worker 失败: %w", err)
	}
	defer rows.Close()

	var workers []Worker
	for rows.Next() {
		var worker Worker
		err := rows.Scan(
			&worker.ID, &worker.GroupName, &worker.Type, &worker.Capabilities,
			&worker.Status, &worker.LastHeartbeat, &worker.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		workers = append(workers, worker)
	}

	return workers, nil
}

// Task log operations
func (d *Database) AddTaskLog(taskID, workerID, logLevel, message string) error {
	query := `
		INSERT INTO task_logs (task_id, worker_id, log_level, message)
		VALUES (?, ?, ?, ?)
	`
	_, err := d.db.Exec(query, taskID, workerID, logLevel, message)
	if err != nil {
		return fmt.Errorf("添加日志失败: %w", err)
	}

	return nil
}

func (d *Database) GetTaskLogs(taskID string) ([]TaskLog, error) {
	query := `
		SELECT id, task_id, worker_id, log_level, message, created_at
		FROM task_logs
		WHERE task_id = ?
		ORDER BY created_at ASC
	`
	rows, err := d.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取日志失败: %w", err)
	}
	defer rows.Close()

	var logs []TaskLog
	for rows.Next() {
		var log TaskLog
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.WorkerID, &log.Level, &log.Message, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// Statistics
func (d *Database) GetGroupStats() ([]GroupStat, error) {
	query := `
		SELECT
			group_name,
			COUNT(*) as total_tasks,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running_tasks,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_tasks,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_tasks
		FROM tasks
		GROUP BY group_name
	`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取统计失败: %w", err)
	}
	defer rows.Close()

	var stats []GroupStat
	for rows.Next() {
		var stat GroupStat
		err := rows.Scan(
			&stat.GroupName, &stat.TotalTasks, &stat.PendingTasks,
			&stat.RunningTasks, &stat.CompletedTasks, &stat.FailedTasks,
		)
		if err != nil {
			return nil, err
		}

		stats = append(stats, stat)
	}

	return stats, nil
}
