#!/usr/bin/env node
/**
 * AgentFlow Programmatic Usage Example
 *
 * This example demonstrates how to use AgentFlow programmatically
 * with the @agentflow/skill package
 */

import { AgentFlowSkill } from '@agentflow/skill';

async function main() {
  console.log('=== AgentFlow Programmatic Usage ===\n');

  // Initialize skill
  const skill = new AgentFlowSkill({
    master_url: 'http://localhost:6767',
    group_name: 'default'
  });

  try {
    // Check health
    console.log('1. Checking Master health...');
    const isHealthy = await skill.checkHealth();
    console.log(`   Status: ${isHealthy ? '✓ Healthy' : '✗ Unhealthy'}\n`);

    if (!isHealthy) {
      console.error('Error: Master server is not running');
      process.exit(1);
    }

    // Create a single task
    console.log('2. Creating a task...');
    const taskId = await skill.createTask({
      title: 'Build Project',
      description: 'npm run build'
    });
    console.log(`   ✓ Task created: ${taskId}\n`);

    // Create parallel tasks
    console.log('3. Creating parallel tasks...');
    const parallelTasks = await skill.executeParallel([
      { title: 'Run tests', description: 'npm test' },
      { title: 'Run linter', description: 'npm run lint' },
      { title: 'Type check', description: 'npm run type-check' }
    ]);
    console.log(`   ✓ Created ${parallelTasks.length} parallel tasks\n`);

    // List all tasks
    console.log('4. Listing tasks...');
    const allTasks = await skill.listTasks();
    console.log(`   Total tasks: ${allTasks.length}\n`);

    // List pending tasks
    console.log('5. Listing pending tasks...');
    const pendingTasks = await skill.listTasks('pending');
    console.log(`   Pending: ${pendingTasks.length}\n`);

    // Wait and check status
    console.log('6. Waiting 3 seconds...');
    await new Promise(resolve => setTimeout(resolve, 3000));

    // Get task status
    if (parallelTasks.length > 0) {
      console.log('7. Checking task status...');
      const task = await skill.getTaskStatus(parallelTasks[0]);
      console.log(`   Task: ${task.id}`);
      console.log(`   Status: ${task.status}`);
      console.log(`   Title: ${task.title}\n`);
    }

    console.log('=== Example Complete ===');
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
}

// Run if executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}

export { main };
