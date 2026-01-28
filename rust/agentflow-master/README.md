# AgentFlow Master - Team C 实现文档

## 概述

AgentFlow Master 是单进程架构的任务调度服务器，集成了 TaskExecutor 和 MemoryCore，直接执行任务而无需远程 Worker。

## 架构设计

```
┌─────────────────────────────────────────────────────┐
│              AgentFlow Master Server                │
│                                                       │
│  ┌─────────────┐      ┌──────────────┐             │
│  │ HTTP API    │◄────►│ TaskExecutor │             │
│  │             │      │              │             │
│  │ - POST /api │      │ - 执行任务   │             │
│  │ - GET /api  │      │ - 状态管理   │             │
│  │ - WebSocket │      │ - 并发控制   │             │
│  └─────────────┘      └──────────────┘             │
│         │                      │                    │
│         └──────────────────────┼──────────────────┐ │
│                                ▼                  │ │
│                        ┌──────────────┐          │ │
│                        │ MemoryCore   │          │ │
│                        │              │          │ │
│                        │ - 键值存储   │          │ │
│                        │ - 过期管理   │          │ │
│                        │ - 检索功能   │          │ │
│                        └──────────────┘          │ │
│                                ▲                  │ │
└────────────────────────────────┼──────────────────┘ │
                                 │                    │
                          ┌──────┴──────┐           │
                          │   SQLite    │           │
                          │  Database   │           │
                          └─────────────┘           │
                                                   │
└───────────────────────────────────────────────────┘
```

## 文件结构

```
agentflow-master/src/
├── main.rs           # 主程序入口，服务器启动逻辑
├── config.rs         # 配置管理（环境变量、配置文件）
├── error.rs          # 错误类型定义
├── executor.rs       # 任务执行器（集成 Team A）
├── memory_core.rs    # 记忆核心（集成 Team B）
└── routes/           # API 路由模块
    ├── mod.rs        # 路由模块导出
    ├── tasks.rs      # 任务管理 API
    ├── memory.rs     # 记忆管理 API
    ├── websocket.rs  # WebSocket 处理
    └── health.rs     # 健康检查 API
```

## 核心功能

### 1. 配置管理 (config.rs)

#### MasterConfig
- **服务器配置**：地址、端口、日志级别
- **数据库配置**：SQLite 连接 URL
- **Sandbox 配置**：工作目录、网络权限、资源限制
- **Memory 配置**：存储后端、TTL、最大条目数
- **Worker 配置**：心跳超时、任务超时、并发数

#### 环境变量
```bash
AGENTFLOW_SERVER_ADDR=0.0.0.0
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_DATABASE_URL=sqlite://agentflow.db
AGENTFLOW_SANDBOX_ENABLED=true
AGENTFLOW_SANDBOX_WORKSPACE=/tmp/agentflow/workspace
AGENTFLOW_MAX_CONCURRENT_TASKS=10
```

### 2. 任务执行器 (executor.rs)

#### TaskExecutor
```rust
pub struct TaskExecutor {
    db: Pool<Sqlite>,                    // 数据库连接池
    running_tasks: Arc<RwLock<Vec<i64>>>, // 运行中的任务列表
    max_concurrent_tasks: usize,         // 最大并发数
}
```

**主要方法：**
- `execute_task(task_id)` - 执行任务，返回结果
- `get_running_tasks()` - 获取运行中的任务列表
- `is_task_running(task_id)` - 检查任务是否在运行

**执行流程：**
1. 检查并发限制
2. 从数据库获取任务信息
3. 标记任务为 Running
4. 执行任务逻辑（模拟 AI Agent 调用）
5. 标记任务为 Completed/Failed
6. 返回执行结果

### 3. 记忆核心 (memory_core.rs)

#### MemoryCore
```rust
pub struct MemoryCore {
    storage: Arc<RwLock<HashMap<String, (MemoryEntry, DateTime<Utc>)>>>,
    default_ttl: i64,
    max_entries: usize,
}
```

**主要方法：**
- `set(key, value, category, task_id, ttl)` - 设置记忆条目
- `get(key)` - 获取记忆条目
- `delete(key)` - 删除记忆条目
- `search(query, category, task_id, limit)` - 搜索记忆
- `cleanup_expired()` - 清理过期条目
- `stats()` - 获取统计信息

**记忆分类：**
- `Execution` - 任务执行记录
- `Context` - 上下文信息
- `Result` - 执行结果
- `Error` - 错误信息
- `Checkpoint` - 检查点

### 4. API 路由

#### 任务管理 API

**创建任务**
```
POST /api/v1/tasks
Content-Type: application/json

{
  "title": "分析代码库",
  "description": "分析项目结构和依赖",
  "group_name": "code-analysis",
  "priority": "high",
  "sandboxed": true,
  "allow_network": false
}

Response:
{
  "success": true,
  "data": {
    "id": 1,
    "task_id": "uuid",
    "title": "分析代码库",
    "status": "pending",
    ...
  }
}
```

**获取任务详情**
```
GET /api/v1/tasks/:id

Response:
{
  "success": true,
  "data": {
    "id": 1,
    "task_id": "uuid",
    "status": "completed",
    "result": "任务执行结果...",
    ...
  }
}
```

**执行任务（SSE 流式输出）**
```
POST /api/v1/tasks/:id/execute

Response (Server-Sent Events):
data: {"type":"start","task_id":1,"message":"开始执行任务"}

data: {"type":"progress","task_id":1,"message":"任务执行中"}

data: {"type":"complete","task_id":1,"result":"任务执行完成"}
```

**列出所有任务**
```
GET /api/v1/tasks

Response:
{
  "success": true,
  "data": [
    {"id": 1, "title": "任务1", "status": "completed"},
    {"id": 2, "title": "任务2", "status": "running"}
  ]
}
```

**取消任务**
```
POST /api/v1/tasks/:id/cancel

Response:
{
  "success": true,
  "data": {"task_id": 1, "cancelled": true}
}
```

**删除任务**
```
DELETE /api/v1/tasks/:id

Response:
{
  "success": true,
  "data": {"task_id": 1, "deleted": true}
}
```

#### 记忆管理 API

**搜索记忆**
```
GET /api/v1/memory/search?q=关键词&category=execution&task_id=xxx&limit=100

Response:
{
  "success": true,
  "data": [
    {
      "key": "task_123_result",
      "value": {...},
      "category": "result",
      "task_id": "task_123",
      "timestamp": 1234567890
    }
  ]
}
```

**获取指定记忆**
```
GET /api/v1/memory/:key

Response:
{
  "success": true,
  "data": {
    "key": "my_key",
    "value": {...},
    "category": "context",
    ...
  }
}
```

**删除记忆**
```
DELETE /api/v1/memory/:key

Response:
{
  "success": true,
  "data": {"key": "my_key", "deleted": true}
}
```

**记忆统计**
```
GET /api/v1/memory/stats

Response:
{
  "success": true,
  "data": {
    "total": 100,
    "active": 85,
    "expired": 15,
    "category_counts": {
      "execution": 30,
      "context": 25,
      "result": 20,
      "error": 10,
      "checkpoint": 15
    }
  }
}
```

#### 健康检查 API

```
GET /health 或 /api/v1/health

Response:
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "0.2.0",
    "uptime": 3600,
    "mode": "master"
  }
}
```

#### WebSocket API

**任务实时更新**
```
WebSocket URL: ws://localhost:6767/ws/task/:id

客户端发送：
{"type": "ping"}

服务器响应：
{"type": "pong", "task_id": 1}

实时事件：
{"type": "status", "task_id": 1, "status": "Running"}
{"type": "progress", "task_id": 1, "progress": 20, "message": "执行进度: 20%"}
{"type": "completed", "task_id": 1, "result": "任务执行完成"}
```

## 数据库 Schema

### tasks 表
```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL UNIQUE,
    parent_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    group_name TEXT NOT NULL DEFAULT 'default',
    completion_criteria TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 1,
    lock_holder TEXT,
    lock_time TEXT,
    result TEXT,
    error TEXT,
    workspace_dir TEXT,
    sandboxed INTEGER NOT NULL DEFAULT 0,
    allow_network INTEGER NOT NULL DEFAULT 0,
    max_memory TEXT,
    max_cpu INTEGER,
    created_by TEXT,
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_group_name ON tasks(group_name);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
```

## 使用示例

### 1. 启动服务器

```bash
# 使用默认配置
cargo run --bin agentflow-master

# 指定端口
cargo run --bin agentflow-master -- --port 8080

# 指定地址和端口
cargo run --bin agentflow-master -- --addr 127.0.0.1 --port 6767

# 指定日志级别
cargo run --bin agentflow-master -- --log-level debug
```

### 2. 创建并执行任务

```bash
# 创建任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试任务",
    "description": "这是一个测试任务",
    "group_name": "test",
    "priority": "high"
  }'

# 执行任务（SSE）
curl -X POST http://localhost:6767/api/v1/tasks/1/execute

# 查询任务状态
curl http://localhost:6767/api/v1/tasks/1

# 列出所有任务
curl http://localhost:6767/api/v1/tasks
```

### 3. 使用记忆 API

```bash
# 搜索记忆
curl "http://localhost:6767/api/v1/memory/search?q=测试"

# 获取记忆统计
curl http://localhost:6767/api/v1/memory/stats

# 获取指定记忆
curl http://localhost:6767/api/v1/memory/my_key
```

### 4. WebSocket 连接

```javascript
// 客户端示例
const ws = new WebSocket('ws://localhost:6767/ws/task/1');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);

  switch(data.type) {
    case 'status':
      console.log('任务状态:', data.status);
      break;
    case 'progress':
      console.log('进度:', data.progress);
      break;
    case 'completed':
      console.log('完成:', data.result);
      break;
    case 'error':
      console.error('错误:', data.error);
      break;
  }
};

// 发送 ping
ws.send(JSON.stringify({type: 'ping'}));
```

## 集成说明

### 与 Team A (TaskExecutor) 集成
- TaskExecutor 提供任务执行能力
- Master 直接调用 TaskExecutor，无需远程 Worker
- 支持并发控制和状态管理

### 与 Team B (MemoryCore) 集成
- MemoryCore 提供记忆存储和检索
- 任务执行过程中可以保存和检索记忆
- 支持按任务、分类搜索

## 技术栈

- **Web 框架**: Axum 0.8
- **异步运行时**: Tokio 1.42
- **数据库**: SQLite + SQLx 0.8
- **序列化**: Serde + Serde JSON
- **WebSocket**: Tokio Tungstenite
- **日志**: Tracing
- **错误处理**: Anyhow + Thiserror

## 配置选项

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| 服务器地址 | AGENTFLOW_SERVER_ADDR | 0.0.0.0 | 监听地址 |
| 服务器端口 | AGENTFLOW_SERVER_PORT | 6767 | 监听端口 |
| 数据库 URL | AGENTFLOW_DATABASE_URL | sqlite://agentflow.db | SQLite 数据库文件 |
| 最大并发任务 | AGENTFLOW_MAX_CONCURRENT_TASKS | 10 | 同时运行的最大任务数 |
| 任务超时 | AGENTFLOW_TASK_TIMEOUT | 300 | 任务执行超时（秒） |
| 日志级别 | AGENTFLOW_LOG_LEVEL | info | 日志级别 |

## 注意事项

1. **单进程架构**: Master 直接执行任务，适合轻量级部署
2. **并发限制**: 通过 `max_concurrent_tasks` 控制并发数
3. **任务超时**: 默认 300 秒，可通过配置调整
4. **记忆清理**: 后台每 5 分钟自动清理过期记忆
5. **优雅关闭**: 支持 Ctrl+C 和 TERM 信号

## 后续优化

1. **任务队列**: 实现优先级队列
2. **任务重试**: 失败任务自动重试
3. **任务依赖**: 支持任务间依赖关系
4. **分布式**: 扩展为分布式架构
5. **认证授权**: 添加 API 认证
6. **指标监控**: Prometheus 指标导出
7. **限流**: API 限流保护

## 测试

```bash
# 运行测试
cargo test -p agentflow-master

# 运行测试并显示输出
cargo test -p agentflow-master -- --nocapture

# 运行特定测试
cargo test -p agentflow-master test_memory_set_get
```

## 开发

```bash
# 格式化代码
cargo fmt -p agentflow-master

# 检查代码
cargo clippy -p agentflow-master

# 构建文档
cargo doc -p agentflow-master --no-deps --open
```

## 许可证

MIT License - 详见项目根目录 LICENSE 文件
