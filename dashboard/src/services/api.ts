/**
 * AgentFlow 分布式执行系统 API 服务
 */

import type {
  ClusterStatus,
  ClusterNode,
  WorkerInfo,
  WorkerResources,
  LockInfo,
  SystemStats,
  Workflow,
  TaskDependencyGraph,
} from '../types/distributed';
import { authenticatedFetch } from '../utils/auth';

// 使用相对路径，在生产环境中由 Rust 后端提供 API
const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// ==================== 集群管理 API ====================

export async function getClusterStatus(): Promise<ClusterStatus> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/cluster/status`);
  if (!response.ok) throw new Error('Failed to fetch cluster status');
  const data = await response.json();
  return data.data;
}

export async function getLeader(): Promise<string | null> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/cluster/leader`);
  if (!response.ok) throw new Error('Failed to fetch leader');
  const data = await response.json();
  return data.data.leader_id;
}

export async function getClusterNodes(): Promise<ClusterNode[]> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/cluster/nodes`);
  if (!response.ok) throw new Error('Failed to fetch cluster nodes');
  const data = await response.json();
  return data.data.nodes;
}

// ==================== Worker 管理 API ====================

export async function getWorkers(): Promise<WorkerInfo[]> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workers`);
  if (!response.ok) throw new Error('Failed to fetch workers');
  const data = await response.json();
  return data.data.workers;
}

export async function getWorker(workerId: string): Promise<WorkerInfo> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workers/${workerId}`);
  if (!response.ok) throw new Error('Failed to fetch worker');
  const data = await response.json();
  return data.data;
}

export async function updateWorkerStatus(
  workerId: string,
  status: 'Active' | 'Busy' | 'Offline' | 'Draining'
): Promise<void> {
  await authenticatedFetch(`${API_BASE_URL}/api/v1/workers/${workerId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  });
}

export async function getWorkerResources(workerId: string): Promise<WorkerResources> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workers/${workerId}/resources`);
  if (!response.ok) throw new Error('Failed to fetch worker resources');
  const data = await response.json();
  return data.data;
}

// ==================== 工作流 API ====================

export async function createWorkflow(workflow: Workflow): Promise<string> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workflows`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(workflow),
  });
  if (!response.ok) throw new Error('Failed to create workflow');
  const data = await response.json();
  return data.data.workflow_id;
}

export async function getWorkflow(name: string): Promise<Workflow> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workflows/${name}`);
  if (!response.ok) throw new Error('Failed to fetch workflow');
  const data = await response.json();
  return data.data;
}

export async function getWorkflows(): Promise<Workflow[]> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workflows`);
  if (!response.ok) throw new Error('Failed to fetch workflows');
  const data = await response.json();
  return data.data.workflows;
}

export async function getWorkflowGraph(name: string): Promise<TaskDependencyGraph> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workflows/${name}/graph`);
  if (!response.ok) throw new Error('Failed to fetch workflow graph');
  const data = await response.json();
  return data.data;
}

export async function getWorkflowExecutionOrder(name: string): Promise<string[]> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/workflows/${name}/order`);
  if (!response.ok) throw new Error('Failed to fetch execution order');
  const data = await response.json();
  return data.data.order;
}

// ==================== 任务 API ====================

export async function createTask(task: {
  title: string;
  description: string;
  priority?: 'Low' | 'Medium' | 'High' | 'Urgent';
  dependencies?: string[];
  group_name?: string;
}): Promise<string> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/tasks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(task),
  });
  if (!response.ok) throw new Error('Failed to create task');
  const data = await response.json();
  return data.data.task_id;
}

export async function getTask(taskId: string): Promise<any> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/tasks/${taskId}`);
  if (!response.ok) throw new Error('Failed to fetch task');
  const data = await response.json();
  return data.data;
}

export async function getTasks(filters?: {
  status?: string;
  group_name?: string;
  limit?: number;
}): Promise<any[]> {
  const params = new URLSearchParams();
  if (filters?.status) params.set('status', filters.status);
  if (filters?.group_name) params.set('group_name', filters.group_name);
  if (filters?.limit) params.set('limit', filters.limit.toString());

  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/tasks?${params}`);
  if (!response.ok) throw new Error('Failed to fetch tasks');
  const data = await response.json();
  return data.data.tasks;
}

export async function updateTaskPriority(
  taskId: string,
  priority: 'Low' | 'Medium' | 'High' | 'Urgent'
): Promise<void> {
  await authenticatedFetch(`${API_BASE_URL}/api/v1/tasks/${taskId}/priority`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ priority }),
  });
}

export async function cancelTask(taskId: string): Promise<void> {
  await authenticatedFetch(`${API_BASE_URL}/api/v1/tasks/${taskId}`, {
    method: 'DELETE',
  });
}

// ==================== 任务队列 API ====================

export async function getQueueStats(): Promise<QueueStats> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/queue/stats`);
  if (!response.ok) throw new Error('Failed to fetch queue stats');
  const data = await response.json();
  return data.data;
}

// ==================== 分布式锁 API ====================

export async function acquireLock(
  lockKey: string,
  ttl?: number,
  metadata?: any
): Promise<boolean> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/locks/acquire`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lock_key: lockKey, ttl, metadata }),
  });
  if (!response.ok) throw new Error('Failed to acquire lock');
  const data = await response.json();
  return data.data.acquired;
}

export async function releaseLock(lockKey: string): Promise<boolean> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/locks/release`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lock_key: lockKey }),
  });
  if (!response.ok) throw new Error('Failed to release lock');
  const data = await response.json();
  return data.data.released;
}

export async function getLockStatus(lockKey: string): Promise<LockStatus> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/locks/${lockKey}`);
  if (!response.ok) throw new Error('Failed to fetch lock status');
  const data = await response.json();
  return data.data;
}

export async function getAllLocks(): Promise<LockInfo[]> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/locks`);
  if (!response.ok) throw new Error('Failed to fetch locks');
  const data = await response.json();
  return data.data.locks;
}

// ==================== 系统统计 API ====================

export async function getSystemStats(): Promise<SystemStats> {
  const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/stats`);
  if (!response.ok) throw new Error('Failed to fetch system stats');
  const data = await response.json();
  return data.data;
}

// ==================== WebSocket 连接 ====================

export function connectWebSocket(onMessage: (message: DistributedServerMessage) => void): WebSocket {
  // 构建 WebSocket URL
  let wsUrl: string;
  if (API_BASE_URL) {
    // 开发环境：使用绝对路径
    wsUrl = API_BASE_URL.replace('http', 'ws') + '/api/v1/stream';
  } else {
    // 生产环境：使用相对路径
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    wsUrl = `${protocol}//${window.location.host}/ws/task`;
  }

  const ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log('WebSocket connected');
  };

  ws.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data);
      onMessage(message);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  ws.onclose = () => {
    console.log('WebSocket disconnected');
  };

  return ws;
}
