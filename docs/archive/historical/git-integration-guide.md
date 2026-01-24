# AgentFlow Git 集成和文件权限控制指南

## 🎯 核心概念

AgentFlow 的 Git 集成系统提供以下核心功能：

### 1. **文件边界（File Boundaries）**
每个 Agent 只能操作被授权的文件，防止越权修改。

### 2. **文件锁（File Locking）**
防止多个 Agent 同时修改同一文件，避免冲突。

### 3. **Git 分支隔离**
每个 Agent 任务在独立分支工作，互不干扰。

### 4. **冲突检测**
自动检测文件锁定冲突和合并冲突。

---

## 📋 文件边界配置

### 配置文件位置
```
.agentflow/boundaries.json
```

### 配置示例

```json
{
  "frontend": [
    {
      "file_pattern": "src/frontend/**/*",
      "access_type": "exclusive",
      "description": "Frontend agent can exclusively modify frontend files"
    },
    {
      "file_pattern": "src/api/**/*",
      "access_type": "readonly",
      "description": "Frontend agent can read API files"
    }
  ],
  "backend": [
    {
      "file_pattern": "src/backend/**/*",
      "access_type": "exclusive",
      "description": "Backend agent can exclusively modify backend files"
    },
    {
      "file_pattern": "src/api/**/*",
      "access_type": "shared",
      "description": "Backend agent shares API files"
    }
  ],
  "database": [
    {
      "file_pattern": "src/database/**/*",
      "access_type": "exclusive",
      "description": "Database agent exclusively manages database files"
    }
  ]
}
```

### 访问类型说明

| access_type | 读权限 | 写权限 | 说明 |
|-------------|--------|--------|------|
| `exclusive` | ✅ | ✅ | 独占访问，只有该 Agent 可以操作 |
| `shared` | ✅ | ✅ | 共享访问，多个 Agent 可以操作（需文件锁） |
| `readonly` | ✅ | ❌ | 只读访问，不能修改 |

---

## 🚀 Python 版本使用

### 1. 初始化 Git 集成管理器

```python
from agentflow.git_integration import GitIntegrationManager

# 初始化
manager = GitIntegrationManager(
    repo_path=".",  # Git 仓库路径
    boundary_config=".agentflow/boundaries.json"
)
```

### 2. 创建 Agent 任务

```python
# 创建新的 Agent 任务
task = manager.create_agent_task(
    agent_id="frontend",
    task_id="TASK-001",
    description="实现用户登录页面"
)

print(f"任务已创建: {task['branch']}")
# 输出: 任务已创建: agent-frontend/task-TASK-001
```

### 3. 验证文件访问权限

```python
# 检查 Agent 是否可以访问文件
allowed, reason = manager.verify_file_access(
    agent_id="frontend",
    file_path="src/frontend/login.tsx",
    access_type="write"
)

if allowed:
    print("✅ 允许访问")
else:
    print(f"❌ 拒绝访问: {reason}")
```

### 4. 安全的文件操作

```python
# 定义文件操作函数
def modify_file(file_path):
    with open(file_path, 'w') as f:
        f.write("new content")

# 使用安全操作（自动加锁/解锁）
try:
    manager.safe_file_operation(
        agent_id="frontend",
        file_path="src/frontend/login.tsx",
        operation=modify_file
    )
    print("✅ 文件修改成功并已提交")
except PermissionError as e:
    print(f"❌ 操作失败: {e}")
```

### 5. 完成任务并合并

```python
# 完成任务并合并到主分支
result = manager.complete_agent_task(
    agent_id="frontend",
    task_id="TASK-001",
    merge_strategy="squash"  # merge | squash | rebase
)

if result['status'] == 'completed':
    print(f"✅ {result['message']}")
elif result['status'] == 'conflict':
    print(f"⚠️ 合并冲突:")
    for conflict in result['conflicts']:
        print(f"  - {conflict}")
```

### 6. 完整工作流示例

```python
from agentflow.git_integration import GitIntegrationManager

manager = GitIntegrationManager(".", ".agentflow/boundaries.json")

# 1. 创建任务
task = manager.create_agent_task("backend", "TASK-001", "添加用户认证")

# 2. 在独立分支上进行开发
# (Agent 在此处执行开发任务)

# 3. 验证文件权限
allowed, _ = manager.verify_file_access("backend", "src/backend/auth.go", "write")
if not allowed:
    raise PermissionError("无权限访问该文件")

# 4. 安全地修改文件
def add_auth_code(file_path):
    with open(file_path, 'a') as f:
        f.write("\ndef authenticate_user():\n    pass\n")

manager.safe_file_operation("backend", "src/backend/auth.go", add_auth_code)

# 5. 完成任务
result = manager.complete_agent_task("backend", "TASK-001", "squash")
print(f"任务结果: {result['status']}")
```

---

## 🔧 Go 版本使用

### 1. 初始化 Git 集成管理器

```go
package main

import (
    "context"
    "github.com/jiangxiaolong/agentflow-go/internal/git"
    "github.com/jiangxiaolong/agentflow-go/internal/database"
)

func main() {
    // 初始化数据库
    db, _ := database.NewDatabase(".agentflow/agentflow.db")
    db.Init()

    // 创建 Git 表
    git.CreateGitTables(db.DB)

    // 初始化 Git 集成管理器
    manager := git.NewGitIntegrationManager(
        ".",  // Git 仓库路径
        ".agentflow/boundaries.json",
        db,
        logger,
    )
}
```

### 2. 创建 Agent 任务

```go
// 创建新的 Agent 任务
task, err := manager.CreateAgentTask(
    context.Background(),
    "frontend",  // Agent ID
    "TASK-001",  // Task ID
    "实现用户登录页面",  // Description
)

if err != nil {
    log.Fatalf("创建任务失败: %v", err)
}

fmt.Printf("任务已创建: %s\n", task.GitBranch)
```

### 3. 验证文件访问权限

```go
// 检查 Agent 是否可以访问文件
allowed, reason := manager.VerifyFileAccess(
    "frontend",
    "src/frontend/login.tsx",
    "write",
)

if allowed {
    fmt.Println("✅ 允许访问")
} else {
    fmt.Printf("❌ 拒绝访问: %s\n", reason)
}
```

### 4. 安全的文件操作

```go
// 定义文件操作函数
operation := func() error {
    return os.WriteFile(
        "src/frontend/login.tsx",
        []byte("new content"),
        0644,
    )
}

// 使用安全操作（自动加锁/解锁）
err := manager.SafeFileOperation(
    context.Background(),
    "frontend",
    "src/frontend/login.tsx",
    operation,
)

if err != nil {
    log.Printf("❌ 操作失败: %v", err)
}
```

### 5. 完成任务并合并

```go
// 完成任务并合并到主分支
result, err := manager.CompleteAgentTask(
    context.Background(),
    "frontend",
    "TASK-001",
    "squash",  // merge | squash | rebase
)

if err != nil {
    log.Fatalf("任务完成失败: %v", err)
}

if result["status"] == "completed" {
    fmt.Println("✅ 任务完成并已合并")
} else if result["status"] == "conflict" {
    fmt.Println("⚠️ 合并冲突:")
    conflicts := result["conflicts"].([]string)
    for _, conflict := range conflicts {
        fmt.Printf("  - %s\n", conflict)
    }
}
```

---

## 🔒 文件锁机制

### 锁类型

| 锁类型 | 说明 | 并发支持 |
|-------|------|---------|
| `read` | 读锁，允许多个 Agent 同时读取 | ✅ 多个读锁可共存 |
| `write` | 写锁，只允许一个 Agent 修改 | ❌ 写锁互斥 |

### 手动加锁/解锁

**Python**:
```python
from agentflow.git_integration import GitFileBoundaryManager

manager = GitFileBoundaryManager(".agentflow/boundaries.json")

# 加锁
if manager.acquire_lock("frontend", "src/app.tsx", "write"):
    print("✅ 写锁已获取")

    # 执行操作
    # ...

    # 解锁
    manager.release_lock("frontend", "src/app.tsx", "write")
    print("✅ 写锁已释放")
else:
    print("❌ 文件已被锁定")
```

**Go**:
```go
manager := git.NewBoundaryManager(".agentflow/boundaries.json", logger)

// 加锁
if manager.AcquireLock("frontend", "src/app.tsx", "write") {
    fmt.Println("✅ 写锁已获取")

    // 执行操作
    // ...

    // 解锁
    manager.ReleaseLock("frontend", "src/app.tsx", "write")
    fmt.Println("✅ 写锁已释放")
} else {
    fmt.Println("❌ 文件已被锁定")
}
```

---

## ⚠️ 冲突检测

### 自动冲突检测

系统会自动检测以下冲突：

1. **文件锁定冲突** - 尝试修改已被其他 Agent 锁定的文件
2. **边界重叠冲突** - 多个 Agent 声明对同一文件的独占访问
3. **Git 合并冲突** - 分支合并时的冲突

**示例**:
```python
# 检查冲突
conflicts = manager.boundary_manager.check_conflicts(
    agent_id="frontend",
    file_paths=["src/app.tsx", "src/utils.ts"]
)

if conflicts:
    print(f"⚠️ 检测到 {len(conflicts)} 个冲突:")
    for conflict in conflicts:
        print(f"  - {conflict.description}")
        print(f"    严重程度: {conflict.severity}")
else:
    print("✅ 无冲突")
```

---

## 📂 典型项目结构示例

### 多 Agent 协作开发

```
my-project/
├── .agentflow/
│   └── boundaries.json          # 文件边界配置
├── src/
│   ├── frontend/               # Frontend Agent 独占
│   │   ├── components/
│   │   ├── pages/
│   │   └── App.tsx
│   ├── backend/                # Backend Agent 独占
│   │   ├── controllers/
│   │   ├── services/
│   │   └── main.go
│   ├── database/               # Database Agent 独占
│   │   ├── migrations/
│   │   ├── models/
│   │   └── schema.sql
│   └── api/                    # 共享区域
│       ├── schemas/
│       └── types.ts
└── .git/
```

**对应配置**:
```json
{
  "agent-frontend": [
    {
      "file_pattern": "src/frontend/**/*",
      "access_type": "exclusive",
      "description": "Frontend UI 代码"
    },
    {
      "file_pattern": "src/api/**/*",
      "access_type": "readonly",
      "description": "可以读取 API 定义"
    }
  ],
  "agent-backend": [
    {
      "file_pattern": "src/backend/**/*",
      "access_type": "exclusive",
      "description": "Backend 业务逻辑"
    },
    {
      "file_pattern": "src/api/**/*",
      "access_type": "shared",
      "description": "API 定义可共享修改"
    }
  ],
  "agent-database": [
    {
      "file_pattern": "src/database/**/*",
      "access_type": "exclusive",
      "description": "数据库结构和迁移"
    }
  ]
}
```

---

## 🎯 实际使用场景

### 场景 1: 多 Agent 并行开发

**任务**: 同时开发用户管理功能

```python
# Backend Agent 开发 API
backend_manager = GitIntegrationManager(".", ".agentflow/boundaries.json")
backend_task = backend_manager.create_agent_task(
    "backend", "TASK-101", "实现用户管理 API"
)

# Frontend Agent 开发 UI
frontend_manager = GitIntegrationManager(".", ".agentflow/boundaries.json")
frontend_task = frontend_manager.create_agent_task(
    "frontend", "TASK-102", "实现用户管理界面"
)

# 两个 Agent 在各自分支上并行工作，互不干扰
# Backend: agent-backend/task-TASK-101
# Frontend: agent-frontend/task-TASK-102
```

### 场景 2: 防止越权修改

```python
# 尝试越权修改
allowed, reason = manager.verify_file_access(
    agent_id="frontend",
    file_path="src/backend/auth.go",  # Backend 专属文件
    access_type="write"
)

# 结果: (False, "Agent frontend not authorized to access src/backend/auth.go")
```

### 场景 3: 共享文件的并发控制

```python
# API 定义文件可以被多个 Agent 读取
# 但同一时间只能有一个 Agent 修改

# Agent Backend 尝试修改 API
if manager.boundary_manager.acquire_lock("backend", "src/api/user.ts", "write"):
    print("✅ Backend 获取到写锁")

    # 执行修改...
    manager.safe_file_operation("backend", "src/api/user.ts", modify_api)

    # 完成后释放锁
    manager.boundary_manager.release_lock("backend", "src/api/user.ts", "write")

# 在此期间，Frontend Agent 尝试修改会失败
# 返回: False (文件已被 backend 锁定)
```

---

## 🔧 高级配置

### 1. 行范围限制

限制 Agent 只能修改文件的特定行范围：

```json
{
  "agent-1": [
    {
      "file_pattern": "src/config.yaml",
      "access_type": "exclusive",
      "line_range": "1-50",  // 只能修改前 50 行
      "description": "配置文件上半部分"
    }
  ],
  "agent-2": [
    {
      "file_pattern": "src/config.yaml",
      "access_type": "exclusive",
      "line_range": "51-100",  // 只能修改后 50 行
      "description": "配置文件下半部分"
    }
  ]
}
```

### 2. 通配符模式

支持多种通配符模式：

```json
{
  "agent-ui": [
    {
      "file_pattern": "src/**/*.tsx",        // 匹配所有 .tsx 文件
      "access_type": "exclusive"
    },
    {
      "file_pattern": "src/components/**/*", // 匹配 components 下所有文件
      "access_type": "exclusive"
    },
    {
      "file_pattern": "src/api/*",           // 匹配 api 目录下的直接文件
      "access_type": "readonly"
    }
  ]
}
```

### 3. 动态边界管理

运行时添加/修改边界：

**Python**:
```python
from agentflow.git_integration import FileBoundary

# 添加新边界
new_boundary = FileBoundary(
    file_pattern="src/tests/**/*",
    access_type="shared",
    description="测试文件可共享"
)

manager.boundary_manager.boundaries["agent-test"].append(new_boundary)
manager.boundary_manager._save_boundaries()
```

**Go**:
```go
newBoundary := git.FileBoundary{
    FilePath:    "src/tests/**/*",
    AccessType:  "shared",
    Description: "测试文件可共享",
}

manager.boundaryManager.Boundaries["agent-test"] =
    append(manager.boundaryManager.Boundaries["agent-test"], newBoundary)
manager.boundaryManager.SaveBoundaries()
```

---

## 📊 监控和调试

### 查看当前锁状态

**Python**:
```python
active_locks = manager.boundary_manager.get_active_locks()
for lock in active_locks:
    print(f"{lock.file_path} - {lock.lock_type} by {lock.agent_id}")
```

**Go**:
```go
activeLocks := manager.GetActiveLocks("")
for _, lock := range activeLocks {
    fmt.Printf("%s - %s by %s\n", lock.FilePath, lock.LockType, lock.AgentID)
}
```

### 查看合并冲突

```python
conflicts = manager.branch_manager.get_merge_conflicts()
for file_path in conflicts:
    print(f"冲突文件: {file_path}")
```

---

## 🛡️ 最佳实践

### 1. 边界设计原则

- ✅ **明确职责** - 每个 Agent 有明确的功能边界
- ✅ **最小权限** - 只授予必要的文件访问权限
- ✅ **避免重叠** - 尽量减少 exclusive 区域的重叠
- ✅ **共享接口** - 使用 shared 区域定义清晰的 API

### 2. 锁的使用

- ✅ **快速加锁/解锁** - 获取锁后尽快释放
- ✅ **使用上下文管理器** - 确保锁一定会被释放
- ✅ **处理锁超时** - 避免死锁

### 3. 分支管理

- ✅ **任务完成后清理** - 及时删除已完成任务的分支
- ✅ **使用有意义的分支名** - 便于识别和调试
- ✅ **选择合适的合并策略** - 根据场景选择 merge/squash/rebase

---

## 📚 相关文档

- [Python Git Integration API](../python/agentflow/git_integration.py)
- [Go Git Integration API](../golang/internal/git/git_integration.go)
- [README.md](../README.md) - 项目整体说明

---

**版本**: v1.0.0
**更新**: 2026-01-22
