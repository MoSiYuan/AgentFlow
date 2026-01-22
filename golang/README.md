# AgentFlow Go - 云端部署版本

> **高性能 Master-Worker 系统** | 专为 Docker/Kubernetes 设计

## 📦 版本定位

这是 AgentFlow 的 **Go 语言版本**，专为云端部署和容器化环境设计。

**推荐场景**:
- ✅ Kubernetes pod 部署
- ✅ Docker 容器化
- ✅ 微服务架构
- ✅ 生产环境
- ✅ 高并发场景（10,000+ req/s）

**不适合场景**:
- ❌ 本地开发（请使用 [Python 版本](../python/)）
- ❌ 快速原型（请使用 [Python 版本](../python/)）

## 🚀 快速开始

### 方式 1: 使用预编译二进制（推荐）

```bash
# 1. 进入 golang 目录
cd golang

# 2. 验证二进制文件
ls -lh bin/
# 应该看到: master, worker

# 3. 启动 Master
./bin/master --mode standalone --port 8848

# 4. 启动 Worker（另一个终端）
./bin/worker --mode standalone --master http://127.0.0.1:8848 --auto

# 5. 创建任务
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{"task_id": "T1", "title": "Test", "description": "shell:echo Hello", "priority": "high"}'
```

### 方式 2: Docker 部署

```bash
# 1. Standalone 模式
cd golang
docker-compose -f deployments/docker/docker-compose.standalone.yml up -d

# 2. Cloud 模式
docker-compose -f deployments/docker/docker-compose.cloud.yml up -d
```

### 方式 3: Kubernetes 部署

```bash
cd golang
kubectl apply -f deployments/k8s/
```

## 📋 性能指标

| 指标 | 数值 | 说明 |
|------|------|------|
| HTTP 吞吐量 | 10,000+ req/s | 高性能 HTTP 处理 |
| 内存使用 | ~20MB/进程 | 低资源占用 |
| 启动时间 | <100ms | 快速冷启动 |
| 二进制大小 | 34MB | 静态链接，无依赖 |
| 并发能力 | 100+ Workers | 真正多进程 |

## 🔧 配置说明

### Master 配置

```bash
./bin/master [flags]

Flags:
  --mode string       # 部署模式: standalone/cloud (default "standalone")
  --host string       # 监听地址 (default "0.0.0.0")
  -p, --port int      # 监听端口 (default 8848)
  --auto-shutdown     # standalone 模式：任务完成后自动关闭
  -h, --help          # 帮助信息
```

### Worker 配置

```bash
./bin/worker [flags]

Flags:
  --mode string       # 部署模式: standalone/cloud (default "standalone")
  -m, --master string # Master URL (default "http://localhost:8848")
  -n, --name string   # Worker 名称 (default: hostname)
  -a, --auto          # 自动模式：自动拉取并执行任务
  --oneshot           # 执行一个任务后退出（standalone 模式）
  -h, --help          # 帮助信息
```

## 🐳 Docker 部署

### Standalone 模式

```yaml
# deployments/docker/docker-compose.standalone.yml
version: '3.8'
services:
  master:
    image: agentflow:latest
    ports:
      - "8848:8848"
    command: ["./master", "--mode", "standalone", "--auto-shutdown"]
    restart: unless-stopped
```

### Cloud 模式

```yaml
# deployments/docker/docker-compose.cloud.yml
version: '3.8'
services:
  master:
    image: agentflow:latest
    ports:
      - "8848:8848"
    command: ["./master", "--mode", "cloud"]
    restart: always

  worker:
    image: agentflow:latest
    depends_on:
      - master
    environment:
      - MASTER_URL=http://master:8848
    command: ["./worker", "--mode", "cloud", "--auto"]
    restart: always
    deploy:
      replicas: 3
```

## ☸️ Kubernetes 部署

```bash
# 部署 Master 和 Worker
kubectl apply -f deployments/k8s/

# 查看 pods
kubectl get pods -l app=agentflow

# 查看 logs
kubectl logs -f deployment/agentflow-master
kubectl logs -f deployment/agentflow-worker
```

## 📚 API 文档

### 核心 API

```bash
# 健康检查
GET /api/health

# 创建任务
POST /api/tasks/create
{
  "task_id": "ID",
  "title": "标题",
  "description": "shell:命令",
  "priority": "high"
}

# 查询状态
GET /api/status
GET /api/tasks/pending
GET /api/tasks/completed
GET /api/workers
```

**完整 API 文档**: [docs/api.md](../docs/api.md)

## 🔨 编译源码

### 前提条件

- Go 1.21+
- Make（可选）

### 编译步骤

```bash
# 1. 克隆仓库
git clone -b feature/1.0.0 https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow/golang

# 2. 编译
make build

# 3. 验证
ls -lh bin/
```

### Make 命令

```bash
make build      # 编译所有二进制
make master     # 只编译 master
make worker     # 只编译 worker
make test       # 运行测试
make clean      # 清理编译文件
```

## 🔍 故障排查

### 问题 1: 端口被占用

```bash
# 查找占用进程
lsof -i:8848

# 杀掉进程
kill -9 $(lsof -ti:8848)

# 或使用其他端口
./bin/master --port 8850
```

### 问题 2: Worker 无法连接 Master

```bash
# 1. 检查 Master 是否运行
curl http://localhost:8848/api/health

# 2. 检查网络
ping localhost

# 3. 使用 127.0.0.1 而非 localhost（IPv6 问题）
./bin/worker --master http://127.0.0.1:8848
```

### 问题 3: Docker 容器无法启动

```bash
# 启动 Docker
open -a Docker  # macOS
systemctl start docker  # Linux

# 检查 Docker 状态
docker ps

# 查看日志
docker-compose logs
```

**更多故障排查**: [docs/installation.md](../docs/installation.md#故障排查)

## 📊 与 Python 版本对比

| 特性 | Go 版本 | Python 版本 |
|------|---------|-------------|
| 适用场景 | 云端/生产 | 本地/开发 |
| 性能 | 10,000+ req/s | 1,000+ req/s |
| 内存 | ~20MB | ~50MB |
| 部署方式 | 二进制/Docker/K8s | pip install |
| 启动时间 | <100ms | ~1s |
| 依赖 | 无（静态链接） | Flask, requests |

**选择建议**:
- 生产环境、云部署 → **Go 版本**
- 本地开发、学习 → [Python 版本](../python/)

## 📚 相关文档

- [主 README](../README.md) - 项目总览
- [安装指南](../docs/installation.md#go-版本) - 详细安装步骤
- [架构设计](../docs/architecture.md) - 系统架构
- [API 文档](../docs/api.md) - REST API 参考

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

详见 [贡献指南](../docs/contributing.md)

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE)

---

**版本**: v1.0.0
**分支**: [feature/1.0.0](https://github.com/MoSiYuan/AgentFlow/tree/feature/1.0.0)
