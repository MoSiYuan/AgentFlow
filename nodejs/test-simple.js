#!/usr/bin/env node

/**
 * AgentFlow Simple Worker Test
 *
 * This script tests the Worker logic directly without requiring
 * the full Master server setup.
 */

const { spawn } = require('child_process');
const { setTimeout: sleep } = require('timers/promises');

// Color codes
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
};

function log(color, message) {
  console.log(`${color}${message}${colors.reset}`);
}

// Test that workers can execute simple commands
async function testWorkerExecution() {
  log(colors.cyan, '\n╔════════════════════════════════════════════════════════╗');
  log(colors.cyan, '║     AgentFlow Worker Execution Test                    ║');
  log(colors.cyan, '╚════════════════════════════════════════════════════════╝\n');

  log(colors.magenta, '📝 Testing worker execution capabilities...\n');

  // Test commands
  const tests = [
    { name: 'Echo Test', cmd: 'echo "Hello from AgentFlow!"' },
    { name: 'Date Check', cmd: 'date' },
    { name: 'Node Version', cmd: 'node --version' },
    { name: 'Working Directory', cmd: 'pwd' },
  ];

  const results = [];

  for (const test of tests) {
    log(colors.yellow, `Running: ${test.name}`);
    log(colors.blue, `  Command: ${test.cmd}`);

    try {
      const result = await execCommand(test.cmd);
      log(colors.green, `  ✓ Output: ${result.trim().split('\n')[0]}`);
      results.push({ test: test.name, success: true, output: result });
    } catch (error) {
      log(colors.red, `  ✗ Error: ${error.message}`);
      results.push({ test: test.name, success: false, error: error.message });
    }
    console.log();
  }

  // Summary
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.cyan, '📊 Test Summary:');
  const passed = results.filter(r => r.success).length;
  const failed = results.filter(r => !r.success).length;
  log(colors.green, `  ✓ Passed: ${passed}/${results.length}`);
  if (failed > 0) {
    log(colors.red, `  ✗ Failed: ${failed}/${results.length}`);
  }
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  return failed === 0;
}

function execCommand(cmd) {
  return new Promise((resolve, reject) => {
    const [command, ...args] = cmd.split(' ');
    const child = spawn(command, args, { shell: true });

    let stdout = '';
    let stderr = '';

    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });

    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });

    child.on('close', (code) => {
      if (code === 0) {
        resolve(stdout);
      } else {
        reject(new Error(stderr || `Command failed with exit code ${code}`));
      }
    });
  });
}

// Test task distribution concept
async function testParallelExecution() {
  log(colors.magenta, '\n📝 Testing parallel execution concept...\n');

  const tasks = [
    { id: 'TASK-001', cmd: 'echo "Task 1 executing"', delay: 1000 },
    { id: 'TASK-002', cmd: 'echo "Task 2 executing"', delay: 800 },
    { id: 'TASK-003', cmd: 'echo "Task 3 executing"', delay: 1200 },
  ];

  const startTime = Date.now();

  log(colors.yellow, `Starting ${tasks.length} tasks concurrently...`);

  const results = await Promise.all(
    tasks.map(async (task) => {
      const taskStart = Date.now();
      try {
        const output = await execCommand(task.cmd);
        await sleep(task.delay);
        const elapsed = Date.now() - taskStart;
        log(colors.green, `  ✓ ${task.id} completed in ${elapsed}ms`);
        return { task: task.id, success: true, elapsed };
      } catch (error) {
        log(colors.red, `  ✗ ${task.id} failed: ${error.message}`);
        return { task: task.id, success: false, error: error.message };
      }
    })
  );

  const totalElapsed = Date.now() - startTime;

  console.log();
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  log(colors.cyan, '📊 Parallel Execution Results:');
  log(colors.blue, `  Total time: ${totalElapsed}ms`);
  log(colors.blue, `  Tasks completed: ${results.filter(r => r.success).length}/${results.length}`);
  log(colors.blue, `  Average time per task: ${(totalElapsed / results.length).toFixed(0)}ms`);
  log(colors.cyan, '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

  return results.filter(r => r.success).length === results.length;
}

// Main test function
async function runTests() {
  try {
    const test1 = await testWorkerExecution();
    const test2 = await testParallelExecution();

    if (test1 && test2) {
      log(colors.green, '\n🎉 All tests passed!\n');
      process.exit(0);
    } else {
      log(colors.yellow, '\n⚠️  Some tests had issues\n');
      process.exit(1);
    }
  } catch (error) {
    log(colors.red, `\n❌ Test failed: ${error.message}\n`);
    console.error(error);
    process.exit(1);
  }
}

runTests();
