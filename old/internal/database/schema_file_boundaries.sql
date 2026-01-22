-- File Boundaries Management Schema for CPDS
-- 用于管理Worker文件编辑边界和冲突处理

-- 1. 文件边界定义表
CREATE TABLE IF NOT EXISTS file_boundaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_pattern TEXT NOT NULL,  -- 文件匹配模式，如 "src/**/*.go"
    line_range TEXT NOT NULL,    -- 行范围，如 "1-100" 或 "1-50,60-100"
    access_type TEXT NOT NULL,   -- 访问类型: "exclusive" | "shared" | "readonly"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, worker_id, file_pattern)
);

CREATE INDEX idx_file_boundaries_task ON file_boundaries(task_id);
CREATE INDEX idx_file_boundaries_worker ON file_boundaries(worker_id);
CREATE INDEX idx_file_boundaries_pattern ON file_boundaries(file_pattern);

-- 2. 编辑冲突记录表
CREATE TABLE IF NOT EXISTS edit_conflicts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conflict_id TEXT NOT NULL UNIQUE,  -- 冲突唯一ID
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    conflict_type TEXT NOT NULL,  -- "overlap" | "concurrent" | "upstream_changed"
    description TEXT,
    severity TEXT DEFAULT "medium",  -- "low" | "medium" | "high" | "critical"
    status TEXT DEFAULT "pending",    -- "pending" | "resolving" | "resolved" | "escalated"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by TEXT,  -- 高权限Agent的worker_id
    resolution TEXT
);

CREATE INDEX idx_edit_conflicts_status ON edit_conflicts(status);
CREATE INDEX idx_edit_conflicts_task ON edit_conflicts(task_id);
CREATE INDEX idx_edit_conflicts_severity ON edit_conflicts(severity);

-- 3. 冲突解决记录表
CREATE TABLE IF NOT EXISTS conflict_resolutions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conflict_id TEXT NOT NULL,
    resolver_agent_id TEXT NOT NULL,  -- 高权限Agent ID
    resolution_action TEXT NOT NULL,  -- "merge" | "override" | "skip" | "manual"
    original_content TEXT,
    resolved_content TEXT,
    explanation TEXT,
    resolved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conflict_id) REFERENCES edit_conflicts(conflict_id)
);

CREATE INDEX idx_conflict_resolutions_conflict ON conflict_resolutions(conflict_id);

-- 4. 文件编辑日志表
CREATE TABLE IF NOT EXISTS file_edit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_start INTEGER NOT NULL,
    line_end INTEGER NOT NULL,
    operation TEXT NOT NULL,  -- "edit" | "delete" | "insert"
    old_content TEXT,
    new_content TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_file_edit_log_file ON file_edit_log(file_path);
CREATE INDEX idx_file_edit_log_task ON file_edit_log(task_id);

-- 5. 高权限Agent任务表
CREATE TABLE IF NOT EXISTS privileged_agent_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL UNIQUE,
    priority INTEGER DEFAULT 100,  -- 优先级，数字越大越高
    task_type TEXT NOT NULL,  -- "conflict_resolution" | "upstream_sync" | "manual_intervention"
    description TEXT,
    conflict_ids TEXT,  -- JSON数组：["conflict-1", "conflict-2"]
    status TEXT DEFAULT "pending",  -- "pending" | "assigned" | "in_progress" | "completed"
    assigned_to TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_privileged_tasks_status ON privileged_agent_tasks(status);
CREATE INDEX idx_privileged_tasks_priority ON privileged_agent_tasks(priority);
