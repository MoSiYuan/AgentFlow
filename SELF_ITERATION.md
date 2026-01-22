# AgentFlow 自迭代开发指南

本指南说明如何使用 AgentFlow 来开发 AgentFlow 自身，实现"用自己开发自己"的元循环。

## 核心理念

```
AgentFlow (当前版本)
    ↓
创建开发任务
    ↓
AgentFlow Workers 执行任务
    ↓
生成新版本代码
    ↓
编译和测试
    ↓
部署新版本
    ↓
AgentFlow (新版本) ← 循环
```

## 快速开始

### 步骤 1: 启动开发环境

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow

# 使用纯标准库版本（无需依赖）
./tests/ctest_pure.go

# 或者完整版（需要先编译）
make build
./bin/agentflow init dev.db
./bin/agentflow master --db dev.db
```

### 步骤 2: 创建第一个迭代任务

```bash
# 任务：优化代码结构
./bin/agentflow add "重构数据库层" \
  --desc "task:refactor:database_layer" \
  --db dev.db

# 任务：添加新功能
./bin/agentflow add "添加任务优先级" \
  --desc "task:implement:priority_queue" \
  --db dev.db

# 任务：编写测试
./bin/agentflow add "编写集成测试" \
  --desc "shell:go test ./tests/integration/..." \
  --db dev.db
```

### 步骤 3: 监控任务执行

```bash
# 实时查看任务状态
watch -n 2 './bin/agentflow list --db dev.db'

# 查看已完成任务
./bin/agentflow list --status completed --db dev.db

# 查看 Worker 状态
./bin/agentflow workers --db dev.db
```

## 完整开发工作流

### 场景 1: 功能开发

```bash
# 1. 创建功能开发任务
af add "实现任务依赖" --d "task:implement:task_dependency"
# 自动分解为：
#   - 设计数据模型（添加 parent_id 字段）
#   - 实现 DAG 依赖检查
#   - 实现依赖触发逻辑
#   - 编写单元测试

# 2. 创建代码质量任务
af add "格式化代码" --d "shell:gofmt -w ."
af add "代码检查" --d "shell:golangci-lint run"

# 3. 创建测试任务
af add "单元测试" --d "shell:go test ./internal/... -v"
af add "集成测试" --d "shell:go test ./tests/integration/... -v"

# 4. 创建构建任务
af add "编译" --d "shell:go build -v ./..."

# 5. 查看结果
af list --s completed
```

### 场景 2: Bug 修复

```bash
# 1. 创建诊断任务
af add "诊断 Bug" --d "shell:grep -r 'bug_pattern' ./..."

# 2. 创建修复任务
af add "修复 Bug" --d "task:fix:bug_123"

# 3. 创建验证任务
af add "验证修复" --d "shell:go test -run TestBug123"

# 4. 创建回归测试
af add "回归测试" --d "shell:go test ./... -v"
```

### 场景 3: 性能优化

```bash
# 1. 基准测试
af add "基准测试" --d "shell:go test -bench=. -benchmem"

# 2. 性能分析
af add "CPU 分析" --d "shell:go test -cpuprofile=cpu.prof"
af add "内存分析" --d "shell:go test -memprofile=mem.prof"

# 3. 优化实现
af add "优化数据库查询" --d "task:optimize:database_query"

# 4. 验证优化
af add "对比基准" --d "shell:go test -bench=. -benchmem | tee before_after.txt"
```

## 自动化迭代流程

### 完整的 CI/CD 循环

```bash
#!/bin/bash
# auto-iterate.sh - AgentFlow 自动迭代脚本

set -e

DB_PATH="iteration.db"
BINARY="./bin/agentflow"

echo "🚀 开始 AgentFlow 自迭代开发..."

# 1. 清理旧数据
rm -f $DB_PATH
$BINARY init $DB_PATH

# 2. 启动 Master（后台）
$BINARY master --db $DB_PATH > /tmp/af-master.log 2>&1 &
MASTER_PID=$!
sleep 2

# 3. 创建迭代任务
echo "📝 创建迭代任务..."

# 代码质量
$BINARY add "格式化" --desc "shell:gofmt -w ." --db $DB_PATH
$BINARY add "Lint" --desc "shell:golangci-lint run" --db $DB_PATH

# 测试
$BINARY add "单元测试" --desc "shell:go test ./internal/... -v" --db $DB_PATH
$BINARY add "集成测试" --desc "shell:go test ./tests/... -v" --db $DB_PATH

# 构建
$BINARY add "编译" --desc "shell:make build" --db $DB_PATH

# 4. 等待所有任务完成
echo "⏳ 等待任务执行..."
while true; do
    COMPLETED=$($BINARY list --status completed --db $DB_PATH | grep -c "completed" || true)
    RUNNING=$($BINARY list --status running --db $DB_PATH | grep -c "running" || true)
    PENDING=$($BINARY list --status pending --db $DB_PATH | grep -c "pending" || true)

    echo "进度: 完成=$COMPLETED, 运行=$RUNNING, 待处理=$PENDING"

    if [ $PENDING -eq 0 ] && [ $RUNNING -eq 0 ]; then
        break
    fi

    sleep 2
done

# 5. 显示结果
echo "✅ 迭代完成！"
$BINARY list --db $DB_PATH

# 6. 清理
kill $MASTER_PID 2>/dev/null || true

echo "🎉 AgentFlow 自迭代完成！"
```

使用方法：

```bash
chmod +x auto-iterate.sh
./auto-iterate.sh
```

## 增量开发模式

### 小步快跑

```bash
# 每次只做一件事
af add "小改进:添加日志" --d "task:minor:add_logging"
af add "小改进:优化查询" --d "task:minor:optimize_query"
af add "小改进:更新文档" --d "task:minor:update_docs"

# 等待完成
af list --s completed

# 验证后继续下一批
af add "小改进:添加测试" --d "task:minor:add_tests"
```

### 功能分支

```bash
# 1. 创建分支
git checkout -b feature/priority-queue

# 2. 创建开发任务
af add "实现优先级队列" --d "task:implement:priority_queue"

# 3. 执行开发
# AgentFlow 会自动创建子任务：
#   - 修改数据模型
#   - 实现调度逻辑
#   - 编写测试
#   - 更新文档

# 4. 验证
af list --s completed
go test ./...

# 5. 提交
git add .
git commit -m "feat: 添加任务优先级队列"
git push origin feature/priority-queue
```

## 高级场景

### 场景 1: 多环境部署

```bash
# 开发环境
af add "构建开发版" --d "shell:go build -tags=dev" -g darwin

# 测试环境
af add "构建测试版" --d "shell:GOOS=linux go build" -g linux

# 生产环境
af add "构建生产版" --d "shell:go build -ldflags='-s -w'" -g linux
```

### 场景 2: A/B 测试

```bash
# 版本 A
af add "构建版本A" --d "shell:go build -o bin/agentflow-a"
af add "测试版本A" --d "shell:./bin/agentflow-a test"

# 版本 B
af add "构建版本B" --d "shell:go build -o bin/agentflow-b"
af add "测试版本B" --d "shell:./bin/agentflow-b test"

# 对比结果
af add "对比性能" --d "shell:hyperfine ./bin/agentflow-a ./bin/agentflow-b"
```

### 场景 3: 紧急修复

```bash
# 1. 热修复流程
af add "创建修复分支" --d "shell:git checkout -b hotfix/fix-bug"

# 2. 快速修复
af add "修复 Bug" --d "task:hotfix:critical_bug"

# 3. 验证
af add "快速验证" --d "shell:go test -run TestBugFix"

# 4. 打标签
af add "打标签" --d "shell:git tag v1.0.1"

# 5. 部署
af add "部署修复" --d "shell:make deploy"
```

## 元循环示例

### 用 AgentFlow 开发 AgentFlow

```bash
# 当前版本的 AgentFlow
CURRENT_VERSION="v1.0.0"

# 1. 创建"开发新功能"任务
af add "开发 v2.0.0" --d "task:major:version_2"

# 这个任务会分解为：
#   - 设计新架构
#   - 实现新功能
#   - 编写测试
#   - 更新文档

# 2. 使用 v1.0.0 开发 v2.0.0
af add "编译 v2.0.0" --d "shell:go build -o bin/agentflow-v2"

# 3. 用 v2.0.0 开发 v2.1.0
./bin/agentflow-v2 init dev2.db
./bin/agentflow-v2 master --db dev2.db &
af-v2 add "开发 v2.1.0" --d "task:minor:version_2.1"

# 4. 无限循环...
```

## 监控和调试

### 查看迭代历史

```bash
# 数据库查询
sqlite3 dev.db <<EOF
SELECT
    datetime(created_at, 'localtime') as time,
    title,
    status
FROM tasks
ORDER BY created_at DESC
LIMIT 20;
EOF

# 统计成功率
sqlite3 dev.db <<EOF
SELECT
    status,
    COUNT(*) as count
FROM tasks
GROUP BY status;
EOF
```

### 性能分析

```bash
# 任务执行时间
sqlite3 dev.db <<EOF
SELECT
    title,
    julianday(completed_at) - julianday(started_at) as duration_seconds
FROM tasks
WHERE status = 'completed'
ORDER BY duration_seconds DESC
LIMIT 10;
EOF
```

## 最佳实践

### 1. 任务分解原则

- 单一职责：每个任务只做一件事
- 可测试：任务结果可验证
- 幂等性：重复执行结果一致
- 超时控制：避免任务卡住

### 2. 版本管理

```bash
# 每次迭代打标签
af add "打标签" --d "shell:git tag -a v\$(date +%Y.%m.%d) -m 'Auto tag'"

# 自动生成 CHANGELOG
af add "生成日志" --d "shell:git log --pretty=format:'- %s' > CHANGELOG.md"
```

### 3. 回滚策略

```bash
# 保留历史版本
af add "备份版本" --d "shell:cp bin/agentflow bin/agentflow.backup"

# 验证失败则回滚
af add "验证新版本" --d "shell:./bin/agentflow test || cp bin/agentflow.backup bin/agentflow"
```

## 故障排查

### 任务失败处理

```bash
# 1. 查看失败任务
af list --s failed

# 2. 查看错误信息
sqlite3 dev.db "SELECT title, error FROM tasks WHERE status='failed'"

# 3. 重试失败任务
FAILED_TASKS=$(sqlite3 dev.db "SELECT id FROM tasks WHERE status='failed'")
for id in $FAILED_TASKS; do
    curl -X POST http://localhost:8848/api/v1/tasks/$id/reset
done
```

### Worker 故障处理

```bash
# 1. 检查 Worker 状态
af workers

# 2. 重启异常 Workers
pkill -9 agentflow-worker
af worker --group default

# 3. 查看日志
tail -f /tmp/agentflow-*.log
```

## 总结

通过 AgentFlow 自迭代开发，你可以：

1. **自动化开发流程**：从编码到测试到部署全自动化
2. **并行执行任务**：充分利用多核 CPU 和多 Worker
3. **版本演进**：用 v1 开发 v2，用 v2 开发 v3
4. **持续改进**：每次迭代都是基于实际需求
5. **元循环**：工具自我演进，形成正反馈

记住：**最好的工具是能够自我改进的工具！**
