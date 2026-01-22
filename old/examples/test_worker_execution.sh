#!/bin/bash
# 测试Worker执行方式的示例脚本

set -e

echo "==================================="
echo "CPDS Worker 执行方式测试"
echo "==================================="
echo ""

# 1. 检查claudecli是否安装
echo "1. 检查claudecli..."
if command -v claudecli &> /dev/null; then
    echo "✅ claudecli found: $(which claudecli)"
    echo "   版本: $(claudecli --version 2>&1 || echo 'unknown')"
else
    echo "❌ claudecli not found"
    echo "   请安装claudecli: npm install -g @anthropic-ai/claude-cli"
    echo "   将使用本地模拟执行"
fi
echo ""

# 2. 测试Master启动
echo "2. 启动Master (standalone模式)..."
./cpds master \
  --mode standalone \
  --auto-shutdown \
  --host localhost \
  --port 8848 \
  > /tmp/cpds-master.log 2>&1 &

MASTER_PID=$!
echo "   Master PID: $MASTER_PID"

# 等待Master启动
sleep 2

# 检查Master是否启动成功
if curl -s http://localhost:8848/api/health > /dev/null; then
    echo "   ✅ Master started successfully"
else
    echo "   ❌ Master failed to start"
    cat /tmp/cpds-master.log
    exit 1
fi
echo ""

# 3. 创建测试任务
echo "3. 创建测试任务..."
TASK_RESPONSE=$(curl -s -X POST http://localhost:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TEST-001",
    "title": "测试Claude执行",
    "description": "使用Go语言实现一个快速排序算法",
    "priority": "high"
  }')

echo "   任务创建响应: $TASK_RESPONSE"
echo ""

# 4. 启动Worker
echo "4. 启动Worker..."
./cpds worker \
  --mode standalone \
  --master http://localhost:8848 \
  --oneshot \
  > /tmp/cpds-worker.log 2>&1 &

WORKER_PID=$!
echo "   Worker PID: $WORKER_PID"
echo ""

# 5. 等待Worker完成
echo "5. 等待Worker执行任务..."
sleep 5

# 6. 检查任务状态
echo "6. 检查任务状态..."
TASK_STATUS=$(curl -s http://localhost:8848/api/tasks/TEST-001)
echo "   任务状态: $TASK_STATUS"
echo ""

# 7. 等待Master自动关闭（如果没有任务了）
echo "7. 等待Master自动关闭..."
sleep 10

# 检查进程状态
if ps -p $MASTER_PID > /dev/null; then
    echo "   Master仍在运行 (可能还有其他任务)"
    kill $MASTER_PID 2>/dev/null || true
else
    echo "   ✅ Master已自动关闭"
fi

if ps -p $WORKER_PID > /dev/null; then
    echo "   Worker仍在运行"
    kill $WORKER_PID 2>/dev/null || true
else
    echo "   ✅ Worker已完成并退出 (oneshot模式)"
fi
echo ""

# 8. 显示日志
echo "8. Worker执行日志:"
echo "==================================="
cat /tmp/cpds-worker.log || true
echo ""
echo "Master日志:"
echo "==================================="
cat /tmp/cpds-master.log || true
echo ""

echo "==================================="
echo "测试完成！"
echo "==================================="
echo ""
echo "说明:"
echo "- 如果安装了claudecli，Worker会使用claudecli执行任务"
echo "- 否则，Worker会降级到本地模拟执行"
echo "- Worker在standalone模式下执行完一个任务后自动退出"
echo "- Master在auto-shutdown模式下，所有任务完成后自动关闭"
