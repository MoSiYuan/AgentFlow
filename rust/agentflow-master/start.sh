#!/bin/bash
# AgentFlow Master 启动脚本

set -e

echo "🚀 启动 AgentFlow Master 服务器"
echo "================================"

# 检查是否安装了 Rust
if ! command -v cargo &> /dev/null; then
    echo "❌ 错误: 未找到 cargo，请先安装 Rust"
    echo "   访问 https://rustup.rs/ 安装 Rust"
    exit 1
fi

# 检查端口是否被占用
PORT=${AGENTFLOW_SERVER_PORT:-6767}
if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "⚠️  警告: 端口 $PORT 已被占用"
    echo "   请检查是否有其他实例正在运行"
    echo "   或设置环境变量 AGENTFLOW_SERVER_PORT 使用其他端口"
    exit 1
fi

# 创建工作目录
WORKSPACE=${AGENTFLOW_SANDBOX_WORKSPACE:-/tmp/agentflow/workspace}
mkdir -p "$WORKSPACE"
echo "📁 工作目录: $WORKSPACE"

# 加载 .env 文件
if [ -f .env ]; then
    echo "📋 加载 .env 配置"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "⚠️  未找到 .env 文件，使用默认配置"
fi

# 构建并运行
echo "🔨 构建 AgentFlow Master..."
cargo build --release --bin agentflow-master

echo ""
echo "✅ 构建完成"
echo ""
echo "🌐 服务器配置:"
echo "   - 地址: ${AGENTFLOW_SERVER_ADDR:-0.0.0.0}"
echo "   - 端口: ${AGENTFLOW_SERVER_PORT:-6767}"
echo "   - 数据库: ${AGENTFLOW_DATABASE_URL:-sqlite://agentflow.db}"
echo "   - 最大并发: ${AGENTFLOW_MAX_CONCURRENT_TASKS:-10}"
echo ""
echo "🌟 启动服务器..."
echo ""

# 运行服务器
cargo run --release --bin agentflow-master
