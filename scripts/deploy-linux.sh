#!/bin/bash
# AgentFlow Linux 快速部署脚本
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

# 检测 Linux 发行版
detect_linux_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
        VERSION=$VERSION_ID
        print_info "Linux 发行版: $DISTRO $VERSION"
    else
        DISTRO="unknown"
    fi
}

# 检查 Linux 环境
check_linux_environment() {
    print_header "检查 Linux 环境"

    detect_linux_distro

    # 检查架构
    ARCH=$(uname -m)
    print_info "系统架构: $ARCH"

    # 检查内存
    if [ -f /proc/meminfo ]; then
        MEM_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        MEM_GB=$((MEM_KB / 1024 / 1024))
        print_info "系统内存: ${MEM_GB}GB"
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

        # 尝试使用 npm
        if command -v npm &> /dev/null; then
            print_info "使用 npm 安装 Claude CLI..."
            sudo npm install -g @anthropic-ai/claude-cli
            print_success "Claude CLI 安装完成"
        else
            print_warning "npm 未安装"
            print_info "请手动安装 Claude CLI:"
            print_info "  方法 1: npm install -g @anthropic-ai/claude-cli"
            print_info "  方法 2: 下载二进制: https://github.com/anthropics/claude-cli/releases"
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

        # 根据发行版选择安装方法
        case "$DISTRO" in
            ubuntu|debian)
                print_info "安装 Python 3..."
                sudo apt update
                sudo apt install -y python3 python3-pip
                ;;
            fedora|rhel|centos)
                print_info "安装 Python 3..."
                sudo dnf install -y python3 python3-pip
                ;;
            arch|manjaro)
                print_info "安装 Python 3..."
                sudo pacman -S python3 python-pip
                ;;
            *)
                print_warning "无法自动安装 Python 3，请手动安装"
                ;;
        esac
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

        # 根据发行版选择安装方法
        case "$DISTRO" in
            ubuntu|debian)
                print_info "安装 Go..."
                sudo apt update
                sudo apt install -y golang-go
                ;;
            fedora|rhel|centos)
                print_info "安装 Go..."
                sudo dnf install -y golang
                ;;
            arch|manjaro)
                print_info "安装 Go..."
                sudo pacman -S go
                ;;
            *)
                print_warning "无法自动安装 Go，请手动安装"
                print_info "推荐方式: 下载官方包: https://go.dev/dl/"
                ;;
        esac
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

# 创建 systemd 服务
create_systemd_service() {
    print_header "创建 systemd 服务"

    # Master 服务
    sudo tee /etc/systemd/system/agentflow-master.service > /dev/null << EOF
[Unit]
Description=AgentFlow Master Service
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$PWD/golang
ExecStart=$PWD/golang/bin/master -config $PWD/golang/config.example.yaml
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    # Worker 服务
    sudo tee /etc/systemd/system/agentflow-worker.service > /dev/null << EOF
[Unit]
Description=AgentFlow Worker Service
After=network.target agentflow-master.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$PWD/golang
ExecStart=$PWD/golang/bin/worker -config $PWD/golang/config.example.yaml
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    print_success "systemd 服务已创建"
    print_info "管理命令:"
    print_info "  启动: sudo systemctl start agentflow-master"
    print_info "  停止: sudo systemctl stop agentflow-master"
    print_info "  自启: sudo systemctl enable agentflow-master"
}

# 启动 Master 服务
start_master() {
    print_header "启动 Master 服务"

    # 检查是否已运行
    if systemctl is-active --quiet agentflow-master 2>/dev/null; then
        print_warning "Master 已在运行"
        return
    fi

    if [ -f "golang/bin/master" ]; then
        print_info "启动 Master (Go 版本)..."
        sudo systemctl start agentflow-master
        print_success "Master 已启动"
        print_info "Master 地址: http://localhost:8848"
    else
        print_error "未找到可执行的 master"
        print_info "请先运行: $0 install && $0 build"
    fi
}

# 启动 Worker
start_worker() {
    print_header "启动 Worker"

    WORKER_GROUP="${1:-default}"

    if systemctl is-active --quiet agentflow-worker 2>/dev/null; then
        print_warning "Worker 已在运行"
        return
    fi

    if [ -f "golang/bin/worker" ]; then
        print_info "启动 Worker (组: $WORKER_GROUP)..."
        sudo systemctl start agentflow-worker
        print_success "Worker 已启动 (组: $WORKER_GROUP)"
    fi
}

# 显示使用指南
show_usage() {
    cat << EOF

${GREEN}AgentFlow Linux 快速部署完成！${NC}

📝 常用命令:
  ${YELLOW}# 查看服务状态${NC}
  sudo systemctl status agentflow-master
  sudo systemctl status agentflow-worker

  ${YELLOW}# 管理服务${NC}
  sudo systemctl start agentflow-master   # 启动 Master
  sudo systemctl stop agentflow-master    # 停止 Master
  sudo systemctl restart agentflow-master  # 重启 Master
  sudo systemctl enable agentflow-master   # 开机自启

  ${YELLOW}# 查看日志${NC}
  sudo journalctl -u agentflow-master -f
  sudo journalctl -u agentflow-worker -f

  ${YELLOW}# 快速任务${NC}
  ./scripts/quick-task.sh "测试任务" "shell:echo Hello World"

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
   4. 使用 systemd 管理服务（推荐生产环境）

EOF
}

# 检查依赖
check() {
    check_linux_environment
    echo ""
}

# 安装所有依赖
install() {
    check_linux_environment
    install_claude_cli
    install_python_dependencies
    install_go_dependencies
    echo ""
    print_success "所有依赖已安装！"
}

# 构建二进制
build() {
    build_go_binaries
    setup_file_boundaries
    echo ""
    print_success "构建完成！"
}

# 完整部署
deploy() {
    print_header "AgentFlow Linux 快速部署"

    check
    install
    build
    create_systemd_service
    start_master

    show_usage
}

# 停止服务
stop() {
    print_header "停止服务"

    sudo systemctl stop agentflow-worker 2>/dev/null || true
    sudo systemctl stop agentflow-master 2>/dev/null || true

    print_success "所有服务已停止"
}

# 状态检查
status() {
    print_header "系统状态"

    # 检查 Master
    if systemctl is-active --quiet agentflow-master 2>/dev/null; then
        print_success "Master 运行中"
    else
        print_warning "Master 未运行"
    fi

    # 检查 Worker
    if systemctl is-active --quiet agentflow-worker 2>/dev/null; then
        print_success "Worker 运行中"
    else
        print_warning "Worker 未运行"
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
        build)
            build
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
            echo "AgentFlow Linux 快速部署脚本"
            echo ""
            echo "用法: $0 <命令>"
            echo ""
            echo "命令:"
            echo "  check      - 检查环境"
            echo "  install    - 安装所有依赖"
            echo "  build      - 编译二进制文件"
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
