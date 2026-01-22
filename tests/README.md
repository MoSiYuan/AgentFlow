# AgentFlow - AI Agent 快速使用指南

> 3 分钟上手，30 个令牌完成

## 最简用法

```bash
# 1. 运行测试（10个故事+评审，3秒完成）
cd tests && go run ctest_pure.go

# 2. 查看结果
cat ctest_stories/story_1.md
```

## 核心概念

**任务类型**:
- `write_story` - 生成内容
- `review_story` - 评审内容
- `task:xxx` - AI 子任务

**文件格式**:
```
write_story:标题:类型:索引
review_story:任务ID:标题:review
```

## 快速集成

### 1. 直接调用

```go
// 创建任务
agentflow add "标题" --desc "write_story:xxx:type:001" --db agentflow.db

// 查看结果
agentflow list --status completed --db agentflow.db
```

### 2. 作为 Skill 使用

```python
# Python 示例
import subprocess

def create_story(title):
    subprocess.run(['agentflow', 'add', title,
                   '--desc', f'write_story:{title}:type:001',
                   '--db', 'agentflow.db'])
```

### 3. API 调用

```bash
# 创建任务
curl -X POST http://localhost:8848/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"测试","description":"write_story:test:type:001"}'
```

## 自定义任务

### 修改内容生成

编辑 `internal/worker/ai_worker.go` 的 `generateCthulhuStory` 函数

### 添加新类型

```go
case "write_code":
    return w.executeCodeTask(ctx, task, params)
```

## 节省 Token 的技巧

1. **使用简短命令**: `agentflow add "T" --desc "s:T:t:1"` 而非完整描述
2. **批量操作**: 一次创建多个任务而非逐个创建
3. **状态查询**: `agentflow list --status completed` 过滤查看
4. **重用 worker_id**: 环境变量设置后无需重复

## 常见问题

**Q: 如何让 Worker 评审其他 Worker 的作品？**
A: 使用 `review_story:task_id:title:review` 格式

**Q: 如何自定义生成内容？**
A: 修改 `generateCthulhuStory` 函数

**Q: 如何查看任务进度？**
A: `agentflow list --status running`

## 完整示例

```bash
# 1. 初始化
agentflow init agentflow.db

# 2. 启动 Master
agentflow master --db agentflow.db &

# 3. 添加10个任务
for i in {1..10}; do
  agentflow add "故事$i" --desc "write_story:故事$i:type:00$i" --db agentflow.db
done

# 4. 查看结果
agentflow list --db agentflow.db
cat ~/agentflow-workspace/cthulhu_stories/story_1.md
```

## 文档位置

- 测试: `tests/ctest_pure.go`
- 结果: `tests/ctest_stories/`
- 报告: `tests/TEST_REPORT.md`
- 完整文档: `docs/`
