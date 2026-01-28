/**
 * AgentFlow Database Layer
 *
 * Provides SQLite database operations for AgentFlow.
 * Compatible with Python and Go versions.
 */

import Database from 'better-sqlite3';
import { homedir } from 'os';
import { resolve } from 'path';
import type {
  Task,
  Worker,
  TaskStatus,
  SystemStats,
  GroupStats,
  TaskFilter,
  WorkerRegistration,
} from '@agentflow/shared';

export class AgentFlowDatabase {
  private db: Database.Database;

  constructor(dbPath: string = '~/.claude/skills/agentflow/agentflow.db') {
    // Expand ~ to home directory
    let expandedPath = dbPath;
    if (dbPath.startsWith('~')) {
      expandedPath = resolve(homedir(), dbPath.slice(1));
    }

    this.db = new Database(expandedPath);
    this.db.pragma('journal_mode = WAL');
    this.db.pragma('foreign_keys = ON');
  }

  /**
   * Initialize database schema
   */
  init(): void {
    this.exec(`
      -- Tasks table
      CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        parent_id INTEGER DEFAULT NULL,
        title TEXT NOT NULL,
        description TEXT,
        group_name TEXT NOT NULL DEFAULT 'default',
        completion_criteria TEXT,
        status TEXT NOT NULL DEFAULT 'pending'
          CHECK(status IN ('pending', 'running', 'completed', 'failed', 'blocked')),
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
        status TEXT NOT NULL DEFAULT 'active'
          CHECK(status IN ('active', 'inactive')),
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

      -- Git tasks table
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

      -- Git locks table
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
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        resolved_at DATETIME,
        resolved_by TEXT,
        resolution TEXT,
        FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
      );

      -- Task relationships table (DAG and orchestration)
      CREATE TABLE IF NOT EXISTS task_relationships (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        predecessor_id INTEGER NOT NULL,
        successor_id INTEGER NOT NULL,
        relationship_type TEXT NOT NULL CHECK(relationship_type IN ('dependency', 'context', 'upgrade', 'parallel', 'sequential')),
        data_flow TEXT,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (predecessor_id) REFERENCES tasks(id) ON DELETE CASCADE,
        FOREIGN KEY (successor_id) REFERENCES tasks(id) ON DELETE CASCADE,
        UNIQUE(predecessor_id, successor_id, relationship_type)
      );

      -- Task checkpoints table (short-term memory)
      CREATE TABLE IF NOT EXISTS task_checkpoints (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        worker_id TEXT NOT NULL,
        checkpoint_name TEXT NOT NULL,
        checkpoint_data TEXT NOT NULL,
        memory_snapshot TEXT,
        state_snapshot TEXT,
        timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
      );

      -- Task versions table (task upgrade mechanism)
      CREATE TABLE IF NOT EXISTS task_versions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        version_number INTEGER NOT NULL,
        title TEXT NOT NULL,
        description TEXT,
        upgrade_reason TEXT,
        upgraded_from INTEGER,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
        FOREIGN KEY (upgraded_from) REFERENCES tasks(id) ON DELETE SET NULL,
        UNIQUE(task_id, version_number)
      );

      -- Claude Integration Tables
      -- AgentFlow 与 Claude CLI 深度集成
      CREATE TABLE IF NOT EXISTS claude_mappings (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        session_uuid TEXT NOT NULL,
        message_uuid TEXT NOT NULL UNIQUE,
        parent_message_uuid TEXT,
        slug TEXT,
        source TEXT DEFAULT 'api',
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
      );

      -- Task Chain Tables
      -- 支持串行、并行、树形任务链
      CREATE TABLE IF NOT EXISTS task_chains (
        id TEXT PRIMARY KEY,
        session_uuid TEXT NOT NULL,
        root_message_uuid TEXT NOT NULL,
        chain_type TEXT NOT NULL CHECK(chain_type IN ('sequential', 'parallel', 'tree')),
        status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'completed', 'failed')),
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        started_at DATETIME,
        completed_at DATETIME
      );

      -- Task Chain Nodes (任务链节点)
      CREATE TABLE IF NOT EXISTS task_chain_nodes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chain_id TEXT NOT NULL,
        task_id INTEGER NOT NULL,
        parent_node_id INTEGER,
        node_order INTEGER NOT NULL DEFAULT 0,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (chain_id) REFERENCES task_chains(id) ON DELETE CASCADE,
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
        FOREIGN KEY (parent_node_id) REFERENCES task_chain_nodes(id) ON DELETE SET NULL
      );
    `);

    // Create indexes
    this.exec(`
      CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
      CREATE INDEX IF NOT EXISTS idx_tasks_group ON tasks(group_name);
      CREATE INDEX IF NOT EXISTS idx_tasks_lock ON tasks(lock_holder, lock_time);
      CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
      CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority DESC);
      CREATE INDEX IF NOT EXISTS idx_workers_group ON workers(group_name);
      CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status);
      CREATE INDEX IF NOT EXISTS idx_task_logs_task ON task_logs(task_id);
      CREATE INDEX IF NOT EXISTS idx_git_tasks_status ON git_tasks(status);
      CREATE INDEX IF NOT EXISTS idx_git_locks_file ON git_locks(file_path);
      CREATE INDEX IF NOT EXISTS idx_git_locks_status ON git_locks(status);
      CREATE INDEX IF NOT EXISTS idx_git_conflicts_status ON git_conflicts(status);
      CREATE INDEX IF NOT EXISTS idx_task_relationships_pred ON task_relationships(predecessor_id);
      CREATE INDEX IF NOT EXISTS idx_task_relationships_succ ON task_relationships(successor_id);
      CREATE INDEX IF NOT EXISTS idx_task_relationships_type ON task_relationships(relationship_type);
      CREATE INDEX IF NOT EXISTS idx_task_checkpoints_task ON task_checkpoints(task_id);
      CREATE INDEX IF NOT EXISTS idx_task_checkpoints_timestamp ON task_checkpoints(timestamp DESC);
      CREATE INDEX IF NOT EXISTS idx_task_versions_task ON task_versions(task_id);
      CREATE INDEX IF NOT EXISTS idx_task_versions_number ON task_versions(task_id, version_number);
      CREATE INDEX IF NOT EXISTS idx_claude_session ON claude_mappings(session_uuid);
      CREATE INDEX IF NOT EXISTS idx_claude_message ON claude_mappings(message_uuid);
      CREATE INDEX IF NOT EXISTS idx_claude_task ON claude_mappings(task_id);
      CREATE INDEX IF NOT EXISTS idx_claude_parent ON claude_mappings(parent_message_uuid);
      CREATE INDEX IF NOT EXISTS idx_claude_slug ON claude_mappings(slug);
      CREATE INDEX IF NOT EXISTS idx_task_chains_session ON task_chains(session_uuid);
      CREATE INDEX IF NOT EXISTS idx_task_chains_status ON task_chains(status);
      CREATE INDEX IF NOT EXISTS idx_task_chains_type ON task_chains(chain_type);
      CREATE INDEX IF NOT EXISTS idx_task_chain_nodes_chain ON task_chain_nodes(chain_id);
      CREATE INDEX IF NOT EXISTS idx_task_chain_nodes_task ON task_chain_nodes(task_id);
      CREATE INDEX IF NOT EXISTS idx_task_chain_nodes_parent ON task_chain_nodes(parent_node_id);
    `);
  }

  /**
   * Execute SQL statement
   */
  private exec(sql: string): void {
    this.db.exec(sql);
  }

  /**
   * Create a new task
   */
  createTask(data: {
    parent_id?: number;
    title: string;
    description: string;
    group_name?: string;
    completion_criteria?: string;
    priority?: number;
    workspace_dir?: string;
    sandboxed?: boolean;
    allow_network?: boolean;
    max_memory?: string;
    max_cpu?: number;
    created_by?: string;
  }): string {
    const stmt = this.db.prepare(`
      INSERT INTO tasks (
        parent_id, title, description, group_name, completion_criteria,
        priority, workspace_dir, sandboxed, allow_network, max_memory, max_cpu, created_by
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);

    const result = stmt.run(
      data.parent_id || null,
      data.title,
      data.description,
      data.group_name || 'default',
      data.completion_criteria || null,
      data.priority || 0,
      data.workspace_dir || null,
      data.sandboxed ? 1 : 0,
      data.allow_network ? 1 : 0,
      data.max_memory || '512M',
      data.max_cpu || 1,
      data.created_by || null
    );

    return `TASK-${String(result.lastInsertRowid).padStart(8, '0').toUpperCase()}`;
  }

  /**
   * Get task by ID
   */
  getTask(id: string): Task | null {
    const stmt = this.db.prepare('SELECT * FROM tasks WHERE id = ?');
    const row = stmt.get(id) as any;
    if (!row) return null;

    return this.mapTask(row);
  }

  /**
   * List tasks with optional filters
   */
  listTasks(filter?: TaskFilter): Task[] {
    let sql = 'SELECT * FROM tasks WHERE 1=1';
    const params: any[] = [];

    if (filter?.status) {
      sql += ' AND status = ?';
      params.push(filter.status);
    }

    if (filter?.group_name) {
      sql += ' AND group_name = ?';
      params.push(filter.group_name);
    }

    if (filter?.parent_id) {
      sql += ' AND parent_id = ?';
      params.push(filter.parent_id);
    }

    sql += ' ORDER BY created_at DESC';

    if (filter?.limit) {
      sql += ' LIMIT ?';
      params.push(filter.limit);
    }

    if (filter?.offset) {
      sql += ' OFFSET ?';
      params.push(filter.offset);
    }

    const stmt = this.db.prepare(sql);
    const rows = stmt.all(...params) as any[];

    return rows.map(row => this.mapTask(row));
  }

  /**
   * Update task status
   */
  updateTaskStatus(
    id: string,
    status: TaskStatus,
    workerId?: string
  ): boolean {
    let sql = 'UPDATE tasks SET status = ?';
    const params: any[] = [status];

    if (status === 'running') {
      sql += ', started_at = CURRENT_TIMESTAMP';
    } else if (status === 'completed' || status === 'failed') {
      sql += ', completed_at = CURRENT_TIMESTAMP';
    }

    if (workerId) {
      sql += ', lock_holder = ?, lock_time = CURRENT_TIMESTAMP';
      params.push(workerId);
    }

    sql += ' WHERE id = ?';
    params.push(id);

    const stmt = this.db.prepare(sql);
    const result = stmt.run(...params);
    return result.changes > 0;
  }

  /**
   * Complete task
   */
  completeTask(id: string, workerId: string, result: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE tasks
      SET status = 'completed',
          result = ?,
          lock_holder = NULL,
          completed_at = CURRENT_TIMESTAMP
      WHERE id = ? AND lock_holder = ?
    `);

    const updateResult = stmt.run(result, id, workerId);
    return updateResult.changes > 0;
  }

  /**
   * Fail task
   */
  failTask(id: string, workerId: string, error: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE tasks
      SET status = 'failed',
          error = ?,
          lock_holder = NULL,
          completed_at = CURRENT_TIMESTAMP
      WHERE id = ? AND lock_holder = ?
    `);

    const result = stmt.run(error, id, workerId);
    return result.changes > 0;
  }

  /**
   * Lock task for worker
   */
  lockTask(id: string, workerId: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE tasks
      SET status = 'running',
          lock_holder = ?,
          lock_time = CURRENT_TIMESTAMP,
          started_at = CURRENT_TIMESTAMP
      WHERE id = ? AND lock_holder IS NULL
    `);

    const result = stmt.run(workerId, id);
    return result.changes > 0;
  }

  /**
   * Unlock task
   */
  unlockTask(id: string, workerId: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE tasks
      SET lock_holder = NULL,
          lock_time = NULL,
          status = 'pending'
      WHERE id = ? AND lock_holder = ?
    `);

    const result = stmt.run(id, workerId);
    return result.changes > 0;
  }

  /**
   * Register worker
   */
  registerWorker(data: WorkerRegistration): boolean {
    const stmt = this.db.prepare(`
      INSERT OR REPLACE INTO workers (id, group_name, type, capabilities, status)
      VALUES (?, ?, ?, ?, 'active')
    `);

    try {
      stmt.run(
        data.worker_id || this.generateWorkerId(),
        data.group_name,
        'local',
        JSON.stringify(data.capabilities)
      );
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Update worker heartbeat
   */
  updateWorkerHeartbeat(id: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE workers
      SET last_heartbeat = CURRENT_TIMESTAMP, status = 'active'
      WHERE id = ?
    `);

    const result = stmt.run(id);
    return result.changes > 0;
  }

  /**
   * List workers
   */
  listWorkers(group?: string): Worker[] {
    let sql = 'SELECT * FROM workers WHERE status = "active"';
    const params: any[] = [];

    if (group) {
      sql += ' AND group_name = ?';
      params.push(group);
    }

    sql += ' ORDER BY last_heartbeat DESC';

    const stmt = this.db.prepare(sql);
    const rows = stmt.all(...params) as any[];

    return rows.map(row => ({
      id: row.id,
      group_name: row.group_name,
      type: row.type,
      capabilities: JSON.parse(row.capabilities || '[]'),
      status: row.status,
      last_heartbeat: new Date(row.last_heartbeat),
      created_at: new Date(row.created_at),
    }));
  }

  /**
   * Cleanup offline workers
   */
  cleanupOfflineWorkers(timeoutSeconds: number = 120): number {
    const stmt = this.db.prepare(`
      UPDATE workers
      SET status = 'inactive'
      WHERE status = 'active'
        AND datetime(last_heartbeat, '+' || ? || ' seconds') < datetime('now')
    `);

    const result = stmt.run(timeoutSeconds);
    return result.changes;
  }

  /**
   * Get system statistics
   */
  getStats(): SystemStats {
    const stmt = this.db.prepare(`
      SELECT
        COUNT(*) as total_tasks,
        SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
        SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running_tasks,
        SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_tasks,
        SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_tasks
      FROM tasks
    `);

    const taskStats = stmt.get() as any;

    const workerStmt = this.db.prepare(`
      SELECT
        COUNT(*) as total_workers,
        SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active_workers
      FROM workers
    `);

    const workerStats = workerStmt.get() as any;

    return {
      total_tasks: taskStats.total_tasks,
      pending_tasks: taskStats.pending_tasks,
      running_tasks: taskStats.running_tasks,
      completed_tasks: taskStats.completed_tasks,
      failed_tasks: taskStats.failed_tasks,
      total_workers: workerStats.total_workers,
      active_workers: workerStats.active_workers,
      uptime_seconds: 0, // To be set by caller
    };
  }

  /**
   * Get group statistics
   */
  getGroupStats(): GroupStats[] {
    const stmt = this.db.prepare(`
      SELECT
        group_name,
        COUNT(*) as total_tasks,
        SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
        SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running_tasks,
        SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_tasks,
        SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_tasks
      FROM tasks
      GROUP BY group_name
    `);

    const rows = stmt.all() as any[];
    return rows.map(row => ({
      group_name: row.group_name,
      total_tasks: row.total_tasks,
      pending_tasks: row.pending_tasks,
      running_tasks: row.running_tasks,
      completed_tasks: row.completed_tasks,
      failed_tasks: row.failed_tasks,
    }));
  }

  /**
   * Close database connection
   */
  close(): void {
    this.db.close();
  }

  /**
   * Map database row to Task object
   */
  private mapTask(row: any): Task {
    return {
      id: row.id,
      parent_id: row.parent_id || undefined,
      title: row.title,
      description: row.description,
      group_name: row.group_name,
      completion_criteria: row.completion_criteria || undefined,
      status: row.status,
      priority: row.priority === 2 ? 'high' : row.priority === 1 ? 'medium' : 'low',
      lock_holder: row.lock_holder || undefined,
      lock_time: row.lock_time ? new Date(row.lock_time) : undefined,
      result: row.result || undefined,
      error: row.error || undefined,
      workspace_dir: row.workspace_dir || undefined,
      sandboxed: row.sandboxed === 1,
      allow_network: row.allow_network === 1,
      max_memory: row.max_memory,
      max_cpu: row.max_cpu,
      created_by: row.created_by || undefined,
      created_at: new Date(row.created_at),
      started_at: row.started_at ? new Date(row.started_at) : undefined,
      completed_at: row.completed_at ? new Date(row.completed_at) : undefined,
    };
  }

  /**
   * Generate unique worker ID
   */
  private generateWorkerId(): string {
    const os = require('os');
    const pid = process.pid;
    const random = Math.random().toString(36).substring(2, 8);
    return `worker-${os.hostname()}-${pid}-${random}`;
  }

  // ==================== Task Relationships (Orchestration) ====================

  /**
   * Add task relationship
   */
  addTaskRelationship(data: {
    predecessor_id: number;
    successor_id: number;
    relationship_type: 'dependency' | 'context' | 'upgrade' | 'parallel' | 'sequential';
    data_flow?: string;
  }): boolean {
    const stmt = this.db.prepare(`
      INSERT OR IGNORE INTO task_relationships (predecessor_id, successor_id, relationship_type, data_flow)
      VALUES (?, ?, ?, ?)
    `);

    try {
      const result = stmt.run(
        data.predecessor_id,
        data.successor_id,
        data.relationship_type,
        data.data_flow || null
      );
      return result.changes > 0;
    } catch (error) {
      console.error('Error adding task relationship:', error);
      return false;
    }
  }

  /**
   * Get task predecessors (dependencies)
   */
  getTaskPredecessors(taskId: number): any[] {
    const stmt = this.db.prepare(`
      SELECT
        tr.id,
        tr.predecessor_id,
        tr.successor_id,
        tr.relationship_type,
        tr.data_flow,
        t.title as predecessor_title,
        t.status as predecessor_status
      FROM task_relationships tr
      JOIN tasks t ON t.id = tr.predecessor_id
      WHERE tr.successor_id = ?
      ORDER BY tr.created_at
    `);

    return stmt.all(taskId);
  }

  /**
   * Get task successors (dependents)
   */
  getTaskSuccessors(taskId: number): any[] {
    const stmt = this.db.prepare(`
      SELECT
        tr.id,
        tr.predecessor_id,
        tr.successor_id,
        tr.relationship_type,
        tr.data_flow,
        t.title as successor_title,
        t.status as successor_status
      FROM task_relationships tr
      JOIN tasks t ON t.id = tr.successor_id
      WHERE tr.predecessor_id = ?
      ORDER BY tr.created_at
    `);

    return stmt.all(taskId);
  }

  /**
   * Check if task can be executed (all dependencies completed)
   */
  canExecuteTask(taskId: number): boolean {
    const stmt = this.db.prepare(`
      SELECT COUNT(*) as count
      FROM task_relationships tr
      JOIN tasks t ON t.id = tr.predecessor_id
      WHERE tr.successor_id = ?
        AND tr.relationship_type = 'dependency'
        AND t.status != 'completed'
    `);

    const result = stmt.get(taskId) as any;
    return result.count === 0;
  }

  /**
   * Get task execution graph (for orchestration)
   */
  getTaskExecutionGraph(): any[] {
    const stmt = this.db.prepare(`
      SELECT
        t.id,
        t.title,
        t.status,
        t.priority,
        COUNT(DISTINCT tr.predecessor_id) as dependency_count
      FROM tasks t
      LEFT JOIN task_relationships tr ON tr.successor_id = t.id AND tr.relationship_type = 'dependency'
      WHERE t.status IN ('pending', 'running')
      GROUP BY t.id
      ORDER BY t.priority DESC, t.created_at
    `);

    return stmt.all();
  }

  // ==================== Task Checkpoints (Short-term Memory) ====================

  /**
   * Create checkpoint
   */
  createCheckpoint(data: {
    task_id: number;
    worker_id: string;
    checkpoint_name: string;
    checkpoint_data: string;
    memory_snapshot?: string;
    state_snapshot?: string;
  }): number {
    const stmt = this.db.prepare(`
      INSERT INTO task_checkpoints (task_id, worker_id, checkpoint_name, checkpoint_data, memory_snapshot, state_snapshot)
      VALUES (?, ?, ?, ?, ?, ?)
    `);

    const result = stmt.run(
      data.task_id,
      data.worker_id,
      data.checkpoint_name,
      data.checkpoint_data,
      data.memory_snapshot || null,
      data.state_snapshot || null
    );

    return result.lastInsertRowid as number;
  }

  /**
   * Get latest checkpoint for task
   */
  getLatestCheckpoint(taskId: number): any | null {
    const stmt = this.db.prepare(`
      SELECT * FROM task_checkpoints
      WHERE task_id = ?
      ORDER BY timestamp DESC
      LIMIT 1
    `);

    const row = stmt.get(taskId) as any;
    return row || null;
  }

  /**
   * Get all checkpoints for task
   */
  getCheckpoints(taskId: number): any[] {
    const stmt = this.db.prepare(`
      SELECT * FROM task_checkpoints
      WHERE task_id = ?
      ORDER BY timestamp ASC
    `);

    return stmt.all(taskId);
  }

  /**
   * Restore task from checkpoint
   */
  restoreFromCheckpoint(checkpointId: number): any | null {
    const stmt = this.db.prepare(`
      SELECT * FROM task_checkpoints
      WHERE id = ?
    `);

    const row = stmt.get(checkpointId) as any;
    return row || null;
  }

  /**
   * Clean old checkpoints (keep last N per task)
   */
  cleanOldCheckpoints(taskId: number, keepCount: number = 5): number {
    const stmt = this.db.prepare(`
      DELETE FROM task_checkpoints
      WHERE task_id = ?
        AND id NOT IN (
          SELECT id FROM task_checkpoints
          WHERE task_id = ?
          ORDER BY timestamp DESC
          LIMIT ?
        )
    `);

    const result = stmt.run(taskId, taskId, keepCount);
    return result.changes;
  }

  // ==================== Task Versions (Upgrade Mechanism) ====================

  /**
   * Create task version
   */
  createTaskVersion(data: {
    task_id: number;
    version_number: number;
    title: string;
    description?: string;
    upgrade_reason?: string;
    upgraded_from?: number;
  }): boolean {
    const stmt = this.db.prepare(`
      INSERT INTO task_versions (task_id, version_number, title, description, upgrade_reason, upgraded_from)
      VALUES (?, ?, ?, ?, ?, ?)
    `);

    try {
      const result = stmt.run(
        data.task_id,
        data.version_number,
        data.title,
        data.description || null,
        data.upgrade_reason || null,
        data.upgraded_from || null
      );
      return result.changes > 0;
    } catch (error) {
      console.error('Error creating task version:', error);
      return false;
    }
  }

  /**
   * Get task version history
   */
  getTaskVersions(taskId: number): any[] {
    const stmt = this.db.prepare(`
      SELECT
        tv.*,
        t.title as current_title
      FROM task_versions tv
      JOIN tasks t ON t.id = tv.task_id
      WHERE tv.task_id = ?
      ORDER BY tv.version_number ASC
    `);

    return stmt.all(taskId);
  }

  /**
   * Get latest task version
   */
  getLatestTaskVersion(taskId: number): any | null {
    const stmt = this.db.prepare(`
      SELECT * FROM task_versions
      WHERE task_id = ?
      ORDER BY version_number DESC
      LIMIT 1
    `);

    const row = stmt.get(taskId) as any;
    return row || null;
  }

  /**
   * Upgrade task (create new version)
   */
  upgradeTask(data: {
    task_id: number;
    new_title: string;
    new_description?: string;
    upgrade_reason: string;
  }): boolean {
    // Get current version
    const latestStmt = this.db.prepare(`
      SELECT MAX(version_number) as max_version FROM task_versions WHERE task_id = ?
    `);
    const latest = latestStmt.get(data.task_id) as any;
    const newVersion = (latest.max_version || 0) + 1;

    // Create version record
    const versionStmt = this.db.prepare(`
      INSERT INTO task_versions (task_id, version_number, title, description, upgrade_reason, upgraded_from)
      VALUES (?, ?, ?, ?, ?, ?)
    `);

    try {
      // Get current task details
      const taskStmt = this.db.prepare('SELECT title, description FROM tasks WHERE id = ?');
      const task = taskStmt.get(data.task_id) as any;

      versionStmt.run(
        data.task_id,
        newVersion,
        data.new_title,
        data.new_description || task.description,
        data.upgrade_reason,
        data.task_id
      );

      // Update task
      const updateStmt = this.db.prepare(`
        UPDATE tasks
        SET title = ?, description = ?
        WHERE id = ?
      `);

      const updateResult = updateStmt.run(data.new_title, data.new_description || task.description, data.task_id);

      return updateResult.changes > 0;
    } catch (error) {
      console.error('Error upgrading task:', error);
      return false;
    }
  }

  // ==================== Git Locks ====================

  /**
   * Acquire git lock for file
   */
  acquireGitLock(data: {
    task_id: string;
    agent_id: string;
    file_path: string;
    lock_type: string;
  }): boolean {
    const stmt = this.db.prepare(`
      INSERT INTO git_locks (task_id, agent_id, file_path, lock_type, status)
      VALUES (?, ?, ?, ?, 'active')
    `);

    try {
      const result = stmt.run(data.task_id, data.agent_id, data.file_path, data.lock_type);
      return result.changes > 0;
    } catch (error) {
      // Lock might already exist
      return false;
    }
  }

  /**
   * Release git lock
   */
  releaseGitLock(taskId: string, filePath: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE git_locks
      SET status = 'released', released_at = CURRENT_TIMESTAMP
      WHERE task_id = ? AND file_path = ? AND status = 'active'
    `);

    const result = stmt.run(taskId, filePath);
    return result.changes > 0;
  }

  /**
   * Check if file is locked
   */
  isFileLocked(filePath: string): boolean {
    const stmt = this.db.prepare(`
      SELECT COUNT(*) as count FROM git_locks
      WHERE file_path = ? AND status = 'active'
    `);

    const result = stmt.get(filePath) as any;
    return result.count > 0;
  }

  /**
   * Get active locks for task
   */
  getActiveLocks(taskId: string): any[] {
    const stmt = this.db.prepare(`
      SELECT * FROM git_locks
      WHERE task_id = ? AND status = 'active'
      ORDER BY acquired_at
    `);

    return stmt.all(taskId);
  }

  /**
   * Release all locks for task
   */
  releaseAllLocks(taskId: string): number {
    const stmt = this.db.prepare(`
      UPDATE git_locks
      SET status = 'released', released_at = CURRENT_TIMESTAMP
      WHERE task_id = ? AND status = 'active'
    `);

    const result = stmt.run(taskId);
    return result.changes;
  }

  /**
   * Clean expired locks (older than timeout minutes)
   */
  cleanExpiredLocks(timeoutMinutes: number = 30): number {
    const stmt = this.db.prepare(`
      UPDATE git_locks
      SET status = 'expired', released_at = CURRENT_TIMESTAMP
      WHERE status = 'active'
        AND datetime(acquired_at, '+' || ? || ' minutes') < datetime('now')
    `);

    const result = stmt.run(timeoutMinutes);
    return result.changes;
  }

  // ==================== Claude Integration ====================

  /**
   * Create Claude ID mapping
   */
  createClaudeMapping(data: {
    task_id: number;
    session_uuid: string;
    message_uuid: string;
    parent_message_uuid?: string;
    slug?: string;
    source?: string;
  }): boolean {
    const stmt = this.db.prepare(`
      INSERT INTO claude_mappings
      (task_id, session_uuid, message_uuid, parent_message_uuid, slug, source)
      VALUES (?, ?, ?, ?, ?, ?)
    `);

    try {
      const result = stmt.run(
        data.task_id,
        data.session_uuid,
        data.message_uuid,
        data.parent_message_uuid || null,
        data.slug || null,
        data.source || 'api'
      );
      return result.changes > 0;
    } catch (error) {
      console.error('Error creating Claude mapping:', error);
      return false;
    }
  }

  /**
   * Get Claude mapping by AgentFlow Task ID
   */
  getClaudeMappingByTaskID(taskId: string): any | null {
    const stmt = this.db.prepare(`
      SELECT id, task_id, session_uuid, message_uuid,
             parent_message_uuid, slug, source, created_at, updated_at
      FROM claude_mappings
      WHERE task_id = ?
    `);

    const row = stmt.get(taskId) as any;
    return row || null;
  }

  /**
   * Get Claude mapping by Message UUID (reverse lookup)
   */
  getClaudeMappingByMessageUUID(messageUUID: string): any | null {
    const stmt = this.db.prepare(`
      SELECT id, task_id, session_uuid, message_uuid,
             parent_message_uuid, slug, source, created_at, updated_at
      FROM claude_mappings
      WHERE message_uuid = ?
    `);

    const row = stmt.get(messageUUID) as any;
    return row || null;
  }

  /**
   * Get task chain by Session UUID
   */
  getTaskChainBySession(sessionUUID: string): any[] {
    const stmt = this.db.prepare(`
      SELECT id, task_id, session_uuid, message_uuid,
             parent_message_uuid, slug, source, created_at, updated_at
      FROM claude_mappings
      WHERE session_uuid = ?
      ORDER BY created_at ASC
    `);

    return stmt.all(sessionUUID);
  }

  /**
   * Get Claude mapping by Slug (friendly name)
   */
  getClaudeMappingBySlug(slug: string): any | null {
    const stmt = this.db.prepare(`
      SELECT id, task_id, session_uuid, message_uuid,
             parent_message_uuid, slug, source, created_at, updated_at
      FROM claude_mappings
      WHERE slug = ?
    `);

    const row = stmt.get(slug) as any;
    return row || null;
  }

  /**
   * Update Claude mapping slug
   */
  updateClaudeMappingSlug(taskId: string, slug: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE claude_mappings
      SET slug = ?, updated_at = CURRENT_TIMESTAMP
      WHERE task_id = ?
    `);

    const result = stmt.run(slug, taskId);
    return result.changes > 0;
  }

  /**
   * Delete Claude mapping
   */
  deleteClaudeMapping(taskId: string): boolean {
    const stmt = this.db.prepare(`DELETE FROM claude_mappings WHERE task_id = ?`);
    const result = stmt.run(taskId);
    return result.changes > 0;
  }

  /**
   * List Claude mappings by session (with pagination)
   */
  listClaudeMappingsBySession(
    sessionUUID: string,
    limit: number = 50,
    offset: number = 0
  ): any[] {
    const stmt = this.db.prepare(`
      SELECT id, task_id, session_uuid, message_uuid,
             parent_message_uuid, slug, source, created_at, updated_at
      FROM claude_mappings
      WHERE session_uuid = ?
      ORDER BY created_at ASC
      LIMIT ? OFFSET ?
    `);

    return stmt.all(sessionUUID, limit, offset);
  }

  // ==================== Task Chain Management ====================

  /**
   * Create task chain
   */
  createTaskChain(data: {
    id: string;
    session_uuid: string;
    root_message_uuid: string;
    chain_type: 'sequential' | 'parallel' | 'tree';
    status?: string;
  }): boolean {
    const stmt = this.db.prepare(`
      INSERT INTO task_chains (id, session_uuid, root_message_uuid, chain_type, status)
      VALUES (?, ?, ?, ?, ?)
    `);

    try {
      const result = stmt.run(
        data.id,
        data.session_uuid,
        data.root_message_uuid,
        data.chain_type,
        data.status || 'pending'
      );
      return result.changes > 0;
    } catch (error) {
      console.error('Error creating task chain:', error);
      return false;
    }
  }

  /**
   * Create task chain node
   */
  createTaskChainNode(data: {
    chain_id: string;
    task_id: number;
    parent_node_id?: number;
    node_order: number;
  }): number {
    const stmt = this.db.prepare(`
      INSERT INTO task_chain_nodes (chain_id, task_id, parent_node_id, node_order)
      VALUES (?, ?, ?, ?)
    `);

    try {
      const result = stmt.run(
        data.chain_id,
        data.task_id,
        data.parent_node_id || null,
        data.node_order
      );
      return result.lastInsertRowid as number;
    } catch (error) {
      console.error('Error creating task chain node:', error);
      return -1;
    }
  }

  /**
   * Get task chain by ID
   */
  getTaskChain(chainId: string): any | null {
    const stmt = this.db.prepare(`
      SELECT id, session_uuid, root_message_uuid, chain_type, status,
             created_at, started_at, completed_at
      FROM task_chains
      WHERE id = ?
    `);

    const row = stmt.get(chainId) as any;
    return row || null;
  }

  /**
   * Get task chain nodes
   */
  getTaskChainNodes(chainId: string): any[] {
    const stmt = this.db.prepare(`
      SELECT id, chain_id, task_id, parent_node_id, node_order, created_at
      FROM task_chain_nodes
      WHERE chain_id = ?
      ORDER BY node_order ASC
    `);

    return stmt.all(chainId);
  }

  /**
   * Get task chains by session UUID
   */
  getTaskChainsBySession(sessionUUID: string): any[] {
    const stmt = this.db.prepare(`
      SELECT id, session_uuid, root_message_uuid, chain_type, status,
             created_at, started_at, completed_at
      FROM task_chains
      WHERE session_uuid = ?
      ORDER BY created_at DESC
    `);

    return stmt.all(sessionUUID);
  }

  /**
   * Update task chain status
   */
  updateTaskChainStatus(chainId: string, status: string): boolean {
    const stmt = this.db.prepare(`
      UPDATE task_chains
      SET status = ?,
          started_at = CASE WHEN ? = 'running' THEN CURRENT_TIMESTAMP ELSE started_at END,
          completed_at = CASE WHEN ? IN ('completed', 'failed') THEN CURRENT_TIMESTAMP ELSE completed_at END
      WHERE id = ?
    `);

    try {
      const result = stmt.run(status, status, status, chainId);
      return result.changes > 0;
    } catch (error) {
      console.error('Error updating task chain status:', error);
      return false;
    }
  }

  /**
   * Delete task chain
   */
  deleteTaskChain(chainId: string): boolean {
    const stmt = this.db.prepare(`DELETE FROM task_chains WHERE id = ?`);
    const result = stmt.run(chainId);
    return result.changes > 0;
  }
}

export default AgentFlowDatabase;

// Export memory system schema
export * from './memory-schema';
