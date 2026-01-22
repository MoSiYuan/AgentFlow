# AgentFlow 快速入门指南

本指南将帮助你快速上手 AgentFlow (Claude Parallel Development System)。

## 前置要求

- Go 1.21 或更高版本
- SQLite 3 (Go 会自动处理)

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go

# 编译
make build

# 或者使用 go build
go build -o cpds ./cmd/cpds
```

### 使用 go install 安装

```bash
go install github.com/jiangxiaolong/agentflow-go/cmd/cpds@latest
```

## 基本使用

### 1. 初始化数据库

```bash
cpds init agentflow.db
```

### 2. 启动 Master 服务

```bash
# 默认配置
cpds master

# 自定义数据库
cpds master --db /path/to/database.db

# 自定义端口
cpds master --port 9000

# 禁用自动启动本地 workers
cpds master --auto-start=false
```

Master 服务将在 `http://localhost:8848` 启动。

### 3. 添加任务

```bash
# 添加简单任务
cpds add "运行测试"

# 添加带描述的任务
cpds add "编译项目" --desc "shell:go build -v ./..." --group linux

# 添加带完成标准的任务
cpds add "部署应用" \
  --desc "script:./deploy.sh" \
  --criteria "应用返回 200 状态码"
```

### 4. 查看任务列表

```bash
# 查看所有任务
cpds list

# 按状态过滤
cpds list --status pending
cpds list --status running
cpds list --status completed

# 按工作组过滤
cpds list --group windows
```

### 5. 查看 Worker 状态

```bash
# 查看所有 workers
cpds workers

# 按工作组过滤
cpds workers --group docker
```

## Worker 使用

### 自动模式（推荐）

启动 Master 时会自动检测本地环境并启动相应的 workers：

- **Linux**: 自动启动 `linux` 组 worker
- **Windows**: 自动启动 `windows` 组 worker
- **macOS**: 自动启动 `darwin` 组 worker
- **Docker**: 自动启动 `docker` 组 worker
- **Kubernetes**: 自动启动 `k8s` 组 worker

### 手动启动 Worker

```bash
# 自动检测工作组
cpds worker

# 指定工作组
cpds worker --group windows

# 连接到远程 Master
cpds worker --master http://master.example.com:8848

# 本地模式（直接读取数据库）
cpds worker --db agentflow.db
```

## 任务格式

任务描述支持以下格式：

### Shell 命令

```bash
cpds add "运行测试" --desc "shell:go test ./..."
```

### 脚本执行

```bash
cpds add "部署" --desc "script:./deploy.sh"
```

### Windows PowerShell

```bash
cpds add "Windows 清理" --desc "shell:Clean-BuildArtifacts.ps1" --group windows
```

## 示例工作流

### 单机开发环境

```bash
# 1. 初始化
cpds init dev.db

# 2. 启动 Master（自动启动本地 workers）
cpds master --db dev.db

# 3. 在另一个终端添加任务
cpds add "格式化代码" --desc "shell:gofmt -w ." --db dev.db
cpds add "运行测试" --desc "shell:go test ./..." --db dev.db
cpds add "构建应用" --desc "shell:go build -v ./..." --db dev.db

# 4. 查看进度
cpds list --db dev.db
```

### 云端分布式环境

```bash
# Master 服务器
cpds master --db /data/agentflow.db --host 0.0.0.0 --port 8848

# Worker 服务器 (Linux)
export MASTER_URL=http://master-server:8848
export WORKER_GROUP=linux
cpds worker

# Worker 服务器 (Windows)
set MASTER_URL=http://master-server:8848
set WORKER_GROUP=windows
cpds worker

# 添加任务
curl -X POST http://master-server:8848/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "数据处理",
    "description": "shell:python process_data.py",
    "group_name": "linux"
  }'
```

## Docker 环境

```bash
# 启动 Master
docker run -d \
  -v /data:/data \
  -p 8848:8848 \
  --name cpds-master \
  agentflow-go master --db /data/agentflow.db

# 启动 Worker
docker run -d \
  -e MASTER_URL=http://cpds-master:8848 \
  -e WORKER_GROUP=docker \
  --name cpds-worker \
  agentflow-go worker
```

## 下一步

- 阅读 [架构设计](ARCHITECTURE.md) 了解系统架构
- 查看 [Worker Group 文档](WORKER_GROUPS.md) 学习工作组配置
- 参考 [API 文档](API.md) 了解 API 接口
- 阅读 [部署指南](DEPLOYMENT.md) 进行生产部署

## 常见问题

### Q: Worker 无法连接到 Master

检查以下几点：
1. Master 服务是否正常运行：`curl http://localhost:8848/health`
2. Worker 的 `--master` 参数是否正确
3. 防火墙是否开放了 8848 端口

### Q: 任务一直处于 pending 状态

可能原因：
1. 没有 Worker 注册到对应的 group
2. Worker 进程崩溃或失去连接
3. 任务锁超时（默认 5 分钟）

使用 `cpds workers` 查看 Worker 状态。

### Q: 如何查看任务执行日志

目前 AgentFlow 将任务执行结果存储在数据库中。你可以使用 API 查询：

```bash
curl http://localhost:8848/api/v1/tasks/1
```

## 获取帮助

- GitHub Issues: https://github.com/jiangxiaolong/agentflow-go/issues
- 文档: https://github.com/jiangxiaolong/agentflow-go/tree/main/docs
