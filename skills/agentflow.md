# AgentFlow Skill

AgentFlow 是一个 AI Agent 任务协作系统，可以将复杂的开发任务分解为多个子任务，并通过 Worker 并发执行。

## 命令格式

### 简化命令（节省 Token）

```bash
# af 是 agentflow 的别名
af init <db>              # 初始化数据库
af master                 # 启动 Master
af add "T" --d "s:T:t:1"  # 添加任务（最简 16 token）
af list --s completed     # 查看已完成任务
af workers                # 查看 Workers
```

### 任务类型

```bash
# 1. AI 智能任务（自动分解）
af add "实现用户登录" --desc "task:implement:login"
# 自动分解为：
# - 设计数据模型
# - 实现 API 端点
# - 编写测试
# - 更新文档

# 2. Shell 命令
af add "运行测试" --desc "shell:go test ./..."

# 3. 脚本执行
af add "部署应用" --desc "script:./deploy.sh"

# 4. 文件操作
af add "写配置" --desc "file:write:config.yaml:key:value"

# 5. 故事生成（演示）
af add "生成故事" --desc "write_story:标题:type:001"

# 6. 评审任务
af add "评审故事" --desc "review_story:任务ID:标题:type"
```

## 典型工作流

### 1. 本地开发模式

```bash
# Terminal 1: 启动 Master
cd /Users/jiangxiaolong/work/project/AgentFlow
make build
./bin/agentflow init dev.db
./bin/agentflow master --db dev.db

# Terminal 2: 添加开发任务
./bin/agentflow add "添加用户认证" --desc "task:implement:auth" --db dev.db
./bin/agentflow add "编写测试" --desc "shell:go test ./..." --db dev.db
./bin/agentflow add "代码检查" --desc "shell:golangci-lint run" --db dev.db

# 查看进度
./bin/agentflow list --status running --db dev.db
./bin/agentflow list --status completed --db dev.db
```

### 2. 批量任务创建

```bash
# 一次性创建多个相关任务
af add "格式化代码" --d "shell:gofmt -w ."
af add "运行测试" --d "shell:go test ./..."
af add "构建应用" --d "shell:go build -v ./..."
af add "启动服务" --d "shell:./bin/agentflow master"

# 查看所有任务
af list
```

### 3. 任务链（任务升级）

```bash
# 主任务会自动创建子任务
af add "开发新功能" --desc "task:implement:feature_name"

# 等待所有子任务完成
af list --status completed
```

## 与 Claude Code 集成

### SDK 方式

```go
// 直接在代码中调用
exec.Command("agentflow", "add", "测试", "--desc", "shell:go test").Run()
```

### 快速命令

```bash
# 1. 创建任务（16 token）
af add "T" --d "s:T:t:1"

# 2. 查询状态（12 token）
af list --s done

# 3. 查看 Workers（8 token）
af workers
```

## 常用场景

### 场景 1: 快速迭代开发

```bash
# 创建开发循环任务
af add "格式化" --d "shell:gofmt -w ."
af add "测试" --d "shell:go test ./..."
af add "构建" --d "shell:go build"
af add "重启" --d "shell:killall agentflow; ./bin/agentflow master"
```

### 场景 2: 并行测试

```bash
# 创建多个并行测试任务
af add "单元测试" --d "shell:go test ./internal/..." -g linux
af add "集成测试" --d "shell:go test ./tests/integration/..." -g linux
af add "E2E 测试" --d "shell:go test ./tests/e2e/..." -g linux
```

### 场景 3: 代码质量检查

```bash
# 多个质量检查并行执行
af add "格式检查" --d "shell:gofmt -l ."
af add "Lint" --d "shell:golangci-lint run"
af add "安全扫描" --d "shell:gosec ./..."
af add "覆盖率" --d "shell:go test -cover ./..."
```

## 调试技巧

```bash
# 查看 Master 日志
tail -f /tmp/agentflow-master.log

# 查看 Worker 日志
tail -f /tmp/agentflow-worker-*.log

# 检查数据库
sqlite3 dev.db "SELECT id, title, status FROM tasks ORDER BY created_at DESC;"

# 测试 API
curl http://localhost:8848/health
curl http://localhost:8848/api/v1/tasks
curl http://localhost:8848/api/v1/workers
```

## 性能优化

```bash
# 1. 批量创建任务
for i in {1..10}; do
  af add "任务$i" --d "shell:echo $i"
done

# 2. 使用 Worker Groups
af add "Linux 任务" --d "shell:make linux" -g linux
af add "Windows 任务" --d "shell:make windows" -g windows

# 3. 并发执行（自动）
af add "测试1" --d "shell:go test ./pkg1"
af add "测试2" --d "shell:go test ./pkg2"
af add "测试3" --d "shell:go test ./pkg3"
# 三个任务会并发执行
```

## 命令别名配置

在 `~/.bashrc` 或 `~/.zshrc` 中添加：

```bash
# AgentFlow 别名
alias af='agentflow'
alias afi='agentflow init'
alias afm='agentflow master'
def af-add() { agentflow add "$1" --desc "$2"; }
alias afl='agentflow list'
alias afc='agentflow list --status completed'
alias afg='agentflow list --status running'
```

## 故障排查

### 任务卡住不动

```bash
# 1. 检查 Master 状态
curl http://localhost:8848/health

# 2. 检查 Worker 连接
af workers

# 3. 重启卡住的任务
af list --status running
# 找到 task_id，然后：
curl -X POST http://localhost:8848/api/v1/tasks/{task_id}/unlock
```

### Worker 未注册

```bash
# 1. 检查 Worker 日志
tail -f /tmp/agentflow-worker-*.log

# 2. 手动启动 Worker
af worker --group default

# 3. 检查网络连接
ping localhost
curl http://localhost:8848/api/v1/workers
```

## 快速参考

| 命令 | 说明 | Token |
|------|------|-------|
| `af add "T" --d "s:T:t:1"` | 添加任务 | 16 |
| `af list --s done` | 查看已完成 | 12 |
| `af workers` | 查看 Workers | 8 |
| `af init db` | 初始化 | 10 |
| `af master` | 启动 Master | 8 |

## 示例：完整开发流程

```bash
# 1. 初始化
af init dev.db

# 2. 启动 Master（后台）
af master --db dev.db &

# 3. 创建开发任务
af add "格式化" --d "shell:gofmt -w ."
af add "测试" --d "shell:go test ./..."
af add "构建" --d "shell:go build"

# 4. 监控进度
watch -n 2 'af list --s completed | wc -l'

# 5. 查看结果
af list --s completed
af list --s failed
```
