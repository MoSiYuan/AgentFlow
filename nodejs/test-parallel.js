#!/usr/bin/env node

/**
 * AgentFlow Parallel Task Test
 *
 * This script tests the Master-Worker system by:
 * 1. Starting a Master server
 * 2. Creating multiple test tasks
 * 3. Spawning multiple workers to execute tasks in parallel
 * 4. Monitoring task execution
 * 5. Verifying all tasks complete successfully
 */

const { spawn } = require('child_process');
const { setTimeout: sleep } = require('timers/promises');
const http = require('http');

// Configuration
const MASTER_PORT = 8848;
const NUM_WORKERS = 3;
const NUM_TASKS = 6;
const TEST_DB = '.claude/cpds-manager/test-agentflow.db';

// Color codes for output
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

// Helper function to make HTTP requests
function httpRequest(method, path, data = null) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'localhost',
      port: MASTER_PORT,
      path: path,
      method: method,
      headers: {
        'Content-Type': 'application/json',
      },
    };

    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => body += chunk);
      res.on('end', () => {
        try {
          resolve({ status: res.statusCode, data: JSON.parse(body) });
        } catch (e) {
          resolve({ status: res.statusCode, data: body });
        }
      });
    });

    req.on('error', reject);
    if (data) {
      req.write(JSON.stringify(data));
    }
    req.end();
  });
}

// Start Master server
function startMaster() {
  log(colors.cyan, '\n🚀 Starting Master server...');
  return spawn('node', ['packages/master/dist/index.js'], {
    stdio: ['ignore', 'pipe', 'pipe'],
    env: { ...process.env, NODE_ENV: 'test' }
  });
}

// Start a Worker
function startWorker(workerId) {
  log(colors.blue, `👷 Starting Worker ${workerId}...`);
  return spawn('node', ['packages/worker/dist/index.js'], {
    stdio: ['ignore', 'pipe', 'pipe'],
    env: { ...process.env }
  });
}

// Wait for server to be ready
async function waitForServer() {
  log(colors.yellow, '⏳ Waiting for Master server to start...');
  for (let i = 0; i < 30; i++) {
    try {
      const response = await httpRequest('GET', '/health');
      if (response.status === 200) {
        log(colors.green, '✓ Master server is ready!');
        return true;
      }
    } catch (e) {
      await sleep(500);
    }
  }
  throw new Error('Master server failed to start');
}

// Create test tasks
async function createTasks() {
  log(colors.magenta, `\n📝 Creating ${NUM_TASKS} test tasks...`);

  const taskTypes = [
    { title: 'Echo Test', description: 'echo "Hello from Worker"' },
    { title: 'Date Check', description: 'date' },
    { title: 'List Files', description: 'ls -la' },
    { title: 'Node Version', description: 'node --version' },
    { title: 'CPU Info', description: 'uname -m' },
    { title: 'Memory Check', description: 'free -h || vm_stat' },
  ];

  const taskIds = [];
  for (let i = 0; i < NUM_TASKS; i++) {
    const taskType = taskTypes[i % taskTypes.length];
    const response = await httpRequest('POST', '/api/v1/tasks', {
      title: `Task ${i + 1}: ${taskType.title}`,
      description: taskType.description,
      group_name: 'test',
      priority: i % 3 === 0 ? 'high' : 'medium'
    });

    if (response.status === 201 && response.data.success) {
      taskIds.push(response.data.data.task_id);
      log(colors.green, `  ✓ Created task: ${response.data.data.task_id}`);
    } else {
      log(colors.red, `  ✗ Failed to create task ${i + 1}`);
    }
  }

  return taskIds;
}

// Monitor task completion
async function monitorTasks(taskIds) {
  log(colors.yellow, '\n⏳ Monitoring task execution...');

  const startTime = Date.now();
  let completed = 0;
  let failed = 0;

  while (completed + failed < taskIds.length) {
    await sleep(1000);

    const response = await httpRequest('GET', '/api/v1/tasks');
    if (response.status === 200 && response.data.success) {
      const tasks = response.data.data.tasks;

      for (const task of tasks) {
        if (taskIds.includes(task.id)) {
          if (task.status === 'completed' && !task.checked) {
            completed++;
            task.checked = true;
            log(colors.green, `  ✓ ${task.id} completed: ${task.title}`);
          } else if (task.status === 'failed' && !task.checked) {
            failed++;
            task.checked = true;
            log(colors.red, `  ✗ ${task.id} failed: ${task.error || 'Unknown error'}`);
          }
        }
      }
    }

    // Show progress
    const total = completed + failed;
    if (total > 0) {
      process.stdout.write(`\r  Progress: ${total}/${taskIds.length} tasks processed (${completed} completed, ${failed} failed)`);
    }
  }

  const elapsed = ((Date.now() - startTime) / 1000).toFixed(2);
  console.log(); // New line after progress
  log(colors.cyan, `\n⏱️  All tasks completed in ${elapsed} seconds`);

  return { completed, failed, elapsed };
}

// Main test function
async function runTest() {
  let master;
  const workers = [];

  try {
    // Start Master
    master = startMaster();
    master.stdout.on('data', (data) => {
      process.stdout.write(`[Master] ${data}`);
    });
    master.stderr.on('data', (data) => {
      process.stderr.write(`[Master Error] ${data}`);
    });

    // Wait for Master to be ready
    await waitForServer();

    // Create tasks
    const taskIds = await createTasks();

    // Wait a bit for tasks to be stored
    await sleep(1000);

    // Start workers
    log(colors.cyan, `\n👷 Starting ${NUM_WORKERS} workers...`);
    for (let i = 0; i < NUM_WORKERS; i++) {
      const worker = startWorker(i + 1);
      workers.push(worker);

      worker.stdout.on('data', (data) => {
        process.stdout.write(`[Worker ${i + 1}] ${data}`);
      });
      worker.stderr.on('data', (data) => {
        process.stderr.write(`[Worker ${i + 1} Error] ${data}`);
      });

      await sleep(500); // Stagger worker starts
    }

    // Monitor task completion
    const results = await monitorTasks(taskIds);

    // Summary
    console.log('\n' + '='.repeat(60));
    log(colors.cyan, '📊 Test Results:');
    log(colors.green, `  ✓ Completed: ${results.completed}/${taskIds.length}`);
    if (results.failed > 0) {
      log(colors.red, `  ✗ Failed: ${results.failed}/${taskIds.length}`);
    }
    log(colors.blue, `  ⏱️  Time: ${results.elapsed}s`);
    log(colors.blue, `  👷 Workers: ${NUM_WORKERS}`);
    log(colors.blue, `  📝 Tasks: ${taskIds.length}`);
    log(colors.blue, `  📈 Throughput: ${(taskIds.length / results.elapsed).toFixed(2)} tasks/sec`);
    console.log('='.repeat(60));

    if (results.failed === 0) {
      log(colors.green, '\n🎉 All tests passed!');
    } else {
      log(colors.red, '\n❌ Some tests failed!');
    }

  } catch (error) {
    log(colors.red, `\n❌ Test failed: ${error.message}`);
    console.error(error);
  } finally {
    // Cleanup
    log(colors.yellow, '\n🧹 Cleaning up...');

    // Stop workers
    for (const worker of workers) {
      worker.kill('SIGTERM');
    }

    // Stop master
    if (master) {
      master.kill('SIGTERM');
    }

    await sleep(1000);
    log(colors.green, '✓ Cleanup complete');
  }
}

// Run the test
log(colors.magenta, '╔════════════════════════════════════════════════════════╗');
log(colors.magenta, '║     AgentFlow Parallel Task Execution Test             ║');
log(colors.magenta, '╚════════════════════════════════════════════════════════╝');

runTest().then(() => {
  process.exit(0);
}).catch((error) => {
  log(colors.red, `Fatal error: ${error.message}`);
  process.exit(1);
});
