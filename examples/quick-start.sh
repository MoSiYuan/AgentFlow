#!/bin/bash
# AgentFlow Quick Start Example
#
# This script demonstrates basic AgentFlow usage

echo "=== AgentFlow Quick Start ==="
echo ""

# 1. Check if AgentFlow is installed
echo "1. Checking AgentFlow installation..."
if command -v agentflow &> /dev/null; then
    echo "   ✓ AgentFlow is installed"
else
    echo "   ✗ AgentFlow not found"
    echo "   Install: npm install -g @agentflow/skill"
    exit 1
fi

# 2. Check Master server
echo ""
echo "2. Checking Master server..."
if curl -s http://localhost:6767/health > /dev/null 2>&1; then
    echo "   ✓ Master server is running"
else
    echo "   ✗ Master server is not running"
    echo "   Start: cd /path/to/AgentFlow/nodejs && node packages/master/dist/index.js"
    exit 1
fi

# 3. Create a simple task
echo ""
echo "3. Creating a test task..."
TASK_ID=$(agentflow create "Test task" -d "echo 'Hello AgentFlow'" 2>&1 | grep -o 'TASK-[0-9]*' | head -1)
echo "   ✓ Task created: $TASK_ID"

# 4. List all tasks
echo ""
echo "4. Listing tasks..."
agentflow list

# 5. Wait for task completion
echo ""
echo "5. Waiting for task completion..."
sleep 2

# 6. Check task status
echo ""
echo "6. Checking task status..."
agentflow status $TASK_ID

echo ""
echo "=== Quick Start Complete ==="
