# AgentFlow Skill - Quick Reference

## Installation

```bash
npm install -g @agentflow/skill
# or
cd /Users/jiangxiaolong/work/project/AgentFlow/nodejs/packages/skill
npm link
```

## CLI Commands

```bash
# Create task
agentflow create "Task title" -d "Description"

# List tasks
agentflow list
agentflow list --status pending

# Task status
agentflow status TASK-00000001

# Execute shell command
agentflow exec "npm test"

# Health check
agentflow health
```

## Programmatic Usage

```typescript
import { AgentFlowSkill } from '@agentflow/skill';

const skill = new AgentFlowSkill({
  master_url: 'http://localhost:8848',
  group_name: 'default'
});

// Create task
const taskId = await skill.createTask({
  title: 'Run Tests',
  description: 'npm test'
});

// Parallel execution
const ids = await skill.executeParallel([
  { title: 'Task 1', description: '...' },
  { title: 'Task 2', description: '...' }
]);

// List tasks
const tasks = await skill.listTasks('pending');
```

## Features

- ✅ Task orchestration (DAG, parallel, sequential)
- ✅ Checkpoint mechanism
- ✅ Git locks
- ✅ Task versioning
- ✅ Local CLI execution
- ✅ Short-term memory

## Environment

```bash
export AGENTFLOW_MASTER_URL="http://localhost:8848"
export AGENTFLOW_GROUP="default"
```

## Documentation

- [API Reference](API.md)
- [Architecture](ARCHITECTURE.md)
- [Getting Started](GETTING_STARTED.md)

---

*Version: 1.0.0 | License: MIT*
