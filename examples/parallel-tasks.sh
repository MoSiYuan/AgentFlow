#!/bin/bash
# Parallel Task Execution Example
#
# This script demonstrates how to create multiple tasks
# that run in parallel using AgentFlow

echo "=== Parallel Task Execution Example ==="
echo ""

# Check Master server
if ! curl -s http://localhost:6767/health > /dev/null 2>&1; then
    echo "Error: Master server not running"
    exit 1
fi

echo "Creating 3 parallel tasks..."
echo ""

# Create multiple tasks in parallel
TASK_1=$(agentflow create "Frontend tests" -d "cd frontend && npm test" 2>&1 | grep -o 'TASK-[0-9]*')
echo "✓ Task 1 created: $TASK_1 (Frontend tests)"

TASK_2=$(agentflow create "Backend tests" -d "cd backend && npm test" 2>&1 | grep -o 'TASK-[0-9]*')
echo "✓ Task 2 created: $TASK_2 (Backend tests)"

TASK_3=$(agentflow create "API tests" -d "cd api && npm test" 2>&1 | grep -o 'TASK-[0-9]*')
echo "✓ Task 3 created: $TASK_3 (API tests)"

echo ""
echo "All tasks created. They will execute in parallel."
echo ""
echo "Monitor progress:"
agentflow list

echo ""
echo "Wait for completion (5 seconds)..."
sleep 5

echo ""
echo "Final status:"
agentflow list --status completed

echo ""
echo "=== Example Complete ==="
