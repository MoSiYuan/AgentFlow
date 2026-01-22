# CPDS Git+SQLite混合架构测试指南

## 测试目标

验证Git+SQLite混合文件锁和任务编排架构的完整功能，使用克苏鲁故事生成作为测试场景。

## 前置准备

### 1. 启动Master服务器

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
go run cmd/master/main.go
```

Master服务器将在 `http://localhost:8848` 启动。

### 2. 启动Claude HTTP服务器（可选）

如果需要实际生成故事内容：

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
go run cmd/simple-server/main.go
```

Claude服务器将在 `http://localhost:8849` 启动。

---

## 测试方案

### 方案A: 快速测试（推荐）

快速测试核心API功能，无需实际生成Git仓库。

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
./scripts/quick_test.sh
```

**测试内容**：
- ✅ 创建Git任务
- ✅ 查询任务详情
- ✅ 分配任务给Worker
- ✅ 检测冲突状态

**预期输出**：
```
🚀 CPDS Git+SQLite混合架构 - 快速测试
======================================

📝 步骤1: 创建Git任务
{
  "success": true,
  "data": {
    "task_id": "STORY-001",
    "git_branch": "pending/STORY-001",
    "files_locked": ["story_001.md"],
    "message": "Git task created successfully..."
  }
}

📖 步骤2: 查询任务详情
{
  "success": true,
  "data": {
    "id": "STORY-001",
    "title": "深海遗迹中的古老召唤",
    "status": "pending",
    ...
  }
}

🎯 步骤3: 分配任务给Worker
Worker ID: test-worker-xxx
{
  "success": true,
  "data": {
    "task_id": "STORY-001",
    "git_branch": "test-worker-xxx/STORY-001",
    "message": "Git task assigned successfully"
  }
}

⚠️  步骤4: 查询冲突状态
待处理冲突数量: 0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ 快速测试完成
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### 方案B: 完整集成测试

完整的端到端测试，包括：
- 创建真实的Git仓库
- 创建和提交故事文件
- 测试分支合并和冲突检测

```bash
cd /Users/jiangxiaolong/work/project/game/AdStar/cpds-go
./scripts/test_git_integration.sh
```

**测试流程**：

#### 1. 创建Git测试仓库
```
📖 创建Git测试仓库...
✅ Git仓库初始化完成
   仓库路径: /tmp/cpds-git-test-xxxxxx/lovecraft-stories
```

#### 2. 创建5个Git任务
```
📝 测试1: 创建Git任务
======================================

创建任务 STORY-001: 深海遗迹中的古老召唤
创建任务 STORY-002: 被诅咒的家族族谱
创建任务 STORY-003: 南极冰层下的未知文明
创建任务 STORY-004: 会自动生长的诡异书籍
创建任务 STORY-005: 梦境中的呓语与现实重叠

✅ 创建了 5 个Git任务
```

#### 3. Worker分配并执行任务
```
🎯 测试3: Worker分配任务
======================================

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎬 处理任务: STORY-001
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

步骤1: 分配Git任务...
{
  "success": true,
  "data": {
    "git_branch": "worker-xxx/STORY-001"
  }
}

步骤2: 检出分支 worker-xxx/STORY-001...
✅ 已检出分支

步骤3: 生成克苏鲁故事...
故事描述: 创作一个关于'深海遗迹中的古老召唤'的克苏鲁神话故事...
故事文件: story_001.md

步骤4: 将故事写入文件...
✅ 故事已写入 story_001.md

步骤5: 提交变更到Git...
✅ 提交完成

步骤6: 尝试合并到主分支...
✅ 合并成功！

✅ 任务 STORY-001 处理完成
```

#### 4. 查看最终结果
```
📊 测试4: 查看最终结果
======================================

🌿 Git分支状态:
  main
  worker-xxx/STORY-001
  worker-xxx/STORY-002
  worker-xxx/STORY-003
  worker-xxx/STORY-004
  worker-xxx/STORY-005

📄 主分支文件列表:
-rw-r--r-- 1 user staff 1.2K Jan 21 18:00 story_001.md
-rw-r--r-- 1 user staff 1.1K Jan 21 18:00 story_002.md
-rw-r--r-- 1 user staff 1.3K Jan 21 18:00 story_003.md
-rw-r--r-- 1 user staff 1.0K Jan 21 18:00 story_004.md
-rw-r--r-- 1 user staff 1.2K Jan 21 18:00 story_005.md

📖 故事内容预览:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📖 story_001.md:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 克苏鲁故事：深海遗迹中的古老召唤

**创作时间**: 2025-01-21 18:00:00
**创作Worker**: worker-xxx
**任务ID**: STORY-001

---

在深渊的深处，某种古老的力量正在苏醒...
...
```

---

## 手动测试步骤

### 1. 创建Git任务

```bash
curl -X POST http://localhost:8848/api/tasks/create-with-git \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "STORY-001",
    "title": "深海遗迹中的古老召唤",
    "description": "创作一个关于深海遗迹中古老召唤的克苏鲁神话故事",
    "priority": "medium",
    "file_boundaries": [
      {
        "file_path": "story_001.md",
        "line_start": 1,
        "line_end": 100,
        "lock_type": "write"
      }
    ]
  }'
```

### 2. 查询任务详情

```bash
curl http://localhost:8848/api/tasks/STORY-001/git
```

### 3. 分配任务给Worker

```bash
curl -X POST http://localhost:8848/api/tasks/assign-git \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "STORY-001",
    "worker_id": "worker-test-001"
  }'
```

### 4. 查询冲突

```bash
curl http://localhost:8848/api/conflicts
```

### 5. 解决冲突

```bash
curl -X POST http://localhost:8848/api/conflicts/resolve \
  -H "Content-Type: application/json" \
  -d '{
    "conflict_id": "CONFLICT-001",
    "resolver_id": "privileged-agent-001",
    "resolution": "merged",
    "resolution_record": {
      "original_content": "...",
      "resolved_content": "...",
      "explanation": "Merged both versions"
    }
  }'
```

---

## 测试场景

### 场景1: 正常流程

**目标**: 验证完整的任务创建→分配→执行→合并流程

**步骤**:
1. 创建Git任务
2. 分配给Worker
3. Worker创建分支并锁定文件
4. Worker编辑并提交
5. 成功合并到主分支

**预期结果**: ✅ 全部成功

---

### 场景2: 文件锁冲突

**目标**: 验证Git文件锁机制

**步骤**:
1. Worker-1锁定 `story_001.md`
2. Worker-2尝试锁定同一文件
3. 检测到冲突

**预期结果**: ✅ Worker-2分配失败，冲突被记录

---

### 场景3: Git合并冲突

**目标**: 验证Git合并冲突检测

**步骤**:
1. Worker-1和Worker-2同时在同一文件的不同行编辑
2. Worker-1先合并成功
3. Worker-2合并时检测到冲突

**预期结果**: ✅ 检测到冲突，自动报告，等待高权限Agent处理

---

## 验证检查点

### ✅ Master API端点

- [ ] `POST /api/tasks/create-with-git` - 创建Git任务
- [ ] `POST /api/tasks/assign-git` - 分配Git任务
- [ ] `GET /api/tasks/:task_id/git` - 获取Git任务详情
- [ ] `POST /api/conflicts` - 报告冲突
- [ ] `GET /api/conflicts` - 获取待处理冲突
- [ ] `POST /api/conflicts/resolve` - 解决冲突

### ✅ Git操作

- [ ] 创建Git分支: `git checkout -b worker-xxx/TASK-001`
- [ ] 锁定文件: `git lfs lock story_001.md`
- [ ] 提交变更: `git commit -m "..."`
- [ ] 合并分支: `git merge worker-xxx/TASK-001`
- [ ] 检测冲突: `git diff --name-only --diff-filter=U`

### ✅ SQLite数据

- [ ] `git_tasks` 表记录任务状态
- [ ] `git_locks` 表记录文件锁
- [ ] `git_conflicts` 表记录冲突

---

## 故障排查

### 问题1: Master服务器启动失败

**检查**:
```bash
# 检查端口占用
lsof -i :8848

# 查看日志
tail -f logs/master.log
```

**解决**: 释放端口或更改Master端口配置

---

### 问题2: Git操作失败

**检查**:
```bash
# 检查Git是否安装
git --version

# 检查Git LFS是否安装
git lfs version

# 查看Git状态
git status
```

**解决**: 安装Git和Git LFS

---

### 问题3: 任务分配失败

**检查**:
```bash
# 查看任务详情
curl http://localhost:8848/api/tasks/STORY-001/git

# 查看冲突
curl http://localhost:8848/api/conflicts
```

**解决**:
- 任务可能已被分配，状态不是 `pending`
- 文件可能已被其他Worker锁定

---

## 性能测试

### 并发测试

```bash
# 同时创建10个任务
for i in {1..10}; do
  curl -X POST http://localhost:8848/api/tasks/create-with-git \
    -H "Content-Type: application/json" \
    -d "{
      \"task_id\": \"STORY-00$i\",
      \"title\": \"并发测试任务$i\",
      \"file_boundaries\": [
        {\"file_path\": \"story_00$i.md\", \"line_start\": 1, \"line_end\": 100}
      ]
    }" &
done
wait
```

### 压力测试

```bash
# 创建100个任务
seq 1 100 | parallel -j 10 curl -X POST http://localhost:8848/api/tasks/create-with-git \
  -H "Content-Type: application/json" \
  -d '{"task_id":"STORY-{}", "title":"任务{}", "file_boundaries":[{"file_path":"story_{}.md","line_start":1,"line_end":100}]}'
```

---

## 清理测试数据

```bash
# 删除测试目录
rm -rf /tmp/cpds-git-test-*

# 清空数据库
rm -f data/master.db

# 重启Master
go run cmd/master/main.go
```

---

## 总结

本测试指南涵盖了：

1. ✅ **快速测试**: 验证核心API功能
2. ✅ **完整测试**: 端到端的Git集成测试
3. ✅ **手动测试**: 灵活的API调用测试
4. ✅ **场景测试**: 覆盖正常和异常情况
5. ✅ **故障排查**: 常见问题解决方案
6. ✅ **性能测试**: 并发和压力测试

通过这些测试，可以全面验证Git+SQLite混合架构的：
- 任务创建和分配
- 文件锁管理
- 冲突检测和处理
- Git分支操作
- SQLite元数据管理

**测试完成后，你将看到5个完整的克苏鲁故事，每个故事都通过独立的Git分支并行创作完成！** 🎉
