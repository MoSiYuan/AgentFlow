# AgentFlow Workspace Rules

## Agent Behavior Guidelines

### 1. Task Execution

**DO:**
- Always clarify requirements before starting
- Break complex tasks into smaller steps
- Use checkpoints for long-running tasks
- Release Git locks after completing work
- Document decisions and assumptions

**DON'T:**
- Don't skip testing phases
- Don't ignore error messages
- Don't modify files without acquiring locks
- Don't make assumptions without clarification

### 2. File Operations

**Git Lock Management:**
```typescript
// ALWAYS acquire lock before editing
await acquireGitLock({
  file_path: '/src/app.ts',
  lock_type: 'write',
  timeout: 30000
});

// ... perform edits ...

// ALWAYS release lock after completion
await releaseGitLock(taskId, '/src/app.ts');
```

**Safe Concurrent Work:**
- Worker 1 edits `app.ts` ✅ (has lock)
- Worker 2 edits `utils.ts` ✅ (different file)
- Worker 3 edits `app.ts` ⏳ (waits for lock)

### 3. Short-term Memory

**Best Practices:**
```typescript
// Remember important context
worker.remember('project_type', 'typescript');
worker.remember('framework', 'react', 3600); // 1 hour TTL
worker.remember('user_choice', 'option_a');

// Recall when needed
const framework = worker.recall('framework');
if (framework === 'react') {
  // Use React-specific approach
}
```

**What to Remember:**
- User preferences and choices
- Important project decisions
- Configuration values
- File paths and locations

### 4. Checkpoint Usage

**Automatic Checkpoints:**
```typescript
// Save at key points
await saveCheckpoint(task, 'step_1_complete', {
  files_processed: 50,
  errors: 0
});

// Resume from checkpoint if task fails
// → No need to redo completed work
```

**When to Create Checkpoints:**
- After completing major steps
- Before risky operations
- After processing batches of items
- When user input is received

### 5. Error Handling

**Retry Strategy:**
```typescript
// Configure retry behavior
const workerConfig = {
  retry_on_failure: true,
  max_retries: 3,
  backoff_multiplier: 2
};
```

**Error Recovery:**
1. Log error details
2. Save checkpoint before retry
3. Increment retry counter
4. Use exponential backoff
5. Fail gracefully after max retries

### 6. Communication

**With Users:**
- Ask clarifying questions when uncertain
- Provide progress updates for long tasks
- Report errors with context
- Suggest solutions when possible

**With Other Agents:**
- Share relevant context
- Respect file locks
- Follow task dependencies
- Update shared memory

## Task Priorities

**Priority Levels:**
1. **Critical** (100): Security fixes, data loss prevention
2. **High** (75): Production bugs, performance issues
3. **Normal** (50): Features, improvements
4. **Low** (25): Documentation, optimizations

**Example:**
```bash
agentflow create "Fix security vulnerability" \
  --priority 100 \
  -d "Patch SQL injection in auth API"
```

## Security Guidelines

**Prohibited Actions:**
- ❌ Never expose API keys or secrets
- ❌ Never execute untrusted code
- ❌ Never disable security checks
- ❌ Never commit sensitive data

**Required Actions:**
- ✅ Validate all inputs
- ✅ Use environment variables for secrets
- ✅ Follow OWASP guidelines
- ✅ Run security audits before deployment

## Code Quality Standards

**Required:**
- Follow project style guide
- Write tests for new code
- Update documentation
- Handle errors appropriately

**Preferred:**
- Add comments for complex logic
- Use meaningful variable names
- Keep functions focused
- DRY (Don't Repeat Yourself)

## Performance Guidelines

**Optimization Tips:**
- Use parallel execution for independent tasks
- Cache expensive operations
- Batch similar operations
- Profile before optimizing

**Example:**
```bash
# Good: Parallel execution
agentflow create "Test all" -d "
  - Parallel run all test suites
"

# Avoid: Sequential when parallel possible
for suite in test_suits; do
  agentflow create "Test $suite"  # Sequential
done
```
