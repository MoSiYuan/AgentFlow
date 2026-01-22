#!/bin/bash
# CPDS Socket简化测试

set -e

MASTER_URL="http://localhost:8848"
SOCKET_PATH="/tmp/cpds-claude.sock"

echo "==================================="
echo "  CPDS Socket 简化测试"
echo "==================================="
echo ""

# ========================================
# Step 1: 启动Socket服务器
# ========================================
echo "📡 Step 1: 启动Socket服务器..."
rm -f "$SOCKET_PATH"
./bin/socket-server > /tmp/cpds_socket_simple.log 2>&1 &
SOCKET_PID=$!
sleep 2

if [ -S "$SOCKET_PATH" ]; then
  echo "✅ Socket服务器已启动 (PID: $SOCKET_PID)"
else
  echo "❌ Socket服务器启动失败"
  cat /tmp/cpds_socket_simple.log
  exit 1
fi

# ========================================
# Step 2: 测试Socket通信
# ========================================
echo ""
echo "🔌 Step 2: 测试Socket通信..."

echo "REQUEST
简单测试prompt
END_REQUEST" | nc -U "$SOCKET_PATH" > /tmp/socket_response.txt 2>&1 &
sleep 5

if grep -q "RESPONSE" /tmp/socket_response.txt; then
  echo "✅ Socket通信正常"
  echo "响应内容:"
  head -5 /tmp/socket_response.txt
else
  echo "⚠️  Socket响应异常（可能claude未配置）"
fi

# ========================================
# Step 3: 启动Master
# ========================================
echo ""
echo "🌐 Step 3: 启动Master服务器..."
rm -rf .claude/cpds-manager
./cpds/cpds master --mode standalone --auto-shutdown --port 8848 > /tmp/cpds_master_simple.log 2>&1 &
sleep 3

if curl -s "$MASTER_URL/api/health" | grep -q "healthy"; then
  echo "✅ Master服务器已启动"
else
  echo "❌ Master服务器启动失败"
  exit 1
fi

# ========================================
# Step 4: 创建任务
# ========================================
echo ""
echo "📝 Step 4: 创建测试任务..."

TASK_RESPONSE=$(curl -s -X POST "$MASTER_URL/api/tasks/create" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "SIMPLE-001",
    "title": "简化测试任务",
    "description": "你是测试助手，请回复：测试成功。",
    "priority": "high",
    "tags": "[\"test\"]"
  }')

if echo "$TASK_RESPONSE" | grep -q "success.*true"; then
  echo "✅ 任务创建成功"
else
  echo "❌ 任务创建失败: $TASK_RESPONSE"
  echo ""
  echo "📹 Master日志:"
  tail -20 /tmp/cpds_master_simple.log
  exit 1
fi

# ========================================
# Step 5: 启动Worker
# ========================================
echo ""
echo "🤖 Step 5: 启动Worker..."

./cpds/cpds worker --mode standalone \
  --master "$MASTER_URL" \
  --name "Test-Agent" \
  --oneshot > /tmp/worker_simple.log 2>&1 &

WORKER_PID=$!
echo "Worker已启动 (PID: $WORKER_PID)"

# 等待worker完成
echo "⏳ 等待Worker执行任务..."
sleep 10

# ========================================
# Step 6: 检查结果
# ========================================
echo ""
echo "📊 Step 6: 检查执行结果..."

# 检查任务状态
TASK_STATUS=$(sqlite3 .claude/cpds-manager/master.db "SELECT status FROM tasks WHERE task_id='SIMPLE-001';" 2>/dev/null || echo "")
echo "任务状态: $TASK_STATUS"

# 检查worker日志
echo ""
echo "📋 Worker执行日志:"
grep -E "socket|executor|Step|completed|failed" /tmp/worker_simple.log || true

# 检查任务输出
echo ""
echo "📝 任务输出:"
sqlite3 .claude/cpds-manager/master.db "SELECT output FROM tasks WHERE task_id='SIMPLE-001';" 2>/dev/null | head -20 || echo "(未找到输出)"

# ========================================
# Step 7: 清理
# ========================================
echo ""
echo "🧹 Step 7: 清理进程..."
kill $SOCKET_PID 2>/dev/null || true
pkill -f "cpds.*master" 2>/dev/null || true
echo "✅ 清理完成"

# ========================================
# 总结
# ========================================
echo ""
echo "==================================="
echo "  测试总结"
echo "==================================="
echo ""
echo "📁 日志文件:"
echo "  - Socket: /tmp/cpds_socket_simple.log"
echo "  - Master: /tmp/cpds_master_simple.log"
echo "  - Worker: /tmp/worker_simple.log"
echo "  - Socket响应: /tmp/socket_response.txt"
echo ""
echo "🔍 查看完整日志:"
echo "  cat /tmp/cpds_socket_simple.log"
echo "  cat /tmp/worker_simple.log"
echo ""
echo "==================================="
