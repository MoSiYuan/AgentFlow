-- Enhanced File Boundaries Management Schema with File Locking
-- 增强版文件边界管理，包含文件结构和锁机制

-- 1. 文件结构表（记录项目文件结构）
CREATE TABLE IF NOT EXISTS file_structure (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL UNIQUE,  -- 文件相对路径，如 "src/controllers/user.go"
    file_type TEXT NOT NULL,          -- "file" | "directory"
    parent_path TEXT,                 -- 父目录路径
    size INTEGER DEFAULT 0,           -- 文件大小（字节）
    checksum TEXT,                    -- 文件内容的MD5/SHA256
    version INTEGER DEFAULT 1,        -- 文件版本号
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_file_structure_path ON file_structure(file_path);
CREATE INDEX idx_file_structure_parent ON file_structure(parent_path);

-- 2. 文件锁表（实现文件级别的并发控制）
CREATE TABLE IF NOT EXISTS file_locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    worker_id TEXT NOT NULL,          -- 持有锁的Worker ID
    task_id TEXT NOT NULL,            -- 关联的任务ID
    lock_type TEXT NOT NULL,          -- "read" | "write" | "exclusive"
    line_start INTEGER,               -- 行锁：起始行（NULL表示文件锁）
    line_end INTEGER,                 -- 行锁：结束行
    acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,            -- 锁过期时间（NULL表示永不过期）
    status TEXT DEFAULT "active",     -- "active" | "released" | "expired"
    UNIQUE(file_path, worker_id, task_id, status)
);

CREATE INDEX idx_file_locks_path ON file_locks(file_path);
CREATE INDEX idx_file_locks_status ON file_locks(status);
CREATE INDEX idx_file_locks_worker ON file_locks(worker_id);

-- 3. 文件边界定义表
CREATE TABLE IF NOT EXISTS file_boundaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_path TEXT NOT NULL,          -- 具体文件路径
    line_start INTEGER NOT NULL,     -- 起始行
    line_end INTEGER NOT NULL,       -- 结束行
    access_type TEXT NOT NULL,       -- "exclusive" | "shared" | "readonly"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_path) REFERENCES file_structure(file_path)
);

CREATE INDEX idx_file_boundaries_task ON file_boundaries(task_id);
CREATE INDEX idx_file_boundaries_worker ON file_boundaries(worker_id);

-- 4. 编辑冲突记录表
CREATE TABLE IF NOT EXISTS edit_conflicts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conflict_id TEXT NOT NULL UNIQUE,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    conflict_type TEXT NOT NULL,     -- "overlap" | "concurrent" | "lock_denied" | "upstream_changed"
    description TEXT,
    severity TEXT DEFAULT "medium",
    status TEXT DEFAULT "pending",
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by TEXT,
    resolution TEXT,
    FOREIGN KEY (file_path) REFERENCES file_structure(file_path)
);

CREATE INDEX idx_edit_conflicts_status ON edit_conflicts(status);
CREATE INDEX idx_edit_conflicts_severity ON edit_conflicts(severity);

-- 5. 冲突解决记录表
CREATE TABLE IF NOT EXISTS conflict_resolutions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conflict_id TEXT NOT NULL,
    resolver_agent_id TEXT NOT NULL,
    resolution_action TEXT NOT NULL,
    original_content TEXT,
    resolved_content TEXT,
    explanation TEXT,
    resolved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conflict_id) REFERENCES edit_conflicts(conflict_id)
);

-- 6. 文件编辑日志表
CREATE TABLE IF NOT EXISTS file_edit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_start INTEGER NOT NULL,
    line_end INTEGER NOT NULL,
    operation TEXT NOT NULL,
    old_content TEXT,
    new_content TEXT,
    checksum_before TEXT,
    checksum_after TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_path) REFERENCES file_structure(file_path)
);

CREATE INDEX idx_file_edit_log_file ON file_edit_log(file_path);
CREATE INDEX idx_file_edit_log_task ON file_edit_log(task_id);

-- 7. 高权限Agent任务表
CREATE TABLE IF NOT EXISTS privileged_agent_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL UNIQUE,
    priority INTEGER DEFAULT 100,
    task_type TEXT NOT NULL,
    description TEXT,
    conflict_ids TEXT,
    file_paths TEXT,                -- JSON数组：需要处理的文件列表
    status TEXT DEFAULT "pending",
    assigned_to TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_privileged_tasks_status ON privileged_agent_tasks(status);
CREATE INDEX idx_privileged_tasks_priority ON privileged_agent_tasks(priority);

-- 8. 等待队列（当Worker无法获取锁时，任务进入等待队列）
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
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_path) REFERENCES file_structure(file_path)
);

CREATE INDEX idx_lock_wait_status ON lock_wait_queue(status);
CREATE INDEX idx_lock_wait_priority ON lock_wait_queue(priority);
