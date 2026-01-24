# AgentFlow Architecture

## Overview

AgentFlow is a **Master-Worker** task orchestration system with 100% API-compatible implementations in Node.js and Go.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                         │
│  - CLI Tools (@agentflow/skill)                                   │
│  - REST API                                                      │
│  - WebSocket                                                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Master Server                            │
│  - Task Scheduler (DAG orchestration)                           │
│  - State Manager (Checkpoints & Versions)                        │
│  - Lock Manager (Git locks)                                     │
│  - HTTP API + WebSocket                                         │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Worker 1  │ │   Worker 2  │ │   Worker N  │
│  (Node.js)  │ │  (Node.js)  │ │  (Node.js)  │
├─────────────┤ ├─────────────┤ ├─────────────┤
│ Local CLI   │ │ Cloud SDK   │ │ ...        │
└─────────────┘ └─────────────┘ └─────────────┘
```

## Core Components

### 1. Master Server

**Responsibilities:**
- Task scheduling and distribution
- DAG-based orchestration (Ralph pattern)
- State persistence (SQLite)
- Worker management
- API gateway

**Implementation:**
- Node.js: `nodejs/packages/master/`
- Go: `golang/master/`

### 2. Worker

**Responsibilities:**
- Execute tasks locally (CLI, skills)
- Manage short-term memory
- Create checkpoints
- Report status

**Execution Modes:**
- Local CLI: Direct command execution
- Skills: `/commit`, `/test`, `/build`
- Claude CLI: Complex AI tasks

**Implementation:**
- `nodejs/packages/worker/`

### 3. Database

**Schema:**
- `tasks` - Task definitions
- `workers` - Worker registry
- `task_relationships` - Dependencies
- `task_checkpoints` - State snapshots
- `task_versions` - Version history
- `git_locks` - File locks

**Implementation:**
- SQLite via `better-sqlite3`
- `nodejs/packages/database/`

### 4. Orchestration

**Ralph Pattern Modes:**
- **Sequential** - One by one
- **Parallel** - Concurrent execution
- **DAG** - Dependency-aware scheduling
- **Pipeline** - Data flow pipelines
- **Conditional** - Branching logic

**Implementation:**
- `nodejs/packages/master/src/orchestrator.ts`

## Key Features

### Task Orchestration

```typescript
// Create dependent tasks
const task1 = await createTask({ title: 'Build' });
const task2 = await createTask({ title: 'Test' });
addRelationship(task1, task2, 'dependency');

// DAG execution order: [[1], [2]]
```

### Checkpoint Mechanism

```typescript
// Worker auto-saves checkpoints
await saveCheckpoint(taskId, 'progress', { step: 5, data: '...' });
await restoreFromCheckpoint(checkpointId);
```

### Git Locks

```typescript
// Prevent concurrent file conflicts
acquireGitLock({ file_path: '/src/app.ts', lock_type: 'write' });
// ... work with file ...
releaseGitLock(taskId, '/src/app.ts');
```

### Task Versioning

```typescript
// Upgrade task to new version
upgradeTask({
  task_id: 1,
  new_title: 'Enhanced Task v2',
  upgrade_reason: 'Requirements changed'
});
```

## Data Flow

### 1. Task Creation
```
Client → Master API → Database → Return Task ID
```

### 2. Task Execution
```
Master → Worker (via polling) → Execute → Report Result → Update Database
```

### 3. Progress Tracking
```
Worker → WebSocket Push → Client → Real-time Updates
```

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Master | Node.js, Go | API Server |
| Worker | Node.js | Task Execution |
| Database | SQLite | State Persistence |
| CLI | TypeScript | User Interface |
| API | REST + WebSocket | Communication |

## Project Structure

```
AgentFlow/
├── .agentflow/              # Configuration & templates ⭐ NEW
│   ├── agents/              # Agent templates (developer, tester, reviewer)
│   ├── skills/              # Skill definitions (git, testing, etc.)
│   ├── workflows/           # Workflow templates (feature-development)
│   ├── examples/            # Usage examples (quick-start)
│   ├── rules/               # Workspace rules (guidelines)
│   └── config.example.json  # Configuration template
├── nodejs/                  # Node.js implementation
│   ├── packages/
│   │   ├── master/          # Master server
│   │   ├── worker/          # Worker
│   │   ├── database/        # Database layer
│   │   ├── shared/          # Shared types
│   │   └── skill/           # CLI tool
├── golang/                  # Go implementation
│   ├── master/              # Go Master
│   └── worker/              # Go Worker
└── docs/                   # Documentation
```

### `.agentflow/` Directory

Inspired by [Antigravity Kit](https://github.com/vudovn/antigravity-kit), the `.agentflow/` directory contains:

- **agents/** - Predefined agent personas with capabilities and workflows
- **skills/** - Reusable task patterns (Git operations, testing, deployment)
- **workflows/** - Multi-agent collaboration templates
- **examples/** - Real-world usage scenarios
- **rules/** - Agent behavior guidelines and best practices

**Initialization:**
```bash
agentflow init              # Create .agentflow/ directory
agentflow status            # Check installation
agentflow update            # Update templates (coming soon)
```

## Scalability

- **Horizontal Scaling**: Add more Workers
- **Task Distribution**: Automatic load balancing
- **State Management**: Centralized database
- **Fault Tolerance**: Checkpoint recovery

## Performance

| Metric | Value |
|--------|-------|
| Task Throughput | ~100-1000 tasks/min (per worker) |
| Parallel Execution | 2.9x speedup (demonstrated) |
| Checkpoint Overhead | <1ms |
| Memory Usage | ~50MB per worker |

## Security

- Task isolation (sandboxing)
- Worker authentication
- API key management (environment variables)
- Git file locking
- Audit logging

---

*Version: 2.0.0 | Last Updated: 2026-01-23*
