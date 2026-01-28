# AgentFlow v3 (Pure Rust) - 快速开始指南

## 🎯 项目概述

AgentFlow v3 是完全用 Rust 重写的版本，采用**单进程架构**（Master = Worker），无需 Node.js 依赖。

**核心特性：**
- ✅ 单二进制、单进程
- ✅ 直接调用 Claude CLI（继承环境变量）
- ✅ 基于 tokio 的高性能异步运行时
- ✅ SQLite 向量记忆系统
- ✅ 完整的沙箱安全控制
- ✅ HTTP API + WebSocket 支持

---

## 📦 安装 Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
```

---

## 🚀 快速开始

### 方式 1: 使用启动脚本（推荐）

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow
./start-rust.sh
```

### 方式 2: 手动编译和运行

```bash
# 1. 进入 Rust 项目目录
cd /Users/jiangxiaolong/work/project/AgentFlow/rust

# 2. 创建 .env 文件
cat > .env << 'EOF'
DATABASE_URL=sqlite://agentflow.db
RUST_LOG=info
SQLX_OFFLINE=true
EOF

# 3. 编译项目
export SQLX_OFFLINE=true
cargo build --release

# 4. 运行服务器
./target/release/agentflow-master
```

### 方式 3: 开发模式

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow/rust
export SQLX_OFFLINE=true
cargo run --bin agentflow-master
```

---

## 📝 使用示例

### 1. 创建任务

```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试任务",
    "description": "请帮我写一个 Hello World 程序",
    "priority": "high"
  }'
```

### 2. 执行任务（SSE 流式）

```bash
curl -X POST http://localhost:6767/api/v1/tasks/1/execute \
  -H "Accept: text/event-stream"
```

### 3. 查询任务状态

```bash
curl http://localhost:6767/api/v1/tasks/1
```

### 4. WebSocket 连接

```javascript
const ws = new WebSocket('ws://localhost:6767/ws/task/1');
ws.onmessage = (event) => {
  console.log('Task update:', JSON.parse(event.data));
};
```

---

## 🔧 配置选项

环境变量配置（`.env` 文件）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `AGENTFLOW_PORT` | 6767 | 服务器端口 |
| `AGENTFLOW_HOST` | 0.0.0.0 | 绑定地址 |
| `DATABASE_URL` | sqlite://agentflow.db | 数据库连接 |
| `RUST_LOG` | info | 日志级别 |
| `SQLX_OFFLINE` | true | SQLx 离线模式 |

---

## 📂 项目结构

```
rust/
├── Cargo.toml                    # Workspace 配置
├── .env                          # 环境变量
├── agentflow-core/               # 核心库
│   ├── src/
│   │   ├── types.rs              # 共享类型
│   │   ├── executor/             # 执行引擎 (Team A)
│   │   │   ├── mod.rs            # TaskExecutor
│   │   │   ├── killer.rs         # ProcessKiller
│   │   │   └── prompt_builder.rs # PromptBuilder
│   │   ├── memory/               # 记忆系统 (Team B)
│   │   │   └── mod.rs            # MemoryCore
│   │   └── sandbox/              # 沙箱安全 (Team B)
│   │       └── mod.rs            # SandboxConfig
└── agentflow-master/             # Master 服务器 (Team C)
    ├── src/
    │   ├── main.rs               # 主程序
    │   ├── config.rs             # 配置管理
    │   ├── executor.rs           # 执行器集成
    │   ├── memory_core.rs        # 记忆核心集成
    │   └── routes/               # API 路由
    │       ├── tasks.rs          # 任务 API
    │       ├── memory.rs         # 记忆 API
    │       ├── websocket.rs      # WebSocket
    │       └── health.rs         # 健康检查
    └── target/
        └── release/
            └── agentflow-master   # 可执行文件
```

---

## 🎨 核心模块

### Team A: 执行引擎
- **TaskExecutor** - 执行 Claude CLI 命令
- **ProcessKiller** - 超时熔断和进程管理
- **PromptBuilder** - 智能 Prompt 构建

### Team B: 记忆与安全
- **MemoryCore** - SQLite 向量记忆系统
- **SandboxConfig** - 路径白名单和沙箱控制
- **PromptBuilder** - Token 管理和截断

### Team C: API 和路由
- **HTTP API** - RESTful 接口（14个端点）
- **WebSocket** - 实时任务更新
- **SSE** - 流式输出支持

---

## 🔒 安全特性

- ✅ **路径白名单** - 只允许访问指定目录
- ✅ **路径穿透检测** - 防止 `../` 攻击
- ✅ **符号链接防护** - 递归检查，深度限制
- ✅ **进程超时熔断** - SIGTERM → 等待 → SIGKILL
- ✅ **进程组级联清理** - 自动清理所有子进程

---

## 📊 性能指标

- **内存占用**: < 100MB (空闲)
- **启动时间**: < 1 秒
- **并发任务**: 支持多任务并发执行
- **响应速度**: API 响应 < 10ms

---

## 🐛 故障排查

### 编译错误

**问题**: `error: set DATABASE_URL to use query macros online`

**解决**: 设置 `SQLX_OFFLINE=true`
```bash
export SQLX_OFFLINE=true
cargo build
```

### 运行错误

**问题**: `Database locked`

**解决**: SQLite 文件权限问题
```bash
chmod 644 agentflow.db
```

### Claude CLI 未找到

**问题**: `No such file or directory (os error 2)`

**解决**: 确保 claude 命令在 PATH 中
```bash
which claude
export PATH="$HOME/.npm-global/bin:$PATH"
```

---

## 📚 相关文档

- **[RUST_V3_IMPLEMENTATION.md](../RUST_V3_IMPLEMENTATION.md)** - 完整实现报告
- **[TEAM_A_IMPLEMENTATION_REPORT.md](../docs/TEAM_A_IMPLEMENTATION_REPORT.md)** - Team A 报告
- **[TEAM_B_IMPLEMENTATION_REPORT.md](../docs/TEAM_B_IMPLEMENTATION_REPORT.md)** - Team B 报告
- **[API.md](rust/agentflow-master/API.md)** - API 文档

---

## 🎯 下一步

1. **编译并运行**: `./start-rust.sh`
2. **测试 API**: 创建测试任务
3. **查看日志**: 观察任务执行过程
4. **扩展功能**: 根据需求定制

---

**享受高性能的 AgentFlow v3 (Pure Rust)! 🚀**
