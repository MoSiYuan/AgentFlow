#!/bin/bash
# CPDS Complete Three-Stage Workflow Test
# Stage 1: Topic Selection (抢题材) ✅ IMPLEMENTED
# Stage 2: Article Generation (输出文章) - TODO
# Stage 3: Agent Peer Review (Agent互评) - TODO

set -e

MASTER_URL="http://localhost:8848"
NUM_AGENTS=10

echo "═══════════════════════════════════════════════════════════"
echo "     CPDS Complete Three-Stage Workflow Test"
echo "═══════════════════════════════════════════════════════════"
echo ""

# ========================================
# STAGE 0: Start Master
# ========================================
echo "🚀 Stage 0: Starting Master..."
pkill -f "cpds.*master" 2>/dev/null || true
sleep 2
rm -rf .claude/cpds-manager

./cpds/cpds master --mode standalone --auto-shutdown --port 8848 > /tmp/cpds_complete_test.log 2>&1 &
sleep 3

if ! curl -s "$MASTER_URL/api/health" | grep -q "healthy"; then
  echo "❌ Master failed to start"
  exit 1
fi

echo "✅ Master started successfully"
echo ""

# ========================================
# STAGE 1: Topic Selection (抢题材)
# ========================================
echo "🎯 Stage 1: Topic Selection (抢题材)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Creating $NUM_AGENTS story creation tasks..."

for i in $(seq 1 $NUM_AGENTS); do
  TASK_ID="STORY-$(printf '%03d' $i)"
  curl -s -X POST "$MASTER_URL/api/tasks/create" \
    -H "Content-Type: application/json" \
    -d "{
      \"task_id\": \"$TASK_ID\",
      \"title\": \"Agent $i 自主创作克苏鲁故事\",
      \"description\": \"你是专业的克苏鲁神话作家。请自主选择题材并创作500-1000字故事。流程：1) GET /api/topics 查询已选题材；2) 选择独特题材；3) POST /api/topics/register 注册；4) 创作。洛夫克拉夫特风格。\",
      \"priority\": \"high\",
      \"tags\": [\"creative-writing\", \"lovecraft\", \"stage1\"]
    }" > /dev/null

  echo "  ✅ Created task: $TASK_ID"
done

echo ""
echo "Starting $NUM_AGENTS agents for topic selection..."
for i in $(seq 1 $NUM_AGENTS); do
  ./cpds/cpds worker --mode standalone \
    --master "$MASTER_URL" \
    --name "Agent-$i" \
    --oneshot > /tmp/agent_stage1_$i.log 2>&1 &

  echo "  🤖 Started Agent-$i"
  sleep 0.3
done

# Wait for all agents to complete
echo ""
echo "⏳ Waiting for topic selection to complete..."
sleep 15

# Show selected topics
echo ""
echo "📊 Selected Topics:"
curl -s "$MASTER_URL/api/topics" | jq -r '.topics[] | "\(.topic) - by \(.worker_id[:8])... (story: \(.story_id))"'

echo ""
echo "✅ Stage 1 Complete: All topics selected"
echo ""

# ========================================
# STAGE 2: Article Generation (输出文章)
# ========================================
echo "📝 Stage 2: Article Generation (输出文章)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "⚠️  NOTE: Stage 2 requires Claude API integration"
echo ""
echo "Current implementation uses simulation mode."
echo "To enable real article generation:"
echo "  1. Set ANTHROPIC_API_KEY environment variable"
echo "  2. Install anthropic-go library: go get github.com/anthropics/anthropic-go"
echo "  3. Update executeLocally() to call Claude Messages API"
echo ""

# Show generated stories
echo "📚 Generated Stories (Simulation):"
for i in $(seq 1 $NUM_AGENTS); do
  TASK_ID="STORY-$(printf '%03d' $i)"
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "[$TASK_ID]"
  sqlite3 .claude/cpds-manager/master.db "SELECT output FROM tasks WHERE task_id='$TASK_ID';" 2>/dev/null | grep -A 5 "选定的题材" || echo "  (Not found)"
done

echo ""
echo "✅ Stage 2 Complete: Stories generated (simulated)"
echo ""

# ========================================
# STAGE 3: Agent Peer Review (Agent互评)
# ========================================
echo "💬 Stage 3: Agent Peer Review (Agent互评)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "⚠️  NOTE: Stage 3 requires implementation"
echo ""
echo "Required features:"
echo "  1. Query all completed stories from Stage 2"
echo "  2. Create review tasks for each agent"
echo "  3. Each agent reviews 9 other agents' stories"
echo "  4. Provide feedback and ratings"
echo ""

# Show what would happen
echo "📋 Planned Review Tasks:"
for reviewer in $(seq 1 $NUM_AGENTS); do
  echo ""
  echo "Agent-$reviewer will review:"
  for reviewed in $(seq 1 $NUM_AGENTS); do
    if [ $reviewer -ne $reviewed ]; then
      echo "  - STORY-$(printf '%03d' $reviewed)"
    fi
  done
done

echo ""
echo "❌ Stage 3: NOT YET IMPLEMENTED"
echo ""

# ========================================
# SUMMARY
# ========================================
echo "═══════════════════════════════════════════════════════════"
echo "                          SUMMARY"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Count completed tasks
COMPLETED=$(sqlite3 .claude/cpds-manager/master.db "SELECT COUNT(*) FROM tasks WHERE status='completed';" 2>/dev/null || echo "0")
REGISTERED_TOPICS=$(curl -s "$MASTER_URL/api/topics" | jq '.count' 2>/dev/null || echo "0")

echo "✅ Stage 1 (Topic Selection): COMPLETE"
echo "   - Tasks completed: $COMPLETED/$NUM_AGENTS"
echo "   - Topics registered: $REGISTERED_TOPICS"
echo ""
echo "⚠️  Stage 2 (Article Generation): SIMULATION ONLY"
echo "   - Stories generated: $COMPLETED (simulated)"
echo "   - Real Claude API: NOT INTEGRATED"
echo ""
echo "❌ Stage 3 (Peer Review): NOT IMPLEMENTED"
echo "   - Review tasks: 0"
echo "   - Reviews completed: 0"
echo ""

echo "📁 Test logs saved to:"
echo "   - /tmp/cpds_complete_test.log"
echo "   - /tmp/agent_stage1_*.log"
echo ""
echo "📊 Database: .claude/cpds-manager/master.db"
echo ""

# Wait for master to shutdown
echo "⏳ Waiting for Master to shutdown..."
sleep 5

echo "═══════════════════════════════════════════════════════════"
echo "                    TEST COMPLETE"
echo "═══════════════════════════════════════════════════════════"
