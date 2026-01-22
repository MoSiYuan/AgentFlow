#!/bin/bash
# CPDS自主选题快速测试脚本

set -e

echo "╔══════════════════════════════════════════════════════════╗"
echo "║     CPDS实战测试：10个Agent自主选题创作              ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# 检查二进制
if [ ! -f "../../../cpds" ]; then
    echo "❌ CPDS binary not found"
    echo "   Run: cd ../.. && go build -o cpds ./cpds"
    exit 1
fi

# 检查Master是否运行
if ! curl -s http://localhost:8848/api/health > /dev/null 2>&1; then
    echo "❌ Master not running"
    echo "   Please start Master first:"
    echo "   ./cpds master --mode standalone --auto-shutdown --port 8848"
    exit 1
fi

echo "✅ Master is running"
echo ""

# 创建10个通用任务
echo "📝 Creating 10 autonomous story tasks..."
echo ""

SUCCESS_COUNT=0

for i in {1..10}; do
    TASK_ID="STORY-AUTO-$(printf '%03d' $i)"
    WORKER_NAME="Agent-$i"

    RESPONSE=$(curl -s -X POST http://localhost:8848/api/tasks/create \
        -H "Content-Type: application/json" \
        -d "{
            \"task_id\": \"$TASK_ID\",
            \"title\": \"$WORKER_NAME 自主创作克苏鲁故事\",
            \"description\": \"你是克苏鲁神话作家。请自主选择题材并创作500-1000字故事。流程：1) GET /api/topics 查询已选题材；2) 选择独特题材；3) POST /api/topics/register 注册；4) 创作。洛夫克拉夫特风格，强调未知恐惧和人类渺小。如果题材冲突请重新选择。\",
            \"priority\": \"high\",
            \"tags\": [\"creative-writing\", \"lovecraft\", \"autonomous\"],
            \"deployment_mode\": \"standalone\"
        }")

    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo "  ✅ Created task for $WORKER_NAME"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        echo "  ❌ Failed for $WORKER_NAME"
    fi
done

echo ""
if [ $SUCCESS_COUNT -eq 10 ]; then
    echo "✅ All 10 tasks created successfully!"
else
    echo "⚠️  Only $SUCCESS_COUNT/10 tasks created"
    exit 1
fi

# 启动10个Agent
echo ""
echo "🤖 Starting 10 Agents..."
echo ""

AGENT_PIDS=()

for i in {1..10}; do
    AGENT_NAME="Agent-$i"
    LOG_FILE="/tmp/cpds_agent_$i.log"

    ../../../cpds worker --mode standalone \
        --master http://localhost:8848 \
        --name "$AGENT_NAME" \
        --oneshot \
        > "$LOG_FILE" 2>&1 &

    PID=$!
    AGENT_PIDS+=($PID)
    echo "  🤖 Started $AGENT_NAME (PID: $PID, Log: $LOG_FILE)"

    sleep 0.2
done

echo ""
echo "⏳ Waiting for agents to select topics and create stories..."
echo "   (This will take 5-10 minutes depending on Claude API response time)"
echo ""

# 监控进度
ELAPSED=0
TIMEOUT=600  # 10分钟超时

while [ $ELAPSED -lt $TIMEOUT ]; do
    sleep 10
    ELAPSED=$((ELAPSED + 10))

    # 检查已完成的Agent数量
    COMPLETED=$(curl -s http://localhost:8848/api/tasks/completed | jq '.data.tasks | length // length' 2>/dev/null || echo 0)
    RUNNING=$(curl -s http://localhost:8848/api/tasks/running | jq '.data.tasks | length // length' 2>/dev/null || echo 0)
    PENDING=$(curl -s http://localhost:8848/api/tasks/pending | jq '.data.tasks | length // length' 2>/dev/null || echo 0)

    # 检查是否有Agent还在运行
    RUNNING_AGENTS=0
    for pid in "${AGENT_PIDS[@]}"; do
        if ps -p $pid > /dev/null 2>&1; then
            RUNNING_AGENTS=$((RUNNING_AGENTS + 1))
        fi
    done

    echo "  ⏱️  ${ELAPSED}s | Completed: $COMPLETED | Running: $RUNNING | Pending: $PENDING | Active Agents: $RUNNING_AGENTS"

    # 如果所有任务完成或没有Agent在运行，退出
    if [ $COMPLETED -ge 10 ] || [ $RUNNING_AGENTS -eq 0 && $PENDING -eq 0 ]; then
        break
    fi
done

echo ""
echo "✅ All agents completed!"
echo ""

# 显示题材选择结果
echo "📋 Selected Topics:"
echo ""
curl -s http://localhost:8848/api/topics | \
    jq -r '.topics[] | "• \(.topic) - selected by \(.worker_id) at \(.selected_at)"'

echo ""
echo "================================"
echo "📊 Test Summary"
echo "================================"
echo ""

# 获取所有任务
TOTAL_TASKS=$(curl -s http://localhost:8848/api/tasks/completed | jq '.data.tasks | length // length' 2>/dev/null || echo 0)

echo "Total tasks completed: $TOTAL_TASKS"

if [ "$TOTAL_TASKS" -ge 10 ]; then
    echo ""
    echo "✅ Test successful!"
    echo ""
    echo "📚 View stories:"
    echo ""
    curl -s http://localhost:8848/api/tasks/completed | \
        jq -r '.data.tasks[] | select(.task_id | startswith("STORY")) | "\(.task_id): \(.title)\n"'
else
    echo "⚠️  Some tasks may have failed"
    echo "   Check agent logs:"
    echo "   for i in {1..10}; do echo \"Agent \$i:\"; tail -20 /tmp/cpds_agent_\$i.log; echo"
fi

echo ""
echo "================================"
echo "✅ Test Complete!"
echo "================================"
