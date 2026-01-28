# Team C: Cloud/Edge Federation & Webhook Implementation Plan

**Version**: AgentFlow v0.2.1
**Date**: 2026-01-28
**Branch**: feature/0.2.1
**Team**: Team C - Cloud/Edge Federation & Webhook Integration

---

## 📋 Executive Summary

This plan details the implementation of distributed node architecture with Zhipu Webhook integration for AgentFlow v0.2.1. The system will evolve from a single-process Master=Worker architecture to a cloud-edge federation where:

- **Cloud Master**: Acts as the central coordinator, receives webhooks from Zhipu AI
- **Edge Nodes**: Register with cloud master, execute tasks locally, report results
- **Security**: Fingerprint-based authentication, rate limiting, dangerous operation detection
- **Communication**: WebSocket for node heartbeat, HTTP for task dispatch

---

## 🎯 Scope & Objectives

### Primary Goals
1. Implement node registration and identity management system
2. Create Hello/Heartbeat protocol using WebSocket
3. Build `/api/v1/webhook/zhipu` endpoint for task intake
4. Implement security middleware (authentication, rate limiting, input validation)
5. Design task routing and dispatch mechanism

### Non-Goals
- Full multi-node task scheduling optimization (deferred to future versions)
- Automatic node discovery (manual registration for v0.2.1)
- Complex load balancing algorithms (simple label-based routing)

---

## 🏗️ Architecture Overview

### Current State (v0.2.0)
```
┌─────────────────────────────────┐
│  agentflow-master               │
│  - HTTP API                     │
│  - WebSocket (task updates)     │
│  - TaskExecutor (local only)    │
│  - MemoryCore                   │
└─────────────────────────────────┘
```

### Target State (v0.2.1)
```
                    ┌──────────────────────┐
                    │  Zhipu AI Webhook    │
                    │  (Mobile App)        │
                    └──────────┬───────────┘
                               │ POST /api/v1/webhook/zhipu
                               ↓
┌──────────────────────────────────────────────────────┐
│  Cloud Master (agentflow server cloud)               │
│                                                       │
│  ┌───────────────────────────────────────────────┐  │
│  │  Webhook Handler                              │  │
│  │  - Bearer token auth                          │  │
│  │  - Rate limiting                              │  │
│  │  - Dangerous op detection                     │  │
│  └───────────┬───────────────────────────────────┘  │
│              ↓                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │  Planner (cli.planner)                        │  │
│  │  - Intent recognition                         │  │
│  │  - Task breakdown                             │  │
│  └───────────┬───────────────────────────────────┘  │
│              ↓                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │  NodeRegistry                                 │  │
│  │  - determine_target(task_plan)                │  │
│  │  - Route to appropriate node                  │  │
│  └───────────┬───────────────────────────────────┘  │
│              ↓                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │  Task Dispatcher                              │  │
│  │  - POST to target node /api/v1/tasks          │  │
│  └───────────────────────────────────────────────┘  │
│                                                       │
│  ┌───────────────────────────────────────────────┐  │
│  │  WebSocket Server                             │  │
│  │  - /ws/node (node connection)                 │  │
│  │  - Hello handshake                            │  │
│  │  - Heartbeat (30s interval)                   │  │
│  └───────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
          ▲                                    │
          │ Hello + Heartbeat                  │ Task dispatch
          │                                    ↓
┌─────────┴────────┐                  ┌────────────────┐
│  Edge Node 1     │                  │  Edge Node 2    │
│  (home_win)      │                  │  (work_mac)     │
│                  │                  │                 │
│  - TaskExecutor  │                  │  - TaskExecutor │
│  - Local SQLite  │                  │  - Local SQLite │
│  - Fingerprint   │                  │  - Fingerprint  │
└──────────────────┘                  └─────────────────┘
```

---

## 📦 Deliverables

### 1. Core Modules

#### A. `rust/agentflow-core/src/node/mod.rs`

**Purpose**: Node registration, identity management, and registry

**Key Types**:
```rust
/// Node status
#[derive(Debug, Clone, PartialEq, sqlx::Type)]
#[sqlx(type_name = "node_status", rename_all = "lowercase")]
pub enum NodeStatus {
    Active,
    Inactive,
}

/// Node information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Node {
    pub id: String,                      // Unique node ID (e.g., "home_win")
    pub fingerprint: String,              // MD5 of public key
    pub endpoint: String,                 // e.g., "http://192.168.1.100:6767"
    pub last_heartbeat: DateTime<Utc>,
    pub labels: HashMap<String, String>,  // e.g., {"os": "windows", "repo": "DiveAdstra"}
    pub status: NodeStatus,
    pub created_at: DateTime<Utc>,
}

/// Node registration request
#[derive(Debug, Deserialize)]
pub struct NodeRegistration {
    pub node_id: String,
    pub fingerprint: String,
    pub endpoint: String,
    pub labels: HashMap<String, String>,
}

/// Hello message (WebSocket)
#[derive(Debug, Serialize, Deserialize)]
pub struct HelloMessage {
    pub node_id: String,
    pub fingerprint: String,
    pub timestamp: DateTime<Utc>,
}

/// Heartbeat message (WebSocket)
#[derive(Debug, Serialize, Deserialize)]
pub struct HeartbeatMessage {
    pub node_id: String,
    pub timestamp: DateTime<Utc>,
}
```

**NodeRegistry API**:
```rust
pub struct NodeRegistry {
    db: DbPool,
    heartbeat_timeout: Duration,  // Default: 90 seconds (3 * 30s)
}

impl NodeRegistry {
    /// Register a new node (called on Hello)
    pub async fn register(&self, info: NodeRegistration) -> Result<Node>;

    /// Update heartbeat timestamp
    pub async fn heartbeat(&self, node_id: &str) -> Result<()>;

    /// Get node by ID
    pub async fn get_node(&self, node_id: &str) -> Result<Option<Node>>;

    /// List all active nodes
    pub async fn list_active_nodes(&self) -> Result<Vec<Node>>;

    /// Mark inactive nodes (called by background task)
    pub async fn mark_inactive_nodes(&self) -> Result<usize>;

    /// Determine target node for task
    /// Strategy: Match task.labels against node.labels
    pub async fn determine_target(&self, task_labels: &HashMap<String, String>)
        -> Result<Option<Node>>;

    /// Add trusted fingerprint (manual admin operation)
    pub async fn add_trusted_fingerprint(&self, fingerprint: String) -> Result<()>;

    /// Check if fingerprint is trusted
    pub async fn is_trusted(&self, fingerprint: &str) -> Result<bool>;
}
```

#### B. `rust/agentflow-master/src/routes/webhook.rs`

**Purpose**: Handle Zhipu AI webhook requests

**Webhook Request Structure**:
```rust
#[derive(Debug, Deserialize)]
pub struct ZhipuWebhookRequest {
    pub text: String,  // Natural language task description
}

#[derive(Debug, Serialize)]
pub struct ZhipuWebhookResponse {
    pub status: String,      // "dispatched" | "rejected" | "error"
    pub task_id: Option<i64>,
    pub target_node: Option<String>,
    pub message: String,
    pub reason: Option<String>,  // If rejected
}
```

**Processing Flow**:
```rust
pub async fn handle_zhipu_webhook(
    State(state): State<AppState>,
    headers: HeaderMap,
    Json(req): Json<ZhipuWebhookRequest>,
) -> Result<Json<ZhipuWebhookResponse>, WebhookError> {
    // 1. Verify Bearer token
    verify_auth_token(&headers, &state.config.webhook.secret)?;

    // 2. Check rate limit (by IP)
    let ip = extract_ip(&headers);
    check_rate_limit(&state.rate_limiter, &ip).await?;

    // 3. Detect dangerous operations
    detect_dangerous_ops(&req.text)?;

    // 4. Call planner for intent + task breakdown
    let plan = call_planner(&req.text).await?;

    // 5. Determine target node
    let target = state.node_registry
        .determine_target(&plan.labels)
        .await?;

    let target = match target {
        Some(t) => t,
        None => return Ok(Json(ZhipuWebhookResponse {
            status: "rejected".to_string(),
            task_id: None,
            target_node: None,
            message: "No suitable node available".to_string(),
            reason: Some("No node matches task requirements".to_string()),
        })),
    };

    // 6. Create task in local DB (for tracking)
    let task_id = create_task(&state.db, &plan, &target.id).await?;

    // 7. Dispatch to target node
    dispatch_task(&target.endpoint, &task_id, &plan).await?;

    // 8. Return response
    Ok(Json(ZhipuWebhookResponse {
        status: "dispatched".to_string(),
        task_id: Some(task_id),
        target_node: Some(target.id.clone()),
        message: format!("已下发至 {} 节点执行", target.id),
        reason: None,
    }))
}
```

#### C. `rust/agentflow-master/src/routes/node_ws.rs`

**Purpose**: WebSocket endpoint for node connection and heartbeat

**WebSocket Protocol**:
```rust
/// WebSocket upgrade endpoint
pub async fn node_websocket(
    ws: WebSocketUpgrade,
    State(state): State<AppState>,
) -> impl IntoResponse {
    ws.on_upgrade(|socket| handle_node_connection(socket, state))
}

async fn handle_node_connection(socket: WebSocket, state: AppState) {
    let (mut sender, mut receiver) = socket.split();

    // Phase 1: Wait for Hello message
    let hello_msg = match receive_hello(&mut receiver).await {
        Ok(msg) => msg,
        Err(e) => {
            send_error(&mut sender, &format!("Invalid Hello: {}", e)).await;
            return;
        }
    };

    // Phase 2: Verify fingerprint
    match state.node_registry.is_trusted(&hello_msg.fingerprint).await {
        Ok(true) => {},
        Ok(false) => {
            send_error(&mut sender, "Fingerprint not in trusted list").await;
            warn!("Untrusted fingerprint: {} (node: {})",
                  hello_msg.fingerprint, hello_msg.node_id);
            return;
        },
        Err(e) => {
            send_error(&mut sender, "Database error").await;
            error!("Failed to verify fingerprint: {}", e);
            return;
        }
    }

    // Phase 3: Register node
    let node = match state.node_registry.register(NodeRegistration {
        node_id: hello_msg.node_id.clone(),
        fingerprint: hello_msg.fingerprint,
        endpoint: extract_endpoint(&sender),  // Extract peer address
        labels: hello_msg.labels,
    }).await {
        Ok(n) => n,
        Err(e) => {
            send_error(&mut sender, &format!("Registration failed: {}", e)).await;
            return;
        }
    };

    info!("Node connected: {} ({})", node.id, node.endpoint);
    send_ack(&mut sender, &node.id).await;

    // Phase 4: Heartbeat loop
    let mut heartbeat_interval = tokio::time::interval(Duration::from_secs(30));
    let mut last_heartbeat = Instant::now();

    loop {
        tokio::select! {
            // Receive messages from node
            msg = receiver.next() => {
                match msg {
                    Some(Ok(Message::Text(text))) => {
                        if let Ok(hb) = serde_json::from_str::<HeartbeatMessage>(&text) {
                            if hb.node_id == node.id {
                                state.node_registry.heartbeat(&node.id).await.ok();
                                last_heartbeat = Instant::now();
                                send_pong(&mut sender).await;
                            }
                        }
                    },
                    Some(Ok(Message::Close(_))) | None => break,
                    _ => {},
                }
            },

            // Send heartbeat requests
            _ = heartbeat_interval.tick() => {
                if last_heartbeat.elapsed() > Duration::from_secs(90) {
                    warn!("Node {} heartbeat timeout", node.id);
                    break;
                }
                send_ping(&mut sender).await;
            },
        }
    }

    info!("Node disconnected: {}", node.id);
}
```

### 2. Database Schema

**New Tables**:
```sql
-- Nodes table
CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    fingerprint TEXT NOT NULL UNIQUE,
    endpoint TEXT NOT NULL,
    last_heartbeat INTEGER NOT NULL,  -- Unix timestamp
    labels TEXT NOT NULL,              -- JSON: {"os": "windows", "repo": "..."}
    status TEXT NOT NULL DEFAULT 'active',  -- active | inactive
    created_at INTEGER NOT NULL,       -- Unix timestamp
    updated_at INTEGER NOT NULL        -- Unix timestamp
);

CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_fingerprint ON nodes(fingerprint);

-- Trusted fingerprints table
CREATE TABLE IF NOT EXISTS trusted_nodes (
    fingerprint TEXT PRIMARY KEY,
    added_at INTEGER NOT NULL,         -- Unix timestamp
    added_by TEXT                      -- Admin identifier
);

-- Tasks table (extend existing)
ALTER TABLE tasks ADD COLUMN target_node_id TEXT;
ALTER TABLE tasks ADD COLUMN from_node_id TEXT;
ALTER TABLE tasks ADD COLUMN labels TEXT;
CREATE INDEX idx_tasks_target_node ON tasks(target_node_id);
```

**Migration Script**: `rust/migrations/002_add_nodes.sql`

### 3. Configuration Extensions

**Add to `rust/agentflow-master/src/config.rs`**:
```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MasterConfig {
    // ... existing fields ...

    /// Node registry configuration
    pub node: NodeConfig,

    /// Webhook configuration
    pub webhook: WebhookConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeConfig {
    /// Heartbeat timeout in seconds
    pub heartbeat_timeout: u64,
    /// Node cleanup interval in seconds
    pub cleanup_interval: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WebhookConfig {
    /// Bearer token secret
    pub secret: String,
    /// Rate limit per IP (requests per minute)
    pub rate_limit_per_ip: u64,
    /// Allowed actions (whitelist)
    pub allowed_actions: Vec<String>,
    /// Dangerous keywords (blacklist)
    pub dangerous_keywords: Vec<String>,
}

impl Default for NodeConfig {
    fn default() -> Self {
        Self {
            heartbeat_timeout: 90,
            cleanup_interval: 300,
        }
    }
}

impl Default for WebhookConfig {
    fn default() -> Self {
        Self {
            secret: "change-me-in-production".to_string(),
            rate_limit_per_ip: 10,
            allowed_actions: vec!["run_test".into(), "benchmark".into(), "status".into()],
            dangerous_keywords: vec!["install".into(), "system".into(), "rm -rf".into()],
        }
    }
}
```

**Environment Variables** (add to `.env.example`):
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

### 4. Security Middleware

#### A. `rust/agentflow-master/src/middleware/auth.rs`

**Bearer Token Authentication**:
```rust
use axum::{
    extract::Request,
    http::HeaderMap,
    middleware::Next,
    response::Response,
};

pub async fn verify_webhook_auth(
    headers: HeaderMap,
    secret: &str,
) -> Result<(), AuthError> {
    let auth_header = headers
        .get("Authorization")
        .ok_or(AuthError::MissingToken)?;

    let token = auth_header
        .to_str()
        .map_err(|_| AuthError::InvalidToken)?
        .strip_prefix("Bearer ")
        .ok_or(AuthError::InvalidToken)?;

    if token == secret {
        Ok(())
    } else {
        Err(AuthError::InvalidToken)
    }
}

#[derive(Debug)]
pub enum AuthError {
    MissingToken,
    InvalidToken,
}

impl IntoResponse for AuthError {
    fn into_response(self) -> Response {
        let (status, message) = match self {
            AuthError::MissingToken => (StatusCode::UNAUTHORIZED, "Missing bearer token"),
            AuthError::InvalidToken => (StatusCode::UNAUTHORIZED, "Invalid bearer token"),
        };
        (status, Json(json!({"error": message}))).into_response()
    }
}
```

#### B. `rust/agentflow-master/src/middleware/rate_limit.rs`

**Rate Limiting (In-Memory)**:
```rust
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

pub struct RateLimiter {
    // IP -> (request_count, window_start)
    requests: Arc<RwLock<HashMap<String, (u64, Instant)>>>,
    max_requests: u64,
    window_duration: Duration,
}

impl RateLimiter {
    pub fn new(max_requests: u64, window_duration: Duration) -> Self {
        Self {
            requests: Arc::new(RwLock::new(HashMap::new())),
            max_requests,
            window_duration,
        }
    }

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
}

#[derive(Debug)]
pub enum RateLimitError {
    TooManyRequests,
}

impl IntoResponse for RateLimitError {
    fn into_response(self) -> Response {
        (
            StatusCode::TOO_MANY_REQUESTS,
            Json(json!({"error": "Rate limit exceeded"})),
        )
            .into_response()
    }
}
```

#### C. `rust/agentflow-master/src/middleware/safety.rs`

**Dangerous Operation Detection**:
```rust
pub fn detect_dangerous_ops(text: &str) -> Result<(), SafetyError> {
    let dangerous_keywords = vec![
        "install", "system", "rm -rf", "sudo", "format",
        "mkfs", "fdisk", "shutdown", "reboot", "passwd",
    ];

    let text_lower = text.to_lowercase();

    for keyword in dangerous_keywords {
        if text_lower.contains(keyword) {
            return Err(SafetyError::DangerousOperation(keyword.to_string()));
        }
    }

    Ok(())
}

#[derive(Debug)]
pub enum SafetyError {
    DangerousOperation(String),
}

impl IntoResponse for SafetyError {
    fn into_response(self) -> Response {
        let message = match self {
            SafetyError::DangerousOperation(op) => {
                format!("Dangerous operation detected: {}", op)
            }
        };
        (
            StatusCode::BAD_REQUEST,
            Json(json!({"error": message, "status": "rejected"})),
        )
            .into_response()
    }
}
```

### 5. Planner Integration

**Planner Response Structure** (`rust/agentflow-core/src/planner.rs`):
```rust
#[derive(Debug, Deserialize)]
pub struct PlannerResponse {
    pub intent: String,                     // e.g., "benchmark"
    pub target_repository: Option<String>,  // e.g., "DiveAdstra"
    pub steps: Vec<String>,
    pub labels: HashMap<String, String>,    // Routing labels
    pub estimated_duration_minutes: u64,
}

/// Call planner CLI
pub async fn call_planner(task_description: &str) -> Result<PlannerResponse> {
    let output = tokio::process::Command::new("claude")
        .args(["-p", "--model", "claude-3-haiku"])
        .arg(format!(
            "Analyze this task and return JSON:\n{}\n\n\
            Return format:\n\
            {{\n\
              \"intent\": \"string\",\n\
              \"target_repository\": \"string or null\",\n\
              \"steps\": [\"step1\", \"step2\"],\n\
              \"labels\": {{\"os\": \"windows\", \"repo\": \"name\"}},\n\
              \"estimated_duration_minutes\": 10\n\
            }}",
            task_description
        ))
        .output()
        .await?;

    if !output.status.success() {
        return Err(anyhow::anyhow!("Planner failed: {}", String::from_utf8_lossy(&output.stderr)));
    }

    let json_str = String::from_utf8(output.stdout)?;
    let response: PlannerResponse = serde_json::from_str(&json_str)?;

    Ok(response)
}
```

### 6. Task Dispatcher

**`rust/agentflow-master/src/dispatcher.rs`**:
```rust
use reqwest::Client;

pub struct TaskDispatcher {
    client: Client,
}

impl TaskDispatcher {
    pub fn new() -> Self {
        Self {
            client: Client::new(),
        }
    }

    pub async fn dispatch_to_node(
        &self,
        node_endpoint: &str,
        task: &CreateTaskRequest,
    ) -> Result<Task> {
        let url = format!("{}/api/v1/tasks", node_endpoint);

        let response = self.client
            .post(&url)
            .json(task)
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(anyhow::anyhow!(
                "Node returned error: {}",
                response.status()
            ));
        }

        let api_response: ApiResponse<Task> = response.json().await?;

        api_response.data.ok_or_else(|| anyhow::anyhow!("No task data returned"))
    }
}
```

---

## 🔄 Detailed Workflow

### Webhook Processing Flow

```
1. Zhipu AI sends POST /api/v1/webhook/zhipu
   ↓
2. Auth middleware verifies Bearer token
   ↓
3. Rate limiter checks IP (10 req/min)
   ↓
4. Safety check for dangerous keywords
   ↓
5. Call planner (CLI) to analyze intent
   ↓
6. NodeRegistry.determine_target()
   - Match task.labels against node.labels
   - Return best node or None
   ↓
7a. If target found:
    - Create task in cloud DB
    - Dispatch to target node via HTTP
    - Return "dispatched" response
   ↓
7b. If no target:
    - Return "rejected" response with reason
```

### Node Registration Flow

```
1. Edge Node starts up
   ↓
2. Connect to Cloud Master: ws://cloud-master:6767/ws/node
   ↓
3. Send Hello message:
   {
     "type": "hello",
     "node_id": "home_win",
     "fingerprint": "abc123...",
     "labels": {"os": "windows", "repo": "DiveAdstra"}
   }
   ↓
4. Cloud Master:
    - Verifies fingerprint in trusted_nodes
    - If not trusted: Close connection, log warning
    - If trusted: Register in nodes table, status=active
   ↓
5. Start heartbeat loop (every 30s)
   ↓
6. Cloud Master updates last_heartbeat
   ↓
7. Background task (every 5min):
    - Mark nodes inactive if last_heartbeat > 90s ago
```

### Task Execution Flow (Distributed)

```
Cloud Side:
1. Receive webhook
2. Call planner → get task plan
3. Select target node
4. Create task in cloud DB (for tracking)
5. POST to target node /api/v1/tasks
   - Include from_node_id = "cloud"
   - Include labels for routing
6. Return response to webhook

Edge Side:
1. Receive task via POST /api/v1/tasks
2. Store in local SQLite
3. Execute via TaskExecutor (same as local mode)
4. Report result back to cloud (optional callback)
```

---

## 🔒 Security Considerations

### 1. Authentication Layers

| Layer | Mechanism | Purpose |
|-------|-----------|---------|
| Webhook | Bearer token | Prevent unauthorized webhook calls |
| Node | Fingerprint whitelist | Prevent untrusted nodes from connecting |
| Rate Limiting | IP-based | Prevent abuse/DoS |
| Safety Check | Keyword detection | Prevent dangerous operations |

### 2. Fingerprint Management

**Initial Setup** (manual):
```bash
# On Cloud Master
agentflow node add-trusted <fingerprint>

# Or via API (admin-only)
curl -X POST http://cloud-master:6767/api/v1/admin/trusted-nodes \
  -H "Authorization: Bearer <admin-secret>" \
  -d '{"fingerprint": "abc123..."}'
```

**Edge Node Generates Fingerprint**:
```bash
# On Edge Node (first time)
agentflow node fingerprint
# Output: abc123def456... (save this)

# Then start node
agentflow server --mode=edge --node-id=home_win
```

### 3. Configuration Best Practices

- Use strong Bearer token (64-char random string)
- Change default secret in production
- Use HTTPS in production (TLS termination)
- Restrict webhook endpoint via firewall if possible
- Log all authentication failures

---

## 📊 Monitoring & Observability

### Key Metrics

- **Node Health**: Active vs inactive nodes count
- **Heartbeat Latency**: Time between heartbeat send/receive
- **Task Dispatch**: Success rate, latency
- **Webhook**: Request rate, rejection rate
- **Security**: Auth failures, rate limit hits

### Logging Strategy

```rust
// Structured logging with tracing
info!(
    node_id = %node.id,
    fingerprint = %node.fingerprint,
    "Node registered"
);

warn!(
    ip = %ip,
    fingerprint = %fingerprint,
    "Untrusted node connection attempt"
);

error!(
    node_id = %node_id,
    error = %e,
    "Failed to dispatch task"
);
```

### Health Check Endpoint

`GET /api/v1/cluster/status`:
```json
{
  "cloud_master": {
    "status": "healthy",
    "uptime_seconds": 3600,
    "mode": "cloud"
  },
  "nodes": {
    "total": 3,
    "active": 2,
    "inactive": 1,
    "nodes": [
      {
        "id": "home_win",
        "status": "active",
        "last_heartbeat": "2026-01-28T10:30:00Z"
      },
      {
        "id": "work_mac",
        "status": "active",
        "last_heartbeat": "2026-01-28T10:29:55Z"
      },
      {
        "id": "old_server",
        "status": "inactive",
        "last_heartbeat": "2026-01-28T10:15:00Z"
      }
    ]
  }
}
```

---

## 🧪 Testing Strategy

### Unit Tests

1. **Node Registry**:
   - Register node
   - Heartbeat updates timestamp
   - Inactive detection after timeout
   - Target selection logic

2. **Security Middleware**:
   - Valid token passes
   - Invalid token rejected
   - Rate limiting enforcement
   - Dangerous keyword detection

3. **Planner Integration**:
   - Mock CLI output
   - Parse JSON response
   - Extract labels

### Integration Tests

1. **Node Registration Flow**:
   ```rust
   #[tokio::test]
   async fn test_node_registration_flow() {
       // Start cloud master
       // Connect edge node via WebSocket
       // Send Hello message
       // Verify node in database
       // Verify heartbeat updates
   }
   ```

2. **Webhook End-to-End**:
   ```rust
   #[tokio::test]
   async fn test_webhook_dispatch() {
       // Start cloud master with mock nodes
       // Send webhook request
       // Verify planner called
       // Verify task dispatched to correct node
       // Verify response format
   }
   ```

### Manual Testing

1. **Local Multi-Node Setup**:
   ```bash
   # Terminal 1: Cloud Master
   AGENTFLOW_MODE=cloud ./target/release/agentflow-master

   # Terminal 2: Edge Node 1
   AGENTFLOW_MODE=edge NODE_ID=test1 ./target/release/agentflow-master

   # Terminal 3: Edge Node 2
   AGENTFLOW_MODE=edge NODE_ID=test2 ./target/release/agentflow-master

   # Terminal 4: Test Webhook
   curl -X POST http://localhost:6767/api/v1/webhook/zhipu \
     -H "Authorization: Bearer test-secret" \
     -d '{"text": "Run benchmark on test1"}'
   ```

---

## 📅 Implementation Phases

### Phase 1: Foundation (Week 1, Days 1-3)

**Goal**: Database schema and core types

- [ ] Create migration file `002_add_nodes.sql`
- [ ] Implement `Node`, `NodeStatus` types in `agentflow-core/src/node/mod.rs`
- [ ] Implement `NodeRegistry` methods: register, heartbeat, get_node
- [ ] Add config structs to `agentflow-master/src/config.rs`
- [ ] Unit tests for NodeRegistry

### Phase 2: WebSocket & Protocol (Week 1, Days 4-5)

**Goal**: Node connection and heartbeat

- [ ] Implement `/ws/node` endpoint in `routes/node_ws.rs`
- [ ] Implement Hello/Heartbeat message handlers
- [ ] Implement fingerprint verification
- [ ] Add background task for inactive node cleanup
- [ ] Integration test for node registration flow

### Phase 3: Security Middleware (Week 2, Days 1-2)

**Goal**: Authentication and safety checks

- [ ] Implement auth middleware (`middleware/auth.rs`)
- [ ] Implement rate limiter (`middleware/rate_limit.rs`)
- [ ] Implement dangerous op detection (`middleware/safety.rs`)
- [ ] Add middleware to webhook route
- [ ] Unit tests for all middleware

### Phase 4: Planner & Dispatch (Week 2, Days 3-4)

**Goal**: Task routing and dispatch

- [ ] Implement planner integration (`agentflow-core/src/planner.rs`)
- [ ] Implement `NodeRegistry::determine_target`
- [ ] Implement task dispatcher (`dispatcher.rs`)
- [ ] Implement webhook handler (`routes/webhook.rs`)
- [ ] End-to-end test

### Phase 5: Integration & Polish (Week 2, Day 5)

**Goal**: Full system integration

- [ ] Update main.rs to load node config
- [ ] Add admin endpoint for trusted node management
- [ ] Add `/api/v1/cluster/status` endpoint
- [ ] Documentation (ZHIPU_INTEGRATION.md)
- [ ] Manual testing with real nodes

---

## 📚 Documentation Deliverables

### 1. `docs/ZHIPU_INTEGRATION.md`

- Zhipu AI webhook configuration
- Setting up cloud master
- Registering edge nodes
- Webhook body examples
- Security best practices
- Troubleshooting guide

### 2. `docs/NODE_ARCHITECTURE.md`

- Cloud-edge architecture diagram
- Node registration flow
- Heartbeat protocol details
- Task routing strategy
- Security model

### 3. Update `README.md`

- Add cloud mode quick start
- Add webhook usage example
- Add node management commands
- Update architecture diagram

---

## 🚀 Future Enhancements (Post v0.2.1)

1. **Automatic Node Discovery**: mDNS / service discovery
2. **Advanced Routing**: Load balancing, task priority queuing
3. **Result Callback**: Nodes report results back to cloud
4. **Task Federation**: Split large tasks across multiple nodes
5. **Web UI**: Dashboard for cluster status
6. **Metrics**: Prometheus integration
7. **TLS**: Mutual TLS for node communication

---

## ✅ Acceptance Criteria

### Functional Requirements

- [ ] Edge node can register with cloud master via WebSocket
- [ ] Fingerprint verification works (trusted nodes allowed, others rejected)
- [ ] Heartbeat mechanism detects inactive nodes within 90 seconds
- [ ] Webhook endpoint accepts requests with valid Bearer token
- [ ] Rate limiter blocks > 10 requests/minute per IP
- [ ] Dangerous operations are rejected with clear error message
- [ ] Planner is called and returns structured task plan
- [ ] Tasks are dispatched to correct node based on labels
- [ ] Cluster status endpoint shows active/inactive nodes

### Performance Requirements

- [ ] Webhook response time < 500ms (excluding planner)
- [ ] Node heartbeat processing < 10ms
- [ ] Can handle 100 concurrent node connections
- [ ] Memory overhead < 50MB for node registry

### Security Requirements

- [ ] All unauthenticated webhook requests return 401
- [ ] Rate limit violations return 429
- [ ] Dangerous operations return 400 with rejection
- [ ] Untrusted node connections are logged and rejected
- [ ] Fingerprint verification uses constant-time comparison

---

## 🎯 Success Metrics

**For v0.2.1 Release**:

1. **Deployment**: 3+ edge nodes successfully register to single cloud master
2. **Webhook**: Zhipu AI can dispatch tasks to Windows/Mac/Linux nodes
3. **Reliability**: 99%+ heartbeat success rate in stable network
4. **Security**: Zero unauthorized access in penetration testing
5. **Performance**: Sub-second webhook-to-dispatch latency

---

## 📝 Notes & Considerations

### Design Decisions

1. **WebSocket vs HTTP for Heartbeat**:
   - Chose WebSocket for bidirectional, low-overhead communication
   - Easier to detect disconnection
   - Allows future push notifications to nodes

2. **In-Memory Rate Limiting**:
   - Simple, fast, no external dependencies
   - Reset on server restart (acceptable for MVP)
   - Future: Redis for distributed rate limiting

3. **Fingerprint-based Auth**:
   - Simpler than full PKI
   - Manual whitelist is secure for small deployments
   - Future: Auto-registration with approval workflow

4. **Planner as CLI**:
   - Leverages existing Claude CLI integration
   - Keeps Rust codebase small
   - Future: HTTP-based planner service

### Known Limitations

1. **Single Cloud Master**: No HA/failover in v0.2.1
2. **Manual Trust Management**: No auto-approval workflow
3. **Simple Routing**: Label-based only, no intelligent scheduling
4. **No Result Callback**: Tasks dispatched fire-and-forget
5. **Limited Scalability**: In-memory rate limiting doesn't scale horizontally

### Dependencies on Other Teams

- **Team A (CLI)**: Need `agentflow server --mode=cloud` command
- **Team B (Memory)**: Optional - could enhance planner with memory search
- **Team D (Packaging)**: Need installation script for easy node deployment

---

## 🔗 References

- **Task Breakdown**: `/docs/plan/v0.2.1_TASK_BREAKDOWN.md`
- **Iteration Plan**: `/docs/plan/v0.2.1迭代计划.md`
- **Cloud Flow**: `/docs/plan/agentFlow云端流程.md`
- **Existing Routes**: `/rust/agentflow-master/src/routes/mod.rs`
- **Config System**: `/rust/agentflow-master/src/config.rs`

---

**Document Version**: 1.0
**Last Updated**: 2026-01-28
**Status**: Ready for Implementation
