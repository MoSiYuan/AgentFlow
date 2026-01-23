# AgentFlow 统一架构设计文档

## 📋 文档说明

本文档定义 AgentFlow 跨语言（Python/Go/Node.js）的统一架构设计，确保三个版本的功能、API、流程完全一致。

---

## 🎯 设计原则

### 1. **API 优先**
- 所有语言版本提供完全一致的 RESTful API
- 客户端可以无缝切换不同语言的 Master/Worker

### 2. **功能对等**
- 核心功能在所有语言版本中完全一致
- 语言特定功能作为扩展提供

### 3. **数据库兼容**
- 所有版本使用相同的 SQLite 数据库 Schema
- 可以跨语言访问同一个数据库

### 4. **配置统一**
- 配置文件格式跨语言一致（YAML/JSON）
- 环境变量命名规范统一

---

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         AgentFlow 系统                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                      Client Layer                         │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  CLI Tool | Web Dashboard | VSCode Plugin | API Client  │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │ REST API                                │
│                       ▼                                         │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Master Server                         │  │
│  │  (Python / Go / Node.js)                                 │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  • Task Management    • Worker Coordination              │  │
│  │  • Status Tracking    • Event Broadcasting               │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                         │
│         ┌─────────────┼─────────────┐                         │
│         ▼             ▼             ▼                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                    │
│  │ Worker 1 │  │ Worker 2 │  │ Worker N │  ...               │
│  │(Python)  │  │   (Go)   │  │(Node.js) │                    │
│  └──────────┘  └──────────┘  └──────────┘                    │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Storage Layer                         │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  SQLite Database | Task Queue | File System | Git        │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📊 核心组件设计

### 1. Master Server

**职责**：
- 任务管理（创建、查询、更新状态）
- Worker 协调（注册、心跳、任务分配）
- 事件广播（任务状态变更、Worker 上下线）
- API 服务（RESTful + WebSocket）

**核心接口**：

```typescript
interface Master {
  // 任务管理
  createTask(task: Partial<Task>): Promise<Task>;
  getTask(id: string): Promise<Task>;
  listTasks(filter?: TaskFilter): Promise<Task[]>;
  updateTaskStatus(id: string, status: TaskStatus): Promise<void>;

  // Worker 管理
  registerWorker(worker: WorkerRegistration): Promise<Worker>;
  updateWorkerHeartbeat(id: string): Promise<void>;
  listWorkers(group?: string): Promise<Worker[]>;

  // 任务执行
  assignTask(taskId: string, workerId: string): Promise<boolean>;
  completeTask(taskId: string, result: TaskResult): Promise<void>;
  failTask(taskId: string, error: Error): Promise<void>;

  // 服务器控制
  start(): Promise<void>;
  stop(): Promise<void>;
}
```

**统一 API 端点**：

```
# 健康检查
GET /health
GET /api/status

# 任务管理
POST   /api/v1/tasks
GET    /api/v1/tasks
GET    /api/v1/tasks/:id
POST   /api/v1/tasks/:id/complete
POST   /api/v1/tasks/:id/fail
GET    /api/v1/tasks/pending

# Worker 管理
GET    /api/v1/workers
POST   /api/v1/workers/:id/heartbeat
GET    /api/v1/workers/available

# 任务执行
POST   /api/v1/tasks/:id/assign
POST   /api/v1/tasks/:id/lock
POST   /api/v1/tasks/:id/unlock

# 统计信息
GET    /api/v1/stats
GET    /api/v1/stats/groups
```

**语言特定优化**：

| 特性 | Python | Go | Node.js |
|------|--------|-----|---------|
| HTTP 框架 | Flask | Gin | Express/Fastify |
| 并发模型 | Threading | Goroutines | Event Loop |
| WebSocket | Flask-SocketIO | gorilla/websocket | ws |
| 数据库 | sqlite3 | mattn/go-sqlite3 | better-sqlite3 |

---

### 2. Worker

**职责**：
- 向 Master 注册并保持心跳
- 从 Master 拉取待执行任务
- 执行任务（Claude CLI / Shell / HTTP）
- 上报任务进度和结果

**核心接口**：

```typescript
interface Worker {
  // Worker 控制
  start(): Promise<void>;
  stop(): Promise<void>;

  // 任务执行
  fetchTask(): Promise<Task | null>;
  executeTask(task: Task): Promise<TaskResult>;
  reportProgress(taskId: string, progress: number, message: string): Promise<void>;

  // 生命周期
  register(): Promise<void>;
  sendHeartbeat(): Promise<void>;
}
```

**执行器优先级**（所有语言统一）：

```
1. HTTP Executor (Claude Server)
   ↓ (不可用)
2. Claude CLI Executor
   ↓ (未安装)
3. Shell Command Executor
```

**语言特定能力**：

| 能力 | Python | Go | Node.js |
|------|--------|-----|---------|
| GUI 自动化 | ✅ 原生 | ❌ | ⚠️ 通过系统命令 |
| 系统调用 | ✅ os | ✅ syscall | ✅ child_process |
| 跨平台编译 | ❌ | ✅ | ✅ pkg/nexe |
| 实时通信 | ⚠️ | ⚠️ | ✅ 原生 WebSocket |

---

### 3. Database Layer

**统一 Schema**：

```sql
-- 任务表
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    group_name TEXT NOT NULL DEFAULT 'default',
    completion_criteria TEXT,
    status TEXT NOT NULL DEFAULT 'pending'
           CHECK(status IN ('pending', 'running', 'completed', 'failed', 'blocked')),
    priority INTEGER DEFAULT 0,
    lock_holder TEXT,
    lock_time DATETIME,
    result TEXT,
    error TEXT,
    workspace_dir TEXT,
    sandboxed INTEGER DEFAULT 0,
    allow_network INTEGER DEFAULT 1,
    max_memory TEXT DEFAULT '512M',
    max_cpu INTEGER DEFAULT 1,
    created_by TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Worker 表
CREATE TABLE workers (
    id TEXT PRIMARY KEY,
    group_name TEXT NOT NULL DEFAULT 'default',
    type TEXT NOT NULL CHECK(type IN ('local', 'remote')),
    capabilities TEXT,
    status TEXT NOT NULL DEFAULT 'active'
           CHECK(status IN ('active', 'inactive')),
    last_heartbeat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 任务日志表
CREATE TABLE task_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    worker_id TEXT NOT NULL,
    log_level TEXT NOT NULL CHECK(log_level IN ('info', 'warning', 'error')),
    message TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Git 集成表
CREATE TABLE git_tasks (
    id TEXT PRIMARY KEY,
    title TEXT,
    description TEXT,
    agent_id TEXT,
    git_branch TEXT NOT NULL UNIQUE,
    file_boundaries TEXT,
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    assigned_at DATETIME,
    completed_at DATETIME
);

CREATE TABLE git_locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    lock_type TEXT NOT NULL,
    acquired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    released_at DATETIME,
    status TEXT DEFAULT 'active',
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);

CREATE TABLE git_conflicts (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    conflict_type TEXT NOT NULL,
    file_paths TEXT,
    description TEXT,
    severity TEXT DEFAULT 'medium',
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    resolved_by TEXT,
    resolution TEXT,
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);
```

**数据库操作接口**：

```typescript
interface Database {
  // 任务操作
  createTask(task: Partial<Task>): Promise<string>;
  getTask(id: string): Promise<Task>;
  listTasks(status?: string, group?: string): Promise<Task[]>;
  updateTaskStatus(id: string, status: TaskStatus, workerId?: string): Promise<boolean>;
  completeTask(id: string, workerId: string, result: string): Promise<boolean>;
  failTask(id: string, workerId: string, error: string): Promise<boolean>;

  // 任务锁定
  lockTask(id: string, workerId: string): Promise<boolean>;
  unlockTask(id: string, workerId: string): Promise<boolean>;

  // Worker 操作
  registerWorker(worker: Worker): Promise<boolean>;
  updateWorkerHeartbeat(id: string): Promise<boolean>;
  listWorkers(group?: string): Promise<Worker[]>;
  cleanupOfflineWorkers(timeout?: number): Promise<number>;

  // 统计信息
  getStats(): Promise<SystemStats>;
  getGroupStats(): Promise<GroupStats[]>;

  // 连接管理
  connect(): Promise<void>;
  close(): Promise<void>;
}
```

---

## 🔄 任务执行流程

### 1. 标准任务流程

```
┌─────────────────────────────────────────────────────────────────┐
│                     任务执行生命周期                             │
└─────────────────────────────────────────────────────────────────┘

1. 任务创建（Client → Master）
   ├─ POST /api/v1/tasks
   ├─ { title, description, group_name, priority }
   └─ → Task { id: "TASK-XXX", status: "pending" }

2. Worker 注册（Worker → Master）
   ├─ POST /api/v1/workers/register
   ├─ { worker_id, group_name, capabilities }
   └─ → Worker registered

3. 任务拉取（Worker → Master）
   ├─ GET /api/v1/tasks/pending?group=default
   ├─ 返回待执行任务列表
   └─ → [Task1, Task2, ...]

4. 任务锁定（Worker → Master）
   ├─ POST /api/v1/tasks/:id/lock
   ├─ { worker_id: "worker-1" }
   └─ → { status: "locked" } 或 { status: "already_locked" }

5. 任务执行（Worker 本地）
   ├─ 尝试优先级 1: HTTP Executor (Claude Server)
   ├─ 尝试优先级 2: Claude CLI Executor
   ├─ 尝试优先级 3: Shell Command Executor
   └─ → TaskResult { success, result, error }

6. 进度上报（Worker → Master，可选）
   ├─ POST /api/v1/tasks/:id/progress
   ├─ { progress: 50, message: "Processing..." }
   └─ → Progress updated

7. 任务完成（Worker → Master）
   ├─ POST /api/v1/tasks/:id/complete
   ├─ { worker_id, result }
   └─ → Task status: "completed"

   或任务失败
   ├─ POST /api/v1/tasks/:id/fail
   ├─ { worker_id, error }
   └─ → Task status: "failed"

8. 任务解锁（自动）
   └─ Task lock released
```

### 2. Git 集成任务流程

```
┌─────────────────────────────────────────────────────────────────┐
│                  Git 集成任务执行流程                            │
└─────────────────────────────────────────────────────────────────┘

1. 创建 Git 任务
   ├─ 创建 Git 分支: agent-{agent_id}/task-{task_id}
   ├─ 记录到 git_tasks 表
   └─ → Task with git_branch

2. 验证文件访问权限
   ├─ 检查 .agentflow/boundaries.json
   ├─ 验证 agent_id 对文件是否有访问权限
   └─ → Allowed / Denied

3. 获取文件锁（如需要）
   ├─ 检查文件是否被锁定
   ├─ 创建锁记录到 git_locks 表
   └─ → Lock acquired / Conflict

4. Agent 在分支上工作
   ├─ 所有提交在独立分支
   ├─ 其他 Agent 不可见
   └─ → Commits on branch

5. 尝试合并到主分支
   ├─ git merge --squash agent-{id}/task-{id}
   ├─ 检测冲突
   └─ → Success / Conflict

6. 冲突处理（如有）
   ├─ 级别 1: 自动重试（最多 3 次）
   ├─ 级别 2: 创建升级任务
   │   ├─ 新任务: UPGRADE-{task_id}
   │   ├─ 包含冲突信息
   │   └─ → Auto resolve
   └─ 级别 3: 人工介入
       ├─ 标记为 needs_manual_resolution
       └─ → Human intervention

7. 清理
   ├─ 删除分支
   ├─ 释放文件锁
   └─ → Task completed
```

---

## 🔧 配置管理

### 统一配置文件格式

**Master 配置** (`master.config.yaml`):

```yaml
# Master 服务器配置
master:
  host: "0.0.0.0"
  port: 8848
  db_path: ".claude/cpds-manager/agentflow.db"
  auto_shutdown: false
  log_level: "info"

# Worker 自动启动
auto_start_workers: true
worker_groups:
  - "default"
  - "darwin"
  - "linux"
  - "windows"

# Git 集成
git:
  enabled: true
  boundary_config: ".agentflow/boundaries.json"
  auto_merge: true
  merge_strategy: "squash"  # squash | merge | rebase

# 任务队列
queue:
  type: "memory"  # memory | redis
  redis_url: "redis://localhost:6379"

# WebSocket
websocket:
  enabled: true
  port: 8849
```

**Worker 配置** (`worker.config.yaml`):

```yaml
# Worker 基础配置
worker:
  id: ""  # 留空自动生成
  master_url: "http://localhost:8848"
  group_name: "default"
  mode: "auto"  # auto | manual | oneshot

# 心跳配置
heartbeat:
  interval: 30  # 秒
  timeout: 120  # 秒，超时判定为离线

# 任务执行
execution:
  max_concurrent: 1
  timeout: 600  # 秒
  retry_on_failure: false
  max_retries: 3

# 执行器优先级
executors:
  - type: "http"        # HTTP Executor (Claude Server)
    enabled: true
    url: "http://localhost:8849"

  - type: "claude_cli"  # Claude CLI
    enabled: true
    command: "claude"

  - type: "shell"       # Shell 命令
    enabled: true
    shell: "/bin/bash"  # Unix, Windows 用 powershell

# Claude 配置
claude:
  model: ""
  max_tokens: 4096
  temperature: 0.7
  timeout: 120

# 工作空间
workspace:
  base_dir: ".agentflow/workspace"
  cleanup_on_completion: false
  sandbox: false
```

### 环境变量规范

```bash
# Master
AGENTFLOW_MASTER_HOST=0.0.0.0
AGENTFLOW_MASTER_PORT=8848
AGENTFLOW_DB_PATH=.claude/cpds-manager/agentflow.db
AGENTFLOW_LOG_LEVEL=info

# Worker
AGENTFLOW_WORKER_ID=worker-1
AGENTFLOW_MASTER_URL=http://localhost:8848
AGENTFLOW_GROUP_NAME=default
AGENTFLOW_MODE=auto

# Git
AGENTFLOW_GIT_ENABLED=true
AGENTFLOW_GIT_BOUNDARY_CONFIG=.agentflow/boundaries.json

# Claude
ANTHROPIC_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-sonnet-4-20250514
```

---

## 📡 事件系统

### 事件类型

```typescript
type EventType =
  | 'task.created'      // 任务创建
  | 'task.assigned'     // 任务分配
  | 'task.started'      // 任务开始
  | 'task.progress'     // 任务进度更新
  | 'task.completed'    // 任务完成
  | 'task.failed'       // 任务失败
  | 'worker.registered' // Worker 注册
  | 'worker.online'     // Worker 上线
  | 'worker.offline'    // Worker 离线
  | 'git.conflict'      // Git 冲突
  | 'git.merged';       // Git 合并

interface Event {
  type: EventType;
  data: any;
  timestamp: Date;
}
```

### WebSocket 事件广播

```typescript
// Master 端
const eventBus = new EventEmitter();

// 监听数据库变化并广播事件
database.on('task_created', (task) => {
  eventBus.emit('task.created', task);
  websocket.broadcast({
    type: 'task.created',
    data: task,
    timestamp: new Date()
  });
});

// Worker/Client 端
const ws = new WebSocket('ws://localhost:8849');
ws.on('message', (data) => {
  const event = JSON.parse(data);
  if (event.type === 'task.created') {
    console.log('New task:', event.data);
  }
});
```

---

## 🧪 测试策略

### 跨语言 API 兼容性测试

```yaml
test_scenarios:
  - name: "Create and execute task"
    steps:
      - POST /api/v1/tasks (create)
      - GET /api/v1/tasks/:id (query)
      - POST /api/v1/tasks/:id/assign (assign)
      - POST /api/v1/tasks/:id/complete (complete)
      - GET /api/v1/tasks/:id (verify status)

  - name: "Worker registration and heartbeat"
    steps:
      - POST /api/v1/workers/register
      - POST /api/v1/workers/:id/heartbeat
      - GET /api/v1/workers (verify)

  - name: "Git integration"
    steps:
      - POST /api/v1/git/tasks (create git task)
      - Verify branch created
      - Lock file
      - Modify file
      - Attempt merge
      - Verify conflict detection
```

---

## 📚 文档规范

### API 文档
- 使用 OpenAPI 3.0 规范
- 所有语言版本 API 完全一致
- 提供 examples 目录

### 配置文档
- 每个配置项都有说明
- 提供默认值和可选值
- 包含环境变量对应关系

### 部署文档
- 三平台部署脚本
- Docker 镜像
- K8s manifests

---

## 🎯 版本兼容性

### API 版本策略
```
/api/v1  - 当前稳定版本
/api/v2  - 未来版本（不破坏兼容性）
```

### 数据库版本管理
- 使用 migrations 管理数据库变更
- 所有语言版本使用相同 migration 文件
- 版本号：`v1.0.0`, `v1.1.0`, etc.

---

**版本**: v1.0.0
**更新**: 2026-01-22
**作者**: AgentFlow Team
