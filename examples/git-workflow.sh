#!/bin/bash
# Git Integration Workflow Example
#
# This demonstrates AgentFlow's Git lock mechanism
# for preventing concurrent conflicts

echo "=== Git Integration Workflow Example ==="
echo ""

# Check if we're in a Git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not in a Git repository"
    echo "Create a test repo first:"
    echo "  mkdir test-repo && cd test-repo && git init"
    exit 1
fi

echo "✓ Git repository detected"
echo ""

# Create multiple tasks that work on different files
echo "Creating tasks for concurrent file editing..."
echo ""

# Task 1: Work on app.ts
TASK_1=$(agentflow create "Update app.ts" \
  -d "Add new function to app.ts" \
  2>&1 | grep -o 'TASK-[0-9]*')
echo "✓ Task 1: $TASK_1 (app.ts)"

# Task 2: Work on utils.ts
TASK_2=$(agentflow create "Update utils.ts" \
  -d "Refactor utils.ts" \
  2>&1 | grep -o 'TASK-[0-9]*')
echo "✓ Task 2: $TASK_2 (utils.ts)"

# Task 3: Also try to work on app.ts (will wait for lock)
TASK_3=$(agentflow create "Fix app.ts" \
  -d "Fix bug in app.ts" \
  2>&1 | grep -o 'TASK-[0-9]*')
echo "✓ Task 3: $TASK_3 (app.ts - will wait for lock)"

echo ""
echo "How Git locks work:"
echo "  - Task 1 acquires lock on app.ts ✓"
echo "  - Task 2 works on utils.ts in parallel ✓"
echo "  - Task 3 waits for app.ts lock to release ⏳"
echo ""

echo "View active locks:"
echo "  sqlite3 .claude/cpds-manager/agentflow.db \\"
echo "    \"SELECT * FROM git_locks WHERE status = 'active'\""
echo ""

echo "=== Example Complete ==="
