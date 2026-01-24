# AgentFlow - AI Agent Task Collaboration System

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8E.svg)](https://golang.org/)
[![Node.js](https://img.shields.io/badge/Node.js-20+-brightgreen.svg)](https://nodejs.org/)

**Master-Worker architecture for asynchronous AI task collaboration with 100% API-compatible dual-language implementation.**

## 🚀 Quick Start

### Option 1: Go Version (Recommended - Zero Dependencies) ⭐

```bash
# Clone repository
git clone https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow

# Use immediately (no installation needed)
./agentflow-go.sh run '["echo hello","echo world"]'

# Output:
# ✓ 准备执行 2 个任务
# ✓ [1/2] 执行: echo hello
# hello
# ✓ [1/2] ✓ 成功
# ✓ [2/2] 执行: echo world
# world
# ✓ [2/2] ✓ 成功
# ✓ 执行完成: 2/2 成功, 0 失败
```

**Features:**
- ✅ Zero dependencies (no Node.js, Python, etc.)
- ✅ Download and use, 30 seconds to start
- ✅ Single binary file (13-16MB)
- ✅ Supports macOS, Linux, Windows
- ✅ 100% API-compatible with Node.js version

### Option 2: Node.js Version (Latest: v20 LTS)

```bash
# Navigate to Node.js directory
cd nodejs

# Set Node.js 20 environment
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"

# Install dependencies
pnpm install

# Build all packages
pnpm run build

# Start Master
node packages/master/dist/index.js --port 6767 --db data/agentflow.db

# Start Worker (another terminal)
node packages/worker/dist/index.js

# Create task
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Task",
    "description": "echo Hello from Node.js 20!",
    "group_name": "default"
  }'
```

### Option 3: Simple CLI (One-Line Execution) ✨

```bash
# Execute tasks in one command
node nodejs/packages/cli/dist/index.js run "echo Hello AgentFlow!"

# With custom title
node nodejs/packages/cli/dist/index.js run "npm test" --title "Run Tests"

# Keep services running
node nodejs/packages/cli/dist/index.js run "echo test" --no-shutdown
```

## 📦 Architecture

```
AgentFlow/
├── cmd/                    # Go implementation
│   ├── agentflow-master/   # Master server (Go)
│   └── agentflow-worker/   # Worker (Go)
├── nodejs/                 # Node.js implementation
│   ├── packages/
│   │   ├── master/        # Master server (Node.js)
│   │   ├── worker/        # Worker (Node.js)
│   │   ├── local-executor/# Automatic management
│   │   ├── cli/           # Command-line tool
│   │   ├── database/      # SQLite database
│   │   ├── shared/        # Shared types
│   │   └── skill/         # Task management API
│   └── package.json
├── deployment/             # Deployment scripts
├── examples/               # Usage examples
└── docs/                  # Documentation
    ├── archive/           # Archived reports
    └── ...
```

## 🎯 Features

### Core Capabilities

- ✅ **Task Orchestration**: DAG-based task dependency resolution
- ✅ **Parallel Execution**: Multi-worker concurrent task processing
- ✅ **API Compatible**: 100% compatible between Go and Node.js versions
- ✅ **Claude CLI Integration**: Automatic use of Claude CLI for complex tasks
- ✅ **Checkpoint Support**: Task state saving and recovery
- ✅ **WebSocket Support**: Real-time task status updates
- ✅ **SQLite Database**: Persistent task storage
- ✅ **RESTful API**: Standard HTTP API for task management

### Task Execution Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  Master Server  │
│  (Port 6767)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Task Queue     │
│  (SQLite DB)    │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐ ┌────────┐
│Worker 1│ │Worker 2│  ...
└────────┘ └────────┘
```

## 📚 Documentation

### Core Documentation

- **[CLI Guide](AGENTFLOW_CLI_GUIDE.md)** - Command-line interface usage
- **[Go Version Guide](docs/GO_VERSION_GUIDE.md)** - Go implementation details
- **[Node.js Guide](docs/NODEJS_GUIDE.md)** - Node.js implementation details

### Archived Reports

Historical development and testing reports are available in [docs/archive/](docs/archive/).

## 🔧 System Requirements

### Go Version
- **OS**: macOS, Linux, Windows
- **Dependencies**: None (zero-deployment)

### Node.js Version
- **Node.js**: v20.19.6 LTS
- **pnpm**: v10.28.1+
- **better-sqlite3**: v12.6.2
- **OS**: macOS, Linux, Windows

## 🚦 Quick Reference

### Go Version Commands

```bash
# Run tasks directly
./agentflow-go.sh run '["echo hello","echo world"]'

# Start Master server
./agentflow-master-darwin-arm64 --port 6767 --db data/agentflow.db

# Start Worker
./agentflow-worker-darwin-arm64 --master http://localhost:6767
```

### Node.js Version Commands

```bash
# Start Master
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
node nodejs/packages/master/dist/index.js --port 6767

# Start Worker
node nodejs/packages/worker/dist/index.js

# Execute with CLI
node nodejs/packages/cli/dist/index.js run "echo hello"

# LocalExecutor (programmatic)
node -e "
const { LocalExecutor } = require('./nodejs/packages/local-executor/dist/index.js');
const executor = new LocalExecutor({
  masterPath: './nodejs/packages/master/dist/index.js',
  masterPort: 6767,
  dbPath: './data/agentflow.db',
  shutdownOnComplete: true
});
executor.executeOne('My Task', 'echo Hello World');
"
```

## 🔄 Version Comparison

| Feature | Go Version | Node.js Version |
|---------|-----------|----------------|
| **Dependencies** | None | Node.js 20 + pnpm |
| **Binary Size** | 13-16 MB | N/A (interpreted) |
| **Startup Time** | <100ms | ~1s |
| **Memory Usage** | ~20MB | ~80MB |
| **Platform Support** | All platforms | Node.js 18-20 |
| **Deployment** | Zero-dep | Requires Node.js 20 |
| **Performance** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Ease of Debug** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Development** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

## 🐛 Bug Fixes

### Latest Fixes (Node.js v20)

1. ✅ **Worker JSON Parse Error** - Fixed 204 No Content handling
2. ✅ **Worker Heartbeat Error** - Silently ignore connection errors during shutdown
3. ✅ **Task ID Format Inconsistency** - Unified ID format handling across APIs

## 📖 Development

### Project Status

- ✅ **Go Version**: Production-ready
- ✅ **Node.js Version**: Production-ready (v20 LTS)
- ❌ **Node.js v22/v24**: Not supported (better-sqlite3 incompatibility)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow

# Go version (ready to use)
./agentflow-go.sh run '["echo test"]'

# Node.js version (requires setup)
cd nodejs
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
pnpm install
pnpm run build
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

For issues, questions, or suggestions, please open an issue on GitHub.

---

**Made with ❤️ by the AgentFlow Team**
