# CPDS实战测试：10个Agent自主选题创作

## 🎯 核心思想

**真正的分布式AI协作**：每个Agent自主决策，动态协调，避免冲突。

### 旧方式（预先分配）❌
```
Master: "Agent 1，你写深海故事"
Agent 1: "好的"
Master: "Agent 2，你写南极故事"
Agent 2: "好的"
...
```

### 新方式（自主选题）✅
```
Agent 1: "我查查已选题材...嗯，我选深海探险吧"
Agent 2: "我想想...南极被选了，我选宇宙恐怖吧"
Agent 3: "梦境题材不错，我要了"
...
```

## 🔄 完整流程

```
1. Master启动
   └─ 初始化空的题材注册表

2. 创建10个通用任务
   └─ 任务描述相同，不指定主题

3. 启动10个Agent
   ├─ Agent 1 查询题材 → 选择"深海探险" → 注册 → 创作
   ├─ Agent 2 查询题材 → 选择"宇宙恐怖" → 注册 → 创作
   ├─ Agent 3 查询题材 → 选择"梦境侵蚀" → 注册 → 创作
   └─ ... (动态协调，避免冲突)

4. 题材列表实时更新
   GET /api/topics →
   {
     "topics": [
       {"topic": "深海探险", "worker_id": "Agent-1", "selected_at": "..."},
       {"topic": "宇宙恐怖", "worker_id": "Agent-2", "selected_at": "..."},
       ...
     ]
   }

5. 所有故事创作完成
   └─ 10个独特的故事，无重复

6. 创建分析任务
   └─ 每个故事被其他9个Agent分析

7. 启动10个Critic Agent
   └─ 互相分析评价

8. 生成最终报告
   └─ 包含所有故事和分析
```

## 🆕 新增API端点

### 1. 获取已选题材列表
```bash
curl http://localhost:8848/api/topics
```

响应：
```json
{
  "success": true,
  "topics": [
    {
      "topic": "深海探险与远古遗迹",
      "worker_id": "Agent-1",
      "story_id": "STORY-AUTO-001",
      "selected_at": "2024-01-21T14:30:00Z"
    }
  ],
  "count": 1
}
```

### 2. 注册题材（原子操作）
```bash
curl -X POST http://localhost:8848/api/topics/register \
  -H "Content-Type: application/json" \
  -d '{
    "worker_id": "Agent-2",
    "topic": "宇宙恐怖与天文观测",
    "story_id": "STORY-AUTO-002"
  }'
```

成功响应：
```json
{
  "success": true,
  "topic": "宇宙恐怖与天文观测",
  "worker_id": "Agent-2",
  "story_id": "STORY-AUTO-002"
}
```

冲突响应：
```json
{
  "error": "Topic already taken",
  "topic": "深海探险",
  "taken_by": "Agent-1"
}
```

### 3. 检查题材可用性
```bash
curl "http://localhost:8848/api/topics/available?topic=深海探险"
```

响应：
```json
{
  "topic": "深海探险",
  "available": false,
  "taken_by": "Agent-1"
}
```

## 📝 快速测试

### 步骤1：启动Master

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
./cpds master --mode standalone --auto-shutdown --port 8848
```

### 步骤2：创建10个通用任务

```bash
# 使用通用模板
for i in {1..10}; do
  TASK_ID="STORY-AUTO-$(printf '%03d' $i)"

  cat > /tmp/task_$i.json << EOF
{
  "task_id": "$TASK_ID",
  "title": "Agent $i 自主创作克苏鲁故事",
  "description": "你是专业的克苏鲁神话作家。请自主选择题材并创作500-1000字故事。流程：1) GET /api/topics 查询已选题材；2)选择独特题材；3) POST /api/topics/register 注册；4)创作。洛夫克拉夫特风格，强调未知恐惧。",
  "priority": "high",
  "tags": ["creative-writing", "lovecraft", "autonomous"],
  "deployment_mode": "standalone"
}
EOF

  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @/tmp/task_$i.json

  echo "✅ Created task $i"
done
```

### 步骤3：启动10个Agent（自主选题+创作）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 启动10个Agent
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Agent-$i" \
    --oneshot \
    > /tmp/agent_$i.log 2>&1 &

  echo "🤖 Started Agent-$i (PID: $!)"
done

echo ""
echo "⏳ Waiting for agents to complete..."
wait

echo ""
echo "✅ All agents finished!"
```

### 步骤4：查看题材选择结果

```bash
# 查看所有已选题材
curl -s http://localhost:8848/api/topics | jq '.topics[]'

# 或更详细的输出
curl -s http://localhost:8848/api/topics | jq '.topics[] | {topic, worker_id}'
```

预期输出：
```
{
  "topic": "深海探险与远古遗迹",
  "worker_id": "Agent-1"
}
{
  "topic": "宇宙恐怖与天文观测",
  "worker_id": "Agent-2"
}
...
```

### 步骤5：查看故事结果

```bash
# 获取所有故事
curl -s http://localhost:8848/api/tasks/completed | \
  jq -r '.data.tasks[] | select(.task_id | startswith("STORY")) | .output' | \
  head -20
```

## 🔥 核心优势

### 1. 真正的自主性
- Agent不是被动执行，而是主动决策
- 每个Agent都可以选择自己喜欢的题材
- 更符合真实的多Agent协作场景

### 2. 动态协调
- 题材列表实时更新
- 冲突检测和避免
- 无需中央调度

### 3. 可扩展性
- 可以支持更多Agent（20个、50个...）
- 题材池可以无限扩展
- 系统会自动协调

### 4. 真实测试
- 测试并发控制（题材注册）
- 测试原子操作
- 测试冲突处理
- 测试分布式决策

## 📊 预期结果

### 题材选择
```
Agent-1  → "深海探险与远古遗迹"
Agent-2  → "宇宙恐怖与天文观测"
Agent-3  → "梦境侵蚀与现实扭曲"
Agent-4  → "禁忌知识与疯狂学者"
Agent-5  → "时间循环与时空裂缝"
Agent-6  → "身体变异与深潜杂交"
Agent-7  → "古老神祇与苏醒仪式"
Agent-8  → "被诅咒的鬼镇"
Agent-9  → "南极探险与非人生物"
Agent-10 → "理智边缘与守夜人"
```

### 故事质量
- 每个故事都符合Agent自主选择的题材
- 避免了重复题材
- 每个故事都是Agent的独立创作

## 🎉 总结

这种新设计实现了：

✅ **真正的分布式AI协作** - Agent自主决策
✅ **动态协调机制** - 实时题材管理
✅ **冲突避免** - 原子性题材注册
✅ **可扩展性** - 支持任意数量的Agent
✅ **真实性** - 模拟真实的多Agent协作场景

这才是真正的**实战测试**！

## 📚 相关文档

- [autonomous_selection.md](autonomous_selection.md) - 详细设计文档
- [topics.go](../../internal/master/topics.go) - 题材管理实现
- [server.go](../../internal/master/server.go) - Master服务器
