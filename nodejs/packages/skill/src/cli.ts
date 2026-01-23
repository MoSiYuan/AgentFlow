#!/usr/bin/env node

/**
 * AgentFlow CLI
 *
 * Command-line interface for AgentFlow task management
 */

import { Command } from 'commander';
import chalk from 'chalk';
import ora from 'ora';
import { existsSync, copyFileSync, mkdirSync, readdirSync } from 'fs';
import { join } from 'path';
import { AgentFlowSkill } from './index';

// Get package directory
const getPackageDir = () => {
  return join(__dirname, '../');
};

const program = new Command();

program
  .name('agentflow')
  .description('AgentFlow - AI Agent Task Collaboration System')
  .version('1.0.0');

// Initialize AgentFlow in current directory
program
  .command('init')
  .description('Initialize AgentFlow in current directory')
  .option('-f, --force', 'Overwrite existing .agentflow directory')
  .action(async (options) => {
    const agentflowDir = join(process.cwd(), '.agentflow');

    if (existsSync(agentflowDir) && !options.force) {
      console.log(chalk.yellow('.agentflow directory already exists'));
      console.log(chalk.gray('Use --force to overwrite'));
      process.exit(1);
    }

    const spinner = ora('Initializing AgentFlow...').start();

    try {
      // Create .agentflow directory structure
      const dirs = ['agents', 'skills', 'workflows', 'examples', 'rules'];
      for (const dir of dirs) {
        mkdirSync(join(agentflowDir, dir), { recursive: true });
      }

      // Copy template files from package
      const templateDir = join(getPackageDir(), 'templates');

      if (existsSync(templateDir)) {
        copyTemplateFiles(templateDir, agentflowDir);
      } else {
        // Create default files if templates don't exist
        createDefaultFiles(agentflowDir);
      }

      spinner.succeed(chalk.green('AgentFlow initialized'));
      console.log();
      console.log(chalk.bold('Next steps:'));
      console.log(chalk.gray('  1. Start Master server: cd nodejs && node packages/master/dist/index.js'));
      console.log(chalk.gray('  2. Create task: agentflow create "My first task" -d "npm test"'));
      console.log(chalk.gray('  3. Check status: agentflow list'));
      console.log();
      console.log(chalk.gray('See .agentflow/ directory for agents, skills, and workflows'));
    } catch (error: any) {
      spinner.fail(chalk.red(`Initialization failed: ${error.message}`));
      process.exit(1);
    }
  });

// Update AgentFlow templates
program
  .command('update')
  .description('Update AgentFlow templates to latest version')
  .action(async () => {
    const spinner = ora('Updating AgentFlow...').start();

    try {
      // TODO: Fetch latest templates from GitHub or npm
      spinner.info(chalk.yellow('Update feature coming soon'));
      console.log(chalk.gray('Manually update by reinstalling: npm install -g @agentflow/skill'));
    } catch (error: any) {
      spinner.fail(chalk.red(`Update failed: ${error.message}`));
      process.exit(1);
    }
  });

// Check AgentFlow installation status
program
  .command('info')
  .description('Check AgentFlow installation status')
  .action(async () => {
    console.log();
    console.log(chalk.bold('AgentFlow Status'));
    console.log(chalk.gray('─'.repeat(50)));

    // Check .agentflow directory
    const agentflowDir = join(process.cwd(), '.agentflow');
    if (existsSync(agentflowDir)) {
      console.log(chalk.green('✓') + ' .agentflow directory exists');

      // List contents
      const dirs = readdirSync(agentflowDir);
      console.log(chalk.gray('  Contents:'), dirs.join(', '));
    } else {
      console.log(chalk.yellow('○') + ' .agentflow directory not found');
      console.log(chalk.gray('  Run: agentflow init'));
    }

    // Check Master server
    console.log();
    const skill = new AgentFlowSkill();
    const spinner = ora('Checking Master...').start();

    try {
      const isHealthy = await skill.checkHealth();
      if (isHealthy) {
        spinner.succeed(chalk.green('✓') + ' Master server is running');
        console.log(chalk.gray(`  URL: ${skill['masterUrl']}`));
      } else {
        spinner.fail(chalk.red('✗') + ' Master server is not responding');
        console.log(chalk.gray('  Start: cd nodejs && node packages/master/dist/index.js'));
      }
    } catch {
      spinner.fail(chalk.red('✗') + ' Cannot connect to Master server');
    }

    console.log();
    console.log(chalk.gray('Version: 1.0.0'));
  });

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

// Helper functions
function copyTemplateFiles(templateDir: string, targetDir: string) {
  const files = ['README.md', 'config.example.json'];
  for (const file of files) {
    const src = join(templateDir, file);
    const dest = join(targetDir, file);
    if (existsSync(src)) {
      copyFileSync(src, dest);
    }
  }
}

function createDefaultFiles(agentflowDir: string) {
  // Create README
  const readme = `# AgentFlow Configuration

This directory contains your AgentFlow configuration.

## Structure

- \`agents/\` - Agent templates
- \`skills/\` - Skill definitions
- \`workflows/\` - Workflow templates
- \`examples/\` - Usage examples
- \`rules/\` - Workspace rules

## Documentation

See [AI Integration Guide](https://github.com/MoSiYuan/AgentFlow) for details.
`;
  require('fs').writeFileSync(join(agentflowDir, 'README.md'), readme);

  // Create config example
  const config = JSON.stringify({
    version: '1.0.0',
    master: { url: 'http://localhost:6767' },
    workers: { default_group: 'default' },
    tasks: { default_priority: 50, max_retries: 3 },
  }, null, 2);
  require('fs').writeFileSync(join(agentflowDir, 'config.example.json'), config);
}

// Parse and execute
program.parse();
