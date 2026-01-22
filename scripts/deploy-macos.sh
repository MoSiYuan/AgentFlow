#!/bin/bash
# AgentFlow macOS 快速部署脚本
# 用于快速部署 Claude 开发环境

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_header() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查 macOS 环境
check_macos_environment() {
    print_header "检查 macOS 环境"

    # 检查 macOS 版本
    if [[ "$OSTYPE" == "darwin"* ]]; then
        MACOS_VERSION=$(sw_vers -productVersion)
        print_success "macOS 版本: $MACOS_VERSION"
    else
        print_error "此脚本仅支持 macOS"
        exit 1
    fi

    # 检查 Homebrew
    if command -v brew &> /dev/null; then
        print_success "Homebrew 已安装"
    else
        print_warning "Homebrew 未安装"
        print_info "安装 Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        print_success "Homebrew 安装完成"
    fi
}

# 安装 Claude CLI
install_claude_cli() {
    print_header "安装 Claude CLI"

    if command -v claudecli &> /dev/null; then
        CLAUDE_VERSION=$(claudecli --version 2>/dev/null || echo "unknown")
        print_success "Claude CLI 已安装: $CLAUDE_VERSION"
        print_info "Claude CLI 路径: $(which claudecli)"
    else
        print_info "Claude CLI 未安装，开始安装..."

        # 使用 Homebrew 安装
        if brew list claudecli &> /dev/null; then
            print_info "Claude CLI 已通过 Homebrew 安装"
        else
            print_warning "Claude CLI 不在 Homebrew 中，请手动安装"
            print_info "安装方式 1: npm install -g @anthropic-ai/claude-cli"
            print_info "安装方式 2: 从源码编译"
            print_info "下载地址: https://github.com/anthropics/claude-cli"
            read -p "按回车键继续手动安装..."
        fi
    fi
}

# 安装 Python 依赖
install_python_dependencies() {
    print_header "安装 Python 依赖"

    # 检查 Python 3
    if command -v python3 &> /dev/null; then
        PYTHON_VERSION=$(python3 --version)
        print_success "Python 3 已安装: $PYTHON_VERSION"
    else
        print_warning "Python 3 未安装"
        print_info "通过 Homebrew 安装 Python 3..."
        brew install python@3.11
    fi

    # 安装 Python 依赖
    if [ -d "python" ]; then
        print_info "安装 Python 依赖..."
        cd python
        pip3 install -r requirements.txt || print_warning "部分依赖安装失败"
        cd ..
        print_success "Python 依赖安装完成"
    fi
}

# 安装 Go 依赖
install_go_dependencies() {
    print_header "安装 Go 依赖"

    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_success "Go 已安装: $GO_VERSION"

        # 设置 Go 代理（中国）
        export GOPROXY=https://goproxy.cn,direct
        print_info "Go 代理已设置: $GOPROXY"

        # 安装依赖
        if [ -d "golang" ]; then
            print_info "安装 Go 模块依赖..."
            cd golang
            go mod tidy
            cd ..
            print_success "Go 依赖安装完成"
        fi
    else
        print_warning "Go 未安装"
        print_info "通过 Homebrew 安装 Go..."
        brew install go
    fi
}

# 编译 Go 二进制文件
build_go_binaries() {
    print_header "编译 Go 二进制文件"

    if [ ! -d "golang" ]; then
        print_error "golang 目录不存在"
        exit 1
    fi

    cd golang

    # 创建输出目录
    mkdir -p bin

    # 编译 master
    print_info "编译 master..."
    go build -o bin/master cmd/master/main.go

    # 编译 worker
    print_info "编译 worker..."
    go build -o bin/worker cmd/worker/main.go

    # 编译 oneshot
    print_info "编译 oneshot..."
    go build -o bin/oneshot cmd/oneshot/main.go

    cd ..

    print_success "编译完成"
    print_info "二进制文件位置:"
    print_info "  - golang/bin/master"
    print_info "  - golang/bin/worker"
    print_info "  - golang/bin/oneshot"
}

# 配置文件边界
setup_file_boundaries() {
    print_header "配置文件边界"

    BOUNDARIES_FILE=".agentflow/boundaries.json"

    if [ -f "$BOUNDARIES_FILE" ]; then
        print_success "文件边界配置已存在: $BOUNDARIES_FILE"
    else
        print_info "创建文件边界配置..."
        mkdir -p .agentflow
        cp .agentflow/boundaries.example.json "$BOUNDARIES_FILE" 2>/dev/null || true

        # 如果示例文件不存在，创建默认配置
        if [ ! -f "$BOUNDARIES_FILE" ]; then
            cat > "$BOUNDARIES_FILE" << 'EOF'
{
  "agent-frontend": [
    {
      "file_pattern": "src/frontend/**/*",
      "access_type": "exclusive",
      "description": "Frontend agent manages frontend UI code"
    },
    {
      "file_pattern": "src/api/**/*",
      "access_type": "readonly",
      "description": "Frontend agent can read API definitions"
    }
  ],
  "agent-backend": [
    {
      "file_pattern": "src/backend/**/*",
      "access_type": "exclusive",
      "description": "Backend agent manages backend code"
    },
    {
      "file_pattern": "src/api/**/*",
      "access_type": "shared",
      "description": "Backend agent shares API files"
    }
  ],
  "agent-database": [
    {
      "file_pattern": "src/database/**/*",
      "access_type": "exclusive",
      "description": "Database agent manages database layer"
    }
  ]
}
EOF
        fi

        print_success "文件边界配置已创建: $BOUNDARIES_FILE"
    fi
}

# 启动 Master 服务
start_master() {
    print_header "启动 Master 服务"

    # 检查是否已运行
    if [ -f "/tmp/agentflow-master.pid" ]; then
        MASTER_PID=$(cat /tmp/agentflow-master.pid)
        if ps -p $MASTER_PID > /dev/null 2>&1; then
            print_warning "Master 已在运行 (PID: $MASTER_PID)"
            return
        fi
    fi

    # 启动 Master
    if [ -f "golang/bin/master" ]; then
        print_info "启动 Master (Go 版本)..."
        golang/bin/master -config golang/config.example.yaml &
        MASTER_PID=$!
        echo $MASTER_PID > /tmp/agentflow-master.pid
        print_success "Master 已启动 (PID: $MASTER_PID)"
        print_info "Master 地址: http://localhost:8848"
    else
        print_warning "未找到编译好的 master，尝试 Python 版本..."
        if [ -f "python/agentflow/__init__.py" ]; then
            cd python
            python3 -m agentflow.cli master --port 8848 &
            MASTER_PID=$!
            echo $MASTER_PID > /tmp/agentflow-master.pid
            cd ..
            print_success "Master 已启动 (PID: $MASTER_PID)"
        else
            print_error "未找到可执行的 master"
        fi
    fi
}

# 启动 Worker
start_worker() {
    print_header "启动 Worker"

    WORKER_GROUP="${1:-default}"

    if [ -f "golang/bin/worker" ]; then
        print_info "启动 Worker (Go 版本, 组: $WORKER_GROUP)..."
        golang/bin/worker -config golang/config.example.yaml &
        WORKER_PID=$!
        print_success "Worker 已启动 (PID: $WORKER_PID, 组: $WORKER_GROUP)"
    else
        print_warning "未找到编译好的 worker，尝试 Python 版本..."
        cd python
        python3 -m agentflow.cli worker --group "$WORKER_GROUP" --auto &
        WORKER_PID=$!
        cd ..
        print_success "Worker 已启动 (PID: $WORKER_PID, 组: $WORKER_GROUP)"
    fi
}

# 显示使用指南
show_usage() {
    cat << EOF

${GREEN}AgentFlow macOS 快速部署完成！${NC}

📝 常用命令:
  ${YELLOW}# 查看系统状态${NC}
  ./scripts/deploy-macos.sh status

  ${YELLOW}# 创建任务${NC}
  ./scripts/quick-task.sh "测试任务" "shell:echo Hello World"

  ${YELLOW}# 停止服务${NC}
  ./scripts/deploy-macos.sh stop

  ${YELLOW}# 查看日志${NC}
  tail -f /tmp/agentflow-*.log

🔗 服务地址:
  - Master API: http://localhost:8848
  - API 文档: http://localhost:8848/docs

📚 文档:
  - 完整指南: docs/git-integration-guide.md
  - README.md: README.md
  - Skill 手册: skills/agentflow.md

🎯 下一步:
   1. 访问 Master API: http://localhost:8848
   2. 创建第一个任务测试系统
  3. 配置文件边界: .agentflow/boundaries.json
  4. 启动多个 Workers 并行处理

EOF
}

# 检查依赖
check() {
    check_macos_environment
    echo ""
}

# 安装所有依赖
install() {
    check_macos_environment
    install_claude_cli
    install_python_dependencies
    install_go_dependencies
    build_go_binaries
    setup_file_boundaries
    echo ""
    print_success "所有依赖已安装！"
}

# 完整部署
deploy() {
    print_header "AgentFlow macOS 快速部署"

    check
    install
    start_master
    start_worker "default"

    show_usage
}

# 停止服务
stop() {
    print_header "停止服务"

    if [ -f "/tmp/agentflow-master.pid" ]; then
        MASTER_PID=$(cat /tmp/agentflow-master.pid)
        if ps -p $MASTER_PID > /dev/null 2>&1; then
            print_info "停止 Master (PID: $MASTER_PID)..."
            kill $MASTER_PID 2>/dev/null || true
            rm -f /tmp/agentflow-master.pid
            print_success "Master 已停止"
        fi
    fi

    # 停止所有 Workers
    pkill -f "agentflow.*worker" 2>/dev/null && print_success "Workers 已停止" || true
}

# 状态检查
status() {
    print_header "系统状态"

    # 检查 Master
    if [ -f "/tmp/agentflow-master.pid" ]; then
        MASTER_PID=$(cat /tmp/agentflow-master.pid)
        if ps -p $MASTER_PID > /dev/null 2>&1; then
            print_success "Master 运行中 (PID: $MASTER_PID)"
        else
            print_warning "Master 未运行"
        fi
    else
        print_warning "Master 未运行"
    fi

    # 检查 Workers
    WORKER_COUNT=$(pgrep -f "agentflow.*worker" | wc -l)
    if [ $WORKER_COUNT -gt 0 ]; then
        print_success "Workers 运行中: $WORKER_COUNT 个"
    else
        print_warning "没有运行的 Workers"
    fi

    # 检查二进制文件
    if [ -f "golang/bin/master" ]; then
        print_success "Go master: golang/bin/master"
    fi
    if [ -f "golang/bin/worker" ]; then
        print_success "Go worker: golang/bin/worker"
    fi
    if [ -f "golang/bin/oneshot" ]; then
        print_success "Go oneshot: golang/bin/oneshot"
    fi

    # 检查配置文件
    if [ -f ".agentflow/boundaries.json" ]; then
        print_success "文件边界配置: .agentflow/boundaries.json"
    fi
}

# 主函数
main() {
    case "${1:-deploy}" in
        check)
            check
            ;;
        install)
            install
            ;;
        deploy)
            deploy
            ;;
        start)
            start_master
            ;;
        worker)
            start_worker "$2"
            ;;
        stop)
            stop
            ;;
        status)
            status
            ;;
        *)
            echo "AgentFlow macOS 快速部署脚本"
            echo ""
            echo "用法: $0 <命令>"
            echo ""
            echo "命令:"
            echo "  check      - 检查环境"
            echo "  install    - 安装所有依赖"
            echo "  deploy     - 完整部署（默认）"
            echo "  start      - 启动 Master"
            echo "  worker [组] - 启动 Worker（可选组名）"
            echo "  stop       - 停止所有服务"
            echo "  status     - 查看系统状态"
            echo ""
            echo "示例:"
            echo "  $0 deploy          # 完整部署"
            echo "  $0 status          # 查看状态"
            echo "  $0 worker backend   # 启动 backend 组 Worker"
            exit 0
            ;;
    esac
}

main "$@"
