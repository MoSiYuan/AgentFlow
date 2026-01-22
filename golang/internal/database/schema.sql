-- CPDS Database Schema
-- SQLite with WAL mode for high concurrency

-- Tasks table with hierarchy and security
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER DEFAULT NULL,
    title TEXT NOT NULL,
    description TEXT,
    group_name TEXT NOT NULL DEFAULT 'default',
    completion_criteria TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'completed', 'failed', 'blocked')),
    priority INTEGER DEFAULT 0,
    lock_holder TEXT,
    lock_time DATETIME,
    result TEXT,
    error TEXT,
    workspace_dir TEXT,
    sandboxed INTEGER DEFAULT 0,
    allow_network INTEGER DEFAULT 1,
    max_memory TEXT DEFAULT '512M',
    max_cpu INTEGER DEFAULT 1,
    created_by TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Workers table
CREATE TABLE IF NOT EXISTS workers (
    id TEXT PRIMARY KEY,
    group_name TEXT NOT NULL DEFAULT 'default',
    type TEXT NOT NULL CHECK(type IN ('local', 'remote')),
    capabilities TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
    last_heartbeat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Task logs table
CREATE TABLE IF NOT EXISTS task_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    worker_id TEXT NOT NULL,
    log_level TEXT NOT NULL CHECK(log_level IN ('info', 'warning', 'error')),
    message TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_group ON tasks(group_name);
CREATE INDEX IF NOT EXISTS idx_tasks_lock ON tasks(lock_holder, lock_time);
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority DESC);
CREATE INDEX IF NOT EXISTS idx_workers_group ON workers(group_name);
CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status);
CREATE INDEX IF NOT EXISTS idx_task_logs_task ON task_logs(task_id);

-- Views for common queries
CREATE VIEW IF NOT EXISTS v_pending_tasks AS
SELECT id, title, description, group_name, created_at
FROM tasks
WHERE status = 'pending'
ORDER BY created_at ASC;

CREATE VIEW IF NOT EXISTS v_running_tasks AS
SELECT id, title, description, group_name, lock_holder, lock_time, started_at
FROM tasks
WHERE status = 'running'
ORDER BY started_at ASC;

CREATE VIEW IF NOT EXISTS v_active_workers AS
SELECT id, group_name, type, last_heartbeat
FROM workers
WHERE status = 'active'
  AND last_heartbeat > datetime('now', '-2 minutes')
ORDER BY last_heartbeat DESC;

CREATE VIEW IF NOT EXISTS v_group_stats AS
SELECT
    group_name,
    COUNT(*) as total_tasks,
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
    SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running_tasks,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_tasks,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_tasks
FROM tasks
GROUP BY group_name;

-- Git Integration Tables
CREATE TABLE IF NOT EXISTS git_tasks (
    id TEXT PRIMARY KEY,
    title TEXT,
    description TEXT,
    agent_id TEXT,
    git_branch TEXT NOT NULL UNIQUE,
    file_boundaries TEXT,
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    assigned_at DATETIME,
    completed_at DATETIME
);

CREATE TABLE IF NOT EXISTS git_locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    lock_type TEXT NOT NULL,
    acquired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    released_at DATETIME,
    status TEXT DEFAULT 'active',
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS git_conflicts (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    conflict_type TEXT NOT NULL,
    file_paths TEXT,
    description TEXT,
    severity TEXT DEFAULT 'medium',
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    resolved_by TEXT,
    resolution TEXT,
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);

-- Git Integration Indexes
CREATE INDEX IF NOT EXISTS idx_git_tasks_status ON git_tasks(status);
CREATE INDEX IF NOT EXISTS idx_git_tasks_agent ON git_tasks(agent_id);
CREATE INDEX IF NOT EXISTS idx_git_locks_file ON git_locks(file_path);
CREATE INDEX IF NOT EXISTS idx_git_locks_status ON git_locks(status);
CREATE INDEX IF NOT EXISTS idx_git_conflicts_status ON git_conflicts(status);
