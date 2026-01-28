#!/bin/bash

# AgentFlow Rust Master 启动脚本

set -e

echo "=== AgentFlow Master 启动脚本 ==="

# 检查 Rust 工具链
echo "1. 检查 Rust 工具链..."
if ! command -v cargo &> /dev/null; then
    echo "错误: Cargo 未找到"
    echo "请安装 Rust: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
    echo "或访问: https://rustup.rs/"
    exit 1
fi

echo "✓ Rust 版本: $(rustc --version)"
echo "✓ Cargo 版本: $(cargo --version)"

# 进入 rust 目录
cd "$(dirname "$0")"

# 检查是否已编译
if [ ! -f "target/release/agentflow-master" ]; then
    echo ""
    echo "2. 编译 AgentFlow Master..."
    cargo build --release
fi

# 启动 Master
echo ""
echo "3. 启动 AgentFlow Master..."
echo "Master 将在 http://localhost:6767 启动"
echo ""
echo "配置:"
echo "  - 节点 ID: ${NODE_ID:-master-1}"
echo "  - 端口: ${PORT:-6767}"
echo "  - Peers: ${PEERS:-master-1:6767,master-2:6768,master-3:6769}"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

# 运行 Master
exec cargo run --bin agentflow-master -- \
    --node-id "${NODE_ID:-master-1}" \
    --port "${PORT:-6767}" \
    --peers "${PEERS:-master-1:6767,master-2:6768,master-3:6769}"
