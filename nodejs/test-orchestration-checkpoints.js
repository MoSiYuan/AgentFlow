#!/usr/bin/env node

/**
 * AgentFlow Orchestration & Checkpoint Integration Test
 *
 * 测试以下功能:
 * 1. 任务关系和 DAG 编排
 * 2. 检查点机制
 * 3. 任务升级
 * 4. Git 锁机制
 * 5. Ralph 编排模式
 */

const { AgentFlowDatabase } = require('./packages/database/dist/index.js');
const { setTimeout: sleep } = require('timers/promises');

// Inline TaskOrchestrator for testing
class TaskOrchestrator {
  constructor(db) {
    this.db = db;
  }

  async createPlan(groupName, mode = 'dag') {
    const tasks = this.db.listTasks({
      group_name: groupName,
      status: 'pending',
    });

    if (tasks.length === 0) {
      return {
        mode,
        tasks: [],
        execution_order: [],
        total_tasks: 0,
        ready_tasks: 0,
        completed_tasks: 0,
      };
    }

    const nodes = [];
    for (const task of tasks) {
      const taskId = parseInt(task.id.replace(/\D/g, ''));
      const predecessors = this.db.getTaskPredecessors(taskId);
      const successors = this.db.getTaskSuccessors(taskId);

      nodes.push({
        task_id: taskId,
        title: task.title,
        status: task.status,
        priority: task.priority,
        dependencies: predecessors.map((p) => p.predecessor_id),
        dependents: successors.map((s) => s.successor_id),
        can_execute: this.db.canExecuteTask(taskId),
      });
    }

    const execution_order = this.calculateExecutionOrder(nodes, mode);

    const ready_tasks = nodes.filter((n) => n.can_execute).length;
    const completed_tasks = nodes.filter((n) => n.status === 'completed').length;

    return {
      mode,
      tasks: nodes,
      execution_order,
      total_tasks: tasks.length,
      ready_tasks,
      completed_tasks,
    };
  }

  calculateExecutionOrder(nodes, mode) {
    switch (mode) {
      case 'sequential':
        return nodes.map((n) => [n.task_id]);
      case 'parallel':
        return nodes.length > 0 ? [nodes.map((n) => n.task_id)] : [];
      case 'dag':
        return this.topologicalSort(nodes);
      case 'pipeline':
        return this.createPipelineStages(nodes);
      case 'conditional':
        return this.topologicalSort(nodes);
      default:
        return this.topologicalSort(nodes);
    }
  }

  topologicalSort(nodes) {
    const inDegree = new Map();
    const adjList = new Map();
    const allTasks = new Set();

    for (const node of nodes) {
      allTasks.add(node.task_id);
      inDegree.set(node.task_id, node.dependencies.length);
      adjList.set(node.task_id, node.dependents);
    }

    const result = [];
    const queue = [];

    for (const taskId of allTasks) {
      if ((inDegree.get(taskId) || 0) === 0) {
        queue.push(taskId);
      }
    }

    while (queue.length > 0) {
      const levelSize = queue.length;
      const currentLevel = [];

      for (let i = 0; i < levelSize; i++) {
        const current = queue.shift();
        currentLevel.push(current);

        const dependents = adjList.get(current) || [];
        for (const dependent of dependents) {
          const newDegree = (inDegree.get(dependent) || 0) - 1;
          inDegree.set(dependent, newDegree);

          if (newDegree === 0) {
            queue.push(dependent);
          }
        }
      }

      result.push(currentLevel);
    }

    return result;
  }

  createPipelineStages(nodes) {
    const depthMap = new Map();

    const getDepth = (taskId, visited = new Set()) => {
      if (visited.has(taskId)) return 0;
      visited.add(taskId);

      const node = nodes.find((n) => n.task_id === taskId);
      if (!node || node.dependencies.length === 0) return 0;

      const maxDepth = Math.max(...node.dependencies.map((depId) => getDepth(depId, visited)));
      return maxDepth + 1;
    };

    const stages = new Map();
    for (const node of nodes) {
      const depth = getDepth(node.task_id);
      if (!stages.has(depth)) {
        stages.set(depth, []);
      }
      stages.get(depth).push(node.task_id);
    }

    return Array.from(stages.values()).sort((a, b) => a[0] - b[0]);
  }

  createCheckpoint(data) {
    return this.db.createCheckpoint(data);
  }

  getLatestCheckpoint(taskId) {
    const checkpoint = this.db.getLatestCheckpoint(taskId);
    if (!checkpoint) return null;

    return {
      id: checkpoint.id,
      task_id: checkpoint.task_id,
      worker_id: checkpoint.worker_id,
      checkpoint_name: checkpoint.checkpoint_name,
      data: JSON.parse(checkpoint.checkpoint_data),
      memory_snapshot: checkpoint.memory_snapshot ? JSON.parse(checkpoint.memory_snapshot) : undefined,
      state_snapshot: checkpoint.state_snapshot ? JSON.parse(checkpoint.state_snapshot) : undefined,
      timestamp: new Date(checkpoint.timestamp),
    };
  }

  upgradeTask(data) {
    return this.db.upgradeTask(data);
  }

  getUpgradeHistory(taskId) {
    return this.db.getTaskVersions(taskId);
  }
}

// Color codes
const colors = {
  reset: '\x1b[0m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
};

function log(color, message) {
  console.log(`${color}${message}${colors.reset}`);
}

// Test database
const dbPath = '/tmp/agentflow-test-orchestration.db';

function cleanDatabase() {
  const fs = require('fs');
  if (fs.existsSync(dbPath)) {
    fs.unlinkSync(dbPath);
  }
}

// Test 1: Task Relationships and DAG Orchestration
async function testTaskRelationships() {
  log(colors.cyan, '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.cyan, '📊 Test 1: Task Relationships and DAG Orchestration');
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  cleanDatabase();
  const db = new AgentFlowDatabase(dbPath);
  db.init();

  // Create test tasks with dependencies
  const task1 = db.createTask({
    title: 'Task 1 - Install Dependencies',
    description: 'npm install',
    group_name: 'test',
  });

  const task2 = db.createTask({
    title: 'Task 2 - Build Project',
    description: 'npm run build',
    group_name: 'test',
  });

  const task3 = db.createTask({
    title: 'Task 3 - Run Tests',
    description: 'npm test',
    group_name: 'test',
  });

  const task4 = db.createTask({
    title: 'Task 4 - Deploy',
    description: 'Deploy to production',
    group_name: 'test',
  });

  // Extract task IDs
  const id1 = parseInt(task1.replace(/\D/g, ''));
  const id2 = parseInt(task2.replace(/\D/g, ''));
  const id3 = parseInt(task3.replace(/\D/g, ''));
  const id4 = parseInt(task4.replace(/\D/g, ''));

  log(colors.blue, '✓ Created 4 tasks');

  // Create relationships: task1 -> task2 -> task4, task1 -> task3 -> task4
  db.addTaskRelationship({
    predecessor_id: id1,
    successor_id: id2,
    relationship_type: 'dependency',
  });

  db.addTaskRelationship({
    predecessor_id: id2,
    successor_id: id4,
    relationship_type: 'dependency',
  });

  db.addTaskRelationship({
    predecessor_id: id1,
    successor_id: id3,
    relationship_type: 'dependency',
  });

  db.addTaskRelationship({
    predecessor_id: id3,
    successor_id: id4,
    relationship_type: 'dependency',
  });

  log(colors.blue, '✓ Created task relationships');

  // Test DAG orchestration
  const orchestrator = new TaskOrchestrator(db);
  const plan = await orchestrator.createPlan('test', 'dag');

  log(colors.green, `\n✓ Orchestration Plan Created:`);
  log(colors.cyan, `  Mode: ${plan.mode}`);
  log(colors.cyan, `  Total tasks: ${plan.total_tasks}`);
  log(colors.cyan, `  Ready tasks: ${plan.ready_tasks}`);
  log(colors.cyan, `  Execution order:`);

  for (let i = 0; i < plan.execution_order.length; i++) {
    const level = plan.execution_order[i];
    const taskNames = level.map((id) => {
      const task = plan.tasks.find((t) => t.task_id === id);
      return task?.title || `Task ${id}`;
    });
    log(colors.yellow, `    Level ${i + 1}: ${taskNames.join(', ')}`);
  }

  // Verify execution order
  const expected = [[id1], [id2, id3], [id4]];
  const correct =
    plan.execution_order.length === expected.length &&
    plan.execution_order[0][0] === id1 &&
    plan.execution_order[1].length === 2 &&
    plan.execution_order[2][0] === id4;

  if (correct) {
    log(colors.green, '\n✅ Test 1 PASSED: DAG orchestration works correctly\n');
    db.close();
    return true;
  } else {
    log(colors.red, '\n❌ Test 1 FAILED: DAG orchestration incorrect\n');
    db.close();
    return false;
  }
}

// Test 2: Checkpoint Mechanism
async function testCheckpoints() {
  log(colors.cyan, '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.cyan, '📊 Test 2: Checkpoint Mechanism');
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  cleanDatabase();
  const db = new AgentFlowDatabase(dbPath);
  db.init();

  const orchestrator = new TaskOrchestrator(db);

  // Create a task
  const task = db.createTask({
    title: 'Test Checkpoint Task',
    description: 'This task will have checkpoints',
    group_name: 'test',
  });

  const taskId = parseInt(task.replace(/\D/g, ''));

  log(colors.blue, `✓ Created task: ${task}`);

  // Create checkpoints
  const checkpoint1 = orchestrator.createCheckpoint({
    task_id: taskId,
    worker_id: 'worker-test-1',
    checkpoint_name: 'task_start',
    checkpoint_data: { status: 'started', progress: 0 },
    memory_snapshot: { context: 'Initial context' },
    state_snapshot: { variables: {} },
  });

  log(colors.green, `✓ Created checkpoint 1: ID ${checkpoint1}`);

  await sleep(100);

  const checkpoint2 = orchestrator.createCheckpoint({
    task_id: taskId,
    worker_id: 'worker-test-1',
    checkpoint_name: 'task_progress',
    checkpoint_data: { status: 'running', progress: 50 },
    memory_snapshot: { context: 'Updated context' },
    state_snapshot: { variables: { temp: 'value' } },
  });

  log(colors.green, `✓ Created checkpoint 2: ID ${checkpoint2}`);

  // Retrieve latest checkpoint
  const latest = orchestrator.getLatestCheckpoint(taskId);

  if (latest && latest.checkpoint_name === 'task_progress') {
    log(colors.cyan, '\n✓ Retrieved latest checkpoint:');
    log(colors.yellow, `  Name: ${latest.checkpoint_name}`);
    log(colors.yellow, `  Data: ${JSON.stringify(latest.data)}`);
    log(colors.yellow, `  Memory: ${JSON.stringify(latest.memory_snapshot)}`);

    log(colors.green, '\n✅ Test 2 PASSED: Checkpoint mechanism works correctly\n');
    db.close();
    return true;
  } else {
    log(colors.red, '\n❌ Test 2 FAILED: Could not retrieve latest checkpoint\n');
    db.close();
    return false;
  }
}

// Test 3: Task Upgrade Mechanism
async function testTaskUpgrade() {
  log(colors.cyan, '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.cyan, '📊 Test 3: Task Upgrade Mechanism');
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  cleanDatabase();
  const db = new AgentFlowDatabase(dbPath);
  db.init();

  const orchestrator = new TaskOrchestrator(db);

  // Create a task
  const task = db.createTask({
    title: 'Original Task',
    description: 'This is the original task description',
    group_name: 'test',
  });

  const taskId = parseInt(task.replace(/\D/g, ''));

  log(colors.blue, `✓ Created task: ${task}`);

  // Upgrade task
  const upgraded = orchestrator.upgradeTask({
    task_id: taskId,
    new_title: 'Upgraded Task',
    new_description: 'This is the upgraded task description with more details',
    upgrade_reason: 'Requirements changed - need to add additional features',
  });

  if (!upgraded) {
    log(colors.red, '\n❌ Test 3 FAILED: Could not upgrade task\n');
    db.close();
    return false;
  }

  log(colors.green, '✓ Task upgraded successfully');

  // Check upgrade history
  const history = orchestrator.getUpgradeHistory(taskId);

  log(colors.cyan, `\n✓ Upgrade history (${history.length} versions):`);
  for (const version of history) {
    log(colors.yellow, `  Version ${version.version_number}: ${version.title}`);
    log(colors.yellow, `    Reason: ${version.upgrade_reason}`);
  }

  // Get latest task
  const currentTask = db.getTask(task);

  if (currentTask && currentTask.title === 'Upgraded Task') {
    log(colors.cyan, '\n✓ Current task title: ' + currentTask.title);
    log(colors.cyan, '✓ Current task description: ' + currentTask.description);

    log(colors.green, '\n✅ Test 3 PASSED: Task upgrade mechanism works correctly\n');
    db.close();
    return true;
  } else {
    log(colors.red, '\n❌ Test 3 FAILED: Task was not upgraded correctly\n');
    db.close();
    return false;
  }
}

// Test 4: Git Lock Mechanism
async function testGitLocks() {
  log(colors.cyan, '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.cyan, '📊 Test 4: Git Lock Mechanism');
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  cleanDatabase();
  const db = new AgentFlowDatabase(dbPath);
  db.init();

  // Test file lock
  const filePath1 = '/src/app.ts';
  const locked1 = db.acquireGitLock({
    task_id: 'git-task-1',
    agent_id: 'agent-1',
    file_path: filePath1,
    lock_type: 'write',
  });

  log(colors.green, `✓ Acquired lock for ${filePath1}: ${locked1 ? 'Success' : 'Failed'}`);

  // Check if file is locked
  const isLocked = db.isFileLocked(filePath1);
  log(colors.cyan, `✓ File ${filePath1} is locked: ${isLocked}`);

  // Try to acquire lock again (should fail)
  const locked2 = db.acquireGitLock({
    task_id: 'git-task-2',
    agent_id: 'agent-2',
    file_path: filePath1,
    lock_type: 'write',
  });

  log(colors.yellow, `✓ Second lock attempt for ${filePath1}: ${locked2 ? 'Success (unexpected!)' : 'Failed (expected)'}`);

  // Release lock
  const released = db.releaseGitLock('git-task-1', filePath1);
  log(colors.green, `✓ Released lock for ${filePath1}: ${released ? 'Success' : 'Failed'}`);

  // Check again
  const stillLocked = db.isFileLocked(filePath1);
  log(colors.cyan, `✓ File ${filePath1} is still locked: ${stillLocked}`);

  // Test multiple locks
  const filePath2 = '/src/utils.ts';
  db.acquireGitLock({
    task_id: 'git-task-3',
    agent_id: 'agent-1',
    file_path: filePath2,
    lock_type: 'read',
  });

  db.acquireGitLock({
    task_id: 'git-task-3',
    agent_id: 'agent-1',
    file_path: '/src/components.ts',
    lock_type: 'write',
  });

  const activeLocks = db.getActiveLocks('git-task-3');
  log(colors.cyan, `\n✓ Active locks for git-task-3: ${activeLocks.length} files`);

  // Test cleanup
  const cleaned = db.releaseAllLocks('git-task-3');
  log(colors.green, `✓ Released all locks for git-task-3: ${cleaned} locks\n`);

  if (isLocked && !stillLocked && activeLocks.length === 2 && cleaned >= 2) {
    log(colors.green, '✅ Test 4 PASSED: Git lock mechanism works correctly\n');
    db.close();
    return true;
  } else {
    log(colors.red, '❌ Test 4 FAILED: Git lock mechanism not working correctly\n');
    db.close();
    return false;
  }
}

// Test 5: Ralph Orchestration Modes
async function testRalphModes() {
  log(colors.cyan, '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.cyan, '📊 Test 5: Ralph Orchestration Modes');
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  cleanDatabase();
  const db = new AgentFlowDatabase(dbPath);
  db.init();
  const orchestrator = new TaskOrchestrator(db);

  // Create test tasks
  for (let i = 1; i <= 5; i++) {
    db.createTask({
      title: `Task ${i}`,
      description: `Description for task ${i}`,
      group_name: 'test',
    });
  }

  log(colors.blue, '✓ Created 5 tasks\n');

  const modes = ['sequential', 'parallel', 'dag', 'pipeline'];
  const results = [];

  for (const mode of modes) {
    const plan = await orchestrator.createPlan('test', mode);

    log(colors.yellow, `Mode: ${mode}`);
    log(colors.cyan, `  Execution order levels: ${plan.execution_order.length}`);
    log(colors.cyan, `  Tasks per level: ${plan.execution_order.map((l) => l.length).join(', ')}`);

    results.push({
      mode,
      levels: plan.execution_order.length,
      total_tasks: plan.total_tasks,
    });
  }

  // Verify results
  const sequential = results.find((r) => r.mode === 'sequential');
  const parallel = results.find((r) => r.mode === 'parallel');
  const dag = results.find((r) => r.mode === 'dag');

  const correct =
    sequential?.levels === 5 && parallel?.levels === 1 && dag?.total_tasks === 5;

  if (correct) {
    log(colors.green, '\n✅ Test 5 PASSED: Ralph orchestration modes work correctly\n');
    db.close();
    return true;
  } else {
    log(colors.red, '\n❌ Test 5 FAILED: Ralph modes not working correctly\n');
    db.close();
    return false;
  }
}

// Main test runner
async function runTests() {
  log(colors.magenta, '\n╔═══════════════════════════════════════════════════════════════╗');
  log(colors.magenta, '║        AgentFlow Orchestration & Checkpoint Test             ║');
  log(colors.magenta, '║              验证高级功能和编排能力                          ║');
  log(colors.magenta, '╚═══════════════════════════════════════════════════════════════╝\n');

  const results = [];

  // Test 1: Task Relationships and DAG
  log(colors.yellow, '阶段 1/5: Task Relationships and DAG...\n');
  const test1 = await testTaskRelationships();
  results.push({ name: 'Task Relationships & DAG', passed: test1 });
  await sleep(1000);

  // Test 2: Checkpoints
  log(colors.yellow, '阶段 2/5: Checkpoint Mechanism...\n');
  const test2 = await testCheckpoints();
  results.push({ name: 'Checkpoint Mechanism', passed: test2 });
  await sleep(1000);

  // Test 3: Task Upgrade
  log(colors.yellow, '阶段 3/5: Task Upgrade...\n');
  const test3 = await testTaskUpgrade();
  results.push({ name: 'Task Upgrade', passed: test3 });
  await sleep(1000);

  // Test 4: Git Locks
  log(colors.yellow, '阶段 4/5: Git Locks...\n');
  const test4 = await testGitLocks();
  results.push({ name: 'Git Locks', passed: test4 });
  await sleep(1000);

  // Test 5: Ralph Modes
  log(colors.yellow, '阶段 5/5: Ralph Orchestration Modes...\n');
  const test5 = await testRalphModes();
  results.push({ name: 'Ralph Modes', passed: test5 });

  // Summary
  console.log('\n' + '═'.repeat(65));
  log(colors.cyan, '📊 测试总结');
  console.log('═'.repeat(65));

  let passed = 0;
  for (const result of results) {
    const status = result.passed ? '✅ PASS' : '❌ FAIL';
    const color = result.passed ? colors.green : colors.red;
    log(color, `  ${status} - ${result.name}`);
    if (result.passed) passed++;
  }

  console.log('═'.repeat(65));
  log(colors.cyan, `总计: ${passed}/${results.length} 测试通过`);
  console.log('═'.repeat(65));

  if (passed === results.length) {
    log(colors.green, '\n🎉 所有高级功能测试通过！AgentFlow 具备完整的任务编排和检查点能力！\n');

    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    log(colors.cyan, '📋 已验证的功能:');
    log(colors.blue, '  1. ✅ Task Relationships  → 任务依赖关系管理');
    log(colors.blue, '  2. ✅ DAG Orchestration   → DAG 编排执行');
    log(colors.blue, '  3. ✅ Checkpoint System   → 检查点和恢复机制');
    log(colors.blue, '  4. ✅ Task Upgrade        → 任务升级和版本控制');
    log(colors.blue, '  5. ✅ Git Locks           → Git 文件锁防止冲突');
    log(colors.blue, '  6. ✅ Ralph Modes         → 多种编排模式');
    log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

    return true;
  } else {
    log(colors.yellow, `\n⚠️  ${results.length - passed} 个测试失败\n`);
    return false;
  }
}

// Run tests
runTests()
  .then((success) => {
    process.exit(success ? 0 : 1);
  })
  .catch((error) => {
    log(colors.red, `\n❌ 测试失败: ${error.message}\n`);
    console.error(error);
    process.exit(1);
  });
