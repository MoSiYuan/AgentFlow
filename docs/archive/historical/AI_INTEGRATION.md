# AgentFlow AI Integration Guide

## Quick Start

```bash
# Install AgentFlow skill
npm install -g @agentflow/skill

# Start Master server
cd /Users/jiangxiaolong/work/project/AgentFlow/nodejs
node packages/master/dist/index.js

# Create task via CLI
agentflow create "Analyze code quality" -d "Review src/ directory"
```

## Claude CLI Integration

### 1. Basic Usage

```bash
# In Claude Code, use the agentflow skill
/agentflow create "Run tests" -d "npm test"

# Check status
/agentflow list

# View task details
/agentflow status TASK-00000001
```

### 2. Advanced Patterns

#### Pattern 1: Parallel Execution

```typescript
// Create multiple tasks in parallel
await skill.executeParallel([
  { title: 'Frontend tests', description: 'cd frontend && npm test' },
  { title: 'Backend tests', description: 'cd backend && npm test' },
  { title: 'API tests', description: 'cd api && npm test' }
]);
```

#### Pattern 2: Sequential Workflow

```typescript
// Create dependent tasks
const buildId = await skill.createTask({
  title: 'Build Project',
  description: 'npm run build'
});

const testId = await skill.createTask({
  title: 'Run Tests',
  description: 'npm test'
});

// Add dependency relationship
// Master will ensure build completes before testing
```

#### Pattern 3: Task with Context

```bash
# Provide context from previous task
agentflow create "Deploy" -d "Deploy the build artifacts from previous task"
```

## Configuration

### Environment Variables

```bash
# Master server URL
export AGENTFLOW_MASTER_URL="http://localhost:6767"

# Default task group
export AGENTFLOW_GROUP="default"

# Anthropic API key (for AI features)
export ANTHROPIC_API_KEY="sk-ant-..."
```

### In Code

```typescript
import { AgentFlowSkill } from '@agentflow/skill';

const skill = new AgentFlowSkill({
  master_url: process.env.AGENTFLOW_MASTER_URL,
  group_name: process.env.AGENTFLOW_GROUP
});
```

## Common Use Cases

### 1. Code Review Workflow

```bash
# Step 1: Analyze code
agentflow create "Code analysis" -d "Review code quality and suggest improvements"

# Step 2: Run tests
agentflow create "Test suite" -d "npm test && npm run coverage"

# Step 3: Build
agentflow create "Build" -d "npm run build"
```

### 2. Deployment Pipeline

```typescript
// Create deployment pipeline
const tasks = [
  { title: 'Run tests', description: 'npm test' },
  { title: 'Build', description: 'npm run build' },
  { title: 'Deploy', description: 'npm run deploy' }
];

for (const task of tasks) {
  const taskId = await skill.createTask(task);
  console.log(`Created task: ${taskId}`);
}
```

### 3. Data Processing

```bash
# Process multiple files in parallel
for file in data/*.json; do
  agentflow create "Process $file" -d "node process.js $file"
done
```

## Agent Short-term Memory

Workers can remember context during task execution:

```typescript
// Worker saves state
worker.remember('file_path', '/src/app.ts');
worker.remember('user_choice', 'option_a', 3600); // 1 hour TTL

// Later in execution
const path = worker.recall('file_path');
const choice = worker.recall('user_choice');
```

## Checkpoints & Recovery

Tasks automatically save checkpoints for recovery:

```typescript
// Automatic checkpoint creation
await skill.createTask({
  title: 'Long running task',
  description: 'Process 1000 files'
});

// If task fails, Worker resumes from last checkpoint
// No manual intervention needed
```

## Git Integration

### Preventing Conflicts

AgentFlow automatically locks files during task execution:

```typescript
// Worker acquires lock
await acquireGitLock({
  file_path: '/src/app.ts',
  lock_type: 'write'
});

// ... process file ...

// Worker releases lock
await releaseGitLock(taskId, '/src/app.ts');
```

### Safe Concurrent Work

Multiple workers can work on different files simultaneously:

```
Worker 1: app.ts    ← Lock acquired ✅
Worker 2: utils.ts ← Lock acquired ✅
Worker 3: app.ts    ← Lock waits ⏳
```

## Error Handling

### Task Retry Logic

```typescript
// Configure retry behavior
const workerConfig = {
  retry_on_failure: true,
  max_retries: 3
};

// Worker automatically retries failed tasks
// With exponential backoff
```

### Checkpoint Recovery

```typescript
// Task fails at step 5/10
// → Checkpoint saved at step 4

// Resume from checkpoint
// → Starts from step 4, not step 1
```

## Best Practices

### 1. Task Granularity

**Good** ✅
```typescript
{
  title: "Run unit tests",
  description: "cd frontend && npm test"
}
```

**Too Broad** ❌
```typescript
{
  title: "CI/CD pipeline",
  description: "Run all tests, build, deploy, notify team"
}
```

### 2. Descriptive Titles

```bash
# Good
agentflow create "Fix login bug" -d "Update authentication logic"

# Bad
agentflow create "Task 1" -d "Do something"
```

### 3. Clear Descriptions

```bash
# Include commands, expected outputs, etc.
agentflow create "Deploy to staging" -d "
  1. Build Docker image
  2. Push to registry
  3. Update Kubernetes deployment
"
```

## Troubleshooting

### Issue: Tasks not executing

**Check:**
```bash
# Is Master running?
curl http://localhost:6767/health

# Are workers registered?
curl http://localhost:6767/api/v1/workers

# Any pending tasks?
curl http://localhost:6767/api/v1/tasks/pending
```

### Issue: Claude CLI not working

**Solution:**
```bash
# Verify Claude CLI installation
which claude

# Check API key
echo $ANTHROPIC_API_KEY

# Test CLI manually
claude --help
```

### Issue: File lock conflicts

**Solution:**
```bash
# View active locks
sqlite3 .claude/cpds-manager/agentflow.db \
  "SELECT * FROM git_locks WHERE status = 'active'"

# Release stale locks (if needed)
# Locks auto-expire after 30 minutes
```

## Examples

### Example 1: Automated Testing

```bash
#!/bin/bash
# test-all.sh

agentflow create "Frontend Tests" -d "cd frontend && npm test"
agentflow create "Backend Tests" -d "cd backend && npm test"
agentflow create "E2E Tests" -d "npm run test:e2e"

# Wait for completion
sleep 60

# Check results
agentflow list --status completed
```

### Example 2: Code Analysis

```bash
# Analyze codebase
agentflow create "Code review" -d "
  Analyze the codebase for:
  1. Security vulnerabilities
  2. Performance issues
  3. Code smell
  4. Suggest refactoring opportunities
"
```

### Example 3: Batch Processing

```typescript
// Process multiple files
const files = await fs.readdir('data/');

const tasks = files.map(file => ({
  title: `Process ${file}`,
  description: `node processor.js data/${file}`
}));

await skill.executeParallel(tasks);
```

## Tips

1. **Start Master before creating tasks**
2. **Use task groups for isolation** (e.g., `dev`, `staging`, `prod`)
3. **Monitor task status** with `agentflow list`
4. **Review results** with `agentflow status <task-id>`
5. **Check logs** for detailed execution information

## Advanced Features

### Task Dependencies

```typescript
// Create dependent tasks
const id1 = await createTask({ title: 'Build' });
const id2 = await createTask({ title: 'Test' });

// Add relationship
addRelationship(id1, id2, 'dependency');
// Test will only run after Build completes
```

### Task Upgrade

```typescript
// Upgrade task with new requirements
upgradeTask({
  task_id: taskId,
  new_title: 'Enhanced Task v2',
  new_description: 'Updated with additional features',
  upgrade_reason: 'Client requested more features'
});
```

### Custom Workflows

```typescript
// Define complex workflow
const workflow = {
  name: 'Release Pipeline',
  tasks: [
    { title: 'Run Tests', description: 'npm test' },
    { title: 'Build', description: 'npm run build' },
    { title: 'Deploy Staging', description: './deploy.sh staging' },
    { title: 'Smoke Tests', description: './smoke-tests.sh' }
  ]
};

await skill.createWorkflow(workflow);
```

---

*Related Docs:*
- [Skill Usage](SKILL.md) - Command-line reference
- [Architecture](ARCHITECTURE.md) - System design
- [Getting Started](../archives/GETTING_STARTED.md) - Installation guide
