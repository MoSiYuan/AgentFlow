# AgentFlow 集群部署指南

**目标**: 在多节点环境中部署 AgentFlow，实现高可用和负载均衡

**适用场景**:
- 单机多实例部署（简单负载均衡）
- 多机多实例部署（高可用）
- 分布式任务执行

---

## 📋 目录

1. [集群架构概述](#集群架构概述)
2. [部署模式](#部署模式)
3. [单机多实例部署](#单机多实例部署)
4. [多机多实例部署](#多机多实例部署)
5. [共享存储方案](#共享存储方案)
6. [升级路线图](#升级路线图)

---

## 集群架构概述

### 当前架构（v3.0）

```
┌─────────────────────────────────────────────────┐
│              AgentFlow v3.0 (单进程)            │
│                                                     │
│  ┌─────────────────────────────────────────────┐ │
│  │         agentflow-master                     │ │
│  │  (Master + Worker + Memory + SQLite)        │ │
│  └─────────────────────────────────────────────┘ │
│                                                     │
│  - 单进程架构（Master = Worker）                  │
│  - 本地 SQLite 存储                              │
│  - 适用于: 个人助手、小型团队                    │
└─────────────────────────────────────────────────┘
```

**特点**:
- ✅ 简单部署
- ✅ 低资源占用
- ⚠️ 单点故障
- ⚠️ 无任务共享

### 目标架构（未来 v3.1+）

```
┌─────────────────────────────────────────────────┐
│              负载均衡器 (Nginx/HAProxy)          │
└──────────────────┬──────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
┌───────────────┐          ┌───────────────┐
│ AgentFlow #1   │          │ AgentFlow #2   │
│ (Port 6767)   │          │ (Port 6768)   │
└───────┬────────┘          └───────┬────────┘
        │                          │
        └──────────┬───────────┘
                   ▼
        ┌────────────────────────┐
        │   共享任务队列 (Redis)   │
        │   共享记忆 (Qdrant)      │
        └────────────────────────┘
```

**特点**:
- ✅ 高可用
- ✅ 负载均衡
- ✅ 任务共享
- ✅ 记忆共享

---

## 部署模式

### 模式对比

| 部署模式 | 复杂度 | 高可用 | 任务共享 | 适用场景 |
|---------|--------|--------|---------|---------|
| **单机单实例** | ⭐ | ❌ | ❌ | 开发、个人使用 |
| **单机多实例** | ⭐⭐ | ⚠️ | ❌ | 小型团队、内网部署 |
| **多机多实例（无状态）** | ⭐⭐⭐ | ✅ | ❌ | 中型团队、外网访问 |
| **多机多实例（共享队列）** | ⭐⭐⭐⭐ | ✅ | ✅ | 大型团队、分布式任务 |

---

## 单机多实例部署

### 架构图

```
                    ┌─────────────┐
                    │   Nginx     │
                    │  (HAProxy)   │
                    └──────┬──────┘
                           │
        ┌──────────────────┴──────────┐
        ▼                                  ▼
┌─────────────┐                  ┌─────────────┐
│ AgentFlow #1 │                  │ AgentFlow #2 │
│ Port: 6767   │                  │ Port: 6768   │
└─────────────┘                  └─────────────┘
```

### 部署步骤

#### 1. 修改端口配置

**实例 1**:
```bash
# /etc/default/agentflow-instance-1
AGENTFLOW_SERVER_PORT=6767
```

**实例 2**:
```bash
# /etc/default/agentflow-instance-2
AGENTFLOW_SERVER_PORT=6768
```

#### 2. 配置 Nginx 负载均衡

```nginx
upstream agentflow_backend {
    least_conn;
    server 127.0.0.1:6767 weight=1;
    server 127.0.0.1:6768 weight=1;
}

server {
    listen 80;

    location / {
        proxy_pass http://agentflow_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

#### 3. 创建 systemd service（多实例）

**实例 1**:
```ini
[Unit]
Description=AgentFlow Master Instance 1
After=network.target

[Service]
Type=simple
User=agentflow
EnvironmentFile=/etc/default/agentflow-instance-1
ExecStart=/opt/agentflow/bin/agentflow-master
Restart=always

[Install]
WantedBy=multi-user.target
```

**实例 2**:
```ini
[Unit]
Description=AgentFlow Master Instance 2
After=network.target

[Service]
Type=simple
User=agentflow
EnvironmentFile=/etc/default/agentflow-instance-2
ExecStart=/opt/agentflow/bin/agentflow-master
Restart=always

[Install]
WantedBy=multi-user.target
```

#### 4. 启动服务

```bash
# 启动服务
sudo systemctl start agentflow-master@instance1
sudo systemctl start agentflow@instance2

# 设置开机自启
sudo systemctl enable agentflow-master@instance1
sudo systemctl enable agentflow@instance2

# 检查状态
sudo systemctl status agentflow-master@*
```

#### 5. 启动 Nginx

```bash
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 局限性说明

**优点**:
- ✅ 简单的负载均衡
- ✅ 提高并发处理能力
- ✅ 单实例故障时其他实例继续服务

**限制**:
- ⚠️ 任务队列不共享（每个实例独立 SQLite）
- ⚠️ 记忆存储不共享
- ⚠️ 单机故障导致所有实例不可用

**适用场景**:
- 内网部署（外网可通过 VPN 访问）
- 任务由外部系统分发
- 无需共享状态

---

## 多机多实例部署

### 架构图

```
                        ┌─────────────┐
                        │   负载均衡   │
                        │  (Cloud LB)  │
                        └──────┬──────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        ▼                      ▼                      ▼
┌──────────────┐      ┌───────────────┐      ┌───────────────┐
│  Server A     │      │  Server B      │      │  Server C     │
│  (AgentFlow)  │      │  (AgentFlow)  │      │  (AgentFlow)  │
│  Port: 6767   │      │  Port: 6767   │      │  Port: 6767   │
└───────┬───────┘      └───────┬───────┘      └───────┬───────┘
        │                      │                      │
        └──────────────────────┴──────────────────────┘
                               ▼
                   ┌────────────────────────┐
                   │   外部任务分发系统        │
                   │  (Your Orchestrator)     │
                   └────────────────────────┘
```

### 部署步骤

#### 1. 服务器准备

**在所有服务器上执行**:

```bash
# 安装 Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"

# 下载 AgentFlow
wget https://github.com/MoSiYuan/AgentFlow/releases/latest/download/agentflow-master-linux-amd64
chmod +x agentflow-master-linux-amd64
sudo cp agentflow-master-linux-amd64 /usr/local/bin/agentflow-master

# 创建用户
sudo useradd -r -s /bin/bash agentflow
```

#### 2. 配置防火墙

```bash
# 允许 SSH
sudo ufw allow 22/tcp

# 允许 AgentFlow 端口
sudo ufw allow 6767/tcp

# 允许 HAProxy 端口（如果使用）
sudo ufw allow 8080/tcp

# 启用防火墙
sudo ufw enable
```

#### 3. 配置 AgentFlow 服务

**在所有服务器上**:

```bash
sudo tee /etc/default/agentflow > /dev/null << 'EOF'
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_LOG_LEVEL=info
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=your_secure_password
AUTH_API_KEY_SECRET=your_shared_secret
EOF
```

```bash
sudo tee /etc/systemd/system/agentflow-master.service > /dev/null << 'EOF'
[Unit]
Description=AgentFlow Master
After=network-online.target

[Service]
Type=simple
User=agentflow
EnvironmentFile=/etc/default/agentflow
ExecStart=/usr/local/bin/agentflow-master
Restart=always

[Install]
WantedBy=multi-user.target
EOF
```

```bash
sudo systemctl daemon-reload
sudo systemctl start agentflow-master
sudo systemctl enable agentflow-master
```

#### 4. 配置负载均衡器

**选项 A: 云厂商 LB**

```bash
# AWS ALB
aws elbv2 create-load-balancer \
  --name agentflow-lb \
  --subnets subnet-1 subnet-2 \
  --security-groups sg-agentflow \
  --instances i-aaa i-bbb i-ccc \
  --listener-port-protocol HTTP Port-6767 \
  --health-check Protocol=HTTP Path=/health \
  TargetGroup-Name agentflow-tg
```

**选项 B: HAProxy**

```bash
# 安装 HAProxy
sudo apt-get install haproxy

# 配置 HAProxy
sudo tee /etc/haproxy/haproxy.cfg > /dev/null << 'EOF'
defaults
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

frontend agentflow_frontend
    bind *:80
    default_backend agentflow_backend

backend agentflow_backend
    balance roundrobin
    server server-a 10.0.1.10:6767 check
    server server-b 10.0.1.11:6767 check
    server server-c 10.0.1.12:6767 check
EOF

# 启动 HAProxy
sudo systemctl restart haproxy
sudo systemctl enable haproxy
```

#### 5. 验证部署

```bash
# 测试各个节点
curl http://10.0.1.10:6767/health
curl http://10.0.1.11:6767/health
curl http://10.0.1.12:6767/health

# 测试负载均衡
for i in {1..10}; do
  curl http://<lb-ip>/health
done
```

### 任务分发策略

由于当前版本任务不共享，需要外部系统分发任务：

**分发算法**:
1. 轮询（Round-robin）
2. 最少连接（Least connections）
3. 随机（Random）

**示例代码（Python）**:

```python
import random
import requests

INSTANCES = [
    "http://10.0.1.10:6767",
    "http://10.0.1.11:6767",
    "http://10.0.1.12:6767",
]

def distribute_task(title: str, description: str):
    """分发任务到可用实例"""
    # 随机选择一个实例
    instance = random.choice(INSTANCES)

    response = requests.post(
        f"{instance}/api/v1/tasks",
        json={
            "title": title,
            "description": description,
            "priority": "Medium"
        }
    )

    return response.json()["data"]["task_id"]

# 使用
task_id = distribute_task(
    title="分析代码",
    description="分析 AgentFlow 代码库"
)
```

---

## 共享存储方案

### 方案 1: Redis 任务队列（推荐）

**架构**:

```
┌─────────────┐   ┌─────────────┐
│ AgentFlow #1│   │ AgentFlow #2│
└──────┬──────┘   └──────┬──────┘
       │                 │
       └────────┬────────┘
                ▼
         ┌──────────────┐
         │    Redis     │
         │  (Queue)     │
         └──────────────┘
```

**实现**（未来 v3.1+）:

```rust
// 使用 Redis 作为任务队列
use redis::AsyncCommands;

async fn push_task_to_queue(task: &Task) -> Result<()> {
    let mut client = redis::Client::open("redis://localhost").await?;
    let _: () = client.lpush("agentflow:tasks", task.serialize()).await?;
    Ok(())
}

async fn pop_task_from_queue() -> Result<Task> {
    let mut client = redis::Client::open("redis://localhost").await?;
    let task = client.rpop("agentflow:tasks").await?;
    Ok(serde_json::from_str(&task)?)
}
```

### 方案 2: 共享记忆存储

**架构**:

```
┌─────────────┐   ┌─────────────┐
│ AgentFlow #1│   │ AgentFlow #2│
└──────┬──────┘   └──────┬──────┘
       │                 │
       └────────┬────────┘
                ▼
         ┌──────────────┐
         │  Qdrant       │
         │  (Vector DB)  │
         └──────────────┘
```

**实现**（未来 v3.1+）:

```rust
// 使用 Qdrant 作为共享记忆
use qdrant_client::prelude::*;

async fn store_memory_shared(memory: &Memory) -> Result<()> {
    let client = QdrantClient::from_url("http://qdrant:6333").await?;

    let client.insert_point("agentflow_memories")
        .insert_point(None)
        .vectors(&[memory.vector])
        .await?;

    Ok(())
}

async fn search_memory_shared(query: &str) -> Result<Vec<Memory>> {
    let client = QdrantClient::from_url("http://qdrant:6333").await?;

    let results = client.search_point("agentflow_memories")
        .vector(query)
        .limit(10)
        .execute()
        .await?;

    Ok(results.into_iter().collect())
}
```

### 方案 3: 数据库集群（生产级）

**架构**:

```
┌────────────────────────────────────────┐
│            PostgreSQL 集群              │
│  ┌────────┐       ┌────────┐               │
│  │ Primary │ ←──→ │ Standby │               │
│  └────────┘       └────────┘               │
│                                             │
│  ┌────────┐       ┌────────┐               │
│  │ Replica │       │ Replica │               │
│  └────────┘       └────────┘               │
└────────────────────────────────────────┘
         ↕               ↕               ↕
    AgentFlow     AgentFlow       AgentFlow
```

---

## 升级路线图

### v3.0（当前版本）

**特点**:
- ✅ 单进程架构
- ✅ 本地 SQLite
- ✅ 简单部署

**限制**:
- ⚠️ 无任务共享
- ⚠️ 无记忆共享
- ⚠️ 单点故障

### v3.1（计划中）

**新增功能**:
- ✅ Redis 任务队列
- ✅ Qdrant 向量数据库
- ✅ 任务分发器
- ✅ 健康检查改进

**部署模式**:
- ✅ 单机多实例（共享队列）
- ✅ 多机多实例（共享队列）
- ⚠️ 仍需外部负载均衡

**升级路径**:
```bash
# 1. 部署 Redis
docker run -d --name redis -p 6379:6379 redis:7

# 2. 部署 Qdrant
docker run -d --name qdrant -p 6333:6333 qdrant/qdrant

# 3. 启动 AgentFlow（使用共享存储）
export AGENTFLOW_QUEUE_TYPE=redis
export AGENTFLOW_QUEUE_URL=redis://localhost:6379
export AGENTFLOW_MEMORY_TYPE=qdrant
export AGENTFLOW_MEMORY_URL=http://localhost:6333
./target/release/agentflow-master
```

### v3.2+（规划中）

**新增功能**:
- ✅ 内置负载均衡器
- ✅ 自动服务发现
- ✅ 故障自动转移
- ✅ Prometheus metrics

**部署模式**:
- ✅ 多机多实例（开箱即用）
- ✅ 自动集群管理
- ✅ 滚动升级

---

## 最小可用集群示例

### 示例架构

**3 个节点，单机多实例**:

```
┌────────────────────────────────────────────┐
│                   Nginx                  │
│            (Load Balancer)                 │
└──────────────────┬─────────────────────────┘
                   │
        ┌──────────┴──────────┐
        ▼                      ▼
    ┌──────┐              ┌──────┐
    │ #1   │              │ #3   │
    │6767 │              │6769 │
    └──────┘              └──────┘
```

**部署脚本**:

```bash
#!/bin/bash

# 1. 创建配置文件
cat > /etc/default/agentflow-inst1 << 'EOF'
AGENTFLOW_SERVER_PORT=6767
EOF

cat > /etc/default/agentflow-inst2 << 'EOF'
AGENTFLOW_SERVER_PORT=6768
EOF

cat > /etc/default/agentflow-inst3 << 'EOF'
AGENTFLOW_SERVER_PORT=6769
EOF

# 2. 配置 Nginx
cat > /etc/nginx/sites-available/agentflow-cluster << 'EOF'
upstream agentflow_cluster {
    least_conn;
    server 127.0.0.1:6767;
    server 127.0.0.1:6768;
    server 127.0.0.1:6769;
}

server {
    listen 80;
    location / {
        proxy_pass http://agentflow_cluster;
    }
}
EOF

# 3. 启动实例
sudo systemctl start agentflow-master@inst1
sudo systemctl start agentflow-master@inst2
sudo systemctl start agentflow@inst3

# 4. 启动 Nginx
sudo nginx -s reload
```

**验证脚本**:

```bash
#!/bin/bash

echo "验证集群部署..."

# 测试各个实例
for port in 6767 6768 6769; do
    echo "测试端口 $port..."
    curl http://localhost:$port/health
done

# 测试负载均衡
echo -e "\n测试负载均衡..."
for i in {1..6}; do
    echo "请求 $i:"
    curl http://localhost/80/health
done

echo -e "\n✓ 集群部署完成"
```

---

## 配置示例

### Nginx 高级配置

```nginx
upstream agentflow_cluster {
    # 负载均衡算法
    least_conn;

    # 服务器列表
    server 10.0.1.10:6767 weight=1 max_fails=3 fail_timeout=30s;
    server 10.0.1.11:6767 weight=1 max_fails=3 fail_timeout=30s;
    server 10.0.1.12:6767 weight=1 max_fails=3 fail_timeout=30s backup;

    # keepalive 连接
    keepalive 32;
}

server {
    listen 80;
    server_name agentflow.example.com;

    # 访问日志
    access_log /var/log/nginx/agentflow-access.log;
    error_log /var/log/nginx/agentflow-error.log;

    # 限流
    limit_req_zone=agentflow burst=20 nodelay;
    limit_req_status 429;

    # 代理配置
    location / {
        proxy_pass http://agentflow_cluster;

        # 代理头
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # 超时
        proxy_read_timeout 300s;
        proxy_connect_timeout 60s;

        # WebSocket
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 健康检查端点
    location /health {
        proxy_pass http://agentflow_cluster/health;
        access_log off;
    }
}
```

---

## 故障排查

### 1. 节点下线

**症状**: 负载均衡器将流量发送到下线节点

**解决**:

```bash
# 1. 检查节点健康状态
curl http://10.0.1.10:6767/health

# 2. 从 Nginx upstream 中移除节点
# 编辑 Nginx 配置，注释掉故障节点
sudo nano /etc/nginx/nginx.conf
# sudo nginx -s reload

# 3. 节点恢复后重新加入
# 取消注释
sudo nginx -s reload
```

### 2. 任务不均衡

**症状**: 某个节点负载过高，其他节点空闲

**原因**:
- 负载均衡算法问题
- 节点配置不一致
- 长连接占用

**解决**:

```bash
# 1. 使用 least_conn 算法
upstream agentflow_cluster {
    least_conn;  # 改为 least_conn
    ...
}

# 2. 调整 keepalive
upstream agentflow_cluster {
    keepalive 16;  # 减少 keepalive 连接数
    ...
}

# 3. 重启 Nginx
sudo nginx -s reload
```

### 3. 数据不一致

**症状**: 不同节点返回不同的数据

**原因**: 当前架构使用本地 SQLite，数据天然不共享

**解决**:

**短期方案**:
- 使用外部分发器（确保任务在同一节点执行）
- 为每个客户端分配固定的节点

**长期方案**:
- 升级到 v3.1，使用 Redis 任务队列
- 升级到 v3.2，使用共享数据库

---

## 最佳实践

### 1. 监控

**关键指标**:
- 节点健康状态
- 请求响应时间
- 错误率
- 任务队列长度

**监控工具**:
- Prometheus + Grafana
- ELK Stack（日志聚合）
- Zabbix

### 2. 备份

**数据备份**:
- 定期备份 SQLite 数据库
- 备份配置文件（`/etc/default/agentflow`）
- 备份 Nginx 配置

**备份脚本**:

```bash
#!/bin/bash
# backup-agentflow.sh

BACKUP_DIR="/backup/agentflow"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR/$DATE"

# 备份数据库
cp -r /var/lib/agentflow "$BACKUP_DIR/$DATE/data"

# 备份配置
cp /etc/default/agentflow "$BACKUP_DIR/$DATE/config"

# 保留最近 7 天的备份
find "$BACKUP_DIR" -type d -mtime +7 -exec rm -rf {} \;

echo "备份完成: $DATE"
```

### 3. 安全

**网络安全**:
- 使用 VPN 或内网隧道
- 配置防火墙规则
- 使用 TLS/SSL

**访问控制**:
- 启用认证（`AUTH_ENABLED=true`）
- 使用强密码（16+ 字符）
- 定期轮换 API Key Secret

---

**文档版本**: 1.0.0
**最后更新**: 2026-01-28
