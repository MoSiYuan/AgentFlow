# AGENTFLOW.md Project Configuration Feature

## Overview

The PromptBuilder now supports reading project-level configuration from `AGENTFLOW.md` files in your workspace. This allows you to provide project-specific instructions, build commands, testing guidelines, and other important information to the AI execution engine.

## Implementation Details

### Modified Files

1. **`rust/agentflow-core/src/executor/prompt_builder.rs`**
   - Added `workspace_path: PathBuf` field to `PromptBuilder` struct
   - Added `with_workspace(path: PathBuf)` constructor method
   - Added `load_project_config()` private method to read AGENTFLOW.md
   - Added `build_project_config_section()` method to format project config
   - Modified `build()` and `build_with_metadata()` methods to include project config

2. **`templates/AGENTFLOW.md.example`** (NEW)
   - Example configuration file template
   - Contains sections for build system, testing, do's/don'ts, debugging, etc.

### How It Works

1. **Prompt Creation Flow**:
   ```
   System Instruction
   ↓
   Project Configuration (if AGENTFLOW.md exists)
   ↓
   Task Metadata (if enabled)
   ↓
   Memory Context (if enabled and available)
   ↓
   Current Task
   ```

2. **File Loading**:
   - Default workspace: `std::env::current_dir()`
   - Config file: `{workspace}/AGENTFLOW.md`
   - If file doesn't exist: No error, prompt builds normally without project config

3. **Integration**:
   - Project config appears under "## 项目专属配置" section
   - Content is injected verbatim from AGENTFLOW.md
   - Positioned after system instruction, before memory/task sections

## Usage

### Basic Usage (Default Workspace)

```rust
use agentflow_core::executor::PromptBuilder;

// Uses current working directory as workspace
let builder = PromptBuilder::new();

// If AGENTFLOW.md exists in current directory, it will be loaded
let prompt = builder.build("Fix login bug", &memories);
```

### Custom Workspace

```rust
use agentflow_core::executor::PromptBuilder;
use std::path::PathBuf;

// Specify a custom workspace path
let workspace = PathBuf::from("/path/to/project");
let builder = PromptBuilder::with_workspace(workspace);

// Will load /path/to/project/AGENTFLOW.md if it exists
let prompt = builder.build("Implement feature", &memories);
```

### With Custom Config

```rust
use agentflow_core::executor::{PromptBuilder, PromptBuilderConfig};
use std::path::PathBuf;

let config = PromptBuilderConfig::default();
let mut builder = PromptBuilder::with_config(config);

// Note: with_config() uses current directory by default
// You would need to modify with_config() to accept workspace path
```

## Setting Up Your Project

1. **Copy the Example**:
   ```bash
   cp templates/AGENTFLOW.md.example AGENTFLOW.md
   ```

2. **Customize for Your Project**:
   ```markdown
   # AgentFlow 项目配置

   ## 1. 构建系统

   ### Rust Project
   - 使用 `cargo build --release` 进行编译
   - 示例：`cargo build --release && cargo test`

   ## 2. 测试工作流

   ### 运行测试
   \`\`\`bash
   cargo test --lib
   cargo test --test integration_tests
   \`\`\`

   ## 3. 关键技能

   ### ✅ DO
   - 提交前运行 `cargo clippy`
   - 遵循 Rust API guidelines

   ### ❌ DON'T
   - 不要修改 `vendor/` 目录
   - 不要提交 `Cargo.lock`

   ## 4. 项目特定信息

   ### 架构
   - MVC 架构，模块位于 `src/modules/`

   ### 依赖
   - tokio: 异步运行时
   - sqlx: 数据库访问
   ```

3. **Place in Project Root**:
   ```
   /my-project/
   ├── AGENTFLOW.md           ← Your project config
   ├── Cargo.toml
   └── src/
   ```

## Example AGENTFLOW.md Sections

### Build System
```markdown
## 1. 构建系统

### Frontend
- npm install
- npm run build

### Backend
- cargo build --release
- make dist
```

### Testing Guidelines
```markdown
## 2. 测试工作流

### Unit Tests
- cargo test --lib

### Integration Tests
- docker-compose up -d
- cargo test --test '*'

### What NOT to do
- Don't use `make test` (it's deprecated)
- Don't run integration tests without database
```

### Do's and Don'ts
```markdown
## 3. 关键技能

### ✅ DO (应该做)
- Run `cargo fmt` before committing
- Update CHANGELOG.md for user-facing changes
- Add tests for new features

### ❌ DON'T (不应该做)
- Don't commit to main directly (use PRs)
- Don't ignore clippy warnings
- Don't use `.unwrap()` in production code

### 🔧 Special Skills
- Database migration failed? Delete `migrations/.tmp/`
- Cache issues? Clear `target/` directory
```

### Debugging Strategies
```markdown
## 4. 调试策略

### Common Issues
- **Compilation error**: Check Rust version in `rust-toolchain.toml`
- **Test failures**: Run with `RUST_LOG=debug` for more info
- **Performance**: Use `cargo flamegraph` to profile

### Log Locations
- Application logs: `logs/app.log`
- Error logs: `logs/error.log`
```

### Project-Specific Info
```markdown
## 5. 项目特定信息

### Architecture
- Layered architecture: handlers → services → repositories
- Event-driven: uses internal event bus for async operations

### Key Dependencies
- `axum`: Web framework
- `sqlx`: Database toolkit
- `tokio`: Async runtime

### Common Tasks
- Add new API endpoint: See `docs/api-guide.md`
- Database migration: See `docs/migrations.md`
- Add dependency: Update `Cargo.toml` and run `cargo fetch`
```

## Benefits

1. **Project Context**: AI understands project-specific workflows
2. **Consistency**: Team follows same build/test procedures
3. **Reduced Errors**: AI knows what NOT to do
4. **Faster Execution**: No trial-and-error for project-specific commands
5. **Team Knowledge**: Captures tribal knowledge in executable form

## Verification

All modifications have been verified:

- ✅ `workspace_path` field added to `PromptBuilder`
- ✅ `with_workspace()` method implemented
- ✅ `load_project_config()` reads AGENTFLOW.md
- ✅ `build_project_config_section()` formats config
- ✅ Both `build()` and `build_with_metadata()` integrate project config
- ✅ Example configuration file created at `templates/AGENTFLOW.md.example`

## Testing

To test the feature:

1. Create a test AGENTFLOW.md:
   ```bash
   echo "# Test Config" > AGENTFLOW.md
   ```

2. Run PromptBuilder:
   ```rust
   let builder = PromptBuilder::new();
   let prompt = builder.build("Test task", &[]);

   assert!(prompt.contains("项目专属配置"));
   assert!(prompt.contains("Test Config"));
   ```

3. Delete AGENTFLOW.md and verify no errors:
   ```bash
   rm AGENTFLOW.md
   ```

4. Run PromptBuilder again - should work normally without project config section

## Migration Guide

No migration needed! The feature is backward compatible:

- Existing code works unchanged
- AGENTFLOW.md is optional
- If missing, PromptBuilder behaves as before
- If present, config is automatically included

## Future Enhancements

Possible future improvements:

1. **Multiple config files**: Support for `AGENTFLOW.{project}.md`
2. **Config validation**: Schema validation for AGENTFLOW.md
3. **Inheritance**: Base config + project-specific overrides
4. **Remote configs**: Load from URLs for distributed teams
5. **Config sections**: Selective inclusion of sections
6. **Priority**: Override with command-line configs

## References

- Implementation: `/rust/agentflow-core/src/executor/prompt_builder.rs`
- Example config: `/templates/AGENTFLOW.md.example`
- Related: Memory system, task execution, prompt building
