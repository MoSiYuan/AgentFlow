# AgentFlow 与 Claude 深度集成方案

## 一、Claude Task 系统深度分析

### 1.1 核心发现

通过深入分析 Claude CLI 的数据结构，发现了以下关键机制：

#### Claude ID 生成系统

**三层 ID 体系**:

1. **Session ID** (UUID格式)
   ```json
   "sessionId": "312e5ca4-5dde-4f47-a6f4-e46bf12b7827"
   ```
   - 全局唯一标识符
   - 用于整个会话周期
   - 格式: 标准UUID v4

2. **Message UUID** (UUID格式)
   ```json
   "uuid": "1c075c2a-1731-4d27-88c3-b6a7bb8664c9",
   "parentUuid": null // 根消息
   ```
   - 每条消息的唯一标识
   - parentUuid 构成消息链/树
   - 支持分支和并行对话

3. **Slug** (友好名称)
   ```json
   "slug": "immutable-brewing-karp"
   ```
   - 人类可读的友好名称
   - 用于 todos 目录
   - 格式: `adjective-noun-name` (3个随机单词)

#### Task/Todo 系统

**数据结构**:
```json
{
  "oldTodos": [
    {
      "content": "研究 Claude CLI task 系统",
      "status": "completed",
      "activeForm": "研究 Claude CLI task 系统"
    }
  ],
  "newTodos": [
    {
      "content": "分析 todos 目录结构",
      "status": "in_progress",
      "activeForm": "分析 todos 目录结构"
    }
  ]
}
```

**特性**:
- ✅ 原子性更新（oldTodos → newTodos）
- ✅ 状态跟踪（pending, in_progress, completed）
- ✅ 活跃形式（activeForm）用于显示
- ✅ 通过 toolUseResult 传递

#### 数据存储格式

**JSON Lines (.jsonl)**:
```
{"type":"user", "uuid":"...", "timestamp":"..."}
{"type":"assistant", "uuid":"...", "parentUuid":"...", "content":[...]}
{"type":"tool_use", "uuid":"...", "name":"Bash", "input":{...}}
{"type":"tool_result", "uuid":"...", "content":"..."}
```

**特点**:
- 每行一个独立的 JSON 对象
- 按时间顺序追加
- 支持流式读取和解析
- 易于备份和迁移

### 1.2 Claude Task 系统架构

```
Claude CLI Task System
├── Session Layer (会话层)
│   ├── Session ID (UUID)
│   ├── Project Directory
│   └── Metadata (branch, cwd, version)
├── Message Layer (消息层)
│   ├── Message UUID
│   ├── Parent UUID (tree structure)
│   ├── Type (user/assistant/tool_use/tool_result)
│   └── Content (text/thinking/tool_use)
└── Task Layer (任务层)
    ├── Slug (friendly name)
    ├── Content (task description)
    ├── Status (pending/in_progress/completed)
    └── Active Form (display text)
```

### 1.3 关键机制分析

#### 1.3.1 消息链机制

```json
{
  "uuid": "msg-3",
  "parentUuid": "msg-2"
}
{
  "uuid": "msg-2",
  "parentUuid": "msg-1"
}
{
  "uuid": "msg-1",
  "parentUuid": null
}
```

**优势**:
- ✅ 支持完整的历史追溯
- ✅ 支持分支和合并
- ✅ 易于实现会话恢复

#### 1.3.2 原子性更新

```json
{
  "toolUseResult": {
    "oldTodos": [...],
    "newTodos": [...]
  }
}
```

**优势**:
- ✅ 状态一致性保证
- ✅ 支持回滚
- ✅ 易于调试和审计

#### 1.3.3 流式存储

```
Session File (.jsonl):
Line 1: {"type":"queue-operation",...}
Line 2: {"type":"user",...}
Line 3: {"type":"assistant",...}
...
Line 11698: {"type":"tool_result",...}
```

**优势**:
- ✅ 支持大文件（40MB+）
- ✅ 流式读取，内存友好
- ✅ 易于追加新数据
- ✅ 支持并发写入

## 二、AgentFlow 当前系统分析

### 2.1 AgentFlow ID 机制

**当前实现**:
```typescript
// Task ID Format
"TASK-00000001" // 格式化字符串

// Database
id: INTEGER PRIMARY KEY AUTOINCREMENT
```

**特点**:
- ✅ 简洁易读
- ✅ 数据库原生支持
- ❌ 不支持分布式
- ❌ 无法与 Claude 对接

### 2.2 AgentFlow Task 系统

**数据结构**:
```typescript
interface Task {
  id: number;                    // 数据库自增ID
  title: string;                 // 任务标题
  description: string;           // 任务描述
  status: 'pending' | 'running' | 'completed' | 'failed';
  priority: 'high' | 'medium' | 'low';
  group_name: string;            // Worker组
  result: string | null;         // 执行结果
  error: string | null;          // 错误信息
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
}
```

**特点**:
- ✅ 关系型数据库存储
- ✅ 支持复杂查询
- ✅ 事务支持
- ❌ 无消息链
- ❌ 无分支支持

## 三、深度集成方案设计

### 3.1 方案概述

**核心思路**: **混合 ID 系统 + 双向同步**

```
┌─────────────────┐         ┌─────────────────┐
│   Claude CLI    │ ◄────► │   AgentFlow     │
│                 │         │                 │
│  Session UUID   │ ─────► │  Task Mapping   │
│  Message UUID   │ ─────► │  Execution ID   │
│  Slug Name      │ ◄───── │  Friendly ID    │
└─────────────────┘         └─────────────────┘
```

### 3.2 混合 ID 系统

#### 3.2.1 ID 映射表

**数据库扩展**:
```sql
CREATE TABLE task_mappings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  agentflow_task_id INTEGER NOT NULL,  -- AgentFlow 原生ID
  claude_session_id TEXT NOT NULL,     -- Claude Session UUID
  claude_message_uuid TEXT NOT NULL,   -- Claude Message UUID
  claude_slug TEXT,                    -- Claude Slug
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (agentflow_task_id) REFERENCES tasks(id)
);

CREATE INDEX idx_claude_session ON task_mappings(claude_session_id);
CREATE INDEX idx_claude_message ON task_mappings(claude_message_uuid);
CREATE INDEX idx_claude_slug ON task_mappings(claude_slug);
```

#### 3.2.2 双向 ID 查询

**AgentFlow → Claude**:
```typescript
async getClaudeContext(taskId: number): Promise<ClaudeContext> {
  const mapping = await this.db.getTaskMapping(taskId);
  return {
    sessionId: mapping.claude_session_id,
    messageUuid: mapping.claude_message_uuid,
    slug: mapping.claude_slug,
    sessionFile: `~/.claude/projects/-${project}/${mapping.claude_session_id}.jsonl`
  };
}
```

**Claude → AgentFlow**:
```typescript
async getAgentFlowTask(claudeMessageUuid: string): Promise<Task> {
  const mapping = await this.db.getMappingByMessageUuid(claudeMessageUuid);
  return await this.db.getTask(mapping.agentflow_task_id);
}
```

### 3.3 统一 Task API

#### 3.3.1 扩展 Task 接口

```typescript
interface UnifiedTask {
  // AgentFlow 原有字段
  id: number;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;

  // Claude 集成字段
  claude: {
    sessionId: string;        // Claude Session UUID
    messageUuid: string;      // Claude Message UUID
    parentUuid?: string;      // Claude Parent UUID (支持任务链)
    slug: string;             // 友好名称
    thread?: string[];        // 消息线程（历史）
  };

  // 元数据
  metadata: {
    source: 'claude' | 'api' | 'cli';
    createdAt: string;
    startedAt?: string;
    completedAt?: string;
  };
}
```

#### 3.3.2 创建任务（集成模式）

```typescript
async createTaskFromClaude(request: ClaudeTaskRequest): Promise<UnifiedTask> {
  // 1. 生成 Claude IDs
  const sessionId = request.sessionId || uuidv4();
  const messageUuid = uuidv4();
  const slug = generateSlug(); // 生成友好名称

  // 2. 创建 AgentFlow Task
  const agentflowTask = await this.db.createTask({
    title: request.title,
    description: request.description,
    priority: request.priority || 'medium',
    group_name: request.group || 'default'
  });

  // 3. 创建映射
  await this.db.createMapping({
    agentflow_task_id: agentflowTask.id,
    claude_session_id: sessionId,
    claude_message_uuid: messageUuid,
    claude_slug: slug
  });

  // 4. 返回统一任务对象
  return {
    ...agentflowTask,
    claude: {
      sessionId,
      messageUuid,
      slug,
      parentUuid: request.parentUuid
    },
    metadata: {
      source: 'claude',
      createdAt: new Date().toISOString()
    }
  };
}
```

### 3.4 消息链同步

#### 3.4.1 任务链支持

```typescript
interface TaskChain {
  rootTaskId: number;
  tasks: UnifiedTask[];
  chainType: 'sequential' | 'parallel' | 'tree';
}

async createTaskChain(request: TaskChainRequest): Promise<TaskChain> {
  const sessionId = uuidv4();
  const rootUuid = uuidv4();
  const rootSlug = generateSlug();

  const tasks: UnifiedTask[] = [];
  let parentUuid = rootUuid;

  // 创建任务链
  for (const taskDef of request.tasks) {
    const task = await this.createTaskFromClaude({
      ...taskDef,
      sessionId,
      parentUuid, // 链式依赖
      parentTaskId: tasks.length > 0 ? tasks[tasks.length - 1].id : undefined
    });

    tasks.push(task);
    parentUuid = task.claude.messageUuid;
  }

  return {
    rootTaskId: tasks[0].id,
    tasks,
    chainType: request.chainType
  };
}
```

#### 3.4.2 分支任务支持

```typescript
// 从现有任务创建分支
async createBranch(
  parentTaskId: number,
  branchTasks: TaskDefinition[]
): Promise<TaskChain> {
  const parentTask = await this.getTask(parentTaskId);
  const parentMapping = await this.db.getTaskMapping(parentTaskId);

  // 使用父任务的消息UUID作为父UUID
  return await this.createTaskChain({
    tasks: branchTasks,
    sessionId: parentMapping.claude_session_id,
    parentUuid: parentMapping.claude_message_uuid
  });
}
```

### 3.5 实时状态同步

#### 3.5.1 Claude → AgentFlow 状态同步

```typescript
// Claude Skill 更新任务状态
async syncFromClaude(update: ClaudeTaskUpdate): Promise<void> {
  const mapping = await this.db.getMappingByMessageUuid(update.messageUuid);

  await this.db.updateTaskStatus(mapping.agentflow_task_id, {
    status: update.status,
    result: update.result,
    error: update.error
  });

  // 触发 Webhook 通知 Worker
  if (update.status === 'pending') {
    await this.notifyWorkers(mapping.agentflow_task_id);
  }
}
```

#### 3.5.2 AgentFlow → Claude 状态同步

```typescript
// Worker 执行完成后同步回 Claude
async syncToClaude(taskId: number, result: TaskResult): Promise<void> {
  const mapping = await this.db.getTaskMapping(taskId);

  // 写入 Claude 会话文件
  const sessionFile = path.join(
    os.homedir(),
    '.claude',
    'projects',
    `-${sanitize(process.cwd())}`,
    `${mapping.claude_session_id}.jsonl`
  );

  const statusMessage = {
    type: 'agentflow_update',
    timestamp: new Date().toISOString(),
    taskId: taskId,
    claudeMessageUuid: mapping.claude_message_uuid,
    status: result.status,
    result: result.output,
    error: result.error
  };

  await fs.appendFile(sessionFile, JSON.stringify(statusMessage) + '\n');
}
```

### 3.6 统一查询 API

#### 3.6.1 多维度查询

```typescript
interface TaskQuery {
  // AgentFlow 查询
  taskId?: number;
  status?: TaskStatus;
  priority?: TaskPriority;
  group?: string;

  // Claude 查询
  sessionId?: string;
  messageUuid?: string;
  slug?: string;

  // 时间范围
  createdAfter?: string;
  createdBefore?: string;
}

async queryTasks(query: TaskQuery): Promise<UnifiedTask[]> {
  // 构建复杂查询
  const sql = `
    SELECT
      t.*,
      m.claude_session_id as sessionId,
      m.claude_message_uuid as messageUuid,
      m.claude_slug as slug
    FROM tasks t
    LEFT JOIN task_mappings m ON t.id = m.agentflow_task_id
    WHERE 1=1
      ${query.taskId ? 'AND t.id = ?' : ''}
      ${query.status ? 'AND t.status = ?' : ''}
      ${query.sessionId ? 'AND m.claude_session_id = ?' : ''}
      ${query.messageUuid ? 'AND m.claude_message_uuid = ?' : ''}
      ${query.slug ? 'AND m.claude_slug = ?' : ''}
    ORDER BY t.created_at DESC
  `;

  return await this.db.all(sql, /* bind parameters */);
}
```

### 3.7 Claude Skill 增强

#### 3.7.1 任务链 Skill

```markdown
---
name: agentflow-chain
description: Create and manage task chains in AgentFlow
parameters:
  - name: tasks
    type: array
    required: true
  - name: chainType
    type: string
    description: "sequential|parallel|tree"
---

## Task Chain Examples

### Sequential Chain
\`\`\`bash
agentflow chain sequential '[
  {"title": "Test", "description": "npm test"},
  {"title": "Build", "description": "npm run build"},
  {"title": "Deploy", "description": "./deploy.sh"}
]'
\`\`\`

### Parallel Chain
\`\`\`bash
agentflow chain parallel '[
  {"title": "Frontend Tests", "description": "cd frontend && npm test"},
  {"title": "Backend Tests", "description": "cd backend && npm test"}
]'
\`\`\`
```

#### 3.7.2 分支任务 Skill

```markdown
---
name: agentflow-branch
description: Create branch from existing task
---

## Branch Task

\`\`\`bash
# Create alternative implementation
agentflow branch TASK-00000001 '[
  {"title": "Try Approach B", "description": "implement method B"}
]'

# A/B testing
agentflow branch TASK-00000002 '[
  {"title": "Test Alternative", "description": "test alternative approach"}
]'
\`\`\`
```

## 四、实施路线图

### 4.1 阶段一：基础集成（1-2周）

**目标**: 实现 ID 映射和基本同步

- ✅ 创建 task_mappings 表
- ✅ 实现双向 ID 查询
- ✅ 扩展 Task 接口
- ✅ 基础状态同步

**交付物**:
- 数据库迁移脚本
- 更新的 Task API
- 集成测试

### 4.2 阶段二：任务链支持（2-3周）

**目标**: 支持复杂任务依赖关系

- ✅ 实现任务链创建
- ✅ 支持串行/并行执行
- ✅ 分支任务支持
- ✅ 任务树可视化

**交付物**:
- 任务链 API
- Claude Skills 更新
- 文档和示例

### 4.3 阶段三：深度集成（3-4周）

**目标**: 完全融合两个系统

- ✅ 实时状态同步
- ✅ 统一查询 API
- ✅ 会话文件集成
- ✅ Web Dashboard

**交付物**:
- 完整集成系统
- 可视化界面
- 性能优化

### 4.4 阶段四：高级特性（4-6周）

**目标**: 企业级功能

- ✅ 任务模板
- ✅ 定时任务
- ✅ 任务调度器
- ✅ 监控和告警

**交付物**:
- 企业级功能
- 运维工具
- 完整文档

## 五、技术挑战和解决方案

### 5.1 ID 冲突处理

**挑战**: 两个系统的 ID 格式完全不同

**解决方案**:
- 使用映射表解耦
- 双向查询优化
- 缓存常用映射

```typescript
class IDMappingCache {
  private cache = new Map<string, number>();

  async get(messageUuid: string): Promise<number> {
    if (this.cache.has(messageUuid)) {
      return this.cache.get(messageUuid)!;
    }

    const mapping = await this.db.getMappingByMessageUuid(messageUuid);
    this.cache.set(messageUuid, mapping.agentflow_task_id);
    return mapping.agentflow_task_id;
  }
}
```

### 5.2 数据一致性

**挑战**: 两个系统可能状态不同步

**解决方案**:
- 定期同步任务
- 事件驱动更新
- 最终一致性模型

```typescript
class TaskSynchronizer {
  async syncTask(taskId: number): Promise<void> {
    const afTask = await this.agentflow.getTask(taskId);
    const claudeCtx = await this.agentflow.getClaudeContext(taskId);

    // 比对状态
    if (afTask.status !== claudeCtx.status) {
      // 决定以谁为准
      const source = this.decideSource(afTask, claudeCtx);
      await this.syncTo(source, afTask, claudeCtx);
    }
  }
}
```

### 5.3 性能优化

**挑战**: 频繁的映射查询影响性能

**解决方案**:
- 多级缓存
- 批量查询
- 异步同步

```typescript
class CachedTaskRepository {
  private l1Cache = new Map<number, UnifiedTask>(); // 内存
  private l2Cache: Redis; // Redis

  async getTask(taskId: number): Promise<UnifiedTask> {
    // L1: 内存缓存
    if (this.l1Cache.has(taskId)) {
      return this.l1Cache.get(taskId)!;
    }

    // L2: Redis 缓存
    const cached = await this.l2Cache.get(`task:${taskId}`);
    if (cached) {
      this.l1Cache.set(taskId, JSON.parse(cached));
      return JSON.parse(cached);
    }

    // L3: 数据库
    const task = await this.db.getTask(taskId);
    const unified = await this.enrichWithClaudeInfo(task);

    // 回写缓存
    this.l1Cache.set(taskId, unified);
    await this.l2Cache.set(`task:${taskId}`, JSON.stringify(unified));

    return unified;
  }
}
```

## 六、可行性评估

### 6.1 技术可行性 ⭐⭐⭐⭐⭐

✅ **完全可行**

**理由**:
1. Claude 使用标准 JSON 格式，易于解析
2. UUID 系统成熟可靠
3. JSON Lines 格式支持流式处理
4. 数据库扩展简单直接

### 6.2 工作量评估

| 阶段 | 工作量 | 复杂度 | 风险 |
|------|--------|--------|------|
| 基础集成 | 1-2周 | 中 | 低 |
| 任务链支持 | 2-3周 | 高 | 中 |
| 深度集成 | 3-4周 | 高 | 中 |
| 高级特性 | 4-6周 | 高 | 高 |
| **总计** | **10-15周** | - | - |

### 6.3 优势分析

#### 相比浅层集成

| 特性 | 浅层集成 | 深度集成 |
|------|----------|----------|
| ID 互通 | ❌ | ✅ |
| 任务链 | ❌ | ✅ |
| 状态同步 | 单向 | 双向 |
| 历史追溯 | ❌ | ✅ |
| 分支支持 | ❌ | ✅ |
| 可视化 | 基础 | 完整 |

#### 核心优势

✅ **真正的统一**
- 两个系统无缝协作
- 用户无需关心底层实现
- 统一的任务视图

✅ **强大的能力**
- 支持复杂任务依赖
- 支持分支和并行
- 完整的历史追溯

✅ **企业级特性**
- 事务支持
- 数据一致性
- 性能优化

### 6.4 潜在风险

#### 6.4.1 复杂度增加

**风险**: 系统复杂度显著提升

**缓解措施**:
- 清晰的架构分层
- 完善的文档
- 充分的测试

#### 6.4.2 性能影响

**风险**: 额外的映射查询影响性能

**缓解措施**:
- 多级缓存
- 异步同步
- 批量操作

#### 6.4.3 维护成本

**风险**: 两个系统的版本升级可能导致兼容性问题

**缓解措施**:
- 版本兼容性测试
- 抽象层隔离变化
- 定期的集成测试

## 七、推荐方案

### 7.1 分阶段实施

**阶段 1**: MVP（最小可行产品）
- 实现基础的 ID 映射
- 支持简单的状态同步
- 验证核心概念

**阶段 2**: 增强功能
- 添加任务链支持
- 优化性能
- 完善文档

**阶段 3**: 企业级
- 完整的深度集成
- 可视化界面
- 运维工具

### 7.2 技术选型

**数据库**: SQLite (现有) + 映射表
**缓存**: Redis (可选)
**同步**: 事件驱动 + 定期同步
**API**: REST + WebSocket (实时更新)

### 7.3 优先级建议

**高优先级**:
1. ✅ ID 映射系统
2. ✅ 基础状态同步
3. ✅ 统一查询 API

**中优先级**:
4. ✅ 任务链支持
5. ✅ 性能优化
6. ✅ 监控和日志

**低优先级**:
7. ⏳ Web Dashboard
8. ⏳ 高级调度
9. ⏳ 多租户支持

## 八、总结

### 8.1 核心观点

1. **✅ 技术上完全可行**
   - Claude 的系统架构清晰
   - ID 机制成熟可靠
   - 数据格式易于集成

2. **✅ 带来显著价值**
   - 真正的系统统一
   - 强大的任务管理能力
   - 企业级特性支持

3. **⚠️ 需要投入资源**
   - 开发周期: 10-15周
   - 技术复杂度: 高
   - 维护成本: 中

### 8.2 建议

**对于小团队/快速原型**:
- 推荐使用当前的**浅层集成**方案
- 简单、直接、足够用

**对于企业级应用**:
- 推荐**深度集成**方案
- 投入资源，长期收益

**分阶段策略**:
1. 先实现 MVP 验证概念
2. 根据使用反馈决定是否继续
3. 逐步增强功能

### 8.3 下一步行动

**立即行动**:
1. ✅ 创建技术原型
2. ✅ 编写 POC 代码
3. ✅ 验证核心假设

**短期计划** (1-2周):
4. ✅ 设计数据库 schema
5. ✅ 实现映射表
6. ✅ 基础同步功能

**中期计划** (1-2月):
7. ✅ 完整的任务链支持
8. ✅ 性能优化
9. ✅ 文档和示例

---

**最后更新**: 2026-01-24
**作者**: AgentFlow Team
**状态**: 深度分析完成，等待决策
**版本**: 1.0.0
