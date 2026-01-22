# AgentFlow 功能对比：old vs golang

**日期**: 2026-01-22
**对比版本**: old/ (完整版) vs golang/ (简化版)

## 📊 总体对比

| 类别 | old 版本 | golang 版本 | 状态 |
|------|----------|-------------|------|
| 核心功能 | 完整 | 基础 | ⚠️ 简化 |
| 代码行数 | ~5000+ 行 | ~500 行 | ⚠️ 减少 90% |
| 内部模块 | 7 个 | 4 个 | ⚠️ 减少 3 个 |
| 数据库功能 | 高级 | 基础 | ⚠️ 简化 |
| Worker 功能 | 多模式 | 单模式 | ⚠️ 简化 |
| Git 集成 | ✅ 完整 | ❌ 缺失 | 🔴 缺失 |
| gRPC 支持 | ✅ 有 | ❌ 无 | 🔴 缺失 |
| 文件锁 | ✅ 有 | ❌ 无 | 🔴 缺失 |
| 边界控制 | ✅ 有 | ❌ 无 | 🔴 缺失 |

---

## 🔴 缺失功能详细列表

### 1. Git 集成功能

**old 版本**:
- `internal/git/client.go` - Git 客户端
- `internal/worker/git_worker.go` - Git Worker
- `internal/database/git_integration.go` - Git 集成数据模型

**功能描述**:
- Git 分支管理
- 文件边界定义（FileBoundaries）
- Git 文件锁（GitLocks）
- Git 任务分配
- 合并冲突检测
- 安全编辑检查

**golang 版本**: ❌ **完全缺失**

**影响**:
- 无法进行多 Worker 协作编辑
- 无法防止文件冲突
- 无法实现 Git 版本控制集成

---

### 2. gRPC / Claude Server

**old 版本**:
- `internal/grpc/claude_server.go` - Claude HTTP 服务器
- `internal/grpc/simple_server.go` - gRPC 服务器
- `internal/grpc/claude.proto` - gRPC 协议定义

**功能描述**:
- 独立的 Claude 执行服务器（端口 8849）
- HTTP API 接口（/execute）
- Token 统计
- 超时控制
- 优先级队列

**golang 版本**: ❌ **完全缺失**

**影响**:
- 无独立的 Claude 服务器
- 无 HTTP 执行接口
- Worker 只能直接调用 Claude CLI

---

### 3. 文件边界系统

**old 版本**:
- `internal/database/file_boundaries.go` - 文件边界管理
- `database/schema_file_boundaries.sql` - 数据库表

**功能描述**:
```go
type FileBoundary struct {
    ID          int64
    TaskID      string
    WorkerID    string
    FilePattern string    // e.g., "src/**/*.go"
    LineRange   string    // e.g., "1-100"
    AccessType  string    // "exclusive" | "shared" | "readonly"
}
```

**用途**:
- 控制哪些 Worker 可以编辑哪些文件
- 防止并发编辑冲突
- 定义文件访问权限

**golang 版本**: ❌ **完全缺失**

---

### 4. 文件锁系统

**old 版本**:
- `internal/database/file_locks.go` - 文件锁管理
- `database/schema_file_locks.sql` - 数据库表

**功能描述**:
```go
type GitLock struct {
    ID         int64
    TaskID     string
    WorkerID   string
    FilePath   string
    LockType   string    // "read" | "write"
    AcquiredAt time.Time
    Status     string    // "active" | "released"
}
```

**用途**:
- 防止多个 Worker 同时编辑同一文件
- 读写锁机制
- 锁冲突检测

**golang 版本**: ❌ **完全缺失**

---

### 5. HTTP 执行器

**old 版本**:
- `internal/worker/http_executor.go` - HTTP 执行器

**功能描述**:
```go
type HTTPExecutor struct {
    serverURL string  // Claude 服务器地址
    logger    *logrus.Logger
    client    *http.Client
}
```

**用途**:
- 通过 HTTP 调用 Claude 服务器
- 优先级 1 的执行方式
- 120 秒超时

**golang 版本**: ❌ **完全缺失**

---

### 6. OneShot Runner

**old 版本**:
- `internal/worker/oneshot_runner.go` - 单次运行器

**功能描述**:
```go
type OneShotRunner struct {
    client         *Client
    claudeExecutor *ClaudeExecutor
    httpExecutor   *HTTPExecutor  // ← 3 种执行器
    logger         *logrus.Logger
}
```

**执行优先级**:
1. HTTP Executor（最快）
2. Claude CLI Executor
3. 本地执行

**golang 版本**: ⚠️ **简化版本**
- 只有基本的 Worker
- 无 OneShot 专用模式
- 无多执行器优先级

---

### 7. 高级数据库功能

**old 版本额外功能**:

#### a) Git 任务表
```sql
CREATE TABLE git_tasks (
    id TEXT PRIMARY KEY,
    title TEXT,
    git_branch TEXT,
    file_boundaries TEXT,  -- JSON
    locks TEXT,            -- JSON
    status TEXT
);
```

#### b) 文件边界表
```sql
CREATE TABLE file_boundaries (
    id INTEGER PRIMARY KEY,
    task_id TEXT,
    worker_id TEXT,
    file_pattern TEXT,
    line_range TEXT,
    access_type TEXT
);
```

#### c) 文件锁表
```sql
CREATE TABLE git_locks (
    id INTEGER PRIMARY KEY,
    task_id TEXT,
    worker_id TEXT,
    file_path TEXT,
    lock_type TEXT,
    status TEXT
);
```

#### d) 编辑冲突表
```sql
CREATE TABLE edit_conflicts (
    id INTEGER PRIMARY KEY,
    conflict_id TEXT,
    task_id TEXT,
    file_path TEXT,
    line_number INTEGER,
    conflict_type TEXT,
    severity TEXT,
    status TEXT
);
```

#### e) 冲突解决记录表
```sql
CREATE TABLE conflict_resolutions (
    id INTEGER PRIMARY KEY,
    conflict_id TEXT,
    resolver_agent_id TEXT,
    resolution_action TEXT,
    explanation TEXT
);
```

**golang 版本**: ❌ **只有基础表**（tasks, workers）

---

### 8. 高级 Master API

**old 版本额外 API**:

| API 端点 | 功能 | golang 版本 |
|---------|------|-------------|
| `POST /api/tasks/create-with-git` | 创建 Git 任务 | ❌ 缺失 |
| `POST /api/tasks/assign-git` | 分配 Git 任务 | ❌ 缺失 |
| `GET /api/tasks/:task_id/git` | 获取 Git 任务详情 | ❌ 缺失 |
| `POST /api/conflicts` | 报告冲突 | ❌ 缺失 |
| `GET /api/conflicts` | 获取待解决冲突 | ❌ 缺失 |
| `POST /api/conflicts/resolve` | 解决冲突 | ❌ 缺失 |
| `GET /api/topics` | Topic 管理 | ❌ 缺失 |
| `POST /api/topics/register` | 注册 Topic | ❌ 缺失 |
| `GET /api/topics/available` | 检查 Topic 可用性 | ❌ 缺失 |
| `GET /api/topics/worker/:worker_id` | 获取 Worker 的 Topics | ❌ 缺失 |

**golang 版本**: ⚠️ **只有基础 API**（创建、分配、完成任务）

---

### 9. 配置系统

**old 版本**:
- `internal/config/config.go` - 完整配置管理

**功能**:
- Master 配置
- Worker 配置
- Claude 配置
- Git 配置
- 数据库配置

**golang 版本**: ⚠️ **只有环境变量**
- 无专门的配置系统
- 无配置文件

---

### 10. API 类型定义

**old 版本**:
- `internal/api/types.go` - API 类型定义

**功能**:
- 统一的请求/响应类型
- 错误处理
- 数据验证

**golang 版本**: ❌ **缺失**

---

## ⚠️ 简化功能

### 1. Worker 功能

**old 版本**:
- GitWorker - Git 集成 Worker
- 普通 Worker - 基础 Worker
- OneShot 模式 - 执行一个任务后退出
- Auto 模式 - 自动拉取任务
- 手动模式 - 等待命令

**golang 版本**:
- 只有基础 Worker
- 简化的 auto 模式

---

### 2. Master 功能

**old 版本**:
- Standalone 模式（带自动关闭）
- Cloud 模式（持续运行）
- Topic 管理
- 冲突管理
- Git 任务管理

**golang 版本**:
- 简化的 HTTP 服务器
- 基础 CRUD API
- 无 Topic/冲突管理

---

## 📈 代码统计

### old 版本

| 模块 | 文件数 | 估计行数 |
|------|--------|----------|
| database | 7 | ~1500 |
| master | 4 | ~800 |
| worker | 7 | ~1200 |
| grpc | 3 | ~400 |
| git | 1 | ~200 |
| config | 1 | ~150 |
| api | 1 | ~100 |
| **总计** | **23** | **~4350** |

### golang 版本

| 模块 | 文件数 | 实际行数 |
|------|--------|----------|
| database | 3 | ~300 |
| master | 1 | ~150 |
| worker | 2 | ~250 |
| model | 1 | ~50 |
| **总计** | **7** | **~750** |

**减少**: ~85% 代码量

---

## 🎯 功能保留情况

### ✅ 完全保留

- ✅ 基础任务管理（创建、分配、完成）
- ✅ Worker 注册和心跳
- ✅ SQLite 数据库
- ✅ RESTful API（基础）
- ✅ Claude CLI 集成
- ✅ Shell 命令执行

### ⚠️ 部分保留

- ⚠️ Master-Worker 架构 - 存在但简化
- ⚠️ 任务优先级 - 有但简化
- ⚠️ 进度跟踪 - 有但简化

### ❌ 完全缺失

- ❌ Git 集成（分支、锁、边界）
- ❌ gRPC/Claude Server
- ❌ HTTP 执行器
- ❌ 文件锁系统
- ❌ 冲突检测和解决
- ❌ Topic 管理
- ❌ OneShot 模式
- ❌ 高级数据库表（6 个表）
- ❌ 高级 API 端点（10 个）

---

## 💡 建议

### 如果需要完整功能

1. **使用 old 版本**
   - 位置: `old/` 目录
   - 功能完整但代码较旧

2. **迁移缺失功能**
   - 从 old/ 复制需要的模块
   - 根据新架构调整
   - 逐步集成

3. **使用 Python 版本**
   - 位置: `python/` 目录
   - 功能简化但易用

### 如果只需要基础功能

**golang 版本已足够**:
- ✅ 任务分配和执行
- ✅ 多进程并发
- ✅ Claude CLI 集成
- ✅ 简单部署

---

## 📋 功能优先级建议

如果要恢复功能，建议按以下优先级：

### 高优先级（核心功能）

1. **HTTP 执行器** - 允许远程 Claude 执行
2. **OneShot 模式** - 简化部署场景
3. **配置系统** - 更灵活的配置

### 中优先级（协作功能）

4. **Git 集成** - 多 Worker 协作
5. **文件边界** - 防止冲突
6. **文件锁** - 安全编辑

### 低优先级（高级功能）

7. **gRPC Server** - 替代方案已存在
8. **Topic 管理** - 可选功能
9. **冲突解决** - 复杂度高

---

**结论**: golang 版本是一个**简化版**，适合基础场景。如果需要完整功能，建议参考 old 版本或逐步添加缺失功能。

---

**文档日期**: 2026-01-22
**对比版本**: old/ vs golang/
**作者**: Claude Sonnet 4.5
