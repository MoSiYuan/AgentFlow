# CPDS三阶段工作流实现状态

## 📊 总体进度

### ✅ Stage 1: 抢题材 - 已完成并测试

**实现内容**:
- ✅ 真实HTTP API调用 `GET /api/topics` 查询已选题材
- ✅ 智能题材选择算法（避开已占用的题材）
- ✅ 真实HTTP API调用 `POST /api/topics/register` 注册题材
- ✅ 409 Conflict处理（题材冲突时自动重选）
- ✅ 10个Agent并发测试通过

**代码位置**:
- `internal/worker/claude_executor.go:231-333`
  - `queryTopics()` - 查询已选题材
  - `selectUniqueTopic()` - 选择独特题材
  - `registerTopic()` - 注册题材（含冲突检测）

**测试结果**:
```
✅ Agent-1: 选择了 "深海遗迹中的古老召唤"
✅ Agent-2: 选择了 "被诅咒的家族族谱" （避开重复）
✅ 10/10 任务成功完成
✅ 所有题材正确注册到Master
```

---

### ⚠️ Stage 2: 输出文章 - 部分完成

**当前状态**:
- ✅ 基本框架已实现（`generateStoryWithClaude`）
- ✅ Prompt模板已准备
- ⚠️ 使用模拟模式（未调用真实Claude API）
- ❌ 需要集成anthropic-go库

**代码位置**:
- `internal/worker/claude_executor.go:335-373`
  - `generateStoryWithClaude()` - 当前返回模拟内容

**待实现功能**:
1. **安装依赖**:
   ```bash
   go get github.com/anthropics/anthropic-go
   ```

2. **更新代码** (`claude_executor.go`):
   ```go
   import "github.com/anthropics/anthropic-go"

   // 更新结构体添加API客户端
   type ClaudeExecutor struct {
       // ... 现有字段
       anthropicClient *anthropic.Client
       apiKey          string
   }

   // 更新generateStoryWithClaude方法
   func (e *ClaudeExecutor) generateStoryWithClaude(topic string, task *database.Task) string {
       message := e.anthropicClient.Messages.New(context.Background())

       resp, err := message.WithModel("claude-3-5-sonnet-20241022").
           WithMaxTokens(4096).
           WithMessages([]anthropic.MessageParam{
               anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
           }).
           Execute()

       if err != nil {
           return fmt.Sprintf("Error: %v", err)
       }

       return resp.Content[0].Text
   }
   ```

3. **设置API Key**:
   ```bash
   export ANTHROPIC_API_KEY="your-key-here"
   ```

---

### ❌ Stage 3: Agent互评 - 未实现

**需求分析**:
每个Agent需要评论其他9个Agent的作品，共90条评论（10×9）。

**实现方案**:

#### 步骤1: 创建评论任务
在Master中添加API端点：
```go
// POST /api/tasks/create-reviews
func (s *Server) handleCreateReviewTasks(c *gin.Context) {
    // 1. 查询所有已完成的story任务
    // 2. 为每个agent创建9个评论任务
    // 3. 评论任务格式：REVIEW-{ReviewerID}-{StoryID}
}
```

#### 步骤2: 评论任务模板
```json
{
  "task_id": "REVIEW-001-STORY-002",
  "title": "Agent-1 评论 Agent-2 的作品",
  "description": "请阅读以下克苏鲁故事并提供建设性评论：
    [Story Content]

    请从以下维度评价（1-10分）：
    1. 恐怖氛围营造
    2. 故事创意
    3. 文笔流畅度
    4. 洛夫克拉夫特风格还原度

    总体评价：___/10
    优缺点分析：
    改进建议：
  ",
  "reviewed_story_id": "STORY-002",
  "reviewer_id": "Agent-1"
}
```

#### 步骤3: Worker执行评论
Worker需要：
1. 获取被评论的故事内容
2. 调用Claude API生成评论
3. 提交评论结果

**代码结构**:
```
internal/worker/
  ├── reviewer.go          - 新增：评论逻辑
  └── claude_executor.go  - 扩展：添加generateReview()
```

#### 步骤4: 汇总评论结果
在Master中添加API端点：
```go
// GET /api/reviews/:story_id
func (s *Server) handleGetReviews(c *gin.Context) {
    // 返回某个故事的所有评论
}

// GET /api/reviews/summary
func (s *Server) handleReviewSummary(c *gin.Context) {
    // 返回所有评论的统计汇总
}
```

---

## 🚀 快速开始测试

### 当前可用功能
```bash
# Stage 1: 抢题材（完整实现）
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
chmod +x docs/cpds-test/run_complete_workflow.sh
./docs/cpds-test/run_complete_workflow.sh
```

### 预期输出
```
✅ Stage 1 (Topic Selection): COMPLETE
   - Tasks completed: 10/10
   - Topics registered: 10

⚠️  Stage 2 (Article Generation): SIMULATION ONLY
   - Stories generated: 10 (simulated)

❌ Stage 3 (Peer Review): NOT IMPLEMENTED
```

---

## 📝 实现清单

### Stage 2: 真实文章生成
- [ ] 安装anthropic-go依赖
- [ ] 实现Claude API客户端初始化
- [ ] 更新`generateStoryWithClaude()`调用真实API
- [ ] 添加错误处理和重试逻辑
- [ ] 测试Token使用统计

### Stage 3: Agent互评系统
- [ ] 设计评论任务数据结构
- [ ] 实现`POST /api/tasks/create-reviews`端点
- [ ] 实现Worker评论逻辑
- [ ] 实现`GET /api/reviews`查询端点
- [ ] 创建评论结果汇总脚本
- [ ] 测试完整评论流程

### 优化和增强
- [ ] 添加并发控制和速率限制
- [ ] 实现评论质量过滤
- [ ] 添加评论可视化Dashboard
- [ ] 支持多轮评论和修改

---

## 🔧 使用git worktree并行开发

由于三个Stage相对独立，可以使用git worktree并行开发：

```bash
# 主分支：Stage 1（已完成）
git checkout main

# 创建新分支用于Stage 2
git worktree add ../cpds-go-stage2 feature/stage2-claude-api
cd ../cpds-go-stage2

# 实现Stage 2的Claude API集成
# ... 开发工作 ...
git add .
git commit -m "feat: integrate Claude API for story generation"

# 创建新分支用于Stage 3
git worktree add ../cpds-go-stage3 feature/stage3-peer-review
cd ../cpds-go-stage3

# 实现Stage 3的评论系统
# ... 开发工作 ...
git add .
git commit -m "feat: implement agent peer review system"
```

这样可以在不同分支上并行开发，互不影响。

---

## 📊 当前系统架构

```
┌─────────────────────────────────────────────────────────┐
│                      Master Server                       │
│  - Task Management (SQLite)                              │
│  - Topic Manager (in-memory)                             │
│  - REST API (9+ endpoints)                               │
└─────────────────────────────────────────────────────────┘
           ↑                    ↓
           │                    │
    [Worker Register]    [Task Distribution]
           │                    │
           ↓                    ↓
┌─────────────────────────────────────────────────────────┐
│                    Worker Agents                         │
│  1. Register Worker                                      │
│  2. Claim Task (POST /api/tasks/assign)                 │
│  3. Query Topics (GET /api/topics)                      │
│  4. Select Unique Topic                                  │
│  5. Register Topic (POST /api/topics/register)          │
│  6. Generate Story (Claude API - TODO)                   │
│  7. Complete Task (POST /api/tasks/complete)            │
└─────────────────────────────────────────────────────────┘
```

---

## 📞 联系和支持

- **测试脚本**: `docs/cpds-test/run_complete_workflow.sh`
- **日志位置**: `/tmp/cpds_complete_test.log`
- **数据库**: `.claude/cpds-manager/master.db`
- **文档**: `docs/cpds-test/WORKFLOW_STATUS.md` (本文件)
