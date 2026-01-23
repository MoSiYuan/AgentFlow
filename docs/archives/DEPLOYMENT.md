# AgentFlow 部署指南

本文档介绍如何在不同环境中部署 AgentFlow。

## 系统要求

### Master

- CPU: 1 核心以上
- 内存: 512MB 以上
- 磁盘: 1GB 以上（SQLite 数据库）
- 网络: 6767 端口
- 操作系统: Linux/macOS/Windows

### Worker

- CPU: 1 核心以上
- 内存: 256MB 以上
- 网络: 能访问 Master (6767 端口)

## 单机部署

### 快速启动

```bash
# 1. 下载或编译
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go
make build

# 2. 初始化数据库
./bin/cpds init agentflow.db

# 3. 启动 Master（自动启动本地 Workers）
./bin/cpds master --db agentflow.db

# 4. 添加任务
./bin/cpds add "测试任务" --desc "shell:echo Hello"
```

### 系统服务 (Linux)

创建 systemd 服务：

```bash
sudo cat > /etc/systemd/system/cpds-master.service << 'EOF'
[Unit]
Description=AgentFlow Master Service
After=network.target

[Service]
Type=simple
User=cpds
WorkingDirectory=/opt/cpds
ExecStart=/opt/cpds/bin/cpds master --db /opt/cpds/data/agentflow.db
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable cpds-master
sudo systemctl start cpds-master
```

## Docker 部署

### Docker Compose

创建 `docker-compose.yml`:

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
    restart: unless-stopped

  worker:
    image: agentflow-go:latest
    command: worker
    environment:
      - MASTER_URL=http://master:6767
      - WORKER_GROUP=docker
    depends_on:
      - master
    restart: unless-stopped
    deploy:
      replicas: 3

volumes:
  cpds-data:
```

启动：

```bash
docker-compose up -d
```

## Kubernetes 部署

### Master Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cpds-master
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cpds-master
  template:
    metadata:
      labels:
        app: cpds-master
    spec:
      containers:
      - name: master
        image: agentflow-go:latest
        command: ["cpds"]
        args: ["master", "--db", "/data/agentflow.db"]
        ports:
        - containerPort: 6767
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: cpds-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: cpds-master
spec:
  selector:
    app: cpds-master
  ports:
  - port: 6767
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cpds-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

### Worker Deployment

```yaml
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
        command: ["cpds"]
        args: ["worker"]
        env:
        - name: MASTER_URL
          value: "http://cpds-master:6767"
        - name: WORKER_GROUP
          value: "k8s"
```

部署：

```bash
kubectl apply -f master-deployment.yaml
kubectl apply -f worker-deployment.yaml
```

## 分布式部署

### Master 服务器

```bash
# 安装
wget https://github.com/jiangxiaolong/agentflow-go/releases/latest/download/cpds-linux-amd64
chmod +x cpds-linux-amd64
sudo mv cpds-linux-amd64 /usr/local/bin/cpds

# 创建用户
sudo useradd -r -s /bin/false cpds

# 创建数据目录
sudo mkdir -p /data/cpds
sudo chown cpds:cpds /data/cpds

# 初始化数据库
sudo -u cpds cpds init /data/cpds/agentflow.db

# 启动服务
sudo -u cpds cpds master --db /data/cpds/agentflow.db --host 0.0.0.0
```

### Worker 服务器

```bash
# Linux Worker
export MASTER_URL=http://master-server:6767
export WORKER_GROUP=linux
cpds worker

# Windows Worker
set MASTER_URL=http://master-server:6767
set WORKER_GROUP=windows
cpds.exe worker
```

## 备份和恢复

### 备份脚本

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/cpds"
DB_PATH="/data/cpds/agentflow.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR
sqlite3 $DB_PATH ".backup $BACKUP_DIR/cpds_$DATE.db"
gzip $BACKUP_DIR/cpds_$DATE.db
find $BACKUP_DIR -name "cpds_*.db.gz" -mtime +7 -delete
```

### 定时备份

```bash
crontab -e
0 2 * * * /opt/cpds/scripts/backup.sh
```

## 故障排除

### Master 无法启动

```bash
# 检查端口
netstat -tunlp | grep 6767

# 检查数据库
ls -la /data/cpds/agentflow.db

# 查看日志
journalctl -u cpds-master -f
```

### Worker 无法连接

```bash
# 测试连通性
telnet master-server 6767

# 检查防火墙
sudo firewall-cmd --add-port=6767/tcp --permanent
```

### 任务卡住

```bash
# 查看锁状态
sqlite3 /data/cpds/agentflow.db "SELECT id, title, lock_holder FROM tasks WHERE status='running'"

# 手动解锁（谨慎）
sqlite3 /data/cpds/agentflow.db "UPDATE tasks SET status='pending', lock_holder=NULL WHERE id=1"
```

## 安全建议

1. 使用 VPN 或 VPC 进行网络隔离
2. 添加 API Key 认证
3. 启用 TLS/HTTPS
4. 定期备份数据库
5. 及时更新版本
