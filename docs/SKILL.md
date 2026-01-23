# AgentFlow Skill - Quick Reference

## Installation

```bash
npm install -g @agentflow/skill
# or
cd /Users/jiangxiaolong/work/project/AgentFlow/nodejs/packages/skill
npm link
```

## CLI Commands

### Project Management

```bash
# Initialize AgentFlow in current directory
agentflow init [--force]

# Check installation status
agentflow info

# Update templates (coming soon)
agentflow update
```

### Task Management

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
  master_url: 'http://localhost:6767',
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
export AGENTFLOW_MASTER_URL="http://localhost:6767"
export AGENTFLOW_GROUP="default"
```

## Documentation

- [Architecture](ARCHITECTURE.md) - System design
- [AI Integration](AI_INTEGRATION.md) - AI guide with examples
- [Project Config](../.agentflow/README.md) - Agent templates & skills

---

*Version: 2.0.0 | License: MIT*
