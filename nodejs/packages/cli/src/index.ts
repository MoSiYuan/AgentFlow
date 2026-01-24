/**
 * AgentFlow CLI Tool (Node.js)
 *
 * Command-line interface for AgentFlow
 */

import { Command } from 'commander';
import { program } from 'commander';
import chalk from 'chalk';
import { Master } from '@agentflow/master';
import { Worker } from '@agentflow/worker';
import { LocalExecutor } from '@agentflow/local-executor';
import type { MasterConfig, WorkerConfig } from '@agentflow/shared';
import * as path from 'path';
import * as os from 'os';

// CLI version
const VERSION = '1.0.0';

program.version(VERSION);

// Master command
const masterCommand = new Command('master')
  .description('Start AgentFlow Master server')
  .option('--host <host>', 'Host to bind to', '0.0.0.0')
  .option('--port <port>', 'Port to listen on', '6767')
  .option('--db <db>', 'Database file path', '.claude/cpds-manager/agentflow.db')
  .option('--auto-shutdown', 'Enable auto-shutdown mode', false)
  .option('--ws-enabled', 'Enable WebSocket server', true)
  .option('--ws-port <port>', 'WebSocket port', '8849')
  .action(async (options) => {
    const config: MasterConfig = {
      host: options.host,
      port: parseInt(options.port),
      db_path: options.db,
      auto_shutdown: options.autoShutdown,
      log_level: 'info',
      git_enabled: true,
      git_boundary_config: '.agentflow/boundaries.json',
      websocket_enabled: options.wsEnabled,
      websocket_port: parseInt(options.wsPort)
    };

    try {
      const master = new Master(config);
      await master.start();

      // Handle graceful shutdown
      process.on('SIGINT', async () => {
        console.log('\n⏹  Shutting down gracefully...');
        await master.stop();
        process.exit(0);
      });

      process.on('SIGTERM', async () => {
        console.log('\n⏹  Shutting down gracefully...');
        await master.stop();
        process.exit(0);
      });

    } catch (error) {
      console.error(chalk.red('Failed to start master:'), error);
      process.exit(1);
    }
  });

// Run command - execute tasks with automatic Master/Worker management
const runCommand = new Command('run')
  .description('Execute tasks (automatically manages Master and Worker)')
  .argument('[command]', 'Command to execute')
  .option('--title <title>', 'Task title')
  .option('--master-host <host>', 'Master host', 'localhost')
  .option('--master-port <port>', 'Master port', '6767')
  .option('--db <path>', 'Database path', path.join(os.tmpdir(), 'agentflow-cli.db'))
  .option('--no-shutdown', 'Keep Master and Worker running after completion')
  .option('--group <group>', 'Worker group name', 'cli')
  .action(async (command, options) => {
    if (!command) {
      console.error(chalk.red('Error: Command is required'));
      console.log(chalk.yellow('Example: agentflow run "echo Hello World"'));
      process.exit(1);
    }

    const executor = new LocalExecutor({
      masterPath: path.join(__dirname, '../../master/dist/index.js'),
      masterHost: options.masterHost,
      masterPort: parseInt(options.masterPort),
      dbPath: options.db,
      shutdownOnComplete: options.shutdown
    });

    try {
      console.log(chalk.cyan(`\n🚀 Executing: ${command}\n`));

      await executor.executeOne(
        options.title || command.split(' ')[0],
        command
      );

      console.log(chalk.green('\n✅ Execution complete!\n'));
    } catch (error) {
      console.error(chalk.red('\n❌ Execution failed:'), error);
      process.exit(1);
    }
  });

// Worker command
const workerCommand = new Command('worker')
  .description('Start AgentFlow Worker')
  .option('--master <url>', 'Master server URL', 'http://localhost:6767')
  .option('--id <id>', 'Worker ID', '')
  .option('--group <group>', 'Worker group name', 'default')
  .option('--mode <mode>', 'Worker mode', 'auto')
  .action(async (options) => {
    const config: WorkerConfig = {
      master_url: options.master,
      id: options.id,
      group_name: options.group,
      mode: options.mode as 'auto' | 'manual' | 'oneshot',
      heartbeat_interval: 30000,
      heartbeat_timeout: 120000,
      max_concurrent: 1,
      task_timeout: 300000,
      retry_on_failure: true,
      max_retries: 3
    };

    try {
      const worker = new Worker(config);

      if (config.mode === 'oneshot') {
        const result = await worker.runOneShot();
        console.log(chalk.green('One-shot execution result:'));
        console.log(result);
        process.exit(0);
      } else {
        await worker.start();

        // Handle graceful shutdown
        process.on('SIGINT', async () => {
          console.log('\n⏹  Stopping worker...');
          await worker.stop();
          process.exit(0);
        });

        process.on('SIGTERM', async () => {
          console.log('\n⏹  Stopping worker...');
          await worker.stop();
          process.exit(0);
        });
      }

    } catch (error) {
      console.error(chalk.red('Failed to start worker:'), error);
      process.exit(1);
    }
  });

// Add commands to program
program.addCommand(runCommand);
program.addCommand(masterCommand);
program.addCommand(workerCommand);

// Parse arguments
program.parse(process.argv);

// Show help if no arguments provided
if (!process.argv.slice(2).length) {
  program.outputHelp();
}
