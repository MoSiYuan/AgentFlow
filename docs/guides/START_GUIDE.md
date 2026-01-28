# 🚀 AgentFlow 启动指南

## ⚡ 快速启动

### 方式 1: 一键启动（推荐）

```bash
./start-all.sh
```

这将自动在新的终端窗口中启动：
- ✅ Dashboard (前端)
- ✅ Rust Master (后端，如果已安装 Rust)

---

### 方式 2: 手动启动

#### 终端 1 - 启动 Dashboard

```bash
cd dashboard
./start-dev.sh
```

或手动执行：

```bash
export PATH="/opt/homebrew/bin:$PATH"
npm install
npm run dev
```

Dashboard 将在 `http://localhost:5173` 启动

#### 终端 2 - 启动 Rust Master（可选）

```bash
cd rust
./start-master.sh
```

或手动执行：

```bash
cargo run --bin agentflow-master -- \
    --node-id master-1 \
    --port 6767 \
    --peers master-1:6767,master-2:6768,master-3:6769
```

Master 将在 `http://localhost:6767` 启动

---

## 📋 启动脚本说明

### 1. `start-all.sh` (主启动脚本)
- 位置: 项目根目录
- 功能: 一键启动所有服务
- 特点: 自动在新终端窗口中启动各服务

### 2. `dashboard/start-dev.sh`
- 位置: `dashboard/` 目录
- 功能: 启动 Dashboard 开发服务器
- 依赖: Node.js, npm

### 3. `rust/start-master.sh`
- 位置: `rust/` 目录
- 功能: 启动 Rust Master 服务器
- 依赖: Rust, Cargo

---

## 🔧 环境要求

### 必需
- **Node.js 18+** - JavaScript 运行时
- **npm 9+** - 包管理器

### 可选
- **Rust 1.70+** - 系统后端（如需启动 Master）
- **Cargo** - Rust 包管理器

### 安装依赖

#### macOS
```bash
# 安装 Node.js
brew install node

# 安装 Rust（可选）
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

#### Linux
```bash
# 安装 Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 安装 Rust（可选）
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

---

## 🌐 访问地址

启动成功后，可以访问：

| 服务 | 地址 | 说明 |
|------|------|------|
| **Dashboard** | http://localhost:5173 | Web 管理界面 |
| **Master API** | http://localhost:6767 | REST API |
| **WebSocket** | ws://localhost:8849 | 实时通信 |
| **API 文档** | http://localhost:6767/docs | API 文档 |

---

## 🛠️ 故障排查

### 问题 1: `node: command not found`

**原因**: Node.js 不在 PATH 中

**解决方案**:
```bash
export PATH="/opt/homebrew/bin:$PATH"
./start-all.sh
```

### 问题 2: `cargo: command not found`

**原因**: Rust 未安装

**解决方案**:
```bash
# 安装 Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

### 问题 3: 端口已被占用

**原因**: 6767 或 5173 端口已被使用

**解决方案**:
```bash
# 查找并终止占用端口的进程
lsof -ti:6767 | xargs kill -9
lsof -ti:5173 | xargs kill -9
```

### 问题 4: Dashboard 启动但无法加载数据

**原因**: Master 未启动或连接失败

**解决方案**:
1. 确认 Master 正在运行: `curl http://localhost:6767/health`
2. 检查 `.env` 中的 `VITE_API_URL` 配置
3. 查看浏览器控制台的错误信息

---

## 📊 验证启动

### 1. 检查 Dashboard

```bash
curl http://localhost:5173
```

应该返回 HTML 页面。

### 2. 检查 Master API

```bash
curl http://localhost:6767/health
```

应该返回健康状态。

### 3. 检查集群状态

```bash
curl http://localhost:6767/api/v1/cluster/status
```

应该返回集群信息。

---

## 🎯 下一步

启动成功后：

1. **打开浏览器** 访问 http://localhost:5173
2. **查看集群拓扑** - 可视化 Master 和 Worker 节点
3. **监控 Workers** - 实时查看 Worker 状态和资源
4. **创建工作流** - 提交 DAG 任务执行
5. **管理分布式锁** - 协调跨节点操作

---

## 💡 开发提示

### Dashboard 热重载

修改代码后会自动刷新页面，无需重启。

### Master 重新编译

修改 Rust 代码后需要重新编译：

```bash
cd rust
cargo build --release
./start-master.sh
```

### 日志查看

- Dashboard 日志: 在浏览器控制台查看
- Master 日志: 在启动 Master 的终端窗口查看

---

**准备好了吗？运行 `./start-all.sh` 开始吧！** 🚀
