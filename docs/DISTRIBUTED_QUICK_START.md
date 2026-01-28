# 分布式执行系统 - 快速开始指南

## 🚀 5 分钟快速开始

### 1. 验证构建

```bash
cd rust
./verify-distributed-build.sh
```

### 2. 启动 Master 集群 (3 个节点)

**终端 1 - Master 1**:
```bash
cargo run --bin agentflow-master -- \
  --node-id master-1 \
  --port 6767 \
  --peers master-1:6767,master-2:6768,master-3:6769
```

**终端 2 - Master 2**:
```bash
cargo run --bin agentflow-master -- \
  --node-id master-2 \
  --port 6768 \
  --peers master-1:6767,master-2:6768,master-3:6769
```

**终端 3 - Master 3**:
```bash
cargo run --bin agentflow-master -- \
  --node-id master-3 \
  --port 6769 \
  --peers master-1:6767,master-2:6768,master-3:6769
```

### 3. 启动 Workers

**Worker 1**:
```bash
cargo run --bin agentflow-worker -- \
  --worker-id worker-1 \
  --master-url http://localhost:6767 \
  --group builders
```

**Worker 2**:
```bash
cargo run --bin agentflow-worker -- \
  --worker-id worker-2 \
  --master-url http://localhost:6767 \
  --group testers
```

### 4. 提交任务

```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Build Project",
    "description": "Build and test the project",
    "priority": "high",
    "dependencies": [],
    "group_name": "builders"
  }'
```

---

## 📋 常用命令

### Master 节点管理

```bash
# 查看集群状态
curl http://localhost:6767/api/v1/cluster/status

# 查看当前 Leader
curl http://localhost:6767/api/v1/cluster/leader

# 添加节点
curl -X POST http://localhost:6767/api/v1/cluster/nodes \
  -H "Content-Type: application/json" \
  -d '{"node_id": "master-4", "address": "master-4:6770"}'
```

### 任务管理

```bash
# 创建工作流
curl -X POST http://localhost:6767/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ci-pipeline",
    "tasks": [
      {"id": "build", "dependencies": []},
      {"id": "test", "dependencies": ["build"]},
      {"id": "deploy", "dependencies": ["test"]}
    ]
  }'

# 查看任务状态
curl http://localhost:6767/api/v1/tasks/{task_id}

# 取消任务
curl -X DELETE http://localhost:6767/api/v1/tasks/{task_id}

# 调整任务优先级
curl -X PATCH http://localhost:6767/api/v1/tasks/{task_id}/priority \
  -H "Content-Type: application/json" \
  -d '{"priority": "urgent"}'
```

### Worker 管理

```bash
# 查看所有 Workers
curl http://localhost:6767/api/v1/workers

# 查看特定 Worker
curl http://localhost:6767/api/v1/workers/{worker_id}

# 标记 Worker 为维护模式
curl -X PATCH http://localhost:6767/api/v1/workers/{worker_id} \
  -H "Content-Type: application/json" \
  -d '{"status": "draining"}'

# 查看 Worker 资源
curl http://localhost:6767/api/v1/workers/{worker_id}/resources
```

### 分布式锁

```bash
# 获取锁
curl -X POST http://localhost:6767/api/v1/locks/acquire \
  -H "Content-Type: application/json" \
  -d '{"lock_key": "deploy-lock", "ttl": 300}'

# 释放锁
curl -X POST http://localhost:6767/api/v1/locks/release \
  -H "Content-Type: application/json" \
  -d '{"lock_key": "deploy-lock"}'

# 查看锁状态
curl http://localhost:6767/api/v1/locks/deploy-lock
```

---

## 🔧 配置示例

### Master 配置 (`master-config.toml`)

```toml
[node]
node_id = "master-1"
listen_address = "0.0.0.0:6767"

[raft]
election_timeout_ms = 5000
heartbeat_interval_ms = 2000
replication_timeout_ms = 3000
peers = ["master-1:6767", "master-2:6768", "master-3:6769"]

[scheduler]
max_concurrent_tasks = 100
task_timeout_secs = 3600
retry_limit = 3

[database]
url = "sqlite:agentflow.db"
```

### Worker 配置 (`worker-config.toml`)

```toml
[worker]
worker_id = "worker-1"
group_name = "builders"
master_url = "http://localhost:6767"

[resources]
cpu_cores = 8
total_memory_mb = 16384
max_concurrent_tasks = 4

[health]
heartbeat_interval_secs = 10
timeout_secs = 30
```

---

## 🧪 测试场景

### 场景 1: Leader 故障转移

```bash
# 1. 启动 3 个 Master 节点（见上面）
# 2. 查看 Leader
curl http://localhost:6767/api/v1/cluster/leader

# 3. 停止 Leader 节点 (Ctrl+C)
# 4. 观察自动选举新的 Leader
curl http://localhost:6768/api/v1/cluster/leader

# 5. 重新启动节点，观察它自动加入集群
```

### 场景 2: 复杂工作流执行

```bash
# 创建 CI/CD 工作流
curl -X POST http://localhost:6767/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d @- << EOF
{
  "name": "ci-cd-pipeline",
  "tasks": [
    {
      "id": "lint",
      "name": "Run Linter",
      "priority": "medium",
      "dependencies": []
    },
    {
      "id": "build",
      "name": "Build Project",
      "priority": "high",
      "dependencies": ["lint"]
    },
    {
      "id": "unit-test",
      "name": "Unit Tests",
      "priority": "high",
      "dependencies": ["build"]
    },
    {
      "id": "integration-test",
      "name": "Integration Tests",
      "priority": "medium",
      "dependencies": ["build"]
    },
    {
      "id": "deploy-staging",
      "name": "Deploy to Staging",
      "priority": "medium",
      "dependencies": ["unit-test", "integration-test"]
    },
    {
      "id": "smoke-test",
      "name": "Smoke Tests",
      "priority": "high",
      "dependencies": ["deploy-staging"]
    },
    {
      "id": "deploy-production",
      "name": "Deploy to Production",
      "priority": "urgent",
      "dependencies": ["smoke-test"]
    }
  ]
}
EOF

# 监控执行进度
curl -s http://localhost:6767/api/v1/workflows/ci-cd-pipeline | jq .
```

### 场景 3: Worker 负载均衡

```bash
# 1. 启动多个 Workers (不同组)
# 2. 提交多个任务到不同组

# Builder 组任务
for i in {1..10}; do
  curl -X POST http://localhost:6767/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{\"title\": \"Build ${i}\", \"group_name\": \"builders\"}"
done

# Tester 组任务
for i in {1..10}; do
  curl -X POST http://localhost:6767/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{\"title\": \"Test ${i}\", \"group_name\": \"testers\"}"
done

# 3. 观察 Worker 负载分布
watch -n 1 'curl -s http://localhost:6767/api/v1/workers | jq .[]'
```

### 场景 4: 分布式锁协调

```bash
# 终端 1: 尝试获取部署锁
curl -X POST http://localhost:6767/api/v1/locks/acquire \
  -H "Content-Type: application/json" \
  -d '{"lock_key": "deploy-production", "ttl": 60}'

# 终端 2: 同时尝试获取同一个锁 (会被阻塞或拒绝)
curl -X POST http://localhost:6768/api/v1/locks/acquire \
  -H "Content-Type: application/json" \
  -d '{"lock_key": "deploy-production", "ttl": 60}'

# 终端 1: 释放锁
curl -X POST http://localhost:6767/api/v1/locks/release \
  -H "Content-Type: application/json" \
  -d '{"lock_key": "deploy-production"}'

# 终端 2: 现在可以获取锁了
```

---

## 🐛 故障排查

### 问题 1: Leader 选举卡住

```bash
# 检查网络连通性
nc -zv master-2 6768
nc -zv master-3 6769

# 检查 Raft 日志
tail -f agentflow-master.log | grep -i raft

# 手动触发选举
curl -X POST http://localhost:6767/api/v1/cluster/election
```

### 问题 2: Worker 无法注册

```bash
# 检查 Master 是否可达
curl http://localhost:6767/health

# 检查 Worker 日志
tail -f agentflow-worker.log

# 手动注册 Worker
curl -X POST http://localhost:6767/api/v1/workers/register \
  -H "Content-Type: application/json" \
  -d @worker-config.json
```

### 问题 3: 任务依赖循环

```bash
# 验证工作流 DAG
curl -X POST http://localhost:6767/api/v1/workflows/validate \
  -H "Content-Type: application/json" \
  -d @workflow.json

# 查看依赖图
curl http://localhost:6767/api/v1/workflows/my-workflow/graph | jq .
```

---

## 📊 监控

### 使用 Prometheus 指标

```bash
# Master 暴露的指标
curl http://localhost:6767/metrics

# 关键指标
# - raft_leader_count
# - task_queue_length
# - worker_count
# - active_locks
# - message_throughput
```

### Grafana 仪表板

导入预配置的仪表板:
```
docs/grafana/distributed-execution-dashboard.json
```

---

## 📚 更多资源

- **完整文档**: [DISTRIBUTED_EXECUTION_SYSTEM.md](./DISTRIBUTED_EXECUTION_SYSTEM.md)
- **实施状态**: [DISTRIBUTED_EXECUTION_STATUS.md](./DISTRIBUTED_EXECUTION_STATUS.md)
- **API 文档**: http://localhost:6767/docs
- **Rust Docs**: `target/doc/agentflow_master/index.html`

---

## 🎯 下一步

1. ✅ 验证构建: `./rust/verify-distributed-build.sh`
2. 🧪 运行测试: `cargo test --package agentflow-master`
3. 🚀 启动集群: 按照上述步骤
4. 📊 配置监控: Prometheus + Grafana
5. 🚢 生产部署: Docker Compose / Kubernetes
