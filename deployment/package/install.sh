#!/bin/bash

###############################################################################
# AgentFlow 快速安装脚本（纯 Rust 版本）
###############################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_logo() {
    echo -e "${CYAN}"
    cat << "LOGO"
  ██████╗ ██████╗ ███████╗ █████╗ ███╗   ███╗███████╗
 ██╔════╝██╔═══██╗██╔════╝██╔══██╗████╗ ████║██╔════╝
 ██║     ██║   ██║███████╗███████║██╔████╔██║███████╗
 ██║     ██║   ██║╚════██║██╔══██║██║╚██╔╝██║╚════██║
 ╚██████╗╚██████╔╝███████║██║  ██║██║ ╚═╝ ██║███████║
  ╚═════╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝
              纯 Rust 版本安装脚本 v1.0
LOGO
    echo -e "${NC}"
}

info() { echo -e "${BLUE}ℹ${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

check_environment() {
    echo -e "\n${CYAN}检查系统环境...${NC}\n"

    if command_exists rustc; then
        RUST_VERSION=$(rustc --version)
        success "Rust: $RUST_VERSION"
    else
        error "Rust 未安装"
        echo "  运行: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
        exit 1
    fi

    if command_exists git; then
        GIT_VERSION=$(git --version)
        success "Git: $GIT_VERSION"
    else
        error "Git 未安装"
        echo "  请访问 https://git-scm.com/ 安装"
        exit 1
    fi

    echo ""
}

compile_master() {
    echo -e "${CYAN}编译 Master 服务器...${NC}\n"

    cd rust
    cargo build --release --bin agentflow-master

    if [ -f "target/release/agentflow-master" ]; then
        success "Master 编译成功"
        SIZE=$(du -h target/release/agentflow-master | cut -f1)
        echo "  二进制文件: target/release/agentflow-master ($SIZE)"
    else
        error "Master 编译失败"
        exit 1
    fi

    cd ..
    echo ""
}

build_dashboard() {
    echo -e "${CYAN}构建 Dashboard（可选）...${NC}\n"

    if ! command_exists node; then
        warn "Node.js 未安装，跳过 Dashboard 构建"
        echo "  安装: https://nodejs.org/"
        return
    fi

    cd dashboard
    info "安装依赖..."
    npm install

    info "构建应用..."
    npm run build

    if [ -d "dist" ] && [ "$(ls -A dist)" ]; then
        success "Dashboard 构建成功"
        echo "  静态文件: dashboard/dist/"
    else
        error "Dashboard 构建失败"
    fi

    cd ..
    echo ""
}

build_helper() {
    echo -e "${CYAN}编译 Helper 工具...${NC}\n"

    cd rust/agentflow-helper
    cargo build --release

    if [ -f "target/release/agentflow-helper" ]; then
        success "Helper 编译成功"
        echo "  二进制文件: rust/agentflow-helper/target/release/agentflow-helper"
    else
        error "Helper 编译失败"
    fi

    cd ../../..
    echo ""
}

create_config() {
    echo -e "${CYAN}创建配置文件...${NC}\n"

    cat > .env << 'ENVCONFIG'
# AgentFlow 配置文件
# 由快速安装脚本生成（纯 Rust 版本）

# ==================== 认证配置 ====================
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=admin
AUTH_SESSION_TTL=86400

# ==================== 服务器配置 ====================
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_LOG_LEVEL=info

# ==================== WebSocket 配置 ====================
AGENTFLOW_WS_ENABLED=true
AGENTFLOW_WS_PORT=8849
ENVCONFIG

    success "配置文件已创建: .env"
    echo ""
}

create_startup_script() {
    echo -e "${CYAN}创建启动脚本...${NC}\n"

    cat > start.sh << 'STARTSCRIPT'
#!/bin/bash

# 加载环境变量
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# 启动 Master
echo "启动 AgentFlow Master..."
./rust/target/release/agentflow-master
STARTSCRIPT

    chmod +x start.sh

    success "启动脚本已创建: start.sh"
    echo ""
}

print_completion() {
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}        ✓ AgentFlow 安装完成！${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}\n"

    echo -e "${CYAN}快速开始:${NC}\n"

    echo -e "  ${YELLOW}方式 1: 使用启动脚本${NC}"
    echo -e "  ${WHITE}./start.sh${NC}\n"

    echo -e "  ${YELLOW}方式 2: 直接运行${NC}"
    echo -e "  ${WHITE}./rust/target/release/agentflow-master${NC}\n"

    echo -e "  ${YELLOW}方式 3: 使用 Helper 工具${NC}"
    echo -e "  ${WHITE}./rust/agentflow-helper/target/release/agentflow-helper check${NC}\n"

    echo -e "${CYAN}访问地址:${NC}\n"
    echo -e "  ${WHITE}Dashboard: http://localhost:6767${NC}"
    echo -e "  ${WHITE}API:       http://localhost:6767/api/v1${NC}\n"

    echo -e "${CYAN}默认账号:${NC}\n"
    echo -e "  ${WHITE}用户名: admin${NC}"
    echo -e "  ${WHITE}密码:   admin${NC}\n"
}

main() {
    print_logo
    check_environment
    compile_master
    build_helper
    build_dashboard
    create_config
    create_startup_script
    print_completion
}

main
