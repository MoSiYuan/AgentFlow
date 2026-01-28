# AgentFlow - AI Agent Task Collaboration System

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Rust](https://img.shields.io/badge/Rust-1.93+-orange.svg)](https://www.rust-lang.org/)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)](https://github.com/MoSiYuan/AgentFlow)

**Single-process, high-performance AI agent task orchestration system written in pure Rust.**

## 🎯 Version 3.0 (Pure Rust)

AgentFlow v3 is a complete rewrite in Rust, featuring a revolutionary **single-process architecture** where the Master server also acts as the Worker, eliminating the need for separate worker processes.

### Key Features

- ✅ **Single Binary** - One executable, no dependencies
- ✅ **Single Process** - Master = Worker, no inter-process communication
- ✅ **High Performance** - Built on Tokio async runtime, < 100MB memory
- ✅ **Direct Execution** - Executes Claude CLI directly via tokio::process
- ✅ **Vector Memory** - SQLite-based vector indexing for semantic retrieval
- ✅ **Sandbox Security** - Complete path validation and process isolation
- ✅ **REST API** - 14 HTTP endpoints
- ✅ **Real-time** - WebSocket and SSE streaming support

## 🚀 Quick Start

### Option 1: One-Click Installation (Recommended)

#### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/MoSiYuan/AgentFlow/main/scripts/install.sh | bash
```

Or download and run manually:

```bash
wget https://raw.githubusercontent.com/MoSiYuan/AgentFlow/main/scripts/install.sh
chmod +x install.sh
./install.sh
```

#### Windows

```powershell
irm https://raw.githubusercontent.com/MoSiYuan/AgentFlow/main/scripts/install.ps1 | iex
```

### Option 2: Build from Source

#### 1. Install Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
```

#### 2. Build AgentFlow

```bash
cd rust
export SQLX_OFFLINE=true
cargo build --release
```

#### 3. Run AgentFlow

```bash
./target/release/agentflow-master
```

Server will start on `http://localhost:6767`

## 📝 Usage Examples

### Operating Modes

AgentFlow supports three operating modes:

#### 1. Local Mode (Default)

Execute tasks locally using Claude CLI:

```bash
agentflow server local
# or simply
agentflow server
```

#### 2. Cloud Mode (with Webhooks)

Integrate with AI platforms like Zhipu AI:

```bash
agentflow server cloud
```

#### 3. Planner-Only Mode

Plan and validate tasks without execution:

```bash
agentflow server planner-only
```

### Create a Task

```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Task",
    "description": "echo Hello from AgentFlow v3!",
    "priority": "high"
  }'
```

### Execute a Task (with SSE streaming)

```bash
curl -X POST http://localhost:6767/api/v1/tasks/1/execute \
  -H "Accept: text/event-stream"
```

### Query Task Status

```bash
curl http://localhost:6767/api/v1/tasks/1
```

## 🌐 Cloud Mode & Zhipu AI Integration

AgentFlow can integrate with Zhipu AI (智谱清言) to create an AI-powered task orchestration system.

### Quick Setup

1. **Configure Zhipu AI Integration**

Edit `~/.agentflow/config.toml`:

```toml
[server]
port = 6767

[webhook]
enabled = true
secret = "your-webhook-secret-key"

[zhipu]
enabled = true
api_key = "your-zhipu-api-key"
model = "glm-4"
```

2. **Start AgentFlow in Cloud Mode**

```bash
agentflow server cloud
```

3. **Setup Public URL** (for testing)

```bash
# Using ngrok
ngrok http 6767
# Output: https://abc123.ngrok.io
```

4. **Configure Zhipu AI Webhook**

In Zhipu AI Console, set webhook URL to:
```
https://abc123.ngrok.io/api/v1/webhook
```

5. **Test Integration**

Send a message through Zhipu AI:
```
"帮我创建一个任务，分析这个项目的代码结构"
```

AgentFlow will receive the webhook, create a task, execute it, and send the result back to Zhipu AI.

### Example Webhook Request

```json
{
  "event": "message.received",
  "timestamp": "2026-01-28T10:30:00Z",
  "data": {
    "message_id": "msg_123",
    "user_id": "user_abc",
    "content": "帮我分析这个Go项目的代码结构",
    "metadata": {
      "source": "zhipu",
      "model": "glm-4"
    }
  }
}
```

### Detailed Documentation

- **[ZHIPU_INTEGRATION.md](docs/ZHIPU_INTEGRATION.md)**: Complete Zhipu AI integration guide
- **[CONFIGURATION.md](docs/CONFIGURATION.md)**: Full configuration reference

---

## ⚡ Distributed Execution Mode (NEW!)

AgentFlow now supports **distributed parallel execution** with Master cluster, workflow orchestration, and intelligent scheduling! (v0.4.0)

### Key Features

- ✅ **Master Cluster** - Raft-based leader election and fault tolerance
- ✅ **DAG Workflows** - Task dependency management and parallel execution
- ✅ **Priority Queue** - Intelligent task scheduling (Urgent > High > Medium > Low)
- ✅ **Worker Registry** - Health checking and load balancing
- ✅ **Agent Communication** - Point-to-point and broadcast messaging
- ✅ **Distributed Locks** - Cross-node coordination

### Quick Start

#### 1. Start Master Cluster (3 nodes)

```bash
# Terminal 1 - Master 1
cargo run --bin agentflow-master -- \
  --node-id master-1 --port 6767 \
  --peers master-1:6767,master-2:6768,master-3:6769

# Terminal 2 - Master 2
cargo run --bin agentflow-master -- \
  --node-id master-2 --port 6768 \
  --peers master-1:6767,master-2:6768,master-3:6769

# Terminal 3 - Master 3
cargo run --bin agentflow-master -- \
  --node-id master-3 --port 6769 \
  --peers master-1:6767,master-2:6768,master-3:6769
```

#### 2. Create Workflow

```bash
curl -X POST http://localhost:6767/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ci-pipeline",
    "tasks": [
      {"id": "build", "dependencies": []},
      {"id": "test", "dependencies": ["build"]},
      {"id": "deploy", "dependencies": ["test"]}
    ]
  }'
```

#### 3. Check Cluster Status

```bash
# View current leader
curl http://localhost:6767/api/v1/cluster/leader

# View all nodes
curl http://localhost:6767/api/v1/cluster/nodes

# View workflow execution
curl http://localhost:6767/api/v1/workflows/ci-pipeline
```

### Verification

```bash
cd rust
./verify-distributed-build.sh
```

### Documentation

- **[Distributed Execution System](docs/DISTRIBUTED_EXECUTION_SYSTEM.md)** - Complete system architecture and API reference
- **[Quick Start Guide](docs/DISTRIBUTED_QUICK_START.md)** - 5-minute setup guide
- **[Implementation Status](docs/DISTRIBUTED_EXECUTION_STATUS.md)** - Technical details and progress
- **[README](rust/README_DISTRIBUTED.md)** - Feature overview and examples

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│   agentflow-master (单一二进制)             │
│                                             │
│  ┌───────────────────────────────────────┐ │
│  │  HTTP/WebSocket API (Axum)            │ │
│  └───────────┬───────────────────────────┘ │
│              ↓                              │
│  ┌───────────────────────────────────────┐ │
│  │  Task Scheduler                       │ │
│  └───────────┬───────────────────────────┘ │
│              ↓                              │
│  ┌───────────────────────────────────────┐ │
│  │  TaskExecutor (tokio::process)       │ │
│  │  - Execute Claude CLI                  │ │
│  │  - ProcessKiller (timeout)            │ │
│  │  - PromptBuilder                       │ │
│  └───────────┬───────────────────────────┘ │
│              ↓                              │
│  ┌───────────────────────────────────────┐ │
│  │  MemoryCore (SQLite)                  │ │
│  │  - Vector indexing                    │ │
│  │  - Semantic search                    │ │
│  └───────────┬───────────────────────────┘ │
│              ↓                              │
│  ┌───────────────────────────────────────┐ │
│  │  Sandbox (Security)                   │ │
│  │  - Path whitelist                     │ │
│  │  - Symlink protection                 │ │
│  └───────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
              ↓
    ┌──────────────────┐
    │  claude CLI      │
    └──────────────────┘
```

## 📦 Project Structure

```
rust/
├── agentflow-core/          # Core library
│   ├── src/
│   │   ├── types.rs        # Shared types
│   │   ├── executor/       # Task execution engine
│   │   ├── memory/         # Memory system
│   │   └── sandbox/        # Security sandbox
│   └── Cargo.toml
│
└── agentflow-master/        # Master server
    ├── src/
    │   ├── main.rs         # Entry point
    │   ├── config.rs       # Configuration
    │   ├── executor.rs     # Executor integration
    │   ├── memory_core.rs  # Memory integration
    │   └── routes/         # API routes
    └── Cargo.toml
```

## 📚 Documentation

### Getting Started
- **[RUST_V3_QUICKSTART.md](RUST_V3_QUICKSTART.md)** - Quick start guide
- **[CONFIGURATION.md](docs/CONFIGURATION.md)** - Complete configuration reference
- **[ZHIPU_INTEGRATION.md](docs/ZHIPU_INTEGRATION.md)** - Zhipu AI integration guide

### Technical Details
- **[TEAM_A_IMPLEMENTATION_REPORT.md](docs/TEAM_A_IMPLEMENTATION_REPORT.md)** - Execution engine report
- **[EXECUTOR_QUICK_REFERENCE.md](docs/EXECUTOR_QUICK_REFERENCE.md)** - Executor API reference
- **[EXECUTOR_EXAMPLES.md](docs/EXECUTOR_EXAMPLES.md)** - Executor usage examples
- **[API.md](rust/agentflow-master/API.md)** - REST API documentation

### Historical
- **[RUST_V3_FINAL_REPORT.md](docs/archive/v3-development/RUST_V3_FINAL_REPORT.md)** - Final development report

## 🔧 Configuration

### Quick Configuration

Create `~/.agentflow/config.toml`:

```toml
[server]
port = 6767

[database]
url = "sqlite://agentflow.db"

[executor]
max_concurrent_tasks = 10
task_timeout = 300

[memory]
backend = "memory"
default_ttl = 3600

[sandbox]
enabled = true
allow_network = false
```

### Environment Variables

Alternatively, use environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTFLOW_SERVER_PORT` | 6767 | Server port |
| `AGENTFLOW_SERVER_ADDR` | 0.0.0.0 | Server address |
| `AGENTFLOW_DATABASE_URL` | sqlite://agentflow.db | Database connection |
| `AGENTFLOW_LOG_LEVEL` | info | Log level |
| `AGENTFLOW_MAX_CONCURRENT_TASKS` | 10 | Max concurrent tasks |
| `SQLX_OFFLINE` | true | SQLx offline mode |

### Configuration Priority

1. Command-line arguments (highest)
2. Environment variables
3. Configuration file (`~/.agentflow/config.toml`)
4. Default values (lowest)

For complete configuration reference, see **[CONFIGURATION.md](docs/CONFIGURATION.md)**.

## 🔒 Security Features

- ✅ **Path Whitelist** - Only allows access to specified directories
- ✅ **Path Traversal Protection** - Detects and blocks `../` attacks
- ✅ **Symlink Protection** - Recursive symlink resolution with depth limit
- ✅ **Process Timeout** - Automatic process termination (SIGTERM → wait → SIGKILL)
- ✅ **Process Group Cleanup** - Cleans up all child processes

## 📊 Performance

- **Memory Usage**: < 100MB (idle)
- **Startup Time**: < 1 second
- **Concurrent Tasks**: 5+ (configurable)
- **API Response**: < 10ms

## 🆚 Historical Context

AgentFlow v0.2.1 is the result of multiple iterations:

- **v1.0**: Initial Node.js version with Master + Worker architecture
- **v2.0**: Added Go version, memory system, and skills integration
- **v3.0**: Complete Rust rewrite with single-process architecture

**Previous versions** (Node.js and Go) are **archived** in `docs/archive/old-versions/` for historical reference.

The Rust v0.2.1 version supersedes all previous versions with:
- **Simpler architecture** - Single process instead of Master + Worker
- **Better performance** - Tokio async runtime, lower memory footprint
- **Zero dependencies** - No need for Node.js runtime
- **Enhanced security** - Complete sandbox and process isolation
- **Cleaner codebase** - 176KB of source code vs 812KB (Node.js)
- **Cloud integration** - Webhook support for AI platforms like Zhipu AI

## 🛠️ Development

### Build

```bash
cd rust
cargo build --release
```

### Test

```bash
cargo test
```

### Run

```bash
cargo run --bin agentflow-master
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

For issues, questions, or suggestions, please open an issue on GitHub.

---

**Made with ❤️ and Rust by the AgentFlow Team**
