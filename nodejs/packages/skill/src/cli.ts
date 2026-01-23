#!/usr/bin/env node

/**
 * AgentFlow CLI
 *
 * Command-line interface for AgentFlow task management
 */

import { Command } from 'commander';
import chalk from 'chalk';
import ora from 'ora';
import { AgentFlowSkill } from './index';

const program = new Command();

program
  .name('agentflow')
  .description('AgentFlow - AI Agent Task Collaboration System')
  .version('1.0.0');

// Create task command
program
  .command('create')
  .description('Create a new task')
  .argument('<title>', 'Task title')
  .option('-d, --description <description>', 'Task description')
  .option('-g, --group <group>', 'Task group name', 'default')
  .action(async (title, options) => {
    const skill = new AgentFlowSkill({ group_name: options.group });
    const spinner = ora('Creating task...').start();

    try {
      const taskId = await skill.createTask({
        title,
        description: options.description || title,
        group_name: options.group,
      });

      spinner.succeed(chalk.green(`Task created: ${taskId}`));
      console.log(chalk.gray(`  Master: ${skill['masterUrl']}`));
      console.log(chalk.gray(`  Group: ${options.group}`));
    } catch (error: any) {
      spinner.fail(chalk.red(`Failed to create task: ${error.message}`));
      process.exit(1);
    }
  });

// Execute shell command
program
  .command('exec')
  .description('Execute a shell command')
  .argument('<command>', 'Shell command to execute')
  .action(async (command) => {
    const skill = new AgentFlowSkill();
    const spinner = ora('Executing command...').start();

    try {
      const result = await skill.executeShell(command);
      spinner.succeed(chalk.green('Command executed'));

      console.log(chalk.gray('Output:'));
      console.log(result);
    } catch (error: any) {
      spinner.fail(chalk.red(`Execution failed: ${error.message}`));
      process.exit(1);
    }
  });

// List tasks
program
  .command('list')
  .description('List all tasks')
  .option('-s, --status <status>', 'Filter by status')
  .action(async (options) => {
    const skill = new AgentFlowSkill();
    const spinner = ora('Fetching tasks...').start();

    try {
      const tasks = await skill.listTasks(options.status);
      spinner.succeed(chalk.green(`Found ${tasks.length} tasks`));

      if (tasks.length === 0) {
        console.log(chalk.yellow('No tasks found'));
        return;
      }

      console.log();
      for (const task of tasks) {
        const statusColor =
          task.status === 'completed'
            ? chalk.green
            : task.status === 'failed'
            ? chalk.red
            : task.status === 'running'
            ? chalk.yellow
            : chalk.gray;

        console.log(statusColor(`[${task.status.toUpperCase()}]`) + ` ${task.id}: ${task.title}`);
        if (task.description) {
          console.log(chalk.gray(`  ${task.description.substring(0, 80)}...`));
        }
      }
    } catch (error: any) {
      spinner.fail(chalk.red(`Failed to list tasks: ${error.message}`));
      process.exit(1);
    }
  });

// Get task status
program
  .command('status')
  .description('Get task status')
  .argument('<taskId>', 'Task ID')
  .action(async (taskId) => {
    const skill = new AgentFlowSkill();
    const spinner = ora('Fetching task status...').start();

    try {
      const task = await skill.getTaskStatus(taskId);
      spinner.succeed(chalk.green('Task found'));

      console.log();
      console.log(chalk.bold('Task Details:'));
      console.log(`  ID:       ${task.id}`);
      console.log(`  Title:    ${task.title}`);
      console.log(`  Status:   ${task.status}`);
      console.log(`  Group:    ${task.group_name}`);
      console.log(`  Priority: ${task.priority}`);
      console.log(`  Created:  ${new Date(task.created_at).toLocaleString()}`);

      if (task.started_at) {
        console.log(`  Started:  ${new Date(task.started_at).toLocaleString()}`);
      }

      if (task.completed_at) {
        console.log(`  Completed: ${new Date(task.completed_at).toLocaleString()}`);
      }

      if (task.result) {
        console.log();
        console.log(chalk.bold('Result:'));
        console.log(chalk.gray(task.result.substring(0, 200)));
        if (task.result.length > 200) {
          console.log(chalk.gray('...'));
        }
      }

      if (task.error) {
        console.log();
        console.log(chalk.bold('Error:'));
        console.log(chalk.red(task.error));
      }
    } catch (error: any) {
      spinner.fail(chalk.red(`Failed to get task status: ${error.message}`));
      process.exit(1);
    }
  });

// Check health
program
  .command('health')
  .description('Check Master server health')
  .action(async () => {
    const skill = new AgentFlowSkill();
    const spinner = ora('Checking Master health...').start();

    try {
      const isHealthy = await skill.checkHealth();

      if (isHealthy) {
        spinner.succeed(chalk.green('Master server is healthy'));
        console.log(chalk.gray(`  URL: ${skill['masterUrl']}`));
      } else {
        spinner.fail(chalk.red('Master server is not responding'));
        console.log(chalk.yellow(`  URL: ${skill['masterUrl']}`));
        process.exit(1);
      }
    } catch (error: any) {
      spinner.fail(chalk.red(`Health check failed: ${error.message}`));
      process.exit(1);
    }
  });

// Create workflow
program
  .command('workflow')
  .description('Create a task workflow')
  .argument('<name>', 'Workflow name')
  .option('-t, --tasks <tasks>', 'JSON array of tasks')
  .action(async (name, options) => {
    if (!options.tasks) {
      console.error(chalk.red('Error: --tasks option is required'));
      process.exit(1);
    }

    const skill = new AgentFlowSkill();
    const spinner = ora('Creating workflow...').start();

    try {
      const tasks = JSON.parse(options.tasks);
      const result = await skill.createWorkflow({
        name,
        tasks,
      });

      spinner.succeed(chalk.green('Workflow created'));
      console.log(chalk.gray(result));
    } catch (error: any) {
      spinner.fail(chalk.red(`Failed to create workflow: ${error.message}`));
      process.exit(1);
    }
  });

// Parse and execute
program.parse();
