# AgentFlow - AI Agent 任务协作系统

> **双版本架构，统一 API** - 本地开发与云端部署的完美结合

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8E.svg)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.8+-blue.svg)](https://python.org/)
[![Claude CLI](https://img.shields.io/badge/Claude%20CLI-1.0.102-blue.svg)](https://github.com/anthropics/claude-code)

---

## 🎯 项目简介

AgentFlow 是一个 **Master-Worker 架构** 的异步 AI 任务协作系统，提供 **两个完全兼容的版本**，满足不同场景需求：

- **🐍 Python 版本** - 本地开发与跨平台自动化
- **🐧 Go 版本** - 云端部署与大规模并行

两个版本 **API 100% 兼容**，可以无缝混合使用。

---

## 📦 版本选择与定位

### 🐍 Python 版本 - 本地开发专家

**核心定位**: 本地部署、跨平台执行、GUI 自动化

**最佳使用场景**:
- ✅ 本地开发环境集成
- ✅ 跨平台脚本执行（Windows/macOS/Linux）
- ✅ GUI 自动化操作（文本编辑器、IDE）
- ✅ 系统级命令调用
- ✅ 快速原型开发与调试
- ✅ 个人项目和轻量级任务

**核心优势**:
```python
# 跨平台 GUI 操作示例
from agentflow import Worker

worker = Worker()
worker.run_task("""
使用 VSCode 打开文件 /path/to/file.py
跳转到第 42 行
选中整行代码
复制到剪贴板
""")
```

**技术特点**:
- 🔹 **零编译** - 即插即用，pip 安装
- 🔹 **跨平台原生** - 完美支持 Windows/macOS/Linux 特有功能
- 🔹 **GUI 集成** - 可调用系统 GUI API（VSCode、TextEdit、等）
- 🔹 **易调试** - Python 生态，pdb/ipdb 调试
- 🔹 **灵活扩展** - 利用 Python 丰富的第三方库

**快速开始**:
```bash
# 安装
cd python
pip install -r requirements.txt

# 启动 Master
python -m agentflow.cli master --port 8848

# 启动 Worker
python -m agentflow.cli worker --auto
```

**文档位置**: [python/README.md](python/README.md)

---

### 🐧 Go 版本 - 云端部署专家

**核心定位**: 容器化部署、云端 AI Agent、大规模并行开发

**最佳使用场景**:
- ✅ Docker/Kubernetes 容器化部署
- ✅ 云端 AI Agent 集群
- ✅ 大规模并行任务处理（1000+ 并发）
- ✅ 生产环境 24/7 运行
- ✅ 微服务架构集成
- ✅ CI/CD 流水线集成

**核心优势**:
```bash
# 云端大规模部署示例
kubectl apply -f golang/deployments/
# → 自动启动 100 个 Worker Pods
# → 每个 Pod 独立执行 AI 任务
# → 支持动态扩缩容
```

**技术特点**:
- 🔹 **高性能** - 10,000+ req/s HTTP 吞吐量
- 🔹 **低资源** - 单进程 ~20MB 内存
- 🔹 **单一二进制** - 无依赖，静态链接
- 🔹 **容器友好** - Docker/K8s 原生支持
- 🔹 **三层执行器**:
  1. HTTP Executor (Claude Server)
  2. Claude CLI Executor
  3. Shell 命令 (fallback)

**快速开始**:
```bash
# 编译（或直接使用预编译二进制）
cd golang
go build -o bin/master cmd/master/main.go
go build -o bin/worker cmd/worker/main.go
go build -o bin/oneshot cmd/oneshot/main.go

# 启动 Master
./bin/master --config config.yaml

# 启动 Worker（持续模式）
./bin/worker --config config.yaml

# 启动 Worker（单次执行）
./bin/oneshot --config config.yaml --timeout 5m
```

**文档位置**: [golang/README.md](golang/README.md)

---

## 🔄 两个版本的协作

### 混合部署架构

```
┌─────────────────────────────────────────────────────────────┐
│                        本地开发机                            │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │  Python      │         │  VSCode GUI  │                 │
│  │  Master      │◄────────┤  Automation  │                 │
│  └──────┬───────┘         └──────────────┘                 │
│         │                                                    │
│         │ REST API                                          │
└─────────┼────────────────────────────────────────────────────┘
          │
          │ HTTPS
          ▼
┌─────────────────────────────────────────────────────────────┐
│                      云端 K8s 集群                          │
│  ┌──────────────┐                                           │
│  │  Go Master   │                                           │
│  └──────┬───────┘                                           │
│         │                                                   │
│         ├──────────────────────────────┐                    │
│         ▼                              ▼                    │
│  ┌──────────┐                    ┌──────────┐              │
│  │ Go Pod 1 │  ...  (100+ Pods) │ Go Pod N │              │
│  │ Worker   │                    │ Worker   │              │
│  └──────────┘                    └──────────┘              │
│         │                              │                    │
│         └────────────┬─────────────────┘                    │
│                      ▼                                       │
│              Claude AI Processing                           │
└─────────────────────────────────────────────────────────────┘
```

**典型工作流**:

1. **本地开发** - 使用 Python 版本
   - 在本地机器上启动 Python Master
   - 执行 GUI 操作（VSCode 编辑、文件操作）
   - 快速迭代和调试

2. **云端执行** - 使用 Go 版本
   - 将大规模计算任务提交到云端 Go Master
   - 100+ Go Workers 并行处理
   - 结果返回本地

3. **混合使用**
   - 本地 Python Master 调度云端 Go Workers
   - 云端 Go Master 调用本地 Python Workers
   - 统一 API，无缝切换

---

## 📊 性能对比

| 指标 | Python 版本 | Go 版本 |
|------|-------------|---------|
| **部署方式** | 本地安装 | Docker/K8s |
| **HTTP 吞吐量** | ~1,000 req/s | ~10,000 req/s |
| **内存使用** | ~50MB/进程 | ~20MB/进程 |
| **启动时间** | ~1s | <100ms |
| **并发能力** | 3-10 Workers | 100-1000 Workers |
| **GUI 支持** | ✅ 原生支持 | ❌ 不支持 |
| **跨平台** | ✅ 完美 | ⚠️ 有限 |
| **容器化** | ⚠️ 需要配置 | ✅ 原生支持 |
| **执行模式** | 交互式 | 批处理/持续 |
| **依赖管理** | pip | 无（静态链接） |

---

## 🚀 快速开始

### 🎯 一键部署（推荐）

使用自动化脚本快速部署 Claude 开发环境：

```bash
# macOS
./scripts/deploy-macos.sh install

# Linux
./scripts/deploy-linux.sh install

# Windows (以管理员身份运行 PowerShell)
scripts\deploy-windows.bat install
```

**脚本功能**:
- ✅ 自动检测环境依赖（Claude CLI、Python、Go）
- ✅ 一键安装所有依赖
- ✅ 编译 Go 版本二进制文件
- ✅ 生成配置文件
- ✅ 启动 Master 和 Worker 服务

**详细文档**: [部署脚本使用指南](docs/deployment-scripts.md)

---

### 场景 1: 本地开发（推荐 Python）

```bash
# 1. 安装 Python 版本
cd python
pip install -r requirements.txt

# 2. 启动 Master
python -m agentflow.cli master --port 8848

# 3. 启动 Worker（新终端）
python -m agentflow.cli worker --auto

# 4. 创建任务
curl -X POST http://127.0.0.1:8848/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "编辑代码文件",
    "description": "使用 VSCode 打开 main.py 并跳转到第 42 行",
    "group_name": "default"
  }'
```

### 场景 2: 云端部署（推荐 Go）

```bash
# 1. 准备配置文件
cat > config.yaml <<EOF
master:
  host: "0.0.0.0"
  port: 8848
worker:
  master_url: "http://master:8848"
  group_name: "production"
EOF

# 2. 启动 Master
./golang/bin/master -config config.yaml

# 3. 启动 Workers（Docker）
docker run -d agentflow-worker \
  -config /config/config.yaml

# 或 Kubernetes
kubectl apply -f golang/deployments/
```

### 场景 3: 混合部署

```bash
# 本地：Python Master
python -m agentflow.cli master --port 8848

# 云端：Go Workers（连接到本地 Master）
./golang/bin/worker \
  -master https://your-local-ip:8848 \
  -group cloud-workers
```

---

## 📚 核心功能

### 通用功能（两个版本都支持）

- ✅ **Master-Worker 架构** - 分布式任务调度
- ✅ **RESTful API** - 完整的任务管理接口
- ✅ **SQLite 持久化** - 任务状态持久化
- ✅ **Claude CLI 集成** - AI 任务执行
- ✅ **优先级队列** - 高/中/低优先级
- ✅ **Worker 组** - 支持多组 Worker
- ✅ **心跳机制** - Worker 健康监控
- ✅ **任务锁定** - 防止并发冲突

### Python 版本独有

- ✅ **GUI 自动化** - VSCode、TextEdit、等
- ✅ **跨平台系统调用** - Windows/macOS/Linux 原生 API
- ✅ **交互式执行** - 支持用户输入
- ✅ **Python 生态集成** - 丰富的第三方库

### Go 版本独有

- ✅ **HTTP 执行器** - 通过 HTTP 调用 Claude Server
- ✅ **OneShot 模式** - 执行单个任务后退出
- ✅ **配置系统** - YAML 文件 + 环境变量
- ✅ **容器化部署** - Docker/K8s 支持
- ✅ **高并发优化** - 1000+ Workers

### 多 Agent 协作（新增）

- ✅ **Git 集成** - 文件边界控制和权限管理
- ✅ **文件锁** - 防止多 Agent 同时修改同一文件
- ✅ **分支隔离** - 每个 Agent 在独立分支工作
- ✅ **冲突检测** - 自动检测文件锁定和合并冲突
- ✅ **任务升级机制** - 三级冲突解决（自动重试 → 升级任务 → 人工）

**详细文档**:
- [Git 集成指南](docs/git-integration-guide.md)
- [Agent 任务升级机制](docs/agent-upgrade-mechanism.md)

---

## 📁 项目结构

```
AgentFlow/
├── python/                 # Python 版本（本地开发）
│   ├── agentflow/          # 核心包
│   │   ├── __init__.py     # 模块入口
│   │   ├── master.py       # Master 实现
│   │   ├── worker.py       # Worker 实现
│   │   ├── database.py     # 数据库层
│   │   └── cli.py          # 命令行工具
│   ├── requirements.txt    # Python 依赖
│   └── README.md           # Python 版本文档
│
├── golang/                 # Go 版本（云端部署）
│   ├── cmd/                # 命令行工具
│   │   ├── master/         # Master 入口
│   │   ├── worker/         # Worker 入口
│   │   └── oneshot/        # OneShot 入口
│   ├── internal/           # 内部包
│   │   ├── master/         # Master 实现
│   │   ├── worker/         # Worker 实现
│   │   ├── database/       # 数据库层
│   │   ├── executor/       # 执行器（HTTP/Claude）
│   │   └── config/         # 配置系统
│   ├── bin/                # 预编译二进制
│   ├── deployments/        # Docker/K8s 配置
│   ├── config.example.yaml # 配置示例
│   └── README.md           # Go 版本文档
│
├── docs/                   # 共享文档
│   ├── architecture.md     # 架构设计
│   ├── api.md             # API 文档
│   └── migration.md       # 迁移指南
│
├── skills/                 # Claude Code Skills
│   └── agentflow.md       # Skill 手册
│
└── README.md              # 本文件
```

---

## 📖 文档

### Python 版本文档
- [Python 版本 README](python/README.md) - 详细使用说明
- [安装指南](docs/installation.md#python-版本)
- [API 参考](docs/python-api.md)

### Go 版本文档
- [Go 版本 README](golang/README.md) - 详细使用说明
- [构建指南](golang/docs/BUILD_GUIDE.md)
- [配置参考](golang/config.example.yaml)

### 共享文档
- [系统架构](docs/architecture.md) - 整体架构设计
- [迁移指南](docs/migration.md) - 两个版本之间迁移
- [Skill 手册](skills/agentflow.md) - Claude Code 集成

---

## 🎯 使用场景示例

### 场景 1: 本地代码重构（Python）

```python
# 使用 Python 版本进行本地代码重构
from agentflow import Master, Worker

master = Master(port=8848)
worker = Worker(mode="auto")

# 创建重构任务
master.create_task(
    title="重构用户认证模块",
    description="""
打开项目 /path/to/project
使用 VSCode 查找所有使用 password 字段的地方
逐个文件进行重构
将 password 重命名为 hashed_password
运行测试确保没有破坏
""",
    group_name="refactoring"
)
```

### 场景 2: 云端批量处理（Go）

```bash
# 使用 Go 版本进行云端批量处理
# 1. 准备 1000 个文件处理任务
for i in {1..1000}; do
  curl -X POST http://cloud-master:8848/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{
      \"title\": \"处理文件 $i\",
      \"description\": \"shell:python process.py file_$i.txt\",
      \"group_name\": \"batch-processing\"
    }"
done

# 2. 启动 100 个 Go Workers 并行处理
kubectl scale deployment agentflow-worker --replicas=100
```

### 场景 3: 混合自动化

```python
# 本地 Python Master + 云端 Go Workers
from agentflow import Master

# 连接到云端 Go Workers
master = Master()
master.create_task(
    title="云端数据分析 + 本地报告生成",
    description="""
步骤 1: 云端执行数据聚合（group: cloud-workers）
  - 连接数据库
  - 执行聚合查询
  - 生成 CSV 报告

步骤 2: 本地生成可视化（group: local-gui）
  - 下载云端 CSV
  - 使用 VSCode 打开
  - 运行 Python 生成图表
  - 在浏览器中打开
""",
    group_name="hybrid"
)
```

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

**贡献指南**:
1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

详见 [贡献指南](docs/contributing.md)

---

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

---

## 🔗 相关链接

- **GitHub**: https://github.com/MoSiYuan/AgentFlow
- **文档**: https://github.com/MoSiYuan/AgentFlow/tree/feature/1.0.0/docs
- **Issue**: https://github.com/MoSiYuan/AgentFlow/issues

---

**版本**: v1.0.0
**更新**: 2026-01-22
**分支**: [feature/1.0.0](https://github.com/MoSiYuan/AgentFlow/tree/feature/1.0.0)
