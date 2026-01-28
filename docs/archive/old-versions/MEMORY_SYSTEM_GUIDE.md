# AgentFlow Memory System Guide

## 概述

AgentFlow 记忆系统为 Master 和 Worker 提供了完整的记忆管理功能，包括：

- **工作记忆**：任务执行的中间状态和上下文
- **长期记忆**：历史任务记录和经验总结
- **会话记忆**：对话历史和用户偏好
- **本地缓存**：Worker 的本地持久化记忆

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                     Master Server                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │         MemoryManager (SQLite)                    │  │
│  │  - task_context (工作记忆)                        │  │
│  │  - task_history (长期记忆)                        │  │
│  │  - experience_summaries (经验总结)                │  │
│  │  - conversations (会话记忆)                       │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                          ↕ (Memory Sync)
┌─────────────────────────────────────────────────────────┐
│                      Workers                            │
│  ┌───────────────────────────────────────────────────┐  │
│  │         WorkerMemory (本地持久化)                 │  │
│  │  - Map-based memory (短期)                       │  │
│  │  - File persistence (本地文件)                   │  │
│  │  - Auto-sync to Master (自动同步)                │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## 功能特性

### 1. 工作记忆 (Working Memory)

存储任务执行的中间状态，包括：
- 任务变量
- 依赖关系
- 执行历史

### 2. 长期记忆 (Long-term Memory)

存储历史任务和经验总结，包括：
- 任务历史记录
- 成功/失败模式
- 优化建议
- 最佳实践

### 3. 会话记忆 (Conversational Memory)

存储对话上下文，包括：
- 会话信息
- 消息历史
- 用户偏好

### 4. Worker 本地记忆

每个 Worker 维护自己的本地记忆：
- 内存中的快速访问
- 本地文件持久化
- 自动过期清理
- 与 Master 同步

## API 接口

### Memory API

#### 1. 同步 Worker 记忆

```bash
POST /api/v1/memory/sync
```

请求体：
```json
{
  "worker_id": "worker-hostname-123-abc",
  "memory_snapshot": {
    "entries": [...],
    "workerId": "worker-hostname-123-abc",
    "snapshotTime": "2025-01-27T10:00:00Z"
  }
}
```

#### 2. 获取任务记忆

```bash
GET /api/v1/memory/task/:id
```

响应：
```json
{
  "success": true,
  "data": {
    "variables": { "var1": "value1" },
    "dependencies": ["TASK-00000001"],
    "executionHistory": [...]
  }
}
```

#### 3. 保存任务上下文

```bash
POST /api/v1/memory/task/:id/context
```

请求体：
```json
{
  "context": {
    "variables": { "key": "value" },
    "dependencies": [],
    "executionHistory": []
  }
}
```

#### 4. 获取经验总结

```bash
GET /api/v1/memory/experiences?context=test&limit=5
```

#### 5. 保存经验总结

```bash
POST /api/v1/memory/experiences
```

请求体：
```json
{
  "summary": {
    "summary_type": "best_practice",
    "pattern_description": "Always test your code",
    "context": { "domain": "testing" },
    "confidence_score": 0.9
  }
}
```

#### 6. 获取记忆统计

```bash
GET /api/v1/memory/stats
```

#### 7. 清理过期记忆

```bash
POST /api/v1/memory/cleanup
```

## 使用示例

### Worker 记忆使用

```typescript
import { Worker } from '@agentflow/worker';

const worker = new Worker({
  master_url: 'http://localhost:6767',
  group_name: 'default'
});

// 存储信息到记忆
worker.remember('task-123:start', {
  taskId: 'TASK-00000123',
  startedAt: new Date().toISOString()
});

// 带分类和 TTL 的存储
worker.remember('task-123:result', { result: 'Success' }, {
  ttl: 3600, // 1 hour
  category: 'result',
  taskId: 'TASK-00000123'
});

// 检索信息
const result = worker.recall('task-123:result');

// 获取任务相关记忆
const taskMemory = worker.getTaskMemory('TASK-00000123');

// 获取记忆统计
const stats = worker.getMemoryStats();
```

### Master 记忆使用

```typescript
import { Master } from '@agentflow/master';

const master = new Master({
  port: 6767,
  db_path: '~/.agentflow/agentflow.db'
});

await master.start();

// 保存任务上下文
master.memoryManager?.saveTaskContext(1, {
  variables: { var1: 'value1' },
  dependencies: ['TASK-00000001'],
  executionHistory: []
});

// 获取任务上下文
const context = master.memoryManager?.getTaskContext(1);

// 记录任务到历史
master.memoryManager?.recordTaskToHistory(task, {
  status: 'completed',
  result: 'Task completed',
  durationMs: 1500,
  workerId: 'worker-1'
});

// 保存经验总结
master.memoryManager?.saveExperienceSummary({
  summary_type: 'best_practice',
  pattern_description: 'Always write tests',
  context: { domain: 'development' },
  confidence_score: 0.95
});

// 检索相关经验
const experiences = master.memoryManager?.retrieveRelevantExperiences('testing code', 5);
```

## 数据库 Schema

### SQLite Schema (Node.js)

参见：[nodejs/packages/database/src/memory-schema.ts](../nodejs/packages/database/src/memory-schema.ts)

### MySQL Schema (Go)

参见：[golang/internal/database/memory_schema.sql](../golang/internal/database/memory_schema.sql)

## 配置选项

### MemoryManager 配置

```typescript
interface MemoryConfig {
  enableVectorSearch: boolean;        // 启用向量搜索
  maxWorkingMemorySize: number;       // 最大工作记忆大小
  longTermRetentionDays: number;      // 长期记忆保留天数
  autoCleanup: boolean;               // 自动清理过期记忆
  cleanupIntervalHours: number;       // 清理间隔（小时）
}
```

### WorkerMemory 配置

```typescript
interface WorkerMemoryConfig {
  enablePersistence: boolean;         // 启用持久化
  persistencePath: string;            // 持久化路径
  syncInterval: number;               // 同步间隔（毫秒）
  maxMemoryEntries: number;           // 最大记忆条目数
}
```

## 最佳实践

### 1. 使用分类

```typescript
// 使用分类来组织记忆
worker.remember('task-1:start', data, { category: 'execution' });
worker.remember('task-1:result', data, { category: 'result' });
worker.remember('task-1:error', data, { category: 'error' });
```

### 2. 设置合理的 TTL

```typescript
// 短期记忆（1小时）
worker.remember('temp-data', data, { ttl: 3600 });

// 中期记忆（1天）
worker.remember('mid-term-data', data, { ttl: 86400 });

// 长期记忆（记录到 Master，不设置 TTL）
master.memoryManager?.recordTaskToHistory(task, result);
```

### 3. 关联任务

```typescript
// 使用 taskId 关联记忆
worker.remember('step-1', data, { taskId: 'TASK-00000123' });
worker.remember('step-2', data, { taskId: 'TASK-00000123' });

// 获取任务的所有记忆
const taskMemory = worker.getTaskMemory('TASK-00000123');
```

### 4. 定期清理

```typescript
// Worker 自动清理过期条目
worker.cleanMemory();

// Master 手动触发清理
await fetch('http://localhost:6767/api/v1/memory/cleanup', {
  method: 'POST'
});
```

## 测试

运行记忆系统测试：

```bash
cd nodejs
npm test -- packages/master/test/memory.test.ts
```

## 性能考虑

1. **内存限制**：Worker 默认最大 1000 条记忆条目，超出时自动淘汰最旧的
2. **持久化开销**：每次写入都会持久化到磁盘，批量操作时注意性能
3. **同步频率**：默认任务完成后同步，可根据需要调整
4. **清理策略**：自动清理过期条目，定期运行完整清理

## 故障排查

### 记忆未保存

- 检查 Worker 是否启用持久化：`enablePersistence: true`
- 检查文件路径是否有写权限
- 查看日志中的错误信息

### 记忆未同步

- 确认 Master 正在运行
- 检查网络连接
- 查看 API 响应状态码

### 记忆过期太快

- 增加 TTL 值
- 检查 `autoCleanup` 配置
- 使用 Master 的长期记忆存储重要数据

## 下一步

- 查看 [MemoryManager 源码](../nodejs/packages/master/src/memory-manager.ts)
- 查看 [WorkerMemory 源码](../nodejs/packages/worker/src/worker-memory.ts)
- 运行 [测试用例](../nodejs/packages/master/test/memory.test.ts)
