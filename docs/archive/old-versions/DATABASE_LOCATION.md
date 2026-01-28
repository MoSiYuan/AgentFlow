# AgentFlow 数据库位置说明

## 概述

AgentFlow 的数据库文件已统一放置在 `~/.claude/skills/` 目录下，与 Claude CLI 的 skills 文件（如 `agentflow.md`）位于同一位置，便于管理和查找。

## 文件结构

```
~/.claude/
└── skills/
    ├── agentflow.db              # AgentFlow 主数据库（SQLite）⭐
    ├── agentflow.md              # AgentFlow Skill 定义
    └── ...                       # 其他 skills
```

## 数据库详情

### AgentFlow 数据库 (`agentflow.db`)

**位置**: `~/.claude/skills/agentflow/agentflow.db`

**包含数据**:
- ✅ 任务表 (tasks)
- ✅ 任务链表 (task_chains, task_chain_nodes)
- ✅ Claude 映射表 (claude_mappings)
- ✅ Git 集成表 (git_tasks, git_branches, git_conflicts)
- ✅ Worker 表 (workers)
- ✅ 任务关系表 (task_relationships)
- ✅ 任务检查点表 (task_checkpoints)
- ✅ 任务版本表 (task_versions)

**表结构**:
- 15+ 核心表
- 完整索引优化
- 外键约束（CASCADE 删除）
- 事务支持

## Claude 会话文件

**位置**: `~/.claude/sessions/{session_uuid}.jsonl`

**用途**: 存储 Claude CLI 的对话历史和状态更新

**格式**: JSON Lines (每行一个 JSON 对象)

**AgentFlow 集成**:
- 状态同步器读写此文件
- 任务状态更新自动追加到会话文件
- 支持增量读取大文件

## 配置方法

### Go 版本

#### 1. 使用默认配置（推荐）
```bash
# 启动 Master，使用默认路径 ~/.claude/skills/agentflow/agentflow.db
./bin/master

# 启动 Worker，使用默认路径
./bin/worker
```

#### 2. 环境变量配置
```bash
# 设置自定义数据库路径
export MASTER_DB_PATH="/custom/path/agentflow.db"
export WORKER_DB_PATH="/custom/path/agentflow.db"

# 或使用相对路径
export MASTER_DB_PATH="~/.claude/skills/agentflow/agentflow.db"
export WORKER_DB_PATH="~/.claude/skills/agentflow/agentflow.db"
```

#### 3. 命令行参数
```bash
# 通过命令行指定路径
./bin/master --db ~/.claude/skills/agentflow/agentflow.db
./bin/worker --db ~/.claude/skills/agentflow/agentflow.db
```

#### 4. 配置文件
创建 `~/.claude/agentflow.yaml`:
```yaml
master:
  host: "0.0.0.0"
  port: 8848
  db_path: "~/.claude/skills/agentflow/agentflow.db"  # 支持 ~ 展开为用户主目录
  auto_start: false

worker:
  master_url: "http://localhost:8848"
  db_path: "~/.claude/skills/agentflow/agentflow.db"
  group_name: "default"
  mode: "auto"

claude:
  server_url: "http://localhost:8849"
  max_tokens: 4096
  temperature: 0.7
  timeout: 120
```

使用配置文件:
```bash
./bin/master --config ~/.claude/agentflow.yaml
```

### Node.js 版本

#### 1. 使用默认配置（推荐）
```bash
# 启动 Master，使用默认路径 ~/.claude/skills/agentflow/agentflow.db
cd nodejs/packages/master
npm start

# 或直接运行
node dist/index.js
```

#### 2. 环境变量配置
```bash
# 设置数据库路径
export AGENTFLOW_DB_PATH="~/.claude/skills/agentflow/agentflow.db"
```

#### 3. 命令行参数
```bash
# 通过 --db 参数指定路径
node dist/index.js --db ~/.claude/skills/agentflow/agentflow.db
```

#### 4. 代码配置
```typescript
import { Master } from '@agentflow/master';

const master = new Master({
  db_path: '~/.claude/skills/agentflow/agentflow.db',  // 支持路径展开
  port: 6767,
  host: '0.0.0.0'
});

await master.start();
```

## 路径展开

### 支持的路径格式

1. **绝对路径**:
   ```
   /Users/username/.claude/agentflow.db
   ```

2. **用户主目录（~）**:
   ```
   ~/.claude/skills/agentflow/agentflow.db  # 自动展开为 /Users/username/.claude/agentflow.db
   ```

3. **相对路径**:
   ```
   ./data/agentflow.db  # 相对于当前工作目录
   ```

### Go 版本路径展开

```go
// 在 internal/config/config.go 中自动处理
// 支持 ~ 开头的路径自动展开

if len(path) > 0 && path[0] == '~' {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("failed to get home directory: %w", err)
    }
    path = filepath.Join(homeDir, path[1:])
}
```

### Node.js 版本路径展开

Node.js 版本在数据库初始化时自动处理：

```typescript
import { homedir } from 'os';
import { resolve } from 'path';

// 在 AgentFlowDatabase 中自动展开 ~
let dbPath = config.db_path;
if (dbPath.startsWith('~')) {
  dbPath = resolve(homedir(), dbPath.slice(1));
}
```

## 查看数据库位置

### 方法 1: 通过 API 查询

```bash
# Go 版本
curl http://localhost:8848/api/v1/stats

# Node.js 版本
curl http://localhost:6767/api/status
```

返回结果包含数据库路径信息。

### 方法 2: 查看进程打开的文件

```bash
# macOS/Linux
lsof -p $(pgrep -f agentflow) | grep \.db

# 或使用 ps 查看命令行参数
ps aux | grep agentflow
```

### 方法 3: 直接检查文件

```bash
# 检查数据库文件是否存在
ls -lh ~/.claude/skills/agentflow/agentflow.db

# 查看文件大小
du -h ~/.claude/skills/agentflow/agentflow.db

# 查看文件修改时间
stat ~/.claude/skills/agentflow/agentflow.db
```

## 数据库管理

### 备份数据库

```bash
# 方法 1: 直接复制
cp ~/.claude/skills/agentflow/agentflow.db ~/.claude/skills/agentflow/agentflow.db.backup.$(date +%Y%m%d)

# 方法 2: 使用 SQLite 备份命令
sqlite3 ~/.claude/skills/agentflow/agentflow.db ".backup ~/.claude/skills/agentflow/agentflow.db.backup"

# 方法 3: 导出 SQL
sqlite3 ~/.claude/skills/agentflow/agentflow.db .dump > ~/.claude/agentflow.sql
```

### 恢复数据库

```bash
# 方法 1: 从备份恢复
cp ~/.claude/skills/agentflow/agentflow.db.backup.20260126 ~/.claude/skills/agentflow/agentflow.db

# 方法 2: 从 SQL 导入
sqlite3 ~/.claude/skills/agentflow/agentflow.db < ~/.claude/agentflow.sql
```

### 清理数据库

```bash
# 完全删除（谨慎操作）
rm ~/.claude/skills/agentflow/agentflow.db

# 清空所有数据（保留表结构）
sqlite3 ~/.claude/skills/agentflow/agentflow.db "DELETE FROM tasks;"
sqlite3 ~/.claude/skills/agentflow/agentflow.db "DELETE FROM task_chains;"
sqlite3 ~/.claude/skills/agentflow/agentflow.db "DELETE FROM claude_mappings;"
# ... 其他表
```

### 查看数据库内容

```bash
# 打开 SQLite 命令行
sqlite3 ~/.claude/skills/agentflow/agentflow.db

# 查看所有表
.tables

# 查看表结构
.schema tasks

# 执行查询
SELECT id, title, status FROM tasks LIMIT 10;

# 查看数据库统计信息
SELECT 'tasks' as table_name, COUNT(*) as count FROM tasks
UNION ALL
SELECT 'claude_mappings', COUNT(*) FROM claude_mappings
UNION ALL
SELECT 'task_chains', COUNT(*) FROM task_chains;
```

## 性能优化

### 索引优化

数据库已创建以下索引以提高查询性能：

```sql
-- Claude 映射索引
CREATE INDEX idx_claude_session ON claude_mappings(session_uuid);
CREATE INDEX idx_claude_message ON claude_mappings(message_uuid);
CREATE INDEX idx_claude_task ON claude_mappings(task_id);

-- 任务链索引
CREATE INDEX idx_task_chains_session ON task_chains(session_uuid);
CREATE INDEX idx_task_chains_status ON task_chains(status);
CREATE INDEX idx_task_chain_nodes_chain ON task_chain_nodes(chain_id);
```

### 数据库大小管理

**预期大小**:
- 空数据库: ~100 KB
- 1000个任务: ~1-2 MB
- 10000个任务: ~10-20 MB
- 100000个任务: ~100-200 MB

**优化建议**:
1. 定期清理已完成的旧任务
2. 定期执行 `VACUUM` 命令压缩数据库
3. 考虑分表存储历史数据

```bash
# 压缩数据库
sqlite3 ~/.claude/skills/agentflow/agentflow.db "VACUUM;"
```

## 故障排查

### 问题 1: 数据库锁定

**症状**: "database is locked" 错误

**解决方案**:
```bash
# 检查是否有其他进程在使用
lsof ~/.claude/skills/agentflow/agentflow.db

# 确保只有一个 Master 进程在运行
pgrep -f agentflow

# 重启进程
```

### 问题 2: 权限错误

**症状**: "permission denied" 错误

**解决方案**:
```bash
# 检查文件权限
ls -la ~/.claude/skills/agentflow/agentflow.db

# 修复权限
chmod 644 ~/.claude/skills/agentflow/agentflow.db
chown $USER:$USER ~/.claude/skills/agentflow/agentflow.db
```

### 问题 3: 数据库损坏

**症状**: "database disk image is malformed" 错误

**解决方案**:
```bash
# 尝试修复
sqlite3 ~/.claude/skills/agentflow/agentflow.db "PRAGMA integrity_check;"
sqlite3 ~/.claude/skills/agentflow/agentflow.db "VACUUM;"

# 如果修复失败，从备份恢复
cp ~/.claude/skills/agentflow/agentflow.db.backup.LATEST ~/.claude/skills/agentflow/agentflow.db
```

## 相关文档

- [架构设计](../ARCHITECTURE.md)
- [Claude 集成文档](./CLAUDE_INTEGRATION.md)
- [部署指南](../deployment/README.md)
- [API 文档](./API.md)

---

**最后更新**: 2026-01-26
**版本**: 1.0
