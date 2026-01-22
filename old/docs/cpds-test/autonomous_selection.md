# CPDS实战测试：10个Agent自主协作创作

## 🎯 测试目标（修订版）

验证CPDS系统实现**真正的自主协作**：
1. **自主选题**：10个Agent自主选择克苏鲁故事题材，避免重复
2. **分布式决策**：每个Agent独立决策，无需中央分配
3. **动态协调**：通过Master协调避免题材冲突
4. **互相评价**：每个Agent分析其他9个故事并评分

## 🔄 新的协作流程

```
┌─────────────────────────────────────────────────────┐
│                  CPDS Master                         │
│  ┌────────────────────────────────────────────┐    │
│  │     题材注册表（实时更新）                  │    │
│  │  - 已选题材列表                             │    │
│  │  - 选中的Worker                            │    │
│  │  - 时间戳                                   │    │
│  └────────────────────────────────────────────┘    │
└───────────────┬─────────────────────────────────────┘
                │
     ┌──────────┴──────────┐
     │  API: GET/POST /api/topics  │
     │  查询可用题材、注册选题     │
     └──────────────────────────┘
                │
     ┌──────────┴───────────────────────────┐
     │                                      │
  ▼  ▼                                 ▼  ▼
┌────┐┌────┐                          ┌────┐┌────┐
│W1  ││W2  │  ...                  │W9  ││W10 │
└────┘└────┘                          └────┘└────┘
  │    │                                 │    │
  └────┴────────────────────────────────┘
       │
       ▼
  自主查询→选择→注册→创作
```

## 📝 新的任务设计

### 故事创作任务（通用模板）

```json
{
  "task_id": "STORY-AUTO-001",
  "title": "自主创作克苏鲁神话故事（Agent 1）",
  "description": "你是一位专业的克苏鲁神话作家。请自主选择一个创作主题并创作故事。\n\n【创作要求】\n1. 必须先查询可用题材列表\n2. 选择一个未被选用的克苏鲁神话主题\n3. 注册你的选题\n4. 开始创作（500-1000字）\n\n【选题流程】\n1. GET /api/topics - 查询已选题材\n2. 选择一个独特的主题（深海、太空、梦境、禁忌、时间等）\n3. POST /api/topics - 注册你的选题\n4. 开始创作\n\n【故事要求】\n- 500-1000字\n- 典型的洛夫克拉夫特风格\n- 强调人类渺小、未知恐惧\n- 具体的感官描写\n- 开放式结局\n\n请确保题材独特性，创作高质量作品。",
  "priority": "high",
  "tags": ["creative-writing", "lovecraft", "autonomous"],
  "deployment_mode": "standalone"
}
```

### 新增API端点

需要在Master中添加题材管理API：

```go
// 获取已选题材列表
GET /api/topics
Response: {
  "topics": [
    {
      "topic": "深海探险",
      "worker_id": "Writer-01",
      "selected_at": "2024-01-21T10:30:00Z"
    }
  ]
}

// 注册题材
POST /api/topics/register
Request: {
  "worker_id": "Writer-01",
  "topic": "深海探险与远古遗迹",
  "story_id": "STORY-001"
}

// 题材是否可用
GET /api/topics/available?topic=深海探险
Response: {
  "available": false,
  "reason": "已被Writer-01选用"
}
```

## 🔧 实现方案

### 1. 在Master中添加题材管理

在 `internal/master/` 中添加：

```go
// topics.go
type TopicManager struct {
    mu          sync.RWMutex
    selected    map[string]*TopicRegistration  // topic -> registration
}

type TopicRegistration struct {
    Topic      string
    WorkerID   string
    StoryID    string
    SelectedAt time.Time
}

// GET /api/topics - 获取已选题材
func (s *Server) handleGetTopics(c *gin.Context) {
    s.topicManager.mu.RLock()
    defer s.topicManager.mu.RUnlock()

    topics := s.topicManager.GetAll()
    c.JSON(http.StatusOK, gin.H{"topics": topics})
}

// POST /api/topics/register - 注册题材
func (s *Server) handleRegisterTopic(c *gin.Context) {
    var req struct {
        WorkerID string `json:"worker_id" binding:"required"`
        Topic    string `json:"topic" binding:"required"`
        StoryID  string `json:"story_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 检查是否已被选用
    if s.topicManager.IsTaken(req.Topic) {
        c.JSON(http.StatusConflict, gin.H{
            "error": "Topic already taken",
            "topic": req.Topic,
            "selected_by": s.topicManager.GetWorkerByTopic(req.Topic)
        })
        return
    }

    // 注册题材
    s.topicManager.Register(req.Topic, req.WorkerID, req.StoryID)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "topic": req.Topic,
        "worker_id": req.WorkerID
    })
}

// GET /api/topics/available - 检查题材是否可用
func (s *Server) handleCheckTopic(c *gin.Context) {
    topic := c.Query("topic")
    if topic == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "topic parameter required"})
        return
    }

    available := !s.topicManager.IsTaken(topic)

    response := gin.H{
        "topic": topic,
        "available": available
    }

    if !available {
        response["reason"] = "Already taken"
        response["selected_by"] = s.topicManager.GetWorkerByTopic(topic)
    }

    c.JSON(http.StatusOK, response)
}
```

### 2. 修改Worker执行逻辑

在 `internal/worker/oneshot_runner.go` 中添加题材选择逻辑：

```go
func (r *OneShotRunner) RunWithTopicSelection() error {
    ctx := context.Background()

    // 1. 注册
    if err := r.client.Register(); err != nil {
        return fmt.Errorf("failed to register: %w", err)
    }

    // 2. 获取任务
    tasks, err := r.client.GetPendingTasks()
    if err != nil || len(tasks) == 0 {
        return nil
    }

    task := tasks[0]

    // 3. 选择题材（自主决策）
    topic, err := r.selectTopic(ctx, task)
    if err != nil {
        return fmt.Errorf("failed to select topic: %w", err)
    }

    r.logger.Infof("Selected topic: %s", topic)

    // 4. 注册题材
    if err := r.client.RegisterTopic(r.client.GetWorkerID(), topic, task.TaskID); err != nil {
        return fmt.Errorf("failed to register topic: %w", err)
    }

    // 5. 开始创作
    result, err := r.executeTask(ctx, task, topic)
    if err != nil {
        return err
    }

    // 6. 完成任务
    r.client.CompleteTask(task.TaskID, &result.Output, nil)

    return nil
}

func (r *OneShotRunner) selectTopic(ctx, task *Task) (string, error) {
    // 查询已选题材
    taken, err := r.client.GetTakenTopics()
    if err != nil {
        return "", err
    }

    r.logger.Infof("Taken topics: %v", taken)

    // 让Claude自主选择题材
    claudeTask := *task
    claudeTask.Description = fmt.Sprintf(`你是Agent %s，需要创作一个克苏鲁神话故事。

【已选题材列表】
%v

【要求】
1. 从上面的列表中选择一个**未被选用**的题材
2. 题材必须符合克苏鲁神话特点（恐怖、未知、古老等）
3. 只返回题材名称，不要其他内容

【可选题材类型】
- 深海探险（潜艇、海底遗迹）
- 南极探险（冰盖、远古城市）
- 宇宙恐怖（天文观测、外星信号）
- 梦境侵蚀（梦境、现实扭曲）
- 禁忌知识（古籍、诅咒）
- 时间循环（时空裂缝）
- 身体变异（基因、杂交）
- 古老神祇（苏醒、祭祀）
- 鬼镇（被诅咒的地方）
- 疯狂与理智（精神病院）

请只返回一个题材名称（10个字以内）：`, r.client.GetWorkerName(), taken)

    result := r.claudeExecutor.ExecuteTask(ctx, &claudeTask)
    if result.Error != nil {
        return "", result.Error
    }

    topic := strings.TrimSpace(result.Output)
    return topic, nil
}
```

## 🎯 新的测试流程

### 步骤1：启动Master

```bash
./cpds master --mode standalone --auto-shutdown --port 8848
```

### 步骤2：创建10个通用任务

```bash
for i in {1..10}; do
  cat > /tmp/task_$i.json << EOF
{
  "task_id": "STORY-AUTO-$(printf '%03d' $i)",
  "title": "自主创作克苏鲁故事",
  "description": "你是专业的克苏鲁神话作家。请自主选择一个独特的题材并创作500-1000字的故事。要求：1)先查询/api/topics获取已选题材；2)选择未被选用的题材；3)POST/api/topics/register注册；4)开始创作。洛夫克拉夫特风格，强调未知恐惧和人类渺小。",
  "priority": "high",
  "tags": ["creative-writing", "lovecraft"],
  "deployment_mode": "standalone"
}
EOF

  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @/tmp/task_$i.json
done
```

### 步骤3：启动10个Workers（自主选题）

```bash
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Agent-$i" \
    --oneshot &
done

wait
echo "✅ 10个Agent完成了自主选题和创作！"
```

### 步骤4：查看结果

```bash
# 查看所有已选题材
curl http://localhost:8848/api/topics | jq '.topics'

# 查看完成的故事
curl -s http://localhost:8848/api/tasks/completed | \
  jq '.data.tasks[] | select(.task_id | startswith("STORY")) | {task_id, assigned_to, output}'
```

## ✨ 实战意义

这种新设计的优势：

1. **真正的自主性** - Agent自主决策，不是被动执行
2. **动态协调** - 通过API实时协调避免冲突
3. **可扩展性** - 可以支持更多Agent和更多题材
4. **真实场景** - 模拟真实的多Agent协作
5. **冲突处理** - 题材冲突时需要重新选择

## 🔥 实现优先级

需要立即实现：
1. ✅ Master添加题材管理API
2. ✅ Worker添加题材选择逻辑
3. ✅ 修改任务配置为通用模板
4. ✅ 测试完整的自主选题流程

这样才是真正的**分布式AI协作系统**！
