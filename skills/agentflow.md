# AgentFlow - AI Agent Task Collaboration System

## 快速启动（3命令）

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

## 核心 API（最常用）

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

## 任务格式

```bash
# Shell 命令
shell:echo "Hello World" && date

# AI 任务（Claude CLI）
ai:解释什么是Agent

# 复杂命令
shell:cd /path && make build && ./app

# 多命令
shell:echo "Step 1" && sleep 1 && echo "Step 2"
```

## 多进程并发

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

## 上下文优化（节省 Token）

### 1. 使用 Task ID 引用
```bash
# ❌ 不好：每次都传递完整上下文
curl -X POST /api/tasks/create \
  -d '{"description": "在前面的任务基础上，继续优化代码..."}'

# ✅ 好：使用 task_id 关联
curl -X POST /api/tasks/create \
  -d '{"task_id": "OPTIMIZE-1", "dependencies": ["BUILD-1", "TEST-1"]}'
```

### 2. 批量创建（减少往返）
```bash
# ❌ 不好：单个创建
curl -X POST /api/tasks/create -d '{"task_id": "T1", ...}'
curl -X POST /api/tasks/create -d '{"task_id": "T2", ...}'
curl -X POST /api/tasks/create -d '{"task_id": "T3", ...}'

# ✅ 好：批量创建
for i in 1 2 3; do
  curl -X POST http://127.0.0.1:8848/api/tasks/create \
    -d "{\"task_id\": \"T$i\", \"description\": \"shell:echo $i\"}"
done
```

### 3. 使用 skill 命令（本文件）
```bash
# 直接使用 /agentflow 命令
/agentflow add "测试" --desc "shell:go test ./..."

# 或调用脚本
bash /Users/jiangxiaolong/work/project/AgentFlow/quick-task.sh "标题" "shell:命令"
```

## 快速脚本

### 创建任务脚本
```bash
#!/bin/bash
# quick-task.sh - 快速创建任务
TITLE="$1"
DESC="$2"
curl -s -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d "{\"task_id\": \"TASK-$(date +%s)\", \"title\": \"$TITLE\", \"description\": \"$DESC\", \"priority\": \"high\"}"
```

### 批量任务脚本
```bash
#!/bin/bash
# batch-tasks.sh - 批量创建任务
for cmd in "$@"; do
  /agentflow add "Execute: $cmd" --desc "shell:$cmd"
done
```

## Claude CLI 集成

### 自动检测
- Worker 自动查找 `~/bin/claudecli`
- 如果找到，使用 Claude CLI 执行 AI 任务
- 如果未找到，回退到模拟模式

### Wrapper 脚本
```bash
# ~/bin/claudecli（自动安装）
#!/bin/bash
# 将旧 API 转换为新的 Claude CLI 格式
exec /opt/homebrew/bin/claude -p "$@"
```

## 调试

```bash
# 查看 Worker 日志
cat /tmp/worker-*.log | tail -20

# 测试 claudecli
export PATH="$HOME/bin:$PATH"
claudecli chat --prompt "测试" --no-interactive

# 检查进程
ps aux | grep -E 'master|worker' | grep -v grep

# API 测试
curl http://127.0.0.1:8848/api/status | jq '.'
```

## 性能指标

| 指标 | 数值 |
|------|------|
| Worker 进程 | 真正多进程（独立 PID） |
| 并发执行 | ✅ 已验证 |
| 内存使用 | ~20MB/进程 |
| 任务完成 | 4-5秒/任务（含 Claude） |
| 吞吐量 | 3 Workers = ~0.6 任务/秒 |

## 故障排查

### claudecli 执行失败
```bash
# 检查 wrapper
ls -l ~/bin/claudecli
cat ~/bin/claudecli

# 手动测试
export PATH="$HOME/bin:$PATH"
claudecli chat --prompt "hi" --no-interactive
```

### Worker 竞态条件
```
Error: failed to assign task: task not found or not pending
```
这是正常的！说明多个 Worker 在并发抢任务。

### Master 无法启动
```bash
# 检查端口
lsof -i:8848

# 杀掉旧进程
kill -9 $(lsof -ti:8848)

# 重新启动
./bin/master --mode standalone --port 8848
```

## 完整示例

```bash
# 1. 启动 Master
./bin/master --mode standalone --port 8848 &
MASTER_PID=$!

# 2. 启动 3 个 Worker
for i in 1 2 3; do
  ./bin/worker --mode standalone --master http://127.0.0.1:8848 \
    --name "worker-$i" --oneshot --auto > /tmp/worker-$i.log 2>&1 &
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

**已验证**: ✅ 多进程并发执行、Claude CLI 集成、任务自动分配
**位置**: `/Users/jiangxiaolong/work/project/AgentFlow`
**文档`: [README.md](README.md), [ARCHITECTURE.md](docs/ARCHITECTURE.md)
