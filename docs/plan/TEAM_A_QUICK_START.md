# Team A Implementation Quick Start Guide

**For**: AgentFlow v0.2.1 - CLI & Configuration Layer
**Date**: 2026-01-28
**Status**: Ready to Implement

---

## 🚀 Quick Start

### Step 1: Add Dependencies (5 minutes)

```bash
# Edit rust/Cargo.toml
# Add to [workspace.dependencies]:
toml = "0.8"
dirs = "5.0"
```

```toml
# rust/Cargo.toml

[workspace.dependencies]
# ... existing dependencies ...
toml = "0.8"
dirs = "5.0"
```

**Verify**: `cargo check` in workspace root

---

### Step 2: Create Configuration Module (2 hours)

```bash
# Create directory
mkdir -p rust/agentflow-core/src/config

# Create file
touch rust/agentflow-core/src/config/mod.rs
```

**Copy implementation from**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-a-cli-implementation.md`

**Key sections**:
- Lines 250-450 in the implementation plan
- Config structures
- TOML loading logic
- Validation functions

**Test**:
```bash
cd rust/agentflow-core
cargo test config
```

---

### Step 3: Update Core Library Exports (5 minutes)

```bash
# Edit rust/agentflow-core/src/lib.rs
```

```rust
// Add config module
pub mod config;

// Re-export config types
pub use config::{Config, ServerMode, ServerConfig, ...};
```

**Verify**: `cargo check --package agentflow-core`

---

### Step 4: Create CLI Module (3 hours)

```bash
# Create file
touch rust/agentflow-master/src/cli.rs
```

**Copy implementation from**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-a-cli-implementation.md`

**Key sections**:
- Lines 450-650 in the implementation plan
- CLI enum definitions
- Command handlers

**Test**:
```bash
cd rust/agentflow-master
cargo test cli
```

---

### Step 5: Modify Main Entry Point (1 hour)

```bash
# Edit rust/agentflow-master/src/main.rs
```

**Changes**:
1. Import CLI module: `mod cli;`
2. Replace `Args` parsing with CLI
3. Move server logic into `Cli::execute()`

**Example**:
```rust
// In main.rs

mod cli;  // NEW

#[tokio::main]
async fn main() -> Result<()> {
    // Parse CLI
    let cli = cli::Cli::parse();

    // Execute command
    cli.execute().await?;

    Ok(())
}
```

**Verify**: `cargo run -- --help`

---

### Step 6: Create Mock Node Endpoints (1 hour)

```bash
# Create file
touch rust/agentflow-master/src/routes/nodes.rs
```

**Copy implementation from**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-a-cli-implementation.md`

**Key sections**:
- Lines 650-700 in the implementation plan
- Mock endpoint handlers

**Update routes**:
```rust
// In rust/agentflow-master/src/routes/mod.rs

pub mod nodes; // NEW

pub fn create_routes() -> Router<AppState> {
    Router::new()
        // ... existing routes ...
        // Add new routes
        .route("/api/v1/nodes", get(nodes::list_nodes))
        .route("/api/v1/cluster/status", get(nodes::cluster_status))
        .route("/api/v1/config", get(nodes::get_config))
}
```

**Verify**: `cargo run -- server local`

---

### Step 7: Create Configuration Example (30 minutes)

```bash
# Create file
touch agentflow.toml.example
```

**Copy from**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-a-cli-implementation.md`

**Key sections**:
- Lines 140-220 in the implementation plan
- Complete TOML example

**Test**:
```bash
agentflow config validate agentflow.toml.example
```

---

## 📋 Implementation Checklist

### Phase 1: Foundation (Week 1)

- [ ] Add `toml` and `dirs` dependencies
- [ ] Create `agentflow-core/src/config/mod.rs`
- [ ] Implement Config structures
- [ ] Implement TOML loading
- [ ] Add configuration validation
- [ ] Write config unit tests
- [ ] Update `agentflow-core/src/lib.rs`
- [ ] Test configuration system

### Phase 2: CLI (Week 1)

- [ ] Create `agentflow-master/src/cli.rs`
- [ ] Implement CLI enum
- [ ] Implement all subcommands
- [ ] Add HTTP client logic
- [ ] Modify `main.rs` for CLI
- [ ] Test CLI commands
- [ ] Fix bugs

### Phase 3: Mock Endpoints (Week 2)

- [ ] Create `agentflow-master/src/routes/nodes.rs`
- [ ] Implement list_nodes endpoint
- [ ] Implement cluster_status endpoint
- [ ] Implement get_config endpoint
- [ ] Update `routes/mod.rs`
- [ ] Test endpoints

### Phase 4: Testing & Docs (Week 2)

- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests
- [ ] Create `agentflow.toml.example`
- [ ] Update README.md
- [ ] Write configuration guide
- [ ] Add usage examples
- [ ] Manual testing
- [ ] Performance testing

---

## 🧪 Testing Commands

### Unit Tests

```bash
# Test configuration
cargo test --package agentflow-core config

# Test CLI
cargo test --package agentflow-master cli

# All tests
cargo test
```

### Manual Testing

```bash
# Build project
cargo build --release

# Test server start
./target/release/agentflow-master server local
./target/release/agentflow-master server cloud

# Test config commands
./target/release/agentflow-master config show
./target/release/agentflow-master config validate
./target/release/agentflow-master config init

# Test task commands (server must be running)
./target/release/agentflow-master task list
./target/release/agentflow-master task create "test task"

# Test memory commands
./target/release/agentflow-master memory stats
./target/release/agentflow-master memory search "test"

# Test node commands
./target/release/agentflow-master node list
```

### API Testing

```bash
# Start server
cargo run -- server local

# Test endpoints in another terminal
curl http://localhost:6767/health
curl http://localhost:6767/api/v1/tasks
curl http://localhost:6767/api/v1/nodes
curl http://localhost:6767/api/v1/cluster/status
curl http://localhost:6767/api/v1/config
```

---

## 📁 File Locations Reference

```
AgentFlow/
├── rust/
│   ├── Cargo.toml                          # EDIT: Add toml, dirs
│   ├── agentflow-core/
│   │   ├── Cargo.toml
│   │   └── src/
│   │       ├── lib.rs                      # EDIT: Export config
│   │       └── config/
│   │           └── mod.rs                  # CREATE: Config module
│   └── agentflow-master/
│       ├── Cargo.toml
│       └── src/
│           ├── main.rs                     # EDIT: Use CLI
│           ├── cli.rs                      # CREATE: CLI module
│           └── routes/
│               ├── mod.rs                  # EDIT: Add node routes
│               └── nodes.rs                # CREATE: Mock endpoints
├── agentflow.toml.example                   # CREATE: Config template
├── docs/
│   └── plan/
│       ├── team-a-cli-implementation.md    # REFERENCE: Full plan
│       ├── TEAM_A_CLI_RESEARCH_SUMMARY.md  # REFERENCE: Summary
│       └── CLI_ARCHITECTURE_DIAGRAM.md     # REFERENCE: Diagrams
└── README.md                                # EDIT: Add CLI docs
```

---

## 🔧 Common Issues & Solutions

### Issue 1: Compilation Error - "toml not found"

**Solution**:
```bash
# Check workspace dependencies
cat rust/Cargo.toml | grep -A 20 "\[workspace.dependencies\]"

# Add toml if missing
toml = "0.8"
```

### Issue 2: "Config file not found"

**Solution**:
```bash
# Create default config
agentflow config init

# Or specify path
agentflow --config ./agentflow.toml server local
```

### Issue 3: "Server already running"

**Solution**:
```bash
# Check if port is in use
lsof -i :6767

# Use different port
agentflow server --port 6768 local
```

### Issue 4: "Permission denied" on config file

**Solution**:
```bash
# Create .agentflow directory
mkdir -p ~/.agentflow

# Set permissions
chmod 755 ~/.agentflow
```

---

## 📚 Documentation Tasks

### README Updates

Add section:
```markdown
## CLI Usage

### Starting the Server

```bash
# Local mode (default)
agentflow server local

# Cloud mode
agentflow server cloud

# Planner-only mode
agentflow server planner-only
```

### Task Management

```bash
# Create a task
agentflow task create "My Task" --description "Task description"

# List tasks
agentflow task list

# Execute a task
agentflow task run 1

# Show task details
agentflow task show 1
```

### Configuration

```bash
# Show configuration
agentflow config show

# Validate configuration
agentflow config validate

# Initialize default config
agentflow config init
```

See [docs/plan/team-a-cli-implementation.md](docs/plan/team-a-cli-implementation.md) for complete documentation.
```

### Configuration Guide

Create: `docs/CONFIGURATION.md`

```markdown
# AgentFlow Configuration Guide

## Configuration Locations

AgentFlow searches for configuration files in this order:

1. `--config <PATH>` (CLI override)
2. `./agentflow.toml` (project-specific)
3. `~/.agentflow/config.toml` (user-specific)
4. Built-in defaults

## Server Modes

### Local Mode
Master = Worker, direct Claude CLI execution.

### Cloud Mode
Distributed planner with node registration.

### Planner-Only Mode
Lightweight planning without direct execution.

## Configuration Options

[See agentflow.toml.example for complete reference]
```

---

## 🎯 Daily Goals

### Day 1-2: Configuration System

**Morning**:
- [ ] Add dependencies
- [ ] Create config module structure
- [ ] Implement basic Config types

**Afternoon**:
- [ ] Implement TOML loading
- [ ] Add validation logic
- [ ] Write basic tests

**EOD Goal**: Config loads from file

### Day 3-4: CLI Structure

**Morning**:
- [ ] Create CLI enum
- [ ] Implement server command
- [ ] Add task commands

**Afternoon**:
- [ ] Add memory commands
- [ ] Add config commands
- [ ] Test CLI parsing

**EOD Goal**: All CLI commands parse correctly

### Day 5-7: Integration

**Morning**:
- [ ] Modify main.rs
- [ ] Integrate config with server
- [ ] Test server startup

**Afternoon**:
- [ ] Add HTTP client logic
- [ ] Test CLI → API flow
- [ ] Fix bugs

**EOD Goal**: `agentflow server local` works

### Week 2: Polish & Handoff

**Day 1-2**: Mock endpoints
**Day 3-4**: Testing
**Day 5-7**: Documentation & handoff

---

## ✅ Ready to Implement!

All planning documents are ready:
1. **Implementation Plan**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-a-cli-implementation.md`
2. **Research Summary**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/TEAM_A_CLI_RESEARCH_SUMMARY.md`
3. **Architecture Diagrams**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/CLI_ARCHITECTURE_DIAGRAM.md`
4. **Quick Start**: This document

**Next Steps**:
1. Review implementation plan
2. Start with Step 1: Add dependencies
3. Follow the checklist
4. Ask questions if blocked

**Estimated Timeline**: 2 weeks
**Team Size**: 1-2 developers
**Priority**: 🔴 Highest (blocks other teams)

Good luck! 🚀
