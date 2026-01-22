# AgentFlow Python - 本地部署版本

> **简化版 AgentFlow** | 专为本地部署设计 | 3 秒启动

## 📚 版本说明

AgentFlow 现在提供两个版本：

### 🐧 Go 版本（云端部署）
- **适用场景**: Kubernetes pod、云端服务器
- **特点**: 高性能、生产级、功能完整
- **位置**: [根目录](../) - 预编译二进制
- **文档**: [README.md](../README.md) | [INSTALL.md](../INSTALL.md)

### 🐍 Python 版本（本地部署）
- **适用场景**: 本地开发、个人使用、快速测试
- **特点**: 安装简单、即插即用、功能完整
- **位置**: [python/](./) 目录（本文件）
- **功能**: 与 Go 版本完全兼容

## 🚀 快速开始（2 步）

### 1. 安装

```bash
cd python
pip install -r requirements.txt
# 或者
pip install -e .
```

### 2. 启动

```bash
# Terminal 1: 启动 Master
python -m agentflow.cli master --port 8848

# Terminal 2: 启动 Worker
python -m agentflow.cli worker --name w1 --auto

# Terminal 3: 创建任务
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{"task_id": "T1", "title": "Test", "description": "shell:echo Hello", "priority": "high"}'
```

## 📦 安装方式

### 方式 1: pip 安装（推荐）

```bash
# 从本地安装
cd /path/to/AgentFlow/python
pip install -e .

# 使用命令
agentflow master --help
agentflow worker --help
```

### 方式 2: 直接运行

```bash
cd python

# 运行 Master
python -m agentflow.cli master --port 8848

# 运行 Worker
python -m agentflow.cli worker --auto
```

### 方式 3: 作为模块导入

```python
from agentflow import Master, Worker

# 启动 Master
master = Master(port=8848, auto_shutdown=True)
master.run()

# 启动 Worker
worker = Worker(master_url="http://127.0.0.1:8848", auto_mode=True)
worker.run()
```

## 🎯 核心功能

### ✅ 与 Go 版本功能一致

- **Master-Worker 架构**: 完全兼容 Go 版本的 API
- **任务管理**: 创建、分配、执行、完成
- **多进程并发**: 真正的并行执行
- **Claude CLI 集成**: 自动调用 Claude CLI
- **SQLite 数据库**: 持久化存储
- **RESTful API**: 与 Go 版本相同的 API 接口

### 🔧 Python 特有优势

- **零编译**: 无需编译，直接运行
- **跨平台**: Windows/macOS/Linux 通用
- **易调试**: Python 代码易于修改和调试
- **轻量级**: 代码量少，易于理解

## 📋 API 示例

### 创建任务

```bash
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TASK-1",
    "title": "测试任务",
    "description": "shell:echo Hello World",
    "priority": "high"
  }'
```

### 查询状态

```bash
# 系统状态
curl http://127.0.0.1:8848/api/status

# 待处理任务
curl http://127.0.0.1:8848/api/tasks/pending

# 在线 Workers
curl http://127.0.0.1:8848/api/workers
```

## 📊 性能对比

| 指标 | Go 版本 | Python 版本 |
|------|---------|-------------|
| 启动时间 | <100ms | ~1s |
| 内存使用 | ~20MB | ~50MB |
| HTTP 吞吐量 | 10,000+ req/s | 1,000+ req/s |
| 任务执行 | 4-5秒 | 4-5秒 |
| 部署难度 | 中等 | 简单 |

**结论**: Python 版本非常适合本地开发和测试。

## 🔧 配置说明

### Master 配置

```bash
python -m agentflow.cli master \
  --host 0.0.0.0 \
  --port 8848 \
  --db ~/.agentflow/tasks.db \
  --auto-shutdown  # standalone 模式
```

### Worker 配置

```bash
python -m agentflow.cli worker \
  --master http://127.0.0.1:8848 \
  --name worker-1 \
  --auto          # 自动模式
  # --oneshot     # 执行一个任务后退出
  # --manual      # 手动模式
```

## 🤖 Claude CLI 集成

Worker 自动按优先级查找 Claude CLI：

1. `~/bin/claudecli` (wrapper 脚本)
2. `claude` (系统 PATH)
3. 模拟模式（无 Claude CLI）

### 安装 claudecli wrapper

参见 [安装指南 > Claude CLI 配置](../INSTALL.md#claude-cli-配置)

## 📝 任务格式

```bash
# Shell 命令
shell:echo "Hello World" && date

# AI 任务（Claude CLI）
ai:解释什么是 Agent

# 复杂命令
shell:cd /path && make build && ./app
```

## 🧪 测试

```bash
# 单个测试任务
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TEST-1",
    "title": "测试",
    "description": "shell:echo Python AgentFlow",
    "priority": "high"
  }'

# 批量测试
for i in {1..5}; do
  curl -X POST http://127.0.0.1:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d "{\"task_id\": \"TEST-$i\", \"title\": \"任务 $i\", \"description\": \"shell:echo Task $i\", \"priority\": \"high\"}"
done
```

## 🔍 故障排查

### Master 无法启动

```bash
# 检查端口
lsof -i:8848

# 使用其他端口
python -m agentflow.cli master --port 8850
```

### Worker 无法连接

```bash
# 检查 Master 是否运行
curl http://127.0.0.1:8848/api/health

# 使用 127.0.0.1 而非 localhost
python -m agentflow.cli worker --master http://127.0.0.1:8848
```

### Claude CLI 不工作

```bash
# 检查 claudecli
which claudecli
claudecli --version

# 手动测试
claudecli chat --prompt "测试" --no-interactive
```

## 📚 相关文档

- [主 README](../README.md) - Go 版本文档
- [安装指南](../INSTALL.md) - 详细安装步骤
- [Skill 手册](../skills/agentflow.md) - Claude Code 集成
- [架构设计](../docs/ARCHITECTURE.md) - 系统架构

## 🆚 版本选择

### 选择 Go 版本，如果：
- ✅ 部署到 Kubernetes/Docker
- ✅ 需要高性能和高并发
- ✅ 生产环境使用
- ✅ 单一二进制文件部署

### 选择 Python 版本，如果：
- ✅ 本地开发环境
- ✅ 快速测试功能
- ✅ 个人使用
- ✅ 需要修改源码

## 📄 许可证

MIT License - 与主项目相同

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**版本**: v1.0.0
**分支**: [feature/1.0.0](https://github.com/MoSiYuan/AgentFlow/tree/feature/1.0.0)
