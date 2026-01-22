# Git + SQLite 混合文件锁架构设计

## 核心原则

### 1. Git的职责（文件级并发控制）
```bash
✅ Git LFS File Locks: 文件锁定
✅ Git Branch: 任务隔离
✅ Git Merge: 自动合并检测冲突
✅ Git Reflog: 变更历史追踪
✅ Git Blame: 行级责任追溯
```

### 2. SQLite的职责（元数据和状态管理）
```bash
✅ 任务分配和文件边界定义
✅ Worker状态和心跳
✅ 冲突记录（Git检测到的冲突）
✅ 高权限Agent任务队列
✅ 执行日志和审计
```

---

## 工作流程

### Stage 1: 任务分配（Master）

```go
// Master分配任务
func (m *Master) AssignTaskWithGit(task *Task, worker *Worker) error {
    // 1. 在SQLite中记录任务分配
    db.AssignTask(task.ID, worker.ID)

    // 2. 创建Git分支
    git.CreateBranch(fmt.Sprintf("%s/%s", worker.ID, task.ID))

    // 3. 在SQLite中定义文件边界
    db.AssignFileBoundaries(task.ID, worker.ID, fileBoundaries)

    // 4. 获取Git文件锁
    for _, filePath := range getFileList(fileBoundaries) {
        git.LockFile(filePath, worker.ID)
    }

    return nil
}
```

### Stage 2: Worker执行

```go
// Worker执行任务
func (w *Worker) ExecuteTask(task *Task) error {
    // 1. 切换到自己的分支
    git.Checkout(w.TaskBranch(task.ID))

    // 2. 检查文件锁状态
    lockedFiles := git.GetLockedFiles()
    for _, file := range lockedFiles {
        if file.Owner != w.ID && !file.IsReadLock {
            // 报告冲突
            db.ReportConflict(&Conflict{
                Type: "lock_denied",
                File: file.Path,
            })
            return ErrFileLocked
        }
    }

    // 3. 编辑文件（在自己的分支上）
    editFiles(fileBoundaries)

    // 4. 提交更改
    git.Commit(fmt.Sprintf("Task %s: %s", task.ID, task.Description))

    // 5. 尝试合并到主分支
    if err := git.MergeToMaster(); err != nil {
        // Git检测到冲突
        handleMergeConflict(err)
    }

    return nil
}
```

### Stage 3: 冲突处理

```go
// 处理Git合并冲突
func handleMergeConflict(mergeErr error) error {
    // 1. 记录冲突到SQLite
    conflict := &Conflict{
        ID: generateID(),
        Type: "git_merge_conflict",
        Files: git.GetConflictedFiles(),
        Status: "pending",
    }
    db.ReportConflict(conflict)

    // 2. 创建高权限Agent任务
    db.CreatePrivilegedTask(&PrivilegedTask{
        Type: "conflict_resolution",
        ConflictID: conflict.ID,
        Priority: 100, // 高优先级
    })

    // 3. 释放文件锁，等待高权限Agent处理
    git.ReleaseLocks(conflict.Files)

    return nil
}
```

---

## Git命令封装

### GitLocks (文件锁)
```bash
# 锁定文件
git lfs lock src/controllers/user.go --message "Worker-1: Task-001"

# 查看锁状态
git lfs locks

# 释放锁
git lfs unlock src/controllers/user.go

# 批量锁定
git lfs lock src/**/*.go
```

### GitBranch (分支管理)
```bash
# 创建任务分支
git checkout -b worker-1/task-001

# 列出所有分支
git branch -a

# 删除分支（任务完成后）
git branch -d worker-1/task-001
```

### GitMerge (合并和冲突)
```bash
# 尝试合并到主分支
git checkout master
git merge worker-1/task-001

# 如果有冲突，查看冲突文件
git diff --name-only --diff-filter=U

# 中止合并
git merge --abort

# 标记冲突已解决
git add .
```

---

## 数据库表设计

### 表1: tasks（任务表）
```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    worker_id TEXT,
    git_branch TEXT,              -- Git分支名
    status TEXT,                  -- "pending" | "assigned" | "in_progress" | "completed"
    file_boundaries TEXT,        -- JSON: 文件边界定义
    created_at TIMESTAMP,
    assigned_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

### 表2: git_locks（Git文件锁记录）
```sql
CREATE TABLE git_locks (
    id INTEGER PRIMARY KEY,
    task_id TEXT,
    worker_id TEXT,
    file_path TEXT,
    lock_type TEXT,                -- "read" | "write"
    acquired_at TIMESTAMP,
    released_at TIMESTAMP,
    status TEXT,                   -- "active" | "released"
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);
```

### 表3: conflicts（Git冲突记录）
```sql
CREATE TABLE conflicts (
    id TEXT PRIMARY KEY,
    task_id TEXT,
    conflict_type TEXT,            -- "git_merge" | "file_locked" | "boundary_overlap"
    file_path TEXT,
    description TEXT,
    git_conflict_mark TEXT,       -- Git冲突标记: "<<<<<<<"
    severity TEXT,
    status TEXT DEFAULT "pending",
    created_at TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by TEXT,               -- 高权限Agent ID
    resolution TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);
```

### 表4: privileged_tasks（高权限任务）
```sql
CREATE TABLE privileged_tasks (
    id TEXT PRIMARY KEY,
    task_type TEXT,                -- "conflict_resolution" | "manual_merge"
    conflict_id TEXT,
    priority INTEGER,
    description TEXT,
    status TEXT DEFAULT "pending",
    assigned_to TEXT,
    created_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

---

## Worker安全检查流程

```go
// Worker在编辑前执行安全检查
func (w *Worker) SafeEditCheck(filePath string, lineStart, lineEnd int) error {
    // 1. 检查SQLite中的文件边界
    boundaries := db.GetFileBoundaries(w.TaskID, w.ID)
    if !isInBoundary(filePath, lineStart, lineEnd, boundaries) {
        return fmt.Errorf("Edit outside assigned boundary")
    }

    // 2. 检查Git文件锁
    locks := git.GetFileLocks(filePath)
    for _, lock := range locks {
        if lock.Owner != w.ID {
            return fmt.Errorf("File locked by %s", lock.Owner)
        }
    }

    // 3. 检查是否有未提交的更改
    if git.HasUncommittedChanges(filePath) {
        return fmt.Errorf("File has uncommitted changes")
    }

    return nil
}
```

---

## 高权限Agent工作流

```bash
# 高权限Agent处理冲突
1. git checkout master
2. git status (查看冲突文件)
3. git diff --check (检查冲突标记)
4. 手动解决冲突或调用Claude API
5. git add .
6. git commit -m "Resolved conflict: ..."
7. 更新SQLite: conflict.status = "resolved"
8. 通知所有Worker可以继续
```

---

## 优势总结

| 特性 | Git实现 | SQLite实现 |
|------|---------|-----------|
| 文件锁 | ✅ Git LFS locks | ❌ |
| 版本控制 | ✅ Git commits | ❌ |
| 冲突检测 | ✅ Git merge | ❌ |
| 历史追踪 | ✅ Git log/blame | ❌ |
| 任务管理 | ❌ | ✅ SQL queries |
| 状态查询 | ❌ | ✅ SQL queries |
| 实时监控 | ❌ | ✅ SQL triggers |

### 最佳实践
1. **所有文件操作通过Git**: 不直接编辑文件
2. **SQLite只存元数据**: 不存储文件内容
3. **Git hooks集成**: 自动触发检查和通知
4. **分支隔离**: 每个任务独立分支
5. **原子操作**: Git提交保证一致性
