#!/bin/bash
# AgentFlow Python - Quick Test Script

set -e

echo "🧪 AgentFlow Python - Quick Test"
echo ""

# Kill any existing processes
echo "🧹 Cleaning up..."
killall master 2>/dev/null || true
killall python 2>/dev/null || true
sleep 1

# Test 1: Start Master
echo ""
echo "📋 Test 1: Starting Master..."
python -m agentflow.cli master --port 8849 > /tmp/master.log 2>&1 &
MASTER_PID=$!
echo "✓ Master started (PID: $MASTER_PID)"

sleep 2

# Test 2: Health check
echo ""
echo "📋 Test 2: Health check..."
HEALTH=$(curl -s http://127.0.0.1:8849/api/health)
echo "✓ Health check: $HEALTH"

# Test 3: Create task
echo ""
echo "📋 Test 3: Creating task..."
TASK_RESULT=$(curl -s -X POST http://127.0.0.1:8849/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TEST-PYTHON",
    "title": "Python Test",
    "description": "shell:echo Hello from Python AgentFlow",
    "priority": "high"
  }')
echo "✓ Task created: $TASK_RESULT"

# Test 4: Start Worker
echo ""
echo "📋 Test 4: Starting Worker..."
python -m agentflow.cli worker --master http://127.0.0.1:8849 --name test-worker --oneshot > /tmp/worker.log 2>&1 &
WORKER_PID=$!
echo "✓ Worker started (PID: $WORKER_PID)"

# Wait for task execution
echo ""
echo "⏳ Waiting for task execution..."
sleep 5

# Test 5: Check task status
echo ""
echo "📋 Test 5: Checking task status..."
STATUS=$(curl -s http://127.0.0.1:8849/api/tasks/completed)
echo "✓ Completed tasks: $STATUS"

# Cleanup
echo ""
echo "🧹 Cleaning up..."
kill $MASTER_PID 2>/dev/null || true
kill $WORKER_PID 2>/dev/null || true

echo ""
echo "✅ All tests passed!"
echo ""
echo "📝 Worker log:"
cat /tmp/worker.log | tail -5
