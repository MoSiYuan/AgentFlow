#!/bin/bash

# AgentFlow 快速启动脚本
# 用于快速启动 AgentFlow 开发环境

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
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

print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# 检查依赖
check_dependencies() {
    print_header "检查依赖"

    # 检查 Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_success "Go 已安装: $GO_VERSION"
    else
        print_error "Go 未安装，请先安装 Go 1.21+"
        exit 1
    fi

    # 检查 SQLite
    if command -v sqlite3 &> /dev/null; then
        SQLITE_VERSION=$(sqlite3 --version | awk '{print $1}')
        print_success "SQLite 已安装: $SQLITE_VERSION"
    else
        print_warning "SQLite 未安装，某些功能可能受限"
    fi
}

# 初始化数据库
init_database() {
    print_header "初始化数据库"

    DB_PATH="${DB_PATH:-agentflow.db}"

    if [ -f "$DB_PATH" ]; then
        print_warning "数据库已存在: $DB_PATH"
        read -p "是否删除并重新初始化？(y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -f "$DB_PATH"
            print_info "已删除旧数据库"
        else
            print_info "使用现有数据库"
            return
        fi
    fi

    # 如果有编译好的二进制文件，使用它
    if [ -f "./bin/agentflow" ]; then
        print_info "使用编译好的 agentflow 初始化数据库"
        ./bin/agentflow init "$DB_PATH"
    else
        print_warning "未找到编译好的二进制文件，使用纯标准库版本"
        print_info "运行演示测试..."
        cd tests
        go run ctest_pure.go
        cd ..
    fi

    print_success "数据库初始化完成: $DB_PATH"
}

# 启动 Master
start_master() {
    print_header "启动 Master 服务"

    DB_PATH="${DB_PATH:-agentflow.db}"
    HOST="${HOST:-localhost}"
    PORT="${PORT:-8848}"

    if [ -f "./bin/agentflow" ]; then
        print_info "启动 Master (数据库: $DB_PATH, 端口: $PORT)"
        ./bin/agentflow master --db "$DB_PATH" --port "$PORT" &
        MASTER_PID=$!
        echo $MASTER_PID > /tmp/agentflow-master.pid
        print_success "Master 已启动 (PID: $MASTER_PID)"
    else
        print_warning "未找到编译好的二进制文件"
        print_info "使用纯标准库版本演示功能..."
        print_info "Master 功能在完整版本中可用"
    fi
}

# 显示帮助
show_help() {
    cat << EOF
AgentFlow 快速启动脚本

用法: $0 [命令]

命令:
  check       检查依赖
  init        初始化数据库
  start       启动 Master 服务
  demo        运行演示（纯标准库版本）
  test        创建测试任务
  status      查看系统状态
  stop        停止所有服务
  clean       清理临时文件
  help        显示此帮助信息

环境变量:
  DB_PATH     数据库路径 (默认: agentflow.db)
  HOST        Master 主机 (默认: localhost)
  PORT        Master 端口 (默认: 8848)

示例:
  $0              # 完整启动流程
  $0 demo         # 运行演示
  $0 init         # 仅初始化数据库
  DB_PATH=test.db $0 start   # 使用自定义数据库

更多文档:
  - 安装指南: INSTALL_GUIDE.md
  - Skill 文档: skills/agentflow.md
  - 自迭代开发: SELF_ITERATION.md

EOF
}

# 运行演示
run_demo() {
    print_header "运行 AgentFlow 演示"

    print_info "使用纯标准库版本（无需外部依赖）"
    cd tests
    go run ctest_pure.go
    cd ..

    print_success "演示完成！"
    print_info "生成的文件在: tests/ctest_stories/"
}

# 创建测试任务
create_test_tasks() {
    print_header "创建测试任务"

    if [ ! -f "./bin/agentflow" ]; then
        print_error "未找到 agentflow 二进制文件，请先编译项目"
        print_info "运行: make build"
        exit 1
    fi

    DB_PATH="${DB_PATH:-agentflow.db}"

    print_info "创建测试任务..."

    # 测试任务 1: 格式化
    ./bin/agentflow add "格式化代码" \
        --desc "shell:gofmt -w ." \
        --db "$DB_PATH"
    print_success "任务 1: 格式化代码"

    # 测试任务 2: 测试
    ./bin/agentflow add "运行测试" \
        --desc "shell:go test ./..." \
        --db "$DB_PATH"
    print_success "任务 2: 运行测试"

    # 测试任务 3: 构建
    ./bin/agentflow add "构建应用" \
        --desc "shell:go build -v ./..." \
        --db "$DB_PATH"
    print_success "任务 3: 构建应用"

    print_info "查看任务: ./bin/agentflow list --db $DB_PATH"
}

# 查看状态
show_status() {
    print_header "系统状态"

    # 检查 Master
    if [ -f "/tmp/agentflow-master.pid" ]; then
        MASTER_PID=$(cat /tmp/agentflow-master.pid)
        if ps -p $MASTER_PID > /dev/null 2>&1; then
            print_success "Master 运行中 (PID: $MASTER_PID)"
        else
            print_warning "Master 未运行"
            rm -f /tmp/agentflow-master.pid
        fi
    else
        print_warning "Master 未运行"
    fi

    # 检查数据库
    DB_PATH="${DB_PATH:-agentflow.db}"
    if [ -f "$DB_PATH" ]; then
        print_success "数据库存在: $DB_PATH"

        if command -v sqlite3 &> /dev/null; then
            print_info "数据库统计:"
            sqlite3 "$DB_PATH" <<EOF
.mode column
.headers on
SELECT status, COUNT(*) as count FROM tasks GROUP BY status;
EOF
        fi
    else
        print_warning "数据库不存在: $DB_PATH"
    fi

    # 检查二进制文件
    if [ -f "./bin/agentflow" ]; then
        print_success "二进制文件存在: ./bin/agentflow"
    else
        print_warning "二进制文件不存在，请先运行: make build"
    fi
}

# 停止服务
stop_services() {
    print_header "停止服务"

    if [ -f "/tmp/agentflow-master.pid" ]; then
        MASTER_PID=$(cat /tmp/agentflow-master.pid)
        if ps -p $MASTER_PID > /dev/null 2>&1; then
            print_info "停止 Master (PID: $MASTER_PID)"
            kill $MASTER_PID
            rm -f /tmp/agentflow-master.pid
            print_success "Master 已停止"
        fi
    fi

    # 清理 Worker 进程
    pkill -f "agentflow worker" 2>/dev/null && print_success "Workers 已停止" || true
}

# 清理临时文件
clean_files() {
    print_header "清理临时文件"

    read -p "是否清理所有临时文件？(y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "取消清理"
        return
    fi

    # 停止服务
    stop_services

    # 清理数据库
    rm -f agentflow.db ctest_pure.db
    print_info "已删除数据库文件"

    # 清理测试输出
    rm -rf tests/ctest_stories
    print_info "已删除测试输出"

    # 清理临时文件
    rm -f /tmp/agentflow-*.log
    print_info "已删除日志文件"

    print_success "清理完成"
}

# 主函数
main() {
    print_header "AgentFlow - AI Agent 任务协作系统"

    case "${1:-start}" in
        check)
            check_dependencies
            ;;
        init)
            init_database
            ;;
        start)
            check_dependencies
            init_database
            start_master
            print_success "AgentFlow 已启动！"
            print_info "Master 运行在: http://localhost:${PORT:-8848}"
            print_info "查看任务: ./bin/agentflow list --db ${DB_PATH:-agentflow.db}"
            print_info "停止服务: $0 stop"
            ;;
        demo)
            run_demo
            ;;
        test)
            create_test_tasks
            ;;
        status)
            show_status
            ;;
        stop)
            stop_services
            ;;
        clean)
            clean_files
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
