# AgentFlow Python - 安装指南

## 📋 前提条件

- Python 3.8+ (`python3 --version`)
- pip (Python 包管理器)

## 🚀 安装方式

### 方式 1: 虚拟环境（推荐）

```bash
# 1. 创建虚拟环境
cd /path/to/AgentFlow/python
python3 -m venv venv

# 2. 激活虚拟环境
source venv/bin/activate

# 3. 安装依赖
pip install -r requirements.txt

# 4. 验证安装
python -m agentflow.cli master --help
```

### 方式 2: 用户安装

```bash
# 使用 --user 标志安装到用户目录
pip install --user Flask requests
export PATH="$HOME/.local/bin:$PATH"

# 验证
python3 -m agentflow.cli master --help
```

### 方式 3: 系统级安装（需要 sudo）

```bash
# 警告：可能影响系统 Python 环境
sudo pip3 install Flask requests
```

## 🧪 测试安装

### 快速测试（代码结构）

```bash
cd python
./quick-test.sh
```

### 完整测试（需要 Flask）

```bash
# 激活虚拟环境（如果使用）
source venv/bin/activate

# 运行测试
./test.sh
```

## 🚀 使用方法

### 启动 Master

```bash
# 方式 1: 使用 CLI
python -m agentflow.cli master --port 8848

# 方式 2: 使用 Python 代码
python << EOF
from agentflow import Master
master = Master(port=8848, auto_shutdown=True)
master.run()
EOF
```

### 启动 Worker

```bash
# 方式 1: 使用 CLI
python -m agentflow.cli worker --auto

# 方式 2: 使用 Python 代码
python << EOF
from agentflow import Worker
worker = Worker(master_url="http://127.0.0.1:8848", auto_mode=True)
worker.run()
EOF
```

### 创建任务

```bash
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TASK-1",
    "title": "测试",
    "description": "shell:echo Hello from Python",
    "priority": "high"
  }'
```

## 🔧 故障排查

### 问题 1: ModuleNotFoundError: No module named 'flask'

**解决方案**:
```bash
# 使用虚拟环境
python3 -m venv venv
source venv/bin/activate
pip install Flask requests
```

### 问题 2: externally-managed-environment 错误

**解决方案**:
```bash
# 方式 1: 使用虚拟环境（推荐）
python3 -m venv venv
source venv/bin/activate
pip install Flask requests

# 方式 2: 使用 pipx
pipx install agentflow

# 方式 3: 使用 --break-system-packages（不推荐）
pip install Flask requests --break-system-packages
```

### 问题 3: Worker 无法连接 Master

**解决方案**:
```bash
# 1. 检查 Master 是否运行
curl http://127.0.0.1:8848/api/health

# 2. 使用 127.0.0.1 而非 localhost
python -m agentflow.cli worker --master http://127.0.0.1:8848
```

## 📚 更多文档

- [Python README](python/README.md) - Python 版本详细文档
- [主 README](README.md) - Go 版本文档
- [安装指南](INSTALL.md) - Go 版本安装指南

## 🆚 Python vs Go 版本

| 特性 | Python 版本 | Go 版本 |
|------|------------|---------|
| 安装 | pip install | 下载二进制 |
| 依赖 | Flask, requests | 无（静态链接） |
| 性能 | 1,000+ req/s | 10,000+ req/s |
| 内存 | ~50MB | ~20MB |
| 跨平台 | ✅ | ✅ |
| 易调试 | ✅ | ❌ |
| 生产环境 | ⚠️  | ✅ |

**建议**:
- 本地开发/学习 → Python 版本
- 生产环境/云部署 → Go 版本
