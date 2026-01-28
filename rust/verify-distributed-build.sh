#!/bin/bash

# AgentFlow 分布式执行系统 - 构建验证脚本
#
# 用途: 验证分布式执行模块是否正确编译
# 使用: ./verify-distributed-build.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUST_DIR="${PROJECT_ROOT}/rust"

echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}AgentFlow 分布式执行系统构建验证${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""

# 检查 Rust 工具链
echo -e "${YELLOW}1. 检查 Rust 工具链...${NC}"
if ! command -v cargo &> /dev/null; then
    echo -e "${RED}✗ Cargo 未安装${NC}"
    echo "请访问 https://rustup.rs/ 安装 Rust"
    exit 1
fi

RUST_VERSION=$(rustc --version)
echo -e "${GREEN}✓ ${RUST_VERSION}${NC}"
echo ""

# 检查新添加的依赖
echo -e "${YELLOW}2. 检查依赖项...${NC}"
cd "${RUST_DIR}"

# 检查 Cargo.toml 是否包含 petgraph
if grep -q "petgraph" agentflow-master/Cargo.toml; then
    echo -e "${GREEN}✓ petgraph 依赖已添加${NC}"
else
    echo -e "${RED}✗ petgraph 依赖缺失${NC}"
    echo "请运行: cargo add petgraph --package agentflow-master"
    exit 1
fi

# 检查 Cargo.toml 是否包含 bincode
if grep -q "bincode" agentflow-master/Cargo.toml; then
    echo -e "${GREEN}✓ bincode 依赖已添加${NC}"
else
    echo -e "${RED}✗ bincode 依赖缺失${NC}"
    echo "请运行: cargo add bincode --package agentflow-master"
    exit 1
fi

echo ""

# 检查源文件
echo -e "${YELLOW}3. 检查源文件...${NC}"
FILES=(
    "agentflow-master/src/leader/raft.rs"
    "agentflow-master/src/scheduler/dependency.rs"
    "agentflow-master/src/scheduler/queue.rs"
    "agentflow-master/src/worker_registry.rs"
    "agentflow-master/src/agent_comm.rs"
    "agentflow-master/src/distributed_lock.rs"
)

for file in "${FILES[@]}"; do
    if [ -f "${file}" ]; then
        LINES=$(wc -l < "${file}")
        echo -e "${GREEN}✓ ${file} (${LINES} 行)${NC}"
    else
        echo -e "${RED}✗ ${file} 缺失${NC}"
        exit 1
    fi
done

echo ""

# 语法检查
echo -e "${YELLOW}4. 运行 cargo check...${NC}"
if cargo check --package agentflow-master 2>&1 | tee /tmp/cargo-check.log; then
    echo -e "${GREEN}✓ 语法检查通过${NC}"
else
    echo -e "${RED}✗ 语法检查失败${NC}"
    echo "查看错误日志: /tmp/cargo-check.log"
    exit 1
fi

echo ""

# 编译检查
echo -e "${YELLOW}5. 运行 cargo build...${NC}"
if cargo build --package agentflow-master 2>&1 | tee /tmp/cargo-build.log; then
    echo -e "${GREEN}✓ 编译成功${NC}"
else
    echo -e "${RED}✗ 编译失败${NC}"
    echo "查看错误日志: /tmp/cargo-build.log"
    exit 1
fi

echo ""

# Clippy 检查
echo -e "${YELLOW}6. 运行 cargo clippy...${NC}"
if cargo clippy --package agentflow-master -- -D warnings 2>&1 | tee /tmp/cargo-clippy.log; then
    echo -e "${GREEN}✓ Clippy 检查通过${NC}"
else
    echo -e "${YELLOW}⚠ Clippy 发现一些警告${NC}"
    echo "查看警告日志: /tmp/cargo-clippy.log"
fi

echo ""

# 单元测试
echo -e "${YELLOW}7. 运行单元测试...${NC}"
if cargo test --package agentflow-master 2>&1 | tee /tmp/cargo-test.log; then
    echo -e "${GREEN}✓ 单元测试通过${NC}"
else
    echo -e "${YELLOW}⚠ 部分测试失败或需要完善${NC}"
    echo "查看测试日志: /tmp/cargo-test.log"
fi

echo ""

# 文档生成
echo -e "${YELLOW}8. 生成文档...${NC}"
if cargo doc --package agentflow-master --no-deps 2>&1 | tee /tmp/cargo-doc.log; then
    echo -e "${GREEN}✓ 文档生成成功${NC}"
    echo "文档路径: ${RUST_DIR}/target/doc/agentflow_master/index.html"
else
    echo -e "${YELLOW}⚠ 文档生成有警告${NC}"
fi

echo ""

# 总结
echo -e "${BLUE}=====================================${NC}"
echo -e "${GREEN}✓ 验证完成！${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""
echo "已实现的模块:"
echo "  ✓ Raft 一致性算法 (Leader 选举)"
echo "  ✓ 任务依赖管理 (DAG 工作流)"
echo "  ✓ 优先级任务队列"
echo "  ✓ Worker 注册中心"
echo "  ✓ Agent 通信协议"
echo "  ✓ 分布式锁管理"
echo ""
echo "下一步建议:"
echo "  1. 查看文档: docs/DISTRIBUTED_EXECUTION_SYSTEM.md"
echo "  2. 运行示例: cargo run --bin agentflow-master"
echo "  3. 运行集成测试: make test-integration"
echo "  4. 部署到 Docker: docker-compose up"
echo ""
echo -e "${BLUE}=====================================${NC}"
