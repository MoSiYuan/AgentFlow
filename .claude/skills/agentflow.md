---
name: agentflow
description: Execute and manage distributed tasks using AgentFlow task management system
parameters:
  - name: operation
    type: string
    description: "Operation type: run, create, status, stats, list"
    required: true
  - name: tasks
    type: array
    description: "Array of task descriptions to execute"
    required: false
  - name: title
    type: string
    description: "Task title for better identification"
    required: false
  - name: description
    type: string
    description: "Detailed task description"
    required: false
  - name: priority
    type: string
    description: "Task priority: high, medium, low"
    required: false
  - name: group
    type: string
    description: "Worker group name for task routing"
    required: false
---

# AgentFlow - Distributed Task Management

AgentFlow is an enterprise-grade task execution system that provides distributed task queues, parallel execution, and persistent storage.

## Quick Start

### Execute Single Task (Auto-managed)

```bash
agentflow run "npm test"
```

This command:
- Automatically starts Master and Worker
- Executes the task
- Cleans up resources on completion

### Execute Multiple Tasks

```bash
agentflow run '["npm test", "npm run lint", "npm run build"]'
```

Tasks will execute in parallel by default.

### Create Task with Options

```bash
agentflow create \
  --title "Run Tests" \
  --description "Execute full test suite" \
  --priority high \
  --group production
```

## Command Reference

### Run Command

Execute tasks with automatic lifecycle management:

```bash
# Basic usage
agentflow run "command here"

# With custom title
agentflow run "npm test" --title "Unit Tests"

# Keep Master/Worker running
agentflow run "npm test" --no-shutdown

# Custom Master URL
agentflow run "npm test" --master-host localhost --master-port 6767

# Custom group
agentflow run "npm test" --group staging
```

### Create Command

Create tasks without executing:

```bash
agentflow create \
  --title "Task Title" \
  --description "Task description or command" \
  --priority high \
  --group default

# Create from JSON
agentflow create '{
  "title": "Deploy App",
  "description": "./deploy.sh",
  "priority": "high",
  "group_name": "production"
}'
```

### Status Command

Check task status:

```bash
# By task ID
agentflow status TASK-00000001

# By numeric ID
agentflow status 1

# Show latest tasks
agentflow status
```

### Stats Command

Show system statistics:

```bash
agentflow stats
```

Returns:
- Total tasks
- Tasks by status
- Active workers
- Worker groups

### List Command

List tasks with filters:

```bash
# List all tasks
agentflow list

# Filter by group
agentflow list --group production

# Filter by status
agentflow list --status pending

# Show recent tasks
agentflow list --limit 10
```

## Task Priorities

- **high**: Execute first, for critical tasks
- **medium**: Default priority
- **low**: Execute when no higher priority tasks

## Worker Groups

Groups allow you to route tasks to specific workers:

```bash
# Create task for production group
agentflow run "./deploy.sh" --group production

# Create task for staging group
agentflow run "./deploy-staging.sh" --group staging

# Create task for testing group
agentflow run "npm test" --group testing
```

## Usage Patterns

### When to Use AgentFlow

✅ **Use AgentFlow for:**
- Long-running background tasks (> 30 seconds)
- Batch operations with multiple tasks
- Parallel task execution
- Distributed workloads across multiple machines
- Production job scheduling
- CI/CD pipelines
- Scheduled/periodic tasks
- Tasks requiring persistence and retry

❌ **Use direct execution for:**
- Quick file edits
- Simple one-line commands
- Interactive debugging
- One-off operations
- Tasks requiring immediate feedback

### Common Workflows

#### 1. CI/CD Pipeline

```bash
# Create parallel test and build tasks
agentflow run '["npm test", "npm run lint"]'
agentflow run "npm run build"

# Deploy after build succeeds
agentflow run "./deploy.sh" --group production --priority high
```

#### 2. Batch Processing

```bash
# Process multiple files
for file in data/*.csv; do
  agentflow run "python process.py $file" --title "Process $file"
done

# Check all tasks
agentflow list
```

#### 3. Multi-Environment Deployment

```bash
# Deploy to staging first
agentflow run "./deploy.sh staging" --group staging

# After staging succeeds, deploy to production
agentflow run "./deploy.sh production" --group production --priority high
```

#### 4. Scheduled Tasks

```bash
# Daily backup
agentflow run "./backup.sh" --title "Daily Backup" --priority low

# Weekly maintenance
agentflow run "./maintenance.sh" --title "Weekly Maintenance" --priority medium
```

## Task Status Flow

Tasks progress through these states:

1. **pending** - Task created, waiting for worker
2. **running** - Worker is executing the task
3. **completed** - Task finished successfully
4. **failed** - Task failed (will retry if configured)

## Examples

### Example 1: Development Workflow

```bash
# Run tests in background while continuing work
agentflow run "npm test" --title "Unit Tests"

# Check status later
agentflow status

# View results when ready
agentflow logs TASK-00000001
```

### Example 2: Release Process

```bash
# Create release tasks
agentflow create --title "Run Tests" --description "npm test" --priority high
agentflow create --title "Build" --description "npm run build" --priority high
agentflow create --title "Deploy Staging" --description "./deploy-staging.sh" --group staging

# Monitor all tasks
agentflow list --status pending
agentflow list --status running
```

### Example 3: Parallel Data Processing

```bash
# Process 10 files in parallel
agentflow run '["python process1.py", "python process2.py", "python process3.py"]' \
  --title "Batch Processing" \
  --group data-processing
```

## Tips

1. **Use descriptive titles** - Makes it easier to identify tasks
2. **Set appropriate priorities** - Ensures important tasks run first
3. **Group related tasks** - Helps with resource allocation
4. **Check status regularly** - Monitor task progress
5. **Use run for simplicity** - Auto-manages Master/Worker lifecycle

## Environment Variables

```bash
# Master server URL
export AGENTFLOW_MASTER_URL="http://localhost:6767"

# Default worker group
export AGENTFLOW_GROUP="default"

# Path to agentflow CLI
export AGENTFLOW_CLI_PATH="agentflow"
```

## Troubleshooting

### "Master not running"
- The `run` command automatically starts Master
- Or manually start: `node nodejs/packages/master/dist/index.js`

### "No workers available"
- The `run` command automatically starts a Worker
- Or manually start: `node nodejs/packages/worker/dist/index.js`

### "Task stuck in pending"
- Check if Worker is running: `agentflow stats`
- Verify group names match
- Check Worker logs for errors

### "Task failed"
- Check task logs: `agentflow logs TASK-ID`
- Verify command is correct
- Check dependencies are installed

## Related Documentation

- [Full Documentation](../docs/INDEX.md)
- [CLI Guide](../AGENTFLOW_CLI_GUIDE.md)
- [Node.js Guide](../docs/NODEJS_GUIDE.md)
- [Go Guide](../docs/GO_VERSION_GUIDE.md)
