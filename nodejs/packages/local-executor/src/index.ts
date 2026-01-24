/**
 * AgentFlow Local Executor
 *
 * Automatically manages Master and Worker for local task execution
 * - Starts Master server
 * - Starts Worker
 * - Creates tasks
 * - Monitors completion
 * - Shuts down when all tasks complete
 */

import { spawn, ChildProcess } from 'child_process';
import { Worker } from '@agentflow/worker';
import { AgentFlowSkill } from '@agentflow/skill';
import type { Task, WorkerConfig } from '@agentflow/shared';

export interface LocalExecutorConfig {
  masterPath: string;
  masterHost?: string;
  masterPort?: number;
  dbPath?: string;
  workerConfig?: WorkerConfig;
  shutdownOnComplete?: boolean;
  shutdownTimeout?: number; // milliseconds
}

export interface TaskSpec {
  title: string;
  description: string;
  priority?: number;
  group_name?: string;
}

export class LocalExecutor {
  private masterPath: string;
  private masterURL: string;
  private masterProcess?: ChildProcess;
  private worker?: Worker;
  private skill: AgentFlowSkill;
  private shutdownOnComplete: boolean;
  private shutdownTimeout: number;

  private running = false;
  private taskIds: string[] = [];

  constructor(config: LocalExecutorConfig) {
    this.masterPath = config.masterPath;
    const host = config.masterHost || 'localhost';
    const port = config.masterPort || 6767;
    this.masterURL = `http://${host}:${port}`;
    this.shutdownOnComplete = config.shutdownOnComplete !== false;
    this.shutdownTimeout = config.shutdownTimeout || 30000; // 30 seconds default

    this.skill = new AgentFlowSkill({
      master_url: this.masterURL
    });
  }

  /**
   * Start Master server as a child process
   */
  async startMaster(): Promise<void> {
    console.log('🚀 Starting Master server...');

    return new Promise((resolve, reject) => {
      const args = [this.masterPath];
      const options = {
        stdio: 'inherit' as any,
        env: {
          ...process.env,
          NODE_ENV: 'development',
          AGENTFLOW_MASTER_URL: this.masterURL
        }
      };

      const proc = spawn('node', args, options);
      this.masterProcess = proc;

      this.masterProcess.on('error', (err) => {
        console.error('❌ Failed to start Master:', err);
        reject(err);
      });

      this.masterProcess.on('exit', (code) => {
        if (code !== 0 && code !== null) {
          console.log(`⚠️  Master process exited with code ${code}`);
        }
      });

      // Wait for Master to be ready
      setTimeout(async () => {
        try {
          const isHealthy = await this.skill.checkHealth();
          if (isHealthy) {
            console.log(`✓ Master server ready at ${this.masterURL}`);
            resolve();
          } else {
            reject(new Error('Master health check failed'));
          }
        } catch (err) {
          reject(err);
        }
      }, 3000); // Wait 3 seconds for startup
    });
  }

  /**
   * Start Worker
   */
  async startWorker(): Promise<void> {
    console.log('🤖 Starting Worker...');

    this.worker = new Worker({
      master_url: this.masterURL,
      group_name: 'local'
    } as any);

    await this.worker.start();
    console.log('✓ Worker started');
  }

  /**
   * Create tasks
   */
  async createTasks(tasks: TaskSpec[]): Promise<string[]> {
    console.log(`📝 Creating ${tasks.length} tasks...`);

    const taskIds: string[] = [];

    for (const task of tasks) {
      const taskId = await this.skill.createTask({
        title: task.title,
        description: task.description,
        group_name: task.group_name || 'local'
      });

      taskIds.push(taskId);
      console.log(`  ✓ Created: ${taskId} - ${task.title}`);
    }

    this.taskIds = taskIds;
    return taskIds;
  }

  /**
   * Monitor task completion
   */
  async monitorTasks(): Promise<void> {
    console.log('⏳ Monitoring task execution...');

    const checkInterval = 2000; // Check every 2 seconds
    const startTime = Date.now();
    const timeout = 300000; // 5 minute timeout

    while (this.running) {
      // Check timeout
      if (Date.now() - startTime > timeout) {
        console.log('⏰ Task monitoring timeout');
        break;
      }

      // Get task statuses
      const tasks = await Promise.all(
        this.taskIds.map(id => this.skill.getTaskStatus(id))
      );

      // Count completed/failed tasks
      const completed = tasks.filter(t => t.status === 'completed').length;
      const failed = tasks.filter(t => t.status === 'failed').length;
      const total = this.taskIds.length;

      // Log progress
      console.log(`  Progress: ${completed}/${total} completed, ${failed} failed`);

      // Check if all tasks are done
      const allDone = completed + failed >= total;

      if (allDone) {
        console.log('\n✓ All tasks completed!');
        this.running = false;
        break;
      }

      // Wait before next check
      await new Promise(resolve => setTimeout(resolve, checkInterval));
    }
  }

  /**
   * Stop Master server
   */
  async stopMaster(): Promise<void> {
    if (!this.masterProcess) {
      return;
    }

    console.log('🛑 Stopping Master server...');

    // Try graceful shutdown via API first
    try {
      await fetch(`${this.masterURL}/api/v1/shutdown`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      });

      // Wait for graceful shutdown
      await new Promise(resolve => setTimeout(resolve, 2000));
    } catch {
      // Ignore errors
    }

    // Kill process if still running
    if (this.masterProcess.pid) {
      this.masterProcess.kill('SIGTERM');

      // Wait a bit
      await new Promise(resolve => setTimeout(resolve, 2000));

      // Force kill if needed
      try {
        this.masterProcess.kill('SIGKILL');
      } catch {
        // Already dead
      }
    }

    console.log('✓ Master server stopped');
  }

  /**
   * Stop Worker
   */
  async stopWorker(): Promise<void> {
    if (!this.worker) {
      return;
    }

    console.log('🤖 Stopping Worker...');
    await this.worker.stop();
    console.log('✓ Worker stopped');
  }

  /**
   * Execute complete workflow
   */
  async execute(tasks: TaskSpec[]): Promise<void> {
    if (this.running) {
      throw new Error('Already running');
    }

    this.running = true;

    try {
      // Step 1: Start Master
      await this.startMaster();

      // Step 2: Start Worker
      await this.startWorker();

      // Step 3: Create tasks
      await this.createTasks(tasks);

      // Step 4: Monitor completion
      await this.monitorTasks();

      // Step 5: Shutdown
      if (this.shutdownOnComplete) {
        console.log('\n🏁 All tasks complete, initiating shutdown...');
        await this.stopWorker();
        await this.stopMaster();
        console.log('\n✅ Execution complete!');
      }

    } catch (error) {
      console.error('❌ Execution error:', error);
      throw error;
    } finally {
      this.running = false;

      // Ensure cleanup
      try {
        await this.stopWorker();
        await this.stopMaster();
      } catch {
        // Ignore cleanup errors
      }
    }
  }

  /**
   * Quick execute - single task
   */
  async executeOne(title: string, description: string): Promise<void> {
    return this.execute([{ title, description }]);
  }
}
