# Build Scripts Cleanup Summary

## Removed Files

### 1. **Makefile** (Root Directory)
**Why:** Obsolete Go build commands
- Referenced non-existent paths (`./cmd/agentflow`, `./cmd/worker`)
- Designed for old Go implementation in `old/` directory
- Current project uses Node.js with `npm run` commands

### 2. **docs/scripts/quick-start.sh**
**Why:** Outdated quick-start script for Go version
- Used `make build` command
- Referenced `./bin/agentflow` binary
- Used old API endpoints

### 3. **docs/scripts/quick-task.sh**
**Why:** Outdated task creation script
- Used wrong API path (`/api/tasks/create` vs `/api/v1/tasks`)
- Superseded by `agentflow create` command

### 4. **scripts/deploy-*.sh** (3 deployment scripts)
**Why:** Outdated deployment scripts
- `deploy-linux.sh` (~14KB)
- `deploy-macos.sh` (~12KB)
- `deploy-windows.bat` (~13KB)

Current installation is simpler:
```bash
npm install -g @agentflow/skill
agentflow init
```

## Current Way to Build/Run

### Node.js Version (Primary)
```bash
cd nodejs

# Install dependencies
pnpm install

# Build all packages
npm run build

# Start Master
node packages/master/dist/index.js

# Start Worker
node packages/worker/dist/index.js

# Run tests
npm test
```

### Go Version (Secondary)
```bash
cd golang

# Build directly
go build -o bin/master ./cmd/master
go build -o bin/worker ./cmd/worker

# Run
./bin/master
./bin/worker
```

### Using AgentFlow CLI
```bash
# Install
npm install -g @agentflow/skill

# Initialize project
agentflow init

# Check status
agentflow info

# Create tasks
agentflow create "My Task" -d "npm test"
```

## Benefits

1. **✅ Simpler** - No build system complexity
2. **✅ Standard** - Uses npm/pnpm (Node.js standard)
3. **✅ Clearer** - Direct commands instead of Makefile targets
4. **✅ Modern** - Follows current Node.js best practices

## Files Archived To

- `archives/old-scripts/` - Deployment scripts
- `archives/quick-start.sh` - Old quick-start script

---

**Date**: 2026-01-23
**AgentFlow Version**: 2.0.0
