#!/bin/bash
# AgentFlow Rust v3 快速启动脚本

set -e

echo "🚀 AgentFlow v3 (Pure Rust) 快速启动"
echo "======================================"
echo ""

# 检查 Rust 是否安装
if ! command -v cargo &> /dev/null; then
    echo "❌ Cargo 未安装，请先安装 Rust:"
    echo "   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
    exit 1
fi

echo "✅ Rust 工具链已安装"
cargo --version
echo ""

# 进入 Rust 项目目录
cd "$(dirname "$0")/rust"

# 创建 .env 文件（如果不存在）
if [ ! -f .env ]; then
    echo "📝 创建 .env 配置文件..."
    cat > .env << 'EOF'
# AgentFlow Master 环境变量配置
AGENTFLOW_PORT=6767
AGENTFLOW_HOST=0.0.0.0
DATABASE_URL=sqlite://agentflow.db
RUST_LOG=info
SQLX_OFFLINE=true
EOF
fi

# 加载环境变量
set -a
source .env
set +a

echo "🔨 编译项目..."
if [ "$1" == "--release" ]; then
    cargo build --release
    echo ""
    echo "✅ 编译完成 (Release 模式)"
    echo ""
    echo "🚀 启动服务器..."
    ./target/release/agentflow-master
else
    cargo build
    echo ""
    echo "✅ 编译完成 (Debug 模式)"
    echo ""
    echo "🚀 启动服务器..."
    cargo run --bin agentflow-master
fi
