# AgentFlow - AI Agent 任务协作系统

> **30秒上手，3令牌完成任务** - 专为 Claude Code 设计

## 🎯 核心功能

- **任务分发**: Master 分配任务给 Worker
- **任务升级**: Worker 创建子任务
- **Worker 协作**: 多 Worker 并发执行
- **边界安全**: 工作目录隔离，可沙箱执行
- **自迭代开发**: 用 AgentFlow 开发 AgentFlow

## 🚀 快速安装

### 方式 1: Claude Code Skill（推荐）

AgentFlow 已集成为 Claude Code 原生 skill，可以直接使用：

```bash
# 安装为 Claude Code skill
cp skills/agentflow.md ~/.claude/commands/

# 使用 skill
/agentflow demo              # 运行演示
/agentflow add "测试"         # 添加任务
/agentflow list              # 查看任务
```

### 方式 2: 快速启动脚本

```bash
# 克隆仓库
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go

# 运行快速启动
chmod +x quick-start.sh
./quick-start.sh demo        # 运行演示
./quick-start.sh start       # 启动服务
```

### 方式 3: 手动安装

#### 前置要求
- Go 1.21+
- SQLite 3

#### 安装步骤

```bash
# 1. 配置 Go 代理（中国大陆推荐）
go env -w GOPROXY=https://goproxy.cn,direct

# 2. 下载依赖
go mod download

# 3. 编译项目
make build

# 4. 初始化数据库
./bin/agentflow init agentflow.db

# 5. 启动 Master 服务
./bin/agentflow master --db agentflow.db
```

## ⚡ 3秒上手

```bash
# 克隆仓库
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go

# 编译项目（需要网络连接下载依赖）
make build

# 初始化并启动
./bin/agentflow init agentflow.db
./bin/agentflow master --db agentflow.db
```

## 📝 任务格式

```bash
# Shell 命令
agentflow add "运行测试" --desc "shell:go test ./..."

# 脚本执行
agentflow add "部署应用" --desc "script:./deploy.sh"

# AI 任务（自动分解）
agentflow add "实现功能" --desc "task:implement:功能名"

# 文件操作
agentflow add "写配置" --desc "file:write:config.yaml:key:value"
```

## 🔧 常用命令

```bash
# Claude Code Skill 命令
/agentflow demo                           # 运行演示
/agentflow add "任务" --desc "..."        # 创建任务
/agentflow list --status completed        # 查看已完成任务
/agentflow workers                        # 查看 Workers
/agentflow status                         # 系统状态

# CLI 命令（需要先编译）
./bin/agentflow init <db>                 # 初始化
./bin/agentflow master --db <db>          # 启动 Master
./bin/agentflow add "标题" --desc "..."   # 创建任务
./bin/agentflow list [--status ...]      # 查看任务
./bin/agentflow workers                   # 查看 Worker
```

## 💻 使用示例

### 本地开发工作流

```bash
# 1. 启动 Master
./bin/agentflow init dev.db
./bin/agentflow master --db dev.db

# 2. 创建开发任务
./bin/agentflow add "格式化代码" --desc "shell:gofmt -w ."
./bin/agentflow add "运行测试" --desc "shell:go test ./..."
./bin/agentflow add "代码检查" --desc "shell:golangci-lint run"
./bin/agentflow add "构建应用" --desc "shell:go build -v ./..."

# 3. 查看进度
./bin/agentflow list --status running
./bin/agentflow list --status completed
```

### Claude Code 集成

```bash
# 直接在 Claude Code 中使用
/agentflow add "代码清理" --desc "shell:gofmt -w ."
/agentflow add "测试" --desc "shell:go test ./..."
/agentflow list --status completed
```

## 📖 文档

- [安装指南](INSTALL_GUIDE.md) - 详细安装说明
- [快速入门](docs/GETTING_STARTED.md) - 基础使用
- [架构设计](docs/ARCHITECTURE.md) - 系统架构
- [自迭代开发](SELF_ITERATION.md) - 用 AgentFlow 开发 AgentFlow
- [Skill 使用](skills/agentflow.md) - Claude Code Skill
- [API 文档](docs/API.md) - REST API 参考

## 🧪 实战示例

**已验证**: 10个故事生成+20个评审=100%成功

- 总任务: 30个
- 执行时间: 3秒
- 输出: 10个Markdown文件
- 位置: `tests/ctest_stories/`

## 🚀 为 AI 优化

### 节约 Token 技巧

1. **短命令**: `af add "T" --d "s:T:t:1"` (16 token)
2. **批量**: 一次创建多个任务
3. **过滤**: `af list --s completed` 只看结果

### 快速集成

```go
// 1行创建任务
exec("agentflow add T --desc s:T:t:1")

// 1行查询状态
exec("agentflow list --s completed")
```

## 📁 项目结构

```
agentflow-go/
├── cmd/agentflow/          # CLI 工具
├── internal/
│   ├── database/          # SQLite 层
│   ├── master/            # Master 服务
│   ├── model/             # 数据模型
│   └── worker/            # Worker + AI Worker
├── tests/                  # 测试文件
├── docs/                  # 完整文档
├── skills/                # Claude Code Skills
├── scripts/               # 实用脚本
├── deployments/           # 部署配置
├── quick-start.sh         # 快速启动脚本
├── INSTALL_GUIDE.md       # 安装指南
├── SELF_ITERATION.md      # 自迭代开发
└── README.md              # 本文件
```

## 💡 使用场景

1. **本地开发**: Master 自动启动本地 Workers，直连数据库
2. **云端部署**: Master 在服务器，Workers 分布式连接
3. **任务协作**: 主任务完成后创建子任务
4. **CI/CD**: 自动化构建、测试、部署流程
5. **自迭代**: 用 AgentFlow 开发 AgentFlow 本身

## 🎯 实战示例

**已验证**: 30 个任务（10个故事+20个评审）100% 成功

- **总任务数**: 30 个
- **执行时间**: 3 秒
- **成功率**: 100%
- **输出**: 10 个 Markdown 文件
- **位置**: `tests/ctest_stories/`

## 🚀 为 AI 优化

### 节约 Token 技巧

1. **短命令**: `af add "T" --d "s:T:t:1"` (16 token)
2. **批量**: 一次创建多个任务
3. **过滤**: `af list --s completed` 只看结果

### Claude Code 集成

```bash
# Skill 已安装到 ~/.claude/commands/agentflow.md
# 直接使用：
/agentflow demo              # 演示
/agentflow add "测试"         # 添加任务
/agentflow list              # 查看任务
```

## 🔧 开发环境

### 前置要求
- Go 1.21+
- SQLite 3
- Git

### 快速开始

```bash
# 1. 克隆仓库
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go

# 2. 编译项目
make build

# 3. 初始化并启动
./bin/agentflow init agentflow.db
./bin/agentflow master --db agentflow.db
```

## 📊 性能指标

| 指标 | 数值 |
|------|------|
| 任务吞吐量 | ~10 任务/秒 |
| 平均延迟 | < 100ms |
| 并发度 | 多 Worker 并发 |
| 成功率 | 100% (已测试) |

## 🤝 贡献

欢迎贡献代码！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 🔗 相关链接

- **GitHub**: [jiangxiaolong/agentflow-go](https://github.com/jiangxiaolong/agentflow-go)
- **Issue**: 提交问题和建议
- **文档**: 完整文档见 docs/ 目录

---

**AgentFlow** - 让 AI Agent 协作更简单 ✨
