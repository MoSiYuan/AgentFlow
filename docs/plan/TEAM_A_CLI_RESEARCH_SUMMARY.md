# Team A CLI Research Summary - AgentFlow v0.2.1

**Date**: 2026-01-28
**Status**: Research Complete, Ready for Implementation
**Branch**: feature/0.2.1

---

## 🎯 Mission Summary

Team A is responsible for implementing the **unified CLI interface with multi-mode support** for AgentFlow v0.2.1. This is the highest priority work that blocks other teams.

---

## 📊 Current State Analysis

### What Exists ✅

1. **Basic CLI Structure**: `agentflow-master` has clap with simple args
   - Location: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/main.rs`
   - Supports: `--config`, `--addr`, `--port`, `--log-level`
   - Currently: Direct server startup, no subcommands

2. **Configuration System**: Environment-based config
   - Location: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/config.rs`
   - Method: `dotenvy` + environment variables
   - No TOML file support yet

3. **Dependencies**: clap already available
   - clap 4.5 with derive features in workspace
   - No need to add clap, but need `toml` and `dirs` crates

4. **Route Structure**: Well-organized API routes
   - Health, tasks, memory, WebSocket already implemented
   - Easy to add new mock endpoints

### What's Missing ❌

1. **Unified CLI Entry Point**: No `agentflow` command
2. **TOML Configuration**: No file-based configuration
3. **Multi-Mode Support**: No local/cloud/planner-only modes
4. **CLI Subcommands**: No task, memory, node management commands
5. **Mock API Endpoints**: No `/api/v1/nodes`, `/api/v1/cluster/status`, `/api/v1/config`

---

## 🏗️ Architecture Overview

### CLI Command Structure

```
agentflow [GLOBAL_OPTIONS] <COMMAND> [ARGS]

Global Options:
  --config <PATH>    Config file path
  --verbose, -v      Verbose output
  --help, -h         Show help
  --version, -V      Show version

Commands:
  server [mode]      Start server (local|cloud|planner-only)
  task <action>      Task management
    create           Create new task
    run <id>         Execute task
    list             List tasks
    show <id>        Show task details
    cancel <id>      Cancel running task
  memory <action>    Memory operations
    search <query>   Search memory
    stats            Memory statistics
  node <action>      Node management (mock)
    list             List nodes
    status <id>      Node status
  config <action>    Configuration operations
    show             Display config
    validate         Validate config file
    init             Generate default config
  info               System information
```

### Configuration Locations (Priority Order)

1. `--config <PATH>` (CLI override)
2. `./agentflow.toml` (project-specific)
3. `~/.agentflow/config.toml` (user-specific)
4. Built-in defaults

### Server Modes

#### Local Mode (Default)
- Master = Worker
- Direct Claude CLI execution
- No node registration
- Current default behavior

#### Cloud Mode
- Distributed planner
- Node registration enabled
- Task routing to edge nodes
- Webhook support (for Zhipu integration)

#### Planner-Only Mode
- Lightweight planning only
- No direct worker execution
- Suitable for pure forwarding scenarios

---

## 📦 Implementation Plan

### File Structure

```
rust/
├── agentflow-core/
│   └── src/
│       └── config/
│           └── mod.rs          # NEW: TOML-based configuration
├── agentflow-master/
│   └── src/
│       ├── cli.rs              # NEW: Unified CLI
│       ├── main.rs             # MODIFY: Use CLI parser
│       └── routes/
│           ├── mod.rs          # MODIFY: Add node routes
│           └── nodes.rs        # NEW: Mock node endpoints
├── Cargo.toml                  # MODIFY: Add toml, dirs
agentflow.toml.example          # NEW: Config template
```

### Dependencies to Add

```toml
# rust/Cargo.toml (workspace dependencies)
toml = "0.8"
dirs = "5.0"

# rust/agentflow-master/Cargo.toml
reqwest = { workspace = true }  # Already in workspace, need to use
```

### Key Deliverables

1. **Configuration Module** (`agentflow-core/src/config/mod.rs`)
   - TOML loading from multiple locations
   - Multi-mode configuration structures
   - Validation and defaults
   - Environment variable overrides
   - Migration layer for backward compatibility

2. **CLI Module** (`agentflow-master/src/cli.rs`)
   - All subcommands implemented
   - HTTP client for API calls
   - Formatted output (text, JSON, table)
   - Error handling

3. **Mock API Endpoints** (`agentflow-master/src/routes/nodes.rs`)
   - `GET /api/v1/nodes` - List nodes (returns empty mock data)
   - `GET /api/v1/cluster/status` - Cluster status (returns local mode)
   - `GET /api/v1/config` - Current configuration (sanitized)

4. **Configuration Template** (`agentflow.toml.example`)
   - Complete example with all options
   - Inline documentation
   - Default values shown

---

## 🔑 Key Design Decisions

### 1. Why TOML Configuration?

**Pros**:
- More readable than JSON/YAML
- Rust ecosystem has excellent TOML support
- Similar to Cargo.toml (familiar to Rust users)
- Supports comments

**Cons**:
- Not as flexible as YAML
- Slightly more verbose than JSON

**Decision**: TOML is the best fit for a Rust project's configuration.

### 2. Why Separate Config Module?

**Rationale**:
- Keeps configuration logic separate from server
- Allows reuse in future worker binaries
- Cleaner separation of concerns
- Easier to test

**Implementation**: Place in `agentflow-core` so both master and future workers can use it.

### 3. Why Mock Endpoints?

**Purpose**:
- Unblocks Teams B, C, D
- Provides API contract upfront
- Allows CLI testing without full implementation
- Clear placeholder for Team C to implement later

**Strategy**: Return descriptive responses with "not yet implemented" messages.

### 4. Why Keep agentflow-master Binary?

**Backward Compatibility**:
- Existing deployments continue working
- Allows gradual migration
- No breaking changes to v3 architecture
- Dual-mode support during transition

---

## 🔄 Migration Strategy

### Phase 1: Dependencies & Config (Week 1, Days 1-4)

1. Add `toml` and `dirs` to workspace dependencies
2. Create `agentflow-core/src/config/mod.rs`
3. Implement TOML loading and validation
4. Test configuration system
5. Create migration layer for existing `config.rs`

### Phase 2: CLI Implementation (Week 1, Days 5-7)

1. Create `agentflow-master/src/cli.rs`
2. Implement all subcommands
3. Add HTTP client logic
4. Modify `main.rs` to use CLI parser
5. Test CLI commands manually

### Phase 3: Mock Endpoints (Week 2, Days 1-2)

1. Create `agentflow-master/src/routes/nodes.rs`
2. Add routes to `routes/mod.rs`
3. Implement mock endpoints
4. Test API responses

### Phase 4: Testing & Documentation (Week 2, Days 3-7)

1. Write unit tests for config
2. Write integration tests for CLI
3. Create `agentflow.toml.example`
4. Update README with CLI usage
5. Write configuration guide

---

## 📋 Configuration Structure

### Minimal Config

```toml
[server]
mode = "local"
port = 6767
```

### Complete Config

```toml
[server]
mode = "local"              # local | cloud | planner-only
host = "0.0.0.0"
port = 6767
log_level = "info"

[database]
url = "sqlite://agentflow.db"
max_connections = 10

[cli.worker_safe]
command = "claude"
args = ["-p", "--dangerously-skip-permissions"]

[cli.planner]
command = "claude"
args = ["-p", "--model", "claude-3-haiku"]

[sandbox]
enabled = true
default_workspace = "/tmp/agentflow/workspace"
whitelist_dirs = []
enable_symlink_protection = true
max_task_seconds = 1800
allow_network = false
max_memory = "1G"
max_cpu = 2

[memory]
backend = "memory"          # memory | sqlite | redis
default_ttl = 3600
max_entries = 10000
enable_persistence = false

[executor]
max_concurrent_tasks = 10
task_timeout = 300
worker_heartbeat_timeout = 60

[webhook]
secret = "change-me-in-production"
rate_limit_per_ip = 10
allowed_actions = ["run_test", "benchmark", "status"]

[cloud]
enable_registration = true
heartbeat_interval = 30
node_timeout = 120
```

---

## 🧪 Testing Strategy

### Unit Tests

- **Configuration Loading**: Test loading from various locations
- **Configuration Validation**: Test invalid configs are rejected
- **CLI Parsing**: Test all subcommands parse correctly
- **Server Mode Parsing**: Test mode string conversion

### Integration Tests

- **Server Startup**: Test all three modes start correctly
- **API Interaction**: Test CLI commands work with server
- **Configuration Overrides**: Test CLI args override config
- **Error Handling**: Test proper error messages

### Manual Testing

- **Smoke Tests**: Basic commands work
- **Edge Cases**: Missing config, invalid config, server not running
- **Performance**: CLI response time <100ms
- **Compatibility**: Existing agentflow-master still works

---

## ⚠️ Risks & Mitigation

### Risk 1: Breaking Changes

**Impact**: High - Could break existing deployments
**Probability**: Medium

**Mitigation**:
- Keep `agentflow-master` binary working
- Provide migration guide
- Support both interfaces during transition
- Extensive testing

### Risk 2: Configuration Complexity

**Impact**: Medium - Users might struggle with config
**Probability**: Medium

**Mitigation**:
- Provide sensible defaults
- Create `agentflow config init` command
- Document all options
- Provide examples

### Risk 3: Dependency Conflicts

**Impact**: Low - Rust makes this easy
**Probability**: Low

**Mitigation**:
- Use workspace dependencies
- Pin versions
- Test on multiple platforms

### Risk 4: Mock Endpoint Confusion

**Impact**: Low - Temporary issue
**Probability**: Medium

**Mitigation**:
- Clear documentation
- Descriptive error messages
- "Not yet implemented" notes
- Team handoff plan

---

## 📅 Timeline

### Week 1: Foundation

**Days 1-2**: Configuration System
- [ ] Add dependencies
- [ ] Create config module
- [ ] Implement TOML loading
- [ ] Test configuration

**Days 3-4**: CLI Structure
- [ ] Create CLI module
- [ ] Implement subcommands
- [ ] Add HTTP client
- [ ] Test CLI parsing

**Days 5-7**: Integration
- [ ] Modify main.rs
- [ ] Test server startup
- [ ] Test all CLI commands
- [ ] Fix bugs

### Week 2: Polish

**Days 1-2**: Mock Endpoints
- [ ] Create nodes.rs
- [ ] Add routes
- [ ] Test API

**Days 3-4**: Testing
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual testing
- [ ] Performance testing

**Days 5-7**: Documentation
- [ ] Create agentflow.toml.example
- [ ] Update README
- [ ] Write usage guide
- [ ] Final review

---

## 🎯 Success Criteria

### Functional Requirements

- ✅ All CLI commands work correctly
- ✅ Configuration loads from multiple locations
- ✅ Server starts in all three modes
- ✅ API endpoints return correct responses
- ✅ Error messages are clear and helpful

### Non-Functional Requirements

- ✅ Zero breaking changes
- ✅ All tests passing (>80% coverage)
- ✅ Clean compilation (no warnings)
- ✅ Complete documentation
- ✅ CLI response <100ms

### User Experience

- ✅ Intuitive command structure
- ✅ Clear error messages
- ✅ Helpful help text
- ✅ Sensible defaults
- ✅ Easy configuration

---

## 🤝 Team Coordination

### Dependencies on Other Teams

**None** - Team A is the foundation, other teams depend on us.

### Other Teams Depend On Us

**Team B** (Memory & Git):
- Needs `memory` section in config
- Timeline: Week 2, after config is stable

**Team C** (Cloud/Edge):
- Needs mock endpoints as contract
- Timeline: Week 2, will replace mocks

**Team D** (Packaging):
- Needs `agentflow.toml.example` for installer
- Timeline: Week 2, needs final template

---

## 📝 Next Steps

### Immediate Actions (Today)

1. **Review Plan**: Team reviews this implementation plan
2. **Approve Design**: Get consensus on architecture
3. **Start Implementation**: Begin with Phase 1

### This Week

1. Add `toml` and `dirs` dependencies
2. Create `config/mod.rs` with basic structures
3. Implement TOML loading logic
4. Test configuration system

### Next Week

1. Implement CLI module
2. Modify main.rs
3. Create mock endpoints
4. Begin testing

---

## 📚 Reference Documents

### Internal Documentation

- **Implementation Plan**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-a-cli-implementation.md`
- **Overall Plan**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/v0.2.1迭代计划.md`
- **Task Breakdown**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/v0.2.1_TASK_BREAKDOWN.md`
- **Previous Work**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/TEAM_A_IMPLEMENTATION_REPORT.md`

### Code Locations

- **Main Entry**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/main.rs`
- **Current Config**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/config.rs`
- **Routes**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/mod.rs`
- **Types**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/types.rs`

### External Resources

- [Clap Documentation](https://docs.rs/clap/latest/clap/)
- [TOML crate](https://docs.rs/toml/latest/toml/)
- [Dirs crate](https://docs.rs/dirs/latest/dirs/)

---

## ✅ Status Summary

| Component | Status | Notes |
|-----------|--------|-------|
| Research | ✅ Complete | Codebase analyzed |
| Planning | ✅ Complete | Implementation plan created |
| Dependencies | ⏳ Pending | Ready to add toml, dirs |
| Config Module | ⏳ Pending | Design complete |
| CLI Module | ⏳ Pending | Design complete |
| Mock Endpoints | ⏳ Pending | Design complete |
| Documentation | ⏳ Pending | Outline complete |
| Testing | ⏳ Pending | Strategy defined |

**Overall Progress**: 20% (Research & Planning Complete)

---

**End of Research Summary**

Ready to begin implementation. Awaiting team review and approval.
