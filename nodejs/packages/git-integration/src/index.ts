/**
 * Git Integration for AgentFlow (Node.js)
 *
 * Provides Git branch management with Claude integration
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import { writeFileSync, readFileSync } from 'fs';
import path from 'path';

const execAsync = promisify(exec);

export interface BranchManagerConfig {
  repoPath: string;
  logger?: any;
}

/**
 * BranchManager manages Git branches
 */
export class BranchManager {
  private repoPath: string;
  private logger?: any;

  constructor(config: BranchManagerConfig) {
    this.repoPath = config.repoPath;
    this.logger = config.logger;
  }

  /**
   * Create a new branch
   */
  async createBranch(branchName: string): Promise<boolean> {
    try {
      const { stdout, stderr } = await execAsync(`git branch ${branchName}`, {
        cwd: this.repoPath
      });

      if (stderr && !stderr.includes('already exists')) {
        throw new Error(stderr);
      }

      this.log('info', `Branch created: ${branchName}`);
      return true;
    } catch (error: any) {
      this.log('error', `Failed to create branch: ${error.message}`);
      return false;
    }
  }

  /**
   * Delete a branch
   */
  async deleteBranch(branchName: string, force = false): Promise<boolean> {
    try {
      const flag = force ? '-D' : '-d';
      await execAsync(`git branch ${flag} ${branchName}`, {
        cwd: this.repoPath
      });

      this.log('info', `Branch deleted: ${branchName}`);
      return true;
    } catch (error: any) {
      this.log('error', `Failed to delete branch: ${error.message}`);
      return false;
    }
  }

  /**
   * Switch to a branch
   */
  async switchBranch(branchName: string): Promise<boolean> {
    try {
      await execAsync(`git checkout ${branchName}`, {
        cwd: this.repoPath
      });

      this.log('info', `Switched to branch: ${branchName}`);
      return true;
    } catch (error: any) {
      this.log('error', `Failed to switch branch: ${error.message}`);
      return false;
    }
  }

  /**
   * Get current branch
   */
  async getCurrentBranch(): Promise<string> {
    try {
      const { stdout } = await execAsync('git rev-parse --abbrev-ref HEAD', {
        cwd: this.repoPath
      });
      return stdout.trim();
    } catch (error: any) {
      this.log('error', `Failed to get current branch: ${error.message}`);
      return 'main';
    }
  }

  /**
   * Generate Claude-friendly branch name
   * Format: claude-{session_uuid前8位}/task-{taskID}
   */
  generateClaudeBranchName(sessionUUID: string, taskID: string): string {
    const prefix = sessionUUID.substring(0, 8);
    return `claude-${prefix}/task-${taskID}`;
  }

  /**
   * Log helper
   */
  private log(level: string, message: string, ...args: any[]): void {
    if (this.logger) {
      this.logger[level](message, ...args);
    } else {
      console[level](message, ...args);
    }
  }
}

/**
 * ClaudeGitTask represents a Git task with Claude integration
 */
export interface ClaudeGitTask {
  id: string;
  title: string;
  description?: string;
  agent_id: string;
  git_branch: string;
  session_uuid: string;
  message_uuid: string;
  parent_message_uuid?: string;
  slug?: string;
  status?: string;
  created_at: Date;
}

/**
 * GitIntegrationManager manages Git integration for AgentFlow Node.js
 */
export class GitIntegrationManager {
  private branchManager: BranchManager;
  private repoPath: string;

  constructor(repoPath: string, logger?: any) {
    this.repoPath = repoPath;
    this.branchManager = new BranchManager({
      repoPath,
      logger
    });
  }

  /**
   * Create Claude Git task with branch
   */
  async createClaudeGitTask(task: {
    id: string;
    title: string;
    description?: string;
    agent_id: string;
    session_uuid: string;
    message_uuid: string;
    parent_message_uuid?: string;
    slug?: string;
  }): Promise<ClaudeGitTask> {
    // Generate branch name
    const branchName = this.branchManager.generateClaudeBranchName(
      task.session_uuid,
      task.id
    );

    // Create the branch
    await this.branchManager.createBranch(branchName);

    const claudeTask: ClaudeGitTask = {
      id: task.id,
      title: task.title,
      description: task.description,
      agent_id: task.agent_id,
      git_branch: branchName,
      session_uuid: task.session_uuid,
      message_uuid: task.message_uuid,
      parent_message_uuid: task.parent_message_uuid,
      slug: task.slug,
      status: 'pending',
      created_at: new Date()
    };

    this.log('info', `Claude Git task created`, {
      task_id: task.id,
      session_uuid: task.session_uuid,
      branch: branchName
    });

    return claudeTask;
  }

  /**
   * Create branches from task chain
   */
  async createBranchesFromChain(
    chainId: string,
    getTaskChainNodes: (chainId: string) => Promise<any[]>,
    getSessionUUID: (taskId: string) => Promise<string>
  ): Promise<string[]> {
    const nodes = await getTaskChainNodes(chainId);
    const branches: string[] = [];

    for (const node of nodes) {
      const sessionUUID = await getSessionUUID(String(node.task_id));
      if (!sessionUUID) {
        this.log('warn', `No session UUID found for task ${node.task_id}`);
        continue;
      }

      const branchName = this.branchManager.generateClaudeBranchName(
        sessionUUID,
        String(node.task_id)
      );

      const created = await this.branchManager.createBranch(branchName);
      if (created) {
        branches.push(branchName);
      }
    }

    this.log('info', `Created ${branches.length} branches from chain ${chainId}`);
    return branches;
  }

  /**
   * Get all branches for a Claude session
   */
  async getClaudeBranches(sessionUUID: string): Promise<string[]> {
    // This would typically query the database for all tasks with this session
    // For now, return empty array as implementation depends on database context
    return [];
  }

  /**
   * Cleanup Claude branches for a session
   */
  async cleanupClaudeBranches(sessionUUID: string): Promise<number> {
    const branches = await this.getClaudeBranches(sessionUUID);
    let deletedCount = 0;

    for (const branch of branches) {
      const deleted = await this.branchManager.deleteBranch(branch);
      if (deleted) {
        deletedCount++;
      }
    }

    this.log('info', `Cleaned up ${deletedCount} branches for session ${sessionUUID}`);
    return deletedCount;
  }

  /**
   * Log helper
   */
  private log(level: string, message: string, ...args: any[]): void {
    console[level](`[GitIntegration] ${message}`, ...args);
  }
}

export default GitIntegrationManager;
