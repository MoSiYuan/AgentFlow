#!/bin/bash
# quick-task.sh - 快速创建 AgentFlow 任务

if [ $# -lt 2 ]; then
  echo "Usage: $0 <title> <description>"
  echo "Example: $0 'Test Task' 'shell:echo Hello'"
  exit 1
fi

TITLE="$1"
DESC="$2"
TASK_ID="TASK-$(date +%s)"

curl -s -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d "{
    \"task_id\": \"$TASK_ID\",
    \"title\": \"$TITLE\",
    \"description\": \"$DESC\",
    \"priority\": \"high\"
  }" | jq '.'

echo "Task created: $TASK_ID"
