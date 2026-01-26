/**
 * Unified Query Unit Tests (Node.js)
 */

import { UnifiedQuery } from './index';
import { AgentFlowDatabase } from '@agentflow/database';

describe('UnifiedQuery', () => {
  let db: AgentFlowDatabase;
  let uq: UnifiedQuery;

  beforeEach(() => {
    // Create in-memory database
    db = new AgentFlowDatabase(':memory:');
    db.init();

    // Create test tasks
    for (let i = 1; i <= 5; i++) {
      const taskID = db.createTask({
        title: `Task ${i}`,
        description: `Description ${i}`,
        group_name: 'default',
        priority: 0
      });

      // Create Claude mappings for some tasks
      if (i <= 3) {
        db.createClaudeMapping({
          task_id: parseInt(taskID),
          session_uuid: 'test-session-123',
          message_uuid: `msg-00${i}`,
          slug: `task-${i}`,
          source: 'test'
        });
      }
    }

    uq = new UnifiedQuery(db);
  });

  afterEach(() => {
    db.close();
  });

  describe('queryTasks', () => {
    it('should query all tasks', () => {
      const result = uq.queryTasks({ limit: 10 });

      expect(result.tasks).toHaveLength(5);
      expect(result.total_count).toBe(5);
      expect(result.limit).toBe(10);
      expect(result.offset).toBe(0);
    });

    it('should query tasks by status', () => {
      const result = uq.queryTasks({
        status: 'pending',
        limit: 10
      });

      // All tasks should be pending initially
      expect(result.tasks.length).toBeGreaterThan(0);
      result.tasks.forEach(task => {
        expect(task.status).toBe('pending');
      });
    });

    it('should query tasks by group', () => {
      const result = uq.queryTasks({
        group: 'default',
        limit: 10
      });

      expect(result.tasks.length).toBeGreaterThan(0);
      result.tasks.forEach(task => {
        expect(task.group).toBe('default');
      });
    });

    it('should query tasks by Claude session UUID', () => {
      const result = uq.queryTasks({
        session_uuid: 'test-session-123',
        limit: 10
      });

      expect(result.tasks).toHaveLength(3);
      result.tasks.forEach(task => {
        expect(task.claude).toBeDefined();
        expect(task.claude?.session_uuid).toBe('test-session-123');
      });
    });

    it('should query tasks by Claude message UUID', () => {
      const result = uq.queryTasks({
        message_uuid: 'msg-001',
        limit: 10
      });

      expect(result.tasks).toHaveLength(1);
      expect(result.tasks[0].claude?.message_uuid).toBe('msg-001');
    });

    it('should query tasks by Claude slug', () => {
      const result = uq.queryTasks({
        slug: 'task-1',
        limit: 10
      });

      expect(result.tasks).toHaveLength(1);
      expect(result.tasks[0].claude?.slug).toBe('task-1');
    });

    it('should support pagination', () => {
      const page1 = uq.queryTasks({ limit: 2, offset: 0 });
      expect(page1.tasks).toHaveLength(2);

      const page2 = uq.queryTasks({ limit: 2, offset: 2 });
      expect(page2.tasks).toHaveLength(2);

      // Verify different tasks on different pages
      expect(page1.tasks[0].id).not.toBe(page2.tasks[0].id);
    });

    it('should support sorting', () => {
      const result = uq.queryTasks({
        order_by: 'title',
        order_desc: false,
        limit: 10
      });

      expect(result.tasks.length).toBeGreaterThanOrEqual(2);
      // Verify ascending order
      if (result.tasks[0].title && result.tasks[1].title) {
        expect(result.tasks[0].title < result.tasks[1].title).toBe(true);
      }
    });

    it('should support multiple filters', () => {
      const result = uq.queryTasks({
        session_uuid: 'test-session-123',
        group: 'default',
        limit: 10
      });

      expect(result.tasks.length).toBeGreaterThan(0);
      result.tasks.forEach(task => {
        expect(task.claude?.session_uuid).toBe('test-session-123');
        expect(task.group).toBe('default');
      });
    });
  });

  describe('getTaskByClaudeMessageUUID', () => {
    it('should get task by message UUID', () => {
      const task = uq.getTaskByClaudeMessageUUID('msg-002');

      expect(task).toBeDefined();
      expect(task?.claude?.message_uuid).toBe('msg-002');
    });

    it('should return null for non-existent message UUID', () => {
      const task = uq.getTaskByClaudeMessageUUID('non-existent');
      expect(task).toBeNull();
    });
  });

  describe('getTasksByClaudeSession', () => {
    it('should get tasks by session UUID', () => {
      const result = uq.getTasksByClaudeSession('test-session-123');

      expect(result.tasks).toHaveLength(3);
      expect(result.total_count).toBe(3);

      result.tasks.forEach(task => {
        expect(task.claude?.session_uuid).toBe('test-session-123');
      });
    });

    it('should support pagination', () => {
      const result = uq.getTasksByClaudeSession('test-session-123', 2, 0);

      expect(result.tasks).toHaveLength(2);
      expect(result.limit).toBe(2);
      expect(result.offset).toBe(0);
    });

    it('should return empty result for non-existent session', () => {
      const result = uq.getTasksByClaudeSession('non-existent');

      expect(result.tasks).toHaveLength(0);
      expect(result.total_count).toBe(0);
    });
  });

  describe('getTaskBySlug', () => {
    it('should get task by slug', () => {
      const task = uq.getTaskBySlug('task-2');

      expect(task).toBeDefined();
      expect(task?.claude?.slug).toBe('task-2');
    });

    it('should return null for non-existent slug', () => {
      const task = uq.getTaskBySlug('non-existent-slug');
      expect(task).toBeNull();
    });
  });

  describe('getTasksWithClaudeInfo', () => {
    it('should get only tasks with Claude info', () => {
      const result = uq.getTasksWithClaudeInfo();

      expect(result.tasks.length).toBe(3);
      result.tasks.forEach(task => {
        expect(task.claude).toBeDefined();
      });
    });

    it('should support pagination', () => {
      const result = uq.getTasksWithClaudeInfo(2, 0);

      expect(result.tasks).toHaveLength(2);
      expect(result.tasks[0].claude).toBeDefined();
      expect(result.tasks[1].claude).toBeDefined();
    });
  });

  describe('getTasksByFilters', () => {
    it('should get tasks by status filter', () => {
      const result = uq.getTasksByFilters({
        status: 'pending',
        limit: 10
      });

      expect(result.tasks.length).toBeGreaterThan(0);
      result.tasks.forEach(task => {
        expect(task.status).toBe('pending');
      });
    });

    it('should get tasks by session and status filters', () => {
      const result = uq.getTasksByFilters({
        session_uuid: 'test-session-123',
        status: 'pending',
        limit: 10
      });

      expect(result.tasks.length).toBeGreaterThan(0);
      result.tasks.forEach(task => {
        expect(task.claude?.session_uuid).toBe('test-session-123');
        expect(task.status).toBe('pending');
      });
    });

    it('should get tasks by group and session filters', () => {
      const result = uq.getTasksByFilters({
        group: 'default',
        session_uuid: 'test-session-123',
        limit: 10
      });

      expect(result.tasks.length).toBeGreaterThan(0);
      result.tasks.forEach(task => {
        expect(task.group).toBe('default');
        expect(task.claude?.session_uuid).toBe('test-session-123');
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty result', () => {
      const result = uq.queryTasks({
        status: 'completed',
        limit: 10
      });

      expect(result.tasks).toHaveLength(0);
      expect(result.total_count).toBe(0);
    });

    it('should handle limit larger than result count', () => {
      const result = uq.queryTasks({
        limit: 1000,
        offset: 0
      });

      expect(result.tasks.length).toBeLessThanOrEqual(1000);
      expect(result.total_count).toBe(5);
    });

    it('should handle offset beyond result count', () => {
      const result = uq.queryTasks({
        limit: 10,
        offset: 100
      });

      expect(result.tasks).toHaveLength(0);
    });
  });
});
