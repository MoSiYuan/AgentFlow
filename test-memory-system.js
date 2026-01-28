#!/usr/bin/env node

/**
 * AgentFlow Memory System Demo
 *
 * 演示记忆系统的各项功能
 */

const Database = require('better-sqlite3');
const path = require('path');

// Import MemoryManager (需要先编译)
// const { MemoryManager } = require('./nodejs/packages/master/dist/memory-manager');

console.log('╔════════════════════════════════════════════════════════════╗');
console.log('║       AgentFlow Memory System Demo                       ║');
console.log('╚════════════════════════════════════════════════════════════╝\n');

// 1. Worker Memory Demo
console.log('📦 1. Worker Memory Demo');
console.log('───────────────────────────────────────────────────────────');

class WorkerMemory {
  constructor(workerId) {
    this.workerId = workerId;
    this.memory = new Map();
  }

  remember(key, value, options = {}) {
    const entry = {
      value,
      expiresAt: options.ttl ? Date.now() + options.ttl * 1000 : null,
      category: options.category || 'context',
      taskId: options.taskId,
      timestamp: Date.now()
    };

    this.memory.set(key, entry);
    console.log(`  ✓ Stored: ${key} (${entry.category})`);
  }

  recall(key) {
    const entry = this.memory.get(key);
    if (!entry) return null;

    if (entry.expiresAt && Date.now() > entry.expiresAt) {
      this.memory.delete(key);
      return null;
    }

    return entry.value;
  }

  getTaskMemory(taskId) {
    const results = [];
    for (const [key, entry] of this.memory.entries()) {
      if (entry.taskId === taskId) {
        results.push({ key, value: entry.value, category: entry.category });
      }
    }
    return results;
  }

  getStats() {
    const byCategory = {};
    for (const entry of this.memory.values()) {
      byCategory[entry.category] = (byCategory[entry.category] || 0) + 1;
    }
    return {
      totalEntries: this.memory.size,
      byCategory
    };
  }
}

const workerMemory = new WorkerMemory('demo-worker-1');

// 存储任务信息
workerMemory.remember('task-001:start', {
  taskId: 'TASK-000001',
  title: 'Build project',
  startedAt: new Date().toISOString()
}, { category: 'execution', taskId: 'TASK-000001' });

workerMemory.remember('task-001:progress', {
  step: 5,
  total: 10,
  status: 'running'
}, { category: 'context', taskId: 'TASK-000001' });

// 检索信息
const taskStart = workerMemory.recall('task-001:start');
console.log(`  → Retrieved: ${taskStart.title} started at ${taskStart.startedAt}`);

// 获取任务记忆
const taskMemory = workerMemory.getTaskMemory('TASK-000001');
console.log(`  → Task memory has ${taskMemory.length} entries`);

// 获取统计
const stats = workerMemory.getStats();
console.log(`  → Stats: ${stats.totalEntries} total entries`);
console.log(`    - Execution: ${stats.byCategory.execution || 0}`);
console.log(`    - Context: ${stats.byCategory.context || 0}\n`);

// 2. Master Memory Demo
console.log('📊 2. Master Memory Manager');
console.log('───────────────────────────────────────────────────────────');

// 创建内存数据库
const db = new Database(':memory:');
db.pragma('foreign_keys = ON');

// 初始化表
db.exec(`
  CREATE TABLE IF NOT EXISTS task_context (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER UNIQUE NOT NULL,
    context_data TEXT NOT NULL,
    variables TEXT,
    dependencies TEXT,
    execution_history TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS task_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    status TEXT,
    result TEXT,
    duration_ms INTEGER,
    worker_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
  );

  CREATE TABLE IF NOT EXISTS experience_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    summary_type TEXT NOT NULL,
    pattern_description TEXT NOT NULL,
    context TEXT,
    confidence_score REAL DEFAULT 0.5,
    usage_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
`);

// 简单的 MemoryManager 实现
class SimpleMemoryManager {
  constructor(db) {
    this.db = db;
  }

  saveTaskContext(taskId, context) {
    const stmt = this.db.prepare(`
      INSERT OR REPLACE INTO task_context
      (task_id, context_data, variables, dependencies, execution_history, updated_at)
      VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
    `);

    stmt.run(
      taskId,
      JSON.stringify(context),
      JSON.stringify(context.variables || {}),
      JSON.stringify(context.dependencies || []),
      JSON.stringify(context.executionHistory || [])
    );

    console.log(`  ✓ Saved context for task ${taskId}`);
  }

  recordTaskToHistory(task, result) {
    const stmt = this.db.prepare(`
      INSERT INTO task_history
      (task_id, title, status, result, duration_ms, worker_id, completed_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);

    stmt.run(
      task.id,
      task.title,
      result.status,
      JSON.stringify(result.result).substring(0, 1000),
      result.durationMs,
      result.workerId,
      result.completedAt || new Date().toISOString()
    );

    console.log(`  ✓ Recorded task "${task.title}" to history`);
  }

  saveExperienceSummary(summary) {
    const stmt = this.db.prepare(`
      INSERT INTO experience_summaries
      (summary_type, pattern_description, context, confidence_score)
      VALUES (?, ?, ?, ?)
    `);

    stmt.run(
      summary.summary_type,
      summary.pattern_description,
      JSON.stringify(summary.context || {}),
      summary.confidence_score
    );

    console.log(`  ✓ Saved experience: "${summary.pattern_description}"`);
  }
}

const memoryManager = new SimpleMemoryManager(db);

// 保存任务上下文
memoryManager.saveTaskContext(1, {
  variables: {
    buildType: 'production',
    target: 'es2015'
  },
  dependencies: ['TASK-000002'],
  executionHistory: [
    { step: 1, action: 'install', result: 'success' },
    { step: 2, action: 'build', result: 'in_progress' }
  ]
});

// 记录任务到历史
memoryManager.recordTaskToHistory(
  { id: 1, title: 'Build project' },
  {
    status: 'completed',
    result: { output: 'Build completed successfully' },
    durationMs: 1500,
    workerId: 'demo-worker-1',
    completedAt: new Date().toISOString()
  }
);

// 保存经验总结
memoryManager.saveExperienceSummary({
  summary_type: 'best_practice',
  pattern_description: 'Always run tests before building',
  context: { domain: 'build', phase: 'pre-build' },
  confidence_score: 0.95
});

memoryManager.saveExperienceSummary({
  summary_type: 'optimization',
  pattern_description: 'Use parallel workers for independent tasks',
  context: { domain: 'performance' },
  confidence_score: 0.88
});

// 查询历史记录
const historyStmt = db.prepare('SELECT * FROM task_history ORDER BY id DESC LIMIT 5');
const history = historyStmt.all();
console.log(`\n  📜 Task History (${history.length} entries):`);
history.forEach(h => {
  console.log(`    - ${h.title}: ${h.status} (${h.duration_ms}ms)`);
});

// 查询经验总结
const expStmt = db.prepare('SELECT * FROM experience_summaries ORDER BY confidence_score DESC');
const experiences = expStmt.all();
console.log(`\n  💡 Experience Summaries (${experiences.length} entries):`);
experiences.forEach(e => {
  console.log(`    - [${e.summary_type}] ${e.pattern_description} (${(e.confidence_score * 100).toFixed(0)}% confidence)`);
});

console.log('\n───────────────────────────────────────────────────────────\n');

// 3. Memory Sync Demo
console.log('🔄 3. Worker ↔ Master Memory Sync');
console.log('───────────────────────────────────────────────────────────');

// Worker 创建快照
const snapshot = {
  entries: Array.from(workerMemory.memory).map(([key, entry]) => ({
    key,
    value: entry.value,
    category: entry.category,
    taskId: entry.taskId,
    timestamp: entry.timestamp
  })),
  workerId: workerMemory.workerId,
  snapshotTime: new Date().toISOString()
};

console.log(`  ✓ Worker created snapshot with ${snapshot.entries.length} entries`);
console.log(`  → Snapshot time: ${snapshot.snapshotTime}`);
console.log(`  → Worker ID: ${snapshot.workerId}`);

// 模拟同步到 Master
console.log('\n  📤 Syncing to Master...');
console.log(`  → POST /api/v1/memory/sync`);
console.log(`  → Body: { worker_id: "${snapshot.workerId}", memory_snapshot: {...} }`);
console.log(`  ✓ Memory synchronized successfully\n`);

// 4. Cleanup Demo
console.log('🧹 4. Memory Cleanup');
console.log('───────────────────────────────────────────────────────────');

// 添加一些过期条目
workerMemory.remember('expired-1', 'data1', { ttl: 1 });
workerMemory.remember('expired-2', 'data2', { ttl: 1 });

setTimeout(() => {
  console.log('  → Cleaning expired entries...');
  const before = workerMemory.memory.size;
  workerMemory.memory.forEach((entry, key) => {
    if (entry.expiresAt && Date.now() > entry.expiresAt) {
      workerMemory.memory.delete(key);
    }
  });
  const after = workerMemory.memory.size;
  console.log(`  ✓ Cleaned up ${before - after} expired entries`);
  console.log(`  → Remaining entries: ${after}\n`);

  // 总结
  console.log('╔════════════════════════════════════════════════════════════╗');
  console.log('║                    Demo Complete!                         ║');
  console.log('╚════════════════════════════════════════════════════════════╝');
  console.log('\n📚 Next Steps:');
  console.log('  1. Build the project: cd nodejs && pnpm run build');
  console.log('  2. Start Master: node nodejs/packages/master/dist/index.js');
  console.log('  3. Start Worker: node nodejs/packages/worker/dist/index.js');
  console.log('  4. Create tasks and watch the memory system in action!');
  console.log('\n📖 Learn more:');
  console.log('  docs/MEMORY_SYSTEM_GUIDE.md - Complete guide\n');

  db.close();
}, 1100);
