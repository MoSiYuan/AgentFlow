# AgentFlow - AI Agent 任务协作系统

> **多进程并发，真 AI 执行** - 分布式任务协作平台

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8E.svg)](https://golang.org/)
[![Claude CLI](https://img.shields.io/badge/Claude%20CLI-1.0.102-blue.svg)](https://github.com/anthropics/claude-code)

## 🎯 核心特性

- ✅ **真正的多进程并发** - 每个 Worker 独立进程，任务自动分配
- ✅ **Claude CLI 深度集成** - AI 任务执行，4-5秒/任务
- ✅ **高性能架构** - 10,000+ req/s，~20MB/进程
- ✅ **完整 REST API** - 任务管理、Worker 监控
- ✅ **即插即用** - 预编译二进制，无需编译
- ✅ **上下文优化** - 节省 token，批量操作

## 📊 性能数据

| 指标 | 数值 | 说明 |
|------|------|------|
| 并发能力 | 3+ Workers | 真正多进程并发 |
| HTTP 吞吐量 | 10,000+ req/s | 高性能处理 |
| 内存使用 | ~20MB/进程 | 低资源占用 |
| 启动时间 | <100ms | 快速启动 |
| 二进制大小 | 34MB | 单文件部署 |
| 任务执行 | 4-5秒 | Claude CLI 集成 |

## 🚀 快速开始

### 3 命令启动（最简单）

```bash
# 1. 启动 Master
cd /path/to/AgentFlow
./bin/master --mode standalone --port 8848

# 2. 启动 Worker（另一个终端）
./bin/worker --mode standalone --master http://127.0.0.1:8848 --name worker1 --auto

# 3. 创建任务
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{"task_id": "TASK-1", "title": "Test", "description": "shell:echo Hello World", "priority": "high"}'
```

### 多进程并发示例

```bash
# 启动 3 个 Worker
for i in 1 2 3; do
  ./bin/worker --mode standalone --master http://127.0.0.1:8848 \
    --name "worker-$i" --auto &
done

# 创建 5 个任务
for i in 1 2 3 4 5; do
  curl -X POST http://127.0.0.1:8848/api/tasks/create \
    -H "Content-Type: application/json" \
    -d "{\"task_id\": \"TASK-$i\", \"title\": \"任务 $i\", \"description\": \"shell:echo Task $i && date\", \"priority\": \"high\"}"
done

# 查看结果
curl http://127.0.0.1:8848/api/status | jq '.'
```

## 📦 安装方式

### 方式 1: Claude Code Skill（推荐）

```bash
# 安装为 skill
cp skills/agentflow.md ~/.claude/commands/

# 使用 skill
/agentflow status           # 查看状态
/agentflow workers          # 查看 Workers
/agentflow list            # 查看任务
```

### 方式 2: 预编译二进制

**无需编译**，直接使用：

```bash
git clone -b feature/1.0.0 https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow

# 启动 Master
./bin/master --mode standalone --port 8848

# 启动 Worker
./bin/worker --mode standalone --master http://127.0.0.1:8848 --name w1 --auto
```

### 方式 3: Docker 部署

```bash
# Standalone 模式
docker-compose -f deployments/docker/docker-compose.standalone.yml up

# Cloud 模式
docker-compose -f deployments/docker/docker-compose.cloud.yml up
```

## 🔧 核心概念

### Master-Worker 架构

```
┌─────────────┐
│   Master    │ ← HTTP API (port 8848)
│  (port 8848)│
└──────┬──────┘
       │
       ├─→ Worker 1 (PID 73811) ✅
       ├─→ Worker 2 (PID 73812) ✅
       └─→ Worker 3 (PID 73813) ✅
              ↓
         任务并发执行
```

### 任务格式

```bash
# Shell 命令
shell:echo "Hello World" && date

# AI 任务（Claude CLI）
ai:解释什么是Agent

# 复杂命令
shell:cd /path && make build && ./app

# 多命令
shell:echo "Step 1" && sleep 1 && echo "Step 2"
```

### REST API

```bash
# 健康检查
GET /api/health

# 系统状态
GET /api/status

# 创建任务
POST /api/tasks/create
{
  "task_id": "TASK-1",
  "title": "任务标题",
  "description": "shell:command",
  "priority": "high"
}

# 查询任务
GET /api/tasks/pending
GET /api/tasks/completed
GET /api/workers
```

## 📖 文档

- [安装指南](INSTALL.md) - 详细安装说明
- [快速入门](docs/GETTING_STARTED.md) - 基础使用教程
- [架构设计](docs/ARCHITECTURE.md) - 系统架构
- [API 文档](docs/API.md) - REST API 参考
- [Skill 手册](skills/agentflow.md) - Claude Code Skill 完整手册

## 💡 使用场景

### 1. 本地开发工作流

```bash
# 代码质量检查
./quick-task.sh "格式化" "shell:gofmt -w ."
./quick-task.sh "测试" "shell:go test ./..."
./quick-task.sh "构建" "shell:go build"
```

### 2. 并发测试

```bash
# 3 个 Worker 并发执行
for i in 1 2 3; do
  ./bin/worker --mode standalone --master http://127.0.0.1:8848 \
    --name "w$i" --auto &
done

# 创建测试任务
for i in 1 2 3; do
  ./quick-task.sh "测试组$i" "shell:go test ./... -run TestGroup$i"
done
```

### 3. AI 任务

```bash
# 使用 Claude CLI 执行 AI 任务
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{"task_id": "AI-1", "title": "AI任务", "description": "解释什么是Agent", "priority": "high"}'
```

## 🎯 特性说明

### 真正的多进程并发

**验证**:
```
Worker-1 PID: 73811 ✅
Worker-2 PID: 73812 ✅
Worker-3 PID: 73813 ✅

并发执行证明：
- 多个 Worker 同时抢任务
- 竞态条件正常（先到先得）
- 任务自动分配
```

### Claude CLI 集成

**自动检测**:
- Worker 自动查找 `~/bin/claudecli`
- 如果找到，使用 Claude CLI 执行
- 如果未找到，回退到模拟模式

**Wrapper 脚本**:
```bash
# ~/bin/claudecli（自动安装）
#!/bin/bash
# 将旧 API 转换为新 Claude CLI 格式
exec /opt/homebrew/bin/claude -p "$@"
```

### 上下文优化

**节省 Token 技巧**:
1. 使用 Task ID 关联（避免重复上下文）
2. 批量创建任务（减少往返）
3. Skill 快捷命令（`/agentflow add`）

## 📁 项目结构

```
AgentFlow/
├── bin/                    # 二进制文件
│   ├── agentflow         # 主程序
│   ├── master            # Master 服务器
│   └── worker            # Worker客户端
├── skills/               # Claude Code Skills
│   └── agentflow.md      # Skill 手册
├── docs/                 # 文档
│   ├── ARCHITECTURE.md
│   ├── API.md
│   └── GETTING_STARTED.md
├── old/                 # 旧代码归档
│   ├── cmd/              # 源代码
│   ├── internal/         # 内部包
│   └── docs/             # 文档
├── quick-task.sh         # 快速任务脚本
├── deployments/         # 部署配置
└── README.md            # 本文档
```

## 🔗 相关链接

- **GitHub**: https://github.com/MoSiYuan/AgentFlow
- **分支**: feature/1.0.0
- **Issue**: 报告问题和建议

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

---

**AgentFlow** - 让 AI Agent 协作更简单 ✨
