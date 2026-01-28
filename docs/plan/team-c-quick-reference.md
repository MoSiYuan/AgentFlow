# Team C Quick Reference - Cloud/Edge Federation & Webhook

**Project**: AgentFlow v0.2.1
**Team**: Team C - Cloud/Edge Federation
**Date**: 2026-01-28

---

## 🎯 Mission Summary

Transform AgentFlow from a **single-node system** to a **cloud-edge federation** where:
- Cloud Master receives webhooks from Zhipu AI
- Edge nodes register with cloud and execute tasks locally
- Security is enforced via fingerprint authentication and rate limiting

---

## 📁 File Locations

### Source Code
```
/Users/jiangxiaolong/work/project/AgentFlow/rust/
├── agentflow-core/src/
│   ├── types.rs              # Existing types
│   ├── database.rs           # DB pool
│   └── node/                 # NEW
│       ├── mod.rs            # Node, NodeRegistry
│       └── planner.rs        # Planner integration
│
└── agentflow-master/src/
    ├── main.rs               # Server startup (MODIFY)
    ├── config.rs             # Configuration (EXTEND)
    ├── executor.rs           # Task executor (existing)
    ├── dispatcher.rs         # NEW - Task dispatcher
    └── routes/
        ├── mod.rs            # Routes (EXTEND)
        ├── tasks.rs          # Task API (existing)
        ├── websocket.rs      # Task WS (existing)
        ├── webhook.rs        # NEW - Zhipu webhook
        └── node_ws.rs        # NEW - Node WS
    └── middleware/           # NEW
        ├── auth.rs           # Bearer token auth
        ├── rate_limit.rs     # Rate limiting
        └── safety.rs         # Dangerous op detection
```

### Documentation
```
/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/
├── team-c-cloud-implementation.md  # Full implementation plan
├── team-c-research-summary.md      # Research findings
└── team-c-architecture-diagram.md  # Architecture diagrams
```

### Database
```
/Users/jiangxiaolong/work/project/AgentFlow/rust/migrations/
└── 002_add_nodes.sql     # NEW - Node tables
```

---

## 🔑 Key Concepts

### 1. Node Types

**Cloud Master**:
- Runs in cloud (AWS/VPS)
- Receives webhooks from Zhipu AI
- Maintains node registry
- Dispatches tasks to edge nodes

**Edge Node**:
- Runs on local machines (home/work)
- Registers with cloud master via WebSocket
- Executes tasks locally
- Reports status via heartbeat

### 2. Communication Protocols

**WebSocket** (`/ws/node`):
- Used for: Node registration, heartbeat
- Direction: Bidirectional
- Frequency: Every 30 seconds
- Messages: Hello, HelloAck, Heartbeat, Ping, Pong

**HTTP** (`/api/v1/tasks`):
- Used for: Task dispatch
- Direction: Cloud → Edge
- Method: POST
- Body: CreateTaskRequest with labels

**Webhook** (`/api/v1/webhook/zhipu`):
- Used for: Task intake from Zhipu AI
- Direction: Zhipu → Cloud
- Method: POST
- Auth: Bearer token

### 3. Security Layers

```
1. Authentication (Bearer token for webhook)
2. Rate Limiting (10 req/min per IP)
3. Authorization (Action whitelist)
4. Safety Check (Dangerous keyword detection)
5. Fingerprint Verification (Node connection)
```

---

## 📊 Database Schema

### New Tables

```sql
-- Nodes
CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    fingerprint TEXT UNIQUE,
    endpoint TEXT,
    last_heartbeat INTEGER,
    labels TEXT,              -- JSON
    status TEXT,              -- active|inactive
    created_at INTEGER,
    updated_at INTEGER
);

-- Trusted Nodes
CREATE TABLE trusted_nodes (
    fingerprint TEXT PRIMARY KEY,
    added_at INTEGER,
    added_by TEXT
);

-- Extend Tasks
ALTER TABLE tasks ADD COLUMN target_node_id TEXT;
ALTER TABLE tasks ADD COLUMN from_node_id TEXT;
ALTER TABLE tasks ADD COLUMN labels TEXT;
```

---

## 🔌 API Endpoints

### New Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | /api/v1/webhook/zhipu | Receive webhook from Zhipu AI |
| WS | /ws/node | Node connection and heartbeat |
| GET | /api/v1/cluster/status | Cluster status (nodes list) |
| POST | /api/v1/admin/trusted-nodes | Add trusted fingerprint |

### Existing Endpoints (Used by Edge)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | /api/v1/tasks | Create task (called by cloud dispatcher) |
| GET | /api/v1/tasks/:id | Get task status |
| POST | /api/v1/tasks/:id/execute | Execute task (local execution) |

---

## 🔄 Request Flow: Webhook to Execution

```
1. Zhipu AI → POST /api/v1/webhook/zhipu
   Body: {"text": "Run benchmark on DiveAdstra"}

2. Cloud Master → Verify Bearer token

3. Cloud Master → Check rate limit (IP-based)

4. Cloud Master → Detect dangerous keywords

5. Cloud Master → Call planner CLI
   Returns: {
     "intent": "benchmark",
     "labels": {"os": "windows", "repo": "DiveAdstra"},
     "steps": [...]
   }

6. Cloud Master → NodeRegistry.determine_target(labels)
   Matches labels against nodes
   Returns: Node {id: "home_win", endpoint: "http://..."}

7. Cloud Master → Create task in local DB

8. Cloud Master → Dispatcher.dispatch(node, task)
   POST http://home_win:6767/api/v1/tasks

9. Edge Node → Receive task, store in local DB

10. Edge Node → TaskExecutor.execute_task()
    Execute steps, update status

11. Cloud Master → Return response to Zhipu AI
    {
      "status": "dispatched",
      "task_id": 123,
      "target_node": "home_win",
      "message": "已下发至 Windows 节点执行"
    }
```

---

## 🔐 Configuration

### Environment Variables

```bash
# Node configuration
AGENTFLOW_NODE_HEARTBEAT_TIMEOUT=90
AGENTFLOW_NODE_CLEANUP_INTERVAL=300

# Webhook configuration
AGENTFLOW_WEBHOOK_SECRET=your-secret-here
AGENTFLOW_WEBHOOK_RATE_LIMIT=10
AGENTFLOW_WEBHOOK_ALLOWED_ACTIONS=run_test,benchmark,status
AGENTFLOW_WEBHOOK_DANGEROUS_KEYWORDS=install,system,rm -rf,sudo
```

### Config File Structure

```toml
[node]
heartbeat_timeout = 90
cleanup_interval = 300

[webhook]
secret = "your-secret-here"
rate_limit_per_ip = 10
allowed_actions = ["run_test", "benchmark", "status"]
dangerous_keywords = ["install", "system", "rm -rf", "sudo"]
```

---

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
cargo test

# Run specific module
cargo test node_registry
cargo test webhook_handler
cargo test rate_limiter
```

### Integration Tests

```bash
# Start cloud master
cargo run --bin agentflow-master -- --mode=cloud

# Start edge node (terminal 2)
cargo run --bin agentflow-master -- --mode=edge --node-id=local1

# Test webhook
curl -X POST http://localhost:6767/api/v1/webhook/zhipu \
  -H "Authorization: Bearer test-secret" \
  -H "Content-Type: application/json" \
  -d '{"text": "Run benchmark on local1"}'
```

---

## 📝 Code Templates

### Node Registration (WebSocket)

```rust
// Edge Node Client
use tokio_tungstenite::connect_async;

let url = "ws://cloud-master:6767/ws/node";
let (ws_stream, _) = connect_async(url).await?;
let (mut write, mut read) = ws_stream.split();

// Send Hello
let hello = json!({
    "type": "hello",
    "node_id": "home_win",
    "fingerprint": "abc123...",
    "labels": {"os": "windows", "repo": "DiveAdstra"}
});
write.send(Message::Text(hello.to_string())).await?;

// Wait for HelloAck
while let Some(msg) = read.next().await {
    match msg? {
        Message::Text(text) => {
            let msg: serde_json::Value = serde_json::from_str(&text)?;
            if msg["type"] == "hello_ack" {
                println!("Registered: {}", msg["node_id"]);
                break;
            }
        }
        _ => {}
    }
}

// Start heartbeat loop
loop {
    tokio::time::sleep(Duration::from_secs(30)).await;
    let hb = json!({
        "type": "heartbeat",
        "node_id": "home_win",
        "timestamp": now()
    });
    write.send(Message::Text(hb.to_string())).await?;
}
```

### Webhook Handler

```rust
pub async fn handle_zhipu_webhook(
    State(state): State<AppState>,
    headers: HeaderMap,
    Json(req): Json<ZhipuWebhookRequest>,
) -> Result<Json<ZhipuWebhookResponse>, WebhookError> {
    // 1. Auth
    verify_auth_token(&headers, &state.config.webhook.secret)?;

    // 2. Rate limit
    let ip = extract_ip(&headers);
    check_rate_limit(&state.rate_limiter, &ip).await?;

    // 3. Safety check
    detect_dangerous_ops(&req.text)?;

    // 4. Call planner
    let plan = call_planner(&req.text).await?;

    // 5. Determine target
    let target = state.node_registry
        .determine_target(&plan.labels)
        .await?
        .ok_or(WebhookError::NoTargetNode)?;

    // 6. Create task
    let task_id = create_task(&state.db, &plan, &target.id).await?;

    // 7. Dispatch
    dispatch_task(&target.endpoint, &task_id, &plan).await?;

    // 8. Return
    Ok(Json(ZhipuWebhookResponse {
        status: "dispatched".to_string(),
        task_id: Some(task_id),
        target_node: Some(target.id),
        message: format!("已下发至 {} 节点执行", target.id),
        reason: None,
    }))
}
```

---

## 🚀 Implementation Checklist

### Phase 1: Foundation (Week 1, Days 1-3)

- [ ] Create `rust/migrations/002_add_nodes.sql`
- [ ] Implement `agentflow-core/src/node/mod.rs`
  - [ ] `Node`, `NodeStatus` types
  - [ ] `NodeRegistry` struct
  - [ ] `register()`, `heartbeat()`, `get_node()`, `determine_target()`
- [ ] Extend `agentflow-master/src/config.rs`
  - [ ] `NodeConfig` struct
  - [ ] `WebhookConfig` struct
- [ ] Unit tests for NodeRegistry

### Phase 2: WebSocket (Week 1, Days 4-5)

- [ ] Implement `agentflow-master/src/routes/node_ws.rs`
  - [ ] `/ws/node` endpoint
  - [ ] Hello message handler
  - [ ] Heartbeat loop
  - [ ] Fingerprint verification
- [ ] Add background cleanup task to `main.rs`
- [ ] Integration test for node registration

### Phase 3: Security (Week 2, Days 1-2)

- [ ] Implement `agentflow-master/src/middleware/auth.rs`
- [ ] Implement `agentflow-master/src/middleware/rate_limit.rs`
- [ ] Implement `agentflow-master/src/middleware/safety.rs`
- [ ] Add middleware to webhook route
- [ ] Unit tests for all middleware

### Phase 4: Planner & Dispatch (Week 2, Days 3-4)

- [ ] Implement `agentflow-core/src/node/planner.rs`
- [ ] Implement `agentflow-master/src/dispatcher.rs`
- [ ] Implement `agentflow-master/src/routes/webhook.rs`
- [ ] Add `/api/v1/cluster/status` endpoint
- [ ] End-to-end test

### Phase 5: Integration & Polish (Week 2, Day 5)

- [ ] Update `main.rs` with all new components
- [ ] Update `routes/mod.rs` with new routes
- [ ] Add admin endpoint for trusted nodes
- [ ] Write documentation
- [ ] Manual multi-node testing

---

## 📚 Documentation Tasks

- [ ] `docs/ZHIPU_INTEGRATION.md`
  - Setup guide
  - Configuration examples
  - Troubleshooting

- [ ] `docs/NODE_ARCHITECTURE.md`
  - Architecture overview
  - Protocol details
  - Security model

- [ ] Update `README.md`
  - Cloud mode examples
  - Webhook usage
  - Multi-node setup

---

## 🐛 Common Issues & Solutions

### Issue: Node Connection Refused

**Cause**: Fingerprint not in trusted_nodes table

**Solution**:
```bash
# Add fingerprint to cloud master
curl -X POST http://cloud-master:6767/api/v1/admin/trusted-nodes \
  -H "Authorization: Bearer <admin-secret>" \
  -d '{"fingerprint": "abc123..."}'
```

### Issue: Rate Limit Exceeded

**Cause**: Too many requests from same IP

**Solution**:
- Increase `AGENTFLOW_WEBHOOK_RATE_LIMIT` in config
- Or wait 1 minute for limit to reset

### Issue: No Target Node

**Cause**: Task labels don't match any node labels

**Solution**:
- Check node labels in database
- Update task labels to match available nodes
- Or add more generic nodes

### Issue: Heartbeat Timeout

**Cause**: Network connectivity or node crash

**Solution**:
- Check network connectivity
- Verify node is still running
- Increase `heartbeat_timeout` in config

---

## 📈 Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| Webhook response time | < 500ms | TBD |
| Heartbeat processing | < 10ms | TBD |
| Max concurrent nodes | 50 | TBD |
| Memory overhead | < 50MB | TBD |
| Task dispatch latency | < 100ms | TBD |

---

## 🔗 Quick Links

- **Full Plan**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-c-cloud-implementation.md`
- **Research**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-c-research-summary.md`
- **Diagrams**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/team-c-architecture-diagram.md`
- **Existing Code**: `/Users/jiangxiaolong/work/project/AgentFlow/rust/`
- **Task Breakdown**: `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/v0.2.1_TASK_BREAKDOWN.md`

---

**Last Updated**: 2026-01-28
**Status**: Research Complete - Ready for Implementation
**Next Step**: Begin Phase 1 - Foundation
