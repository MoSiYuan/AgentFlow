/**
 * AgentFlow Shared Type Definitions
 *
 * This file contains all type definitions used across the AgentFlow system.
 * All language versions (Python/Go/Node.js) should use equivalent types.
 */

/**
 * Task status enumeration
 */
export type TaskStatus =
  | 'pending'     // Task is waiting to be executed
  | 'running'     // Task is currently being executed
  | 'completed'   // Task completed successfully
  | 'failed'      // Task failed with error
  | 'blocked';    // Task is blocked (waiting for dependencies)

/**
 * Task priority levels
 */
export type TaskPriority = 'low' | 'medium' | 'high';

/**
 * Worker type
 */
export type WorkerType = 'local' | 'remote';

/**
 * Worker status
 */
export type WorkerStatus = 'active' | 'inactive';

/**
 * Lock type for file locking
 */
export type LockType = 'read' | 'write';

/**
 * Access type for file boundaries
 */
export type AccessType = 'exclusive' | 'shared' | 'readonly';

/**
 * Merge strategy for Git integration
 */
export type MergeStrategy = 'merge' | 'squash' | 'rebase';

/**
 * Event types for WebSocket
 */
export type EventType =
  | 'task.created'
  | 'task.assigned'
  | 'task.started'
  | 'task.progress'
  | 'task.completed'
  | 'task.failed'
  | 'worker.registered'
  | 'worker.online'
  | 'worker.offline'
  | 'git.conflict'
  | 'git.merged';

/**
 * Task interface
 */
export interface Task {
  id: string;
  parent_id?: string;
  title: string;
  description: string;
  group_name: string;
  completion_criteria?: string;
  status: TaskStatus;
  priority: TaskPriority;
  lock_holder?: string;
  lock_time?: Date;
  result?: string;
  error?: string;
  workspace_dir?: string;
  sandboxed: boolean;
  allow_network: boolean;
  max_memory: string;
  max_cpu: number;
  created_by?: string;
  created_at: Date;
  started_at?: Date;
  completed_at?: Date;
}

/**
 * Worker interface
 */
export interface Worker {
  id: string;
  group_name: string;
  type: WorkerType;
  capabilities: string[];
  status: WorkerStatus;
  last_heartbeat: Date;
  created_at: Date;
}

/**
 * Worker registration data
 */
export interface WorkerRegistration {
  worker_id?: string;
  worker_name: string;
  group_name: string;
  platform: NodeJS.Platform;
  capabilities: string[];
}

/**
 * Task creation request
 */
export interface CreateTaskRequest {
  parent_id?: string;
  title: string;
  description: string;
  group_name?: string;
  completion_criteria?: string;
  priority?: TaskPriority;
  workspace_dir?: string;
  sandboxed?: boolean;
  allow_network?: boolean;
  max_memory?: string;
  max_cpu?: number;
}

/**
 * Task result
 */
export interface TaskResult {
  success: boolean;
  result?: string;
  error?: string;
  output?: string;
  duration_ms: number;
}

/**
 * Task progress update
 */
export interface TaskProgress {
  task_id: string;
  progress: number; // 0-100
  message: string;
}

/**
 * File boundary definition
 */
export interface FileBoundary {
  file_pattern: string;
  access_type: AccessType;
  line_range?: string;
  description: string;
}

/**
 * File lock
 */
export interface FileLock {
  file_path: string;
  agent_id: string;
  lock_type: LockType;
  acquired_at: Date;
  status: 'active' | 'released';
}

/**
 * Git conflict
 */
export interface GitConflict {
  conflict_id: string;
  task_id: string;
  agent_id: string;
  conflict_type: 'file_locked' | 'boundary_overlap' | 'merge_conflict';
  file_paths: string[];
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  status: 'pending' | 'resolving' | 'resolved';
}

/**
 * Git task
 */
export interface GitTask {
  id: string;
  title?: string;
  description?: string;
  agent_id: string;
  git_branch: string;
  file_boundaries: FileBoundary[];
  status: TaskStatus;
  created_at: Date;
  assigned_at?: Date;
  completed_at?: Date;
}

/**
 * System statistics
 */
export interface SystemStats {
  total_tasks: number;
  pending_tasks: number;
  running_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  total_workers: number;
  active_workers: number;
  uptime_seconds: number;
}

/**
 * Group statistics
 */
export interface GroupStats {
  group_name: string;
  total_tasks: number;
  pending_tasks: number;
  running_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
}

/**
 * Event interface
 */
export interface Event {
  type: EventType;
  data: any;
  timestamp: Date;
}

/**
 * Task filter for queries
 */
export interface TaskFilter {
  status?: TaskStatus;
  group_name?: string;
  parent_id?: string;
  limit?: number;
  offset?: number;
}

/**
 * Master configuration
 */
export interface MasterConfig {
  host: string;
  port: number;
  db_path: string;
  auto_shutdown: boolean;
  log_level: 'debug' | 'info' | 'warn' | 'error';
  git_enabled: boolean;
  git_boundary_config: string;
  websocket_enabled: boolean;
  websocket_port: number;
}

/**
 * Worker configuration
 */
export interface WorkerConfig {
  id?: string;
  master_url: string;
  group_name: string;
  mode: 'auto' | 'manual' | 'oneshot';
  heartbeat_interval: number;
  heartbeat_timeout: number;
  max_concurrent: number;
  task_timeout: number;
  retry_on_failure: boolean;
  max_retries: number;
}

/**
 * Executor configuration
 */
export interface ExecutorConfig {
  type: 'http' | 'claude_cli' | 'shell';
  enabled: boolean;
  priority: number;
  config?: Record<string, any>;
}

/**
 * HTTP response wrapper
 */
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

/**
 * Health check response
 */
export interface HealthResponse {
  status: 'ok' | 'error';
  version: string;
  uptime: number;
  mode: 'standalone' | 'cloud';
}
