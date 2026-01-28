# AgentFlow 生产部署指南

**目标**: 将 AgentFlow 作为生产服务部署，确保稳定、安全、可观测

**适用场景**:
- 单机生产部署
- 服务化部署（systemd/supervisor）
- 容器化部署（Docker/K8s）
- 高可用部署

---

## 📋 目录

1. [部署方式对比](#部署方式对比)
2. [本地二进制部署](#本地二进制部署)
3. [Docker 部署](#docker-部署)
4. [Kubernetes 部署](#kubernetes-部署)
5. [安全加固](#安全加固)
6. [运维与可观测性](#运维与可观测性)
7. [故障排查](#故障排查)

---

## 部署方式对比

| 部署方式 | 复杂度 | 可扩展性 | 运维成本 | 推荐场景 |
|---------|--------|----------|----------|----------|
| **本地二进制** | ⭐⭐ | ⭐ | 低 | 个人项目、小型团队 |
| **Docker** | ⭐⭐⭐ | ⭐⭐⭐ | 中 | 中大型项目 |
| **K8s** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 高 | 大型生产环境 |

---

## 本地二进制部署

### 1. 下载预编译二进制

```bash
# 从 GitHub Releases 下载
wget https://github.com/MoSiYuan/AgentFlow/releases/latest/download/agentflow-master-linux-amd64
chmod +x agentflow-master-linux-amd64
```

### 2. 创建部署目录

```bash
# 创建专用用户
sudo useradd -r -s /bin/bash agentflow
sudo mkdir -p /opt/agentflow
sudo chown agentflow:agentflow /opt/agentflow

# 复制二进制
sudo cp agentflow-master-linux-amd64 /opt/agentflow/bin/
sudo chmod +x /opt/agentflow/bin/agentflow-master
```

### 3. 配置环境变量

```bash
# 创建环境配置文件
sudo tee /etc/default/agentflow > /dev/null << 'EOF'
# AgentFlow 环境配置

# 数据目录
AGENTFLOW_DATA_DIR=/var/lib/agentflow
AGENTFLOW_LOG_DIR=/var/log/agentflow

# 服务配置
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_LOG_LEVEL=info

# 认证配置
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=your_secure_password_here
AUTH_API_KEY_SECRET=your_api_key_secret_at_least_32_chars

# 资源限制
RUST_LOG=info
RUST_BACKTRACE=1
EOF
```

### 4. 创建 systemd service

```bash
sudo tee /etc/systemd/system/agentflow-master.service > /dev/null << 'EOF'
[Unit]
Description=AgentFlow Master Server
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
User=agentflow
Group=agentflow
EnvironmentFile=/etc/default/agentflow

# 工作目录
WorkingDirectory=/opt/agentflow

# 二进制路径
ExecStart=/opt/agentflow/bin/agentflow-master
ExecReload=/bin/kill -HUP $MAINPID

# 重启策略
Restart=always
RestartSec=5
StartLimitBurst=3
StartLimitInterval=60

# 资源限制
MemoryMax=100M
CPUQuota=50%
LimitNOFILE=65536
TimeoutStartSec=300
TimeoutStopSec=30

# 安全加固
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/agentflow
ReadWritePaths=/var/log/agentflow

# 日志
StandardOutput=journal
StandardError=journal
SyslogIdentifier=agentflow-master

[Install]
WantedBy=multi-user.target
EOF
```

### 5. 配置日志轮转

```bash
sudo tee /etc/logrotate.d/agentflow > /dev/null << 'EOF'
/var/log/agentflow/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 agentflow agentflow
    sharedscripts
    postrotate
        systemctl reload agentflow-master >/dev/null 2>&1 || true
    endscript
}
EOF
```

### 6. 启动服务

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start agentflow-master

# 设置开机自启
sudo systemctl enable agentflow-master

# 查看状态
sudo systemctl status agentflow-master

# 查看日志
sudo journalctl -u agentflow-master -f
```

---

## Docker 部署

### 1. Dockerfile（多阶段构建）

```dockerfile
# 构建阶段
FROM rust:1.75-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache musl-dev sqlite-dev gcc

# 设置工作目录
WORKDIR /build

# 复制 Cargo 配置
COPY rust/Cargo.toml rust/Cargo.lock ./
COPY rust/agentflow-core ./agentflow-core
COPY rust/agentflow-master ./agentflow-master

# 构建（使用预编译依赖加速）
RUN cargo build --release && \
    mv target/release/agentflow-master /tmp/ && \
    cargo clean

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    sqlite \
    ca-certificates \
    && addgroup -g agentflow && \
    adduser -D -u 1000 -G agentflow agentflow

# 创建目录
RUN mkdir -p /var/lib/agentflow /var/log/agentflow && \
    chown -R agentflow:agentflow /var/lib/agentflow /var/log/agentflow

# 复制二进制
COPY --from=builder /tmp/agentflow-master /opt/agentflow/bin/

# 创建非 root 用户
USER agentflow

# 工作目录
WORKDIR /opt/agentflow

# 暴露端口
EXPOSE 6767 8849

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:6767/health || exit 1

# 启动服务
CMD ["/opt/agentflow/bin/agentflow-master"]
```

### 2. 构建镜像

```bash
# 构建镜像
docker build -t agentflow-master:v1.0 .

# 查看镜像
docker images | grep agentflow
```

### 3. Docker Compose 部署

```yaml
version: '3.8'

services:
  agentflow-master:
    image: agentflow-master:v1.0
    container_name: agentflow-master
    restart: unless-stopped

    environment:
      - AGENTFLOW_SERVER_PORT=6767
      - AGENTFLOW_LOG_LEVEL=info
      - AUTH_ENABLED=true
      - AUTH_USERNAME=admin
      - AUTH_PASSWORD=${AUTH_PASSWORD:-admin}
      - AUTH_API_KEY_SECRET=${AUTH_API_KEY_SECRET}

    ports:
      - "6767:6767"
      - "8849:8849"

    volumes:
      # 数据持久化
      - agentflow-data:/var/lib/agentflow
      - agentflow-logs:/var/log/agentflow

      # 配置文件挂载（可选）
      # - ./config/agentflow.env:/etc/default/agentflow:ro

    # 资源限制
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 100M
        reservations:
          cpus: '0.25'
          memory: 50M

    # 健康检查
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:6767/health"]
      interval: 30s
      timeout: 3s
      retries: 3

    # 日志配置
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  agentflow-data:
  agentflow-logs:
```

### 4. 启动

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f agentflow-master

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

---

## Kubernetes 部署

### 1. Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: agentflow
```

### 2. ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentflow-config
  namespace: agentflow
data:
  AGENTFLOW_SERVER_PORT: "6767"
  AGENTFLOW_LOG_LEVEL: "info"
  AUTH_ENABLED: "true"
```

### 3. Secret

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: agentflow-secret
  namespace: agentflow
type: Opaque
stringData:
  AUTH_PASSWORD: YWRtaW4=  # admin (base64)
  AUTH_API_KEY_SECRET: eW91cl9fc2VyeV9fc2VyX2F0X2xlYXN0XzMyMGNoYXJz  # your_secret (base64)
```

### 4. Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentflow-master
  namespace: agentflow
spec:
  replicas: 2

  selector:
    matchLabels:
      app: agentflow-master

  template:
    metadata:
      labels:
        app: agentflow-master
        version: v1.0

    spec:
      containers:
      - name: agentflow-master
        image: agentflow-master:v1.0
        imagePullPolicy: IfNotPresent

        ports:
        - containerPort: 6767
          name: http
        - containerPort: 8849
          name: websocket

        env:
        - name: AGENTFLOW_SERVER_PORT
          valueFrom:
            configMapKeyRef:
              name: agentflow-config
              key: AGENTFLOW_SERVER_PORT

        - name: AUTH_USERNAME
          value: "admin"

        - name: AUTH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: agentflow-secret
              key: AUTH_PASSWORD

        - name: AUTH_API_KEY_SECRET
          valueFrom:
            secretKeyRef:
              name: agentflow-secret
              key: AUTH_API_KEY_SECRET

        resources:
          requests:
            cpu: 100m
            memory: 50M
          limits:
            cpu: 500m
            memory: 100M

        livenessProbe:
          httpGet:
            path: /health
            port: 6767
          initialDelaySeconds: 10
          periodSeconds: 30
          timeoutSeconds: 3
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /health
            port: 6767
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 2
          failureThreshold: 2

        volumeMounts:
        - name: data
          mountPath: /var/lib/agentflow
        - name: logs
          mountPath: /var/log/agentflow

      volumes:
      - name: data
        emptyDir: {}
      - name: logs
        emptyDir: {}
```

### 5. Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: agentflow-master
  namespace: agentflow
spec:
  selector:
    app: agentflow-master

  ports:
  - port: 6767
    targetPort: 6767
    name: http
  - port: 8849
    targetPort: 8849
    name: websocket

  type: ClusterIP
```

### 6. Ingress（可选）

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agentflow-ingress
  namespace: agentflow
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  ingressClassName: nginx
  rules:
  - host: agentflow.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: agentflow-master
            port:
              number: 6767
  tls:
  - hosts:
    - agentflow.example.com
    secretName: agentflow-tls
```

### 7. 部署到 K8s

```bash
# 应用所有配置
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml

# 查看状态
kubectl get pods -n agentflow
kubectl get svc -n agentflow
kubectl get ingress -n agentflow

# 查看日志
kubectl logs -f deployment/agentflow-master -n agentflow

# 扩容
kubectl scale deployment/agentflow-master --replicas=3 -n agentflow
```

---

## 安全加固

### 1. 反向代理（Nginx）

```nginx
# /etc/nginx/sites-available/agentflow
server {
    listen 443 ssl http2;
    server_name agentflow.example.com;

    # SSL 证书
    ssl_certificate /etc/ssl/certs/agentflow.crt;
    ssl_certificate_key /etc/ssl/certs/agentflow.key;

    # 安全头
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # 限流
    limit_req_zone=agentflow burst=20 nodelay;
    limit_req_status 429;

    # 代理到 AgentFlow
    location / {
        proxy_pass http://localhost:6767;

        # 代理头
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时
        proxy_read_timeout 300s;
        proxy_connect_timeout 60s;

        # 最大请求体
        client_max_body_size 10M;
    }

    # WebSocket 代理
    location /ws/ {
        proxy_pass http://localhost:8849;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

# 限流配置
http {
    limit_req_zone $binary_remote_addr zone=agentflow:10m rate=10r/s;
}
```

### 2. 防火墙配置

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

### 3. 访问控制

**基于 IP 的访问控制**（Nginx）:

```nginx
# 仅允许特定 IP
location /api/v1/ {
    allow 192.168.1.0/24;
    deny all;

    proxy_pass http://localhost:6767;
}
```

**基于认证的访问控制**:

```bash
# 启用认证
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=strong_password_here
```

---

## 运维与可观测性

### 1. 健康检查

```bash
# 基础健康检查
curl http://localhost:6767/health

# 预期返回
{
  "status": "ok",
  "version": "1.0.0",
  "auth_enabled": true
}
```

### 2. 日志管理

#### systemd journal

```bash
# 实时查看
sudo journalctl -u agentflow-master -f

# 查看最近 100 行
sudo journalctl -u agentflow-master -n 100

# 查看启动错误
sudo journalctl -u agentflow-master -b
```

#### 日志文件

```bash
# 查看 AgentFlow 日志
sudo tail -f /var/log/agentflow/agentflow.log
```

### 3. 性能监控（未来）

**计划添加 Prometheus metrics**:

```rust
// 暴露 /metrics 端点
#[derive(prometheus_endpoint::PrometheusEndpoint)]
struct AgentFlowMetrics {
    tasks_total: prometheus::IntCounter,
    tasks_completed: prometheus::IntCounter,
    tasks_failed: prometheus::IntCounter,
}
```

---

## 故障排查

### 1. 服务启动失败

**检查步骤**:

```bash
# 1. 查看服务状态
sudo systemctl status agentflow-master

# 2. 查看启动日志
sudo journalctl -u agentflow-master -n 50 --no-pager

# 3. 检查端口占用
sudo netstat -tulpn | grep 6767

# 4. 检查权限
sudo ls -la /opt/agentflow/bin/
sudo ls -la /var/lib/agentflow
```

### 2. 性能问题

**检查资源使用**:

```bash
# CPU 和内存
top -p $(pgrep agentflow-master)

# 磀查文件描述符
lsof -p $(pgrep agentflow-master) | wc -l

# 检查线程
ps -T $(pgrep agentflow-master)
```

### 3. 连接超时

**可能原因**:
- 防火墙阻止
- 反向代理配置错误
- API Key 过期

**排查**:

```bash
# 测试本地连接
curl -v http://localhost:6767/health

# 测试认证
curl -X POST http://localhost:6767/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# 检查 API Key
echo "sk_$(date +%s)_$(openssl rand -hex 16)" | \
  curl -H "Authorization: Bearer $(stdin)" \
  http://localhost:6767/health
```

---

## 部署清单

### 部署前检查

- [ ] Rust 工具链已安装
- [ ] Claude CLI 已配置
- [ ] 防火墙规则已配置
- [ ] SSL 证书已准备（如使用 HTTPS）

### 部署后验证

- [ ] 服务成功启动
- [ ] 健康检查正常
- [ ] 日志正常输出
- [ ] API 认证测试通过
- [ ] WebSocket 连接正常
- [ ] 开机自启已配置

---

**文档版本**: 1.0.0
**最后更新**: 2026-01-28
