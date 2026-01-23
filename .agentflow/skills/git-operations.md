# Git Operations Skill

## Description

Automated Git workflow management including commits, branch management, and conflict resolution.

## Capabilities

- Automated commits with conventional messages
- Branch creation and management
- Merge and conflict resolution
- PR creation and review
- Git lock management for parallel work

## Usage Examples

```bash
# Create commit with conventional format
agentflow create "Commit changes" \
  -d "/commit -m 'feat: add user authentication'"

# Create PR
agentflow create "Create PR" \
  -d "/pr create --title 'Feature: Auth' --body 'Implementation complete'"

# Merge with conflict resolution
agentflow create "Merge feature" \
  -d "Merge feature/auth into main with automatic conflict resolution"
```

## Git Lock Integration

When multiple agents work concurrently, AgentFlow automatically manages file locks:

```typescript
// Agent 1: Working on app.ts
await acquireGitLock({ file_path: 'src/app.ts', lock_type: 'write' });

// Agent 2: Wants to edit same file
// → Automatically waits or works on different file

// Lock expires after 30 minutes
// → Automatic cleanup of stale locks
```

## Best Practices

1. **Use Conventional Commits**: `feat:`, `fix:`, `docs:`, etc.
2. **Lock Files**: Always acquire locks before editing
3. **Test Before Push**: Run tests in parallel before committing
4. **Document Changes**: Update README with notable changes

## Workflow Templates

### Feature Branch Workflow
```bash
agentflow create "Feature implementation" -d "
  1. Create branch feature/user-auth
  2. Implement authentication
  3. Run tests (npm test)
  4. Create PR to main
"
```

### Parallel Development
```bash
# Create multiple tasks for different files
agentflow create "Implement auth" -d "Work on auth.ts"
agentflow create "Implement UI" -d "Work on login.tsx"
agentflow create "Write tests" -d "Work on auth.test.ts"
# All run in parallel with automatic file locking
```
