#!/bin/bash

# AgentFlow 一键启动脚本

# 设置 PATH
export PATH="/opt/homebrew/bin:$PATH"

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
RUST_DIR="$PROJECT_ROOT/rust"
DASHBOARD_DIR="$PROJECT_ROOT/dashboard"

echo "=========================================="
echo "    AgentFlow 一键启动脚本"
echo "=========================================="
echo ""
echo "此脚本将在新的终端窗口中启动："
echo "  1. Rust Master (后端)"
echo "  2. Dashboard (前端)"
echo ""

# 检查依赖
echo "检查依赖..."

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装"
    echo "请运行: brew install node"
    exit 1
fi
echo "✓ Node.js: $(node --version)"

# 检查 Rust（可选）
if command -v cargo &> /dev/null; then
    echo "✓ Rust: $(rustc --version)"
    RUST_AVAILABLE=true
else
    echo "⚠️  Rust 未安装（可选）"
    echo "   如需启动 Master，请安装 Rust"
    RUST_AVAILABLE=false
fi

echo ""
echo "=========================================="

# macOS: 使用 osascript 打开新终端窗口
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "在 macOS 上启动..."

    # 启动 Dashboard
    echo "1. 启动 Dashboard..."
    osascript -e "tell application \"Terminal\"" \
        -e "do script \"cd $DASHBOARD_DIR && ./start-dev.sh\"" \
        -e "end tell" > /dev/null 2>&1 &

    # 如果 Rust 可用，启动 Master
    if [ "$RUST_AVAILABLE" = true ]; then
        echo "2. 启动 Rust Master..."
        sleep 2
        osascript -e "tell application \"Terminal\"" \
            -e "do script \"cd $RUST_DIR && ./start-master.sh\"" \
            -e "end tell" > /dev/null 2>&1 &
    fi

# Linux: 使用 gnome-terminal 或 xterm
elif command -v gnome-terminal &> /dev/null; then
    echo "在 Linux 上启动..."

    # 启动 Dashboard
    gnome-terminal --working-directory="$DASHBOARD_DIR" -- ./start-dev.sh &

    # 如果 Rust 可用，启动 Master
    if [ "$RUST_AVAILABLE" = true ]; then
        gnome-terminal --working-directory="$RUST_DIR" -- ./start-master.sh &
    fi

else
    echo "❌ 不支持的操作系统或缺少终端模拟器"
    echo ""
    echo "请手动启动："
    echo ""
    echo "终端 1 - Dashboard:"
    echo "  cd $DASHBOARD_DIR"
    echo "  ./start-dev.sh"
    echo ""
    if [ "$RUST_AVAILABLE" = true ]; then
        echo "终端 2 - Rust Master:"
        echo "  cd $RUST_DIR"
        echo "  ./start-master.sh"
    fi
    exit 1
fi

echo ""
echo "=========================================="
echo "✓ 启动命令已执行！"
echo ""
echo "Dashboard: http://localhost:5173"
if [ "$RUST_AVAILABLE" = true ]; then
    echo "Master API: http://localhost:6767"
    echo "WebSocket: ws://localhost:8849"
fi
echo ""
echo "请查看新打开的终端窗口"
echo "=========================================="
