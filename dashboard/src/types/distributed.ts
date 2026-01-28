/**
 * 分布式执行系统类型定义
 */

// ==================== Raft 集群状态 ====================

export interface ClusterStatus {
  leader_id: string | null;
  nodes: ClusterNode[];
  term: number;
  total_nodes: number;
  healthy_nodes: number;
}

export interface ClusterNode {
  node_id: string;
  state: 'Leader' | 'Follower' | 'Candidate';
  term: number;
  voted_for: string | null;
  last_leader_contact: string;
  peers: string[];
}

// ==================== Worker 信息 ====================

export interface WorkerInfo {
  id: string;
  name: string;
  group_name: string | null;
  status: 'Active' | 'Busy' | 'Offline' | 'Draining';
  resources: WorkerResources;
  capabilities: string[];
  registered_at: string;
  last_heartbeat: string;
}

export interface WorkerResources {
  cpu_cores: number;
  total_memory_mb: number;
  available_memory_mb: number;
  gpu_count: number;
  concurrent_tasks: number;
  max_tasks: number;
  custom_attributes: Record<string, string>;
}

export interface WorkerHealth {
  worker_id: string;
  status: 'healthy' | 'unhealthy' | 'unknown';
  last_check: string;
  consecutive_failures: number;
  error_message: string | null;
}

// ==================== 工作流和任务 ====================

export interface Workflow {
  name: string;
  tasks: WorkflowTask[];
  created_at: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
}

export interface WorkflowTask {
  id: string;
  name: string;
  priority: 'Low' | 'Medium' | 'High' | 'Urgent';
  dependencies: string[];
  estimated_duration_secs: number;
  status: 'Pending' | 'Ready' | 'Running' | 'Completed' | 'Failed' | 'Skipped';
  result: any;
  error: string | null;
}

export interface TaskDependencyGraph {
  tasks: Map<string, WorkflowTask>;
  execution_order: string[];
  critical_path: string[];
  ready_tasks: string[];
  completed_tasks: string[];
  failed_tasks: string[];
}

// ==================== 任务队列 ====================

export interface QueueStats {
  total_tasks: number;
  pending_tasks: number;
  running_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  avg_wait_time_ms: number;
  priority_distribution: {
    urgent: number;
    high: number;
    medium: number;
    low: number;
  };
}

export interface QueuedTask {
  task_id: string;
  priority: number;
  enqueued_at: string;
  waiting_duration_secs: number;
  retry_count: number;
}

// ==================== Agent 通信 ====================

export interface AgentMessage {
  id: string;
  from: string;
  to: string;
  msg_type: MessageType;
  content: any;
  timestamp: string;
  requires_response: boolean;
  correlation_id: string | null;
  ttl: number | null;
}

export type MessageType =
  | 'TaskRequest'
  | 'TaskResponse'
  | 'StatusUpdate'
  | 'ResourceQuery'
  | 'ResourceResponse'
  | 'CollaborationRequest'
  | 'CollaborationResponse'
  | 'Notification'
  | 'Heartbeat'
  | 'Error';

// ==================== 分布式锁 ====================

export interface LockInfo {
  lock_key: string;
  owner: string;
  acquired_at: string;
  expires_at: string;
  metadata: any;
}

export interface LockStatus {
  locked: boolean;
  lock_info: LockInfo | null;
  can_acquire: boolean;
}

// ==================== 系统统计 ====================

export interface SystemStats {
  cluster: ClusterStats;
  workers: WorkerStats;
  tasks: TaskStats;
  locks: LockStats;
}

export interface ClusterStats {
  total_nodes: number;
  leader_id: string | null;
  current_term: number;
  election_timeout_ms: number;
  heartbeat_interval_ms: number;
}

export interface WorkerStats {
  total_workers: number;
  active_workers: number;
  busy_workers: number;
  offline_workers: number;
  total_capacity: number;
  used_capacity: number;
}

export interface TaskStats {
  total_tasks: number;
  pending_tasks: number;
  running_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  avg_execution_time_ms: number;
}

export interface LockStats {
  total_locks: number;
  active_locks: number;
  expired_locks: number;
}

// ==================== API 响应类型 ====================

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// ==================== WebSocket 消息类型扩展 ====================

export type DistributedServerMessage =
  | { type: 'CLUSTER_STATUS'; status: ClusterStatus }
  | { type: 'WORKER_REGISTERED'; worker: WorkerInfo }
  | { type: 'WORKER_UPDATED'; worker_id: string; resources: WorkerResources }
  | { type: 'WORKER_UNREGISTERED'; worker_id: string }
  | { type: 'WORKFLOW_CREATED'; workflow: Workflow }
  | { type: 'WORKFLOW_UPDATED'; workflow: Workflow }
  | { type: 'TASK_QUEUED'; task: QueuedTask }
  | { type: 'TASK_STARTED'; task_id: string; worker_id: string }
  | { type: 'TASK_COMPLETED'; task_id: string; result: any }
  | { type: 'TASK_FAILED'; task_id: string; error: string }
  | { type: 'LEADER_CHANGED'; leader_id: string; term: number }
  | { type: 'LOCK_ACQUIRED'; lock_key: string; owner: string }
  | { type: 'LOCK_RELEASED'; lock_key: string }
  | { type: 'AGENT_MESSAGE'; message: AgentMessage }
  | { type: 'ALERT'; level: 'error' | 'warning'; message: string };
