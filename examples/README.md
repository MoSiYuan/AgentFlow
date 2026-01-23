# AgentFlow Examples

This directory contains practical examples for using AgentFlow.

## Quick Start

### 1. Basic Usage ([quick-start.sh](quick-start.sh))

```bash
chmod +x examples/quick-start.sh
./examples/quick-start.sh
```

Demonstrates:
- Checking AgentFlow installation
- Verifying Master server status
- Creating your first task
- Monitoring task progress

### 2. Parallel Execution ([parallel-tasks.sh](parallel-tasks.sh))

```bash
chmod +x examples/parallel-tasks.sh
./examples/parallel-tasks.sh
```

Demonstrates:
- Creating multiple tasks simultaneously
- Parallel task execution
- Monitoring concurrent work

### 3. Programmatic Usage ([programmatic-usage.js](programmatic-usage.js))

```bash
node examples/programmatic-usage.js
```

Demonstrates:
- Using AgentFlow with Node.js
- Health checks
- Parallel task creation
- Task status monitoring

### 4. Git Integration ([git-workflow.sh](git-workflow.sh))

```bash
chmod +x examples/git-workflow.sh
./examples/git-workflow.sh
```

Demonstrates:
- Git lock mechanism
- Concurrent file editing
- Conflict prevention

## Real-World Scenarios

### CI/CD Pipeline

```bash
#!/bin/bash
# ci-pipeline.sh

# Run tests in parallel
agentflow create "Unit tests" -d "npm run test:unit"
agentflow create "Integration tests" -d "npm run test:integration"
agentflow create "E2E tests" -d "npm run test:e2e"

# Build
agentflow create "Build" -d "npm run build"

# Deploy
agentflow create "Deploy staging" -d "./deploy.sh staging"

# Monitor progress
watch -n 5 "agentflow list --status running"
```

### Code Review Workflow

```bash
#!/bin/bash
# review-workflow.sh

# Create review tasks
agentflow create "Security review" \
  -d "Review for security vulnerabilities"

agentflow create "Performance review" \
  -d "Analyze performance bottlenecks"

agentflow create "Code quality review" \
  -d "Check code quality and best practices"

# Check results
sleep 10
agentflow list --status completed
```

### Multi-Agent Collaboration

```bash
#!/bin/bash
# multi-agent-workflow.sh

# Developer agent tasks
agentflow create "Implement feature" -d "Use developer agent"
agentflow create "Write tests" -d "Use tester agent"

# Reviewer agent task
agentflow create "Review PR" -d "Use reviewer agent"

# All agents work with Git locks preventing conflicts
```

## Advanced Examples

### Using with Claude Code

```typescript
// In Claude Code or with Claude CLI
const skill = new AgentFlowSkill();

// Create task from AI
await skill.createTask({
  title: 'Refactor code',
  description: 'Improve code structure'
});
```

### Task Dependencies

```bash
# Create dependent tasks
BUILD_ID=$(agentflow create "Build" -d "npm run build")
TEST_ID=$(agentflow create "Test" -d "npm test")

# Master orchestrates execution order
# Build → Test
```

### Custom Workflows

See [`.agentflow/workflows/](../.agentflow/workflows/) for predefined workflow templates.

## Prerequisites

1. **Install AgentFlow**
   ```bash
   npm install -g @agentflow/skill
   ```

2. **Start Master Server**
   ```bash
   cd /path/to/AgentFlow/nodejs
   node packages/master/dist/index.js
   ```

3. **Verify Installation**
   ```bash
   agentflow info
   ```

## Troubleshooting

### Master server not running
```bash
# Start Master
cd /path/to/AgentFlow/nodejs
node packages/master/dist/index.js

# Check health
curl http://localhost:6767/health
```

### Tasks not executing
```bash
# Check workers
curl http://localhost:6767/api/v1/workers

# Check pending tasks
curl http://localhost:6767/api/v1/tasks/pending
```

### Git lock conflicts
```bash
# View active locks
sqlite3 .claude/cpds-manager/agentflow.db \
  "SELECT * FROM git_locks WHERE status = 'active'"

# Locks auto-expire after 30 minutes
```

## More Examples

- [`.agentflow/examples/`](../.agentflow/examples/) - More usage examples
- [docs/AI_INTEGRATION.md](../docs/AI_INTEGRATION.md) - Complete AI integration guide
- [docs/SKILL.md](../docs/SKILL.md) - Command reference

## Contributing

Have a great example? Submit a PR!

---

**AgentFlow Version**: 2.0.0
