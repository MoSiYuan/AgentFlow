# AgentFlow - AI Agent 任务协作系统

> **多进程并发，真 AI 执行** - 分布式任务协作平台

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8E.svg)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.8+-blue.svg)](https://python.org/)
[![Claude CLI](https://img.shields.io/badge/Claude%20CLI-1.0.102-blue.svg)](https://github.com/anthropics/claude-code)

## 🎯 项目简介

AgentFlow 是一个 Master-Worker 架构的异步任务协作系统，支持真正的多进程并发和 Claude CLI 深度集成。

## 📦 版本选择

AgentFlow 提供两个版本，功能完全兼容，API 100% 相同：

### 🐧 Go 版本（云端部署）

**推荐场景**: Kubernetes pod、云端服务器、生产环境

- ✅ 高性能（10,000+ req/s）
- ✅ 低资源占用（~20MB）
- ✅ 单一二进制文件
- ✅ Docker/Kubernetes 友好

**位置**: [golang/](golang/) | **文档**: [docs/installation.md](docs/installation.md#go-版本)

```bash
cd golang
./bin/master --mode standalone --port 8848
./bin/worker --mode standalone --master http://127.0.0.1:8848 --auto
```

### 🐍 Python 版本（本地部署）

**推荐场景**: 本地开发、个人使用、快速测试

- ✅ 零编译，即插即用
- ✅ 跨平台（Windows/macOS/Linux）
- ✅ 易调试和修改
- ✅ pip 安装

**位置**: [python/](python/) | **文档**: [docs/installation.md](docs/installation.md#python-版本)

```bash
cd python
pip install -r requirements.txt
python -m agentflow.cli master --port 8848
python -m agentflow.cli worker --auto
```

## 🚀 快速开始

### 1. 选择版本

```bash
# 云端/生产 → Go 版本
cd golang

# 本地/开发 → Python 版本
cd python
```

### 2. 启动 Master

```bash
# Go 版本
./bin/master --mode standalone --port 8848

# Python 版本
python -m agentflow.cli master --port 8848
```

### 3. 启动 Worker

```bash
# Go 版本
./bin/worker --mode standalone --master http://127.0.0.1:8848 --auto

# Python 版本
python -m agentflow.cli worker --auto
```

### 4. 创建任务

```bash
curl -X POST http://127.0.0.1:8848/api/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TASK-1",
    "title": "测试任务",
    "description": "shell:echo Hello AgentFlow",
    "priority": "high"
  }'
```

## 📊 核心特性

- ✅ **真正的多进程并发** - 每个 Worker 独立进程，任务自动分配
- ✅ **Claude CLI 深度集成** - AI 任务执行，4-5秒/任务
- ✅ **完整 REST API** - 任务管理、Worker 监控
- ✅ **SQLite 持久化** - 任务状态持久化存储
- ✅ **上下文优化** - 节省 token，批量操作
- ✅ **跨平台支持** - Windows/macOS/Linux

## 📋 性能对比

| 指标 | Go 版本 | Python 版本 |
|------|---------|-------------|
| HTTP 吞吐量 | 10,000+ req/s | 1,000+ req/s |
| 内存使用 | ~20MB/进程 | ~50MB/进程 |
| 启动时间 | <100ms | ~1s |
| 并发能力 | 3+ Workers | 3+ Workers |
| 二进制大小 | 34MB | N/A |
| 依赖管理 | 无（静态链接） | Flask, requests |

## 📁 项目结构

```
AgentFlow/
├── golang/              # Go 版本（云端部署）
│   ├── bin/            # 预编译二进制
│   ├── internal/       # 源代码
│   └── deployments/    # Docker/K8s 配置
│
├── python/             # Python 版本（本地部署）
│   ├── agentflow/      # Python 包
│   └── requirements.txt
│
├── docs/               # 所有文档
│   ├── installation.md # 安装指南
│   ├── architecture.md # 架构设计
│   └── scripts/        # 实用脚本
│
├── skills/             # Claude Code Skill
│   └── agentflow.md    # Skill 手册
│
└── README.md           # 本文件
```

## 📚 文档

- [安装指南](docs/installation.md) - Go 和 Python 版本安装步骤
- [架构设计](docs/architecture.md) - 系统架构和设计理念
- [API 文档](docs/api.md) - REST API 完整参考
- [Skill 手册](skills/agentflow.md) - Claude Code 集成指南

## 🎯 使用场景

### 云端部署（Go 版本）
- ✅ Kubernetes pod 部署
- ✅ Docker 容器化
- ✅ 微服务架构
- ✅ 高并发场景

### 本地开发（Python 版本）
- ✅ 本地开发环境
- ✅ 快速功能测试
- ✅ 学习和调试
- ✅ 个人项目

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

详见 [贡献指南](docs/contributing.md)

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 🔗 相关链接

- **GitHub**: https://github.com/MoSiYuan/AgentFlow
- **分支**: [feature/1.0.0](https://github.com/MoSiYuan/AgentFlow/tree/feature/1.0.0)
- **Issue**: https://github.com/MoSiYuan/AgentFlow/issues

---

**版本**: v1.0.0
**更新**: 2026-01-22
