# CPDS实战测试：克苏鲁故事创作与分析

## 测试概览

**目标**：10个Worker并行创作10个克苏鲁故事，然后互相评价

**流程**：
1. Master启动
2. 创建10个故事任务
3. 启动10个Worker（并行创作）
4. 创建90个分析任务
5. 启动10个Worker（并行分析）
6. 生成报告

## 一键执行

### 步骤1：启动Master（终端1）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
./cpds master --mode standalone --auto-shutdown --port 8848
```

### 步骤2：批量创建任务（终端2）

```bash
# 创建10个故事任务
for i in {1..10}; do
  curl -X POST http://localhost:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d @docs/cpds-test/tasks/story_$(printf '%02d' $i).json
done

# 查看任务状态
curl http://localhost:8848/api/status | jq '.'
```

### 步骤3：启动10个Writer Workers（终端3）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 并行启动10个Worker
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Writer-$i" \
    --oneshot \
    > /tmp/worker_$i.log 2>&1 &
done

# 监控进度
watch -n 2 'curl -s http://localhost:8848/api/status | jq'
```

### 步骤4：查看故事结果

```bash
# 获取所有已完成的故事
curl -s http://localhost:8848/api/tasks/completed | \
  jq '.data.tasks[] | select(.task_id | startswith("STORY"))'
```

### 步骤5：创建分析任务

使用Python脚本批量创建90个分析任务：

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go/docs/cpds-test

# 创建分析任务
python3 create_analysis_tasks.py
```

**create_analysis_tasks.py**:

```python
#!/usr/bin/env python3
import requests
import json

MASTER_URL = "http://localhost:8848"

# 获取已完成的故事
resp = requests.get(f"{MASTER_URL}/api/tasks/completed")
tasks = resp.json()['data']['tasks']
stories = [t for t in tasks if t['task_id'].startswith('STORY')]

print(f"Found {len(stories)} stories")
print("Creating analysis tasks...")

count = 0
for story in stories:
    story_id = story['task_id']
    story_num = story_id.split('-')[1]
    story_content = story.get('output', '')

    # 为每个故事创建9个分析（排除作者）
    for critic_num in range(1, 11):
        if int(story_num) == critic_num:
            continue

        analysis_id = f"ANALYSIS-{story_num}-{critic_num:02d}"

        task = {
            "task_id": analysis_id,
            "title": f"分析故事{story_num}",
            "description": f"""分析以下克苏鲁故事的文学价值：

【故事内容】
{story_content}

【要求】
1. 识别克苏鲁元素（3-5个）
2. 评价恐怖氛围
3. 评价创意和文笔
4. 给出评分（1-10分）
5. 总结论（2-3句）
字数：300-500字""",
            "priority": "medium",
            "tags": ["analysis", "critique"],
            "deployment_mode": "standalone"
        }

        resp = requests.post(f"{MASTER_URL}/api/tasks/create", json=task)
        if resp.status_code == 200:
            print(f"✅ {analysis_id}")
            count += 1
        else:
            print(f"❌ {analysis_id}: {resp.text}")

print(f"\n✅ Created {count} analysis tasks")
```

### 步骤6：启动分析Workers（终端3）

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go

# 清理之前的Worker
pkill -f "cpds worker"
sleep 2

# 启动10个分析Worker
for i in {1..10}; do
  ./cpds worker --mode standalone \
    --master http://localhost:8848 \
    --name "Critic-$i" \
    --oneshot \
    > /tmp/critic_$i.log 2>&1 &
done

# 监控进度
watch -n 2 'curl -s http://localhost:8848/api/status | jq'
```

### 步骤7：生成报告

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go/docs/cpds-test

python3 generate_report.py
```

**generate_report.py**:

```python
#!/usr/bin/env python3
import requests
from datetime import datetime

MASTER_URL = "http://localhost:8848"

# 获取所有任务
resp = requests.get(f"{MASTER_URL}/api/tasks/completed")
tasks = resp.json()['data']['tasks']

stories = [t for t in tasks if t['task_id'].startswith('STORY')]
analyses = [t for t in tasks if t['task_id'].startswith('ANALYSIS')]

# 生成报告
with open('final_report.md', 'w', encoding='utf-8') as f:
    f.write("# CPDS实战测试报告\n\n")
    f.write(f"**时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")

    # 故事列表
    f.write("## 一、故事列表\n\n")
    for story in stories:
        f.write(f"### {story['title']}\n\n")
        output = story.get('output', '')
        f.write(f"**内容**:\n\n{output}\n\n")
        f.write("---\n\n")

    # 分析汇总
    f.write("## 二、分析汇总\n\n")
    for story in stories:
        story_num = story['task_id'].split('-')[1]
        f.write(f"### 故事{story_num}的评价\n\n")

        story_analyses = [a for a in analyses
                          if a['task_id'][8:11] == story_num]

        for analysis in story_analyses:
            critic_num = analysis['task_id'][12:14]
            f.write(f"#### Critic{crit_num}:\n\n")
            f.write(f"{analysis.get('output', '')}\n\n")
            f.write("---\n\n")

    # 统计
    f.write("## 三、统计信息\n\n")
    f.write(f"- 故事总数: {len(stories)}\n")
    f.write(f"- 分析总数: {len(analyses)}\n")
    f.write(f"- 每个故事平均收到: {len(analyses)/len(stories):.1f} 份评价\n")

print("✅ 报告已生成: final_report.md")
print(f"   故事: {len(stories)}")
print(f"   分析: {len(analyses)}")
```

## API快速参考

### 查看状态
```bash
curl http://localhost:8848/api/status | jq
```

### 查看待处理任务
```bash
curl http://localhost:8848/api/tasks/pending | jq
```

### 查看已完成任务
```bash
curl http://localhost:8848/api/tasks/completed | jq
```

### 查看在线Workers
```bash
curl http://localhost:8848/api/workers | jq
```

## 测试结果

完成后的文件结构：
```
docs/cpds-test/
├── final_report.md          # 综合报告
├── tasks/                   # 任务定义
├── logs/                    # Worker日志
└── stories/                 # 如果要保存故事（手动保存）
```

## 预期完成时间

- **故事创作**: 约5-10分钟（取决于Claude API响应）
- **分析评价**: 约10-20分钟
- **总计**: 约15-30分钟
