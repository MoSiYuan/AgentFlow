# AgentFlow AI 功能更新总结

## 新增功能（面向 AI Agent）

### 1. 任务升级机制

**Worker 可以创建子任务**

```bash
# 主任务
cpds add "实现用户登录" --desc "task:implement:login"

# 自动分解为：
# - 设计数据模型
# - 实现 API 端点
# - 编写测试
# - 更新文档
```

**数据库支持任务层次结构**
- 新增 `parent_id` 字段
- 支持任务树结构
- 自动聚合子任务结果

### 2. 边界安全机制

**沙箱执行**
```bash
export SANDBOXED=true
cpds worker
# 所有命令在 Docker 容器中执行
```

**路径限制**
```bash
export WORKSPACE_DIR=/tmp/workspace
export RESTRICT_PATH=true
cpds worker
# 只能操作指定目录
```

**只读模式**
```bash
export READ_ONLY=true
cpds worker
# 只能查询，不能修改
```

**资源限制**
```bash
export MAX_MEMORY=512M
export MAX_CPU=1
cpds worker
```

### 3. AI Worker 实现

**新增文件**: `internal/worker/ai_worker.go`

**功能**:
- 解析 AI 任务描述
- 自动创建子任务
- 等待子任务完成
- 聚合结果
- 安全执行控制

### 4. AI 任务类型

```bash
# AI 智能任务（自动分解）
task:implement:feature_name
task:test
task:build

# Shell 命令
shell:command arguments

# 脚本执行
script:/path/to/script.sh

# 文件操作
file:write:path:content
file:read:path
file:delete:path
```

### 5. SDK 支持

**Python SDK** (`sdk/python/cpds_ai.py`)
```python
from cpds_ai import AgentFlowClient, develop

# 快速开发
task = develop("用户登录功能", wait=True)
print(f"任务完成: {task.result}")
```

**TypeScript SDK** (`sdk/typescript/cpds_ai.ts`)
```typescript
import { AgentFlowClient, develop } from './cpds_ai';

// 快速开发
const task = await develop("用户登录功能");
console.log(`任务完成: ${task.id}`);
```

### 6. 配置文件

**Worker 配置** (`config/ai_worker.env`)
```bash
# Master 连接
MASTER_URL=http://localhost:8848

# 安全配置
WORKSPACE_DIR=/tmp/cpds-workspace
SANDBOXED=false
RESTRICT_PATH=false
READ_ONLY=false

# 资源限制
MAX_MEMORY=512M
MAX_CPU=1
```

### 7. 文档

**新增文档**:
- [AI_QUICKSTART.md](docs/AI_QUICKSTART.md) - 3 分钟上手
- [AI_DEPLOYMENT.md](docs/AI_DEPLOYMENT.md) - 详细部署指南
- [AI_README.md](docs/AI_README.md) - AI 专用说明

**数据库更新**:
- `tasks` 表新增字段：
  - `parent_id` - 父任务 ID
  - `priority` - 任务优先级
  - `workspace_dir` - 工作目录
  - `sandboxed` - 是否沙箱执行
  - `allow_network` - 是否允许网络
  - `max_memory` - 最大内存
  - `max_cpu` - 最大 CPU
  - `created_by` - 创建者

## 使用场景

### 场景 1：本地 AI 开发

```bash
# 1. 启动
make build && ./bin/cpds init agentflow.db && ./bin/cpds master

# 2. Claude Code 调用
import { develop } from './cpds_ai';
await develop("用户登录功能");

# 3. 自动执行
# - 设计数据模型
# - 实现 API
# - 编写测试
# - 更新文档
```

### 场景 2：云端 Master + 本地 AI

```bash
# 服务器
./bin/cpds master --db /data/agentflow.db --host 0.0.0.0

# 本地 AI
export MASTER_URL=http://server.com:8848
./bin/cpds worker

# 创建任务
cpds add "训练模型" --desc "task:train:model" --group gpu
```

### 场景 3：安全沙箱环境

```bash
# 启动沙箱 Worker
export SANDBOXED=true
export WORKSPACE_DIR=/tmp/safe-workspace
export RESTRICT_PATH=true
./bin/cpds worker

# AI 可以安全执行未知命令
cpds add "测试代码" --desc "shell:make test"
```

## 部署模式对比

| 模式 | 优势 | 适用场景 |
|------|------|----------|
| 本地模式 | 性能高、无网络开销 | AI 本地开发 |
| 云端模式 | 分布式、可扩展 | 多机器协作 |
| 沙箱模式 | 安全隔离 | 执行不可信代码 |

## 下一步

1. 运行 `go mod tidy` 下载依赖
2. 运行 `make build` 编译项目
3. 运行 `./bin/cpds init agentflow.db` 初始化
4. 运行 `./bin/cpds master` 启动服务
5. 使用 SDK 或 CLI 创建任务

## 安全建议

1. **生产环境**: 启用沙箱模式
2. **路径限制**: 限制工作目录
3. **资源限制**: 设置 CPU 和内存上限
4. **审计日志**: 启用操作审计
5. **网络隔离**: 使用 VPN 或专用网络
