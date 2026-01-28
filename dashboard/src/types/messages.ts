export type ClientMessage =
  | { type: 'AUTH'; token: string }
  | { type: 'CHAT'; text: string; targetContext: string }
  | { type: 'SUBSCRIBE'; topics: string[] };

export type ServerMessage =
  | { type: 'NODE_STATUS'; node: NodeStatus }
  | { type: 'TASK_UPDATE'; task: TaskInfo }
  | { type: 'CHAT_RESPONSE'; text: string }
  | { type: 'ALERT'; level: 'error' | 'warning'; message: string };

export interface NodeStatus {
  node_id: string;
  mode: 'leader' | 'master';
  cpu_usage: number;
  memory_usage: number;
  active_tasks: number;
  last_updated: string;
}

export interface TaskInfo {
  task_id: string;
  project_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: string;
  created_at: string;
}
