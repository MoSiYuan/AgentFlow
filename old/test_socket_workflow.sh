#!/bin/bash
# CPDS Socket工作流完整测试

set -e

MASTER_URL="http://localhost:8848"
SOCKET_SERVER="/Users/jiangxiaolong/work/project/game/AdStar/cpds-go/bin/socket-server"
SOCKET_PATH="/tmp/cpds-claude.sock"

echo "═══════════════════════════════════════════════════════════"
echo "     CPDS Socket工作流测试 - 完整三阶段"
echo "═══════════════════════════════════════════════════════════"
echo ""

# ========================================
# Stage 0: 启动Socket服务器
# ========================================
echo "🚀 Stage 0: 启动Socket服务器"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 清理旧的socket文件
rm -f "$SOCKET_PATH"

# 启动socket服务器
echo "启动Socket服务器..."
nohup "$SOCKET_SERVER" > /tmp/cpds_socket_test.log 2>&1 &
SOCKET_PID=$!
sleep 2

# 验证socket是否创建
if [ -S "$SOCKET_PATH" ]; then
  echo "✅ Socket服务器已启动 (PID: $SOCKET_PID)"
  echo "   Socket: $SOCKET_PATH"
else
  echo "❌ Socket服务器启动失败"
  cat /tmp/cpds_socket_test.log
  exit 1
fi

echo ""

# ========================================
# Stage 1: 启动Master
# ========================================
echo "🌐 Stage 1: 启动Master服务器"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

pkill -f "cpds.*master" 2>/dev/null || true
sleep 2
rm -rf .claude/cpds-manager

./cpds/cpds master --mode standalone --auto-shutdown --port 8848 > /tmp/cpds_master_socket.log 2>&1 &
sleep 3

if curl -s "$MASTER_URL/api/health" | grep -q "healthy"; then
  echo "✅ Master服务器已启动"
else
  echo "❌ Master服务器启动失败"
  exit 1
fi

echo ""

# ========================================
# Stage 2: 抢题材
# ========================================
echo "🎯 Stage 2: 抢题材测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "创建3个任务..."
for i in {1..3}; do
  TASK_ID="SOCKET-TEST-$(printf '%03d' $i)"
  curl -s -X POST "$MASTER_URL/api/tasks/create" \
    -H "Content-Type: application/json" \
    -d "{
      \"task_id\": \"$TASK_ID\",
      \"title\": \"Socket测试任务 $i\",
      \"description\": \"你是克苏鲁神话作家。请自主选择题材并创作500-1000字故事。\",
      \"priority\": \"high\",
      \"tags\": \"[\\\"creative-writing\\\", \\\"socket-test\\\"]\"
    }" > /dev/null

  echo "  ✅ $TASK_ID"
done

echo ""
echo "启动3个Agent（通过Socket执行）..."
for i in {1..3}; do
  ./cpds/cpds worker --mode standalone \
    --master "$MASTER_URL" \
    --name "Socket-Agent-$i" \
    --oneshot > /tmp/socket_agent_$i.log 2>&1 &

  echo "  🤖 Socket-Agent-$i (PID: $!)"
  sleep 0.5
done

# 等待任务完成
echo ""
echo "⏳ 等待任务完成（通过Socket调用Claude）..."
echo "   注意：首次调用可能需要1-2分钟..."
sleep 15

echo ""
echo "📊 执行结果："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 检查任务状态
COMPLETED=$(sqlite3 .claude/cpds-manager/master.db "SELECT COUNT(*) FROM tasks WHERE status='completed';" 2>/dev/null || echo "0")
PENDING=$(sqlite3 .claude/cpds-manager/master.db "SELECT COUNT(*) FROM tasks WHERE status='pending';" 2>/dev/null || echo "0")

echo "已完成: $COMPLETED"
echo "待处理: $PENDING"

echo ""
echo "📋 选定的题材："
curl -s "$MASTER_URL/api/topics" | jq -r '.topics[] | "  - \(.topic)"' 2>/dev/null || echo "  (无)"

echo ""
echo "📚 生成的故事（预览）："
for i in {1..3}; do
  TASK_ID="SOCKET-TEST-$(printf '%03d' $i)"
  echo ""
  echo "[$TASK_ID]"
  sqlite3 .claude/cpds-manager/master.db "SELECT output FROM tasks WHERE task_id='$TASK_ID';" 2>/dev/null | head -20
  echo "..."
done

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "                          测试总结"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "✅ Socket服务器: 运行中 (PID: $SOCKET_PID)"
echo "✅ Master服务器: 运行中"
echo "✅ Worker: 通过Socket调用Claude"
echo ""
echo "📁 日志文件:"
echo "   - Socket服务器: /tmp/cpds_socket_test.log"
echo "   - Master: /tmp/cpds_master_socket.log"
echo "   - Workers: /tmp/socket_agent_*.log"
echo ""
echo "🔍 查看Socket日志:"
echo "   tail -f /tmp/cpds_socket_test.log"
echo ""
echo "🛑 停止服务:"
echo "   kill $SOCKET_PID  # Socket服务器"
echo "   pkill -f 'cpds.*master'  # Master服务器"
echo ""
echo "═══════════════════════════════════════════════════════════"
