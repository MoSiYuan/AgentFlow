package git

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
	"github.com/sirupsen/logrus"
)

// FileBoundary represents a file boundary for an agent
type FileBoundary struct {
	FilePath    string `json:"file_path"`
	AccessType  string `json:"access_type"`  // "exclusive" | "shared" | "readonly"
	LineRange   string `json:"line_range"`   // e.g., "1-100"
	Description string `json:"description"`
}

// FileLock represents a file lock
type FileLock struct {
	FilePath   string    `json:"file_path"`
	AgentID    string    `json:"agent_id"`
	LockType   string    `json:"lock_type"` // "read" | "write"
	AcquiredAt time.Time `json:"acquired_at"`
	Status     string    `json:"status"` // "active" | "released"
}

// GitConflict represents a Git conflict
type GitConflict struct {
	ConflictID   string   `json:"conflict_id"`
	TaskID       string   `json:"task_id"`
	AgentID      string   `json:"agent_id"`
	ConflictType string   `json:"conflict_type"` // "file_locked" | "boundary_overlap" | "merge_conflict"
	FilePaths    []string `json:"file_paths"`
	Severity     string   `json:"severity"` // "low" | "medium" | "high" | "critical"
	Description  string   `json:"description"`
	Status       string   `json:"status"` // "pending" | "resolving" | "resolved"
}

// GitTask represents a task with Git integration
type GitTask struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	AgentID        string         `json:"agent_id"`
	GitBranch      string         `json:"git_branch"`
	FileBoundaries []FileBoundary `json:"file_boundaries"`
	Locks          []FileLock     `json:"locks"`
	Status         string         `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	AssignedAt     *time.Time     `json:"assigned_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
}

// BoundaryManager manages file boundaries and permissions
type BoundaryManager struct {
	boundaries map[string][]FileBoundary
	locks      map[string]*FileLock
	mutex      sync.RWMutex
	configPath string
	logger     *logrus.Logger
}

// NewBoundaryManager creates a new boundary manager
func NewBoundaryManager(configPath string, logger *logrus.Logger) *BoundaryManager {
	bm := &BoundaryManager{
		boundaries: make(map[string][]FileBoundary),
		locks:      make(map[string]*FileLock),
		configPath: configPath,
		logger:     logger,
	}

	bm.loadBoundaries()
	return bm
}

// loadBoundaries loads boundaries from config file
func (bm *BoundaryManager) loadBoundaries() {
	data, err := os.ReadFile(bm.configPath)
	if err != nil {
		bm.logger.Warnf("Failed to load boundaries: %v, creating defaults", err)
		bm.createDefaultBoundaries()
		return
	}

	var boundaries map[string][]FileBoundary
	if err := json.Unmarshal(data, &boundaries); err != nil {
		bm.logger.Errorf("Failed to parse boundaries: %v", err)
		bm.createDefaultBoundaries()
		return
	}

	bm.boundaries = boundaries
}

// createDefaultBoundaries creates default boundary configuration
func (bm *BoundaryManager) createDefaultBoundaries() {
	bm.boundaries = map[string][]FileBoundary{
		"frontend": {
			{
				FilePath:    "src/frontend/**/*",
				AccessType:  "exclusive",
				Description: "Frontend agent can exclusively modify frontend files",
			},
			{
				FilePath:    "src/api/**/*",
				AccessType:  "readonly",
				Description: "Frontend agent can read API files",
			},
		},
		"backend": {
			{
				FilePath:    "src/backend/**/*",
				AccessType:  "exclusive",
				Description: "Backend agent can exclusively modify backend files",
			},
			{
				FilePath:    "src/api/**/*",
				AccessType:  "shared",
				Description: "Backend agent shares API files",
			},
		},
		"database": {
			{
				FilePath:    "src/database/**/*",
				AccessType:  "exclusive",
				Description: "Database agent exclusively manages database files",
			},
		},
	}

	bm.saveBoundaries()
}

// saveBoundaries saves boundaries to config file
func (bm *BoundaryManager) saveBoundaries() {
	data, err := json.MarshalIndent(bm.boundaries, "", "  ")
	if err != nil {
		bm.logger.Errorf("Failed to marshal boundaries: %v", err)
		return
	}

	// Create directory if not exists
	dir := filepath.Dir(bm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		bm.logger.Errorf("Failed to create directory: %v", err)
		return
	}

	if err := os.WriteFile(bm.configPath, data, 0644); err != nil {
		bm.logger.Errorf("Failed to save boundaries: %v", err)
	}
}

// GetAgentBoundaries returns boundaries for an agent
func (bm *BoundaryManager) GetAgentBoundaries(agentID string) []FileBoundary {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	return bm.boundaries[agentID]
}

// CanAccessFile checks if an agent can access a file
func (bm *BoundaryManager) CanAccessFile(agentID, filePath, accessType string) bool {
	boundaries := bm.GetAgentBoundaries(agentID)
	if len(boundaries) == 0 {
		return false
	}

	for _, boundary := range boundaries {
		if bm.matchesPattern(filePath, boundary.FilePath) {
			if boundary.AccessType == "exclusive" {
				return true
			}
			if boundary.AccessType == "shared" {
				return true
			}
			if boundary.AccessType == "readonly" && accessType == "read" {
				return true
			}
		}
	}

	return false
}

// matchesPattern checks if file path matches pattern
func (bm *BoundaryManager) matchesPattern(filePath, pattern string) bool {
	// Simple pattern matching (can be enhanced with filepath.Match)
	filePath = filepath.Clean(filePath)
	pattern = filepath.Clean(pattern)

	if strings.Contains(pattern, "**") {
		// Matches any subdirectory
		basePattern := strings.Split(pattern, "**")[0]
		return strings.HasPrefix(filePath, basePattern)
	}

	if strings.Contains(pattern, "*") {
		// Use filepath.Match for glob patterns
		matched, _ := filepath.Match(pattern, filePath)
		return matched
	}

	return filePath == pattern
}

// AcquireLock acquires a file lock
func (bm *BoundaryManager) AcquireLock(agentID, filePath, lockType string) bool {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	lockKey := fmt.Sprintf("%s:%s", filePath, lockType)

	// Check if lock already exists
	if existingLock, ok := bm.locks[lockKey]; ok {
		if existingLock.Status == "active" {
			if lockType == "write" || existingLock.LockType == "write" {
				return false
			}
		}
	}

	// Create lock
	bm.locks[lockKey] = &FileLock{
		FilePath:   filePath,
		AgentID:    agentID,
		LockType:   lockType,
		AcquiredAt: time.Now(),
		Status:     "active",
	}

	bm.logger.WithFields(logrus.Fields{
		"agent_id":  agentID,
		"file_path": filePath,
		"lock_type": lockType,
	}).Info("Lock acquired")

	return true
}

// ReleaseLock releases a file lock
func (bm *BoundaryManager) ReleaseLock(agentID, filePath, lockType string) bool {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	lockKey := fmt.Sprintf("%s:%s", filePath, lockType)

	if lock, ok := bm.locks[lockKey]; ok {
		if lock.AgentID == agentID {
			lock.Status = "released"
			delete(bm.locks, lockKey)

			bm.logger.WithFields(logrus.Fields{
				"agent_id":  agentID,
				"file_path": filePath,
				"lock_type": lockType,
			}).Info("Lock released")

			return true
		}
	}

	return false
}

// GetActiveLocks returns all active locks for a file
func (bm *BoundaryManager) GetActiveLocks(filePath string) []*FileLock {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	var activeLocks []*FileLock
	for _, lock := range bm.locks {
		if lock.FilePath == filePath && lock.Status == "active" {
			activeLocks = append(activeLocks, lock)
		}
	}

	return activeLocks
}

// CheckConflicts checks for potential conflicts
func (bm *BoundaryManager) CheckConflicts(agentID string, filePaths []string) []*GitConflict {
	var conflicts []*GitConflict

	for _, filePath := range filePaths {
		activeLocks := bm.GetActiveLocks(filePath)

		for _, lock := range activeLocks {
			if lock.AgentID != agentID && lock.LockType == "write" {
				conflicts = append(conflicts, &GitConflict{
					ConflictID:   generateID(),
					TaskID:       "",
					AgentID:      agentID,
					ConflictType: "file_locked",
					FilePaths:    []string{filePath},
					Severity:     "high",
					Description:  fmt.Sprintf("File %s is locked by %s", filePath, lock.AgentID),
					Status:       "pending",
				})
			}
		}
	}

	return conflicts
}

// BranchManager manages Git branches for agents
type BranchManager struct {
	repoPath string
	logger   *logrus.Logger
}

// NewBranchManager creates a new branch manager
func NewBranchManager(repoPath string, logger *logrus.Logger) *BranchManager {
	return &BranchManager{
		repoPath: repoPath,
		logger:   logger,
	}
}

// CreateAgentBranch creates a new branch for an agent task
func (bm *BranchManager) CreateAgentBranch(agentID, taskID string) (string, error) {
	branchName := fmt.Sprintf("agent-%s/task-%s", agentID, taskID)

	// Create branch
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = bm.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %s", string(output))
	}

	bm.logger.WithFields(logrus.Fields{
		"branch": branchName,
		"agent":  agentID,
		"task":   taskID,
	}).Info("Branch created")

	return branchName, nil
}

// DeleteAgentBranch deletes an agent branch
func (bm *BranchManager) DeleteAgentBranch(branchName string, force bool) error {
	args := []string{"branch", "-d"}
	if force {
		args[1] = "-D"
	}
	args = append(args, branchName)

	cmd := exec.Command("git", args...)
	cmd.Dir = bm.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete branch: %s", string(output))
	}

	bm.logger.WithField("branch", branchName).Info("Branch deleted")
	return nil
}

// SwitchBranch switches to a branch
func (bm *BranchManager) SwitchBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = bm.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to switch branch: %s", string(output))
	}

	return nil
}

// CommitChanges commits changes on agent's branch
func (bm *BranchManager) CommitChanges(agentID, message string) (string, error) {
	// Add changes
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = bm.repoPath

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to add files: %s", string(output))
	}

	// Commit
	commitMessage := fmt.Sprintf("[%s] %s", agentID, message)
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Dir = bm.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to commit: %s", string(output))
	}

	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = bm.repoPath

	hashOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %v", err)
	}

	commitHash := strings.TrimSpace(string(hashOutput))

	bm.logger.WithFields(logrus.Fields{
		"agent":      agentID,
		"commit":     commitHash,
		"message":    message,
	}).Info("Changes committed")

	return commitHash, nil
}

// MergeToMain merges agent branch to main
func (bm *BranchManager) MergeToMain(branchName, strategy string) (bool, error) {
	// Switch to main
	err := bm.SwitchBranch("main")
	if err != nil {
		// Try master instead
		if err := bm.SwitchBranch("master"); err != nil {
			return false, fmt.Errorf("no main or master branch found")
		}
	}

	// Merge
	var cmd *exec.Cmd
	switch strategy {
	case "merge":
		cmd = exec.Command("git", "merge", branchName)
	case "squash":
		cmd = exec.Command("git", "merge", "--squash", branchName)
	case "rebase":
		cmd = exec.Command("git", "rebase", branchName)
	default:
		return false, fmt.Errorf("unknown merge strategy: %s", strategy)
	}

	cmd.Dir = bm.repoPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		bm.logger.WithFields(logrus.Fields{
			"branch": branchName,
			"error":  string(output),
		}).Warn("Merge failed, possible conflicts")
		return false, nil
	}

	bm.logger.WithField("branch", branchName).Info("Merge successful")
	return true, nil
}

// GetMergeConflicts gets list of files with merge conflicts
func (bm *BranchManager) GetMergeConflicts() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = bm.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	conflicts := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, c := range conflicts {
		if c != "" {
			result = append(result, c)
		}
	}

	return result, nil
}

// GitIntegrationManager manages Git integration for AgentFlow
type GitIntegrationManager struct {
	boundaryManager *BoundaryManager
	branchManager   *BranchManager
	db              *database.Database
	logger          *logrus.Logger
}

// NewGitIntegrationManager creates a new Git integration manager
func NewGitIntegrationManager(repoPath, configPath string, db *database.Database, logger *logrus.Logger) *GitIntegrationManager {
	return &GitIntegrationManager{
		boundaryManager: NewBoundaryManager(configPath, logger),
		branchManager:   NewBranchManager(repoPath, logger),
		db:              db,
		logger:          logger,
	}
}

// CreateAgentTask creates a new task for an agent with Git integration
func (gm *GitIntegrationManager) CreateAgentTask(ctx context.Context, agentID, taskID, description string) (*GitTask, error) {
	// Create agent branch
	branchName, err := gm.branchManager.CreateAgentBranch(agentID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	task := &GitTask{
		ID:             taskID,
		Description:    description,
		AgentID:        agentID,
		GitBranch:      branchName,
		FileBoundaries: gm.boundaryManager.GetAgentBoundaries(agentID),
		Status:         "in_progress",
		CreatedAt:      time.Now(),
	}

	// Store in database
	if err := gm.storeGitTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

// VerifyFileAccess verifies if agent can access a file
func (gm *GitIntegrationManager) VerifyFileAccess(agentID, filePath, accessType string) (bool, string) {
	if !gm.boundaryManager.CanAccessFile(agentID, filePath, accessType) {
		return false, fmt.Sprintf("Agent %s not authorized to access %s", agentID, filePath)
	}

	// Check locks
	conflicts := gm.boundaryManager.CheckConflicts(agentID, []string{filePath})
	if len(conflicts) > 0 && accessType == "write" {
		return false, fmt.Sprintf("File locked: %s", conflicts[0].Description)
	}

	return true, ""
}

// SafeFileOperation performs a safe file operation with locking
func (gm *GitIntegrationManager) SafeFileOperation(ctx context.Context, agentID, filePath string, operation func() error) error {
	// Check boundaries
	allowed, reason := gm.VerifyFileAccess(agentID, filePath, "write")
	if !allowed {
		return fmt.Errorf("access denied: %s", reason)
	}

	// Acquire lock
	if !gm.boundaryManager.AcquireLock(agentID, filePath, "write") {
		return fmt.Errorf("could not acquire lock for %s", filePath)
	}

	defer gm.boundaryManager.ReleaseLock(agentID, filePath, "write")

	// Perform operation
	if err := operation(); err != nil {
		return err
	}

	// Commit changes
	_, err := gm.branchManager.CommitChanges(agentID, fmt.Sprintf("Modified %s", filePath))
	return err
}

// CompleteAgentTask completes an agent task and merges to main
func (gm *GitIntegrationManager) CompleteAgentTask(ctx context.Context, agentID, taskID, mergeStrategy string) (map[string]interface{}, error) {
	branchName := fmt.Sprintf("agent-%s/task-%s", agentID, taskID)

	// Check for conflicts
	success, err := gm.branchManager.MergeToMain(branchName, mergeStrategy)
	if err != nil {
		return nil, err
	}

	if !success {
		conflicts, _ := gm.branchManager.GetMergeConflicts()
		return map[string]interface{}{
			"status":    "conflict",
			"branch":    branchName,
			"conflicts": conflicts,
			"message":   fmt.Sprintf("Merge conflicts detected in %d files", len(conflicts)),
		}, nil
	}

	// Clean up
	gm.branchManager.DeleteAgentBranch(branchName, true)

	// Update task status
	now := time.Now()
	task := &GitTask{
		ID:          taskID,
		AgentID:     agentID,
		GitBranch:   branchName,
		Status:      "completed",
		CompletedAt: &now,
	}

	if err := gm.updateGitTask(task); err != nil {
		gm.logger.WithError(err).Warn("Failed to update task status")
	}

	return map[string]interface{}{
		"status":  "completed",
		"branch":  branchName,
		"message": fmt.Sprintf("Task %s completed and merged successfully", taskID),
	}, nil
}

// storeGitTask stores Git task in database
func (gm *GitIntegrationManager) storeGitTask(task *GitTask) error {
	boundariesJSON, _ := json.Marshal(task.FileBoundaries)

	query := `
		INSERT INTO git_tasks (id, title, description, agent_id, git_branch, file_boundaries, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			agent_id = excluded.agent_id,
			git_branch = excluded.git_branch,
			file_boundaries = excluded.file_boundaries,
			status = excluded.status
	`

	_, err := gm.db.DB.Exec(query,
		task.ID, task.Title, task.Description, task.AgentID,
		task.GitBranch, string(boundariesJSON), task.Status, task.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store git task: %w", err)
	}

	gm.logger.WithFields(logrus.Fields{
		"task_id":   task.ID,
		"agent_id":  task.AgentID,
		"branch":    task.GitBranch,
	}).Info("Git task created")

	return nil
}

// updateGitTask updates Git task in database
func (gm *GitIntegrationManager) updateGitTask(task *GitTask) error {
	boundariesJSON, _ := json.Marshal(task.FileBoundaries)

	query := `
		UPDATE git_tasks
		SET status = ?, completed_at = ?
		WHERE id = ?
	`

	_, err := gm.db.DB.Exec(query, task.Status, task.CompletedAt, task.ID)

	if err != nil {
		return fmt.Errorf("failed to update git task: %w", err)
	}

	return nil
}

// CreateGitTables creates Git-integrated tables in database
func CreateGitTables(db *sql.DB) error {
	schema := `
	-- Git tasks table
	CREATE TABLE IF NOT EXISTS git_tasks (
		id TEXT PRIMARY KEY,
		title TEXT,
		description TEXT,
		agent_id TEXT,
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
		agent_id TEXT NOT NULL,
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
		agent_id TEXT NOT NULL,
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
	CREATE INDEX IF NOT EXISTS idx_git_tasks_agent ON git_tasks(agent_id);
	CREATE INDEX IF NOT EXISTS idx_git_locks_file ON git_locks(file_path);
	CREATE INDEX IF NOT EXISTS idx_git_locks_status ON git_locks(status);
	CREATE INDEX IF NOT EXISTS idx_git_conflicts_status ON git_conflicts(status);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create Git tables: %w", err)
	}

	return nil
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
