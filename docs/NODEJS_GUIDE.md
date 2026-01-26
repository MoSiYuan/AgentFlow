# Node.js 版本开发指南

本文档提供 AgentFlow Node.js 版本的详细开发和使用指南。

## 目录

- [环境设置](#环境设置)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [API 参考](#api-参考)
- [故障排除](#故障排除)

## 环境设置

### 系统要求

- **Node.js**: v20.19.6 LTS（必需，v22/v24 不支持）
- **pnpm**: v10.28.1+
- **better-sqlite3**: v12.6.2
- **操作系统**: macOS, Linux, Windows

### 安装步骤

#### 1. 安装 Node.js 20

```bash
# macOS (使用 Homebrew)
brew install node@20

# 设置环境变量
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"

# 验证安装
node --version  # 应显示 v20.x.x
npm --version
```

#### 2. 安装 pnpm

```bash
npm install -g pnpm@latest
pnpm --version
```

#### 3. 安装项目依赖

```bash
cd nodejs
pnpm install
```

#### 4. 编译 better-sqlite3

```bash
npm rebuild better-sqlite3
```

#### 5. 构建项目

```bash
pnpm run build
```

## 项目结构

```
nodejs/
├── packages/
│   ├── master/           # Master 服务器
│   │   ├── src/
│   │   │   └── index.ts   # 主服务器逻辑
│   │   └── dist/          # 编译输出
│   ├── worker/           # Worker 实现
│   │   ├── src/
│   │   │   └── index.ts   # Worker 逻辑
│   │   └── dist/
│   ├── local-executor/   # 自动管理工具
│   │   ├── src/
│   │   │   └── index.ts   # LocalExecutor 实现
│   │   └── dist/
│   ├── cli/              # 命令行工具
│   │   ├── src/
│   │   │   └── index.ts   # CLI 实现
│   │   └── dist/
│   ├── database/         # SQLite 数据库
│   ├── shared/           # 共享类型定义
│   ├── skill/            # 任务管理 API
│   └── package.json
├── package.json
└── pnpm-workspace.yaml   # pnpm 工作区配置
```

## 开发指南

### 启动 Master 服务器

```bash
# 基本启动（使用默认数据库路径 ~/.claude/skills/agentflow/agentflow.db）
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
node packages/master/dist/index.js

# 自定义端口和数据库
node packages/master/dist/index.js --port 6767 --db /custom/path/agentflow.db

# 使用默认配置（端口 6767）
node packages/master/dist/index.js
```

### 启动 Worker

```bash
# 基本启动（连接到默认 Master）
node packages/worker/dist/index.js

# 连接到自定义 Master
AGENTFLOW_MASTER_URL=http://localhost:8080 node packages/worker/dist/index.js

# 自定义工作组
AGENTFLOW_GROUP_NAME=production node packages/worker/dist/index.js
```

### 创建任务

#### 方式 1: 使用 curl

```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试任务",
    "description": "echo Hello World",
    "priority": "high",
    "group_name": "default"
  }'
```

#### 方式 2: 使用 CLI

```bash
node packages/cli/dist/index.js run "echo Hello World"
```

#### 方式 3: 使用 LocalExecutor

```javascript
const { LocalExecutor } = require('./packages/local-executor/dist/index.js');

async function main() {
  const executor = new LocalExecutor({
    masterPath: './packages/master/dist/index.js',
    masterPort: 6767,
    dbPath: '~/.claude/skills/agentflow/agentflow.db',
    shutdownOnComplete: true  // 完成后自动关闭
  });

  // 执行单个任务
  await executor.executeOne('My Task', 'echo Hello World');

  // 执行多个任务
  await executor.execute([
    { title: 'Task 1', description: 'echo First' },
    { title: 'Task 2', description: 'echo Second' },
    { title: 'Task 3', description: 'echo Third' }
  ]);
}

main();
```

## API 参考

### Master API

#### 健康检查

```http
GET /health
```

**响应:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": 123.456,
  "mode": "cloud"
}
```

#### 创建任务

```http
POST /api/v1/tasks
Content-Type: application/json
```

**请求体:**
```json
{
  "title": "任务标题",
  "description": "任务描述（Shell 命令或复杂任务）",
  "priority": "high|medium|low",
  "group_name": "default",
  "parent_id": 1
}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "task_id": "TASK-00000001",
    "message": "Task created successfully"
  }
}
```

#### 查询任务状态

```http
GET /api/v1/tasks/{id}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "任务标题",
    "status": "pending|running|completed|failed",
    "result": "执行结果...",
    "error": "错误信息（如有）"
  }
}
```

#### 获取待处理任务

```http
GET /api/v1/tasks/pending?group={group_name}
```

#### 统计信息

```http
GET /api/v1/stats
```

### Worker API

#### 注册 Worker

```http
POST /api/v1/workers
Content-Type: application/json
```

**请求体:**
```json
{
  "worker_id": "worker-1",
  "worker_name": "Worker 1",
  "group_name": "default",
  "platform": "darwin",
  "capabilities": ["shell", "typescript", "claude-cli"]
}
```

#### Worker 心跳

```http
POST /api/v1/workers/{worker_id}/heartbeat
```

## 任务执行策略

Worker 根据任务描述自动选择执行方式：

### 1. Skill 调用

以 `/` 开头的命令：

```bash
description: "/commit"
→ 执行 git add -A && git commit
```

支持的 skills:
- `/commit` - Git commit
- `/test` - npm test
- `/build` - npm run build
- `/lint` - npm run lint
- `/agentflow` - agentflow 命令

### 2. 简单 Shell 命令

简短、无特殊字符的命令：

```bash
description: "echo Hello World"
description: "ls -la"
description: "npm test"
→ 直接使用 Shell 执行
```

### 3. Claude CLI（复杂任务）

满足以下任一条件：

- 包含代码块（`` ``` ``）
- 长度 > 200 字符
- 包含 AI 提示词

```javascript
description: `
分析以下代码的性能：
function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n-1) + fibonacci(n-2);
}
`
→ 调用 Claude CLI 执行
```

## 开发工作流

### 1. 修改代码

```bash
# 编辑源文件
vim packages/worker/src/index.ts

# 重新编译
pnpm run build --filter @agentflow/worker
```

### 2. 运行测试

```bash
# 启动 Master（终端 1）
node packages/master/dist/index.js --port 6767

# 启动 Worker（终端 2）
node packages/worker/dist/index.js

# 创建测试任务（终端 3）
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","description":"echo test"}'
```

### 3. 调试

启用详细日志：

```bash
# Master 日志
node packages/master/dist/index.js --port 6767 2>&1 | tee master.log

# Worker 日志
node packages/worker/dist/index.js 2>&1 | tee worker.log
```

### 4. 清理

```bash
# 停止所有进程
pkill -f "packages/master/dist/index.js"
pkill -f "packages/worker/dist/index.js"

# 清理测试数据库
rm -f test.db
```

## 故障排除

### 问题 1: better-sqlite3 模块未找到

**错误:**
```
Error: Cannot find module 'better-sqlite3'
```

**解决方案:**
```bash
npm rebuild better-sqlite3
```

### 问题 2: Node.js 版本不兼容

**错误:**
```
Error: NODE_MODULE_VERSION 137 (Node.js v24) not supported
```

**解决方案:**
```bash
# 切换到 Node.js 20
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
node --version  # 确认是 v20.x.x
```

### 问题 3: Worker 无法注册

**错误:**
```
Error: Failed to register worker: Not Found
```

**解决方案:**

确保 Master 先启动并且监听在正确的端口：

```bash
curl http://localhost:6767/health
# 应返回: {"status":"ok",...}
```

### 问题 4: 任务一直 pending

**检查:**
1. Worker 是否已注册
2. Worker 的 `group_name` 是否与任务的 `group_name` 匹配
3. 查看统计信息：
   ```bash
   curl http://localhost:6767/api/v1/stats
   ```

### 问题 5: Claude CLI 未找到

**错误:**
```
Error: Claude CLI not found
```

**解决方案:**
```bash
npm install -g @anthropic-ai/claude-code
claude --version
```

## 性能优化

### 1. 调整 Worker 并发数

修改 Worker 配置：

```typescript
const worker = new Worker({
  master_url: 'http://localhost:6767',
  group_name: 'default',
  max_concurrent: 5,  // 增加并发数
  task_timeout: 300000,
  retry_on_failure: true,
  max_retries: 3
});
```

### 2. 使用内存数据库（开发环境）

```bash
node packages/master/dist/index.js --db :memory:
```

**注意**: 生产环境使用默认数据库路径 `~/.claude/skills/agentflow/agentflow.db`

### 3. 生产环境优化

- 使用文件数据库（持久化）
- 增加数据库缓存大小
- 使用连接池
- 启用 WebSocket（实时更新）

## 部署

### 开发环境

```bash
# 启动 Master（使用默认数据库路径）
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
node packages/master/dist/index.js --port 6767

# 启动多个 Worker
node packages/worker/dist/index.js &
node packages/worker/dist/index.js &
node packages/worker/dist/index.js &
```

### 生产环境

使用进程管理器（PM2）：

```bash
npm install -g pm2

# 启动 Master（使用默认数据库路径）
pm2 start packages/master/dist/index.js --name agentflow-master -- --port 6767

# 启动 Workers
pm2 start packages/worker/dist/index.js --name agentflow-worker-1 -- --instances 3
pm2 save
pm2 startup
```

## 最佳实践

1. **始终使用 Node.js 20 LTS**
   - v22/v24 与 better-sqlite3 不兼容

2. **使用 LocalExecutor 进行自动化**
   - 自动管理 Master 和 Worker 生命周期
   - 自动清理资源

3. **合理设置任务组**
   - 不同环境使用不同的 `group_name`
   - 便于任务隔离和管理

4. **监控 Worker 状态**
   - 定期检查心跳
   - 失败后自动重启

5. **使用 CLI 进行快速测试**
   ```bash
   node packages/cli/dist/index.js run "npm test"
   ```

## 相关文档

- [CLI 使用指南](../AGENTFLOW_CLI_GUIDE.md)
- [主 README](../README.md)
- [Go 版本指南](./GO_VERSION_GUIDE.md)

---

**最后更新**: 2026-01-24
**维护者**: AgentFlow Team
