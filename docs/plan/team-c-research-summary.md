# Team C Research Summary - Cloud/Edge Federation & Webhook

**Date**: 2026-01-28
**Branch**: feature/0.2.1
**Research Focus**: Distributed node architecture with Zhipu Webhook integration

---

## 🔍 Current Codebase Analysis

### Existing Architecture

**File**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/main.rs`

The current system follows a **single-process Master=Worker architecture**:

```rust
// Current startup flow
main() -> init_database() -> TaskExecutor::new() -> MemoryCore::new()
    -> create_app() -> axum::serve()
```

**Key Components**:

1. **Database (SQLite)**: Initialized in `init_database()`
   - Tables: `tasks` (with indexes on status, group_name, parent_id, created_at)
   - No node-related tables yet

2. **Task Executor**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/executor.rs`
   - Executes tasks via `TaskExecutor::execute_task(task_id)`
   - Uses in-memory `running_tasks: Vec<i64>` for concurrency control
   - Updates task status (Running -> Completed/Failed)

3. **Routes**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/mod.rs`
   ```rust
   - GET  /health, /api/v1/health
   - POST /api/v1/tasks
   - GET  /api/v1/tasks
   - GET  /api/v1/tasks/:id
   - DELETE /api/v1/tasks/:id
   - POST /api/v1/tasks/:id/execute
   - POST /api/v1/tasks/:id/cancel
   - GET  /api/v1/memory/search
   - POST /api/v1/memory/search
   - GET  /api/v1/memory/:key
   - DELETE /api/v1/memory/:key
   - GET  /api/v1/memory/stats
   - GET  /ws/task/:id (WebSocket for task updates)
   ```

4. **WebSocket**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/websocket.rs`
   - Current implementation: Task progress streaming
   - Protocol: ping/pong, subscribe
   - Can be extended for node heartbeat

### Type System

**File**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/types.rs`

Existing relevant types:
- `Task`: Core task structure
- `TaskStatus`: Pending, Running, Completed, Failed, Blocked
- `TaskPriority`: Low, Medium, High
- `Worker`, `WorkerStatus`, `WorkerType`: **Legacy from multi-worker architecture**
- `CreateTaskRequest`: Task creation payload

**Note**: `Worker` types exist but are not used in current single-process architecture. We can repurpose or replace with `Node` types.

### Configuration System

**File**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/config.rs`

Current configuration structure:
```rust
pub struct MasterConfig {
    pub server_addr: String,
    pub server_port: u16,
    pub database_url: String,
    pub sandbox: SandboxConfig,
    pub memory: MemoryConfig,
    pub log_level: String,
    pub worker_heartbeat_timeout: u64,  // 60s - NOT USED currently
    pub task_timeout: u64,              // 300s
    pub max_concurrent_tasks: usize,    // 10
}
```

**Loading**: Environment variables via `dotenvy`, with defaults

**Extensions Needed**:
- `NodeConfig`: heartbeat_timeout, cleanup_interval
- `WebhookConfig`: secret, rate_limit, allowed_actions, dangerous_keywords

### Dependencies

**File**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/Cargo.toml`

Already available:
```toml
axum = { version = "0.8", features = ["ws"] }  # WebSocket support ✓
tokio = { version = "1.42", features = ["full"] }  # Async runtime ✓
sqlx = { version = "0.8", features = ["sqlite", "chrono"] }  # Database ✓
serde = { version = "1.0", features = ["derive"] }  # Serialization ✓
chrono = { version = "0.4", features = ["serde"] }  # Time handling ✓
uuid = { version = "1.11", features = ["v4"] }  # UUID generation ✓
```

**Need to add**:
```toml
reqwest = { version = "0.12", features = ["json"] }  # HTTP client (already in workspace)
```

---

## 📐 Architecture Design Insights

### 1. WebSocket Infrastructure Already Exists

**Current Usage**: Task progress streaming (`/ws/task/:id`)

**Reusability for Node Heartbeat**:
- Existing message parsing logic can be extended
- `split()` method already used for bidirectional communication
- Can add new message types: `hello`, `heartbeat`, `node_info`

**Extension Strategy**:
```rust
// Current: /ws/task/:id
// Add: /ws/node (new endpoint for node connections)

// Existing code pattern:
let (mut sender, mut receiver) = socket.split();
while let Some(msg) = receiver.next().await {
    match msg {
        Ok(Message::Text(text)) => { /* parse and handle */ }
        // ...
    }
}
```

### 2. Task Execution Flow

**Current Flow** (Single-node):
```
POST /api/v1/tasks → create_task() → DB
POST /api/v1/tasks/:id/execute → TaskExecutor.execute_task()
  → mark_task_running() → do_execute() → mark_task_completed()
```

**Distributed Flow** (Cloud/Edge):
```
Cloud Side:
POST /api/v1/webhook/zhipu → verify_auth → call_planner()
  → NodeRegistry.determine_target()
  → HTTP POST to {target_node}/api/v1/tasks

Edge Side:
POST /api/v1/tasks → create_task() → DB (local)
  → TaskExecutor.execute_task() → (same as current)
```

**Key Insight**: Edge nodes can use **existing task execution logic** without modification. Only cloud side needs new routing/dispatch logic.

### 3. Database Schema Extensions

**Current Tables** (from `main.rs` migration):
```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL UNIQUE,
    parent_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    group_name TEXT NOT NULL DEFAULT 'default',
    completion_criteria TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 1,
    lock_holder TEXT,
    lock_time TEXT,
    result TEXT,
    error TEXT,
    workspace_dir TEXT,
    sandboxed INTEGER NOT NULL DEFAULT 0,
    allow_network INTEGER NOT NULL DEFAULT 0,
    max_memory TEXT,
    max_cpu INTEGER,
    created_by TEXT,
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

**Extensions Needed**:
```sql
-- New tables
CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    fingerprint TEXT NOT NULL UNIQUE,
    endpoint TEXT NOT NULL,
    last_heartbeat INTEGER NOT NULL,
    labels TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE trusted_nodes (
    fingerprint TEXT PRIMARY KEY,
    added_at INTEGER NOT NULL,
    added_by TEXT
);

-- Extend existing tasks table
ALTER TABLE tasks ADD COLUMN target_node_id TEXT;
ALTER TABLE tasks ADD COLUMN from_node_id TEXT;
ALTER TABLE tasks ADD COLUMN labels TEXT;
```

### 4. State Management Pattern

**Current AppState** (`routes/mod.rs`):
```rust
#[derive(Clone)]
pub struct AppState {
    pub executor: TaskExecutor,
    pub memory: MemoryCore,
    pub start_time: chrono::DateTime<chrono::Utc>,
}
```

**Proposed Extension**:
```rust
#[derive(Clone)]
pub struct AppState {
    pub executor: TaskExecutor,
    pub memory: MemoryCore,
    pub start_time: chrono::DateTime<chrono::Utc>,
    // New fields:
    pub node_registry: Arc<NodeRegistry>,
    pub config: MasterConfig,
    pub rate_limiter: Arc<RateLimiter>,
}
```

**Pattern**: Use `Arc` for shared state to avoid cloning expensive resources.

---

## 🔒 Security Analysis

### Current Security Posture

**Existing Mechanisms**:
- CORS enabled (`tower-http` with `Any` origin - **needs tightening for production**)
- Trace logging for debugging
- No authentication on API endpoints (currently acceptable for local mode)

**Gaps to Address**:

1. **No Authentication Middleware**: Need to implement
   - Bearer token verification for webhook
   - Fingerprint-based auth for node connections

2. **No Rate Limiting**: Need to add
   - In-memory rate limiter (simple, fast)
   - IP-based tracking
   - Configurable limits

3. **No Input Validation**: Need to enhance
   - Dangerous operation detection
   - Action whitelisting

### Recommended Security Layers

```
Layer 1: Network (Infrastructure)
  - HTTPS in production (TLS termination at reverse proxy)
  - Firewall rules to restrict webhook endpoint access

Layer 2: Authentication (Application)
  - Bearer token for webhook
  - Fingerprint whitelist for nodes

Layer 3: Authorization (Business Logic)
  - Action whitelist (allowed_actions config)
  - Dangerous keyword detection

Layer 4: Rate Limiting (DoS Prevention)
  - IP-based rate limiting
  - Per-node rate limiting (future)

Layer 5: Audit (Observability)
  - Structured logging for all auth attempts
  - Failed login alerts
```

---

## 🌐 Webhook Integration Design

### Zhipu AI Webhook Requirements

**Based on task specification**:

Endpoint: `POST /api/v1/webhook/zhipu`

Request:
```json
{
  "text": "在 DiveAdstra 上跑一次 DX12 性能测试"
}
```

Response (Success):
```json
{
  "status": "dispatched",
  "task_id": 123,
  "target_node": "home_win",
  "message": "已下发至 Windows 节点执行"
}
```

Response (Rejection):
```json
{
  "status": "rejected",
  "task_id": null,
  "target_node": null,
  "message": "No suitable node available",
  "reason": "No node matches task requirements"
}
```

### Planner Integration Strategy

**Current Understanding**: Planner is a CLI command that:
1. Takes natural language task description
2. Returns structured JSON with intent, steps, labels

**Implementation Approach**:
```rust
// Use tokio::process::Command to call planner
let output = Command::new("claude")
    .args(["-p", "--model", "claude-3-haiku"])
    .arg(planner_prompt)
    .output()
    .await?;
```

**Planner Prompt Template**:
```
Analyze this task and return JSON:

Task: {user_input}

Return format:
{
  "intent": "benchmark|test|deploy|...",
  "target_repository": "repo-name or null",
  "steps": ["step1", "step2", ...],
  "labels": {
    "os": "windows|linux|macos",
    "repo": "repository-name",
    "requires_gpu": true|false
  },
  "estimated_duration_minutes": 10
}
```

### Node Selection Algorithm

**Label Matching Strategy**:

```rust
// Task labels from planner: {"os": "windows", "repo": "DiveAdstra"}
// Node labels from registration: {"os": "windows", "repo": "DiveAdstra", "gpu": "rtx3080"}

// Score-based matching:
fn score_node(task_labels: &HashMap<String, String>, node: &Node) -> i32 {
    let mut score = 0;
    for (key, task_value) in task_labels {
        if let Some(node_value) = node.labels.get(key) {
            if node_value == task_value {
                score += 10;  // Exact match
            } else {
                score -= 5;   // Mismatch
            }
        } else {
            score += 1;      // Node has more capabilities than required
        }
    }
    score
}

// Select node with highest score
```

---

## 🔄 Node Communication Protocol

### WebSocket Message Types

**Define message enum**:
```rust
#[derive(Debug, Serialize, Deserialize)]
#[serde(tag = "type")]
enum NodeMessage {
    #[serde(rename = "hello")]
    Hello { node_id: String, fingerprint: String, labels: HashMap<String, String> },

    #[serde(rename = "hello_ack")]
    HelloAck { node_id: String, timestamp: i64 },

    #[serde(rename = "heartbeat")]
    Heartbeat { node_id: String, timestamp: i64 },

    #[serde(rename = "ping")]
    Ping { timestamp: i64 },

    #[serde(rename = "pong")]
    Pong { timestamp: i64 },

    #[serde(rename = "task_dispatch")]
    TaskDispatch { task_id: i64, task: CreateTaskRequest },

    #[serde(rename = "error")]
    Error { message: String },
}
```

### Connection State Machine

```
Edge Node                    Cloud Master
   |                               |
   |----- WebSocket Upgrade ------->|
   |                               |
   |<----- Accept ------------------|
   |                               |
   |------ Hello Message --------->|
   |  {node_id, fingerprint}       |
   |                               |
   |         Verify Fingerprint    |
   |         (check trusted_nodes) |
   |                               |
   |<----- Hello Ack --------------|
   |  (node registered)            |
   |                               |
   |<----- Ping -------------------|
   |                               |
   |------ Pong ------------------>|
   |                               |
   |------ Heartbeat (30s) ------->|
   |                               |
   |<----- Pong -------------------|
   |                               |
   ... (heartbeat loop) ...
```

### Timeout Handling

**Configurable timeouts**:
- Heartbeat interval: 30 seconds (Edge → Cloud)
- Heartbeat timeout: 90 seconds (Cloud → mark inactive)
- Reconnection backoff: 5s, 10s, 30s, 60s (max)

**Background cleanup task**:
```rust
tokio::spawn(async move {
    let mut interval = tokio::time::interval(Duration::from_secs(300));
    loop {
        interval.tick().await;
        let inactive_count = node_registry.mark_inactive_nodes().await?;
        info!("Marked {} nodes as inactive", inactive_count);
    }
});
```

---

## 📊 Rate Limiting Strategy

### Implementation Options

**Option 1: In-Memory (Chosen for v0.2.1)**:
- Pros: Simple, fast, no external dependencies
- Cons: Not distributed, resets on restart
- Implementation: `HashMap<Ip, (count, window_start)>`

**Option 2: Redis (Future)**:
- Pros: Distributed, persistent
- Cons: Additional dependency, complexity
- Implementation: `INCR` + `EXPIRE` commands

### In-Memory Implementation

```rust
pub struct RateLimiter {
    requests: Arc<RwLock<HashMap<String, (u64, Instant)>>>,
    max_requests: u64,
    window_duration: Duration,
}

impl RateLimiter {
    pub async fn check(&self, ip: &str) -> Result<(), RateLimitError> {
        let mut requests = self.requests.write().await;
        let now = Instant::now();

        let entry = requests.entry(ip.to_string()).or_insert((0, now));

        // Reset if window expired
        if now.duration_since(entry.1) > self.window_duration {
            *entry = (0, now);
        }

        // Check limit
        if entry.0 >= self.max_requests {
            return Err(RateLimitError::TooManyRequests);
        }

        entry.0 += 1;
        Ok(())
    }

    // Cleanup expired entries (call periodically)
    pub async fn cleanup(&self) {
        let mut requests = self.requests.write().await;
        let now = Instant::now();
        requests.retain(|_, (_, start)| now.duration_since(*start) < self.window_duration);
    }
}
```

---

## 🧪 Testing Approach

### Unit Test Coverage

**NodeRegistry**:
- Register node → verify in database
- Heartbeat → verify timestamp update
- get_node → retrieve correct node
- mark_inactive_nodes → correct nodes marked
- determine_target → label matching logic
- add_trusted_fingerprint → whitelist works
- is_trusted → verification works

**Security Middleware**:
- Valid token → passes
- Invalid token → 401
- Missing token → 401
- Rate limit under threshold → passes
- Rate limit exceeded → 429
- Dangerous keyword detected → 400

**Planner Integration**:
- Mock CLI output → parse JSON correctly
- Extract labels → HashMap populated
- Invalid JSON → error handling

### Integration Test Scenarios

**Scenario 1: Node Registration**
```
1. Start cloud master with empty database
2. Add trusted fingerprint to database
3. Connect edge node via WebSocket
4. Send Hello message with trusted fingerprint
5. Verify Hello Ack received
6. Verify node in database with status=active
7. Stop sending heartbeats for > 90s
8. Verify background task marks node inactive
```

**Scenario 2: Webhook Dispatch**
```
1. Start cloud master with 2 registered nodes
2. Send webhook request with task description
3. Verify auth check passes
4. Verify planner called with correct input
5. Verify task dispatched to correct node
6. Verify webhook response contains task_id and target_node
```

**Scenario 3: Security Rejection**
```
1. Start cloud master
2. Send webhook without Bearer token → 401
3. Send webhook with invalid token → 401
4. Send webhook with dangerous keyword → 400
5. Send 11 requests in 1 minute → 429 on 11th
6. Connect node with untrusted fingerprint → connection rejected
```

### Manual Testing Setup

**Local Multi-Node Testing**:
```bash
# Terminal 1: Cloud Master
export AGENTFLOW_MODE=cloud
export AGENTFLOW_SERVER_PORT=6767
cargo run --bin agentflow-master

# Terminal 2: Edge Node 1
export AGENTFLOW_MODE=edge
export AGENTFLOW_NODE_ID=local1
export AGENTFLOW_SERVER_PORT=6768
cargo run --bin agentflow-master

# Terminal 3: Edge Node 2
export AGENTFLOW_MODE=edge
export AGENTFLOW_NODE_ID=local2
export AGENTFLOW_SERVER_PORT=6769
cargo run --bin agentflow-master

# Terminal 4: Test Webhook
curl -X POST http://localhost:6767/api/v1/webhook/zhipu \
  -H "Authorization: Bearer test-secret" \
  -H "Content-Type: application/json" \
  -d '{"text": "Run benchmark on local1"}'
```

---

## 🚀 Implementation Considerations

### File Structure

**New Files to Create**:
```
rust/agentflow-core/src/
  node/
    mod.rs              # Node, NodeRegistry, related types
    planner.rs          # Planner integration

rust/agentflow-master/src/
  routes/
    webhook.rs          # Webhook endpoint handler
    node_ws.rs          # Node WebSocket endpoint
  middleware/
    mod.rs
    auth.rs             # Bearer token verification
    rate_limit.rs       # Rate limiting
    safety.rs           # Dangerous operation detection
  dispatcher.rs         # Task dispatcher to edge nodes
  config.rs             # Extended with NodeConfig, WebhookConfig

rust/migrations/
  002_add_nodes.sql     # Database schema for nodes

docs/
  ZHIPU_INTEGRATION.md  # User-facing documentation
  NODE_ARCHITECTURE.md  # Architecture documentation
```

**Files to Modify**:
```
rust/agentflow-master/src/
  main.rs               # Add NodeRegistry to AppState, cleanup task
  routes/mod.rs         # Add new routes, update AppState

rust/agentflow-core/src/
  lib.rs                # Export node module
```

### Configuration Migration

**Current `.env.example**:
```bash
AGENTFLOW_SERVER_ADDR=0.0.0.0
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_DATABASE_URL=sqlite://agentflow.db
AGENTFLOW_WORKER_HEARTBEAT_TIMEOUT=60  # Legacy, not used
```

**Additions Needed**:
```bash
# Node configuration
AGENTFLOW_NODE_HEARTBEAT_TIMEOUT=90
AGENTFLOW_NODE_CLEANUP_INTERVAL=300

# Webhook configuration
AGENTFLOW_WEBHOOK_SECRET=change-me-in-production
AGENTFLOW_WEBHOOK_RATE_LIMIT=10
AGENTFLOW_WEBHOOK_ALLOWED_ACTIONS=run_test,benchmark,status
AGENTFLOW_WEBHOOK_DANGEROUS_KEYWORDS=install,system,rm -rf,sudo
```

### Error Handling Strategy

**Error Types**:
```rust
pub enum NodeError {
    NodeNotFound(String),
    InvalidFingerprint(String),
    NodeAlreadyRegistered(String),
    DatabaseError(sqlx::Error),
    HeartbeatTimeout(String),
}

pub enum WebhookError {
    Unauthorized,
    RateLimited,
    DangerousOperation(String),
    PlannerFailed(String),
    NoTargetNode(String),
    DispatchFailed(String),
}
```

**Logging Levels**:
- `ERROR`: Authentication failures, dispatch failures
- `WARN`: Untrusted connection attempts, rate limit hits
- `INFO`: Node registration, task dispatch, heartbeat
- `DEBUG`: Message parsing, label matching details

---

## 📈 Performance Considerations

### Expected Load

**Assumptions**:
- 3-10 edge nodes (small deployment)
- 1-5 webhook requests per minute
- 1 heartbeat per node every 30 seconds

**Resource Usage Estimates**:

| Component | Memory | CPU |
|-----------|--------|-----|
| NodeRegistry | ~1KB per node | Negligible |
| RateLimiter | ~100 bytes per IP | Negligible |
| WebSocket (per node) | ~10KB connection overhead | Minimal |
| Background tasks | Negligible | Minimal (5min interval) |

**Bottlenecks to Watch**:
1. Planner CLI execution time (likely dominant factor)
2. Database write latency (SQLite: sub-millisecond)
3. Network latency to edge nodes (varies)

### Scalability Limits (v0.2.1)

| Metric | Limit | Reason |
|--------|-------|--------|
| Max nodes | ~50 | In-memory state, single master |
| Webhook RPS | ~100 | Planner is blocking, single-threaded |
| Concurrent tasks | Per-node limit | Existing TaskExecutor limit |
| Heartbeat frequency | 30s minimum | Configured interval |

**Future Scaling** (post v0.2.1):
- Redis for distributed rate limiting
- Multiple cloud masters (HA)
- Async planner (HTTP service)
- Connection pooling for edge nodes

---

## ✅ Deliverables Checklist

### Code Deliverables

- [ ] `rust/agentflow-core/src/node/mod.rs` (400-600 lines)
  - Node, NodeStatus, NodeRegistration types
  - NodeRegistry implementation
  - Database integration

- [ ] `rust/agentflow-core/src/node/planner.rs` (100-200 lines)
  - Planner CLI integration
  - Prompt building
  - Response parsing

- [ ] `rust/agentflow-master/src/routes/webhook.rs` (200-300 lines)
  - Webhook handler
  - Auth, rate limit, safety checks
  - Planner call and dispatch logic

- [ ] `rust/agentflow-master/src/routes/node_ws.rs` (250-350 lines)
  - WebSocket upgrade handler
  - Hello/Heartbeat protocol
  - Connection management

- [ ] `rust/agentflow-master/src/middleware/mod.rs` (50 lines)
- [ ] `rust/agentflow-master/src/middleware/auth.rs` (100 lines)
- [ ] `rust/agentflow-master/src/middleware/rate_limit.rs` (150 lines)
- [ ] `rust/agentflow-master/src/middleware/safety.rs` (100 lines)

- [ ] `rust/agentflow-master/src/dispatcher.rs` (100-150 lines)
  - HTTP client for task dispatch
  - Error handling

- [ ] `rust/migrations/002_add_nodes.sql` (50 lines)
  - nodes table
  - trusted_nodes table
  - tasks table extensions

### Configuration

- [ ] Update `rust/agentflow-master/src/config.rs`
  - Add NodeConfig struct
  - Add WebhookConfig struct
  - Update MasterConfig

- [ ] Update `rust/.env.example`
  - Add node configuration variables
  - Add webhook configuration variables

### Main Integration

- [ ] Update `rust/agentflow-master/src/main.rs`
  - Initialize NodeRegistry
  - Initialize RateLimiter
  - Add background cleanup task
  - Update AppState creation

- [ ] Update `rust/agentflow-master/src/routes/mod.rs`
  - Add webhook route
  - Add node WebSocket route
  - Add cluster status route
  - Update AppState

### Tests

- [ ] Unit tests for NodeRegistry (50-70% coverage)
- [ ] Unit tests for middleware (80%+ coverage)
- [ ] Integration test for node registration
- [ ] Integration test for webhook dispatch
- [ ] Manual testing guide

### Documentation

- [ ] `docs/ZHIPU_INTEGRATION.md` (User guide)
- [ ] `docs/NODE_ARCHITECTURE.md` (Architecture doc)
- [ ] Update `README.md` with cloud mode examples
- [ ] Update `CHANGELOG.md`

---

## 🎯 Success Criteria

### Functional Tests (Must Pass)

```bash
# Test 1: Node Registration
./test-node-registration.sh
# ✓ Node connects successfully
# ✓ Fingerprint verified
# ✓ Heartbeat updates timestamp
# ✓ Inactive detection works

# Test 2: Webhook Dispatch
./test-webhook-dispatch.sh
# ✓ Auth required
# ✓ Rate limit enforced
# ✓ Dangerous ops rejected
# ✓ Tasks dispatched to correct node
# ✓ Response format correct

# Test 3: Multi-Node Setup
./test-multi-node.sh
# ✓ 3 nodes register
# ✓ Tasks routed correctly
# ✓ Cluster status accurate
```

### Performance Benchmarks

- [ ] Webhook response (excluding planner) < 100ms
- [ ] Node heartbeat processing < 5ms
- [ ] Can handle 10 concurrent node connections
- [ ] Memory increase < 20MB with 5 nodes

### Security Validation

- [ ] No unauthenticated access to webhook
- [ ] Rate limit cannot be bypassed
- [ ] Dangerous operations always rejected
- [ ] Untrusted nodes rejected
- [ ] All auth attempts logged

---

## 🔄 Timeline Estimation

### Phase 1: Foundation (2-3 days)
- Database schema + migration
- Core types (Node, NodeRegistry)
- Configuration extensions

### Phase 2: WebSocket (2-3 days)
- Node connection endpoint
- Hello/Heartbeat protocol
- Fingerprint verification

### Phase 3: Security (2 days)
- Auth middleware
- Rate limiting
- Safety checks

### Phase 4: Planner & Dispatch (2-3 days)
- Planner integration
- Task dispatcher
- Webhook handler

### Phase 5: Integration & Testing (2 days)
- Main.rs integration
- End-to-end testing
- Documentation

**Total**: 10-13 days (2 weeks)

---

## 📚 References

### Existing Code
- `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/main.rs`
- `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/config.rs`
- `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/mod.rs`
- `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/websocket.rs`
- `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/types.rs`

### Documentation
- `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/v0.2.1_TASK_BREAKDOWN.md`
- `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/v0.2.1迭代计划.md`
- `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/agentFlow云端流程.md`

### External References
- Axum WebSocket docs: https://docs.rs/axum/latest/axum/extract/ws/
- Tower middleware: https://docs.rs/tower/latest/tower/
- SQLx SQLite: https://docs.rs/sqlx/latest/sqlx/sqlite/

---

**Research Completed**: 2026-01-28
**Next Step**: Begin Phase 1 implementation
**Status**: Ready for development
