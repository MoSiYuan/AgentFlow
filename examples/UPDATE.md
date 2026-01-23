# Examples Directory Update Summary

## Changes Made

### Removed (Archived)

Old Python examples based on Go implementation:
- `git_integration_example.py` → `archives/old-examples/`
- `mobile_client.py` → `archives/old-examples/`

These examples were obsolete because:
1. Based on Go version of AgentFlow
2. Used Python SDK that no longer exists
3. Don't align with current Node.js/CLI approach

### Added

New examples aligned with current AgentFlow v2.0.0:

#### 1. **quick-start.sh** - Basic Usage
```bash
./examples/quick-start.sh
```
- Checks AgentFlow installation
- Verifies Master server
- Creates first task
- Monitors progress

#### 2. **parallel-tasks.sh** - Parallel Execution
```bash
./examples/parallel-tasks.sh
```
- Creates multiple tasks simultaneously
- Demonstrates concurrent execution
- Shows Git lock mechanism in action

#### 3. **programmatic-usage.js** - Node.js API
```bash
node examples/programmatic-usage.js
```
- Uses `@agentflow/skill` package
- Health checks
- Parallel task creation
- Status monitoring

#### 4. **git-workflow.sh** - Git Integration
```bash
./examples/git-workflow.sh
```
- Demonstrates Git lock mechanism
- Shows concurrent file editing
- Conflict prevention

#### 5. **README.md** - Documentation
- Usage instructions for all examples
- Real-world scenarios (CI/CD, code review, multi-agent)
- Troubleshooting guide
- Links to more resources

## Key Improvements

### Before
- ❌ Python-based (doesn't match Node.js implementation)
- ❌ Go version specific
- ❌ Outdated SDK usage
- ❌ No clear documentation

### After
- ✅ Shell scripts (universal, easy to run)
- ✅ Node.js programmatic example
- ✅ Aligned with current CLI
- ✅ Comprehensive README
- ✅ Real-world scenarios included
- ✅ Executable permissions set

## Usage

### Run All Examples

```bash
# Make scripts executable (already done)
chmod +x examples/*.sh

# Quick start
./examples/quick-start.sh

# Parallel tasks
./examples/parallel-tasks.sh

# Git workflow
./examples/git-workflow.sh

# Programmatic
node examples/programmatic-usage.js
```

## Example Content Highlights

### Real-World Scenarios in README

1. **CI/CD Pipeline**
   - Parallel test execution
   - Build automation
   - Deployment staging

2. **Code Review Workflow**
   - Security review
   - Performance analysis
   - Code quality check

3. **Multi-Agent Collaboration**
   - Developer agent
   - Tester agent
   - Reviewer agent

### Advanced Topics

- Using with Claude Code
- Task dependencies
- Custom workflows
- Troubleshooting common issues

## Testing

All examples have been tested and verified:
- ✅ Scripts are executable
- ✅ Syntax is correct
- ✅ Documentation is complete
- ✅ References are up-to-date

## Future Enhancements

Potential additions:
- Docker deployment example
- Kubernetes workflow example
- Monitoring/alerting setup
- Integration with CI/CD platforms (GitHub Actions, GitLab CI)

---

**Updated**: 2026-01-23
**AgentFlow Version**: 2.0.0
