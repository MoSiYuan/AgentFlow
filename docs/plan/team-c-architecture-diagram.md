# Team C Architecture Diagrams

**Project**: AgentFlow v0.2.1 - Cloud/Edge Federation & Webhook
**Date**: 2026-01-28

---

## 1. High-Level System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         External World                          │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  │ Webhook
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Zhipu AI (Mobile App)                      │
│                    "Run benchmark on home_win"                  │
└────────────────────────────┬────────────────────────────────────┘
                             │ POST /api/v1/webhook/zhipu
                             │ Authorization: Bearer <secret>
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Cloud Master (AWS/VPS)                       │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Security Layer                                          │  │
│  │  ├─ Bearer Token Verification                            │  │
│  │  ├─ Rate Limiting (10 req/min per IP)                    │  │
│  │  └─ Dangerous Operation Detection                        │  │
│  └──────────────────────┬───────────────────────────────────┘  │
│                         ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Planner CLI (claude -p --model claude-3-haiku)          │  │
│  │  Input: "Run benchmark on home_win"                      │  │
│  │  Output: {                                               │  │
│  │    intent: "benchmark",                                  │  │
│  │    labels: {"os": "windows", "repo": "DiveAdstra"},      │  │
│  │    steps: ["git pull", "cargo test", ...]               │  │
│  │  }                                                       │  │
│  └──────────────────────┬───────────────────────────────────┘  │
│                         ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  NodeRegistry (SQLite)                                   │  │
│  │  ┌─────────────────────────────────────────────────┐    │  │
│  │  │ nodes table                                      │    │  │
│  │  │ ┌──────────┬─────────────┬─────────┬──────────┐  │    │  │
│  │  │ │ id       │ fingerprint │ endpoint│ status   │  │    │  │
│  │  │ ├──────────┼─────────────┼─────────┼──────────┤  │    │  │
│  │  │ │home_win  │ abc123...   │ http:// │ active   │  │    │  │
│  │  │ │work_mac  │ def456...   │ http:// │ active   │  │    │  │
│  │  │ │old_srv   │ ghi789...   │ http:// │ inactive │  │    │  │
│  │  │ └──────────┴─────────────┴─────────┴──────────┘  │    │  │
│  │  └─────────────────────────────────────────────────┘    │  │
│  │                                                          │  │
│  │  determine_target(task_labels):                         │  │
│  │    Match task.labels against node.labels                │  │
│  │    Score: exact_match=+10, partial=-5                   │  │
│  │    Return: highest_score_node                           │  │
│  └──────────────────────┬───────────────────────────────────┘  │
│                         ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Task Dispatcher (reqwest::Client)                       │  │
│  │  POST http://home_win:6767/api/v1/tasks                  │  │
│  │  Body: {                                                 │  │
│  │    "title": "Benchmark DiveAdstra",                      │  │
│  │    "from_node_id": "cloud",                              │  │
│  │    "labels": {"os": "windows", "repo": "DiveAdstra"}     │  │
│  │  }                                                       │  │
│  └──────────────────────┬───────────────────────────────────┘  │
└──────────────────────────┼───────────────────────────────────────┘
                           │
                           │ HTTP (Task Dispatch)
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Edge Nodes (Home/Office)                     │
│                                                                  │
│  ┌─────────────────────┐  ┌─────────────────────┐              │
│  │  home_win (Windows) │  │  work_mac (macOS)   │              │
│  │                     │  │                     │              │
│  │  ┌───────────────┐  │  │  ┌───────────────┐  │              │
│  │  │ HTTP Server   │  │  │  │ HTTP Server   │  │              │
│  │  │ :6767         │  │  │  │ :6767         │  │              │
│  │  └───────┬───────┘  │  │  └───────┬───────┘  │              │
│  │          │           │  │          │           │              │
│  │          ▼           │  │          ▼           │              │
│  │  ┌───────────────┐  │  │  ┌───────────────┐  │              │
│  │  │TaskExecutor   │  │  │  │TaskExecutor   │  │              │
│  │  │- Execute task │  │  │  │- Execute task │  │              │
│  │  │- Update DB    │  │  │  │- Update DB    │  │              │
│  │  └───────┬───────┘  │  │  └───────┬───────┘  │              │
│  │          │           │  │          │           │              │
│  │          ▼           │  │          ▼           │              │
│  │  ┌───────────────┐  │  │  ┌───────────────┐  │              │
│  │  │Local SQLite   │  │  │  │Local SQLite   │  │              │
│  │  │tasks table    │  │  │  │tasks table    │  │              │
│  │  └───────────────┘  │  │  └───────────────┘  │              │
│  │                     │  │                     │              │
│  │  ┌───────────────┐  │  │  ┌───────────────┐  │              │
│  │  │WebSocket      │  │  │  │WebSocket      │  │              │
│  │  │Client         │  │  │  │Client         │  │              │
│  │  └───────┬───────┘  │  │  └───────┬───────┘  │              │
│  └──────────┼───────────┘  └──────────┼───────────┘              │
│             │                        │                          │
└─────────────┼────────────────────────┼──────────────────────────┘
              │                        │
              │ WebSocket (Hello + Heartbeat)
              │                        │
              └────────────────────────┘
                       │
                       ▼
            ┌──────────────────────┐
            │ Cloud Master         │
            │ WebSocket Server     │
            │ /ws/node             │
            └──────────────────────┘
```

---

## 2. Node Registration Flow

```
Edge Node Startup                    Cloud Master
     │                                    │
     │ 1. Generate Fingerprint           │
     │    $ agentflow node fingerprint   │
     │    Output: abc123def456...        │
     │                                    │
     │ 2. Start WebSocket Client         │
     │    $ agentflow server --mode=edge │
     │                                    │
     │ 3. Connect                         │
     │ ──── WebSocket Upgrade ───────────>│
     │                                    │
     │                                    │ Check connection
     │                                    │
     │ 4. Send Hello Message              │
     │ ──── {                             │
     │       "type": "hello",             │
     │       "node_id": "home_win",       │
     │       "fingerprint": "abc123...",  │
     │       "labels": {                  │
     │         "os": "windows",           │
     │         "repo": "DiveAdstra"       │
     │       }                            │
     │     } ─────────────────────────────>│
     │                                    │
     │                                    │ 5. Verify Fingerprint
     │                                    │    SELECT * FROM trusted_nodes
     │                                    │    WHERE fingerprint = 'abc123...'
     │                                    │
     │        ┌───────────────────────────┴────────┐
     │        │                                    │
     │   Trusted?                              Untrusted?
     │        │                                    │
     │        ▼                                    ▼
     │  6. Register Node                    6E. Reject Connection
     │     INSERT INTO nodes (...)          ──── {                │
     │     VALUES ('home_win', ...)              "type": "error", │
     │     UPDATE status = 'active'             "message":        │
     │                                          "Fingerprint     │
     │                                          "not in trusted  │
     │                                          "list"           │
     │                                        } ───────────────>│
     │        │                                    │
     │  7. Send Hello Ack                    ── Close Connection │
     │     ──── {                             │
     │            "type": "hello_ack",        │
     │            "node_id": "home_win",      │
     │            "timestamp": 1706456400     │
     │          } ──────────────────────────> │
     │        │                                    │
     │  8. Start Heartbeat Loop                 │
     │     (every 30 seconds)                    │
     │        │                                    │
     │     ──── Heartbeat ─────────────────────>│
     │           {                               │
     │             "type": "heartbeat",          │
     │             "node_id": "home_win",        │
     │             "timestamp": 1706456430       │
     │           }                               │
     │        │                                    │
     │        │ 9. Update Heartbeat              │
     │        │    UPDATE nodes                  │
     │        │    SET last_heartbeat = ...      │
     │        │    WHERE id = 'home_win'         │
     │        │                                    │
     │     <──── Pong ───────────────────────────│
     │           {                               │
     │             "type": "pong",               │
     │             "timestamp": 1706456430       │
     │           }                               │
     │        │                                    │
     │     ... (repeat every 30s) ...             │
     │                                            │
     │  Background Task (every 5min):             │
     │  UPDATE nodes                              │
     │  SET status = 'inactive'                   │
     │  WHERE last_heartbeat < (now() - 90s)      │
```

---

## 3. Webhook Processing Flow

```
Zhipu AI App                     Cloud Master                    Edge Node
     │                                │                              │
     │ 1. User sends message           │                              │
     │    "Run DX12 benchmark"         │                              │
     │                                │                              │
     │ 2. HTTP POST                    │                              │
     │ ──────────────────────────────>│                              │
     │   POST /api/v1/webhook/zhipu   │                              │
     │   Authorization: Bearer <secret>│                              │
     │   {"text": "Run DX12..."}      │                              │
     │                                │                              │
     │                        ┌───────┴────────┐                       │
     │                        │ 3. Verify Auth│                       │
     │                        │   Check token │                       │
     │                        └───────┬────────┘                       │
     │                                │                              │
     │                        ┌───────┴────────┐                       │
     │                        │ 4. Rate Limit │                       │
     │                        │   Check IP    │                       │
     │                        └───────┬────────┘                       │
     │                                │                              │
     │                        ┌───────┴────────┐                       │
     │                        │ 5. Safety Check│                       │
     │                        │   Keywords     │                       │
     │                        └───────┬────────┘                       │
     │                                │                              │
     │                        ┌───────┴────────┐                       │
     │                Safe?  │                │  Dangerous?            │
     │                        │                │                       │
     │                        ▼                ▼                       │
     │                   6. Call Planner   5E. Reject               │
     │                        │            ──── {                     │
     │                        │                  "status": "rejected", │
     │                        │                  "reason":             │
     │                        │                  "Dangerous keyword..."│
     │                        │                } ─────────────────>│
     │                        │                                    │
     │                   ┌────┴────┐                                │
     │                   │ Planner │                                │
     │                   │ CLI     │                                │
     │                   └────┬────┘                                │
     │                        │                                    │
     │                   7. Parse Response                          │
     │                      {                                      │
     │                        "intent": "benchmark",                │
     │                        "labels": {                           │
     │                          "os": "windows",                    │
     │                          "repo": "DiveAdstra"                │
     │                        },                                   │
     │                        "steps": [...]                        │
     │                      }                                      │
     │                        │                                    │
     │                   ┌────┴────┐                                │
     │                   │Node     │                                │
     │                   │Registry │                                │
     │                   └────┬────┘                                │
     │                        │                                    │
     │                   8. Determine Target                        │
     │                      Match labels:                           │
     │                      {"os": "windows", "repo": "DiveAdstra"}│
     │                      → Find node with highest score         │
     │                        │                                    │
     │                 ┌──────┴──────┐                              │
     │         Target found?      No target?                       │
     │                 │              │                             │
     │                 ▼              ▼                             │
     │           9. Create Task   8E. Reject                       │
     │              INSERT INTO   ──── {                           │
     │              tasks (...)       "status": "rejected",        │
     │              VALUES (...)      "reason": "No suitable node"  │
     │              │               } ─────────────────────────>│
     │              │                                                │
     │           10. Dispatch to Node                               │
     │              POST http://home_win:6767/api/v1/tasks          │
     │              Body: {                                         │
     │                "title": "Benchmark DiveAdstra",              │
     │                "description": "...",                         │
     │                "labels": {...},                              │
     │                "from_node_id": "cloud"                       │
     │              }                                               │
     │              │                                               │
     │              └──────────────────────────────────────────────>│
     │                                                              │
     │           11. Receive Task                                  │
     │              Store in local DB                              │
     │              Execute via TaskExecutor                        │
     │              │                                               │
     │              ▼                                               │
     │           12. Execute Task                                  │
     │              - git pull                                     │
     │              - cargo test --release                         │
     │              - Collect results                              │
     │              │                                               │
     │              ▼                                               │
     │           13. Update Task Status                            │
     │              UPDATE tasks SET status = 'completed'           │
     │              WHERE id = ...                                  │
     │                                                              │
     │           14. Return Response (Cloud Master)                │
     │              ──── {                                         │
     │                    "status": "dispatched",                  │
     │                    "task_id": 123,                          │
     │                    "target_node": "home_win",               │
     │                    "message": "已下发至 Windows 节点执行"    │
     │                  } ─────────────────────────────────────>│
```

---

## 4. Database Schema

```
┌─────────────────────────────────────────────────────────────────┐
│                        Cloud Master DB                          │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ Table: nodes                                                     │
├──────────┬─────────────┬─────────┬──────────┬─────────────────┤
│ Column   │ Type        │ Index   │ Nullable │ Description     │
├──────────┼─────────────┼─────────┼──────────┼─────────────────┤
│ id       │ TEXT        │ PK      │ NO       │ Node ID         │
│ fingerprint│ TEXT      │ UNIQUE  │ NO       │ MD5(pubkey)     │
│ endpoint │ TEXT        │         │ NO       │ http://...      │
│ last_heartbeat│ INTEGER│         │ NO       │ Unix timestamp  │
│ labels   │ TEXT        │         │ NO       │ JSON: {...}     │
│ status   │ TEXT        │ INDEX   │ NO       │ active|inactive │
│ created_at│ INTEGER    │         │ NO       │ Unix timestamp  │
│ updated_at│ INTEGER    │         │ NO       │ Unix timestamp  │
└──────────┴─────────────┴─────────┴──────────┴─────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ Table: trusted_nodes                                             │
├──────────┬─────────────┬─────────┬──────────┬─────────────────┤
│ Column   │ Type        │ Index   │ Nullable │ Description     │
├──────────┼─────────────┼─────────┼──────────┼─────────────────┤
│ fingerprint│ TEXT      │ PK      │ NO       │ MD5(pubkey)     │
│ added_at │ INTEGER     │         │ NO       │ Unix timestamp  │
│ added_by │ TEXT        │         │ YES      │ Admin ID        │
└──────────┴─────────────┴─────────┴──────────┴─────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ Table: tasks (extended)                                          │
├────────────────┬─────────┬──────────┬─────────────────────────┤
│ Column         │ Type    │ Nullable │ Description             │
├────────────────┼─────────┼──────────┼─────────────────────────┤
│ (existing cols)│ ...     │          │                         │
├────────────────┼─────────┼──────────┼─────────────────────────┤
│ target_node_id │ TEXT    │ YES      │ Where task is executed  │
│ from_node_id   │ TEXT    │ YES      │ Source node (cloud)     │
│ labels         │ TEXT    │ YES      │ JSON: {...}             │
└────────────────┴─────────┴──────────┴─────────────────────────┘

Foreign Keys:
  - target_node_id → nodes(id)
  - from_node_id → nodes(id)

Indexes:
  - idx_tasks_target_node ON tasks(target_node_id)
  - idx_nodes_status ON nodes(status)
  - idx_nodes_fingerprint ON nodes(fingerprint)
```

```
┌─────────────────────────────────────────────────────────────────┐
│                      Edge Node DB                               │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│ Table: tasks (same as cloud, local execution)                    │
├────────────────┬─────────┬──────────┬─────────────────────────┤
│ Column         │ Type    │ Nullable │ Description             │
├────────────────┼─────────┼──────────┼─────────────────────────┤
│ id             │ INTEGER │ NO       │ Primary key             │
│ task_id        │ TEXT    │ NO       │ UUID                    │
│ title          │ TEXT    │ NO       │ Task title              │
│ description    │ TEXT    │ YES      │ Task description        │
│ status         │ TEXT    │ NO       │ pending|running|...      │
│ priority       │ INTEGER │ NO       │ 0|1|2                   │
│ result         │ TEXT    │ YES      │ Execution result        │
│ error          │ TEXT    │ YES      │ Error message           │
│ target_node_id │ TEXT    │ YES      │ Always NULL (local)     │
│ from_node_id   │ TEXT    │ YES      │ "cloud"                 │
│ labels         │ TEXT    │ YES      │ Routing labels          │
│ created_at     │ TEXT    │ NO       │ ISO timestamp           │
│ started_at     │ TEXT    │ YES      │ ISO timestamp           │
│ completed_at   │ TEXT    │ YES      │ ISO timestamp           │
└────────────────┴─────────┴──────────┴─────────────────────────┘

Note: Edge nodes do NOT have nodes or trusted_nodes tables.
They only execute tasks and store results locally.
```

---

## 5. Security Layer Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    Request Flow (Webhook)                       │
└─────────────────────────────────────────────────────────────────┘

Incoming Request
│
▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: Network / Infrastructure                             │
│ ┌─────────────────────────────────────────────────────────┐  │
│ │ • Reverse Proxy (Nginx/Caddy) with TLS                 │  │
│ │ • Firewall rules (allow Zhipu AI IP range)             │  │
│ │ • DDoS protection (Cloudflare/AWS Shield)              │  │
│ └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Authentication                                      │
│ ┌─────────────────────────────────────────────────────────┐  │
│ │ • Extract Authorization header                          │  │
│ │ • Verify Bearer token format                           │  │
│ │ • Compare with configured secret                       │  │
│ │ • Return 401 if invalid/missing                        │  │
│ │                                                          │  │
│ │ Code: middleware/auth.rs                                │  │
│ └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: Rate Limiting                                       │
│ ┌─────────────────────────────────────────────────────────┐  │
│ │ • Extract client IP from headers                        │  │
│ │ • Check in-memory rate limiter                          │  │
│ │ • Increment request count for IP                        │  │
│ │ • Return 429 if exceeds limit (10 req/min)             │  │
│ │                                                          │  │
│ │ Code: middleware/rate_limit.rs                          │  │
│ └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 4: Authorization (Business Logic)                     │
│ ┌─────────────────────────────────────────────────────────┐  │
│ │ • Parse request body                                    │  │
│ │ • Check action against allowed_actions whitelist        │  │
│ │ • Detect dangerous keywords                             │  │
│ │ • Return 400 if action not allowed or dangerous         │  │
│ │                                                          │  │
│ │ Code: middleware/safety.rs                              │  │
│ └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 5: Business Logic                                     │
│ ┌─────────────────────────────────────────────────────────┐  │
│ │ • Call planner for intent analysis                     │  │
│ │ • Determine target node                                │  │
│ │ • Dispatch task                                        │  │
│ │                                                          │  │
│ │ Code: routes/webhook.rs                                │  │
│ └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
│
▼
Response
```

---

## 6. Component Interaction Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Component Map                           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Rust Crates & Modules                                           │
└─────────────────────────────────────────────────────────────────┘

agentflow-core/
├── types.rs              (Task, TaskStatus, Worker types)
├── database.rs           (DbPool, create_pool)
├── memory/
│   └── mod.rs            (MemoryCore - already exists)
├── executor/
│   └── mod.rs            (TaskExecutor - already exists)
└── node/                 (NEW)
    ├── mod.rs            (Node, NodeRegistry, NodeStatus)
    └── planner.rs        (Planner CLI integration)

agentflow-master/
├── main.rs               (Server startup, AppState)
├── config.rs             (MasterConfig - EXTEND with NodeConfig, WebhookConfig)
├── executor.rs           (TaskExecutor - already exists)
├── memory_core.rs        (MemoryCore - already exists)
├── dispatcher.rs         (NEW - Task dispatcher to nodes)
├── routes/
│   ├── mod.rs            (Route definitions - EXTEND)
│   ├── tasks.rs          (Task API - already exists)
│   ├── health.rs         (Health check - already exists)
│   ├── memory.rs         (Memory API - already exists)
│   ├── websocket.rs      (Task WebSocket - already exists)
│   ├── webhook.rs        (NEW - Zhipu webhook endpoint)
│   └── node_ws.rs        (NEW - Node WebSocket endpoint)
└── middleware/           (NEW)
    ├── mod.rs
    ├── auth.rs           (Bearer token verification)
    ├── rate_limit.rs     (Rate limiting)
    └── safety.rs         (Dangerous operation detection)

migrations/
├── 001_init_db.sql       (Existing tasks table)
└── 002_add_nodes.sql     (NEW - nodes, trusted_nodes tables)
```

---

## 7. Data Flow: Task Creation to Execution

```
Zhipu AI Request
    {"text": "Run benchmark on DiveAdstra"}
                │
                ▼
        ┌───────────────┐
        │ Webhook       │
        │ Handler      │
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │ Auth Check    │ ← Bearer token
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │ Rate Limit    │ ← IP-based, 10/min
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │ Safety Check  │ ← Dangerous keywords
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │ Planner CLI   │ ← Intent + task breakdown
        └───────┬───────┘
                │
                │ Returns:
                │ {
                │   "intent": "benchmark",
                │   "labels": {"os": "windows", "repo": "DiveAdstra"},
                │   "steps": [...]
                │ }
                ▼
        ┌───────────────┐
        │ NodeRegistry  │ ← Determine target node
        └───────┬───────┘
                │
                │ Returns: Node { id: "home_win", endpoint: "http://..." }
                ▼
        ┌───────────────┐
        │ Create Task   │ ← INSERT INTO cloud DB
        └───────┬───────┘
                │
                │ task_id = 123
                ▼
        ┌───────────────┐
        │ Dispatcher    │ ← HTTP POST to edge node
        └───────┬───────┘
                │
                │ POST http://home_win:6767/api/v1/tasks
                │ Body: {title, description, labels, from_node_id}
                ▼
┌───────────────────────────┐
│ Edge Node (home_win)      │
│                           │
│ ┌─────────────────────┐   │
│ │ Receive Task        │   │
│ └─────────┬───────────┘   │
│           │               │
│           ▼               │
│ ┌─────────────────────┐   │
│ │ Store in Local DB   │   │
│ └─────────┬───────────┘   │
│           │               │
│           ▼               │
│ ┌─────────────────────┐   │
│ │ TaskExecutor        │   │
│ │ - Execute steps     │   │
│ │ - Update status     │   │
│ └─────────┬───────────┘   │
│           │               │
│           ▼               │
│ ┌─────────────────────┐   │
│ │ Mark Completed      │   │
│ └─────────────────────┘   │
└───────────────────────────┘
```

---

## 8. Configuration File Structure

```toml
# ~/.agentflow/config.toml or ./agentflow.toml

[server]
mode = "cloud"              # local | cloud | planner-only
host = "0.0.0.0"
port = 6767

[database]
url = "sqlite://agentflow.db"

[node]
heartbeat_timeout = 90      # seconds
cleanup_interval = 300      # seconds (5 minutes)

[webhook]
secret = "your-secret-here-change-in-production"
rate_limit_per_ip = 10      # requests per minute
allowed_actions = ["run_test", "benchmark", "status"]
dangerous_keywords = ["install", "system", "rm -rf", "sudo", "format"]

[cli.worker_safe]
command = "claude"
args = ["-p", "--dangerously-skip-permissions"]

[cli.planner]
command = "claude"
args = ["-p", "--model", "claude-3-haiku"]

[sandbox]
enabled = true
default_workspace = "/tmp/agentflow/workspace"
allow_network = false
max_memory = "1G"
max_cpu = 2
timeout = 300

[memory]
backend = "memory"
default_ttl = 3600
max_entries = 10000
enable_persistence = false

[logging]
level = "info"
```

---

**Document Version**: 1.0
**Last Updated**: 2026-01-28
**Status**: Research Complete - Ready for Implementation
