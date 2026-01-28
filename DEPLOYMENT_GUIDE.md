# AgentFlow 部署指南

**更新时间**: 2026-01-28
**版本**: v1.0.0

---

## 📦 编译产物位置

### 1. Rust Master 服务器

**路径**: `rust/target/release/agentflow-master`
**大小**: 8.1 MB
**功能**:
- HTTP API 服务
- WebSocket 实时通信
- 双认证系统（Session + API Key）
- 任务调度和管理

**编译命令**:
```bash
cd rust
cargo build --release --bin agentflow-master
```

### 2. Rust Worker 节点

**路径**: `rust/target/release/agentflow-worker`
**大小**: ~7 MB
**功能**:
- 任务执行
- 心跳上报
- 状态管理

**编译命令**:
```bash
cd rust
cargo build --release --bin agentflow-worker
```

### 3. React Dashboard

**路径**: `dashboard/dist/`
**大小**: ~2 MB (压缩后)
**功能**:
- Web 管理界面
- 任务监控
- 集群管理
- 日志查看

**构建命令**:
```bash
cd dashboard
npm install
npm run build
```

---

## 🚀 快速部署方式

### 方式 1: 一键安装脚本（推荐）

```bash
cd deployment/package
./install.sh
```

**自动完成**:
1. ✅ 检查系统环境（Node.js, Rust, Git）
2. ✅ 编译 Master 服务器
3. ✅ 构建 Dashboard
4. ✅ 创建配置文件 (`.env`)
5. ✅ 创建启动脚本 (`start.sh`)

**启动服务**:
```bash
./start.sh
```

---

### 方式 2: AgentFlow Helper（交互式向导）

```bash
# 安装 Helper
npm install -g @agentflow/helper

# 运行初始化向导
agentflow-helper init

# 或直接使用 npx
npx agentflow-helper init
```

**Helper 功能**:
- 🔧 交互式安装向导
- 🏥 环境健康检查
- ⚙️ 配置管理
- 🚀 服务管理

**命令列表**:
```bash
agentflow-helper init              # 初始化向导
agentflow-helper doctor            # 环境检查
agentflow-helper install all       # 安装所有组件
agentflow-helper config --list     # 查看配置
agentflow-helper server start      # 启动服务
```

---

### 方式 3: 手动部署

#### 步骤 1: 编译 Master

```bash
cd rust
cargo build --release --bin agentflow-master
```

#### 步骤 2: 构建 Dashboard

```bash
cd dashboard
npm install
npm run build
```

#### 步骤 3: 创建配置文件

```bash
cat > .env << 'EOF'
# 认证配置
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=admin
AUTH_SESSION_TTL=86400

# 服务器配置
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_LOG_LEVEL=info
EOF
```

#### 步骤 4: 启动服务

```bash
# 加载环境变量并启动
export $(cat .env | grep -v '^#' | xargs)
./rust/target/release/agentflow-master
```

---

### 方式 4: Docker 部署

```bash
# 构建镜像
docker build -t agentflow-master:latest .

# 运行容器
docker run -d \
  --name agentflow-master \
  -p 6767:6767 \
  -p 8849:8849 \
  -e AUTH_ENABLED=true \
  -e AUTH_USERNAME=admin \
  -e AUTH_PASSWORD=admin \
  agentflow-master:latest
```

---

## 🔧 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 | 必需 |
|--------|------|--------|------|
| `AUTH_ENABLED` | 是否启用认证 | false | 否 |
| `AUTH_USERNAME` | 管理员用户名 | admin | 是 |
| `AUTH_PASSWORD` | 管理员密码 | admin | 是 |
| `AUTH_SESSION_TTL` | Session 有效期（秒） | 86400 | 否 |
| `AUTH_API_KEY_SECRET` | API Key 密钥 | - | 是 |
| `AGENTFLOW_SERVER_PORT` | Master 服务端口 | 6767 | 否 |
| `AGENTFLOW_LOG_LEVEL` | 日志级别 | info | 否 |

### 认证系统

AgentFlow 支持双认证方式：

#### 1. 用户 Session 认证（Dashboard）
- 用户名/密码登录
- Session ID 存储在 localStorage
- 自动携带认证头

#### 2. API Key 认证（Master 之间）
- 格式: `sk_{timestamp}_{signature}`
- 签名算法: HMAC-SHA256
- 有效期: 5 分钟

---

## 📍 服务地址

部署成功后，可访问以下地址：

- **Dashboard**: http://localhost:6767
- **API Root**: http://localhost:6767/api/v1
- **WebSocket**: ws://localhost:8849
- **Health Check**: http://localhost:6767/health

---

## 🔍 验证安装

### 1. 检查 Master 状态

```bash
curl http://localhost:6767/health
```

**预期输出**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "auth_enabled": true
}
```

### 2. 测试登录

```bash
curl -X POST http://localhost:6767/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

**预期输出**:
```json
{
  "success": true,
  "message": "登录成功",
  "session_id": "uuid-string"
}
```

### 3. 检查进程

```bash
ps aux | grep agentflow-master
```

### 4. 检查端口

```bash
lsof -i :6767
lsof -i :8849
```

---

## 🛠️ 常见问题

### Q1: 编译失败 - 缺少 Rust

**错误**: `command not found: cargo`

**解决**:
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

### Q2: 端口被占用

**错误**: `Address already in use (os error 48)`

**解决**:
```bash
# 查看占用进程
lsof -i :6767

# 杀死进程
kill -9 <PID>

# 或更换端口
export AGENTFLOW_SERVER_PORT=8080
```

### Q3: Dashboard 无法访问

**可能原因**: Dashboard 未构建

**解决**:
```bash
cd dashboard
npm install
npm run build
```

### Q4: 认证失败

**可能原因**: Session 过期或 API Key 无效

**解决**:
- 刷新页面重新登录
- 检查 `AUTH_API_KEY_SECRET` 是否配置
- 查看 Master 日志

---

## 📊 性能优化

### 1. 生产环境配置

```bash
# 使用 Release 模式编译
cargo build --release

# 启用优化
export RUSTFLAGS="-C target-cpu=native"

# 减少二进制大小
export RUSTFLAGS="-C target-cpu=native -C opt-level=z -C link-arg=-s"
```

### 2. 系统资源

**最低配置**:
- CPU: 2 核
- 内存: 2 GB
- 磁盘: 100 MB

**推荐配置**:
- CPU: 4+ 核
- 内存: 4+ GB
- 磁盘: 1 GB

### 3. 并发优化

```bash
# 调整连接数
export AGENTFLOW_MAX_CONNECTIONS=1000

# 调整 Worker 数量
export AGENTFLOW_MAX_WORKERS=50
```

---

## 🔒 安全建议

### 1. 生产环境

```bash
# 使用强密码
export AUTH_PASSWORD="$(openssl rand -base64 24)"

# 使用强 API Key Secret
export AUTH_API_KEY_SECRET="$(openssl rand -hex 32)"

# 启用 HTTPS（通过反向代理）
```

### 2. 配置防火墙

```bash
# 仅允许本地访问
ufw allow from 127.0.0.1 to any port 6767

# 或限制特定 IP
ufw allow from 192.168.1.0/24 to any port 6767
```

### 3. 使用 Nginx 反向代理

```nginx
server {
    listen 443 ssl;
    server_name agentflow.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:6767;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 📚 更多文档

- [AgentFlow Helper 使用指南](../tools/agentflow-helper/README.md)
- [双认证系统指南](AUTH_GUIDE.md)
- [单机部署指南](SINGLE_DEPLOYMENT_GUIDE.md)
- [系统架构文档](docs/ARCHITECTURE.md)
- [功能特性文档](docs/FEATURES.md)

---

**文档版本**: 1.0.0
**最后更新**: 2026-01-28
