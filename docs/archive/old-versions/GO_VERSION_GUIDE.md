# Go 版本使用指南

本文档提供 AgentFlow Go 版本的详细使用指南。

## 目录

- [快速开始](#快速开始)
- [架构说明](#架构说明)
- [构建指南](#构建指南)
- [部署指南](#部署指南)
- [配置选项](#配置选项)

## 快速开始

### 零依赖运行（推荐）

Go 版本无需任何依赖，直接使用预编译的二进制文件：

```bash
# macOS (ARM64)
./agentflow-master-darwin-arm64 --port 6767 --db data/agentflow.db
./agentflow-worker-darwin-arm64 --master http://localhost:6767

# Linux (AMD64)
./agentflow-master-linux-amd64 --port 6767 --db data/agentflow.db
./agentflow-worker-linux-amd64 --master http://localhost:6767

# Windows (AMD64)
./agentflow-master-windows-amd64.exe --port 6767 --db data/agentflow.db
./agentflow-worker-windows-amd64.exe --master http://localhost:6767
```

### 使用便捷脚本

```bash
# Linux/macOS
./agentflow-go.sh run '["echo hello","echo world"]'

# Windows
agentflow-go.bat run '["echo hello","echo world"]'
```

## 架构说明

### 组件

#### Master Server

**职责:**
- 任务队列管理
- Worker 注册与调度
- 任务状态跟踪
- 提供 REST API

**启动:**
```bash
./agentflow-master-darwin-arm64 \
  --port 6767 \
  --db data/agentflow.db \
  --auto-shutdown=false
```

#### Worker

**职责:**
- 从 Master 获取任务
- 执行任务（Shell 或调用 Claude CLI）
- 报告执行结果
- 定期发送心跳

**启动:**
```bash
./agentflow-worker-darwin-arm64 \
  --master http://localhost:6767 \
  --group default \
  --mode auto \
  --max-concurrent 3
```

## 构建指南

### 环境要求

- **Go**: 1.21+
- **OS**: macOS, Linux, Windows

### 编译步骤

```bash
# 进入 Master 目录
cd cmd/agentflow-master

# 编译（当前平台）
go build -o agentflow-master

# 交叉编译所有平台
./build-all.sh

# 进入 Worker 目录
cd ../../agentflow-worker

# 编译
go build -o agentflow-worker

# 交叉编译
./build-all.sh
```

### 编译输出

```
bin/
├── agentflow-master-darwin-arm64
├── agentflow-master-darwin-amd64
├── agentflow-master-linux-amd64
├── agentflow-master-windows-amd64.exe
├── agentflow-worker-darwin-arm64
├── agentflow-worker-darwin-amd64
├── agentflow-worker-linux-amd64
└── agentflow-worker-windows-amd64.exe
```

## 部署指南

### 单机部署

```bash
# 1. 启动 Master
./agentflow-master-darwin-arm64 \
  --port 6767 \
  --db /var/lib/agentflow/agentflow.db \
  > /var/log/agentflow-master.log 2>&1 &

MASTER_PID=$!

# 2. 启动多个 Worker
./agentflow-worker-darwin-arm64 \
  --master http://localhost:6767 \
  --group production \
  > /var/log/agentflow-worker-1.log 2>&1 &

WORKER1_PID=$!

./agentflow-worker-darwin-arm64 \
  --master http://localhost:6767 \
  --group production \
  > /var/log/agentflow-worker-2.log 2>&1 &

WORKER2_PID=$!

# 3. 保存 PID
echo $MASTER_PID > /var/run/agentflow-master.pid
echo $WORKER1_PID > /var/run/agentflow-worker-1.pid
echo $WORKER2_PID > /var/run/agentflow-worker-2.pid
```

### 系统服务（systemd）

#### Master Service

```ini
[Unit]
Description=AgentFlow Master Server
After=network.target

[Service]
Type=simple
User=agentflow
WorkingDirectory=/opt/agentflow
ExecStart=/opt/agentflow/agentflow-master-linux-amd64 \
  --port 6767 \
  --db /var/lib/agentflow/agentflow.db
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

#### Worker Service

```ini
[Unit]
Description=AgentFlow Worker
After=network.target agentflow-master.service

[Service]
Type=simple
User=agentflow
WorkingDirectory=/opt/agentflow
ExecStart=/opt/agentflow/agentflow-worker-linux-amd64 \
  --master http://localhost:6767 \
  --group production
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**安装服务:**

```bash
# 复制服务文件
sudo cp agentflow-master.service /etc/systemd/system/
sudo cp agentflow-worker.service /etc/systemd/system/

# 启用并启动
sudo systemctl enable agentflow-master
sudo systemctl enable agentflow-worker
sudo systemctl start agentflow-master
sudo systemctl start agentflow-worker

# 查看状态
sudo systemctl status agentflow-master
sudo systemctl status agentflow-worker
```

### Docker 部署

#### Dockerfile

```dockerfile
FROM alpine:latest

WORKDIR /app

# 复制二进制文件
COPY bin/agentflow-master-linux-amd64 /app/agentflow-master
COPY bin/agentflow-worker-linux-amd64 /app/agentflow-worker

# 创建数据目录
RUN mkdir -p /data

# 暴露端口
EXPOSE 6767

# 启动 Master
CMD ["/app/agentflow-master", "--port", "6767", "--db", "/data/agentflow.db"]
```

#### Docker Compose

```yaml
version: '3.8'

services:
  master:
    image: agentflow:latest
    ports:
      - "6767:6767"
    volumes:
      - agentflow-data:/data
    restart: always

  worker:
    image: agentflow:latest
    depends_on:
      - master
    environment:
      - AGENTFLOW_MASTER_URL=http://master:6767
    deploy:
      replicas: 3
    restart: always

volumes:
  agentflow-data:
```

## 配置选项

### Master 配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--port` | int | 6767 | 监听端口 |
| `--host` | string | 0.0.0.0 | 绑定地址 |
| `--db` | string | :memory: | 数据库路径 |
| `--auto-shutdown` | bool | false | 无任务时自动关闭 |
| `--ws-enabled` | bool | true | 启用 WebSocket |
| `--ws-port` | int | 8849 | WebSocket 端口 |

### Worker 配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--master` | string | http://localhost:6767 | Master URL |
| `--id` | string | auto-generated | Worker ID |
| `--group` | string | default | 工作组名 |
| `--mode` | string | auto | 运行模式 (auto/manual/oneshot) |
| `--max-concurrent` | int | 3 | 最大并发数 |
| `--task-timeout` | int | 300000 | 任务超时 (ms) |
| `--heartbeat-interval` | int | 30000 | 心跳间隔 (ms) |
| `--retry-on-failure` | bool | true | 失败重试 |
| `--max-retries` | int | 3 | 最大重试次数 |

### 环境变量

```bash
# Master
export AGENTFLOW_PORT=6767
export AGENTFLOW_DB=/data/agentflow.db
export AGENTFLOW_AUTO_SHUTDOWN=false

# Worker
export AGENTFLOW_MASTER_URL=http://localhost:6767
export AGENTFLOW_GROUP_NAME=production
export AGENTFLOW_MODE=auto
```

## API 使用示例

### 创建任务

```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "部署应用",
    "description": "./deploy.sh",
    "priority": "high",
    "group_name": "production"
  }'
```

### 查询任务状态

```bash
# 方式 1: 通过 ID
curl http://localhost:6767/api/v1/tasks/1

# 方式 2: 通过 TASK-ID
curl http://localhost:6767/api/v1/tasks/TASK-00000001
```

### 获取统计信息

```bash
curl http://localhost:6767/api/v1/stats
```

### 获取待处理任务

```bash
curl http://localhost:6767/api/v1/tasks/pending?group=production
```

## 监控和日志

### 日志级别

```bash
# 设置日志级别
export LOG_LEVEL=debug  # debug | info | warn | error

# 启动 Master
./agentflow-master-darwin-arm64 --port 6767
```

### 日志输出

```bash
# 标准输出
./agentflow-master-darwin-arm64 --port 6767 2>&1 | tee /var/log/agentflow.log

# 后台运行并记录日志
nohup ./agentflow-master-darwin-arm64 \
  --port 6767 \
  >> /var/log/agentflow.log 2>&1 &
```

### 健康检查

```bash
# 基本健康检查
curl http://localhost:6767/health

# 监控脚本
while true; do
  curl -s http://localhost:6767/health || echo "Master down!"
  sleep 60
done
```

## 性能调优

### 1. 数据库优化

```bash
# 使用 WAL 模式（更好的并发性能）
./agentflow-master-darwin-arm64 \
  --db /var/lib/agentflow/agentflow.db \
  --db-cache-size 10000
```

### 2. Worker 并发优化

```bash
# 增加并发数
./agentflow-worker-darwin-arm64 \
  --master http://localhost:6767 \
  --max-concurrent 10 \
  --task-timeout 120000
```

### 3. 网络优化

```bash
# 调整心跳间隔（减少网络负载）
./agentflow-worker-darwin-arm64 \
  --master http://localhost:6767 \
  --heartbeat-interval 60000
```

## 故障排除

### 问题 1: 端口已被占用

**错误:**
```
bind: address already in use
```

**解决方案:**
```bash
# 查找占用端口的进程
lsof -i :6767

# 杀死进程
kill -9 <PID>

# 或使用其他端口
./agentflow-master-darwin-arm64 --port 6768
```

### 问题 2: 数据库锁定

**错误:**
```
database is locked
```

**解决方案:**
```bash
# 确保只有一个 Master 实例
ps aux | grep agentflow-master

# 检查数据库文件权限
ls -la /var/lib/agentflow/agentflow.db
```

### 问题 3: Worker 无法连接 Master

**检查:**
```bash
# Master 是否运行
curl http://localhost:6767/health

# 网络连通性
telnet localhost 6767

# 防火墙
sudo ufw status
sudo ufw allow 6767
```

## 最佳实践

1. **生产环境**
   - 使用文件数据库（持久化）
   - 启用多个 Worker 实例
   - 配置自动重启
   - 监控日志和性能

2. **开发环境**
   - 使用内存数据库
   - 单个 Worker 实例
   - 启用详细日志

3. **安全**
   - 使用反向代理 (nginx)
   - 启用 HTTPS
   - 限制 API 访问
   - 定期备份数据库

4. **高可用**
   - 部署多个 Master 实例（需要负载均衡器）
   - 使用分布式数据库
   - 配置健康检查和自动故障转移

## 相关文档

- [Node.js 版本指南](./NODEJS_GUIDE.md)
- [主 README](../README.md)
- [API 文档](./API_REFERENCE.md)

---

**最后更新**: 2026-01-24
**维护者**: AgentFlow Team
