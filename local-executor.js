#!/usr/bin/env node
/**
 * AgentFlow 简化本地执行器
 *
 * 自动管理 Master 和 Worker 执行任务
 * 用法: agentflow run "task1" "task2"
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// 加载配置
function loadConfig() {
  const configPath = path.join(process.cwd(), '.agentflow', 'config.json');
  if (fs.existsSync(configPath)) {
    return JSON.parse(fs.readFileSync(configPath, 'utf8'));
  }
  // 默认配置
  return {
    master: {
      path: path.join(__dirname, 'nodejs/packages/master/dist/index.js'),
      url: 'http://localhost:6767'
    },
    worker: {
      group: 'local'
    }
  };
}

async function executeLocal(tasks) {
  const config = loadConfig();

  let masterProcess = null;
  let workerProcess = null;

  const log = {
    info: (msg) => console.log(`  ${msg}`),
    error: (msg) => console.error(`  ✗ ${msg}`),
    success: (msg) => console.log(`  ✓ ${msg}`)
  };

  const masterURL = config.master.url;
  const groupName = config.worker.group;

  // 健康检查函数
  async function checkHealth(url, retries = 10) {
    while (retries > 0) {
      try {
        const response = await fetch(`${url}/health`);
        return response.ok;
      } catch {
        retries--;
        await new Promise(r => setTimeout(r, 500));
      }
    }
    return false;
  }

  // 创建任务函数
  async function createTask(title, description, groupName) {
    const response = await fetch(`${masterURL}/api/v1/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title,
        description,
        priority: 50,
        group_name: groupName
      })
    });

    if (!response.ok) {
      throw new Error(`Failed to create task: ${response.statusText}`);
    }

    const data = await response.json();
    return data.data.task_id;
  }

  // 获取任务状态函数
  async function getTaskStatus(taskId) {
    try {
      const response = await fetch(`${masterURL}/api/v1/tasks/${taskId}`);
      if (response.ok) {
        const data = await response.json();
        return data.data;
      }
    } catch {}
    return { status: 'pending' };
  }

  try {
    // 启动 Master
    log.info('正在启动服务...');

    // 使用 pnpm exec 来解决模块路径问题
    const masterDir = path.dirname(config.master.path);
    const masterBin = path.join(path.dirname(masterDir), 'node_modules', '.bin', 'master');

    // 检查是否有 master 命令
    let usePnpm = false;
    try {
      if (require('fs').existsSync(path.join(masterDir, '../../node_modules/.pnpm'))) {
        usePnpm = true;
      }
    } catch {}

    masterProcess = spawn(usePnpm ? 'pnpm' : 'node', [
      usePnpm ? '--filter=@agentflow/master' : 'exec',
      usePnpm ? 'start' : path.basename(config.master.path)
    ].filter(Boolean), {
      stdio: 'pipe',
      env: { ...process.env, NODE_ENV: 'development' },
      cwd: usePnpm ? path.join(__dirname, 'nodejs') : masterDir,
      shell: usePnpm
    });

    let masterStarted = false;
    let masterError = false;

    masterProcess.on('error', (err) => {
      if (!masterStarted) {
        log.error(`Master 启动失败: ${err.message}`);
        masterError = true;
      }
    });

    // 收集输出但不显示（兼容性问题）
    masterProcess.stderr.on('data', (data) => {
      if (data.toString().includes('Cannot find module')) {
        // 忽略 pnpm 模块解析警告
      }
    });

    // 等待 Master 就绪
    await new Promise(resolve => setTimeout(resolve, 3000));

    const isHealthy = await checkHealth(masterURL);
    if (!isHealthy) {
      log.error('Master 启动失败');
      process.exit(1);
    }
    log.success('服务已就绪');

    // 启动 Worker
    log.info('正在注册工作节点...');
    const workerCode = `
      const { Worker } = require('@agentflow/worker');
      const worker = new Worker({
        master_url: '${masterURL}',
        group_name: '${groupName}'
      });
      worker.start();
    `;

    workerProcess = spawn('node', ['-e', workerCode], {
      stdio: 'pipe',
      env: { ...process.env },
      cwd: path.join(__dirname, 'nodejs')
    });

    workerProcess.on('error', (err) => {
      log.error(`Worker 启动失败: ${err.message}`);
    });

    await new Promise(resolve => setTimeout(resolve, 2000));
    log.success('工作节点已就绪');

    // 创建任务
    log.info('正在创建任务...');
    const taskIds = [];

    for (let i = 0; i < tasks.length; i++) {
      try {
        const taskId = await createTask(`任务 ${i + 1}`, tasks[i], groupName);
        taskIds.push(taskId);
        log.success(`${tasks[i]}`);
      } catch (err) {
        log.error(`创建任务失败: ${tasks[i]}`);
      }
    }

    console.log();
    log.info(`共创建 ${taskIds.length} 个任务`);

    if (taskIds.length === 0) {
      log.error('没有任务被创建');
      process.exit(1);
    }

    // 监控执行
    log.info('正在执行任务...');
    const startTime = Date.now();
    const maxWait = 300000; // 5分钟

    while (Date.now() - startTime < maxWait) {
      const statuses = await Promise.all(
        taskIds.map(async (id) => await getTaskStatus(id))
      );

      const completed = statuses.filter(t => t.status === 'completed').length;
      const failed = statuses.filter(t => t.status === 'failed').length;
      const total = taskIds.length;

      process.stdout.write(`\r  进度: ${completed}/${total} 完成, ${failed} 失败`);

      if (completed + failed >= total) {
        console.log();
        log.success('所有任务已完成');
        break;
      }

      await new Promise(resolve => setTimeout(resolve, 3000));
    }

    // 关闭服务
    console.log();
    log.info('正在关闭服务...');

    workerProcess.kill('SIGTERM');
    await new Promise(resolve => setTimeout(resolve, 1000));

    try {
      await fetch(`${masterURL}/api/v1/shutdown`, { method: 'POST' });
      await new Promise(resolve => setTimeout(resolve, 2000));
    } catch {}

    masterProcess.kill('SIGTERM');
    await new Promise(resolve => setTimeout(resolve, 1000));

    try { masterProcess.kill('SIGKILL'); } catch {}
    try { workerProcess.kill('SIGKILL'); } catch {}

    log.success('执行完成\n');

  } catch (error) {
    log.error(error.message);

    // 清理
    try {
      if (masterProcess) masterProcess.kill();
      if (workerProcess) workerProcess.kill();
    } catch {}

    process.exit(1);
  }
}

// 解析命令行参数
let tasks = [];
const args = process.argv.slice(2);

// 支持两种格式:
// 1. JSON 数组: node local-executor.js '["task1","task2"]'
// 2. 多个参数: node local-executor.js "task1" "task2"
if (args.length === 1 && args[0].startsWith('[')) {
  try {
    tasks = JSON.parse(args[0]);
  } catch (e) {
    console.error('错误: JSON 格式不正确');
    console.error('示例: node local-executor.js \'["npm test","npm run build"]\'');
    process.exit(1);
  }
} else if (args.length > 0) {
  tasks = args;
}

if (tasks.length === 0) {
  console.log('用法: node local-executor.js <任务1> <任务2> ...');
  console.log('或: node local-executor.js \'["任务1","任务2"]\'');
  console.log('');
  console.log('示例:');
  console.log('  node local-executor.js "npm test" "npm run build"');
  console.log('  node local-executor.js \'["npm test","npm run build"]\'');
  console.log('');
  console.log('或使用 CLI:');
  console.log('  agentflow run ["npm test","npm run build"]');
  process.exit(1);
}

// 执行
console.log('AgentFlow 本地执行\n');
executeLocal(tasks).catch(err => {
  console.error('致命错误:', err);
  process.exit(1);
});
