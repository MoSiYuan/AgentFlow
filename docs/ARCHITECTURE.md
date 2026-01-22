# AgentFlow 架构设计文档

## 系统概述

AgentFlow (Claude Parallel Development System) 是一个基于 Master-Worker 架构的分布式任务执行系统，专为多 AI Agent 协作开发设计。

## 核心架构

```
┌─────────────────────────────────────────────────────────────┐
│                         Master                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ HTTP Server │  │   SQLite    │  │  Scheduler  │        │
│  │   (Gin)     │  │  Database   │  │   Engine    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│         │                                    │              │
└─────────┼────────────────────────────────────┼──────────────┘
          │                                    │
          │ HTTP API                           │ Process
          │                                    │
    ┌─────▼─────┐                      ┌──────▼──────┐
    │   Local   │                      │    Local    │
    │  Worker   │                      │   Worker    │
    │ (Subprocess)                      │  (Subprocess)│
    └───────────┘                      └─────────────┘

┌─────────────────────────────────────────────────────────────┐
│                       Remote Workers                        │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐      │
│  │ Linux   │  │ Windows │  │ Docker  │  │   K8s   │      │
│  │ Worker  │  │ Worker  │  │ Worker  │  │ Worker  │      │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. Master

Master 是系统的核心节点，负责：

- **任务管理**: 接收、存储、分发任务
- **Worker 协调**: 跟踪 Worker 状态，分配任务
- **锁机制**: 确保任务原子性执行
- **HTTP API**: 提供任务查询和管理接口
- **本地 Worker 管理**: 自动启动和管理本地 Worker 进程

#### Master 数据流

```
Client Request
    │
    ├─> POST /api/v1/tasks ──> 创建任务
    │                              │
    │                              └─> 存入 SQLite (status=pending)
    │
    ├─> GET /api/v1/tasks ──> 查询任务
    │                              │
    │                              └─> 从 SQLite 读取
    │
    └─> Worker Polling ──> 获取待执行任务
                                   │
                                   ├─> 锁定任务 (UPDATE ... WHERE status='pending')
                                   │
                                   └─> 返回任务详情
```

### 2. SQLite 数据库

单文件数据库，存储所有状态：

```sql
tasks (
    id INTEGER PRIMARY KEY,
    title TEXT,
    description TEXT,
    group_name TEXT,
    status TEXT CHECK(status IN ('pending', 'running', 'completed', 'failed')),
    lock_holder TEXT,
    lock_time DATETIME,
    result TEXT,
    error TEXT,
    created_at DATETIME,
    started_at DATETIME,
    completed_at DATETIME
)

workers (
    id TEXT PRIMARY KEY,
    group_name TEXT,
    type TEXT CHECK(type IN ('local', 'remote')),
    capabilities TEXT,
    status TEXT CHECK(status IN ('active', 'inactive')),
    last_heartbeat DATETIME,
    created_at DATETIME
)

task_logs (
    id INTEGER PRIMARY KEY,
    task_id INTEGER,
    worker_id TEXT,
    log_level TEXT,
    message TEXT,
    created_at DATETIME,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
)
```

#### 任务锁机制

```sql
-- 原子级锁获取
UPDATE tasks
SET status = 'running',
    lock_holder = ?,
    lock_time = CURRENT_TIMESTAMP
WHERE id = ?
  AND status = 'pending'
  AND (lock_holder IS NULL 
       OR lock_holder = ? 
       OR lock_time < datetime('now', '-5 minutes'))
```

### 3. Worker

Worker 负责执行任务，分为两种类型：

#### Local Worker（本地 Worker）

- 作为 Master 的子进程运行
- 直接访问 SQLite 数据库
- 无网络开销，性能最高
- Master 自动启动和管理

```
Master Process
    │
    ├─> fork/exec ──> Local Worker 1 (linux group)
    │                      │
    │                      └─> 直接读取 SQLite
    │
    ├─> fork/exec ──> Local Worker 2 (docker group)
    │                      │
    │                      └─> 直接读取 SQLite
    │
    └─> fork/exec ──> Local Worker 3 (windows group)
                           │
                           └─> 直接读取 SQLite
```

#### Remote Worker（远程 Worker）

- 独立进程，通过网络连接 Master
- 通过 HTTP API 获取和更新任务
- 适合云端分布式部署
- 支持多种工作组

```
Remote Worker Loop:
    1. GET /api/v1/tasks/pending?group=docker
    2. POST /api/v1/tasks/:id/lock {worker_id}
    3. 执行任务 (shell command / script)
    4. POST /api/v1/tasks/:id/complete {worker_id, result}
       or
       POST /api/v1/tasks/:id/fail {worker_id, error}
```

### 4. Worker Groups

Worker 按环境/能力分组，实现任务智能分发：

```go
// 预定义组
const (
    GroupDefault = "default"
    GroupLinux   = "linux"
    GroupWindows = "windows"
    GroupDarwin  = "darwin"
    GroupDocker  = "docker"
    GroupK8s     = "k8s"
)

// 自动检测
func detectGroup() string {
    if inDocker() { return "docker" }
    if inKubernetes() { return "k8s" }
    return runtime.GOOS // linux, windows, darwin
}
```

## 执行流程

### 单机模式

```
1. Master 启动
   │
   ├─> 初始化 SQLite 数据库
   │
   ├─> 检测本地环境 (linux/docker/etc)
   │
   └─> 启动本地 Workers

2. Agent 创建任务
   │
   └─> POST /api/v1/tasks
       │
       └─> 存入 SQLite (status=pending, group_name=docker)

3. Local Worker 轮询
   │
   ├─> 查询 SQLite: SELECT * FROM tasks WHERE status='pending' AND group_name='docker'
   │
   ├─> 锁定任务: UPDATE tasks SET status='running' WHERE id=?
   │
   ├─> 执行命令: docker build -t app .
   │
   └─> 完成任务: UPDATE tasks SET status='completed', result=...
```

### 分布式模式

```
1. Master Server 启动 (云端)
   │
   └─> 监听 http://0.0.0.0:8848

2. Workers 连接 (不同机器)
   │
   ├─> Linux Server: worker --master=http://master:8848 --group=linux
   ├─> Windows Server: worker --master=http://master:8848 --group=windows
   └─> Docker Container: worker --master=http://master:8848 --group=docker

3. Agent 创建任务
   │
   └─> POST /api/v1/tasks {group_name: "linux"}

4. Linux Worker 轮询
   │
   ├─> GET /api/v1/tasks/pending?group=linux
   │
   ├─> POST /api/v1/tasks/1/lock
   │
   ├─> 执行任务
   │
   └─> POST /api/v1/tasks/1/complete
```

## 并发控制

### SQLite 并发

- **WAL 模式**: 允许并发读
- **锁超时**: 5 分钟自动释放
- **原子操作**: 使用 SQL WHERE 条件实现乐观锁

```go
// 乐观锁示例
locked, err := db.Exec(`
    UPDATE tasks
    SET status = 'running', lock_holder = ?
    WHERE id = ? AND status = 'pending'
`, workerID, taskID)

affected, _ := db.RowsAffected()
if affected == 0 {
    // 锁定失败，任务已被其他 worker 获取
    return
}
```

### Worker 并发

- 每个 Worker 独立运行
- 同时只能执行一个任务
- 5 秒轮询间隔
- 30 秒心跳间隔

## API 设计

### RESTful API

```
POST   /api/v1/tasks              创建任务
GET    /api/v1/tasks              列出任务
GET    /api/v1/tasks/:id          获取任务详情
POST   /api/v1/tasks/:id/lock     锁定任务
POST   /api/v1/tasks/:id/unlock   解锁任务
POST   /api/v1/tasks/:id/complete 完成任务
POST   /api/v1/tasks/:id/fail     标记任务失败

GET    /api/v1/workers            列出 workers
POST   /api/v1/workers/:id/heartbeat 发送心跳

GET    /api/v1/tasks/pending      获取待执行任务 (Worker 专用)
GET    /api/v1/stats              获取统计信息
GET    /health                    健康检查
```

### 响应格式

```json
{
  "tasks": [
    {
      "id": "1",
      "title": "运行测试",
      "description": "shell:go test ./...",
      "group_name": "linux",
      "status": "pending",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

## 性能优化

### SQLite 优化

```sql
PRAGMA journal_mode = WAL;        -- 写前日志，提高并发
PRAGMA synchronous = NORMAL;      -- 平衡性能和安全
PRAGMA cache_size = -64000;       -- 64MB 缓存
PRAGMA temp_store = MEMORY;       -- 临时表在内存
PRAGMA mmap_size = 30000000000;   -- 内存映射 I/O
PRAGMA page_size = 4096;          -- 页大小优化
```

### 索引优化

```sql
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_group ON tasks(group_name);
CREATE INDEX idx_tasks_lock ON tasks(lock_holder, lock_time);
CREATE INDEX idx_workers_group ON workers(group_name);
```

### 查询优化

```sql
-- 使用视图简化查询
CREATE VIEW v_pending_tasks AS
SELECT id, title, description, group_name, created_at
FROM tasks
WHERE status = 'pending'
ORDER BY created_at ASC;
```

## 错误处理

### 任务失败处理

```go
func (w *Worker) executeTask(task *Task) (result string, err error) {
    // 执行命令
    output, err := runCommand(task.Description)
    if err != nil {
        // 标记任务失败
        w.db.FailTask(task.ID, w.ID, err.Error())
        return "", err
    }
    
    // 标记任务完成
    w.db.CompleteTask(task.ID, w.ID, output)
    return output, nil
}
```

### Worker 故障恢复

- **心跳超时**: 2 分钟无心跳标记为 inactive
- **任务锁超时**: 5 分钟自动释放
- **自动重试**: Worker 可重启并继续执行

## 扩展性

### 水平扩展

- 添加更多 Worker 节点
- Master 无状态，可部署多个实例（需负载均衡）
- SQLite 可迁移到 PostgreSQL/MySQL（需适配层）

### 垂直扩展

- 增加 SQLite cache_size
- 调整 Worker 并发数
- 优化任务执行脚本

## 安全考虑

### 当前实现

- 无身份认证
- 无加密通信
- 本地网络信任

### 生产环境建议

- 添加 TLS/HTTPS
- 实现身份认证（JWT/API Key）
- 网络隔离（VPN/VPC）
- 审计日志

## 监控和可观测性

### 日志

```go
logger.WithFields(logrus.Fields{
    "task_id": taskID,
    "worker_id": workerID,
    "group": groupName,
}).Info("任务已完成")
```

### 指标

- 任务吞吐量
- 任务延迟
- Worker 数量
- 失败率

### 健康检查

```bash
$ curl http://localhost:8848/health
{"status":"ok"}
```

## 未来改进

- [ ] 支持 PostgreSQL/MySQL
- [ ] 任务优先级队列
- [ ] 任务依赖关系
- [ ] 定时任务
- [ ] Web UI 界面
- [ ] Prometheus metrics
- [ ] 分布式追踪 (OpenTelemetry)
- [ ] 任务执行日志流
