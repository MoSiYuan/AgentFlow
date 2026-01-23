# AgentFlow Worker Groups 文档

Worker Groups 是 AgentFlow 的核心功能之一，用于按环境、能力或用途对 Worker 进行分组，实现任务的智能分发。

## 概述

Worker Group 允许你：

- 按操作系统分组 (Linux, Windows, macOS)
- 按运行环境分组 (Docker, Kubernetes, 裸机)
- 按能力分组 (GPU, 高性能, 存储)
- 实现任务定向分发

## 预定义组

AgentFlow 内置以下工作组：

| 组名 | 说明 | 自动检测 |
|------|------|----------|
| `default` | 默认组，未指定组时使用 | ❌ |
| `linux` | Linux 系统 | ✅ |
| `windows` | Windows 系统 | ✅ |
| `darwin` | macOS 系统 | ✅ |
| `docker` | Docker 容器环境 | ✅ |
| `k8s` | Kubernetes 环境 | ✅ |

## 自动检测

Worker 启动时会自动检测运行环境：

```go
func detectGroup() string {
    // 检查环境变量
    if group := os.Getenv("WORKER_GROUP"); group != "" {
        return group
    }

    // 检查 Docker
    if _, err := os.Stat("/.dockerenv"); err == nil {
        return "docker"
    }

    // 检查 Kubernetes
    if _, err := os.Stat("/var/run/secrets/kubernetes.io"); err == nil {
        return "k8s"
    }

    // 默认使用操作系统
    switch runtime.GOOS {
    case "windows":
        return "windows"
    case "linux":
        return "linux"
    case "darwin":
        return "darwin"
    default:
        return "default"
    }
}
```

## 使用场景

### 1. 按操作系统分组

**场景**: 需要在不同操作系统上执行任务

```bash
# 添加 Linux 任务
cpds add "编译 Linux 二进制" \
  --desc "shell:GOOS=linux go build -o app-linux ." \
  --group linux

# 添加 Windows 任务
cpds add "编译 Windows 二进制" \
  --desc "shell:GOOS=windows go build -o app.exe ." \
  --group windows

# Linux Worker 会自动获取 linux 组任务
# Windows Worker 会自动获取 windows 组任务
```

### 2. Docker 环境隔离

**场景**: 容器内执行任务，避免污染主机环境

```bash
# 启动 Docker Worker
docker run -d \
  -e MASTER_URL=http://master:6767 \
  -e WORKER_GROUP=docker \
  -v /var/run/docker.sock:/var/run/docker.sock \
  agentflow-go worker

# 添加 Docker 任务
cpds add "构建镜像" \
  --desc "shell:docker build -t myapp:latest ." \
  --group docker
```

### 3. Kubernetes 部署

**场景**: 在 K8s 集群中运行任务

```yaml
# Deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cpds-worker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cpds-worker
  template:
    metadata:
      labels:
        app: cpds-worker
    spec:
      containers:
      - name: worker
        image: agentflow-go:latest
        args: ["worker"]
        env:
        - name: MASTER_URL
          value: "http://cpds-master:6767"
        - name: WORKER_GROUP
          value: "k8s"
```

```bash
# 添加 K8s 任务
cpds add "更新 Deployment" \
  --desc "shell:kubectl set image deployment/myapp app=myapp:v2" \
  --group k8s
```

### 4. 自定义组

**场景**: 按业务逻辑或能力分组

```bash
# GPU Worker
export WORKER_GROUP="gpu"
cpds worker

# 添加 GPU 任务
cpds add "训练模型" \
  --desc "shell:python train.py --gpu" \
  --group gpu

# 高性能 Worker
export WORKER_GROUP="high-mem"
cpds worker

# 添加大内存任务
cpds add "大数据处理" \
  --desc "shell:spark-submit process.py" \
  --group high-mem
```

## 配置方法

### 方法 1: 环境变量

```bash
export WORKER_GROUP=docker
cpds worker
```

### 方法 2: 命令行参数

```bash
cpds worker --group docker
```

### 方法 3: 自动检测

不指定时，Worker 自动检测环境：

```bash
cpds worker  # 自动检测并设置组
```

## Docker Compose 示例

多组 Worker 同时运行：

```yaml
version: '3.8'
services:
  master:
    image: agentflow-go:latest
    command: master --db /data/agentflow.db
    ports:
      - "6767:6767"
    volumes:
      - cpds-data:/data

  worker-linux:
    image: agentflow-go:latest
    command: worker
    environment:
      - MASTER_URL=http://master:6767
      - WORKER_GROUP=linux
    depends_on:
      - master

  worker-docker:
    image: agentflow-go:latest
    command: worker
    environment:
      - MASTER_URL=http://master:6767
      - WORKER_GROUP=docker
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    depends_on:
      - master

  worker-windows:
    image: agentflow-go:windows
    command: worker
    environment:
      - MASTER_URL=http://master:6767
      - WORKER_GROUP=windows
    depends_on:
      - master

volumes:
  cpds-data:
```

## 任务分配策略

### 1. 精确匹配

任务 `group_name` 必须与 Worker `group_name` 完全匹配：

```bash
# 任务
cpds add "任务" --group docker

# 只有 docker 组 Worker 会执行
cpds worker --group docker
```

### 2. 默认组

未指定组的任务分配到 `default` 组：

```bash
cpds add "简单任务"  # --group 默认为 default

# default 组或未指定组的 Worker 会执行
cpds worker  # 自动检测可能不是 default
```

### 3. 查询特定组任务

```bash
# 列出 docker 组任务
cpds list --group docker

# 列出 windows 组任务
cpds list --group windows
```

## 最佳实践

### 1. 命名规范

- 使用小写字母: `gpu`, `high-mem`
- 使用连字符分隔: `windows-gpu`, `linux-highmem`
- 语义化命名: `build`, `test`, `deploy`

### 2. 环境隔离

```bash
# 开发环境
export WORKER_GROUP=dev
cpds worker

# 测试环境
export WORKER_GROUP=test
cpds worker

# 生产环境
export WORKER_GROUP=prod
cpds worker
```

### 3. 能力标注

```bash
# GPU Worker
export WORKER_GROUP=gpu-tesla-v100
cpds worker

# 添加需要特定 GPU 的任务
cpds add "训练大模型" \
  --desc "shell:python train.py --gpu v100" \
  --group gpu-tesla-v100
```

### 4. 地理分布

```bash
# 北京区域
export WORKER_GROUP=cn-beijing
cpds worker

# 纽约区域
export WORKER_GROUP=us-newyork
cpds worker

# 添加区域相关任务
cpds add "部署到北京" --group cn-beijing
cpds add "部署到纽约" --group us-newyork
```

## 调试

### 查看 Worker 组

```bash
$ cpds workers

共有 3 个 Worker

🟢 [worker-001] local
   工作组: linux
   状态: active
   最后心跳: 2024-01-01 12:00:00

🟢 [worker-002] local
   工作组: docker
   状态: active
   最后心跳: 2024-01-01 12:00:05

🟢 [worker-003] remote
   工作组: windows
   状态: active
   最后心跳: 2024-01-01 12:00:03
```

### 查看特定组任务

```bash
$ cpds list --group docker

共有 2 个任务

⏳ [5] 构建 Docker 镜像
   工作组: docker
   状态: pending
   创建时间: 2024-01-01 10:00:00

▶️ [6] 运行容器测试
   工作组: docker
   状态: running
   执行者: worker-002
   创建时间: 2024-01-01 10:05:00
```

### API 查询

```bash
# 查询所有 docker 组 workers
curl "http://localhost:6767/api/v1/workers?group=docker"

# 查询所有 docker 组待执行任务
curl "http://localhost:6767/api/v1/tasks?status=pending&group=docker"
```

## 常见问题

### Q: 任务一直 pending，没有 Worker 执行？

检查任务的 `group_name` 是否有对应的 Worker：

```bash
# 查看任务组
cpds list --group <task_group>

# 查看该组的 Worker
cpds workers --group <task_group>

# 如果没有 Worker，启动对应组的 Worker
cpds worker --group <task_group>
```

### Q: 如何在同一个机器上运行多个组的 Worker？

使用不同的终端或后台进程：

```bash
# 终端 1
cpds worker --group linux &

# 终端 2
cpds worker --group docker &

# 终端 3
cpds worker --group gpu &
```

### Q: Worker 可以属于多个组吗？

当前一个 Worker 只能属于一个组。如需多组支持，启动多个 Worker 进程，每个指定不同的组。

### Q: 如何动态更改 Worker 组？

重启 Worker 并指定新组：

```bash
# 停止当前 Worker (Ctrl+C)
# 启动新组 Worker
cpds worker --group new-group
```

## 未来改进

- [ ] Worker 多组支持
- [ ] 组优先级
- [ ] 组间任务转发
- [ ] 组能力标注 (CPU, 内存, GPU)
- [ ] 动态组创建
- [ ] 组负载均衡
