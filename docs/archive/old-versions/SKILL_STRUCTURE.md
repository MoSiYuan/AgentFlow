# AgentFlow Skill 部署结构

## 目录结构

AgentFlow skill 部署后的完整目录结构：

```
~/.claude/skills/agentflow/
├── agentflow.db                    # AgentFlow 主数据库（SQLite）⭐
│
├── README.md                       # Skill 说明文档
│
├── bin/                            # 可执行文件
│   ├── master                      # Master 服务器（Go）
│   ├── worker                      # Worker 进程（Go）
│   └── agentflow                   # AgentFlow CLI 工具
│
├── scripts/                        # 工具脚本
│   ├── setup.sh                    # 初始化脚本
│   ├── backup.sh                   # 数据库备份
│   ├── restore.sh                  # 数据库恢复
│   └── status.sh                   # 状态检查
│
├── config/                         # 配置文件
│   ├── agentflow.yaml             # 主配置文件
│   ├── logging.yaml               # 日志配置
│   └── deployment.yaml            # 部署配置
│
├── logs/                           # 日志文件
│   ├── master.log                 # Master 日志
│   ├── worker.log                 # Worker 日志
│   └── access.log                 # API 访问日志
│
├── data/                           # 数据目录
│   ├── workspace/                 # 任务工作空间
│   └── cache/                     # 缓存文件
│
└── docs/                           # 文档
    ├── API.md                      # API 文档
    ├── DATABASE.md                 # 数据库说明
    └── DEPLOYMENT.md               # 部署指南
```

## 数据库详情

### 主数据库: `agentflow.db`

**位置**: `~/.claude/skills/agentflow/agentflow.db`

**大小**:
- 初始: ~100 KB
- 1000个任务: ~1-2 MB
- 10000个任务: ~10-20 MB

**内容**:
- 任务管理（tasks）
- 任务链（task_chains, task_chain_nodes）
- Claude 映射（claude_mappings）
- Git 集成（git_tasks, git_branches）
- Worker 管理（workers）
- 任务关系（task_relationships）
- 任务检查点（task_checkpoints）
- 任务版本（task_versions）

**性能**:
- 索引优化: 20+ 索引
- 查询响应: < 100ms
- 并发支持: 100+ workers

## 配置文件

### 主配置: `config/agentflow.yaml`

```yaml
# Master 配置
master:
  host: "0.0.0.0"
  port: 8848
  db_path: "~/.claude/skills/agentflow/agentflow.db"
  auto_start: false
  git_enabled: true
  git_boundary_config: ".agentflow/boundaries.json"

# Worker 配置
worker:
  master_url: "http://localhost:8848"
  db_path: "~/.claude/skills/agentflow/agentflow.db"
  group_name: "default"
  mode: "auto"

# Claude 集成配置
claude:
  enabled: true
  sync_dir: "~/.claude/sessions"
  auto_sync: true
  sync_interval: 5  # 秒
```

## 日志文件

### 日志位置

```
~/.claude/skills/agentflow/logs/
├── master.log      # Master 服务器日志
├── worker.log      # Worker 进程日志
└── access.log      # HTTP API 访问日志
```

### 日志级别

- `DEBUG`: 详细调试信息
- `INFO`: 一般信息（默认）
- `WARN`: 警告信息
- `ERROR`: 错误信息

### 日志轮转

日志文件自动轮转：
- 单个文件最大 10 MB
- 保留最近 7 天的日志
- 压缩归档旧日志

## 数据目录

### 工作空间: `data/workspace/`

每个任务的工作空间：

```
data/workspace/{task_id}/
├── input/              # 输入文件
├── output/             # 输出文件
├── temp/               # 临时文件
└── logs/               # 任务日志
```

### 缓存: `data/cache/`

缓存文件：
- Git 对象缓存
- 任务结果缓存
- 会话状态缓存

## 部署步骤

### 1. 创建目录结构

```bash
# 创建 skill 目录
mkdir -p ~/.claude/skills/agentflow/{bin,scripts,config,logs,data/{workspace,cache},docs}

# 进入 skill 目录
cd ~/.claude/skills/agentflow
```

### 2. 复制可执行文件

```bash
# 复制 Go 版本
cp /path/to/agentflow/bin/master ./bin/
cp /path/to/agentflow/bin/worker ./bin/

# 或复制 Node.js 版本
cp -r /path/to/agentflow/nodejs ./nodejs
```

### 3. 初始化配置

```bash
# 创建配置文件
cat > config/agentflow.yaml <<EOF
master:
  host: "0.0.0.0"
  port: 8848
  db_path: "~/.claude/skills/agentflow/agentflow.db"
  auto_start: false
EOF

# 运行初始化脚本
./scripts/setup.sh
```

### 4. 启动服务

```bash
# 启动 Master
./bin/master

# 启动 Worker
./bin/worker
```

### 5. 验证安装

```bash
# 检查数据库
ls -lh agentflow.db

# 检查 API
curl http://localhost:8848/health

# 检查状态
./scripts/status.sh
```

## 备份策略

### 数据库备份

```bash
# 自动备份（每天）
0 2 * * * ~/.claude/skills/agentflow/scripts/backup.sh

# 手动备份
./scripts/backup.sh
```

### 备份文件

```
~/.claude/skills/agentflow/backups/
├── agentflow.db.20260126       # 完整备份
├── agentflow.db.20260125       # 完整备份
└── agentflow.db.20260124       # 完整备份
```

### 恢复

```bash
# 从备份恢复
./scripts/restore.sh agentflow.db.20260126
```

## 环境变量

```bash
# AgentFlow 配置
export AGENTFLOW_HOME="~/.claude/skills/agentflow"
export AGENTFLOW_DB="$AGENTFLOW_HOME/agentflow.db"
export AGENTFLOW_CONFIG="$AGENTFLOW_HOME/config"
export AGENTFLOW_LOGS="$AGENTFLOW_HOME/logs"

# Git 集成
export GIT_REPO_PATH="$AGENTFLOW_HOME/workspace"
export GIT_BOUNDARY_CONFIG="$AGENTFLOW_HOME/config/git-boundaries.json"

# Claude 集成
export CLAUDE_SESSION_DIR="~/.claude/sessions"
export CLAUDE_SYNC_ENABLED="true"
export CLAUDE_SYNC_INTERVAL="5"
```

## 进程管理

### 使用 systemd（Linux）

```ini
# /etc/systemd/system/agentflow-master.service
[Unit]
Description=AgentFlow Master Server
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/home/your-user/.claude/skills/agentflow
ExecStart=/home/your-user/.claude/skills/agentflow/bin/master
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 使用 launchd（macOS）

```xml
# ~/Library/LaunchAgents/com.agentflow.master.plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.agentflow.master</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/your-user/.claude/skills/agentflow/bin/master</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>/Users/your-user/.claude/skills/agentflow</string>
</dict>
</plist>
```

## 监控和检查

### 健康检查

```bash
# API 健康检查
curl http://localhost:8848/health

# 数据库完整性
sqlite3 ~/.claude/skills/agentflow/agentflow.db "PRAGMA integrity_check;"

# 磁盘使用
du -sh ~/.claude/skills/agentflow/
```

### 性能监控

```bash
# 查看进程
ps aux | grep agentflow

# 查看资源使用
top -p $(pgrep agentflow)

# 查看连接
lsof -i :8848
```

## 更新和维护

### 更新版本

```bash
# 1. 备份数据库
./scripts/backup.sh

# 2. 停止服务
killall master

# 3. 更新可执行文件
cp /path/to/new/bin/master ./bin/master

# 4. 重启服务
./bin/master
```

### 清理旧数据

```bash
# 清理旧任务（30天前）
sqlite3 agentflow.db "DELETE FROM tasks WHERE created_at < datetime('now', '-30 days');"

# 压缩数据库
sqlite3 agentflow.db "VACUUM;"

# 清理日志
find logs/ -name "*.log" -mtime +7 -delete
```

## 故障排查

### 常见问题

**1. 数据库锁定**
```bash
lsof agentflow.db
kill -9 $(pgrep -f agentflow)
```

**2. 权限错误**
```bash
chmod 755 bin/
chmod 644 agentflow.db
```

**3. 端口占用**
```bash
lsof -i :8848
kill -9 $(lsof -t -i :8848)
```

## 相关文档

- [数据库详细说明](./DATABASE_LOCATION.md)
- [快速开始指南](./DATABASE_PATH.md)
- [API 文档](./API.md)
- [部署指南](./DEPLOYMENT.md)

---

**最后更新**: 2026-01-26
**版本**: 1.0
