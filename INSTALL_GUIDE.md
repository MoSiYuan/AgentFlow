# AgentFlow 安装指南

本指南提供多种安装方式，适用于不同网络环境和场景。

## 方案 1: 快速体验版（无需网络，纯标准库）

适用于：网络受限环境、快速体验

```bash
# 进入测试目录
cd tests

# 运行纯标准库版本（已包含所有核心功能演示）
go run ctest_pure.go

# 查看生成的故事
cat ctest_stories/story_1.md
```

**功能**：
- ✅ 任务创建和管理
- ✅ Worker 并发执行
- ✅ 故事生成和评审
- ✅ 任务结果聚合

---

## 方案 2: 完整版安装（需要网络连接）

适用于：生产环境、完整功能

### 前置要求
- Go 1.21+
- 网络连接（可访问 GitHub 或 Go 代理）

### 安装步骤

```bash
# 1. 配置 Go 代理（中国大陆推荐）
go env -w GOPROXY=https://goproxy.cn,direct

# 2. 下载依赖
go mod download

# 3. 编译项目
make build

# 或者手动编译
go build -o bin/agentflow ./cmd/agentflow
go build -o bin/master ./cmd/master
go build -o bin/worker ./cmd/worker

# 4. 验证安装
./bin/agentflow version
```

### 验证安装

```bash
# 初始化数据库
./bin/agentflow init agentflow.db

# 启动 Master 服务
./bin/agentflow master --db agentflow.db

# 在另一个终端，添加任务
./bin/agentflow add "测试任务" --desc "shell:echo 'Hello AgentFlow'"

# 查看任务列表
./bin/agentflow list
```

---

## 方案 3: Docker 部署（推荐生产环境）

### 使用 Docker Compose

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 手动 Docker 部署

```bash
# 构建 Master 镜像
docker build -t agentflow:master -f Dockerfile.master .

# 构建 Worker 镜像
docker build -t agentflow:worker -f Dockerfile.worker .

# 启动 Master
docker run -d \
  -v $(pwd)/data:/data \
  -p 8848:8848 \
  --name agentflow-master \
  agentflow:master

# 启动 Worker
docker run -d \
  -e MASTER_URL=http://agentflow-master:8848 \
  --name agentflow-worker \
  agentflow:worker
```

---

## 方案 4: 离线安装（企业内网环境）

### 准备工作（在有网络的机器上）

```bash
# 1. 导出 Go 依赖
go mod vendor

# 2. 打包项目
tar czf agentflow.tar.gz \
  --exclude='.git' \
  --exclude='bin' \
  --exclude='*.db' \
  .

# 3. 传输到目标机器
scp agentflow.tar.gz target-server:/tmp/
```

### 在目标机器上安装

```bash
# 1. 解压
cd /opt
tar xzf /tmp/agentflow.tar.gz
cd agentflow-go

# 2. 使用 vendor 目录编译
go build -mod=vendor -o bin/agentflow ./cmd/agentflow
go build -mod=vendor -o bin/master ./cmd/master
go build -mod=vendor -o bin/worker ./cmd/worker

# 3. 验证
./bin/agentflow version
```

---

## 环境变量配置

```bash
# Master 配置
export AGENTFLOW_DB_PATH=/data/agentflow.db
export AGENTFLOW_HOST=0.0.0.0
export AGENTFLOW_PORT=8848
export AGENTFLOW_AUTO_START_WORKERS=true

# Worker 配置
export MASTER_URL=http://localhost:8848
export WORKER_GROUP=default
export WORKER_ID=worker-$(hostname)

# 安全配置
export SANDBOXED=false
export WORKSPACE_DIR=/tmp/agentflow-workspace
export RESTRICT_PATH=false
export READ_ONLY=false
```

---

## 常见问题

### Q1: go mod download 失败

**解决方案**：
```bash
# 方法 1: 使用国内代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 方法 2: 直接编译（使用缓存）
go build -mod=mod ./...

# 方法 3: 清理缓存后重试
go clean -modcache
go mod download
```

### Q2: SQLite 编译错误

**解决方案**：
```bash
# macOS
brew install sqlite3
pkg-config --cflags sqlite3

# Ubuntu/Debian
sudo apt-get install libsqlite3-dev

# CentOS/RHEL
sudo yum install sqlite-devel
```

### Q3: CGO 相关错误

```bash
# 启用 CGO
export CGO_ENABLED=1

# 或者使用纯 Go 的 SQLite 驱动
# 修改 go.mod，使用 modernc.org/sqlite
```

### Q4: 端口被占用

```bash
# 查找占用进程
lsof -i :8848

# 或者修改端口
./bin/agentflow master --port 9000
```

---

## 下一步

安装完成后，请参考：

- [快速入门](docs/GETTING_STARTED.md) - 基本使用
- [架构设计](docs/ARCHITECTURE.md) - 系统架构
- [AI 快速开始](docs/AI_QUICKSTART.md) - AI 功能
- [开发指南](docs/AI_DEPLOYMENT.md) - 开发和部署

---

## 获取帮助

- GitHub Issues: https://github.com/jiangxiaolong/agentflow-go/issues
- 文档: https://github.com/jiangxiaolong/agentflow-go/tree/main/docs
- 邮件: support@example.com
