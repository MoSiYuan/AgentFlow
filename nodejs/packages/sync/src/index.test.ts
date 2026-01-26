/**
 * Task Synchronizer Unit Tests (Node.js)
 */

import { TaskSynchronizer } from './index';
import { AgentFlowDatabase } from '@agentflow/database';
import { rmSync, mkdirSync, existsSync } from 'fs';
import path from 'path';
import { readFileSync, writeFileSync } from 'fs';

describe('TaskSynchronizer', () => {
  let db: AgentFlowDatabase;
  let tempDir: string;
  let synchronizer: TaskSynchronizer;

  beforeEach(() => {
    // Create temporary directory for Claude files
    tempDir = path.join('/tmp', 'test-sync-' + Date.now());
    mkdirSync(tempDir, { recursive: true });

    // Create in-memory database
    db = new AgentFlowDatabase(':memory:');
    db.init();

    // Create test task
    const taskID = db.createTask({
      title: 'Test Task',
      description: 'Test Description',
      group_name: 'default',
      priority: 0
    });

    // Create Claude mapping
    db.createClaudeMapping({
      task_id: parseInt(taskID),
      session_uuid: 'test-session-123',
      message_uuid: 'test-message-456',
      slug: 'test-task',
      source: 'test'
    });

    synchronizer = new TaskSynchronizer(db, {
      claudeDir: tempDir
    });
  });

  afterEach(() => {
    // Cleanup
    if (existsSync(tempDir)) {
      rmSync(tempDir, { recursive: true, force: true });
    }
    db.close();
  });

  describe('syncTaskStatus', () => {
    it('should sync task status to Claude', async () => {
      await synchronizer.syncTaskStatus('1', 'to_claude');

      // Verify session file was created
      const sessionFile = path.join(tempDir, 'test-session-123.jsonl');
      expect(existsSync(sessionFile)).toBe(true);

      // Verify file content
      const content = readFileSync(sessionFile, 'utf-8');
      expect(content.length).toBeGreaterThan(0);

      const lines = content.trim().split('\n');
      expect(lines.length).toBe(1);

      const update = JSON.parse(lines[0]);
      expect(update.type).toBe('agentflow_status_update');
      expect(update.data.task_id).toBe('1');
    });

    it('should sync task status from Claude', async () => {
      // Create session file with status update
      const sessionFile = path.join(tempDir, 'test-session-123.jsonl');
      const update = {
        type: 'agentflow_status_update',
        timestamp: new Date().toISOString(),
        data: {
          task_id: '1',
          message_uuid: 'test-message-456',
          status: 'completed'
        }
      };

      writeFileSync(sessionFile, JSON.stringify(update) + '\n');

      // Update task status to pending first
      db.updateTaskStatus('1', 'pending');

      // Sync from Claude
      await synchronizer.syncTaskStatus('1', 'from_claude');

      // Verify task status was updated
      const task = db.getTask('1');
      expect(task.status).toBe('completed');
    });

    it('should sync bidirectionally', async () => {
      // Create session file with status update
      const sessionFile = path.join(tempDir, 'test-session-123.jsonl');
      const update = {
        type: 'agentflow_status_update',
        timestamp: new Date().toISOString(),
        data: {
          task_id: '1',
          message_uuid: 'test-message-456',
          status: 'in_progress'
        }
      };

      writeFileSync(sessionFile, JSON.stringify(update) + '\n');

      // Sync bidirectionally
      await synchronizer.syncTaskStatus('1', 'bidirectional');

      // Verify session file has new entries
      const content = readFileSync(sessionFile, 'utf-8');
      const lines = content.trim().split('\n');
      expect(lines.length).toBeGreaterThanOrEqual(2);
    });

    it('should handle non-existent task gracefully', async () => {
      await expect(
        synchronizer.syncTaskStatus('999', 'to_claude')
      ).rejects.toThrow();
    });
  });

  describe('syncTaskChain', () => {
    beforeEach(() => {
      // Create additional tasks
      const taskID2 = db.createTask({
        title: 'Task 2',
        description: 'Description',
        group_name: 'default',
        priority: 0
      });

      const taskID3 = db.createTask({
        title: 'Task 3',
        description: 'Description',
        group_name: 'default',
        priority: 0
      });

      // Create Claude mappings for tasks
      db.createClaudeMapping({
        task_id: parseInt(taskID2),
        session_uuid: 'test-session-123',
        message_uuid: 'test-message-789',
        slug: 'test-task-2'
      });

      db.createClaudeMapping({
        task_id: parseInt(taskID3),
        session_uuid: 'test-session-123',
        message_uuid: 'test-message-012',
        slug: 'test-task-3'
      });

      // Create task chain
      db.createTaskChain({
        id: 'chain-1',
        session_uuid: 'test-session-123',
        root_message_uuid: 'test-message-456',
        chain_type: 'sequential'
      });

      // Create chain nodes
      db.createTaskChainNode({
        chain_id: 'chain-1',
        task_id: 1,
        node_order: 0
      });

      db.createTaskChainNode({
        chain_id: 'chain-1',
        task_id: parseInt(taskID2),
        node_order: 1
      });

      db.createTaskChainNode({
        chain_id: 'chain-1',
        task_id: parseInt(taskID3),
        node_order: 2
      });
    });

    it('should sync all tasks in chain', async () => {
      await synchronizer.syncTaskChain('chain-1', 'to_claude');

      // Verify session file was created
      const sessionFile = path.join(tempDir, 'test-session-123.jsonl');
      expect(existsSync(sessionFile)).toBe(true);

      const content = readFileSync(sessionFile, 'utf-8');
      const lines = content.trim().split('\n');
      expect(lines.length).toBe(3);
    });
  });

  describe('getSyncStatus', () => {
    it('should return sync status', () => {
      const status = synchronizer.getSyncStatus();

      expect(status).toHaveProperty('auto_sync');
      expect(status).toHaveProperty('sync_interval');
      expect(status).toHaveProperty('claude_dir');
      expect(status.auto_sync).toBe(true);
      expect(status.claude_dir).toBe(tempDir);
    });
  });

  describe('validateClaudeSession', () => {
    it('should validate existing session', async () => {
      // Create session file by syncing
      await synchronizer.syncTaskStatus('1', 'to_claude');

      const result = synchronizer.validateClaudeSession('test-session-123');
      expect(result.valid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should return error for non-existent session', () => {
      const result = synchronizer.validateClaudeSession('non-existent-session');
      expect(result.valid).toBe(false);
      expect(result.error).toContain('does not exist');
    });
  });

  describe('setAutoSync', () => {
    it('should set auto-sync enabled state', () => {
      synchronizer.setAutoSync(false);

      const status = synchronizer.getSyncStatus();
      expect(status.auto_sync).toBe(false);

      synchronizer.setAutoSync(true);

      const status2 = synchronizer.getSyncStatus();
      expect(status2.auto_sync).toBe(true);
    });
  });

  describe('setSyncInterval', () => {
    it('should set sync interval', () => {
      synchronizer.setSyncInterval(10000);

      const status = synchronizer.getSyncStatus();
      expect(status.sync_interval).toBe(10000);
    });
  });

  describe('startAutoSync/stopAutoSync', () => {
    it('should start and stop auto-sync', () => {
      synchronizer.startAutoSync();

      let status = synchronizer.getSyncStatus();
      expect(status.is_running).toBe(true);

      synchronizer.stopAutoSync();

      status = synchronizer.getSyncStatus();
      expect(status.is_running).toBeUndefined();
    });

    it('should not start auto-sync if disabled', () => {
      synchronizer.setAutoSync(false);
      synchronizer.startAutoSync();

      const status = synchronizer.getSyncStatus();
      expect(status.is_running).toBeUndefined();
    });
  });

  describe('readStatusUpdates', () => {
    it('should filter only agentflow_status_update types', async () => {
      const sessionFile = path.join(tempDir, 'test-session-123.jsonl');

      const updates = [
        {
          type: 'agentflow_status_update',
          timestamp: new Date().toISOString(),
          data: { task_id: '1', message_uuid: 'test-message-456', status: 'pending' }
        },
        {
          type: 'other_type',
          timestamp: new Date().toISOString(),
          data: { task_id: '1', message_uuid: 'test-message-456', status: 'running' }
        },
        {
          type: 'agentflow_status_update',
          timestamp: new Date().toISOString(),
          data: { task_id: '1', message_uuid: 'test-message-456', status: 'completed' }
        }
      ];

      writeFileSync(sessionFile, updates.map(u => JSON.stringify(u)).join('\n') + '\n');

      // Access private method through test
      const synchronizerAny = synchronizer as any;
      const readUpdates = synchronizerAny.readStatusUpdates(sessionFile);

      expect(readUpdates).toHaveLength(2);
      expect(readUpdates[0].type).toBe('agentflow_status_update');
      expect(readUpdates[1].type).toBe('agentflow_status_update');
    });
  });
});
