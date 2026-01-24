# AgentFlow AI 部署指南

> 面向 AI Agent 的部署文档

## 快速部署（AI 专用）

### 模式一：云端 Master + 本地 Worker

**适用场景**：分布式开发，AI 在本地终端执行任务

```bash
# === 云端 Master 服务器 ===
# SSH 登录服务器
ssh user@server.example.com

# 安装 AgentFlow
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go
make build

# 启动 Master
./bin/cpds master \
  --db /data/agentflow.db \
  --host 0.0.0.0 \
  --port 6767

# === 本地 Worker（AI 所在机器）===
export MASTER_URL=http://server.example.com:6767
export WORKER_GROUP=local
./bin/cpds worker
```

### 模式二：本地 Master + Worker

**适用场景**：单机开发，AI 在本地执行所有任务

```bash
# 1. 安装
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go
make build

# 2. 初始化
./bin/cpds init agentflow.db

# 3. 启动 Master（自动启动本地 Workers）
./bin/cpds master --db agentflow.db

# 4. 添加任务（AI 调用）
./bin/cpds add "实现用户登录功能" \
  --desc "task:implement_login" \
  --group local
```

## AI 使用方式

### 作为 Skill 调用

```bash
# AI Skill: cpds
# 功能：创建并执行开发任务

# 创建任务
cpds add "任务描述" \
  --desc "task_type:params" \
  --group worker_group

# 查看状态
cpds list --status running

# 查看结果
cpds list --status completed
```

### 任务描述格式

| 前缀 | 说明 | 示例 |
|------|------|------|
| `shell:` | Shell 命令 | `shell:go test ./...` |
| `script:` | 脚本文件 | `script:./deploy.sh` |
| `task:` | AI 任务（由 Claude 执行） | `task:implement_feature:{...}` |
| `file:` | 文件操作 | `file:write:README.md:content` |

## 边界安全机制

### 1. 沙箱执行

```bash
# Docker 沙箱
docker run --rm -v $(pwd):/workspace cpds-sandbox \
  cpds worker --group sandbox

# 限制执行目录
export WORKSPACE_DIR=/tmp/cpds-workspace
export RESTRICT_PATH=true
```

### 2. 权限控制

```bash
# 只读模式
export READ_ONLY=true

# 禁止网络
export DISABLE_NETWORK=true

# 资源限制
export MAX_MEMORY=512M
export MAX_CPU=1
```

### 3. 审计日志

```bash
# 所有操作记录
export AUDIT_LOG=/var/log/cpds/audit.log
export ENABLE_AUDIT=true
```

## 任务升级机制

Worker 可以创建子任务：

```go
// Worker 执行时，可以创建新的任务
curl -X POST $MASTER_URL/api/v1/tasks \
  -d '{
    "title": "子任务：编写测试",
    "description": "task:write_test:login.go",
    "group_name": "testing"
  }'
```

示例：

1. 主任务：`实现用户登录功能`
   - 子任务：`设计数据模型`
   - 子任务：`实现 API 端点`
   - 子任务：`编写单元测试`
   - 子任务：`更新文档`

## 全自动执行流程

```
1. AI 创建主任务
   ↓
2. Master 分配给 Worker
   ↓
3. Worker 执行，遇到需要分解的任务
   ↓
4. Worker 创建子任务
   ↓
5. 其他 Worker 执行子任务
   ↓
6. 子任务完成，聚合结果
   ↓
7. 主任务完成
```

## 环境变量配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MASTER_URL` | Master 地址 | http://localhost:6767 |
| `WORKER_GROUP` | Worker 组 | auto-detected |
| `WORKSPACE_DIR` | 工作目录 | /tmp/cpds-workspace |
| `RESTRICT_PATH` | 路径限制 | false |
| `READ_ONLY` | 只读模式 | false |
| `DISABLE_NETWORK` | 禁用网络 | false |
| `MAX_MEMORY` | 最大内存 | 512M |
| `MAX_CPU` | 最大 CPU | 1 |
| `ENABLE_AUDIT` | 启用审计 | false |
| `AUDIT_LOG` | 审计日志 | /var/log/cpds/audit.log |

## AI 调用示例

### Claude Code 使用

```typescript
// Claude Code Skill
interface AgentFlowTask {
  title: string;
  description: string;
  group: string;
  priority?: number;
}

async function createTask(task: AgentFlowTask) {
  const result = await exec('cpds', [
    'add',
    task.title,
    '--desc', task.description,
    '--group', task.group
  ]);
  return result;
}

// 示例：创建开发任务
await createTask({
  title: '实现用户注册功能',
  description: 'task:implement:registration',
  group: 'local'
});
```

### Command Line 调用

```bash
# 创建任务
cpds add "实现登录API" \
  --desc "task:implement_api:POST /api/login" \
  --group local

# 等待完成
while true; do
  status=$(cpds list --status running | grep "实现登录API")
  if [ -z "$status" ]; then
    break
  fi
  sleep 2
done

# 获取结果
cpds list --status completed | grep "实现登录API"
```

## 故障处理

### Worker 连接失败

```bash
# 检查 Master
curl $MASTER_URL/health

# 检查网络
telnet $MASTER_URL 6767

# 重启 Worker
pkill -f "cpds worker"
cpds worker
```

### 任务卡住

```bash
# 查看锁状态
cpds list --status running

# 强制解锁（紧急情况）
sqlite3 agentflow.db "UPDATE tasks SET status='pending', lock_holder=NULL WHERE id=1"
```

## 最佳实践

1. **任务分解**：大任务分解为小任务
2. **明确描述**：使用标准化的任务描述格式
3. **合理分组**：按环境和能力分组
4. **资源限制**：设置合理的 CPU 和内存限制
5. **日志记录**：启用审计日志
6. **定期备份**：备份 SQLite 数据库

## 安全建议

1. 使用 VPN 或专用网络
2. 启用 HTTPS/TLS
3. 实现 API Key 认证
4. 限制 Worker 权限
5. 使用沙箱环境执行任务
6. 定期更新依赖
