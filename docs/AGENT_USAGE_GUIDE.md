# AgentFlow Agent 分场景引导指南

**最后更新**: 2026-01-28
**适用版本**: v0.2.1+

---

## 📖 文档概述

本指南帮助你根据不同的使用场景，快速上手 AgentFlow AI Agent 任务协作系统。无论你是个人开发者、小团队，还是企业用户，都能找到适合你的配置和使用方式。

### 什么是 AgentFlow？

AgentFlow 是一个**纯 Rust 编写的单进程高性能 AI Agent 任务编排系统**，具有以下特点：

- ✅ **单二进制** - 一个可执行文件，无需依赖
- ✅ **单进程** - Master = Worker，无进程间通信
- ✅ **高性能** - 基于 Tokio 异步运行时，内存 < 100MB
- ✅ **直接执行** - 通过 tokio::process 直接执行 Claude CLI
- ✅ **向量记忆** - 基于 SQLite 的向量索引和语义检索
- ✅ **沙箱安全** - 完整的路径验证和进程隔离
- ✅ **REST API** - 14 个 HTTP 端点
- ✅ **实时流传输** - WebSocket 和 SSE 流式传输支持

---

## 🚀 快速开始

### 5 分钟快速上手

#### 1️⃣ 安装 AgentFlow

**Linux/macOS**:
```bash
curl -fsSL https://raw.githubusercontent.com/MoSiYuan/AgentFlow/main/scripts/install.sh | bash
```

**Windows**:
```powershell
irm https://raw.githubusercontent.com/MoSiYuan/AgentFlow/main/scripts/install.ps1 | iex
```

#### 2️⃣ 创建配置文件

```bash
mkdir -p ~/.agentflow
cat > ~/.agentflow/config.toml << 'EOF'
[server]
port = 6767

[database]
url = "sqlite://agentflow.db"

[executor]
max_concurrent_tasks = 5
task_timeout = 300
EOF
```

#### 3️⃣ 启动服务器

```bash
agentflow server local
```

服务器将在 `http://localhost:6767` 启动。

#### 4️⃣ 创建第一个任务

```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello AgentFlow",
    "description": "echo \"Hello from AgentFlow!\"",
    "priority": "medium"
  }'
```

#### 5️⃣ 查看任务状态

```bash
curl http://localhost:6767/api/v1/tasks/1
```

恭喜！你已经成功运行了第一个 AgentFlow 任务！🎉

---

## 🎯 场景一：单机任务自动化

### 适用场景

- 个人开发者日常自动化
- 本地脚本执行
- 开发环境工具链集成
- CI/CD 本地测试

### 特点

✅ **简单快速** - 无需网络，即开即用
✅ **低延迟** - 本地执行，响应迅速
✅ **零依赖** - 单机运行，无需额外服务
✅ **低成本** - 无需云服务费用

### 配置步骤

#### 1. 基础配置

创建 `~/.agentflow/config.toml`:

```toml
[server]
host = "127.0.0.1"  # 仅本地访问
port = 6767

[database]
url = "sqlite://agentflow.db"

[executor]
max_concurrent_tasks = 5      # 同时运行的任务数
task_timeout = 300            # 任务超时时间（秒）
backend = "claude-cli"        # 使用 Claude CLI

[memory]
backend = "sqlite"             # 持久化记忆
database_url = "sqlite://agentflow_memory.db"
default_ttl = 3600            # 记忆保留时间（秒）

[sandbox]
enabled = true                 # 启用沙箱
allow_network = false         # 禁止网络访问
default_workspace = "./workspace"  # 工作空间
```

#### 2. 启动本地模式

```bash
# 方式 1: 使用默认配置
agentflow server local

# 方式 2: 指定配置文件
agentflow server --config ~/.agentflow/config.toml

# 方式 3: 使用环境变量
AGENTFLOW_SERVER_PORT=8080 agentflow server
```

#### 3. 配置项目专属规则（可选）

在项目根目录创建 `AGENTFLOW.md`:

```markdown
# AgentFlow 项目配置

## 1. 构建系统

### Rust 项目
- 使用 `cargo build --release` 进行编译
- 运行测试：`cargo test --lib`

## 2. 测试工作流

### 正确的测试命令
```bash
# 不要运行 `make test`
# 正确命令：
cargo test --lib
cargo test --test integration_tests
```

## 3. 关键技能

### ✅ DO (应该做)
- 提交代码前运行 `cargo clippy`
- 遵循 Rust API 指南

### ❌ DON'T (不应该做)
- 不要修改 `vendor/` 目录
- 不要提交 `Cargo.lock`

### 🔧 Special Skills
- 编译错误时，删除 `target/` 目录
- 测试失败时，检查 `.db-shm` 和 `.db-wal` 文件

## 4. 调试策略

### 常见问题
- **编译错误**: 检查 Rust 版本
- **测试失败**: 使用 `RUST_LOG=debug` 查看详细信息
```

### 示例场景

#### 示例 1：自动化测试

```bash
# 创建测试任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "运行单元测试",
    "description": "cd /path/to/project && cargo test --lib",
    "priority": "high"
  }'
```

#### 示例 2：代码生成

```bash
# 创建代码生成任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "生成 REST API",
    "description": "使用 Axum 创建一个 CRUD API",
    "priority": "medium",
    "metadata": {
      "framework": "axum",
      "database": "sqlite"
    }
  }'
```

#### 示例 3：定时任务

使用 cron 配合 AgentFlow：

```bash
# 编辑 crontab
crontab -e

# 每天凌晨 2 点运行测试
0 2 * * * curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "每日测试", "description": "cargo test"}'
```

### 高级功能

#### 记忆系统

AgentFlow 会自动记录执行历史，并在下次任务中复用经验：

```bash
# 查看记忆
curl http://localhost:6767/api/v1/memory?search=编译错误

# 清除旧记忆
curl -X DELETE http://localhost:6767/api/v1/memory?before=2026-01-01
```

#### 任务优先级

```bash
# 紧急任务优先执行
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "紧急修复",
    "description": "修复登录 bug",
    "priority": "urgent"
  }'
```

#### 实时日志流

```bash
# 使用 SSE 流式查看日志
curl -N http://localhost:6767/api/v1/tasks/1/execute
```

---

## 👥 场景二：团队协作

### 适用场景

- 小团队（2-10 人）任务管理
- 代码审查自动化
- 团队知识共享
- CI/CD 流水线集成

### 特点

✅ **多用户支持** - API Key 认证
✅ **任务队列** - 按优先级调度
✅ **记忆共享** - 团队知识复用
✅ **权限控制** - 细粒度权限管理

### 配置步骤

#### 1. 服务器配置

创建 `~/.agentflow/config.toml`:

```toml
[server]
host = "0.0.0.0"  # 允许外部访问
port = 6767

[database]
url = "postgresql://agentflow:password@db.example.com/agentflow"
max_connections = 20

[executor]
max_concurrent_tasks = 20
task_timeout = 600

[memory]
backend = "postgresql"  # 团队共享记忆
database_url = "postgresql://agentflow:password@db.example.com/agentflow_memory"
enable_persistence = true
default_ttl = 86400  # 保留 24 小时

[auth]
enabled = true
api_keys = [
  "sk-team-alice-1234567890abcdef",
  "sk-team-bob-1234567890abcdef",
  "sk-team-charlie-1234567890abcdef"
]
header = "X-API-Key"

[cors]
enabled = true
allowed_origins = ["https://team-dashboard.example.com"]
allow_credentials = true
```

#### 2. 启动服务器

```bash
# 使用 systemd 管理（Linux）
sudo tee /etc/systemd/system/agentflow.service << EOF
[Unit]
Description=AgentFlow Server
After=network.target

[Service]
Type=simple
User=agentflow
WorkingDirectory=/opt/agentflow
ExecStart=/usr/local/bin/agentflow server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable agentflow
sudo systemctl start agentflow
```

#### 3. 配置 Nginx 反向代理

```nginx
# /etc/nginx/sites-available/agentflow
server {
    listen 80;
    server_name agentflow.team.example.com;

    location / {
        proxy_pass http://localhost:6767;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-API-Key $http_x_api_key;
    }
}
```

### 示例场景

#### 示例 1：代码审查自动化

```bash
# 创建代码审查任务
curl -X POST http://agentflow.team.example.com/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-team-alice-1234567890abcdef" \
  -d '{
    "title": "审查 PR #123",
    "description": "审查 feature/login 分支的代码变更",
    "priority": "high",
    "group": "code-review"
  }'
```

#### 示例 2：团队知识共享

```bash
# 查询团队记忆
curl http://agentflow.team.example.com/api/v1/memory/search \
  -H "X-API-Key: sk-team-alice-1234567890abcdef" \
  -G \
  --data-urlencode "query=如何修复数据库迁移失败"

# 输出：
# {
#   "results": [
#     {
#       "content": "数据库迁移失败时，删除 *.db-shm 和 *.db-wal 文件",
#       "relevance": 0.95,
#       "timestamp": "2026-01-27T10:30:00Z"
#     }
#   ]
# }
```

#### 示例 3：CI/CD 集成

在 GitHub Actions 中使用 AgentFlow：

```yaml
# .github/workflows/agentflow-review.yml
name: AgentFlow Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - name: Create review task
        run: |
          curl -X POST ${{ secrets.AGENTFLOW_URL }}/api/v1/tasks \
            -H "Content-Type: application/json" \
            -H "X-API-Key: ${{ secrets.AGENTFLOW_API_KEY }}" \
            -d '{
              "title": "Review PR ${{ github.event.number }}",
              "description": "Review branch ${{ github.event.pull_request.head.ref }}",
              "priority": "high",
              "metadata": {
                "pr_url": "${{ github.event.pull_request.html_url }}",
                "author": "${{ github.event.pull_request.user.login }}"
              }
            }'
```

### 团队最佳实践

#### 1. API Key 管理

- 为每个团队成员生成独立的 API Key
- 定期轮换 API Key
- 使用环境变量管理密钥

```bash
# 生成新的 API Key
openssl rand -hex 32

# 在 .bashrc 中配置
export AGENTFLOW_API_KEY="sk-your-key-here"
export AGENTFLOW_URL="http://agentflow.team.example.com"
```

#### 2. 任务分组

使用 `group` 字段组织任务：

```bash
# 开发组任务
curl -X POST $AGENTFLOW_URL/api/v1/tasks \
  -H "X-API-Key: $AGENTFLOW_API_KEY" \
  -d '{"title": "实现新功能", "group": "development"}'

# 测试组任务
curl -X POST $AGENTFLOW_URL/api/v1/tasks \
  -H "X-API-Key: $AGENTFLOW_API_KEY" \
  -d '{"title": "运行集成测试", "group": "testing"}'
```

#### 3. 任务监控

```bash
# 查看所有进行中的任务
curl $AGENTFLOW_URL/api/v1/tasks?status=running \
  -H "X-API-Key: $AGENTFLOW_API_KEY" | jq .

# 查看特定组的任务
curl "$AGENTFLOW_URL/api/v1/tasks?group=code-review" \
  -H "X-API-Key: $AGENTFLOW_API_KEY" | jq .
```

---

## 🤖 场景三：AI 助手集成

### 适用场景

- 接入 Claude、ChatGPT 等助手
- 智谱清言集成
- 多 AI 模型协同
- 自然语言任务创建

### 特点

✅ **Webhook 接入** - 接收 AI 平台推送
✅ **意图解析** - 自动理解用户需求
✅ **双向通信** - 任务结果回调
✅ **安全验证** - 签名验证、IP 白名单

### 配置步骤

#### 1. AgentFlow 配置

创建 `~/.agentflow/config.toml`:

```toml
[server]
host = "0.0.0.0"
port = 6767

[webhook]
enabled = true
path = "/api/v1/webhook"
secret = "your-webhook-secret-key"
algorithm = "sha256"
signature_header = "X-Webhook-Signature"

# IP 白名单（智谱清言 IP 段）
ip_whitelist_enabled = true
ip_whitelist = ["203.119.0.0/16", "43.243.0.0/16"]

# 速率限制
rate_limit = 100
rate_limit_burst = 20

# 认证
require_auth = true
auth_header = "X-API-Key"
auth_key = "your-expected-api-key"

[zhipu]
enabled = true
api_key = "your-zhipu-api-key"
model = "glm-4"
endpoint = "https://open.bigmodel.cn/api/paas/v4"
max_tokens = 8192
temperature = 0.7
timeout = 300
callback_url = "https://your-domain.com/api/v1/callback"
```

#### 2. 启动 Cloud 模式

```bash
agentflow server cloud
```

#### 3. 配置公网 URL（测试用）

使用 ngrok 创建隧道：

```bash
# 安装 ngrok
brew install ngrok  # macOS
# 或从 https://ngrok.com/ 下载

# 启动隧道
ngrok http 6767

# 输出：
# Forwarding https://abc123.ngrok.io -> http://localhost:6767
```

#### 4. 配置智谱清言 Webhook

在智谱清言控制台设置 Webhook URL：

```
URL: https://abc123.ngrok.io/api/v1/webhook
Method: POST
Content-Type: application/json
Events:
  ✅ 消息接收 (message.received)
  ✅ 用户查询 (user.query)
  ✅ 任务请求 (task.request)
```

### 示例场景

#### 示例 1：简单任务创建

用户在智谱清言中输入：

```
"帮我创建一个任务，分析这个项目的代码结构"
```

智谱清言发送 Webhook 到 AgentFlow：

```json
{
  "event": "message.received",
  "timestamp": "2026-01-28T10:30:00Z",
  "data": {
    "message_id": "msg_123456789",
    "user_id": "user_abc123",
    "conversation_id": "conv_xyz789",
    "content": "帮我创建一个任务，分析这个项目的代码结构",
    "metadata": {
      "source": "zhipu",
      "model": "glm-4",
      "language": "zh-CN"
    }
  }
}
```

AgentFlow 返回：

```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "task_id": "task_abc123",
    "status": "pending",
    "title": "分析项目代码结构",
    "priority": "medium"
  }
}
```

#### 示例 2：复杂任务带参数

用户输入：

```
"执行以下任务：
1. 阅读 README.md
2. 总结项目架构
3. 生成架构图

优先级：高
分组：文档"
```

Webhook 请求：

```json
{
  "event": "user.query",
  "timestamp": "2026-01-28T10:35:00Z",
  "data": {
    "message_id": "msg_987654321",
    "user_id": "user_def456",
    "content": "执行以下任务：\n1. 阅读README.md\n2. 总结项目架构\n3. 生成架构图\n\n优先级：高\n分组：文档",
    "parameters": {
      "priority": "high",
      "group": "documentation",
      "timeout": 600
    },
    "metadata": {
      "source": "zhipu",
      "reply_to_webhook": "https://api.zhipu.ai/v1/webhook/callback"
    }
  }
}
```

#### 示例 3：任务状态回调

任务完成后，AgentFlow 回调智谱清言：

```json
{
  "event": "task.completed",
  "timestamp": "2026-01-28T10:40:00Z",
  "data": {
    "task_id": "task_abc123",
    "status": "completed",
    "result": "任务完成！\n\n项目架构分析：\n1. 前端：React + TypeScript\n2. 后端：Rust (Axum)\n3. 数据库：SQLite\n4. AI 执行：Claude CLI\n\n架构图已生成：architecture.png",
    "duration": 45.2,
    "tokens_used": 1250
  }
}
```

### 协同场景示例

#### 场景 1：通勤路上（手机）

1. 打开智谱清言，输入：
   ```
   "DiveAdstra 编译一下，如果成功就跑个 DX12 测试。"
   ```

2. 智谱清言 → AgentFlow (Webhook) → CLI Planner → 任务分发

3. 智谱清言回复：
   ```
   "指令已下发，正在执行。"
   ```

#### 场景 2：回到电脑前

1. 打开浏览器访问 AgentFlow 控制台

2. 查看实时日志：
   ```
   [Bash] msbuild DiveAdstra.sln ...
   [Build] Success!
   [Bash] ./Binaries/DX12Benchmark.exe ...
   [Result] FPS: 120
   ```

### 安全最佳实践

#### 1. Webhook 签名验证

```bash
# 生成共享密钥
openssl rand -hex 32 > /etc/agentflow/webhook.secret
chmod 600 /etc/agentflow/webhook.secret

# 在配置文件中引用
cat >> ~/.agentflow/config.toml << EOF
[webhook]
secret = "$(cat /etc/agentflow/webhook.secret)"
EOF
```

#### 2. IP 白名单

```bash
# 获取智谱清言 IP 段
curl https://open.bigmodel.cn/api/ip-ranges

# 配置到 AgentFlow
cat >> ~/.agentflow/config.toml << EOF
[webhook]
ip_whitelist_enabled = true
ip_whitelist = ["203.119.0.0/16"]
EOF
```

#### 3. 速率限制

```toml
[webhook]
# 每分钟最多 100 个请求
rate_limit = 100
# 突发流量最多 20 个
rate_limit_burst = 20
```

---

## 🌐 场景四：分布式任务调度

### 适用场景

- 大规模并行任务处理
- 多服务器协同工作
- 高可用任务调度
- 复杂工作流编排

### 特点

✅ **Master 集群** - 基于 Raft 的领导选举
✅ **DAG 工作流** - 任务依赖管理
✅ **优先级队列** - 智能任务调度
✅ **Worker 注册** - 健康检查和负载均衡

### 配置步骤

#### 1. 启动 Master 集群（3 个节点）

**终端 1 - Master 1**:
```bash
cargo run --bin agentflow-master -- \
  --node-id master-1 \
  --port 6767 \
  --peers master-1:6767,master-2:6768,master-3:6769
```

**终端 2 - Master 2**:
```bash
cargo run --bin agentflow-master -- \
  --node-id master-2 \
  --port 6768 \
  --peers master-1:6767,master-2:6768,master-3:6769
```

**终端 3 - Master 3**:
```bash
cargo run --bin agentflow-master -- \
  --node-id master-3 \
  --port 6769 \
  --peers master-1:6767,master-2:6768,master-3:6769
```

#### 2. 启动 Workers

**Worker 1**:
```bash
cargo run --bin agentflow-worker -- \
  --worker-id worker-1 \
  --master-url http://localhost:6767 \
  --group builders
```

**Worker 2**:
```bash
cargo run --bin agentflow-worker -- \
  --worker-id worker-2 \
  --master-url http://localhost:6767 \
  --group testers
```

#### 3. 验证集群状态

```bash
# 查看当前 Leader
curl http://localhost:6767/api/v1/cluster/leader

# 查看所有节点
curl http://localhost:6767/api/v1/cluster/nodes

# 查看所有 Workers
curl http://localhost:6767/api/v1/workers
```

### 示例场景

#### 示例 1：CI/CD 工作流

创建完整的 CI/CD 流水线：

```bash
curl -X POST http://localhost:6767/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ci-cd-pipeline",
    "tasks": [
      {
        "id": "lint",
        "name": "Run Linter",
        "priority": "medium",
        "dependencies": [],
        "group": "builders"
      },
      {
        "id": "build",
        "name": "Build Project",
        "priority": "high",
        "dependencies": ["lint"],
        "group": "builders"
      },
      {
        "id": "unit-test",
        "name": "Unit Tests",
        "priority": "high",
        "dependencies": ["build"],
        "group": "testers"
      },
      {
        "id": "integration-test",
        "name": "Integration Tests",
        "priority": "medium",
        "dependencies": ["build"],
        "group": "testers"
      },
      {
        "id": "deploy-staging",
        "name": "Deploy to Staging",
        "priority": "medium",
        "dependencies": ["unit-test", "integration-test"],
        "group": "deployers"
      },
      {
        "id": "smoke-test",
        "name": "Smoke Tests",
        "priority": "high",
        "dependencies": ["deploy-staging"],
        "group": "testers"
      },
      {
        "id": "deploy-production",
        "name": "Deploy to Production",
        "priority": "urgent",
        "dependencies": ["smoke-test"],
        "group": "deployers"
      }
    ]
  }'
```

#### 示例 2：批量数据处理

并行处理多个数据文件：

```bash
# 创建任务组
for file in data_*.csv; do
  curl -X POST http://localhost:6767/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{
      \"title\": \"Process $file\",
      \"description\": \"python process_data.py $file\",
      \"priority\": \"medium\",
      \"group\": \"processors\"
    }"
done
```

#### 示例 3：Leader 故障转移

```bash
# 1. 查看当前 Leader
curl http://localhost:6767/api/v1/cluster/leader
# 输出：{"leader_id": "master-1", "term": 1}

# 2. 停止 Leader (Ctrl+C 在 master-1 终端)

# 3. 观察自动选举
curl http://localhost:6768/api/v1/cluster/leader
# 输出：{"leader_id": "master-2", "term": 2}

# 4. 重新启动 master-1，观察它自动加入集群作为 Follower
```

### 高级功能

#### 分布式锁

```bash
# 获取部署锁
curl -X POST http://localhost:6767/api/v1/locks/acquire \
  -H "Content-Type: application/json" \
  -d '{"lock_key": "deploy-production", "ttl": 300}'

# 返回：{"lock_id": "lock-xyz", "acquired": true}

# 释放锁
curl -X POST http://localhost:6767/api/v1/locks/release \
  -H "Content-Type: application/json" \
  -d '{"lock_key": "deploy-production"}'
```

#### 任务优先级调整

```bash
# 提高任务优先级
curl -X PATCH http://localhost:6767/api/v1/tasks/task-123/priority \
  -H "Content-Type: application/json" \
  -d '{"priority": "urgent"}'
```

#### Worker 维护模式

```bash
# 标记 Worker 为维护模式（不再接收新任务）
curl -X PATCH http://localhost:6767/api/v1/workers/worker-1 \
  -H "Content-Type: application/json" \
  -d '{"status": "draining"}'

# 恢复正常模式
curl -X PATCH http://localhost:6767/api/v1/workers/worker-1 \
  -H "Content-Type: application/json" \
  -d '{"status": "active"}'
```

---

## 🔧 场景五：开发环境定制

### 适用场景

- 自定义 Worker 开发
- Skill 插件开发
- 企业级定制
- 集成内部系统

### 特点

✅ **模块化设计** - 易于扩展
✅ **插件系统** - 自定义 Skill
✅ **API 完整** - 所有功能可编程
✅ **类型安全** - Rust 强类型保证

### 开发自定义 Skill

#### 1. Skill 定义

在项目 `AGENTFLOW.md` 中定义 Skill：

```markdown
# AgentFlow 项目配置

## 3. 自定义 Skills

### skill: deploy-staging
**描述**: 部署到测试环境
**命令**:
```bash
kubectl apply -f k8s/staging/
kubectl rollout status deployment/web-app -n staging
```
**参数**:
- `env`: 环境名称（默认：staging）
- `timeout`: 超时时间（默认：300）

### skill: backup-database
**描述**: 备份数据库
**命令**:
```bash
pg_dump $DATABASE_URL | gzip > backup_$(date +%Y%m%d).sql.gz
```
**参数**:
- `database`: 数据库名称
- `retention`: 保留天数（默认：7）
```

#### 2. Rust 插件开发

创建自定义 Worker 插件：

```rust
// agentflow-worker-plugin/src/lib.rs
use agentflow_core::WorkerPlugin;

pub struct CustomWorkerPlugin {
    config: PluginConfig,
}

impl WorkerPlugin for CustomWorkerPlugin {
    fn name(&self) -> &str {
        "custom-worker"
    }

    fn execute(&self, task: &Task) -> Result<TaskResult> {
        match task.group.as_str() {
            "deploy" => self.deploy(task),
            "backup" => self.backup(task),
            _ => Err(Error::UnsupportedGroup),
        }
    }

    fn health_check(&self) -> HealthStatus {
        HealthStatus {
            healthy: true,
            message: "OK".to_string(),
        }
    }
}

impl CustomWorkerPlugin {
    fn deploy(&self, task: &Task) -> Result<TaskResult> {
        // 自定义部署逻辑
        let env = task.metadata.get("env").unwrap_or(&"staging".to_string());
        let cmd = format!("kubectl apply -f k8s/{}/", env);

        let output = std::process::Command::new("kubectl")
            .args(["apply", "-f", &format!("k8s/{}/", env)])
            .output()?;

        Ok(TaskResult {
            status: TaskStatus::Completed,
            output: String::from_utf8_lossy(&output.stdout).to_string(),
        })
    }

    fn backup(&self, task: &Task) -> Result<TaskResult> {
        // 自定义备份逻辑
        Ok(TaskResult {
            status: TaskStatus::Completed,
            output: "Backup completed".to_string(),
        })
    }
}
```

#### 3. 集成到 Worker

```rust
// agentflow-worker/src/main.rs
use agentflow_worker_plugin::CustomWorkerPlugin;

#[tokio::main]
async fn main() {
    let plugin = CustomWorkerPlugin::new(config);

    let worker = Worker::builder()
        .with_plugin(plugin)
        .build();

    worker.run().await;
}
```

### 企业级集成示例

#### 1. 集成内部监控系统

```rust
pub struct MonitoringPlugin {
    prometheus_url: String,
}

impl WorkerPlugin for MonitoringPlugin {
    fn on_task_complete(&self, task: &Task, result: &TaskResult) {
        // 上报指标到 Prometheus
        self.prometheus_push(
            "agentflow_task_duration_seconds",
            result.duration.as_secs_f64(),
            &[("task_group", &task.group)],
        );
    }
}
```

#### 2. 集成日志系统

```rust
pub struct LoggingPlugin {
    elasticsearch_url: String,
}

impl WorkerPlugin for LoggingPlugin {
    fn on_task_start(&self, task: &Task) {
        let log_entry = json!({
            "timestamp": Utc::now(),
            "task_id": task.id,
            "title": task.title,
            "group": task.group,
            "priority": task.priority,
        });

        self.elasticsearch_index("agentflow-tasks", &log_entry);
    }
}
```

#### 3. 集成权限系统

```rust
pub struct AuthPlugin {
    ldap_server: String,
}

impl WorkerPlugin for AuthPlugin {
    fn can_execute_task(&self, user: &User, task: &Task) -> bool {
        // 检查 LDAP 权限
        self.ldap_check_permission(user, &task.group)
    }
}
```

### 测试与调试

#### 单元测试

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_custom_skill_deploy() {
        let plugin = CustomWorkerPlugin::default();
        let task = Task {
            group: "deploy".to_string(),
            metadata: hashmap!{
                "env".to_string() => "staging".to_string()
            },
        };

        let result = plugin.deploy(&task).unwrap();
        assert_eq!(result.status, TaskStatus::Completed);
    }
}
```

#### 集成测试

```bash
# 运行所有测试
cargo test --workspace

# 运行特定测试
cargo test --package agentflow-worker-plugin test_deploy

# 运行集成测试
cargo test --test '*'
```

---

## 📊 场景对比表

| 场景 | 适用团队规模 | 技术复杂度 | 部署方式 | 推荐配置 |
|------|-------------|-----------|---------|---------|
| **单机任务自动化** | 1 人 | ⭐ 简单 | 单机运行 | - 5 个并发任务<br>- SQLite 存储<br>- 内存记忆<br>- 无需认证 |
| **团队协作** | 2-10 人 | ⭐⭐ 中等 | 内网服务器 | - 20 个并发任务<br>- PostgreSQL 存储<br>- 持久化记忆<br>- API Key 认证<br>- Nginx 反向代理 |
| **AI 助手集成** | 1-50 人 | ⭐⭐⭐ 较高 | 公网服务器 | - 10 个并发任务<br>- SQLite 存储<br>- Webhook 接入<br>- IP 白名单<br>- 签名验证<br>- 速率限制 |
| **分布式任务调度** | 10+ 人 | ⭐⭐⭐⭐ 高 | Master + Worker 集群 | - 100 个并发任务<br>- PostgreSQL 存储<br>- 3 Master 节点<br>- 多 Worker 分组<br>- 分布式锁 |
| **开发环境定制** | 任意 | ⭐⭐⭐⭐⭐ 极高 | 自定义部署 | - 自定义插件<br>- 企业级集成<br>- 监控告警<br>- 权限系统<br>- 日志收集 |

### 选择建议

#### 个人开发者

**推荐场景**: 单机任务自动化

**原因**:
- 零成本，无需服务器
- 配置简单，5 分钟上手
- 满足日常开发需求

**典型用途**:
- 自动化测试
- 代码生成
- 文档生成
- 脚本自动化

#### 小团队（2-10 人）

**推荐场景**: 团队协作

**原因**:
- 共享记忆，知识复用
- 任务队列，有序执行
- 权限控制，安全隔离
- 成本可控，一台服务器

**典型用途**:
- 代码审查
- CI/CD 流水线
- 团队知识管理
- 自动化部署

#### 中大型团队（10+ 人）

**推荐场景**: 分布式任务调度

**原因**:
- 高可用，Master 集群
- 高并发，多 Worker 并行
- 复杂工作流，DAG 编排
- 企业级，监控告警

**典型用途**:
- 大规模 CI/CD
- 批量数据处理
- 微服务部署
- 复杂工作流

#### AI 驱动团队

**推荐场景**: AI 助手集成

**原因**:
- 自然语言交互
- 移动端支持
- 智能意图解析
- 双向通信

**典型用途**:
- 移动办公
- 语音指令
- 智能助手
- 任务自动化

---

## ❓ 常见问题

### 场景选择疑问

#### Q1: 我应该选择哪个场景？

**A**: 根据团队规模和需求选择：

- **个人开发** → 单机任务自动化
- **2-10 人团队** → 团队协作
- **10+ 人团队** → 分布式任务调度
- **需要 AI 助手** → AI 助手集成（可叠加其他场景）
- **企业定制** → 开发环境定制

#### Q2: 可以从单机升级到分布式吗？

**A**: 可以！AgentFlow 支持平滑升级：

```bash
# 1. 备份数据库
sqlite3 agentflow.db .dump > backup.sql

# 2. 导出数据到 PostgreSQL
sqlite3 agentflow.db .dump | psql postgresql://user:pass@host/db

# 3. 更新配置
cat > ~/.agentflow/config.toml << EOF
[database]
url = "postgresql://user:pass@host/db"
EOF

# 4. 重启服务
agentflow server
```

#### Q3: 可以同时使用多个场景吗？

**A**: 可以！例如：

- **团队协作 + AI 助手集成**: 团队成员可以通过智谱清言下达任务
- **分布式 + 开发定制**: 在分布式集群中运行自定义插件
- **单机 + AI 助手**: 个人开发者也可以使用移动端控制

### 配置问题

#### Q4: 端口 6767 被占用怎么办？

**A**: 修改配置文件：

```toml
[server]
port = 8080  # 或其他可用端口
```

或使用环境变量：

```bash
AGENTFLOW_SERVER_PORT=8080 agentflow server
```

#### Q5: 如何查看日志？

**A**: 根据部署方式查看：

```bash
# 直接运行
agentflow server --log-level debug

# systemd
sudo journalctl -u agentflow -f

# Docker
docker logs -f agentflow

# 指定日志文件
agentflow server --log-file /var/log/agentflow/app.log
```

#### Q6: 数据库连接失败？

**A**: 检查以下几点：

```bash
# 1. 验证数据库 URL
echo $DATABASE_URL

# 2. 测试连接
psql $DATABASE_URL  # PostgreSQL
sqlite3 agentflow.db "SELECT 1"  # SQLite

# 3. 检查防火墙
sudo ufw status

# 4. 查看详细错误
RUST_LOG=debug agentflow server
```

### 故障排除

#### Q7: 任务卡住不执行？

**A**: 检查任务队列：

```bash
# 查看队列状态
curl http://localhost:6767/api/v1/queue/status

# 查看进行中的任务
curl http://localhost:6767/api/v1/tasks?status=running

# 取消卡住的任务
curl -X DELETE http://localhost:6767/api/v1/tasks/{task_id}

# 重启 Executor
curl -X POST http://localhost:6767/api/v1/executor/restart
```

#### Q8: Worker 无法注册到 Master？

**A**: 检查网络连接：

```bash
# 1. 测试 Master 可达性
curl http://master-host:6767/health

# 2. 检查 Worker 日志
tail -f /var/log/agentflow-worker.log

# 3. 手动注册
curl -X POST http://master-host:6767/api/v1/workers/register \
  -H "Content-Type: application/json" \
  -d @worker-config.json
```

#### Q9: Webhook 接收不到请求？

**A**: 检查配置和网络：

```bash
# 1. 验证 Webhook 已启用
curl http://localhost:6767/api/v1/config | jq .webhook

# 2. 检查公网 URL
curl https://your-public-url.com/health

# 3. 测试 Webhook 端点
curl -X POST http://localhost:6767/api/v1/webhook \
  -H "Content-Type: application/json" \
  -d '{"event": "test", "data": {}}'

# 4. 查看日志
tail -f /var/log/agentflow/webhook.log
```

#### Q10: 记忆系统不工作？

**A**: 检查记忆配置：

```bash
# 1. 查看记忆配置
curl http://localhost:6767/api/v1/config | jq .memory

# 2. 测试记忆存储
curl -X POST http://localhost:6767/api/v1/memory \
  -H "Content-Type: application/json" \
  -d '{"content": "Test memory"}'

# 3. 搜索记忆
curl http://localhost:6767/api/v1/memory/search?query=test

# 4. 清除缓存
curl -X DELETE http://localhost:6767/api/v1/memory/cache
```

---

## 📚 进阶资源

### 相关文档

#### 快速开始
- **[README.md](../README.md)** - 项目主文档
- **[RUST_V3_QUICKSTART.md](../RUST_V3_QUICKSTART.md)** - Rust v3 快速开始
- **[DISTRIBUTED_QUICK_START.md](./DISTRIBUTED_QUICK_START.md)** - 分布式快速开始

#### 配置指南
- **[CONFIGURATION.md](./CONFIGURATION.md)** - 完整配置参考
- **[AGENTFLOW_PROJECT_CONFIG.md](./AGENTFLOW_PROJECT_CONFIG.md)** - 项目配置指南
- **[WEBHOOK_QUICK_START.md](./WEBHOOK_QUICK_START.md)** - Webhook 快速开始

#### 集成指南
- **[ZHIPU_INTEGRATION.md](./ZHIPU_INTEGRATION.md)** - 智谱清言集成
- **[MEMORY_WORKFLOW_DESIGN.md](./MEMORY_WORKFLOW_DESIGN.md)** - 记忆系统设计
- **[EXECUTOR_EXAMPLES.md](./EXECUTOR_EXAMPLES.md)** - 执行器示例

#### 技术细节
- **[DISTRIBUTED_EXECUTION_SYSTEM.md](./DISTRIBUTED_EXECUTION_SYSTEM.md)** - 分布式系统架构
- **[EXECUTOR_QUICK_REFERENCE.md](./EXECUTOR_QUICK_REFERENCE.md)** - 执行器 API 参考
- **[API.md](../rust/agentflow-master/API.md)** - REST API 文档

#### 版本规划
- **[VERSION_ROADMAP.md](./VERSION_ROADMAP.md)** - 版本路线图
- **[v0.3.0-index.md](./v0.3.0-index.md)** - v0.3.0 规划索引

### API 文档

#### REST API

完整的 REST API 文档：

```bash
# 启动服务器后访问
open http://localhost:6767/docs

# 或查看离线文档
cat rust/agentflow-master/API.md
```

#### WebSocket API

实时通信接口：

```javascript
// 连接 WebSocket
const ws = new WebSocket('ws://localhost:6767/api/v1/ws');

// 监听任务更新
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Task update:', data);
};

// 发送命令
ws.send(JSON.stringify({
  action: 'subscribe',
  channel: 'tasks'
}));
```

### 社区资源

#### GitHub

- **仓库地址**: https://github.com/MoSiYuan/AgentFlow
- **问题反馈**: https://github.com/MoSiYuan/AgentFlow/issues
- **功能请求**: https://github.com/MoSiYuan/AgentFlow/discussions

#### 示例项目

- **CI/CD 示例**: `examples/ci-cd/`
- **Webhook 示例**: `examples/webhook/`
- **插件开发**: `examples/plugins/`
- **部署脚本**: `scripts/`

### 学习路径

#### 初级（1-2 周）

1. ✅ 阅读 README.md，了解项目概况
2. ✅ 完成快速开始（5 分钟）
3. ✅ 配置单机模式
4. ✅ 创建第一个任务
5. ✅ 理解记忆系统

#### 中级（3-4 周）

1. ✅ 配置团队协作模式
2. ✅ 学习 API 使用
3. ✅ 集成 Webhook
4. ✅ 配置 CI/CD
5. ✅ 自定义 AGENTFLOW.md

#### 高级（5-8 周）

1. ✅ 部署分布式集群
2. ✅ 开发自定义插件
3. ✅ 集成监控系统
4. ✅ 优化性能
5. ✅ 企业级定制

---

## 🎯 总结

AgentFlow 是一个功能强大、灵活可扩展的 AI Agent 任务协作系统，支持从个人开发者到企业级团队的各种场景。

### 核心优势

- 🚀 **简单** - 5 分钟快速上手
- 🔒 **安全** - 沙箱隔离、权限控制
- 📈 **可扩展** - 插件系统、自定义开发
- 🤖 **AI 驱动** - Claude CLI、智谱清言集成
- 🌐 **分布式** - Master 集群、Worker 并行
- 💾 **智能记忆** - 自动学习、经验复用

### 下一步

1. **立即开始**: 按照 [快速开始](#-快速开始) 部署你的第一个 AgentFlow
2. **选择场景**: 根据 [场景对比表](#-场景对比表) 选择适合你的场景
3. **深入阅读**: 查看 [相关文档](#相关文档) 了解更多细节
4. **社区交流**: 加入 GitHub Discussion 与其他用户交流

### 获取帮助

- 📖 **文档**: 查看本文档和相关技术文档
- 🐛 **Bug**: 在 GitHub 提交 Issue
- 💡 **建议**: 在 GitHub Discussion 提出想法
- 📧 **邮件**: support@agentflow.example.com

---

**Happy Flow with AgentFlow!** 🎉

如有任何问题或建议，欢迎随时联系！
