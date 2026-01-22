# AgentFlow 安装指南

**版本**: v1.0.0
**更新**: 2026-01-22
**分支**: feature/1.0.0

## 📋 目录

- [系统要求](#系统要求)
- [安装方式](#安装方式)
- [配置说明](#配置说明)
- [验证安装](#验证安装)
- [故障排查](#故障排查)

## 系统要求

### 最低要求

- **操作系统**: Linux/macOS/Windows
- **内存**: 512MB 可用内存
- **磁盘**: 100MB 可用空间
- **网络**: （可选）用于 Claude CLI 功能

### 推荐配置

- **操作系统**: Linux/macOS
- **CPU**: 2 核心以上
- **内存**: 2GB+ 可用内存
- **网络**: 稳定连接（Claude CLI 功能）

### 依赖软件

#### 必需依赖

无（预编译二进制包含所有依赖）

#### 可选依赖（Claude CLI 功能）

- **Claude Code CLI**: 1.0.102+
- **安装方式**: `brew install claude`

```bash
# macOS/Linux
brew install claude

# 验证安装
claude --version
# 输出: 1.0.102 (Claude Code)
```

## 安装方式

### 方式 1: Claude Code Skill（推荐）

#### 优势

- ✅ 集成到 Claude Code
- ✅ 快速调用
- ✅ 上下文优化
- ✅ 节省 token

#### 安装步骤

```bash
# 1. 克隆仓库
git clone -b feature/1.0.0 https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow

# 2. 安装 skill
cp skills/agentflow.md ~/.claude/commands/

# 3. 验证安装
/agentflow status
```

#### 使用示例

```bash
# 查看状态
/agentflow status

# 查看 Workers
/agentflow workers

# 查看任务
/agentflow list

# 创建任务（通过 API）
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{"task_id": "T1", "title": "Test", "description": "shell:echo Hello", "priority": "high"}'
```

### 方式 2: 预编译二进制（无需编译）

#### 优势

- ✅ 无需编译
- ✅ 开箱即用
- ✅ 稳定可靠
- ✅ 跨平台

#### 安装步骤

```bash
# 1. 下载二进制
git clone -b feature/1.0.0 https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow

# 2. 验证二进制
ls -lh bin/
# 应该看到:
# agentflow (34MB)
# master (34MB)
# worker (34MB)

# 3. 设置权限（Linux/macOS）
chmod +x bin/*

# 4. 测试运行
./bin/master --help
./bin/worker --help
```

#### 启动服务

```bash
# Terminal 1: 启动 Master
./bin/master --mode standalone --port 8848

# Terminal 2: 启动 Worker
./bin/worker --mode standalone --master http://127.0.0.1:8848 --name worker1 --auto

# Terminal 3: 测试
curl http://127.0.0.1:8848/api/health
```

### 方式 3: Docker 部署

#### 优势

- ✅ 环境隔离
- ✅ 易于扩展
- ✅ 持久化存储
- ✅ 自动重启

#### 安装步骤

```bash
# 1. 克隆仓库
git clone -b feature/1.0.0 https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow

# 2. Standalone 模式
docker-compose -f deployments/docker/docker-compose.standalone.yml up -d

# 3. Cloud 模式
docker-compose -f deployments/docker/docker-compose.cloud.yml up -d

# 4. 验证
curl http://localhost:8848/api/health
```

#### Docker Compose 配置

**Standalone 模式** (`docker-compose.standalone.yml`):
```yaml
version: '3.8'
services:
  master:
    image: agentflow:latest
    ports:
      - "8848:8848"
    volumes:
      - agentflow-data:/data
    command: ["./master", "--mode", "standalone", "--auto-shutdown"]
    restart: unless-stopped

volumes:
  agentflow-data:
```

**Cloud 模式** (`docker-compose.cloud.yml`):
```yaml
version: '3.8'
services:
  master:
    image: agentflow:latest
    ports:
      - "8848:8848"
    volumes:
      - agentflow-data:/data
    command: ["./master", "--mode", "cloud"]
    restart: always

  worker:
    image: agentflow:latest
    depends_on:
      - master
    environment:
      - MASTER_URL=http://master:8848
      - MODE=cloud
    restart: always
    deploy:
      replicas: 3

volumes:
  agentflow-data:
```

## 配置说明

### Master 配置

#### 命令行参数

```bash
./bin/master [flags]

Flags:
  --mode string       # 部署模式: standalone/cloud (default "standalone")
  --host string       # 监听地址 (default "0.0.0.0")
  -p, --port int        # 监听端口 (default 8848)
  --auto-shutdown      # standalone 模式：任务完成后自动关闭
  -h, --help          # 帮助信息
```

#### 环境变量

```bash
export AGENTFLOW_MODE=standalone
export AGENTFLOW_HOST=0.0.0.0
export AGENTFLOW_PORT=8848
export AGENTFLOW_AUTO_SHUTDOWN=true
```

### Worker 配置

#### 命令行参数

```bash
./bin/worker [flags]

Flags:
  --mode string       # 部署模式: standalone/cloud (default "standalone")
  -m, --master string  # Master URL (default "http://localhost:8848")
  -n, --name string     # Worker 名称 (default: hostname)
  -a, --auto           # 自动模式：自动拉取并执行任务
  --oneshot            # 执行一个任务后退出（standalone 模式）
  -h, --help          # 帮助信息
```

#### 环境变量

```bash
export AGENTFLOW_MODE=standalone
export AGENTFLOW_MASTER_URL=http://localhost:8848
export AGENTFLOW_WORKER_NAME=worker1
export AGENTFLOW_AUTO=true
```

### Claude CLI 配置

#### 自动检测

Worker 会自动查找 `~/bin/claudecli`。

#### 手动安装

```bash
# 1. 创建 wrapper
mkdir -p ~/bin
cat > ~/bin/claudecli << 'EOF'
#!/bin/bash
# AgentFlow Claude CLI wrapper
# 转换旧 API 为新 Claude CLI 格式

PROMPT=""
PRINT_MODE=false
CONTEXT=""

# 解析参数
for ((i=1; i<=$#; i++)); do
    case "${!i}" in
        chat)
            # 忽略 chat 子命令
            ;;
        --prompt)
            PROMPT="${!i+1}"
            ((i++))
            ;;
        --no-interactive)
            PRINT_MODE=true
            ;;
        --context)
            CONTEXT="${!i+1}"
            ((i++))
            ;;
        *)
            # 其他参数
            ;;
    esac
done

# 执行 Claude CLI
if [ "$PRINT_MODE" = true ]; then
    if [ -n "$CONTEXT" ]; then
        PROMPT="$CONTEXT

$PROMPT"
    fi
    exec /opt/homebrew/bin/claude -p "$PROMPT"
else
    exec /opt/homebrew/bin/claude "$PROMPT"
fi
EOF

chmod +x ~/bin/claudecli

# 2. 验证
~/bin/claudecli chat --prompt "test" --no-interactive
```

## 验证安装

### 1. 检查二进制

```bash
cd AgentFlow
ls -lh bin/
```

**预期输出**:
```
agentflow (34MB)
master (34MB)
worker (34MB)
```

### 2. 测试 Master

```bash
./bin/master --mode standalone --port 8848 &
MASTER_PID=$!

sleep 2

# 测试 API
curl http://localhost:8848/api/health

# 停止 Master
kill $MASTER_PID
```

**预期输出**:
```json
{"mode":"standalone","status":"healthy"}
```

### 3. 测试 Worker

```bash
./bin/master --mode standalone --port 8848 &
sleep 2

# 启动 Worker
./bin/worker worker --mode standalone --master http://127.0.0.1:8848 \
  --name test-worker --auto

sleep 3

# 检查状态
curl http://localhost:8848/api/workers

# 清理
killall master worker 2>/dev/null || true
```

**预期输出**:
```json
{"success":true,"data":{"workers":[...]}}
```

### 4. 测试任务创建

```bash
# 启动 Master
./bin/master --mode standalone --port 8848 &
sleep 2

# 创建测试任务
curl -X POST http://localhost:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TEST-INSTALL",
    "title": "安装测试",
    "description": "shell:echo \"AgentFlow installed successfully\"",
    "priority": "high"
  }'

# 查询任务
curl http://localhost:8848/api/tasks/pending

# 清理
killall master 2>/dev/null || true
```

## 故障排查

### 问题 1: 端口被占用

**错误信息**:
```
listen tcp 0.0.0.0:8848: bind: address already in use
```

**解决方案**:
```bash
# 查找占用进程
lsof -i:8848

# 杀掉进程
kill -9 $(lsof -ti:8848)

# 或使用其他端口
./bin/master --port 8850
```

### 问题 2: Worker 无法连接 Master

**错误信息**:
```
Error: failed to register: connection refused
```

**解决方案**:
```bash
# 1. 检查 Master 是否运行
curl http://localhost:8848/api/health

# 2. 检查网络
ping localhost

# 3. 使用 127.0.0.1 而非 localhost（IPv6 问题）
./bin/worker --master http://127.0.0.1:8848
```

### 问题 3: claudecli 执行失败

**错误信息**:
```
level=warning msg="claudecli execution failed: exit status 1"
```

**解决方案**:
```bash
# 1. 检查 wrapper
ls -l ~/bin/claudecli

# 2. 手动测试
export PATH="$HOME/bin:$PATH"
claudecli chat --prompt "test" --no-interactive

# 3. 检查 Claude CLI
which claude
claude --version

# 4. 重新安装 wrapper
rm ~/bin/claudecli
# 重新执行"手动安装 claudecli"步骤
```

### 问题 4: 任务无法分配

**错误信息**:
```
Error: failed to assign task: task not found or not pending
```

**解决方案**:
```
这是正常的！说明：
- 多个 Worker 在并发抢任务
- 第一个 Worker 成功获取任务
- 其他 Worker 看到任务已不在 pending 状态
- 这是 Master-Worker 架构的正常行为
```

### 问题 5: Docker 容器无法启动

**错误信息**:
```
Error: Cannot connect to Docker daemon
```

**解决方案**:
```bash
# 启动 Docker
open -a Docker

# 检查 Docker 状态
docker ps

# 查看日志
docker-compose logs
```

## 高级配置

### 1. 自定义数据库位置

```bash
# 使用自定义数据库
./bin/master --db /path/to/custom.db

# 或通过环境变量
export AGENTFLOW_DB=/path/to/custom.db
```

### 2. 多 Master 部署

```bash
# Master 1
./bin/master --port 8848 &

# Master 2
./bin/master --port 8849 &
```

### 3. Worker Groups

```bash
# 创建专用 Worker 组
./bin/worker --group linux --name linux-worker &
./bin/worker --group windows --name windows-worker &
```

### 4. 任务优先级

```bash
# 创建不同优先级的任务
curl -X POST http://localhost:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "HIGH-1",
    "title": "高优先级",
    "description": "shell:echo high priority",
    "priority": "high"
  }'
```

## 升级和卸载

### 升级

```bash
# 1. 备份当前版本
cp -r AgentFlow AgentFlow.backup

# 2. 拉取最新版本
git pull origin feature/1.0.0

# 3. 替换二进制（如果需要）
# ./bin/master --stop
# cp new_bin/* ./bin/

# 4. 重启服务
killall master worker
./bin/master --mode standalone --port 8848 &
```

### 卸载

```bash
# 1. 停止所有进程
killall master worker 2>/dev/null || true

# 2. 清理数据（可选）
rm -f .claude/cpds-manager/*.db*

# 3. 删除二进制
rm -rf bin/
```

## 下一步

安装完成后，建议：

1. ✅ 阅读 [快速入门](docs/GETTING_STARTED.md)
2. ✅ 查看 [架构设计](docs/ARCHITECTURE.md)
3. ✅ 参考 [Skill 手册](skills/agentflow.md)
4. ✅ 尝试第一个任务

---

**安装支持**: 如有问题，请查看 [故障排查](#故障排查) 或提交 Issue
