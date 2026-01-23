# AgentFlow 功能迁移任务列表

**创建日期**: 2026-01-22
**源版本**: old/ (完整版)
**目标版本**: golang/ (简化版)
**分支**: feature/1.0.0

## 🎯 迁移目标

将 old 版本的高级功能迁移到新的 golang 版本，使其功能完整。

## 📋 迁移任务优先级

### 🔴 高优先级（核心功能）

#### 任务 1: HTTP 执行器
**状态**: ⏳ 待开始
**预计工时**: 2-3 小时
**复杂度**: 中等

**描述**:
从 `old/internal/worker/http_executor.go` 迁移 HTTP 执行器功能。

**功能点**:
- [ ] 创建 `internal/worker/http_executor.go`
- [ ] 实现独立的 Claude HTTP 服务器
- [ ] 支持 `/execute` API 端点
- [ ] Token 统计功能
- [ ] 超时控制（120 秒）
- [ ] 集成到 Worker 执行流程

**文件**:
- 新建: `golang/internal/worker/http_executor.go`
- 新建: `golang/internal/executor/server.go` (Claude 服务器)
- 修改: `golang/internal/worker/worker.go` (集成 HTTP 执行器)

**API 端点**:
```go
POST /execute
{
  "task_id": "TASK-001",
  "title": "Test",
  "description": "prompt",
  "worker_id": "WORKER-001"
}
```

**测试**:
```bash
# 启动 Claude 服务器
go run internal/executor/server.go

# 测试执行
curl -X POST http://localhost:8849/execute \
  -H "Content-Type: application/json" \
  -d '{"task_id": "TEST", "description": "echo hello"}'
```

---

#### 任务 2: OneShot 模式
**状态**: ⏳ 待开始
**预计工时**: 1-2 小时
**复杂度**: 低

**描述**:
实现 OneShot 运行模式，执行一个任务后自动退出。

**功能点**:
- [ ] 添加 `--oneshot` 参数支持
- [ ] 实现单次任务执行逻辑
- [ ] 完成后自动退出
- [ ] 返回码处理（成功/失败）

**文件**:
- 修改: `golang/cmd/worker/main.go` (添加 oneshot flag)
- 修改: `golang/internal/worker/worker.go` (实现 oneshot 逻辑)

**测试**:
```bash
# Terminal 1
./bin/master --port 8848

# Terminal 2: 创建任务
curl -X POST http://127.0.0.1:8848/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Oneshot Test", "description": "shell:echo test"}'

# Terminal 3: oneshot 执行
./bin/worker --master http://127.0.0.1:8848 --oneshot
# 应该执行一个任务后自动退出
```

---

#### 任务 3: 配置系统
**状态**: ⏳ 待开始
**预计工时**: 2-3 小时
**复杂度**: 中等

**描述**:
从 `old/internal/config/config.go` 迁移配置管理系统。

**功能点**:
- [ ] 创建 `internal/config/config.go`
- [ ] 支持环境变量
- [ ] 支持配置文件（YAML/TOML）
- [ ] 配置验证
- [ ] 默认值处理

**配置结构**:
```go
type Config struct {
    // Master
    MasterHost string
    MasterPort int
    MasterDBPath string

    // Worker
    WorkerID string
    WorkerName string
    WorkerGroup string

    // Claude
    ClaudeServerURL string
    ClaudeTimeout time.Duration

    // Git
    GitEnabled bool
    GitRepoPath string
}
```

**文件**:
- 新建: `golang/internal/config/config.go`
- 修改: `golang/cmd/master/main.go` (使用配置)
- 修改: `golang/cmd/worker/main.go` (使用配置)

**测试**:
```bash
# 环境变量
export AGENTFLOW_MASTER_PORT=8850
./bin/master

# 配置文件
./bin/master --config config.yaml
```

---

### 🟡 中优先级（协作功能）

#### 任务 4: Git 集成基础
**状态**: ⏳ 待开始
**预计工时**: 4-6 小时
**复杂度**: 高

**描述**:
从 `old/internal/git/` 和 `old/internal/worker/git_worker.go` 迁移 Git 集成功能。

**功能点**:
- [ ] 创建 `internal/git/client.go`
- [ ] Git 分支创建
- [ ] Git 提交
- [ ] Git 状态检查
- [ ] 文件变更检测

**文件**:
- 新建: `golang/internal/git/client.go`
- 新建: `golang/internal/git/operations.go`

**API**:
```bash
# 创建 Git 分支
git checkout -b worker-1/task-001

# 提交变更
git add .
git commit -m "Task 001: Implement feature"

# 检查状态
git status
```

---

#### 任务 5: 文件边界系统
**状态**: ⏳ 待开始
**预计工时**: 4-5 小时
**复杂度**: 高

**描述**:
从 `old/internal/database/file_boundaries.go` 迁移文件边界管理。

**数据库表**:
```sql
CREATE TABLE file_boundaries (
    id INTEGER PRIMARY KEY,
    task_id TEXT,
    worker_id TEXT,
    file_pattern TEXT,
    line_range TEXT,
    access_type TEXT  -- "exclusive" | "shared" | "readonly"
);
```

**功能点**:
- [ ] 创建数据库表
- [ ] 创建 `internal/database/file_boundaries.go`
- [ ] 文件边界分配
- [ ] 边界冲突检测
- [ ] API 端点实现

**API 端点**:
- `POST /api/v1/file-boundaries` - 创建边界
- `GET /api/v1/file-boundaries/:task_id` - 查询边界
- `POST /api/v1/file-boundaries/check` - 检查冲突

---

#### 任务 6: 文件锁系统
**状态**: ⏳ 待开始
**预计工时**: 3-4 小时
**复杂度**: 高

**描述**:
从 `old/internal/database/file_locks.go` 迁移文件锁机制。

**数据库表**:
```sql
CREATE TABLE git_locks (
    id INTEGER PRIMARY KEY,
    task_id TEXT,
    worker_id TEXT,
    file_path TEXT,
    lock_type TEXT,  -- "read" | "write"
    status TEXT,     -- "active" | "released"
    acquired_at TIMESTAMP,
    released_at TIMESTAMP
);
```

**功能点**:
- [ ] 创建数据库表
- [ ] 创建 `internal/database/file_locks.go`
- [ ] 文件锁获取
- [ ] 文件锁释放
- [ ] 锁冲突检测
- [ ] 自动解锁机制

**API 端点**:
- `POST /api/v1/locks/acquire` - 获取锁
- `POST /api/v1/locks/release` - 释放锁
- `GET /api/v1/locks/:file_path` - 查询锁状态

---

### 🟢 低优先级（高级功能）

#### 任务 7: gRPC 服务
**状态**: ⏳ 待开始
**预计工时**: 3-4 小时
**复杂度**: 高

**描述**:
从 `old/internal/grpc/` 迁移 gRPC/Claude Server。

**注意**: gRPC 可能不是必需的，如果 HTTP 执行器已足够。

**决策点**: 是否需要 gRPC？

---

#### 任务 8: Topic 管理
**状态**: ⏳ 待开始
**预计工时**: 2-3 小时
**复杂度**: 中等

**描述**:
从 `old/internal/master/topics.go` 迁移 Topic 管理功能。

**功能点**:
- [ ] Topic 注册
- [ ] Topic 订阅
- [ ] Topic 广播
- [ ] Worker-Topic 映射

**API 端点**:
- `POST /api/v1/topics/register` - 注册 Topic
- `GET /api/v1/topics` - 列出 Topics
- `GET /api/v1/topics/available` - 查询可用 Topics
- `GET /api/v1/topics/worker/:worker_id` - Worker 的 Topics

---

#### 任务 9: 冲突检测与解决
**状态**: ⏳ 待开始
**预计工时**: 4-5 小时
**复杂度**: 高

**描述**:
从 `old/internal/database/` 的冲突相关模块迁移。

**数据库表**:
```sql
CREATE TABLE edit_conflicts (
    id INTEGER PRIMARY KEY,
    conflict_id TEXT,
    task_id TEXT,
    worker_id TEXT,
    file_path TEXT,
    line_number INTEGER,
    conflict_type TEXT,
    severity TEXT,
    status TEXT
);

CREATE TABLE conflict_resolutions (
    id INTEGER PRIMARY KEY,
    conflict_id TEXT,
    resolver_agent_id TEXT,
    resolution_action TEXT,
    explanation TEXT,
    resolved_at TIMESTAMP
);
```

**功能点**:
- [ ] 冲突检测
- [ ] 冲突上报
- [ ] 冲突解决记录
- [ ] 冲突统计

**API 端点**:
- `POST /api/v1/conflicts` - 上报冲突
- `GET /api/v1/conflicts` - 查询冲突
- `POST /api/v1/conflicts/:id/resolve` - 解决冲突

---

## 📊 迁移进度追踪

| 任务 | 状态 | 进度 | 预计完成时间 |
|------|------|------|--------------|
| 1. HTTP 执行器 | ⏳ 待开始 | 0% | 2-3 小时 |
| 2. OneShot 模式 | ⏳ 待开始 | 0% | 1-2 小时 |
| 3. 配置系统 | ⏳ 待开始 | 0% | 2-3 小时 |
| 4. Git 集成基础 | ⏳ 待开始 | 0% | 4-6 小时 |
| 5. 文件边界系统 | ⏳ 待开始 | 0% | 4-5 小时 |
| 6. 文件锁系统 | ⏳ 待开始 | 0% | 3-4 小时 |
| 7. gRPC 服务 | ⏳ 待开始 | 0% | 3-4 小时 |
| 8. Topic 管理 | ⏳ 待开始 | 0% | 2-3 小时 |
| 9. 冲突检测 | ⏳ 待开始 | 0% | 4-5 小时 |

**总预计工时**: 约 25-35 小时

---

## 🛠️ 使用方法

### 开始迁移

1. **选择任务**: 从高优先级开始
2. **查看 old 代码**: 参考 `old/internal/` 中的实现
3. **创建新文件**: 在 `golang/internal/` 中创建对应文件
4. **实现功能**: 按功能点逐步实现
5. **测试验证**: 使用测试用例验证
6. **提交代码**: 完成后提交到 git

### 迁移流程

```bash
# 1. 开始任务
/agentflow migrate start <任务编号>

# 2. 查看任务详情
/agentflow migrate status <任务编号>

# 3. 标记完成
/agentflow migrate complete <任务编号>
```

### 创建迁移任务

```bash
# 为任务 1 创建子任务
curl -X POST http://127.0.0.1:8848/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Migrate HTTP Executor",
    "description": "migration:task-1",
    "group_name": "migration"
  }'
```

---

## 📝 迁移最佳实践

### 代码复用策略
1. **复制粘贴**: 直接复制 old/ 代码到 golang/
2. **调整适配**: 修改包名和导入路径
3. **简化优化**: 去除不必要的复杂度
4. **测试验证**: 确保功能正常

### 测试策略
1. **单元测试**: 为每个功能编写测试
2. **集成测试**: 测试功能间的协作
3. **端到端测试**: 完整流程测试
4. **性能测试**: 确保性能可接受

### 文档更新
- [ ] 更新 API 文档
- [ ] 更新架构设计文档
- [ ] 更新迁移记录
- [ ] 更新 README

---

## 🎯 里程碑

### Milestone 1: 核心功能恢复
**目标**: 完成任务 1-3
- ✅ HTTP 执行器
- ✅ OneShot 模式
- ✅ 配置系统

**预计完成**: 1-2 天

### Milestone 2: 协作功能恢复
**目标**: 完成任务 4-6
- ✅ Git 集成
- ✅ 文件边界
- ✅ 文件锁

**预计完成**: 3-5 天

### Milestone 3: 完整功能恢复
**目标**: 完成任务 7-9
- ✅ gRPC 服务
- ✅ Topic 管理
- ✅ 冲突检测

**预计完成**: 1-2 周

---

## 📚 参考资料

- [golang/docs/FEATURE_COMPARISON.md](golang/docs/FEATURE_COMPARISON.md) - 功能对比
- [golang/docs/BUILD_GUIDE.md](golang/docs/BUILD_GUIDE.md) - 编译指南
- [old/](old/) - 源代码目录

---

**创建者**: Claude Sonnet 4.5
**最后更新**: 2026-01-22
**状态**: 准备开始迁移
