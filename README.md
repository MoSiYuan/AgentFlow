# AgentFlow - AI Agent Task Collaboration System

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Node.js](https://img.shields.io/badge/Node.js-18%2B-brightgreen.svg)](https://nodejs.org/)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8E.svg)](https://golang.org/)

Master-Worker architecture for asynchronous AI task collaboration with 100% API-compatible dual-language implementation.

## Quick Start

```bash
# Install
npm install -g @agentflow/skill

# Create task
agentflow create "Run tests" -d "npm test"

# List tasks
agentflow list

# Check status
agentflow status TASK-00000001
```

## Features

- ✅ **Task Orchestration** - DAG, parallel, sequential workflows
- ✅ **Checkpoint Mechanism** - State persistence and recovery
- ✅ **Git Locks** - Prevent concurrent conflicts
- ✅ **Task Versioning** - Upgrade and history tracking
- ✅ **Local CLI Execution** - Direct command execution
- ✅ **Short-term Memory** - Agent context management

## Architecture

```
┌─────────────┐
│   Master    │ Task Scheduler
│  (Node.js)  │ HTTP API + WebSocket
└─────┬───────┘
      │
      ├──▶ ┌─────────────┐
      │    │   Worker 1  │ Local CLI
      │    │  (Node.js)  │ Skills Execution
      │    └─────────────┘
      │
      ├──▶ ┌─────────────┐
      │    │   Worker 2  │ Cloud SDK
      │    │  (Node.js)  │ AI Processing
      │    └─────────────┘
      │
      └──▶ ...
```

## Usage

### CLI

```bash
agentflow create "My Task" -d "Description"
agentflow list --status pending
agentflow exec "npm run build"
```

### Programmatic

```typescript
import { AgentFlowSkill } from '@agentflow/skill';

const skill = new AgentFlowSkill({
  master_url: 'http://localhost:8848'
});

// Create task
const taskId = await skill.createTask({
  title: 'Deploy',
  description: 'Build and deploy app'
});

// Parallel execution
await skill.executeParallel([
  { title: 'Test', description: 'npm test' },
  { title: 'Lint', description: 'npm run lint' }
]);
```

## Documentation

- [Skill Usage](docs/SKILL.md)
- [API Reference](docs/API.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Getting Started](docs/GETTING_STARTED.md)

## Project Structure

```
AgentFlow/
├── nodejs/              # Node.js implementation
│   ├── packages/
│   │   ├── master/      # Master server
│   │   ├── worker/      # Worker (CLI execution)
│   │   ├── database/    # SQLite database layer
│   │   ├── shared/      # Type definitions
│   │   ├── skill/       # CLI skill ⭐ NEW
│   │   └── cli/         # Main CLI
│   └── test-*.js        # Integration tests
├── golang/              # Go implementation
│   ├── master/          # Go Master server
│   └── worker/          # Go Worker
└── docs/                # Documentation
```

## Installation

### From Source

```bash
# Node.js version
cd nodejs
npm install
npm run build
npm link

# Go version
cd golang
make build
make install
```

### Using Skill Package

```bash
cd nodejs/packages/skill
npm link
agentflow --help
```

## Environment Variables

```bash
export AGENTFLOW_MASTER_URL="http://localhost:8848"
export AGENTFLOW_GROUP="default"
export ANTHROPIC_API_KEY="sk-ant-..."  # For AI features
```

## Development

```bash
# Install dependencies
cd nodejs && pnpm install

# Build all packages
npm run build

# Run tests
npm test

# Start Master
node packages/master/dist/index.js

# Start Worker
node packages/worker/dist/index.js
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built for [Claude Code](https://github.com/anthropics/claude-code)
- Inspired by modern task orchestration systems
- Powered by [Anthropic Claude](https://www.anthropic.com/claude)

---

**Version**: 2.0.0 | **Status**: Production Ready
