# AGENTFLOW.md Feature Demonstration

## Test Scenario

This demonstrates how the AGENTFLOW.md project configuration feature works.

## Step 1: Create Test Configuration

Create a test `AGENTFLOW.md` file in the project root:

```markdown
# Test Project Configuration

## Build Commands
- Use `cargo build` not `make`
- Example: `cargo build --release`

## Test Commands
- Run `cargo test` for unit tests
- Skip integration tests with `--skip integration`

## Important Notes
- This is a test configuration
- It should appear in the prompt under "## 项目专属配置"
```

## Step 2: Create Test Prompt

When PromptBuilder builds a prompt, it will include:

```
## 系统指令

## Role
你是 AgentFlow Execution Engine...

## Core Directives
1. **Action First**: ...

## 项目专属配置

# Test Project Configuration

## Build Commands
- Use `cargo build` not `make`
- Example: `cargo build --release`

## Test Commands
- Run `cargo test` for unit tests
- Skip integration tests with `--skip integration`

## Important Notes
- This is a test configuration
- It should appear in the prompt under "## 项目专属配置"

## 当前任务

Implement a new feature...
```

## Step 3: Verify Integration

The project configuration is:
1. ✅ Read from workspace path
2. ✅ Formatted with "## 项目专属配置" header
3. ✅ Inserted after system instruction
4. ✅ Placed before memory and task sections
5. ✅ Only included if file exists (no errors if missing)

## Code Flow

```rust
// 1. Create builder (uses current directory)
let builder = PromptBuilder::new();

// 2. Build prompt
let prompt = builder.build("Fix bug", &memories);

// Internally:
// - build_system_section() → "## 系统指令\n..."
// - load_project_config() → Some(content) or None
// - build_project_config_section(content) → "## 项目专属配置\n..."
// - build_memory_section(memories) → "## 相关上下文\n..."
// - build_task_section(task) → "## 当前任务\n..."
// - Join with "\n\n"
```

## File Locations

- **Implementation**: `rust/agentflow-core/src/executor/prompt_builder.rs`
- **Example**: `templates/AGENTFLOW.md.example`
- **Your Config**: `AGENTFLOW.md` (project root)

## Cleanup

To remove test config:
```bash
rm AGENTFLOW.md
```

The PromptBuilder will continue to work normally without the project config section.
