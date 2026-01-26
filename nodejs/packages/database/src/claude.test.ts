/**
 * Claude Integration Unit Tests
 */

import { AgentFlowDatabase } from './index';

describe('Claude Mapping', () => {
  let db: AgentFlowDatabase;

  beforeEach(() => {
    db = new AgentFlowDatabase(':memory:');
    db.init();
  });

  afterEach(() => {
    db.close();
  });

  describe('createClaudeMapping', () => {
    it('should create a new Claude mapping', () => {
      const result = db.createClaudeMapping({
        task_id: 1,
        session_uuid: 'test-session-123',
        message_uuid: 'msg-001',
        parent_message_uuid: null,
        slug: 'test-slug',
        source: 'test'
      });

      expect(result).toBe(true);
    });

    it('should create mapping with parent UUID', () => {
      const result = db.createClaudeMapping({
        task_id: 2,
        session_uuid: 'test-session-456',
        message_uuid: 'msg-002',
        parent_message_uuid: 'msg-001',
        slug: 'test-slug-2'
      });

      expect(result).toBe(true);
    });

    it('should handle duplicate message UUID', () => {
      db.createClaudeMapping({
        task_id: 3,
        session_uuid: 'test-session-789',
        message_uuid: 'msg-dup',
        slug: 'test-slug-3'
      });

      const result = db.createClaudeMapping({
        task_id: 4,
        session_uuid: 'test-session-789',
        message_uuid: 'msg-dup', // duplicate
        slug: 'test-slug-4'
      });

      // Should fail due to unique constraint on message_uuid
      expect(result).toBe(false);
    });
  });

  describe('getClaudeMappingByTaskID', () => {
    it('should retrieve mapping by task ID', () => {
      db.createClaudeMapping({
        task_id: 1,
        session_uuid: 'test-session-123',
        message_uuid: 'msg-001',
        slug: 'test-slug'
      });

      const mapping = db.getClaudeMappingByTaskID('1');
      expect(mapping).not.toBeNull();
      expect(mapping?.task_id).toBe(1);
      expect(mapping?.session_uuid).toBe('test-session-123');
      expect(mapping?.message_uuid).toBe('msg-001');
    });

    it('should return null for non-existent task', () => {
      const mapping = db.getClaudeMappingByTaskID('999');
      expect(mapping).toBeNull();
    });
  });

  describe('getClaudeMappingByMessageUUID', () => {
    it('should retrieve mapping by message UUID (reverse lookup)', () => {
      db.createClaudeMapping({
        task_id: 1,
        session_uuid: 'test-session-123',
        message_uuid: 'msg-001',
        slug: 'test-slug'
      });

      const mapping = db.getClaudeMappingByMessageUUID('msg-001');
      expect(mapping).not.toBeNull();
      expect(mapping?.task_id).toBe(1);
    });

    it('should return null for non-existent message UUID', () => {
      const mapping = db.getClaudeMappingByMessageUUID('non-existent');
      expect(mapping).toBeNull();
    });
  });

  describe('getTaskChainBySession', () => {
    it('should retrieve all mappings for a session', () => {
      db.createClaudeMapping({
        task_id: 1,
        session_uuid: 'test-session-multi',
        message_uuid: 'msg-001',
        slug: 'slug-1'
      });

      db.createClaudeMapping({
        task_id: 2,
        session_uuid: 'test-session-multi',
        message_uuid: 'msg-002',
        slug: 'slug-2'
      });

      const mappings = db.getTaskChainBySession('test-session-multi');
      expect(mappings).toHaveLength(2);
      expect(mappings[0].task_id).toBe(1);
      expect(mappings[1].task_id).toBe(2);
    });

    it('should return empty array for non-existent session', () => {
      const mappings = db.getTaskChainBySession('non-existent');
      expect(mappings).toHaveLength(0);
    });
  });

  describe('getClaudeMappingBySlug', () => {
    it('should retrieve mapping by slug', () => {
      db.createClaudeMapping({
        task_id: 1,
        session_uuid: 'test-session-123',
        message_uuid: 'msg-001',
        slug: 'friendly-name'
      });

      const mapping = db.getClaudeMappingBySlug('friendly-name');
      expect(mapping).not.toBeNull();
      expect(mapping?.task_id).toBe(1);
    });

    it('should return null for non-existent slug', () => {
      const mapping = db.getClaudeMappingBySlug('non-existent');
      expect(mapping).toBeNull();
    });
  });

  describe('updateClaudeMappingSlug', () => {
    it('should update slug for a task', () => {
      db.createClaudeMapping({
        task_id: 1,
        session_uuid: 'test-session-123',
        message_uuid: 'msg-001',
        slug: 'old-slug'
      });

      const result = db.updateClaudeMappingSlug('1', 'new-slug');
      expect(result).toBe(true);

      const mapping = db.getClaudeMappingByTaskID('1');
      expect(mapping?.slug).toBe('new-slug');
    });

    it('should return false for non-existent task', () => {
      const result = db.updateClaudeMappingSlug('999', 'new-slug');
      expect(result).toBe(false);
    });
  });

  describe('deleteClaudeMapping', () => {
    it('should delete mapping', () => {
      db.createClaudeMapping({
        task_id: 1,
        session_uuid: 'test-session-123',
        message_uuid: 'msg-001',
        slug: 'test-slug'
      });

      const result = db.deleteClaudeMapping('1');
      expect(result).toBe(true);

      const mapping = db.getClaudeMappingByTaskID('1');
      expect(mapping).toBeNull();
    });

    it('should return false for non-existent task', () => {
      const result = db.deleteClaudeMapping('999');
      expect(result).toBe(false);
    });
  });

  describe('listClaudeMappingsBySession', () => {
    beforeEach(() => {
      // Create multiple mappings
      for (let i = 1; i <= 5; i++) {
        db.createClaudeMapping({
          task_id: i,
          session_uuid: 'test-session-pagination',
          message_uuid: `msg-${i}`,
          slug: `slug-${i}`
        });
      }
    });

    it('should return paginated results', () => {
      const page1 = db.listClaudeMappingsBySession('test-session-pagination', 2, 0);
      expect(page1).toHaveLength(2);
      expect(page1[0].task_id).toBe(1);
      expect(page1[1].task_id).toBe(2);

      const page2 = db.listClaudeMappingsBySession('test-session-pagination', 2, 2);
      expect(page2).toHaveLength(2);
      expect(page2[0].task_id).toBe(3);
      expect(page2[1].task_id).toBe(4);
    });

    it('should respect limit and offset', () => {
      const all = db.listClaudeMappingsBySession('test-session-pagination', 10, 0);
      expect(all).toHaveLength(5);
    });
  });
});
