# AgentFlow for Claude Code - 使用指南

> 让 Claude Code 更高效地使用 AgentFlow skill

## 🎯 最简流程

```bash
# 1. 运行测试（3秒，完成30个任务）
cd tests && go run ctest_pure.go

# 2. 查看结果
cat ctest_stories/story_1.md
```

## 📝 任务模板

### 故事生成
```
write_story:标题:类型:索引
```

**示例**:
```bash
agentflow add "深海探险" --desc "write_story:深海探险:cthulhu:001"
```

### 评审任务
```
review_story:任务ID:标题:review
```

**示例**:
```bash
agentflow add "评审:深海探险" --desc "review_story:1:深海探险:review"
```

### AI 子任务
```
task:implement:功能名
task:test:模块名
task:build:项目名
```

## 🔧 常用场景

### 1. 文档生成

```bash
# 创建文档任务
agentflow add "API文档" \
  --desc "write_doc:api:spec" \
  --db agentflow.db

# 后续自动创建
# - 评审任务
# - 格式化任务
# - 发布任务
```

### 2. 代码开发

```bash
# 开发任务（自动分解）
agentflow add "用户认证" \
  --desc "task:implement:auth" \
  --db agentflow.db

# 自动生成子任务：
# - 设计数据模型
# - 实现 API
# - 编写测试
# - 更新文档
```

### 3. 测试执行

```bash
# 批量测试
agentflow add "单元测试" --desc "shell:go test ./..." --db agentflow.db
agentflow add "集成测试" --desc "shell:go test ./tests/integration/..." --db agentflow.db
agentflow add "E2E测试" --desc "script:./e2e.sh" --db agentflow.db
```

### 4. 代码审查

```bash
# 创建审查任务
agentflow add "审查 PR #123" \
  --desc "review_pr:123:feature-branch" \
  --db agentflow.db
```

## 💡 Token 优化

### 最短命令格式

| 操作 | 常规格式 | Token优化 |
|------|---------|-----------|
| 初始化 | `agentflow init agentflow.db` | ∼ 原样 |
| 创建任务 | `agentflow add "标题" --desc "write_story:T:t:1"` | 可缩写描述 |
| 查询全部 | `agentflow list` | ∼ 原样 |
| 查询完成 | `agentflow list --status completed` | 可缩写为 `--s` |

### 批量操作

```bash
# 一次性创建10个任务
for i in {1..10}; do
  agentflow add "任务$i" --desc "task:test:module$i"
done

# 查询一次
agentflow list --status completed > results.txt
```

### 环境变量

```bash
export AgentFlow_DB=agentflow.db
export MASTER_URL=http://localhost:6767
export WORKER_GROUP=local

# 使用环境变量
agentflow add "测试" --db "$AgentFlow_DB"  # 无需每次指定
```

## 🤖 Claude Code 技巧

### 技巧 1: 快速测试

```typescript
// 在 Claude Code 中
await exec("cd tests && go run ctest_pure.go")
```

### 技巧 2: 读取结果

```typescript
// 查看特定故事
const story = await fs.readFile("tests/ctest_stories/story_1.md", "utf-8")
console.log(story)
```

### 技巧 3: 批量创建

```typescript
// 批量创建任务
for (let i = 1; i <= 10; i++) {
  await exec(`agentflow add "任务${i}" --desc "write_story:任务${i}:t:00${i}"`)
}
```

### 技巧 4: 状态监控

```typescript
// 等待完成
while (true) {
  const result = await exec("agentflow list --status running")
  if (!result.includes("▶️")) break  // 无运行中任务
  await sleep(2000)
}
```

## 🎨 自定义内容

### 修改故事模板

编辑 `internal/worker/ai_worker.go`:

```go
// 在 generateCthulhuStory 函数中
func generateCthulhuStory(title, storyType, workerID string) string {
    // 自定义你的内容生成逻辑
    return fmt.Sprintf("# %s\n\n自定义内容...", title)
}
```

### 添加新任务类型

```go
case "write_doc":
    return w.executeDocTask(ctx, task, params)

case "review_pr":
    return w.executePRReview(ctx, task, params)
```

## 📊 快速参考

### 任务状态

- `pending` - 待执行
- `running` - 执行中
- `completed` - 已完成
- `failed` - 失败

### Worker 状态

- `active` - 活跃
- `inactive` - 非活跃（超时）

### 命令别名（可自己设置）

```bash
alias af='agentflow'
alias afl='af list --s completed'
alias afw='af workers'
alias afa='af add'
```

## 🚨 故障排除

### 任务不执行

```bash
# 检查 Master 是否运行
curl http://localhost:6767/health

# 查看 Worker 状态
agentflow workers

# 查看 pending 任务
agentflow list --status pending
```

### 文件未生成

```bash
# 检查 workspace
ls ~/agentflow-workspace/

# 查看日志
tail -f /tmp/cpds_master.log
```

### Token 使用建议

1. **初始化阶段**: 只执行一次（~5 token）
2. **创建任务**: 每个任务 ~10-20 token
3. **查询状态**: 每次 ~5 token
4. **查看结果**: 使用文件读取，避免重复查询

**典型工作流**（~50 token）:
```
1. 初始化: 5 token
2. 创建10个任务: 150 token
3. 查询状态: 5 token
4. 读取结果: 10 token（文件读取）
5. 清理: 5 token
```

## 📚 更多信息

- [完整文档](../README.md)
- [测试指南](tests/README.md)
- [部署指南](AI_DEPLOYMENT.md)
