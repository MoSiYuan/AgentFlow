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
import type { MasterConfig, WorkerConfig } from '@agentflow/shared';

// CLI version
const VERSION = '1.0.0';

program.version(VERSION);

// Master command
const masterCommand = new Command('master')
  .description('Start AgentFlow Master server')
  .option('--host <host>', 'Host to bind to', '0.0.0.0')
  .option('--port <port>', 'Port to listen on', '8848')
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

// Worker command
const workerCommand = new Command('worker')
  .description('Start AgentFlow Worker')
  .option('--master <url>', 'Master server URL', 'http://localhost:8848')
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
program.addCommand(masterCommand);
program.addCommand(workerCommand);

// Parse arguments
program.parse(process.argv);

// Show help if no arguments provided
if (!process.argv.slice(2).length) {
  program.outputHelp();
}
