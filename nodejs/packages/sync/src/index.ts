/**
 * Task Synchronizer for AgentFlow Node.js
 *
 * Handles synchronization between AgentFlow and Claude CLI
 */

import { readFileSync, writeFileSync, openSync, closeSync, writeSync, statSync, mkdirSync, existsSync } from 'fs';
import { join, dirname } from 'path';

export type SyncDirection = 'to_claude' | 'from_claude' | 'bidirectional';

export interface ClaudeStatusUpdate {
  type: string;
  timestamp: string;
  data: ClaudeStatusUpdateData;
}

export interface ClaudeStatusUpdateData {
  task_id: string;
  message_uuid: string;
  status: string;
  result?: string;
  error?: string;
}

export interface SynchronizerConfig {
  claudeDir: string;
  autoSync?: boolean;
  syncInterval?: number; // milliseconds
  logger?: any;
}

/**
 * TaskSynchronizer handles synchronization between AgentFlow and Claude
 */
export class TaskSynchronizer {
  private claudeDir: string;
  private autoSync: boolean;
  private syncInterval: number;
  private logger?: any;
  private syncTimer?: NodeJS.Timeout;

  constructor(private db: any, config: SynchronizerConfig) {
    this.claudeDir = config.claudeDir;
    this.autoSync = config.autoSync !== false;
    this.syncInterval = config.syncInterval || 5000; // 5 seconds default
    this.logger = config.logger;
  }

  /**
   * Set auto-sync enabled state
   */
  setAutoSync(enabled: boolean): void {
    this.autoSync = enabled;
  }

  /**
   * Set sync interval
   */
  setSyncInterval(interval: number): void {
    this.syncInterval = interval;
  }

  /**
   * Synchronize task status between AgentFlow and Claude
   */
  async syncTaskStatus(
    taskID: string,
    direction: SyncDirection
  ): Promise<void> {
    // Get task from database
    const task = this.db.getTask(taskID);
    if (!task) {
      throw new Error(`Task not found: ${taskID}`);
    }

    // Get Claude mapping
    const mapping = this.db.getClaudeMappingByTaskID(taskID);
    if (!mapping) {
      this.log('warn', `No Claude mapping found for task ${taskID}`);
      return;
    }

    switch (direction) {
      case 'to_claude':
        await this.syncToClaude(task, mapping);
        break;
      case 'from_claude':
        await this.syncFromClaude(mapping);
        break;
      case 'bidirectional':
        await this.syncToClaude(task, mapping);
        await this.syncFromClaude(mapping);
        break;
      default:
        throw new Error(`Invalid sync direction: ${direction}`);
    }
  }

  /**
   * Sync AgentFlow task status to Claude session file
   */
  private async syncToClaude(
    task: any,
    mapping: any
  ): Promise<void> {
    const sessionFile = this.getClaudeSessionFile(mapping.session_uuid);

    // Create status update
    const update: ClaudeStatusUpdate = {
      type: 'agentflow_status_update',
      timestamp: new Date().toISOString(),
      data: {
        task_id: mapping.task_id,
        message_uuid: mapping.message_uuid,
        status: task.status
      }
    };

    // Marshal to JSON
    const data = JSON.stringify(update);

    // Append to session file (JSON Lines format)
    try {
      await this.appendToFile(sessionFile, data);
      this.log('info', `Synced task ${task.id} status to Claude: ${task.status}`);
    } catch (error: any) {
      throw new Error(`Failed to write to session file: ${error.message}`);
    }
  }

  /**
   * Sync status updates from Claude session file to AgentFlow
   */
  private async syncFromClaude(mapping: any): Promise<void> {
    const sessionFile = this.getClaudeSessionFile(mapping.session_uuid);

    // Check if file exists
    if (!existsSync(sessionFile)) {
      this.log('warn', `Claude session file not found: ${sessionFile}`);
      return;
    }

    try {
      // Read and parse status updates
      const updates = this.readStatusUpdates(sessionFile);

      // Process updates for this task
      for (const update of updates) {
        if (update.data.message_uuid === mapping.message_uuid) {
          // Update task status in AgentFlow
          try {
            await this.updateTaskStatusInAgentFlow(mapping.task_id, update.data.status);
            this.log('info', `Synced task ${mapping.task_id} status from Claude: ${update.data.status}`);
          } catch (error: any) {
            this.log('error', `Failed to update task status: ${error.message}`);
          }
        }
      }
    } catch (error: any) {
      throw new Error(`Failed to read status updates: ${error.message}`);
    }
  }

  /**
   * Get Claude session file path
   */
  private getClaudeSessionFile(sessionUUID: string): string {
    return join(this.claudeDir, `${sessionUUID}.jsonl`);
  }

  /**
   * Append data to file (JSON Lines format)
   */
  private async appendToFile(filePath: string, data: string): Promise<void> {
    // Ensure directory exists
    const dir = dirname(filePath);
    if (!existsSync(dir)) {
      mkdirSync(dir, { recursive: true, mode: 0o755 });
    }

    // Open file in append mode
    const fd = openSync(filePath, 'a');
    try {
      // Write data with newline
      writeSync(fd, data + '\n');
    } finally {
      closeSync(fd);
    }
  }

  /**
   * Read status updates from Claude session file
   */
  private readStatusUpdates(filePath: string): ClaudeStatusUpdate[] {
    const content = readFileSync(filePath, 'utf-8');
    const lines = content.split('\n').filter(line => line.trim() !== '');

    const updates: ClaudeStatusUpdate[] = [];

    for (const line of lines) {
      try {
        const update: ClaudeStatusUpdate = JSON.parse(line);
        // Only process agentflow_status_update types
        if (update.type === 'agentflow_status_update') {
          updates.push(update);
        }
      } catch (error: any) {
        this.log('warn', `Failed to decode status update: ${error.message}`);
      }
    }

    return updates;
  }

  /**
   * Update task status in AgentFlow database
   */
  private async updateTaskStatusInAgentFlow(taskID: string, status: string): Promise<void> {
    const validStatuses = ['pending', 'in_progress', 'completed', 'failed'];

    if (!validStatuses.includes(status)) {
      throw new Error(`Invalid status: ${status}`);
    }

    this.db.updateTaskStatus(taskID, status);
  }

  /**
   * Synchronize entire task chain status
   */
  async syncTaskChain(chainID: string, direction: SyncDirection): Promise<void> {
    const nodes = this.db.getTaskChainNodes(chainID);

    for (const node of nodes) {
      const taskID = String(node.task_id);
      try {
        await this.syncTaskStatus(taskID, direction);
      } catch (error: any) {
        this.log('error', `Failed to sync task ${taskID}: ${error.message}`);
      }
    }

    this.log('info', `Synced ${nodes.length} tasks in chain ${chainID}`);
  }

  /**
   * Start auto-sync in the background
   */
  startAutoSync(): void {
    if (!this.autoSync) {
      this.log('info', 'Auto-sync is disabled');
      return;
    }

    if (this.syncTimer) {
      this.log('warn', 'Auto-sync is already running');
      return;
    }

    this.log('info', 'Starting auto-sync');

    this.syncTimer = setInterval(async () => {
      try {
        await this.syncAllActiveTasks();
      } catch (error: any) {
        this.log('error', `Auto-sync failed: ${error.message}`);
      }
    }, this.syncInterval);
  }

  /**
   * Stop auto-sync
   */
  stopAutoSync(): void {
    if (this.syncTimer) {
      clearInterval(this.syncTimer);
      this.syncTimer = undefined;
      this.log('info', 'Auto-sync stopped');
    }
  }

  /**
   * Synchronize all active tasks
   */
  private async syncAllActiveTasks(): Promise<void> {
    // Get all tasks with in_progress status
    const tasks = this.db.getTasksByStatus('in_progress');

    for (const task of tasks) {
      // Check if task has Claude mapping
      const mapping = this.db.getClaudeMappingByTaskID(task.id);
      if (!mapping) {
        continue;
      }

      // Sync bidirectionally
      try {
        await this.syncTaskStatus(task.id, 'bidirectional');
      } catch (error: any) {
        this.log('error', `Failed to sync task ${task.id}: ${error.message}`);
      }
    }
  }

  /**
   * Get synchronization status
   */
  getSyncStatus(): any {
    return {
      auto_sync: this.autoSync,
      sync_interval: this.syncInterval,
      claude_dir: this.claudeDir,
      is_running: !!this.syncTimer
    };
  }

  /**
   * Validate Claude session file accessibility
   */
  validateClaudeSession(sessionUUID: string): { valid: boolean; error?: string } {
    const sessionFile = this.getClaudeSessionFile(sessionUUID);

    try {
      // Check if file exists
      if (!existsSync(sessionFile)) {
        return { valid: false, error: `Session file does not exist: ${sessionFile}` };
      }

      // Check if file is readable
      const stats = statSync(sessionFile);
      if (!stats.isFile()) {
        return { valid: false, error: `Path is not a file: ${sessionFile}` };
      }

      return { valid: true };
    } catch (error: any) {
      return { valid: false, error: error.message };
    }
  }

  /**
   * Log helper
   */
  private log(level: string, message: string, ...args: any[]): void {
    if (this.logger) {
      this.logger[level](message, ...args);
    } else {
      console[level](`[TaskSynchronizer] ${message}`, ...args);
    }
  }
}

export default TaskSynchronizer;
