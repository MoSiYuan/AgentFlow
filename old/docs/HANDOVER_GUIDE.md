# CPDS Git+SQLite混合架构 - 交接文档

## 📋 项目概述

**项目名称**: CPDS (Claude Parallel Development System)
**核心功能**: Git+SQLite混合文件锁和任务编排架构
**实现时间**: 2025-01-21 ~ 2025-01-22
**技术栈**: Go + Git + SQLite + Gin + Git LFS

---

## 🎯 核心功能模块

### 1. Git客户端封装 (`internal/git/client.go`)

**功能**:
- Git仓库操作（初始化、分支管理）
- Git LFS文件锁定
- 合并冲突检测
- 提交历史管理

**核心方法**:
```go
// 分支管理
CreateBranch(branchName string) error
CheckoutBranch(branchName string) error
DeleteBranch(branchName string, force bool) error
GetCurrentBranch() (string, error)

// 文件锁定（Git LFS）
LockFile(filePath, ownerID string) error
UnlockFile(filePath string) error
GetLockedFiles() ([]FileLock, error)

// 合并与冲突检测
MergeToMaster(branchName string) error
HasConflicts() bool
GetConflictedFiles() ([]string, error)
AbortMerge() error

// 提交管理
AddAll() error
Commit(message string) error
HasUncommittedChanges() bool
```

**依赖**:
- `github.com/sirupsen/logrus` (日志)
- 标准库: `os/exec`, `fmt`, `strings`, `time`

**文件大小**: 10KB

---

### 2. SQLite集成层 (`internal/database/git_integration.go`)

**功能**:
- Git任务管理（创建、分配、查询、更新）
- 文件边界管理
- 文件锁记录
- 冲突记录和处理
- 高权限Agent任务队列

**核心结构**:
```go
type GitTask struct {
    ID              string
    Title           string
    Description     string
    WorkerID        string
    GitBranch       string           // "worker-1/task-001"
    FileBoundaries  []GitFileBoundary
    Locks           []GitLock
    Status          string           // pending|assigned|in_progress|completed
}

type GitLock struct {
    ID         int64
    TaskID     string
    WorkerID   string
    FilePath   string
    LockType   string           // "read" | "write"
    AcquiredAt time.Time
    ReleasedAt *time.Time
    Status     string           // "active" | "released"
}

type GitConflict struct {
    ID           string
    TaskID       string
    WorkerID     string
    ConflictType string           // "git_merge" | "file_locked" | "boundary_overlap"
    FilePaths    []string
    Description  string
    Severity     string           // "low" | "medium" | "high" | "critical"
    Status       string           // "pending" | "resolving" | "resolved"
}
```

**核心方法**:
```go
CreateGitTables() error                     // 创建Git集成表
CreateGitTask(task *GitTask) error          // 创建Git任务
AssignTaskWithGit(task *GitTask, gitClient *git.GitClient) error
GetGitTask(taskID string) (*GitTask, error)
ReportGitConflict(conflict *GitConflict) error
ResolveGitConflict(conflictID, resolverID, resolution string) error
```

**依赖**:
- `github.com/jiangxiaolong/cpds-go/internal/git` (本地Git客户端)
- `database/sql`
- `encoding/json`
- 标准库: `fmt`, `time`

**文件大小**: 16KB

---

### 3. Master API端点 (`internal/master/handlers.go`)

**新增端点**:

| 端点 | 方法 | 功能 |
|------|------|------|
| `/api/tasks/create-with-git` | POST | 创建Git任务 |
| `/api/tasks/assign-git` | POST | 分配任务（创建分支+锁文件） |
| `/api/tasks/:task_id/git` | GET | 获取Git任务详情 |
| `/api/conflicts` | POST | 报告冲突 |
| `/api/conflicts` | GET | 查询待处理冲突 |
| `/api/conflicts/resolve` | POST | 解决冲突 |

**请求/响应示例**:
```json
// 创建Git任务
POST /api/tasks/create-with-git
{
  "task_id": "TASK-001",
  "title": "实现用户认证",
  "file_boundaries": [
    {
      "file_path": "src/auth/user.go",
      "line_start": 1,
      "line_end": 100,
      "lock_type": "write"
    }
  ]
}

// 响应
{
  "success": true,
  "data": {
    "task_id": "TASK-001",
    "git_branch": "pending/TASK-001",
    "files_locked": ["src/auth/user.go"]
  }
}
```

**依赖**:
- `github.com/gin-gonic/gin` (HTTP框架)
- `github.com/jiangxiaolong/cpds-go/internal/git`
- `github.com/jiangxiaolong/cpds-go/internal/database`
- `github.com/jiangxiaolong/cpds-go/internal/api`

---

### 4. Worker Git集成 (`internal/worker/git_worker.go`)

**功能**:
- 安全编辑检查（三层验证）
- 分支检出
- 提交和合并
- 冲突报告

**安全检查流程**:
```go
// 1. 检查文件边界（SQLite）
gitTask := gw.getGitTask(taskID)
if !gw.isInBoundary(filePath, gitTask.FileBoundaries) {
    return "file not in assigned boundaries"
}

// 2. 检查Git文件锁（Git LFS）
lockedFiles := gw.git.GetLockedFiles()
for _, lock := range lockedFiles {
    if lock.FilePath == filePath && lock.Owner.ID != gw.client.workerID {
        return "file locked by " + lock.Owner.Name
    }
}

// 3. 检查未提交变更
if gw.git.HasUncommittedChanges() {
    return "file has uncommitted changes"
}
```

**依赖**:
- `github.com/jiangxiaolong/cpds-go/internal/git`
- `github.com/jiangxiaolong/cpds-go/internal/database`
- `github.com/jiangxiaolong/cpds-go/internal/worker`
- `github.com/sirupsen/logrus`

**文件大小**: 5.7KB

---

## 📂 需要复制的文件清单

### 核心代码文件（3个，必须）

```
internal/git/
└── client.go                    (10KB) - Git客户端封装

internal/database/
└── git_integration.go            (16KB) - SQLite集成层

internal/worker/
└── git_worker.go                (5.7KB) - Worker Git集成
```

**总计**: ~32KB核心代码

### API类型定义（1个文件）

```
internal/api/
└── types.go                     - Git-Integrated Task类型
```

**需要添加的类型**:
- `CreateGitTaskRequest`
- `GitFileBoundaryRequest`
- `CreateGitTaskResponse`
- `GetGitTaskResponse`
- `ConflictListResponse`

### 测试脚本（3个，技术测试）

```
scripts/
├── quick_test.sh                 - 快速API测试
├── full_test.sh                  - 完整API功能测试
└── test_git_integration.sh    - Git集成测试
```

### 文档（6个，纯技术）

```
docs/
├── GIT_LOCK_ARCHITECTURE.md       - 架构设计文档
├── GIT_INTEGRATION_IMPLEMENTATION.md  - 实现文档
├── TESTING_GUIDE.md              - 测试指南
├── TEST_RESULTS.md                - 测试结果
└── COMPLETE_TEST_SUMMARY.md      - 完整测试总结
```

---

## 🎯 核心架构原则

### Git职责（文件级并发控制）

- ✅ Git LFS File Locks: 文件锁定
- ✅ Git Branch: 任务隔离
- ✅ Git Merge: 自动合并检测冲突
- ✅ Git Reflog: 变更历史追踪
- ✅ Git Blame: 行级责任追溯

### SQLite职责（元数据和状态管理）

- ✅ 任务分配和文件边界定义
- ✅ Worker状态和心跳
- ✅ 冲突记录（Git检测到的冲突）
- ✅ 高权限Agent任务队列
- ✅ 执行日志和审计

---

## 🔧 集成步骤指南

### 步骤1: 复制核心代码

```bash
# 假设您的项目路径是 /path/to/your-project

# 1. 复制Git客户端
mkdir -p /path/to/your-project/internal/git
cp /path/to/cpds-go/internal/git/client.go /path/to/your-project/internal/git/

# 2. 复制SQLite集成层
cp /path/to/cpds-go/internal/database/git_integration.go /path/to/your-project/internal/database/

# 3. 复制Worker集成（如果需要）
mkdir -p /path/to/your-project/internal/worker
cp /path/to/cpds-go/internal/worker/git_worker.go /path/to/your-project/internal/worker/
```

### 步骤2: 添加API类型

编辑您的项目文件 `/path/to/your-project/internal/api/types.go`，添加：

```go
// === Git-Integrated Task Types ===

type CreateGitTaskRequest struct {
    TaskID        string                       `json:"task_id" binding:"required"`
    Title         string                       `json:"title" binding:"required"`
    Description   string                       `json:"description"`
    Priority      string                       `json:"priority"`
    FileBoundaries []GitFileBoundaryRequest    `json:"file_boundaries" binding:"required"`
    DeploymentMode string                      `json:"deployment_mode"`
}

type GitFileBoundaryRequest struct {
    FilePath  string `json:"file_path" binding:"required"`
    LineStart int    `json:"line_start" binding:"required"`
    LineEnd   int    `json:"line_end" binding:"required"`
    LockType  string `json:"lock_type"`
}

type CreateGitTaskResponse struct {
    TaskID    string   `json:"task_id"`
    GitBranch string   `json:"git_branch"`
    Files     []string `json:"files_locked"`
    Message   string   `json:"message"`
}

type GetGitTaskResponse struct {
    GitTask        *database.GitTask
    ConflictedFiles []string
}

type ConflictListResponse struct {
    Conflicts []database.GitConflict
    Total     int
}
```

### 步骤3: 添加数据库表

在您的数据库初始化脚本中添加：

```sql
-- Git任务表
CREATE TABLE IF NOT EXISTS git_tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    worker_id TEXT,
    git_branch TEXT NOT NULL UNIQUE,
    file_boundaries TEXT,              -- JSON
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Git文件锁表
CREATE TABLE IF NOT EXISTS git_locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    lock_type TEXT NOT NULL,
    acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP,
    status TEXT DEFAULT 'active',
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);

-- Git冲突表
CREATE TABLE IF NOT EXISTS git_conflicts (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    conflict_type TEXT NOT NULL,
    file_paths TEXT,                  -- JSON数组
    description TEXT,
    severity TEXT DEFAULT 'medium',
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by TEXT,
    resolution TEXT,
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_git_tasks_status ON git_tasks(status);
CREATE INDEX IF NOT EXISTS idx_git_tasks_worker ON git_tasks(worker_id);
CREATE INDEX IF NOT EXISTS idx_git_locks_file ON git_locks(file_path);
CREATE INDEX IF NOT EXISTS idx_git_conflicts_status ON git_conflicts(status);
```

### 步骤4: 安装依赖

```bash
cd /path/to/your-project

# 添加必要的Go依赖
go get github.com/sirupsen/logrus
go get github.com/gin-gonic/gin
go get github.com/mattn/go-sqlite3
```

---

## 🔄 复用现有项目公用库的优化建议

### 1. 日志库

**当前实现**: `github.com/sirupsen/logrus`

**优化建议**:
```go
// 如果项目使用其他日志库（如zap, log15），创建适配层

type Logger interface {
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Debug(args ...interface{})
}

// 修改 internal/git/client.go
type GitClient struct {
    repoPath string
    logger   Logger  // 使用接口而不是具体实现
}
```

### 2. 数据库连接池

**优化建议**:
```go
// 如果项目使用ORM（如GORM）
type GitTaskModel struct {
    gorm.Model
    ID          string
    Title       string
    Description string
    WorkerID    sql.NullString
    GitBranch   string
    FileBoundaries string  // JSON
    Status      string
}

// 在git_integration.go中使用GORM
func (db *Database) CreateGitTaskWithGorm(task *GitTask) error {
    model := &GitTaskModel{...}
    return db.db.Create(model).Error
}
```

### 3. 配置管理

**优化建议**:
```go
// 使用项目的配置系统
type GitConfig struct {
    RepoPath    string `yaml:"repo_path"`
    LFSEnabled bool   `yaml:"lfs_enabled"`
    AutoLock    bool   `yaml:"auto_lock"`
}

// 在git_integration.go中
func (db *Database) AssignTaskWithGitConfig(task *GitTask, cfg GitConfig) error {
    gitClient := git.NewGitClient(cfg.RepoPath, db.logger)
    // ...
}
```

---

## 🎛️ 配置和部署

### 环境变量

```bash
# Git配置
export GIT_REPO_PATH=/path/to/your/project
export GIT_LFS_ENABLED=true
export GIT_AUTO_LOCK=true

# SQLite配置
export DB_PATH=./data/app.db
export GIT_INTEGRATION_ENABLED=true
```

### Docker配置

```dockerfile
FROM golang:1.21-alpine

# 安装Git和Git LFS
RUN apk add --no-cache git git-lfs

WORKDIR /app
COPY . .
RUN go build -o server
CMD ["./server"]
```

---

## 🧪 测试验证

### 单元测试

```bash
# 测试Git客户端
go test -v ./internal/git/...

# 测试SQLite集成
go test -v ./internal/database/... -run TestGit
```

### 集成测试

```bash
# 运行快速测试
./scripts/quick_test.sh

# 运行完整测试
./scripts/full_test.sh
```

---

## ⚠️ 注意事项

### 1. 命名冲突

- CPDS使用 `internal/git/` 作为包名
- 如果冲突，建议重命名为 `internal/scm/git/` 或 `internal/vcs/git/`

### 2. 数据库表名

- Git相关表都以 `git_` 前缀
- 避免与现有表名冲突
- 如果冲突，可以改为 `scm_tasks`, `scm_locks` 等

### 3. API路径

- CPDS使用 `/api/tasks/*` 路径
- 如果需要，可以改为 `/api/scm/tasks/*`

---

## 🚀 快速开始示例

### 最小集成示例

```go
package main

import (
    "fmt"
    "github.com/jiangxiaolong/cpds-go/internal/git"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    gitClient := git.NewGitClient(".", logger)

    // 创建分支
    err := gitClient.CreateBranch("feature/task-001")
    if err != nil {
        logger.Fatal(err)
    }

    // 锁定文件
    err = gitClient.LockFile("src/file.go", "worker-001")
    if err != nil {
        logger.Warnf("File lock failed: %v", err)
    }

    fmt.Println("Git task started!")
}
```

---

## 📚 技术文档

### 核心文档

1. **docs/GIT_LOCK_ARCHITECTURE.md**
   - Git+SQLite混合架构设计原则
   - 职责分离说明
   - 文件锁机制详解

2. **docs/GIT_INTEGRATION_IMPLEMENTATION.md**
   - 完整实现文档
   - API端点说明
   - 数据表结构

3. **docs/TESTING_GUIDE.md**
   - 测试指南
   - 测试场景说明
   - 故障排查

4. **docs/TEST_RESULTS.md**
   - 测试结果
   - 性能指标
   - 功能覆盖

---

## 📞 技术支持

### 常见问题

**Q1: 如何与现有任务系统集成？**
A: 在现有任务系统中添加Git元数据字段，或使用`git_integration.go`作为独立的Git任务管理器。

**Q2: 如何处理数据库迁移？**
A: 使用提供的SQL脚本创建表，然后启动应用自动迁移。

**Q3: 如何调试Git操作？**
A: 使用`logrus.SetLevel(logrus.DebugLevel)`查看详细日志。

**Q4: Git LFS是否必须？**
A: 不是必须的。系统有降级模式，Git LFS不可用时使用模拟锁定。

**Q5: 支持哪些Git操作？**
A: 所有标准Git操作：分支、提交、合并、冲突检测等。

---

## 📦 完整复制命令

```bash
# 1. 复制核心代码
cp internal/git/*.go <your-project>/internal/git/
cp internal/database/git_integration.go <your-project>/internal/database/
cp internal/worker/git_worker.go <your-project>/internal/worker/

# 2. 复制API类型（需要手动编辑）
# 打开 internal/api/types.go，复制Git-Integrated Task Types部分

# 3. 复制文档
mkdir -p <your-project>/docs/cpds
cp docs/GIT_LOCK_ARCHITECTURE.md docs/GIT_INTEGRATION_IMPLEMENTATION.md <your-project>/docs/cpds/

# 4. 复制测试脚本
cp scripts/quick_test.sh scripts/full_test.sh <your-project>/scripts/

# 5. 安装依赖
go get github.com/sirupsen/logrus
go get github.com/gin-gonic/gin
go get github.com/mattn/go-sqlite3

# 6. 创建数据库表
# 在您的数据库初始化脚本中添加docs中提供的SQL
```

---

**文档版本**: v2.0 (已删除故事和评审相关内容)
**最后更新**: 2025-01-22
**项目**: CPDS (Claude Parallel Development System) - Git+SQLite混合架构

**核心功能**: 文件级并发控制、任务编排、冲突检测

---

**这是纯技术的Git+SQLite混合架构实现，可独立用于任何并发控制和版本管理场景！**
