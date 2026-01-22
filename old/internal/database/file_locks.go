package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// FileStructure represents a file or directory in the project
type FileStructure struct {
	ID        int64     `json:"id"`
	FilePath string    `json:"file_path"`
	FileType string    `json:"file_type"`      // "file" | "directory"
	Parent   string    `json:"parent_path"`
	Size     int64     `json:"size"`
	Checksum string    `json:"checksum"`
	Version  int       `json:"version"`
	Created  time.Time `json:"created_at"`
	Updated  time.Time `json:"updated_at"`
}

// FileLock represents a lock on a file or file range
type FileLock struct {
	ID        int64      `json:"id"`
	FilePath  string     `json:"file_path"`
	WorkerID  string     `json:"worker_id"`
	TaskID    string     `json:"task_id"`
	LockType  string     `json:"lock_type"`  // "read" | "write" | "exclusive"
	LineStart *int       `json:"line_start"` // NULL for file lock
	LineEnd   *int       `json:"line_end"`
	Acquired  time.Time  `json:"acquired_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	Status    string     `json:"status"`
}

// LockWaitEntry represents an entry in the lock wait queue
type LockWaitEntry struct {
	ID           int64     `json:"id"`
	TaskID       string    `json:"task_id"`
	WorkerID     string    `json:"worker_id"`
	FilePath     string    `json:"file_path"`
	LockType     string    `json:"lock_type"`
	LineStart    *int      `json:"line_start"`
	LineEnd      *int      `json:"line_end"`
	Priority     int       `json:"priority"`
	Status       string    `json:"status"`
	RequestedAt  time.Time `json:"requested_at"`
}

// InitializeFileLockSchema initializes the file lock tables
func (db *Database) InitializeFileLockSchema() error {
	schema := `
	-- File Structure Table
	CREATE TABLE IF NOT EXISTS file_structure (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL UNIQUE,
		file_type TEXT NOT NULL,
		parent_path TEXT,
		size INTEGER DEFAULT 0,
		checksum TEXT,
		version INTEGER DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_file_structure_path ON file_structure(file_path);

	-- File Locks Table
	CREATE TABLE IF NOT EXISTS file_locks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		worker_id TEXT NOT NULL,
		task_id TEXT NOT NULL,
		lock_type TEXT NOT NULL,
		line_start INTEGER,
		line_end INTEGER,
		acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP,
		status TEXT DEFAULT "active",
		UNIQUE(file_path, worker_id, task_id, status)
	);

	CREATE INDEX IF NOT EXISTS idx_file_locks_path ON file_locks(file_path);
	CREATE INDEX IF NOT EXISTS idx_file_locks_status ON file_locks(status);

	-- Lock Wait Queue
	CREATE TABLE IF NOT EXISTS lock_wait_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		worker_id TEXT NOT NULL,
		file_path TEXT NOT NULL,
		lock_type TEXT NOT NULL,
		line_start INTEGER,
		line_end INTEGER,
		priority INTEGER DEFAULT 50,
		status TEXT DEFAULT "waiting",
		requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_lock_wait_status ON lock_wait_queue(status);
	`

	_, err := db.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize file lock schema: %w", err)
	}

	db.logger.Info("File lock schema initialized successfully")
	return nil
}

// === File Structure Management ===

// RegisterFileStructure registers a file or directory in the database
func (db *Database) RegisterFileStructure(fs *FileStructure) error {
	fs.Created = time.Now()
	fs.Updated = time.Now()

	query := `
		INSERT INTO file_structure (file_path, file_type, parent_path, size, checksum, version)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			size = excluded.size,
			checksum = excluded.checksum,
			version = version + 1,
			updated_at = CURRENT_TIMESTAMP
	`

	result, err := db.db.Exec(query, fs.FilePath, fs.FileType, fs.Parent,
		fs.Size, fs.Checksum, fs.Version)
	if err != nil {
		return fmt.Errorf("failed to register file structure: %w", err)
	}

	id, _ := result.LastInsertId()
	fs.ID = id

	db.logger.Debugf("Registered file structure: %s (type: %s)", fs.FilePath, fs.FileType)
	return nil
}

// GetFileStructure retrieves file structure by path
func (db *Database) GetFileStructure(filePath string) (*FileStructure, error) {
	query := `
		SELECT id, file_path, file_type, parent_path, size, checksum, version, created_at, updated_at
		FROM file_structure
		WHERE file_path = ?
	`

	var fs FileStructure
	err := db.db.QueryRow(query, filePath).Scan(
		&fs.ID, &fs.FilePath, &fs.FileType, &fs.Parent, &fs.Size,
		&fs.Checksum, &fs.Version, &fs.Created, &fs.Updated)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // File not found
		}
		return nil, fmt.Errorf("failed to get file structure: %w", err)
	}

	return &fs, nil
}

// GetAllFiles retrieves all files in the project
func (db *Database) GetAllFiles() ([]FileStructure, error) {
	query := `
		SELECT id, file_path, file_type, parent_path, size, checksum, version, created_at, updated_at
		FROM file_structure
		WHERE file_type = 'file'
		ORDER BY file_path
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []FileStructure
	for rows.Next() {
		var fs FileStructure
		err := rows.Scan(&fs.ID, &fs.FilePath, &fs.FileType, &fs.Parent,
			&fs.Size, &fs.Checksum, &fs.Version, &fs.Created, &fs.Updated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, fs)
	}

	return files, nil
}

// === File Lock Management ===

// AcquireFileLock attempts to acquire a lock on a file
func (db *Database) AcquireFileLock(lock *FileLock) (bool, error) {
	// Check if file is locked
	locked, existingLock, err := db.IsFileLocked(lock.FilePath, lock.LockType, lock.LineStart, lock.LineEnd)
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}

	if locked {
		// File is locked, add to wait queue
		if err := db.AddToWaitQueue(lock.TaskID, lock.WorkerID, lock.FilePath,
			lock.LockType, lock.LineStart, lock.LineEnd); err != nil {
			return false, fmt.Errorf("failed to add to wait queue: %w", err)
		}

		db.logger.Warnf("File %s locked by %s, added to wait queue",
			lock.FilePath, existingLock.WorkerID)
		return false, nil
	}

	// Acquire the lock
	lock.Acquired = time.Now()
	lock.Status = "active"

	query := `
		INSERT INTO file_locks
		(file_path, worker_id, task_id, lock_type, line_start, line_end, acquired_at, expires_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.db.Exec(query, lock.FilePath, lock.WorkerID, lock.TaskID,
		lock.LockType, lock.LineStart, lock.LineEnd, lock.Acquired, lock.ExpiresAt, lock.Status)

	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	db.logger.Infof("Lock acquired: %s -> %s (type: %s)", lock.FilePath, lock.WorkerID, lock.LockType)
	return true, nil
}

// ReleaseFileLock releases a file lock
func (db *Database) ReleaseFileLock(filePath, workerID, taskID string) error {
	query := `
		UPDATE file_locks
		SET status = 'released'
		WHERE file_path = ? AND worker_id = ? AND task_id = ? AND status = 'active'
	`

	result, err := db.db.Exec(query, filePath, workerID, taskID)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		db.logger.Infof("Lock released: %s by %s", filePath, workerID)

		// Process wait queue
		go db.ProcessWaitQueue(filePath)
	}

	return nil
}

// IsFileLocked checks if a file is locked
func (db *Database) IsFileLocked(filePath, lockType string, lineStart, lineEnd *int) (bool, *FileLock, error) {
	query := `
		SELECT id, file_path, worker_id, task_id, lock_type, line_start, line_end,
		       acquired_at, expires_at, status
		FROM file_locks
		WHERE file_path = ? AND status = 'active'
		ORDER BY acquired_at ASC
		LIMIT 1
	`

	row := db.db.QueryRow(query, filePath)

	var lock FileLock
	var ls, le sql.NullInt64
	var expiresAt sql.NullTime

	err := row.Scan(&lock.ID, &lock.FilePath, &lock.WorkerID, &lock.TaskID,
		&lock.LockType, &ls, &le, &lock.Acquired, &expiresAt, &lock.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil // No lock found
		}
		return false, nil, fmt.Errorf("failed to check lock: %w", err)
	}

	// Parse nullable fields
	if ls.Valid {
		val := int(ls.Int64)
		lock.LineStart = &val
	}
	if le.Valid {
		val := int(le.Int64)
		lock.LineEnd = &val
	}
	if expiresAt.Valid {
		lock.ExpiresAt = &expiresAt.Time
	}

	// Check if lock has expired
	if lock.ExpiresAt != nil && time.Now().After(*lock.ExpiresAt) {
		// Lock expired, release it
		db.ReleaseFileLock(filePath, lock.WorkerID, lock.TaskID)
		return false, nil, nil
	}

	// Check lock compatibility
	if lockType == "read" && lock.LockType == "read" {
		// Read locks are compatible
		return false, nil, nil
	}

	// Check for line overlap
	if lock.LineStart != nil && lineStart != nil {
		if rangesOverlap(*lock.LineStart, *lock.LineEnd, *lineStart, *lineEnd) {
			return true, &lock, nil
		}
		return false, nil, nil
	}

	// File-level lock
	return true, &lock, nil
}

// rangesOverlap checks if two ranges overlap
func rangesOverlap(start1, end1, start2, end2 int) bool {
	return start1 <= end2 && start2 <= end1
}

// AddToWaitQueue adds a lock request to the wait queue
func (db *Database) AddToWaitQueue(taskID, workerID, filePath, lockType string,
	lineStart, lineEnd *int) error {

	var ls, le sql.NullInt64
	if lineStart != nil {
		ls.Int64 = int64(*lineStart)
		ls.Valid = true
	}
	if lineEnd != nil {
		le.Int64 = int64(*lineEnd)
		le.Valid = true
	}

	query := `
		INSERT INTO lock_wait_queue
		(task_id, worker_id, file_path, lock_type, line_start, line_end, status)
		VALUES (?, ?, ?, ?, ?, ?, 'waiting')
	`

	_, err := db.db.Exec(query, taskID, workerID, filePath, lockType, ls, le)
	if err != nil {
		return fmt.Errorf("failed to add to wait queue: %w", err)
	}

	return nil
}

// ProcessWaitQueue processes the wait queue for a file
func (db *Database) ProcessWaitQueue(filePath string) error {
	// Get the highest priority waiting task
	query := `
		SELECT id, task_id, worker_id, file_path, lock_type, line_start, line_end, priority
		FROM lock_wait_queue
		WHERE file_path = ? AND status = 'waiting'
		ORDER BY priority DESC, requested_at ASC
		LIMIT 1
	`

	row := db.db.QueryRow(query, filePath)

	var entry LockWaitEntry
	var ls, le sql.NullInt64

	err := row.Scan(&entry.ID, &entry.TaskID, &entry.WorkerID, &entry.FilePath,
		&entry.LockType, &ls, &le, &entry.Priority)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No waiting tasks
		}
		return fmt.Errorf("failed to query wait queue: %w", err)
	}

	// Parse line range
	if ls.Valid {
		val := int(ls.Int64)
		entry.LineStart = &val
	}
	if le.Valid {
		val := int(le.Int64)
		entry.LineEnd = &val
	}

	// Try to acquire lock for this task
	lock := &FileLock{
		FilePath:  entry.FilePath,
		WorkerID:  entry.WorkerID,
		TaskID:    entry.TaskID,
		LockType:  entry.LockType,
		LineStart: entry.LineStart,
		LineEnd:   entry.LineEnd,
	}

	acquired, err := db.AcquireFileLock(lock)
	if err != nil {
		return fmt.Errorf("failed to acquire lock from wait queue: %w", err)
	}

	if acquired {
		// Remove from wait queue and notify worker
		_, err = db.db.Exec("DELETE FROM lock_wait_queue WHERE id = ?", entry.ID)
		if err != nil {
			return fmt.Errorf("failed to remove from wait queue: %w", err)
		}

		db.logger.Infof("Lock granted from wait queue: %s -> %s", filePath, entry.WorkerID)

		// TODO: Notify worker that lock is available
		// This could be done via WebSocket, HTTP callback, or worker polling
	}

	return nil
}

// GetActiveLocks retrieves all active locks
func (db *Database) GetActiveLocks() ([]FileLock, error) {
	query := `
		SELECT id, file_path, worker_id, task_id, lock_type, line_start, line_end,
		       acquired_at, expires_at, status
		FROM file_locks
		WHERE status = 'active'
		ORDER BY acquired_at DESC
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active locks: %w", err)
	}
	defer rows.Close()

	var locks []FileLock
	for rows.Next() {
		var lock FileLock
		var ls, le sql.NullInt64
		var expiresAt sql.NullTime

		err := rows.Scan(&lock.ID, &lock.FilePath, &lock.WorkerID, &lock.TaskID,
			&lock.LockType, &ls, &le, &lock.Acquired, &expiresAt, &lock.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lock: %w", err)
		}

		if ls.Valid {
			val := int(ls.Int64)
			lock.LineStart = &val
		}
		if le.Valid {
			val := int(le.Int64)
			lock.LineEnd = &val
		}
		if expiresAt.Valid {
			lock.ExpiresAt = &expiresAt.Time
		}

		locks = append(locks, lock)
	}

	return locks, nil
}

// ReleaseAllWorkerLocks releases all locks held by a worker
func (db *Database) ReleaseAllWorkerLocks(workerID string) (int, error) {
	query := `
		UPDATE file_locks
		SET status = 'released'
		WHERE worker_id = ? AND status = 'active'
	`

	result, err := db.db.Exec(query, workerID)
	if err != nil {
		return 0, fmt.Errorf("failed to release worker locks: %w", err)
	}

	rows, _ := result.RowsAffected()
	db.logger.Infof("Released %d locks for worker %s", rows, workerID)

	return int(rows), nil
}
