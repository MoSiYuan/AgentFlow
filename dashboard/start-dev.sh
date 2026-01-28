#!/bin/bash

# AgentFlow Dashboard 启动脚本

set -e

echo "=== AgentFlow Dashboard 启动脚本 ==="

# 设置 PATH 包含 Homebrew
export PATH="/opt/homebrew/bin:$PATH"

# 检查依赖
echo "1. 检查依赖..."
if ! command -v node &> /dev/null; then
    echo "错误: Node.js 未找到"
    echo "请确保已安装 Node.js: brew install node"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "错误: npm 未找到"
    exit 1
fi

echo "✓ Node.js 版本: $(node --version)"
echo "✓ npm 版本: $(npm --version)"

# 进入 dashboard 目录
cd "$(dirname "$0")"

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo ""
    echo "2. 安装依赖..."
    npm install
fi

# 启动开发服务器
echo ""
echo "3. 启动开发服务器..."
echo "Dashboard 将在 http://localhost:5173 启动"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

npm run dev
