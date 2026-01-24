#!/bin/bash
# AgentFlow Node.js 版本设置脚本
# 自动安装和配置 Node.js 20 LTS + better-sqlite3

set -e

echo "========================================="
echo "AgentFlow Node.js 环境配置"
echo "========================================="
echo ""

# 检测操作系统
OS="$(uname -s)"
case "${OS}" in
    Linux*)     MACHINE=Linux;;
    Darwin*)    MACHINE=Mac;;
    MINGW*)     MACHINE=Windows;;
    *)          MACHINE="UNKNOWN:${OS}"
esac

echo "检测到操作系统: $MACHINE"
echo ""

# 检查 Node.js
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo "当前 Node.js 版本: $NODE_VERSION"

    # 检查版本是否合适
    MAJOR_VERSION=$(echo $NODE_VERSION | cut -d'.' -f1 | sed 's/v//')

    if [ "$MAJOR_VERSION" -ge 24 ]; then
        echo "⚠️  警告: Node.js v24 与 better-sqlite3 不兼容"
        echo "建议: 使用 Node.js 18-20 LTS"
        echo ""
        read -p "是否继续安装? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "已取消安装"
            exit 1
        fi
    fi
else
    echo "❌ 未检测到 Node.js"
    echo ""
    echo "请安装 Node.js:"
    echo "  macOS:   brew install node@20"
    echo "  Ubuntu:  curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -"
    echo "           sudo apt-get install -y nodejs"
    echo ""
    exit 1
fi

echo ""
echo "========================================="
echo "开始安装依赖..."
echo "========================================="
echo ""

# 进入 Node.js 目录
if [ ! -d "nodejs" ]; then
    echo "❌ 错误: 未找到 nodejs 目录"
    echo "请在 AgentFlow 根目录运行此脚本"
    exit 1
fi

cd nodejs

# 检查包管理器
if command -v pnpm &> /dev/null; then
    PKG_MANAGER="pnpm"
    echo "使用包管理器: pnpm"
elif command -v npm &> /dev/null; then
    PKG_MANAGER="npm"
    echo "使用包管理器: npm"
else
    echo "❌ 错误: 未找到 pnpm 或 npm"
    exit 1
fi

echo ""
echo "安装依赖..."
$PKG_MANAGER install

echo ""
echo "========================================="
echo "构建项目..."
echo "========================================="
echo ""

$PKG_MANAGER run build

echo ""
echo "========================================="
echo "验证安装..."
echo "========================================="
echo ""

# 检查 better-sqlite3
echo "检查 better-sqlite3..."
node -e "const Database = require('better-sqlite3'); console.log('✓ better-sqlite3 版本:', Database.Database.prototype.constructor.name);" 2>&1

if [ $? -eq 0 ]; then
    echo "✓ better-sqlite3 安装成功"
else
    echo "❌ better-sqlite3 加载失败"
    echo ""
    echo "可能的原因:"
    echo "  1. Node.js 版本不兼容（需要 v18-20）"
    echo "  2. 缺少 C++ 编译工具"
    echo "  3. better-sqlite3 原生模块未编译"
    echo ""
    echo "解决方案:"
    echo "  - 降级到 Node.js 20 LTS"
    echo "  - 或使用 Go 版本（推荐）"
    exit 1
fi

echo ""
echo "========================================="
echo "测试 Master 启动..."
echo "========================================="
echo ""

# 创建测试数据库目录
mkdir -p ../test-data

# 启动 Master（后台）
echo "启动 Master 服务..."
node packages/master/dist/index.js --port 6767 --db ../test-data/test.db &
MASTER_PID=$!

# 等待启动
sleep 3

# 测试健康检查
echo "测试健康检查..."
if curl -s http://localhost:6767/health > /dev/null; then
    echo "✓ Master 启动成功"

    # 停止 Master
    kill $MASTER_PID 2>/dev/null || true
    echo "✓ 测试完成"
else
    echo "❌ Master 启动失败"
    kill $MASTER_PID 2>/dev/null || true
    exit 1
fi

echo ""
echo "========================================="
echo "安装完成！"
echo "========================================="
echo ""
echo "快速开始:"
echo ""
echo "  # 启动 Master"
echo "  cd nodejs"
echo "  node packages/master/dist/index.js"
echo ""
echo "  # 使用 CLI"
echo "  npm link"
echo "  agentflow create '{\"title\":\"测试\",\"detail\":\"echo hello\"}'"
echo ""
echo "  # 或使用 Go 版本（推荐）"
echo "  ./agentflow-go.sh run '[\"echo hello\",\"echo world\"]'"
echo ""
