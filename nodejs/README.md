# AgentFlow Node.js - 零依赖部署版本

**开箱即用，无需安装 Node.js 和依赖**

---

## ⚡ 快速开始（3 种方式）

### 方式 1: 独立可执行文件 ⭐ 推荐

```bash
# 开发者：打包
cd nodejs && ./package.sh

# 用户：直接运行
./dist/agentflow-master
```

**特点**:
- ✅ 无需 Node.js
- ✅ 单文件部署
- ✅ 开箱即用

### 方式 2: Docker 🐳

```bash
cd nodejs
npm run docker:compose
```

**特点**:
- ✅ 完全隔离
- ✅ 一键启动
- ✅ 自动重启

### 方式 3: 完整依赖包

```bash
cd nodejs
./package.sh
cd dist/bundle && ./start-master.sh
```

**特点**:
- ✅ 包含所有依赖
- ✅ 可调试
- ✅ 跨平台

---

## 📦 关于打包

### 为什么需要打包？

AgentFlow Node.js 版本依赖以下内容：
- Node.js 20 LTS
- pnpm 包管理器
- better-sqlite3 (需要编译)
- 其他 npm 包

为了让用户无需安装这些依赖，我们提供打包方案。

### 打包后的文件

```
nodejs/
├── dist/
│   ├── agentflow-master        # 独立可执行文件 (⭐ 推荐)
│   └── bundle/                 # 完整依赖包
│       ├── master/
│       ├── worker/
│       ├── node_modules/       # 已包含所有依赖
│       └── start-master.sh
├── Dockerfile.standalone        # Docker 镜像
└── docker-compose.yml          # Docker Compose 配置
```

---

## 🚀 部署流程

### 开发者：打包发布

```bash
# 1. 准备环境
cd nodejs
pnpm install
npm rebuild better-sqlite3
pnpm run build

# 2. 打包
chmod +x package.sh
./package.sh

# 3. 分发
# - 独立可执行文件: dist/agentflow-master
# - Docker 镜像: docker build -t agentflow:latest .
# - 完整包: dist/bundle/
```

### 用户：部署使用

#### 选项 A: 独立可执行文件

```bash
# 1. 下载文件
wget https://github.com/xxx/agentflow/releases/download/v1.0.0/agentflow-master

# 2. 添加执行权限
chmod +x agentflow-master

# 3. 运行
./agentflow-master
```

#### 选项 B: Docker

```bash
# 1. 拉取镜像
docker pull agentflow:latest

# 2. 运行容器
docker run -d -p 6767:6767 agentflow:latest
```

#### 选项 C: 完整包

```bash
# 1. 下载并解压
tar -xzf agentflow-bundle.tar.gz
cd bundle

# 2. 运行
./start-master.sh
```

---

## 📊 打包方案对比

| 特性 | 独立可执行文件 | Docker | 完整包 |
|------|---------------|--------|--------|
| **文件大小** | ~100MB | ~200MB | ~50MB |
| **需要 Node.js** | ❌ | ❌ | ✅ |
| **需要 Docker** | ❌ | ✅ | ❌ |
| **部署难度** | ⭐ 最简单 | ⭐⭐ 简单 | ⭐⭐ 中等 |
| **启动速度** | 快 | 中 | 快 |
| **可调试性** | ❌ | ❌ | ✅ |
| **推荐场景** | 生产环境 | 云部署 | 开发/测试 |

---

## 🎯 使用场景

### 本地开发

```bash
# 使用完整依赖包（可调试）
cd nodejs/dist/bundle
./start-master.sh
```

### 单机部署

```bash
# 使用独立可执行文件（最简单）
./agentflow-master
```

### 服务器部署

```bash
# 使用 Docker（推荐）
docker-compose up -d
```

### Kubernetes

```bash
# 使用 K8s 部署文件
kubectl apply -f deployment/k8s/
```

---

## 🔧 配置

### 默认配置

- **端口**: 6767
- **数据库**: `~/.claude/skills/agentflow/agentflow.db`
- **日志**: 控制台输出

### 自定义配置

```bash
# 命令行参数
agentflow-master --port 8080 --db /data/agentflow.db

# 环境变量
export AGENTFLOW_DB_PATH="/data/agentflow.db"
export AGENTFLOW_PORT=8080
agentflow-master
```

---

## 📈 性能

### 独立可执行文件

- **启动时间**: < 1s
- **内存占用**: ~50MB
- **并发能力**: 100+ Workers

### Docker

- **启动时间**: ~2s
- **内存占用**: ~100MB (容器开销)
- **并发能力**: 100+ Workers

---

## 🧪 验证部署

```bash
# 健康检查
curl http://localhost:6767/health

# 查看状态
curl http://localhost:6767/api/status

# 创建任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"测试任务","description":"echo Hello"}'
```

---

## 📚 文档

- [快速开始](./QUICK_START.md) - 5分钟快速部署
- [打包指南](./PACKAGING_GUIDE.md) - 详细打包说明
- [开发指南](../docs/NODEJS_GUIDE.md) - 开发者指南
- [状态报告](./NODEJS_STATUS.md) - 版本状态

---

## 🎉 总结

**推荐部署方式**:

1. **生产环境**: 独立可执行文件 (`dist/agentflow-master`)
2. **云环境**: Docker (`docker-compose up`)
3. **开发环境**: 完整依赖包 (`dist/bundle/`)

**一键部署命令**:

```bash
# 独立可执行文件
cd nodejs && ./package.sh && sudo cp dist/agentflow-master /usr/local/bin/

# Docker
cd nodejs && npm run docker:compose
```

---

**版本**: 1.0.0
**最后更新**: 2026-01-26
**维护者**: AgentFlow Team
