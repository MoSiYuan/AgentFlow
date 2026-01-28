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

### 1. Install Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
```

### 2. Build AgentFlow

```bash
cd rust
export SQLX_OFFLINE=true
cargo build --release
```

### 3. Run AgentFlow

```bash
./target/release/agentflow-master
```

Server will start on `http://localhost:6767`

## 📝 Usage Examples

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

- **[RUST_V3_QUICKSTART.md](RUST_V3_QUICKSTART.md)** - Quick start guide
- **[RUST_V3_IMPLEMENTATION.md](RUST_V3_IMPLEMENTATION.md)** - Implementation details
- **[RUST_V3_FINAL_REPORT.md](RUST_V3_FINAL_REPORT.md)** - Final report

## 🔧 Configuration

Environment variables (`.env` file):

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTFLOW_PORT` | 6767 | Server port |
| `DATABASE_URL` | sqlite://agentflow.db | Database connection |
| `RUST_LOG` | info | Log level |
| `SQLX_OFFLINE` | true | SQLx offline mode |

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

AgentFlow v3 is the result of multiple iterations:

- **v1.0**: Initial Node.js version with Master + Worker architecture
- **v2.0**: Added Go version, memory system, and skills integration
- **v3.0**: Complete Rust rewrite with single-process architecture

**Previous versions** (Node.js and Go) are **archived** in `docs/archive/old-versions/` for historical reference.

The Rust v3 version supersedes all previous versions with:
- **Simpler architecture** - Single process instead of Master + Worker
- **Better performance** - Tokio async runtime, lower memory footprint
- **Zero dependencies** - No need for Node.js runtime
- **Enhanced security** - Complete sandbox and process isolation
- **Cleaner codebase** - 176KB of source code vs 812KB (Node.js)

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
