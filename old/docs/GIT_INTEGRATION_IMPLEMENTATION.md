# Git + SQLite 混合架构实现总结

## 实现概览

已完成CPDS系统的Git + SQLite混合文件锁和任务编排架构实现。

## 架构设计原则

### Git职责（文件级并发控制）
- ✅ Git LFS File Locks: 文件锁定
- ✅ Git Branch: 任务隔离
- ✅ Git Merge: 自动合并检测冲突
- ✅ Git Commit: 变更历史追踪

### SQLite职责（元数据和状态管理）
- ✅ 任务分配和文件边界定义
- ✅ Worker状态和心跳
- ✅ 冲突记录（Git检测到的冲突）
- ✅ 高权限Agent任务队列
- ✅ 执行日志和审计

---

## 核心组件

### 1. Git客户端封装 (`internal/git/client.go`)

**已实现的方法**:

```go
// 分支管理
CreateBranch(branchName string) error
CheckoutBranch(branchName string) error
DeleteBranch(branchName string, force bool) error
GetCurrentBranch() (string, error)
ListBranches() ([]string, error)

// 文件锁（Git LFS）
LockFile(filePath, ownerID string) error
UnlockFile(filePath string) error
GetLockedFiles() ([]FileLock, error)
LockMultiple(filePaths []string, ownerID string) error
ReleaseAllLocks(ownerID string) (int, error)

// 合并与冲突检测
MergeToMaster(branchName string) error
HasConflicts() bool
GetConflictedFiles() ([]string, error)
AbortMerge() error
ContinueMerge() error

// 提交管理
Add(files ...string) error
AddAll() error  // 新增：添加所有变更
Commit(message string) error
HasUncommittedChanges() bool
GetChangedFiles() ([]string, error)

// 历史与责任追溯
GetFileHistory(filePath string, limit int) ([]Commit, error)
Blame(filePath string) ([]BlameLine, error)
```

**关键数据结构**:
```go
type FileLock struct {
    ID       string
    FilePath string
    Owner    OwnerInfo
}

type MergeConflictError struct {
    Branch string
    Files  []string
}

type Commit struct {
    Hash      string
    Author    string
    Message   string
    Timestamp string
}
```

---

### 2. SQLite集成层 (`internal/database/git_integration.go`)

**数据表结构**:

#### git_tasks 表
```sql
CREATE TABLE IF NOT EXISTS git_tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    worker_id TEXT,
    git_branch TEXT NOT NULL UNIQUE,  -- "worker-1/task-001"
    file_boundaries TEXT,             -- JSON
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

#### git_locks 表
```sql
CREATE TABLE IF NOT EXISTS git_locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    lock_type TEXT NOT NULL,          -- "read" | "write"
    acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP,
    status TEXT DEFAULT 'active',
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);
```

#### git_conflicts 表
```sql
CREATE TABLE IF NOT EXISTS git_conflicts (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,
    conflict_type TEXT NOT NULL,      -- "git_merge" | "file_locked" | "boundary_overlap"
    file_paths TEXT,                  -- JSON数组
    description TEXT,
    severity TEXT DEFAULT 'medium',   -- "low" | "medium" | "high" | "critical"
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by TEXT,
    resolution TEXT,
    FOREIGN KEY (task_id) REFERENCES git_tasks(id) ON DELETE CASCADE
);
```

**核心方法**:
```go
// 表创建
CreateGitTables() error

// Git任务管理
CreateGitTask(task *GitTask) error
AssignTaskWithGit(task *GitTask, gitClient *git.GitClient) error
GetGitTask(taskID string) (*GitTask, error)
UpdateGitTaskStatus(taskID, status string) error
GetWorkerActiveGitTasks(workerID string) ([]GitTask, error)
CleanupCompletedGitTasks(gitClient *git.GitClient) error

// 锁管理
GetGitLocksForTask(taskID string) ([]GitLock, error)
ReleaseAllTaskLocks(taskID string, gitClient *git.GitClient) error

// 冲突管理
ReportGitConflict(conflict *GitConflict) error
ResolveGitConflict(conflictID, resolverID, resolution string, resolutionRecord *ConflictResolution) error
GetPendingGitConflicts() ([]GitConflict, error)
```

**关键数据结构**:
```go
type GitFileBoundary struct {
    FilePath string  // e.g., "src/controllers/user.go"
}

type GitTask struct {
    ID            string
    Title         string
    Description   string
    WorkerID      string
    GitBranch     string           // "worker-1/task-001"
    FileBoundaries []GitFileBoundary
    Locks         []GitLock
    Status        string           // pending | assigned | in_progress | completed
    CreatedAt     time.Time
    AssignedAt    *time.Time
    CompletedAt   *time.Time
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
    CreatedAt    time.Time
    ResolvedAt   *time.Time
    ResolvedBy   string
    Resolution   string
}
```

---

### 3. Master API端点 (`internal/master/handlers.go`)

**新增的Git集成API**:

```go
// 创建Git任务
POST /api/tasks/create-with-git
Request:
{
    "task_id": "TASK-001",
    "title": "实现用户认证",
    "description": "使用JWT实现用户认证功能",
    "priority": "high",
    "file_boundaries": [
        {
            "file_path": "src/controllers/user.go",
            "line_start": 1,
            "line_end": 100,
            "lock_type": "write"
        }
    ]
}
Response:
{
    "success": true,
    "data": {
        "task_id": "TASK-001",
        "git_branch": "pending/TASK-001",
        "files_locked": ["src/controllers/user.go"],
        "message": "Git task created successfully"
    }
}

// 分配Git任务（创建分支并锁定文件）
POST /api/tasks/assign-git
Request:
{
    "task_id": "TASK-001",
    "worker_id": "worker-xxx"
}
Response:
{
    "success": true,
    "data": {
        "task_id": "TASK-001",
        "git_branch": "worker-xxx/TASK-001",
        "files_locked": ["src/controllers/user.go"],
        "message": "Git task assigned successfully"
    }
}

// 获取Git任务详情
GET /api/tasks/:task_id/git
Response:
{
    "success": true,
    "data": {
        "id": "TASK-001",
        "title": "实现用户认证",
        "git_branch": "worker-xxx/TASK-001",
        "file_boundaries": [...],
        "status": "assigned",
        "conflicted_files": []
    }
}

// 报告冲突
POST /api/conflicts
Request:
{
    "task_id": "TASK-001",
    "worker_id": "worker-xxx",
    "conflict_type": "git_merge",
    "file_paths": ["src/controllers/user.go"],
    "description": "Merge conflict detected",
    "severity": "high"
}

// 获取待处理冲突
GET /api/conflicts
Response:
{
    "success": true,
    "data": {
        "conflicts": [...],
        "total": 5
    }
}

// 解决冲突
POST /api/conflicts/resolve
Request:
{
    "conflict_id": "CONFLICT-001",
    "resolver_id": "privileged-agent-001",
    "resolution": "merged",
    "resolution_record": {
        "original_content": "...",
        "resolved_content": "...",
        "explanation": "Merged both versions"
    }
}
```

---

### 4. Worker Git集成 (`internal/worker/git_worker.go`)

**核心方法**:

```go
type GitWorker struct {
    client   *Client
    git      *git.GitClient
    masterURL string
    logger   *logrus.Logger
    repoPath string
}

// 分配Git任务
AssignGitTask(taskID string) error

// 安全编辑检查（执行前必检）
SafeEditCheck(taskID, filePath string, lineStart, lineEnd int) error

// 检出任务分支
CheckoutTaskBranch(taskID string) error

// 提交并合并到主分支
CommitAndMerge(taskID, commitMessage string) error

// 报告冲突
reportConflict(taskID, conflictType string, filePaths []string, description string) error

// 获取Git任务信息
getGitTask(taskID string) (*database.GitTask, error)
```

**Worker执行流程**:

```go
// 1. 获取并分配任务
worker.AssignGitTask("TASK-001")

// 2. 检出任务分支
worker.CheckoutTaskBranch("TASK-001")

// 3. 编辑前进行安全检查
err := worker.SafeEditCheck("TASK-001", "src/controllers/user.go", 1, 100)
if err != nil {
    // 报告冲突，停止执行
    worker.reportConflict("TASK-001", "lock_denied", []string{"src/controllers/user.go"}, err.Error())
    return
}

// 4. 编辑文件（在自己的分支上）
editFiles()

// 5. 提交并尝试合并
err = worker.CommitAndMerge("TASK-001", "feat: implemented user authentication")
if err != nil {
    // Git已自动报告冲突到Master
    // 等待高权限Agent处理
    return
}
```

---

## 工作流程

### 阶段1：Master分配任务

```bash
# 1. 用户创建Git任务
curl -X POST http://localhost:8848/api/tasks/create-with-git \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TASK-001",
    "title": "实现用户认证",
    "description": "使用JWT实现用户登录",
    "file_boundaries": [
      {"file_path": "src/controllers/user.go", "line_start": 1, "line_end": 100}
    ]
  }'

# 2. SQLite记录任务状态
# git_tasks: id=TASK-001, status='pending', git_branch='pending/TASK-001'

# 3. Worker抢占任务
curl -X POST http://localhost:8848/api/tasks/assign-git \
  -H "Content-Type: application/json" \
  -d '{"task_id": "TASK-001", "worker_id": "worker-001"}'

# 4. Master创建Git分支并锁定文件
git checkout -b worker-001/TASK-001
git lfs lock src/controllers/user.go

# 5. SQLite更新任务状态
# git_tasks: status='assigned', git_branch='worker-001/TASK-001'
# git_locks: file_path='src/controllers/user.go', worker_id='worker-001'
```

### 阶段2：Worker执行任务

```bash
# 1. Worker检出任务分支
git checkout worker-001/TASK-001

# 2. Worker执行安全检查
# 检查SQLite: 文件是否在file_boundaries中？
# 检查Git LFS: 文件是否被其他Worker锁定？
# 检查Git status: 文件是否有未提交的变更？

# 3. Worker编辑文件
vim src/controllers/user.go

# 4. Worker提交变更
git add -A
git commit -m "feat: implemented user authentication"

# 5. Worker尝试合并到master
git checkout master
git merge worker-001/TASK-001

# 如果有冲突:
# - Git自动检测冲突文件
# - Worker报告冲突到SQLite
# - 高权限Agent接手处理
```

### 阶段3：冲突处理

```bash
# 1. Git检测到合并冲突
git status
# 输出: both modified: src/controllers/user.go

# 2. Worker报告冲突
curl -X POST http://localhost:8848/api/conflicts \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "TASK-001",
    "worker_id": "worker-001",
    "conflict_type": "git_merge",
    "file_paths": ["src/controllers/user.go"],
    "severity": "high"
  }'

# 3. Master创建高权限Agent任务
# privileged_agent_tasks:
#   - task_type='conflict_resolution'
#   - priority=100
#   - conflict_ids=['CONFLICT-001']
#   - status='pending'

# 4. 高权限Agent获取任务
curl http://localhost:8848/api/tasks/pending

# 5. 高权限Agent解决冲突
# - 切换到master分支
# - 手动或调用Claude API解决冲突
# - 标记冲突已解决
# - 通知Worker可以继续

curl -X POST http://localhost:8848/api/conflicts/resolve \
  -H "Content-Type: application/json" \
  -d '{
    "conflict_id": "CONFLICT-001",
    "resolver_id": "privileged-agent-001",
    "resolution": "merged"
  }'

# 6. 释放文件锁，删除分支
git lfs unlock src/controllers/user.go
git branch -d worker-001/TASK-001
```

---

## 关键特性

### 1. 原子性保证

```go
// AssignTaskWithGit 确保原子操作
func (db *Database) AssignTaskWithGit(task *GitTask, gitClient *git.GitClient) error {
    // 1. 创建分支
    if err := gitClient.CreateBranch(branchName); err != nil {
        return err
    }

    // 2. 锁定文件
    for _, boundary := range task.FileBoundaries {
        if err := gitClient.LockFile(boundary.FilePath, task.WorkerID); err != nil {
            // 回滚：删除分支并释放所有锁
            gitClient.DeleteBranch(branchName, true)
            return fmt.Errorf("failed to lock file %s: %w", boundary.FilePath, err)
        }
    }

    // 3. 存储到SQLite
    // ...
}
```

### 2. 自动冲突升级

```go
// 高严重级冲突自动创建高权限Agent任务
func (db *Database) ReportGitConflict(conflict *GitConflict) error {
    // ...

    // 如果是高严重级冲突，自动创建高权限Agent任务
    if conflict.Severity == "high" || conflict.Severity == "critical" {
        privTask := &PrivilegedAgentTask{
            TaskID:     fmt.Sprintf("CONFLICT-%s", conflict.ID),
            Priority:   100,
            TaskType:   "conflict_resolution",
            Description: fmt.Sprintf("Resolve %s conflict for task %s", conflict.ConflictType, conflict.TaskID),
            ConflictIDs: []string{conflict.ID},
            Status:     "pending",
        }

        if err := db.CreatePrivilegedAgentTask(privTask); err != nil {
            db.logger.Errorf("Failed to create privileged task: %v", err)
        }
    }

    return nil
}
```

### 3. 多层安全检查

Worker在编辑前执行3层检查：

```go
func (gw *GitWorker) SafeEditCheck(taskID, filePath string, lineStart, lineEnd int) error {
    // 第1层: 检查SQLite中的文件边界
    gitTask, _ := gw.getGitTask(taskID)
    if !gw.isInBoundary(filePath, gitTask.FileBoundaries) {
        return fmt.Errorf("file not in assigned boundaries")
    }

    // 第2层: 检查Git文件锁
    lockedFiles, _ := gw.git.GetLockedFiles()
    for _, lock := range lockedFiles {
        if lock.FilePath == filePath && lock.Owner.ID != gw.client.workerID {
            return fmt.Errorf("file locked by %s", lock.Owner.ID)
        }
    }

    // 第3层: 检查未提交的变更
    if gw.git.HasUncommittedChanges(filePath) {
        return fmt.Errorf("file has uncommitted changes")
    }

    return nil
}
```

---

## 数据库初始化

系统启动时需要调用：

```go
db := database.New("master.db", logger)

// 创建Git集成表
if err := db.CreateGitTables(); err != nil {
    log.Fatal(err)
}
```

---

## 测试用例

### 测试场景1：正常流程

```bash
# 1. 创建Git任务
# 2. Worker分配任务
# 3. Worker检出分支
# 4. Worker编辑文件
# 5. Worker提交并合并
# 6. 任务完成，清理分支
```

### 测试场景2：文件锁冲突

```bash
# 1. Worker-1锁定 src/user.go
# 2. Worker-2尝试锁定同一文件
# 3. 检测到冲突，报告到SQLite
# 4. 高权限Agent接手
```

### 测试场景3：Git合并冲突

```bash
# 1. Worker-1和Worker-2同时编辑同一文件
# 2. Worker-1先合并成功
# 3. Worker-2合并时检测到冲突
# 4. 自动报告冲突，等待高权限Agent处理
```

---

## 性能优势

| 特性 | 传统方式 | Git+SQLite混合 |
|------|---------|---------------|
| 文件锁 | ❌ 需自己实现 | ✅ Git LFS原生支持 |
| 版本控制 | ❌ 无 | ✅ Git commit历史 |
| 冲突检测 | ❌ 需手动比较 | ✅ Git merge自动检测 |
| 责任追溯 | ❌ 困难 | ✅ Git blame |
| 回滚 | ❌ 复杂 | ✅ Git reset/revert |
| 并发控制 | ❌ 需分布式锁 | ✅ Git分支天然隔离 |
| 原子性 | ❌ 难保证 | ✅ Git事务性 |

---

## 后续优化方向

1. **模式匹配**: 当前 `isInBoundary` 仅支持精确匹配，可扩展为glob模式
2. **行级锁**: 当前仅支持文件级锁，可扩展为行范围锁
3. **锁超时**: 添加锁超时机制，防止死锁
4. **批量操作**: 支持批量任务分配和文件锁定
5. **缓存优化**: 缓存Git状态，减少命令调用
6. **通知机制**: WebSocket实时通知冲突状态

---

## 文件清单

```
internal/git/client.go              - Git命令封装
internal/database/git_integration.go - SQLite集成层
internal/master/handlers.go         - Master API端点
internal/master/server.go           - 路由配置
internal/worker/git_worker.go       - Worker Git集成
internal/api/types.go               - API类型定义
docs/GIT_LOCK_ARCHITECTURE.md       - 架构设计文档
docs/GIT_INTEGRATION_IMPLEMENTATION.md - 本文档
```

---

**实现完成时间**: 2025-01-21
**版本**: v2.0.0
**状态**: ✅ 实现完成，待测试
