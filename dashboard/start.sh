#!/bin/bash

# AgentFlow Dashboard 启动脚本

# 设置正确的 PATH
export PATH="/opt/homebrew/Cellar/node/25.5.0/bin:/opt/homebrew/bin:$PATH"

# 进入脚本所在目录
cd "$(dirname "$0")"

echo "=== AgentFlow Dashboard 启动 ==="
echo "Node.js: $(node --version)"
echo "npm: $(npm --version)"
echo ""

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo "安装依赖..."
    npm install
fi

echo "启动开发服务器..."
echo "访问: http://localhost:5173"
echo ""

npm run dev
