/**
 * AgentFlow Skill
 *
 * Claude CLI skill for task orchestration and execution
 * Provides a simple interface to create and manage tasks through AgentFlow
 */

import { spawn } from 'child_process';
import { promisify } from 'util';
import { writeFile, readFile } from 'fs/promises';
import { join } from 'path';
import { tmpdir } from 'os';

const execAsync = promisify(require('child_process').exec);

export interface AgentFlowConfig {
  master_url?: string;
  group_name?: string;
}

export interface Task {
  title: string;
  description: string;
  group_name?: string;
}

export class AgentFlowSkill {
  private masterUrl: string;
  private groupName: string;

  constructor(config: AgentFlowConfig = {}) {
    this.masterUrl = config.master_url || process.env.AGENTFLOW_MASTER_URL || 'http://localhost:6767';
    this.groupName = config.group_name || process.env.AGENTFLOW_GROUP || 'default';
  }

  /**
   * Create a new task
   */
  async createTask(task: Task): Promise<string> {
    const response = await fetch(`${this.masterUrl}/api/v1/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: task.title,
        description: task.description,
        group_name: task.group_name || this.groupName,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to create task: ${response.statusText}`);
    }

    const data = (await response.json()) as any;
    return data.data.task_id;
  }

  /**
   * Execute a shell command and return the result
   */
  async executeShell(command: string): Promise<string> {
    const { stdout } = await execAsync(command, {
      timeout: 30000,
      maxBuffer: 1024 * 1024 * 10,
    });
    return stdout.trim();
  }

  /**
   * Execute multiple tasks in parallel
   */
  async executeParallel(tasks: Task[]): Promise<string[]> {
    const promises = tasks.map((task) => this.createTask(task));
    const taskIds = await Promise.all(promises);

    return taskIds;
  }

  /**
   * Execute tasks sequentially
   */
  async executeSequential(tasks: Task[]): Promise<string[]> {
    const results: string[] = [];

    for (const task of tasks) {
      const taskId = await this.createTask(task);
      results.push(taskId);
    }

    return results;
  }

  /**
   * Create a task workflow with dependencies
   */
  async createWorkflow(workflow: {
    name: string;
    tasks: Task[];
    dependencies?: Array<[number, number]>; // [from, to] task indices
  }): Promise<string> {
    // Create all tasks
    const taskIds: string[] = [];
    for (const task of workflow.tasks) {
      const taskId = await this.createTask(task);
      taskIds.push(taskId);
    }

    return `Workflow "${workflow.name}" created with ${taskIds.length} tasks`;
  }

  /**
   * Get task status
   */
  async getTaskStatus(taskId: string): Promise<any> {
    // Extract numeric ID from TASK-XXXXXXX format if needed
    const numericId = taskId.replace(/\D/g, '');
    const response = await fetch(`${this.masterUrl}/api/v1/tasks/${numericId}`);

    if (!response.ok) {
      throw new Error(`Failed to get task status: ${response.statusText}`);
    }

    const data = (await response.json()) as any;
    return data.data;
  }

  /**
   * List all tasks
   */
  async listTasks(status?: string): Promise<any[]> {
    const url = status
      ? `${this.masterUrl}/api/v1/tasks?status=${status}`
      : `${this.masterUrl}/api/v1/tasks`;

    const response = await fetch(url);

    if (!response.ok) {
      throw new Error(`Failed to list tasks: ${response.statusText}`);
    }

    const data = (await response.json()) as any;
    return data.data.tasks;
  }

  /**
   * Check Master status
   */
  async checkHealth(): Promise<boolean> {
    try {
      const response = await fetch(`${this.masterUrl}/health`);
      return response.ok;
    } catch {
      return false;
    }
  }
}

/**
 * Convenience function to create an AgentFlow skill instance
 */
export function createSkill(config?: AgentFlowConfig): AgentFlowSkill {
  return new AgentFlowSkill(config);
}

export default AgentFlowSkill;
