# CPDS实战测试完整指南

## 📋 测试目标

使用CPDS系统实现10个Agent的协作：
1. **10个Writer Worker** - 并行创作10个不同主题的克苏鲁故事
2. **90个Critic Worker** - 每个Worker分析其他9个故事
3. **自动汇总** - 生成完整的测试报告

## 🚀 快速开始

### 第一步：启动Master

打开**终端1**：

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
./cpds master --mode standalone --auto-shutdown --port 8848
```

**预期输出**：
```
╔═══════════════════════════════════════════════════════════╗
║           🚀 CPDS Master Server Started                   ║
╠═══════════════════════════════════════════════════════════╣
║  Mode:       standalone                                    ║
║  Host:       0.0.0.0                                      ║
║  Port:       8848                                         ║
║  Auto-Shutdown: ENABLED                                   ║
╚═══════════════════════════════════════════════════════════╝
```

---

### 第二步：创建故事任务

打开**终端2**，批量创建10个故事创作任务：

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 创建10个故事任务
for i in {1..10}; do
  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @docs/cpds-test/tasks/story_$(printf '%02d' $i).json
  echo "Created task $i"
done

# 查看任务状态
curl -s http://localhost:8848/api/status | jq '.'
```

**预期输出**：
```
Created task 1
Created task 2
...
Created task 10

{
  "pending_tasks": 10,
  "in_progress_tasks": 0,
  "completed_tasks": 0,
  "online_workers": 0
}
```

---

### 第三步：启动10个Writer Workers

打开**终端3**，并行启动10个Worker：

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 并行启动10个Writer Worker
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Writer-$i" \
    --oneshot \
    > /tmp/worker_$i.log 2>&1 &
  echo "Started Writer-$i (PID: $!)"
done

echo "Waiting for workers to complete..."
wait

echo "All workers finished!"
```

**预期输出**：
```
Started Writer-1 (PID: 12345)
Started Writer-2 (PID: 12346)
...
Started Writer-10 (PID: 12354)
Waiting for workers to complete...

🤖 Starting CPDS Worker in standalone mode
   Master: http://localhost:8848
   One-Shot: ENABLED

📡 Registering to master...
✅ Registered: Writer-1

📋 Found task: STORY-001 - 《深海之影》...
⚙️  Executing task...
✅ Task completed: STORY-001

All workers finished!
```

---

### 第四步：查看故事结果

```bash
# 获取所有已完成的故事
curl -s http://localhost:8848/api/tasks/completed | \
  jq '.data.tasks[] | select(.task_id | startswith("STORY")) | {task_id, title}'
```

**或者保存到文件**：

```bash
# 创建stories目录并保存
mkdir -p docs/cpds-test/stories

for i in {1..10}; do
  TASK_ID="STORY-$(printf '%03d' $i)"
  curl -s "http://localhost:8848/api/tasks/$TASK_ID" | \
    jq -r '.data.output' > docs/cpds-test/stories/story_$i.md
  echo "Saved story_$i.md"
done
```

---

### 第五步：创建分析任务

使用Python脚本批量创建90个分析任务（每个故事9个分析）：

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go/docs/cpds-test

# 运行脚本
python3 create_analysis_tasks.py
```

**预期输出**：
```
============================================================
CPDS实战测试 - 创建分析任务
============================================================

✅ Master is running

📚 Fetching completed stories...
✅ Found 10 stories

📝 Creating analysis tasks...

Story 001: 《深海之影》- 深海探险主题克苏鲁故事
  ✅ ANALYSIS-001-02
  ✅ ANALYSIS-001-03
  ...

Story 010: 《最后的守夜人》- 疯狂与理智主题克苏鲁故事
  ✅ ANALYSIS-010-01
  ...

============================================================
Task Creation Summary
============================================================
✅ Successfully created: 90 analysis tasks

✅ Analysis tasks are ready!
```

---

### 第六步：启动Critic Workers

在**终端3**（清理旧Worker后启动新的）：

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 清理之前的Worker
pkill -f "cpds worker"
sleep 2

# 启动10个Critic Worker
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Critic-$i" \
    --oneshot \
    > /tmp/critic_$i.log 2>&1 &
  echo "Started Critic-$i (PID: $!)"
done

echo "Waiting for critics to complete..."
wait

echo "All critics finished!"
```

---

### 第七步：生成最终报告

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go/docs/cpds-test

python3 generate_report.py
```

**预期输出**：
```
============================================================
CPDS实战测试 - 生成报告
============================================================

📊 Fetching task data...
  Stories: 10
  Analyses: 90

📝 Generating report...

✅ Report generated: final_report.md

============================================================
Report Summary
============================================================
Stories: 10
Analyses: 90
Total tasks: 100

View report:
  cat final_report.md
  or
  less final_report.md
```

---

## 📊 查看报告

```bash
# 查看报告
less docs/cpds-test/final_report.md

# 或在浏览器中打开（如果有Markdown预览）
open docs/cpds-test/final_report.md
```

---

## 🎯 API查询示例

### 查看实时状态（监控进度）

```bash
# 系统状态
watch -n 2 'curl -s http://localhost:8848/api/status | jq'

# 待处理任务
curl -s http://localhost:8848/api/tasks/pending | jq '.data.tasks | length'

# 执行中任务
curl -s http://localhost:8848/api/tasks/running | jq '.data.tasks[]'

# 已完成任务数
curl -s http://localhost:8848/api/tasks/completed | jq '.data.tasks | length'

# 在线Workers
curl -s http://localhost:8848/api/workers | jq '.data.workers | length'
```

---

## ✅ 成功标准

测试成功需要满足：

- [x] Master启动并运行正常
- [x] 10个故事任务创建成功
- [x] 10个Writer Worker完成任务
- [x] 每个故事内容不重复
- [x] 90个分析任务创建成功
- [x] 10个Critic Worker完成分析
- [x] 最终报告生成成功
- [x] Master在所有任务完成后自动关闭

---

## 📈 性能指标

- **并行度**: 10个Worker同时工作
- **任务数**: 100个任务（10创作 + 90分析）
- **预计时间**: 15-30分钟（取决于Claude API响应）
- **成功率**: 100%（所有任务成功完成）

---

## 🔧 故障排查

### Master无法启动

```bash
# 检查端口占用
lsof -i :8848

# 如果被占用，杀掉进程
killall cpds
```

### Worker无法连接

```bash
# 测试Master连接
curl http://localhost:8848/api/health

# 检查防火墙
# macOS
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate
```

### 任务卡住

```bash
# 查看Worker日志
tail -f /tmp/worker_*.log

# 重启Worker
pkill -f "cpds worker"
# 然后重新启动
```

---

## 📁 文件结构

测试完成后生成的文件：

```
docs/cpds-test/
├── final_report.md          # 最终报告（包含所有故事和分析）
├── create_analysis_tasks.py # 创建分析任务的脚本
├── generate_report.py       # 生成报告的脚本
├── tasks/                   # 任务定义文件
│   ├── story_01.json
│   ├── ...
│   └── story_10.json
└── logs/                    # Worker日志（如果有）
```

---

## 🎉 测试完成

所有任务完成后，Master会自动关闭（`--auto-shutdown`模式）。

您可以查看 `final_report.md` 来欣赏10个Agent协作创作的克苏鲁故事和互相评价！
