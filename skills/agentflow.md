# AgentFlow - AI Agent Task Collaboration System

> **Master-Worker 架构的异步任务协作系统** | 真正多进程并发 | 4-5秒/任务

## 📚 相关文档

- 📖 [完整文档](README.md) - 架构设计、性能指标、使用场景
- 📦 [安装指南](INSTALL.md) - 3种安装方式、配置说明、故障排查
- 🔧 [本 Skill 手册](skills/agentflow.md) - 快速调用、Token 优化、实战示例

## 🚀 快速启动（3命令）

```bash
# Terminal 1: 启动 Master
cd /Users/jiangxiaolong/work/project/AgentFlow
./bin/master --mode standalone --port 8848

# Terminal 2: 启动 Worker
./bin/worker --mode standalone --master http://127.0.0.1:8848 --name w1 --auto

# Terminal 3: 创建任务
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{"task_id": "T1", "title": "Test", "description": "prompt", "priority": "high"}'
```

**详细安装步骤**: 请参考 [安装指南](INSTALL.md#安装方式)

## 🎯 核心 API（最常用）

```bash
# 健康检查
GET /api/health

# 创建任务
POST /api/tasks/create
{"task_id": "ID", "title": "标题", "description": "shell:命令", "priority": "high"}

# 查询状态
GET /api/status
GET /api/tasks/pending
GET /api/tasks/completed
GET /api/workers
```

**完整 API 文档**: 请参考 [README.md#核心-api](README.md#核心-api)

## 📋 任务格式

```bash
# Shell 命令
shell:echo "Hello World" && date

# AI 任务（自动调用 Claude CLI）
ai:解释什么是 Agent

# 复杂命令
shell:cd /path && make build && ./app

# 多命令流水线
shell:echo "Step 1" && sleep 1 && echo "Step 2"
```

**任务执行优先级**:
1. **HTTP Executor** - 如果 Claude Server 可用（最快）
2. **Claude CLI** - 如果找到 `~/bin/claudecli`（4-5秒）
3. **本地模拟** - 回退模式（<2ms）

**Claude CLI 配置**: 参见 [安装指南 > Claude CLI 配置](INSTALL.md#claude-cli-配置)

## ⚡ 多进程并发

```bash
# 启动 3 个 Worker（真正的并发）
./bin/worker --mode standalone --master http://127.0.0.1:8848 --name w1 --auto &
./bin/worker --mode standalone --master http://127.0.0.1:8848 --name w2 --auto &
./bin/worker --mode standalone --master http://127.0.0.1:8848 --name w3 --auto &

# 创建 5 个任务
for i in 1 2 3 4 5; do
  curl -X POST http://127.0.0.1:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d "{\"task_id\": \"T$i\", \"title\": \"Task $i\", \"description\": \"shell:echo Task $i && date\", \"priority\": \"high\"}"
done

# 查看 3 个 Worker 并发执行
curl http://127.0.0.1:8848/api/status | jq '.data'
```

**性能特点**:
- ✅ 真正多进程（独立 PID，非多线程）
- ✅ 并发执行（3 Workers ≈ 0.6 任务/秒）
- ✅ 任务自动分配（Master 调度）
- ✅ 吞吐量线性扩展

## 💰 上下文优化（节省 Token）

### 技巧 1: 使用 Task ID 引用

```bash
# ❌ 不好：每次都传递完整上下文
curl -X POST /api/tasks/create \
  -d '{"description": "在前面的任务基础上，继续优化代码..."}'

# ✅ 好：使用 task_id 关联
curl -X POST /api/tasks/create \
  -d '{"task_id": "OPTIMIZE-1", "dependencies": ["BUILD-1", "TEST-1"]}'
```

**节省**: ~500-1000 tokens/任务

### 技巧 2: 批量创建（减少往返）

```bash
# ❌ 不好：单个创建（多次上下文切换）
curl -X POST /api/tasks/create -d '{"task_id": "T1", ...}'
curl -X POST /api/tasks/create -d '{"task_id": "T2", ...}'
curl -X POST /api/tasks/create -d '{"task_id": "T3", ...}'

# ✅ 好：批量创建（一次性提交）
for i in 1 2 3; do
  curl -X POST http://127.0.0.1:8848/api/tasks/create \
    -d "{\"task_id\": \"T$i\", \"description\": \"shell:echo $i\"}"
done
```

**节省**: ~300 tokens/批次

### 技巧 3: 使用 Skill 快捷命令

```bash
# ✅ 使用本 skill（直接集成上下文）
/agentflow status          # 查看 Workers
/agentflow workers         # 查看任务状态
/agentflow list            # 列出所有任务

# ✅ 使用快速脚本
bash /Users/jiangxiaolong/work/project/AgentFlow/quick-task.sh "标题" "shell:命令"
```

**节省**: ~200-400 tokens/命令

### 技巧 4: 描述简洁化

```bash
# ❌ 不好：冗长描述
"description": "请先切换到项目目录，然后执行 go build 命令编译整个项目，最后运行生成的二进制文件"

# ✅ 好：简洁命令
"description": "shell:cd /path && go build && ./app"
```

**节省**: ~100-200 tokens/任务

## 🛠️ 快速脚本

### 创建任务脚本

[quick-task.sh](quick-task.sh) - 快速创建单个任务

```bash
#!/bin/bash
# quick-task.sh - 快速创建任务
TITLE="$1"
DESC="$2"
TASK_ID="TASK-$(date +%s)"

curl -s -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d "{
    \"task_id\": \"$TASK_ID\",
    \"title\": \"$TITLE\",
    \"description\": \"$DESC\",
    \"priority\": \"high\"
  }" | jq '.'
```

**使用示例**:
```bash
bash quick-task.sh "测试任务" "shell:echo Hello World"
```

### 批量任务脚本

```bash
#!/bin/bash
# batch-tasks.sh - 批量创建任务
for cmd in "$@"; do
  bash /Users/jiangxiaolong/work/project/AgentFlow/quick-task.sh "Execute: $cmd" "shell:$cmd"
done
```

**使用示例**:
```bash
bash batch-tasks.sh "echo Task1" "echo Task2" "echo Task3"
```

## 🤖 Claude CLI 集成

### 自动检测机制

Worker 自动按以下顺序查找 `claudecli`:
1. `~/bin/claudecli` (推荐位置)
2. `$PATH` 中的 `claudecli`
3. 如果未找到 → 回退到模拟模式

### Wrapper 脚本安装

**自动安装** (推荐):
```bash
# 安装指南包含完整的 wrapper 脚本
# 参见: INSTALL.md > Claude CLI 配置
```

**手动验证**:
```bash
# 1. 检查 wrapper
ls -l ~/bin/claudecli

# 2. 测试执行
export PATH="$HOME/bin:$PATH"
claudecli chat --prompt "测试" --no-interactive

# 3. 检查 Claude CLI
which claude
claude --version  # 应显示: 1.0.102+
```

### 执行优先级

1. **HTTP Executor** - Claude Server 可用时（最快）
2. **Claude CLI** - 使用 `claudecli` wrapper（4-5秒）
3. **本地模拟** - 无 Claude CLI 时（<2ms）

**详细配置**: 参见 [安装指南 > Claude CLI 配置](INSTALL.md#claude-cli-配置)

## 🔍 调试

### 查看日志

```bash
# Worker 日志
cat /tmp/worker-*.log | tail -20

# Master 日志（如果运行在 terminal）
# 查看 Master 的标准输出
```

### 测试 claudecli

```bash
# 1. 设置 PATH
export PATH="$HOME/bin:$PATH"

# 2. 测试 wrapper
claudecli chat --prompt "测试" --no-interactive

# 3. 验证输出
# 预期: Claude 的响应内容
# Exit code: 0
```

### 检查进程

```bash
# 查看 Master 和 Worker 进程
ps aux | grep -E 'master|worker' | grep -v grep

# 验证多进程
# 应该看到多个不同的 PID
```

### API 测试

```bash
# 健康检查
curl http://127.0.0.1:8848/api/health

# 查看状态（格式化输出）
curl http://127.0.0.1:8848/api/status | jq '.'

# 查看 Workers
curl http://127.0.0.1:8848/api/workers | jq '.'
```

## 📊 性能指标

| 指标 | 数值 | 说明 |
|------|------|------|
| Worker 进程 | 真正多进程 | 独立 PID，非多线程 |
| 并发执行 | ✅ 已验证 | 3 Workers 并发执行 |
| 内存使用 | ~20MB/进程 | 轻量级设计 |
| 任务完成 | 4-5秒/任务 | 含 Claude CLI 调用 |
| 吞吐量 | ~0.6 任务/秒 | 3 Workers 并发 |
| HTTP 吞吐量 | 10,000+ req/s | Master API 性能 |
| 启动时间 | <100ms | Worker 冷启动 |
| 二进制大小 | 34MB | 静态链接 Go 二进制 |

**更多性能数据**: 参见 [README.md#性能指标](README.md#性能指标)

## 🔧 故障排查

### 问题 1: claudecli 执行失败

**症状**:
```
level=warning msg="claudecli execution failed: exit status 1"
```

**解决方案**:
```bash
# 1. 检查 wrapper 是否存在
ls -l ~/bin/claudecli
cat ~/bin/claudecli

# 2. 手动测试
export PATH="$HOME/bin:$PATH"
claudecli chat --prompt "hi" --no-interactive

# 3. 检查 Claude CLI
which claude
claude --version

# 4. 重新安装 wrapper
# 参见: INSTALL.md > Claude CLI 配置 > 手动安装
```

### 问题 2: Worker 竞态条件（正常行为）

**症状**:
```
Error: failed to assign task: task not found or not pending
```

**说明**: 这说明多个 Worker 在并发抢任务，证明系统真正并发执行 ✅

### 问题 3: Master 无法启动

**症状**:
```
listen tcp 0.0.0.0:8848: bind: address already in use
```

**解决方案**:
```bash
# 1. 检查端口占用
lsof -i:8848

# 2. 杀掉旧进程
kill -9 $(lsof -ti:8848)

# 3. 重新启动
./bin/master --mode standalone --port 8848
```

**更多故障排查**: 参见 [安装指南 > 故障排查](INSTALL.md#故障排查)

## 📝 完整示例

### 场景：3 Workers 并发执行 5 个任务

```bash
# 1. 启动 Master
./bin/master --mode standalone --port 8848 &
MASTER_PID=$!

# 2. 启动 3 个 Worker
for i in 1 2 3; do
  ./bin/worker --mode standalone --master http://127.0.0.1:8848 \
    --name "worker-$i" --auto > /tmp/worker-$i.log 2>&1 &
done

# 3. 创建 5 个并发任务
for i in 1 2 3 4 5; do
  curl -s -X POST http://127.0.0.1:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d "{
      \"task_id\": \"TASK-$i\",
      \"title\": \"并发任务 $i\",
      \"description\": \"shell:echo 'Worker executing task $i' && date\",
      \"priority\": \"high\"
    }"
done

# 4. 等待完成
sleep 5

# 5. 查看结果
curl -s http://127.0.0.1:8848/api/status | jq '.data'
curl -s http://127.0.0.1:8848/api/tasks/completed | jq '.data.tasks | length'

# 6. 清理
kill $MASTER_PID
```

---

**项目位置**: `/Users/jiangxiaolong/work/project/AgentFlow`
**当前分支**: `feature/1.0.0`
**版本**: v1.0.0

**相关文档**:
- 📖 [README.md](README.md) - 完整项目文档
- 📦 [INSTALL.md](INSTALL.md) - 安装指南
- 🏗️ [ARCHITECTURE.md](docs/ARCHITECTURE.md) - 架构设计

**已验证特性**: ✅ 多进程并发执行 | ✅ Claude CLI 集成 | ✅ 任务自动分配 | ✅ 真实 AI 执行
