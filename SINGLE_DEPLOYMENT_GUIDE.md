# AgentFlow 单机部署指南

本文档说明如何将 AgentFlow Master 部署为单机应用，Rust 后端直接提供 React 前端和 API 服务。

## 架构概述

```
┌─────────────────────────────────────────────────────────┐
│              AgentFlow Master (Rust)                    │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │           API 服务 (Axum)                          │ │
│  │  - /api/v1/tasks (任务管理)                       │ │
│  │  - /api/v1/memory (记忆管理)                      │ │
│  │  - /api/v1/login (登录认证)                       │ │
│  │  - /ws/task (WebSocket)                           │ │
│  └───────────────────────────────────────────────────┘ │
│                         ↓                               │
│  ┌───────────────────────────────────────────────────┐ │
│  │        静态文件服务 (React Dashboard)             │ │
│  │  - 所有非 API 路径返回 index.html                 │ │
│  │  - API 请求使用相对路径                          │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │           认证中间件                               │ │
│  │  - Session 管理                                    │ │
│  │  - 用户名密码验证                                  │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## 功能特性

### ✅ 已实现

1. **Rust 后端提供 React 静态文件**
   - 自动 SPA 路由支持
   - 从 `dashboard/dist/` 目录提供文件
   - 所有非 API 路径返回 `index.html`

2. **简单的用户名密码认证**
   - Session 管理
   - Bearer Token 认证
   - 可通过环境变量配置

3. **相对路径 API 调用**
   - 生产环境自动使用相对路径
   - 开发环境仍可使用绝对路径

## 环境变量配置

创建 `.env` 文件在项目根目录或设置环境变量：

```bash
# AgentFlow Master 配置
AGENTFLOW_SERVER_ADDR=0.0.0.0
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_DATABASE_URL=sqlite://agentflow.db
AGENTFLOW_LOG_LEVEL=info

# 认证配置（可选）
AUTH_ENABLED=true              # 是否启用认证（默认: false）
AUTH_USERNAME=admin            # 用户名（默认: admin）
AUTH_PASSWORD=admin            # 密码（默认: admin）
AUTH_SESSION_TTL=86400         # Session 过期时间，秒（默认: 86400 = 24小时）

# Worker 配置
AGENTFLOW_MAX_CONCURRENT_TASKS=10
AGENTFLOW_TASK_TIMEOUT=300
```

## 构建和部署

### 1. 构建 React 前端

```bash
cd dashboard
npm install
npm run build
```

构建产物将生成在 `dashboard/dist/` 目录。

### 2. 构建 Rust 后端

```bash
cd rust
cargo build --release
```

生成的二进制文件：`target/release/agentflow-master`

### 3. 配置环境变量

创建 `.env` 文件（可选）：

```bash
# .env
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=your_secure_password
```

### 4. 运行服务器

```bash
# 方式 1: 使用默认配置
./target/release/agentflow-master

# 方式 2: 加载 .env 文件
./target/release/agentflow-master

# 方式 3: 指定端口
./target/release/agentflow-master --port 8080

# 方式 4: 指定日志级别
./target/release/agentflow-master --log-level debug
```

服务器启动后，访问：

- **Dashboard**: http://localhost:6767
- **API**: http://localhost:6767/api/v1/...
- **WebSocket**: ws://localhost:6767/ws/task/...

## 认证流程

### 1. 登录

```bash
curl -X POST http://localhost:6767/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

响应：

```json
{
  "success": true,
  "message": "登录成功",
  "session_id": "uuid-string"
}
```

### 2. 访问受保护的 API

使用返回的 `session_id` 作为 Bearer Token：

```bash
curl http://localhost:6767/api/v1/tasks \
  -H "Authorization: Bearer uuid-string"
```

### 3. 前端集成

在 React 应用中，登录后存储 session_id 并在所有后续请求中携带：

```typescript
// 登录
const response = await fetch('/api/v1/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username, password }),
});
const { session_id } = await response.json();

// 存储到 localStorage
localStorage.setItem('session_id', session_id);

// 后续请求
const tasks = await fetch('/api/v1/tasks', {
  headers: {
    'Authorization': `Bearer ${session_id}`
  }
});
```

## 开发模式

### 开发环境配置

在开发环境中，可以单独运行前端开发服务器：

```bash
# 终端 1: 启动前端开发服务器
cd dashboard
npm run dev

# 终端 2: 启动后端服务器
cd rust
cargo run
```

前端开发服务器将代理 API 请求到后端。

### 生产环境配置

在生产环境中，只需要运行 Rust 服务器：

```bash
./target/release/agentflow-master
```

Rust 服务器将同时提供 API 和静态文件。

## 测试清单

### 基本功能测试

- [ ] 服务器成功启动
- [ ] 访问 http://localhost:6767 显示 React Dashboard
- [ ] 前端能正确加载所有资源（JS/CSS）

### API 测试

- [ ] 健康检查：`GET /health`
- [ ] 健康检查：`GET /api/v1/health`
- [ ] 登录：`POST /api/v1/login`
- [ ] 任务列表：`GET /api/v1/tasks`
- [ ] 创建任务：`POST /api/v1/tasks`

### 认证测试

- [ ] 未启用认证时，所有 API 可直接访问
- [ ] 启用认证后，未登录访问受保护 API 返回 401
- [ ] 登录成功返回 session_id
- [ ] 使用 session_id 可访问受保护 API
- [ ] Session 过期后返回 401

### SPA 路由测试

- [ ] 直接访问 http://localhost:6767 返回 index.html
- [ ] 刷新页面（F5）保持当前路由
- [ ] 前端路由正常工作

### WebSocket 测试

- [ ] WebSocket 连接成功
- [ ] 能接收实时消息
- [ ] 连接断开后能自动重连

## 故障排除

### 问题 1: 404 Not Found

**症状**: 访问根路径返回 404

**原因**: React 前端未构建或 `dashboard/dist/` 目录不存在

**解决**:

```bash
cd dashboard
npm run build
```

### 问题 2: API 请求失败

**症状**: 前端加载但 API 请求失败

**原因**:
1. 后端未启动
2. 端口配置错误
3. CORS 问题

**解决**:

```bash
# 检查后端是否运行
curl http://localhost:6767/health

# 检查端口
lsof -i :6767
```

### 问题 3: 401 Unauthorized

**症状**: API 返回 401 错误

**原因**:
1. 认证已启用但未登录
2. Session 过期
3. Token 格式错误

**解决**:

```bash
# 重新登录
curl -X POST http://localhost:6767/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# 检查认证是否启用
curl http://localhost:6767/health | jq '.auth_enabled'
```

### 问题 4: WebSocket 连接失败

**症状**: WebSocket 无法连接

**原因**:
1. WebSocket 路由配置错误
2. 认证失败
3. 防火墙阻止

**解决**:

```bash
# 测试 WebSocket 连接
wscat -c ws://localhost:6767/ws/task/123

# 检查认证
curl -H "Authorization: Bearer your-session-id" \
  http://localhost:6767/api/v1/tasks
```

## 性能优化

### 静态文件缓存

Rust 服务器可以配置静态文件缓存：

```rust
use tower_http::services::ServeDir;

let spa_service = ServeDir::new("dashboard/dist")
    .precompressed_gzip(true)
    .precompressed_br(true)
    .fallback(ServeFile::new("dashboard/dist/index.html"));
```

### 压缩

启用 HTTP 响应压缩：

```rust
use tower_http::compression::CompressionLayer;

let app = Router::new()
    .layer(CompressionLayer::new())
    // ...
```

### 并发配置

调整工作线程数：

```bash
# 设置 Tokio 运行时线程数
TOKIO_WORKER_THREADS=4 ./target/release/agentflow-master
```

## 安全建议

### 生产环境配置

1. **启用认证**:
   ```bash
   AUTH_ENABLED=true
   AUTH_USERNAME=your_secure_username
   AUTH_PASSWORD=your_very_secure_password
   ```

2. **使用 HTTPS**:
   在反向代理（如 Nginx）后面部署并启用 SSL

3. **限制访问**:
   配置防火墙规则，只允许必要的端口

4. **定期更新**:
   及时更新依赖和安全补丁

### 密码管理

- 使用强密码（至少 12 位，包含大小写字母、数字、特殊字符）
- 定期更换密码
- 不要在代码中硬编码密码
- 使用环境变量或密钥管理服务

## 监控和日志

### 日志级别

```bash
# 开发环境
./agentflow-master --log-level debug

# 生产环境
./agentflow-master --log-level info

# 问题排查
./agentflow-master --log-level trace
```

### 日志输出

日志输出到 stdout，可以重定向到文件：

```bash
./agentflow-master > agentflow.log 2>&1
```

或使用 systemd/journald 管理：

```bash
journalctl -u agentflow-master -f
```

## 部署示例

### Systemd 服务

创建 `/etc/systemd/system/agentflow-master.service`:

```ini
[Unit]
Description=AgentFlow Master Server
After=network.target

[Service]
Type=simple
User=agentflow
Group=agentflow
WorkingDirectory=/opt/agentflow
Environment="AUTH_ENABLED=true"
Environment="AUTH_USERNAME=admin"
Environment="AUTH_PASSWORD=your_password"
ExecStart=/opt/agentflow/agentflow-master
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable agentflow-master
sudo systemctl start agentflow-master
sudo systemctl status agentflow-master
```

### Docker 部署

创建 `Dockerfile`:

```dockerfile
FROM rust:1.75 as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/target/release/agentflow-master /usr/local/bin/
COPY --from=builder /app/dashboard/dist /var/www/agentflow
EXPOSE 6767
CMD ["agentflow-master"]
```

构建和运行：

```bash
docker build -t agentflow-master .
docker run -p 6767:6767 -e AUTH_ENABLED=true agentflow-master
```

## 总结

通过本部署方案，AgentFlow Master 可以：

✅ **单二进制部署**: 只需运行一个可执行文件
✅ **内置静态文件服务**: 无需额外配置 Web 服务器
✅ **简单的认证机制**: 用户名密码 + Session 管理
✅ **开发友好**: 开发环境可独立运行前后端
✅ **生产就绪**: 支持环境变量配置、日志、监控

---

**最后更新**: 2026-01-28
**文档版本**: 1.0.0
