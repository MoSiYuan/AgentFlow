#!/bin/bash

export PATH="/opt/homebrew/Cellar/node@20/20.19.6/bin:/Users/jiangxiaolong/work/project/AgentFlow/dashboard/node_modules/.bin:$PATH"

cd /Users/jiangxiaolong/work/project/AgentFlow/dashboard

echo "Starting dev server..."
npm run dev &
DEV_PID=$!

echo "Waiting for server to start..."
sleep 3

if ps -p $DEV_PID > /dev/null; then
  echo "✅ Dev server started successfully (PID: $DEV_PID)"
  echo "Server is running at: http://localhost:5173"
  kill $DEV_PID
  echo "Server stopped."
  exit 0
else
  echo "❌ Dev server failed to start"
  exit 1
fi
