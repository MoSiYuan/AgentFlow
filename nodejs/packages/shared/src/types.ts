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

/**
 * Task relationship types
 */
export type TaskRelationshipType =
  | 'dependency'  // Successor depends on predecessor (predecessor must complete first)
  | 'context'     // Successor uses predecessor's output as context
  | 'upgrade'     // Successor is an upgraded version of predecessor
  | 'parallel'    // Tasks can run in parallel
  | 'sequential'; // Tasks must run in sequence

/**
 * Task relationship
 */
export interface TaskRelationship {
  id: number;
  predecessor_id: number;
  successor_id: number;
  relationship_type: TaskRelationshipType;
  data_flow?: string;  // JSON string describing data flow
  created_at: Date;
}

/**
 * Task checkpoint (short-term memory)
 */
export interface TaskCheckpoint {
  id: number;
  task_id: number;
  worker_id: string;
  checkpoint_name: string;
  checkpoint_data: string;    // JSON string
  memory_snapshot?: string;   // JSON string of agent's short-term memory
  state_snapshot?: string;    // JSON string of execution state
  timestamp: Date;
}

/**
 * Task version (upgrade mechanism)
 */
export interface TaskVersion {
  id: number;
  task_id: number;
  version_number: number;
  title: string;
  description?: string;
  upgrade_reason?: string;
  upgraded_from?: number;     // ID of the original task
  created_at: Date;
}

/**
 * Orchestration modes (Ralph pattern)
 */
export type OrchestrationMode =
  | 'sequential'   // Execute tasks one by one
  | 'parallel'     // Execute tasks concurrently
  | 'dag'          // Directed Acyclic Graph - respect dependencies
  | 'conditional'  // Branch based on conditions
  | 'pipeline';    // Pipeline with data flow

/**
 * Orchestration graph node
 */
export interface OrchestrationNode {
  task_id: number;
  title: string;
  status: TaskStatus;
  priority: TaskPriority;
  dependencies: number[];      // Task IDs that must complete first
  dependents: number[];        // Task IDs that depend on this
  can_execute: boolean;        // Whether all dependencies are met
}

/**
 * Orchestration plan
 */
export interface OrchestrationPlan {
  mode: OrchestrationMode;
  tasks: OrchestrationNode[];
  execution_order: number[][];  // Array of task IDs that can run in parallel
  total_tasks: number;
  ready_tasks: number;
  completed_tasks: number;
  estimated_completion?: Date;
}

/**
 * Checkpoint restore options
 */
export interface CheckpointRestoreOptions {
  restore_memory: boolean;      // Restore short-term memory
  restore_state: boolean;       // Restore execution state
  continue_from: 'checkpoint' | 'beginning';
}

/**
 * Task upgrade request
 */
export interface TaskUpgradeRequest {
  task_id: number;
  new_title: string;
  new_description?: string;
  upgrade_reason: string;
  preserve_relationships: boolean;
}
