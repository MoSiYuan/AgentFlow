# 🎯 CPDS实战测试：10个Agent自主选题创作

## ⚡ 核心理念

**让每个Agent自主选择题材，而不是预先分配**，这样才能真正测试分布式AI协作的能力！

### 对比

| 方式 | 旧方式（预先分配） | 新方式（自主选题） |
|------|-------------------|-----------------|
| 题材来源 | Master预先指定 | Agent自主查询和选择 |
| 题材冲突 | 可能重复 | 实时协调避免 |
| Agent角色 | 被动执行 | 主动决策 |
| 真实感 | 像脚本执行 | 像真实协作 |
| 扩展性 | 受限 | 无限 |

---

## 🚀 5分钟快速开始

### 步骤1：启动Master（终端1）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
./cpds master --mode standalone --auto-shutdown --port 8848
```

### 步骤2：创建10个通用任务（终端2）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 创建10个通用任务（使用通用模板）
for i in {1..10}; do
  TASK_ID="STORY-AUTO-$(printf '%03d' $i)"

  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d "{
      \"task_id\": \"$TASK_ID\",
      \"title\": \"Agent $i 自主创作克苏鲁故事\",
      \"description\": \"你是克苏鲁神话作家。请自主选择题材并创作500-1000字故事。流程：1) GET /api/topics 查询已选题材；2) 选择独特题材；3) POST /api/topics/register 注册；4) 创作。洛夫克拉夫特风格。\",
      \"priority\": \"high\",
      \"tags\": [\"creative-writing\", \"lovecraft\", \"autonomous\"],
      \"deployment_mode\": \"standalone\"
    }"

  echo "✅ Created task $i"
  sleep 0.5
done
```

### 步骤3：启动10个Agent（终端2继续）

```bash
# 并行启动10个Agent
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Agent-$i" \
    --oneshot \
    > /tmp/agent_$i.log 2>&1 &

  echo "🤖 Started Agent-$i (PID: $!)"
  sleep 0.2
done

echo ""
echo "⏳ Waiting for agents to select topics and create stories..."
wait

echo ""
echo "✅ All agents finished!"
```

### 步骤4：查看题材选择结果

```bash
# 查看所有已选题材
echo "📋 Selected topics:"
curl -s http://localhost:8848/api/topics | \
  jq -r '.topics[] | "• \(.topic) - by \(.worker_id)"'

# 或JSON格式
curl -s http://localhost:8848/api/topics | jq '.topics'
```

### 步骤5：查看故事结果

```bash
# 查看所有故事标题
curl -s http://localhost:8848/api/tasks/completed | \
  jq -r '.data.tasks[] | select(.task_id | startswith("STORY")) | "\(.title)\n\n\(.output[:200])...\n---"'
```

---

## 🔥 实时监控

### 监控题材选择

```bash
# 实时查看已选题材（自动刷新）
watch -n 2 'curl -s http://localhost:8848/api/topics | jq'
```

### 监控任务进度

```bash
# 查看系统状态
watch -n 2 'curl -s http://localhost:8848/api/status | jq'
```

### 查看Agent日志

```bash
# 查看特定Agent的执行过程
tail -f /tmp/agent_1.log
```

---

## 📊 新增API

### 获取已选题材

```bash
curl http://localhost:8848/api/topics
```

### 注册题材（Agent内部调用）

```bash
curl -X POST http://localhost:8848/api/topics/register \
  -H "Content-Type: application/json" \
  -d '{
    "worker_id": "Agent-1",
    "topic": "深海探险与远古遗迹",
    "story_id": "STORY-AUTO-001"
  }'
```

### 检查题材可用性

```bash
curl "http://localhost:8848/api/topics/available?topic=深海探险"
```

---

## ✨ 预期结果

### 题材分布示例

```
✅ Agent-1  → "深海探险与远古遗迹"
✅ Agent-2  → "宇宙恐怖与天文观测"
✅ Agent-3  → "梦境侵蚀与现实扭曲"
✅ Agent-4  → "禁忌知识与疯狂学者"
✅ Agent-5  → "时间循环与时空裂缝"
✅ Agent-6  → "身体变异与深潜杂交"
✅ Agent-7  → "古老神祇与苏醒仪式"
✅ Agent-8  → "被诅咒的鬼镇"
✅ Agent-9  → "南极探险与非人生物"
✅ Agent-10 → "理智边缘与守夜人"
```

### 故事统计

- 总故事数：10个
- 总字数：5000-10000字
- 平均字数：500-1000字/个
- 题材重复率：0%（通过协调机制）

---

## 🎯 测试价值

这个测试验证了：

### 1. 分布式决策
- ✅ 10个Agent独立决策
- ✅ 无中央调度器
- ✅ 自主选题机制

### 2. 动态协调
- ✅ 实时题材共享
- ✅ 冲突检测和避免
- ✅ 原子性操作

### 3. 并发控制
- ✅ 10个Agent同时查询
- ✅ 题材注册互斥
- ✅ 无死锁无竞争

### 4. API稳定性
- ✅ 新的题材管理端点
- ✅ 高并发请求处理
- ✅ 数据一致性

### 5. 真实场景
- ✅ 模拟真实多Agent协作
- ✅ Agent自主性而非被动
- ✅ 动态环境适应

---

## 🔧 故障排查

### 题材冲突

如果两个Agent同时选择同一题材：
```
Agent-1: POST /api/topics/register {"topic": "深海", ...}
Agent-2: POST /api/topics/register {"topic": "深海", ...}

结果:
Agent-1: ✅ 200 OK
Agent-2: ❌ 409 Conflict - "Topic already taken"
Agent-2: 重新选择其他题材
```

### Agent处理冲突

Agent会在任务描述中说明：
```
如果题材已被占用，请重新选择。
查询GET /api/topics了解已选题材，
选择一个未被选用的题材。
```

---

## 📚 完整文档

- **[AUTONOMOUS_TEST_GUIDE.md](AUTONOMOUS_TEST_GUIDE.md)** - 详细测试指南
- **[autonomous_selection.md](autonomous_selection.md)** - 设计文档
- **[topics.go](../../internal/master/topics.go)** - 实现代码

---

## 🎉 立即开始

```bash
# 1. 启动Master
./cpds master --mode standalone --auto-shutdown --port 8848

# 2. 创建任务 + 启动Agent
cd docs/cpds-test
./quick_test.sh

# 或手动执行
for i in {1..10}; do
  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @tasks/story_auto_template.json
done
```

**测试关键点**：
- ✅ 每个Agent自主选择题材
- ✅ 10个题材各不相同
- ✅ 无题材冲突
- ✅ 所有故事创作成功

这才是真正的**分布式AI协作实战测试**！
