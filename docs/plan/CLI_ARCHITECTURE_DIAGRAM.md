# AgentFlow v0.2.1 CLI Architecture Diagrams

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         AgentFlow CLI                           │
│                      (agentflow command)                        │
└─────────────────────────────────────────────────────────────────┘
                                  │
                  ┌───────────────┼───────────────┐
                  │               │               │
                  ▼               ▼               ▼
         ┌────────────┐   ┌────────────┐   ┌────────────┐
         │   SERVER   │   │    TASK    │   │   MEMORY   │
         │   MODES    │   │ MANAGEMENT │   │  OPERATIONS│
         └────────────┘   └────────────┘   └────────────┘
              │                  │                │
         ┌────┴────┐         ┌───┴────┐       ┌───┴────┐
         │         │         │        │       │        │
         ▼         ▼         ▼        ▼       ▼        ▼
     ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
     │LOCAL │ │CLOUD │ │CREATE│ │ RUN  │ │SEARCH│ │STATS │
     └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘
         │         │         │        │       │        │
         └────┬────┘         └────────┴───────┴────────┘
              │
              ▼
    ┌─────────────────────┐
    │   HTTP API Client   │
    └─────────────────────┘
              │
              ▼
    ┌─────────────────────┐
    │  AgentFlow Server   │
    │  (agentflow-master) │
    └─────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
    ▼         ▼         ▼
┌────────┐ ┌──────┐ ┌──────┐
│ TASKS  │ │NODES │ │MEMORY│
│  API   │ │ API  │ │  API │
└────────┘ └──────┘ └──────┘
```

## Configuration Loading Flow

```
┌────────────────────────────────────────────────────┐
│              Configuration Sources                 │
│  (Priority: High → Low)                           │
└────────────────────────────────────────────────────┘
           │
    ┌──────┴──────┐
    │             │
    ▼             ▼
┌─────────┐  ┌─────────────┐  ┌────────────────────┐
│--config │  │./agentflow. │  │~/.agentflow/config.│
│ <PATH>  │  │   toml      │  │      toml          │
└─────────┘  └─────────────┘  └────────────────────┘
    │             │                      │
    └─────────────┴──────────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │  Config Loaded │
         └────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
    ▼                           ▼
┌─────────────┐         ┌──────────────┐
│ Apply CLI   │         │ Apply ENV    │
│ Overrides   │         │ Overrides    │
└─────────────┘         └──────────────┘
    │                           │
    └─────────────┬─────────────┘
                  │
                  ▼
         ┌────────────────┐
         │  Final Config  │
         └────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │  Validate      │
         └────────────────┘
                  │
         ┌────────┴────────┐
         │                 │
         ▼                 ▼
    ┌────────┐       ┌──────────┐
    │  Valid │       │ Invalid  │
    └────────┘       └──────────┘
         │                 │
         ▼                 ▼
    ┌────────┐       ┌──────────┐
    │  Use   │       │  Error   │
    │ Config │       │  Message │
    └────────┘       └──────────┘
```

## Server Mode Decision Tree

```
┌────────────────────────────────────────┐
│     User runs: agentflow server       │
└────────────────────────────────────────┘
                  │
                  ▼
    ┌──────────────────────────────┐
    │ What mode is specified?      │
    └──────────────────────────────┘
           │           │           │
           ▼           ▼           ▼
      ┌───────┐  ┌─────────┐  ┌──────────┐
      │ local │  │  cloud  │  │planner-  │
      │       │  │         │  │  only    │
      └───────┘  └─────────┘  └──────────┘
           │           │           │
           ▼           ▼           ▼
    ┌─────────┐  ┌──────────┐  ┌──────────┐
    │Master = │  │Enable    │  │Planner   │
    │Worker   │  │Node Reg  │  │Only Mode │
    │         │  │Webhook   │  │No Worker │
    │Direct   │  │Task Route│  │          │
    │Claude   │  │          │  │          │
    └─────────┘  └──────────┘  └──────────┘
           │           │           │
           └───────────┴───────────┘
                       │
                       ▼
            ┌────────────────────┐
            │  Start HTTP Server │
            │  on configured port│
            └────────────────────┘
```

## CLI Command Hierarchy

```
agentflow
│
├─ server [mode]
│  ├─ --mode <local|cloud|planner-only>
│  ├─ --port <PORT>
│  ├─ --addr <ADDR>
│  └─ --config <PATH>
│
├─ task
│  ├─ create <title>
│  │  ├─ --description <TEXT>
│  │  ├─ --group <NAME>
│  │  ├─ --workspace <PATH>
│  │  ├─ --sandboxed
│  │  └─ --allow-network
│  │
│  ├─ run <id>
│  │  ├─ --wait <true|false>
│  │  └─ --stream
│  │
│  ├─ list
│  │  ├─ --status <STATUS>
│  │  ├─ --group <NAME>
│  │  ├─ --limit <N>
│  │  └─ --format <text|json|table>
│  │
│  ├─ show <id>
│  │
│  └─ cancel <id>
│
├─ memory
│  ├─ search <query>
│  │  ├─ --limit <N>
│  │  └─ --format <text|json>
│  │
│  └─ stats
│     └─ --format <text|json>
│
├─ node
│  ├─ list
│  │  └─ --format <text|json|table>
│  │
│  └─ status <id>
│     └─ --format <text|json>
│
├─ config
│  ├─ show
│  │  ├─ --reveal
│  │  └─ --format <text|json|toml>
│  │
│  ├─ validate [file]
│  │
│  └─ init
│     ├─ --output <PATH>
│     └─ --force
│
└─ info
```

## API Endpoint Structure

```
/api/v1/
├─ tasks
│  ├─ GET    /api/v1/tasks           → List tasks
│  ├─ POST   /api/v1/tasks           → Create task
│  ├─ GET    /api/v1/tasks/:id       → Get task
│  ├─ DELETE /api/v1/tasks/:id       → Delete task
│  ├─ POST   /api/v1/tasks/:id/execute → Execute task (SSE)
│  └─ POST   /api/v1/tasks/:id/cancel  → Cancel task
│
├─ memory
│  ├─ GET    /api/v1/memory/search   → Search memory
│  ├─ POST   /api/v1/memory/search   → Search memory
│  ├─ GET    /api/v1/memory/:key     → Get memory entry
│  ├─ DELETE /api/v1/memory/:key     → Delete memory entry
│  └─ GET    /api/v1/memory/stats    → Memory statistics
│
├─ nodes (NEW - Mock)
│  ├─ GET    /api/v1/nodes           → List nodes
│  └─ GET    /api/v1/cluster/status  → Cluster status
│
├─ config (NEW)
│  └─ GET    /api/v1/config          → Current config
│
├─ webhook (FUTURE - Team C)
│  └─ POST   /api/v1/webhook/zhipu   → Zhipu webhook
│
└─ health
   ├─ GET    /health                → Health check
   └─ GET    /api/v1/health          → Health check

WebSocket:
└─ /ws/task/:id                       → Task execution stream
```

## Module Dependency Graph

```
┌─────────────────────────────────────────────────┐
│                  Cargo.toml                     │
│              (Workspace Dependencies)            │
└─────────────────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
┌───────────────┐       ┌──────────────┐
│agentflow-core│       │agentflow-    │
│               │       │master        │
├───────────────┤       ├──────────────┤
│types          │       │main.rs       │◄─────────┐
│database       │       │cli.rs (NEW)  │          │
│memory/        │       │config.rs     │          │
│executor/      │       │executor.rs   │          │
│sandbox/       │       │routes/       │          │
│config/ (NEW)  │       │  - tasks     │          │
│               │       │  - memory    │          │
└───────────────┘       │  - nodes (NEW)│         │
        │               │  - health    │         │
        │               │  - websocket │         │
        └───────────────│error.rs      │         │
                        │memory_core.rs│         │
                        └──────────────┘         │
                                                  │
                                ┌─────────────────┘
                                │
                                ▼
                         User executes:
                         agentflow server local
                         agentflow task create "test"
                         etc.
```

## Configuration Schema

```
Config
├─ server: ServerConfig
│  ├─ mode: ServerMode (enum)
│  │  ├─ Local
│  │  ├─ Cloud
│  │  └─ PlannerOnly
│  ├─ host: String
│  ├─ port: u16
│  └─ log_level: String
│
├─ database: DatabaseConfig
│  ├─ url: String
│  └─ max_connections: u32
│
├─ cli: CliConfig
│  ├─ worker_safe: CommandConfig
│  │  ├─ command: String
│  │  ├─ args: Vec<String>
│  │  └─ working_dir: Option<PathBuf>
│  │
│  └─ planner: CommandConfig
│     ├─ command: String
│     ├─ args: Vec<String>
│     └─ working_dir: Option<PathBuf>
│
├─ sandbox: SandboxConfig
│  ├─ enabled: bool
│  ├─ default_workspace: PathBuf
│  ├─ whitelist_dirs: Vec<PathBuf>
│  ├─ enable_symlink_protection: bool
│  ├─ max_task_seconds: u64
│  ├─ allow_network: bool
│  ├─ max_memory: Option<String>
│  └─ max_cpu: Option<i32>
│
├─ memory: MemoryConfig
│  ├─ backend: String
│  ├─ default_ttl: u64
│  ├─ max_entries: usize
│  ├─ enable_persistence: bool
│  ├─ database_url: Option<String>
│  └─ redis_url: Option<String>
│
├─ executor: ExecutorConfig
│  ├─ max_concurrent_tasks: usize
│  ├─ task_timeout: u64
│  └─ worker_heartbeat_timeout: u64
│
├─ webhook: WebhookConfig
│  ├─ secret: Option<String>
│  ├─ rate_limit_per_ip: usize
│  └─ allowed_actions: Vec<String>
│
└─ cloud: CloudConfig
   ├─ discovery_endpoint: Option<String>
   ├─ enable_registration: bool
   ├─ heartbeat_interval: u64
   └─ node_timeout: u64
```

## Implementation Flow

```
┌─────────────────────────────────────────────────────┐
│           PHASE 1: Dependencies & Config            │
│                  (Week 1, Days 1-4)                │
└─────────────────────────────────────────────────────┘
  │
  ├─ Add toml, dirs to workspace dependencies
  ├─ Create agentflow-core/src/config/mod.rs
  ├─ Implement Config structures
  ├─ Implement TOML loading
  ├─ Add validation logic
  ├─ Write unit tests
  └─ ✓ Configuration system complete

┌─────────────────────────────────────────────────────┐
│            PHASE 2: CLI Implementation              │
│                  (Week 1, Days 5-7)                │
└─────────────────────────────────────────────────────┘
  │
  ├─ Create agentflow-master/src/cli.rs
  ├─ Implement CLI enum and subcommands
  ├─ Add HTTP client logic
  ├─ Implement command handlers
  ├─ Modify main.rs to use CLI
  ├─ Test CLI commands
  └─ ✓ CLI complete

┌─────────────────────────────────────────────────────┐
│           PHASE 3: Mock API Endpoints               │
│                  (Week 2, Days 1-2)                │
└─────────────────────────────────────────────────────┘
  │
  ├─ Create agentflow-master/src/routes/nodes.rs
  ├─ Implement GET /api/v1/nodes (mock)
  ├─ Implement GET /api/v1/cluster/status (mock)
  ├─ Implement GET /api/v1/config
  ├─ Add routes to routes/mod.rs
  ├─ Test API endpoints
  └─ ✓ Mock endpoints complete

┌─────────────────────────────────────────────────────┐
│         PHASE 4: Testing & Documentation            │
│                  (Week 2, Days 3-7)                │
└─────────────────────────────────────────────────────┘
  │
  ├─ Write unit tests (target >80% coverage)
  ├─ Write integration tests
  ├─ Create agentflow.toml.example
  ├─ Update README.md
  ├─ Write configuration guide
  ├─ Add usage examples
  ├─ Manual testing
  └─ ✓ All deliverables complete

┌─────────────────────────────────────────────────────┐
│                  HANDOFF TO TEAMS                   │
└─────────────────────────────────────────────────────┘
  │
  ├─ Team B (Memory): Config system ready
  ├─ Team C (Cloud): Mock API contract defined
  └─ Team D (Packaging): Config template ready
```

## Testing Pyramid

```
                    ┌─────────────┐
                    │   Manual    │
                    │   Testing   │
                    │  (E2E, UX)  │
                    └──────┬──────┘
                           │
                ┌──────────┴──────────┐
                │                     │
         ┌──────┴──────┐       ┌──────┴──────┐
         │Integration  │       │  Integration│
         │  Tests      │       │   (API)     │
         │(CLI, Config)│       └─────────────┘
         └──────┬──────┘
                │
         ┌──────┴──────┐
         │             │
    ┌────┴────┐  ┌────┴────┐
    │  Unit   │  │  Unit   │
    │  Tests  │  │  Tests  │
    │(Config) │  │  (CLI)  │
    └─────────┘  └─────────┘
```

## Success Metrics

```
┌────────────────────────────────────────────────────┐
│              Functional Requirements               │
└────────────────────────────────────────────────────┘
✅ agentflow server local works
✅ agentflow server cloud works (with mocks)
✅ agentflow task create/list/run work
✅ agentflow memory search/stats work
✅ agentflow config init/validate/show work
✅ Configuration loads from multiple locations
✅ API endpoints return correct responses

┌────────────────────────────────────────────────────┐
│            Non-Functional Requirements             │
└────────────────────────────────────────────────────┘
✅ Zero breaking changes
✅ All tests passing (>80% coverage)
✅ Clean compilation (no warnings)
✅ Complete documentation
✅ CLI response <100ms
✅ Sensible defaults
✅ Clear error messages

┌────────────────────────────────────────────────────┐
│                 Team Dependencies                  │
└────────────────────────────────────────────────────┘
✅ Team A (CLI): Foundation complete
✅ → Team B (Memory): Can use config system
✅ → Team C (Cloud): Has API contract
✅ → Team D (Packaging): Has config template
```
