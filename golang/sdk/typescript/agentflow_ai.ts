#!/usr/bin/env tsx
/**
 * CPDS AI SDK - TypeScript SDK for Claude Code
 *
 * 专为 Claude Code 设计的 CPDS SDK
 */

import { execSync, spawn } from "child_process";
import { promisify } from "util";
import * as fs from "fs";
import * as path from "path";

const exec = promisify(execSync);

/**
 * 任务状态枚举
 */
export enum TaskStatus {
  PENDING = "pending",
  RUNNING = "running",
  COMPLETED = "completed",
  FAILED = "failed",
}

/**
 * 任务类型枚举
 */
export enum TaskType {
  AI = "task",       // AI 智能任务
  SHELL = "shell",   // Shell 命令
  SCRIPT = "script", // 脚本文件
  FILE = "file",     // 文件操作
}

/**
 * 任务接口
 */
export interface Task {
  id: string;
  title: string;
  description: string;
  group: string;
  status: TaskStatus;
  result?: string;
  error?: string;
}

/**
 * CPDS 客户端配置
 */
export interface CPDSConfig {
  dbPath?: string;
  masterUrl?: string;
  workerGroup?: string;
}

/**
 * CPDS 客户端类
 */
export class CPDSClient {
  private dbPath: string;
  private masterUrl: string;
  private workerGroup: string;

  constructor(config: CPDSConfig = {}) {
    this.dbPath = config.dbPath || process.env.CPDS_DB_PATH || "agentflow.db";
    this.masterUrl =
      config.masterUrl || process.env.MASTER_URL || "http://localhost:6767";
    this.workerGroup =
      config.workerGroup || process.env.WORKER_GROUP || "local";
  }

  /**
   * 创建任务
   */
  async createTask(
    title: string,
    description: string,
    options: {
      group?: string;
      wait?: boolean;
      timeout?: number;
    } = {}
  ): Promise<Task> {
    const group = options.group || this.workerGroup;
    const cmd = [
      "agentflow",
      "add",
      title,
      "--desc",
      description,
      "--group",
      group,
    ];

    const result = exec(cmd.join(" "), { encoding: "utf-8" });
    const taskId = this.parseTaskId(result);

    const task: Task = {
      id: taskId,
      title,
      description,
      group,
      status: TaskStatus.PENDING,
    };

    if (options.wait) {
      return this.waitForTask(taskId, options.timeout || 300);
    }

    return task;
  }

  /**
   * 等待任务完成
   */
  async waitForTask(
    taskId: string,
    timeout: number = 300,
    checkInterval: number = 2000
  ): Promise<Task> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeout * 1000) {
      const task = await this.getTask(taskId);

      if (
        task.status === TaskStatus.COMPLETED ||
        task.status === TaskStatus.FAILED
      ) {
        return task;
      }

      await this.sleep(checkInterval);
    }

    throw new Error(`Task ${taskId} timeout after ${timeout}s`);
  }

  /**
   * 获取任务详情
   */
  async getTask(taskId: string): Promise<Task> {
    const result = exec("agentflow list", { encoding: "utf-8" });
    return this.parseTaskFromList(result, taskId);
  }

  /**
   * 列出任务
   */
  async listTasks(options: {
    status?: TaskStatus;
    group?: string;
  } = {}): Promise<Task[]> {
    let cmd = ["agentflow", "list"];

    if (options.status) {
      cmd.push("--status", options.status);
    }
    if (options.group) {
      cmd.push("--group", options.group);
    }

    const result = exec(cmd.join(" "), { encoding: "utf-8" });
    return this.parseTaskList(result);
  }

  /**
   * 创建 AI 任务（自动分解）
   */
  async createAITask(
    feature: string,
    options: { wait?: boolean; timeout?: number } = {}
  ): Promise<Task> {
    return this.createTask(`实现${feature}`, `task:implement:${feature}`, {
      ...options,
      wait: options.wait !== false, // 默认等待
    });
  }

  /**
   * 创建 Shell 任务
   */
  async createShellTask(
    title: string,
    command: string,
    options: { group?: string; wait?: boolean } = {}
  ): Promise<Task> {
    return this.createTask(title, `shell:${command}`, options);
  }

  /**
   * 创建测试任务
   */
  async createTestTask(options: { wait?: boolean } = {}): Promise<Task> {
    return this.createTask("运行测试", "task:test", {
      ...options,
      wait: options.wait !== false,
    });
  }

  /**
   * 创建构建任务
   */
  async createBuildTask(options: { wait?: boolean } = {}): Promise<Task> {
    return this.createShellTask("构建项目", "go build -v ./...", {
      ...options,
      wait: options.wait !== false,
    });
  }

  /**
   * 解析任务 ID
   */
  private parseTaskId(output: string): string {
    const lines = output.split("\n");
    for (const line of lines) {
      if (line.includes("ID:")) {
        const match = line.match(/ID:\s*(\d+)/);
        if (match) {
          return match[1];
        }
      }
    }
    throw new Error("Cannot parse task ID from output");
  }

  /**
   * 从列表输出解析任务
   */
  private parseTaskFromList(output: string, taskId: string): Task {
    const tasks = this.parseTaskList(output);
    const task = tasks.find((t) => t.id === taskId);
    if (!task) {
      throw new Error(`Task ${taskId} not found`);
    }
    return task;
  }

  /**
   * 解析任务列表
   */
  private parseTaskList(output: string): Task[] {
    const tasks: Task[] = [];
    const lines = output.split("\n");

    for (const line of lines) {
      if (line.includes("[")) {
        try {
          const task = this.parseTaskLine(line);
          if (task) {
            tasks.push(task);
          }
        } catch (e) {
          // 忽略解析失败的行
        }
      }
    }

    return tasks;
  }

  /**
   * 解析单行任务
   */
  private parseTaskLine(line: string): Task | null {
    const match = line.match(/\[(\d+)\]/);
    if (!match) {
      return null;
    }

    const id = match[1];
    let status = TaskStatus.PENDING;

    if (line.includes("⏳")) {
      status = TaskStatus.PENDING;
    } else if (line.includes("▶️")) {
      status = TaskStatus.RUNNING;
    } else if (line.includes("✅")) {
      status = TaskStatus.COMPLETED;
    } else if (line.includes("❌")) {
      status = TaskStatus.FAILED;
    }

    // 提取标题（简化版）
    const parts = line.split("]");
    const title = parts.length > 2 ? parts[2].trim() : "Unknown";

    return {
      id,
      title,
      description: "",
      group: this.workerGroup,
      status,
    };
  }

  /**
   * 延时函数
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

/**
 * 快捷函数：创建并等待任务
 */
export async function quickTask(
  title: string,
  description: string,
  wait: boolean = true
): Promise<Task> {
  const client = new CPDSClient();
  return client.createTask(title, description, { wait });
}

/**
 * 快捷函数：开发功能
 */
export async function develop(
  feature: string,
  wait: boolean = true
): Promise<Task> {
  const client = new CPDSClient();
  return client.createAITask(feature, { wait });
}

// === 使用示例 ===

async function main() {
  // 示例 1：创建简单任务
  console.log("示例 1：创建简单任务");
  const task1 = await quickTask("系统信息", "shell:uname -a", true);
  console.log(`任务完成: ${task1.id}`);

  // 示例 2：开发功能
  console.log("\n示例 2：开发功能");
  const task2 = await develop("用户登录功能", true);
  console.log(`开发完成: ${task2.id}`);

  // 示例 3：批量创建任务
  console.log("\n示例 3：批量创建任务");
  const client = new CPDSClient();

  const tasks = await Promise.all([
    client.createShellTask("格式化代码", "gofmt -w ."),
    client.createShellTask("运行测试", "go test ./..."),
    client.createShellTask("构建应用", "go build -v ./..."),
  ]);

  // 等待所有任务完成
  for (const task of tasks) {
    await client.waitForTask(task.id);
    console.log(`✓ ${task.title} 完成`);
  }

  console.log("\n所有任务完成！");
}

if (require.main === module) {
  main().catch(console.error);
}
