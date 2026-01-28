# AgentFlow Master Makefile

.PHONY: all build run test clean check fmt clippy doc help

# 默认目标
all: build

# 构建项目
build:
	@echo "🔨 构建 AgentFlow Master..."
	cargo build --release

# 运行服务器
run:
	@echo "🚀 启动 AgentFlow Master..."
	cargo run --bin agentflow-master

# 开发模式运行（自动重新加载）
dev:
	@echo "🔥 开发模式启动..."
	cargo watch -x run

# 运行测试
test:
	@echo "🧪 运行测试..."
	cargo test -p agentflow-master

# 测试（带输出）
test-verbose:
	@echo "🧪 运行测试（详细输出）..."
	cargo test -p agentflow-master -- --nocapture

# 集成测试（需要服务器运行）
test-integration:
	@echo "🔗 运行集成测试..."
	cargo test -p agentflow-master --test integration_test

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	cargo clean
	rm -rf target/

# 检查代码
check:
	@echo "🔍 检查代码..."
	cargo check -p agentflow-master

# 格式化代码
fmt:
	@echo "✨ 格式化代码..."
	cargo fmt -p agentflow-master

# 代码检查
clippy:
	@echo "📎 代码检查..."
	cargo clippy -p agentflow-master -- -D warnings

# 生成文档
doc:
	@echo "📚 生成文档..."
	cargo doc -p agentflow-master --no-deps --open

# 运行示例
example-client:
	@echo "🎯 运行客户端示例..."
	cargo run --example client

# 安装依赖
install:
	@echo "📦 安装依赖..."
	cargo fetch

# 数据库迁移
migrate:
	@echo "🗄️  创建数据库..."
	cargo run --bin agentflow-master -- --migrate

# 显示帮助信息
help:
	@echo "AgentFlow Master Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make build         - 构建项目"
	@echo "  make run           - 运行服务器"
	@echo "  make dev           - 开发模式（自动重新加载）"
	@echo "  make test          - 运行测试"
	@echo "  make test-verbose  - 运行测试（详细输出）"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make clean         - 清理构建文件"
	@echo "  make check         - 检查代码"
	@echo "  make fmt           - 格式化代码"
	@echo "  make clippy        - 代码检查"
	@echo "  make doc           - 生成文档"
	@echo "  make example-client - 运行客户端示例"
	@echo "  make install       - 安装依赖"
	@echo "  make migrate       - 数据库迁移"
	@echo "  make help          - 显示帮助信息"
