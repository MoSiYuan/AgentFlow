# Quick Start Example

## Complete Workflow Example

This example demonstrates a complete AgentFlow workflow from setup to completion.

## Step 1: Initialize AgentFlow

```bash
# Install AgentFlow skill
npm install -g @agentflow/skill

# Initialize project (creates .agentflow/ directory)
agentflow init

# Start Master server
cd /Users/jiangxiaolong/work/project/AgentFlow/nodejs
node packages/master/dist/index.js
```

## Step 2: Create Your First Task

```bash
# Simple task
agentflow create "Run tests" -d "npm test"

# Task with detailed description
agentflow create "Build project" -d "
  1. Clean previous build: rm -rf dist/
  2. Compile TypeScript: npm run build
  3. Verify output: ls -la dist/
"
```

## Step 3: Check Task Status

```bash
# List all tasks
agentflow list

# List specific status
agentflow list --status completed
agentflow list --status pending
agentflow list --status failed

# View task details
agentflow status TASK-00000001
```

## Step 4: Parallel Execution

```bash
# Create multiple tasks that run in parallel
agentflow create "Frontend tests" -d "cd frontend && npm test"
agentflow create "Backend tests" -d "cd backend && npm test"
agentflow create "API tests" -d "cd api && npm test"

# All execute simultaneously
# Git locks prevent file conflicts
```

## Step 5: Sequential Workflow

```bash
# Create dependent tasks
BUILD_ID=$(agentflow create "Build" -d "npm run build")
TEST_ID=$(agentflow create "Test" -d "npm test")

# Master orchestrates execution order
# Build completes → Test starts
```

## Real-World Example: CI/CD Pipeline

```bash
#!/bin/bash
# ci-cd.sh

# 1. Run tests
agentflow create "Unit tests" -d "npm run test:unit"
agentflow create "Integration tests" -d "npm run test:integration"

# 2. Build
agentflow create "Build" -d "npm run build"

# 3. Deploy
agentflow create "Deploy staging" -d "./deploy.sh staging"
agentflow create "Smoke tests" -d "./smoke-tests.sh"

# 4. Monitor progress
watch -n 5 "agentflow list --status running"

# 5. View results
agentflow list --status completed
```

## Example Output

```
$ agentflow list

╔═══════════════════════════════════════════════════════════╗
║                    AgentFlow Tasks                        ║
╠═══════════════════════════════════════════════════════════╣
║ TASK ID       │ TITLE             │ STATUS    │ WORKER   ║
╠═══════════════════════════════════════════════════════════╣
║ TASK-00000001 │ Run tests         │ completed │ worker-1 ║
║ TASK-00000002 │ Build project     │ completed │ worker-2 ║
║ TASK-00000003 │ Deploy staging    │ running   │ worker-3 ║
║ TASK-00000004 │ Smoke tests       │ pending   │ -        ║
╚═══════════════════════════════════════════════════════════╝
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
  title: 'Run tests',
  description: 'npm test'
});

// Parallel execution
const ids = await skill.executeParallel([
  { title: 'Frontend tests', description: 'cd frontend && npm test' },
  { title: 'Backend tests', description: 'cd backend && npm test' }
]);

// List tasks
const tasks = await skill.listTasks('completed');
console.log('Completed tasks:', tasks);
```

## Next Steps

- Explore [Agents](../agents/) - Available agent templates
- Try [Skills](../skills/) - Reusable task patterns
- Use [Workflows](../workflows/) - Predefined collaboration patterns
- Read [Documentation](../../docs/AI_INTEGRATION.md) - Complete guide
