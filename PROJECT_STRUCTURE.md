# AgentFlow 项目结构

本文档描述 AgentFlow 项目的完整目录结构和文件说明。

## 目录结构

```
agentflow-go/
├── cmd/                          # 命令行工具入口
│   ├── agentflow/                     # 统一 CLI 工具
│   │   └── main.go               # 主程序：init, master, worker, add, list, workers
│   ├── master/                   # Master 独立程序（可选）
│   │   └── main.go
│   └── worker/                   # Worker 独立程序（可选）
│       └── main.go
│
├── internal/                     # 内部包（不对外暴露）
│   ├── database/                 # 数据库层
│   │   ├── database.go           # 数据库操作 CRUD
│   │   ├── models.go             # 数据模型定义
│   │   └── schema.sql            # 数据库架构（嵌入式）
│   ├── master/                   # Master 实现
│   │   └── master.go             # Master 核心逻辑 + HTTP API
│   └── worker/                   # Worker 实现
│       └── worker.go             # Worker 核心逻辑 + 任务执行
│
├── pkg/                          # 公共库（可被外部使用）
│   ├── sqliteutil/               # SQLite 工具函数
│   │   └── util.go
│   └── httputil/                 # HTTP 工具函数
│       └── util.go
│
├── docs/                         # 项目文档
│   ├── GETTING_STARTED.md        # 快速入门
│   ├── ARCHITECTURE.md           # 架构设计
│   ├── API.md                    # API 文档
│   ├── WORKER_GROUPS.md          # Worker Group 说明
│   └── DEPLOYMENT.md             # 部署指南
│
├── scripts/                      # 实用脚本
│   ├── backup.sh                 # 数据库备份
│   ├── health_check.sh           # 健康检查
│   └── example_tasks.sh          # 示例任务创建
│
├── deployments/                  # 部署配置
│   ├── k8s/                      # Kubernetes 部署文件
│   │   ├── namespace.yaml
│   │   ├── master.yaml           # Master Deployment + Service
│   │   ├── worker.yaml           # Worker Deployment + HPA
│   │   └── ingress.yaml          # Ingress 配置
│   └── docker-compose.yml        # Docker Compose 配置
│
├── .github/                      # GitHub 配置
│   ├── workflows/                # CI/CD 工作流
│   │   └── ci.yml
│   └── ISSUE_TEMPLATE/           # Issue 模板
│       ├── bug.yml
│       └── feature.yml
│
├── bin/                          # 编译输出（gitignore）
│   ├── agentflow
│   ├── master
│   └── worker
│
├── go.mod                        # Go 模块定义
├── go.sum                        # 依赖锁定文件
├── Makefile                      # 构建脚本
├── Dockerfile                    # Docker 镜像构建
├── .dockerignore                 # Docker 忽略文件
├── .gitignore                    # Git 忽略文件
├── LICENSE                       # MIT 许可证
├── README.md                     # 项目说明
├── CONTRIBUTING.md               # 贡献指南
└── PROJECT_STRUCTURE.md          # 本文件
```

## 核心文件说明

### cmd/cpds/main.go

统一的 CLI 工具，支持以下命令：

- `init <db>` - 初始化数据库
- `master` - 启动 Master 服务
- `worker` - 启动 Worker
- `add <title>` - 添加任务
- `list` - 列出任务
- `workers` - 列出 Worker
- `version` - 显示版本

### internal/database/

数据库层，封装所有 SQLite 操作：

- `database.go` - CRUD 操作、锁机制、统计查询
- `models.go` - Task, Worker, TaskLog, GroupStat 结构体
- `schema.sql` - 表结构、索引、视图定义

### internal/master/

Master 服务实现：

- HTTP API (Gin 框架)
- 任务分发逻辑
- Worker 管理
- 本地 Worker 进程启动
- 请求日志中间件

### internal/worker/

Worker 实现：

- 任务轮询和执行
- Shell 命令和脚本执行
- 自动组检测
- 心跳机制
- 错误处理和重试

### docs/

详细文档：

- **GETTING_STARTED.md** - 5分钟快速上手
- **ARCHITECTURE.md** - 系统架构和设计原理
- **API.md** - 完整 API 参考
- **WORKER_GROUPS.md** - Worker Group 功能详解
- **DEPLOYMENT.md** - 生产环境部署指南

### scripts/

运维脚本：

- **backup.sh** - 数据库自动备份
- **health_check.sh** - 系统健康检查
- **example_tasks.sh** - 创建示例任务

### deployments/

部署配置：

- **k8s/** - Kubernetes 部署文件
  - Master Deployment + Service + PVC
  - Worker Deployment + HPA
  - Ingress 配置
- **docker-compose.yml** - Docker Compose 一键部署

## 数据流

### 单机模式

```
Agent → CLI (agentflow add) → SQLite → Local Worker → 执行 → SQLite
```

### 分布式模式

```
Agent → HTTP API → SQLite → Remote Worker 轮询 → 执行 → HTTP API → SQLite
```

## 构建流程

```bash
# 开发
make dev-init      # 初始化开发环境
make build         # 编译
make test          # 测试

# Docker
make docker-build  # 构建镜像
make docker-run    # 运行容器

# 生产
make release       # 构建多平台发布版本
```

## 依赖关系

```
cmd/cpds/main.go
    ├── internal/database
    ├── internal/master
    └── internal/worker
        └── internal/database
```

## 扩展点

### 添加新的 Worker 类型

1. 在 `internal/worker/worker.go` 添加检测逻辑
2. 更新 `WORKER_GROUPS.md` 文档

### 添加新的 API 端点

1. 在 `internal/master/master.go` 注册路由
2. 实现处理函数
3. 更新 `docs/API.md`

### 支持新数据库

1. 创建 `internal/database/postgres.go`
2. 实现相同接口
3. 添加配置选项

## 测试策略

```bash
# 单元测试
go test ./internal/database/...
go test ./internal/master/...
go test ./internal/worker/...

# 集成测试
go test ./tests/integration/...

# 端到端测试
go test ./tests/e2e/...
```

## 发布流程

1. 更新版本号
2. 运行完整测试
3. 构建 Docker 镜像
4. 创建 Git tag
5. 推送到 GitHub
6. 发布 Release

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 运行测试
5. 创建 Pull Request

详细步骤见 [CONTRIBUTING.md](CONTRIBUTING.md)
