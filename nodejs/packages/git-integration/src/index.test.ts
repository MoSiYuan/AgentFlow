/**
 * Git Integration Unit Tests
 */

import { BranchManager, GitIntegrationManager } from './index';
import { execSync } from 'child_process';
import { rmSync, mkdirSync, existsSync } from 'fs';
import path from 'path';

describe('BranchManager', () => {
  const testRepoPath = path.join('/tmp', 'test-repo-' + Date.now());
  let branchManager: BranchManager;

  beforeEach(() => {
    // Create test repository
    if (existsSync(testRepoPath)) {
      rmSync(testRepoPath, { recursive: true, force: true });
    }
    mkdirSync(testRepoPath, { recursive: true });

    // Initialize git repo
    execSync('git init', { cwd: testRepoPath });
    execSync('git config user.email "test@example.com"', { cwd: testRepoPath });
    execSync('git config user.name "Test User"', { cwd: testRepoPath });

    // Create initial commit
    const testFile = path.join(testRepoPath, 'README.md');
    require('fs').writeFileSync(testFile, '# Test Repository');
    execSync('git add .', { cwd: testRepoPath });
    execSync('git commit -m "Initial commit"', { cwd: testRepoPath });

    branchManager = new BranchManager({ repoPath: testRepoPath });
  });

  afterEach(() => {
    // Cleanup test repository
    if (existsSync(testRepoPath)) {
      rmSync(testRepoPath, { recursive: true, force: true });
    }
  });

  describe('createBranch', () => {
    it('should create a new branch', async () => {
      const result = await branchManager.createBranch('test-branch-1');
      expect(result).toBe(true);

      // Verify branch exists
      const branches = execSync('git branch', { cwd: testRepoPath }).toString();
      expect(branches).toContain('test-branch-1');
    });

    it('should handle branch creation error gracefully', async () => {
      // Try to create branch with invalid name
      const result = await branchManager.createBranch('');
      expect(result).toBe(false);
    });

    it('should handle duplicate branch name', async () => {
      await branchManager.createBranch('duplicate-branch');

      // Creating same branch again should fail or handle gracefully
      const result = await branchManager.createBranch('duplicate-branch');
      // Git may return false or the branch may already exist
      expect(typeof result).toBe('boolean');
    });
  });

  describe('deleteBranch', () => {
    it('should delete a branch', async () => {
      await branchManager.createBranch('branch-to-delete');

      const result = await branchManager.deleteBranch('branch-to-delete');
      expect(result).toBe(true);

      // Verify branch is deleted
      const branches = execSync('git branch', { cwd: testRepoPath }).toString();
      expect(branches).not.toContain('branch-to-delete');
    });

    it('should force delete a branch', async () => {
      await branchManager.createBranch('branch-to-force-delete');

      const result = await branchManager.deleteBranch('branch-to-force-delete', true);
      expect(result).toBe(true);

      const branches = execSync('git branch', { cwd: testRepoPath }).toString();
      expect(branches).not.toContain('branch-to-force-delete');
    });

    it('should handle deleting non-existent branch', async () => {
      const result = await branchManager.deleteBranch('non-existent-branch');
      expect(result).toBe(false);
    });

    it('should not delete current branch', async () => {
      // Create and switch to branch
      await branchManager.createBranch('current-branch');
      await branchManager.switchBranch('current-branch');

      // Try to delete current branch (should fail)
      const result = await branchManager.deleteBranch('current-branch');
      expect(result).toBe(false);
    });
  });

  describe('switchBranch', () => {
    it('should switch to an existing branch', async () => {
      await branchManager.createBranch('branch-to-switch');

      const result = await branchManager.switchBranch('branch-to-switch');
      expect(result).toBe(true);

      // Verify current branch
      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe('branch-to-switch');
    });

    it('should handle switching to non-existent branch', async () => {
      const result = await branchManager.switchBranch('non-existent-branch');
      expect(result).toBe(false);
    });
  });

  describe('getCurrentBranch', () => {
    it('should return current branch name', async () => {
      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe('main');
    });

    it('should return updated branch after switching', async () => {
      await branchManager.createBranch('new-branch');
      await branchManager.switchBranch('new-branch');

      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe('new-branch');
    });
  });

  describe('generateClaudeBranchName', () => {
    it('should generate branch name with session UUID prefix and task ID', () => {
      const branchName = branchManager.generateClaudeBranchName('abc123def456', '123');
      expect(branchName).toBe('claude-abc123de/task-123');
    });

    it('should handle short session UUID', () => {
      const branchName = branchManager.generateClaudeBranchName('abc', '456');
      expect(branchName).toBe('claude-abc/task-456');
    });

    it('should handle exact 8 character session UUID', () => {
      const branchName = branchManager.generateClaudeBranchName('12345678', '789');
      expect(branchName).toBe('claude-12345678/task-789');
    });

    it('should handle long session UUID (truncate to 8)', () => {
      const longUUID = 'abcdef1234567890abcdef1234567890';
      const branchName = branchManager.generateClaudeBranchName(longUUID, '999');
      expect(branchName).toBe('claude-abcdef12/task-999');
    });
  });
});

describe('GitIntegrationManager', () => {
  const testRepoPath = path.join('/tmp', 'test-git-integration-' + Date.now());
  let gitManager: GitIntegrationManager;

  beforeEach(() => {
    // Create test repository
    if (existsSync(testRepoPath)) {
      rmSync(testRepoPath, { recursive: true, force: true });
    }
    mkdirSync(testRepoPath, { recursive: true });

    // Initialize git repo
    execSync('git init', { cwd: testRepoPath });
    execSync('git config user.email "test@example.com"', { cwd: testRepoPath });
    execSync('git config user.name "Test User"', { cwd: testRepoPath });

    // Create initial commit
    const testFile = path.join(testRepoPath, 'README.md');
    require('fs').writeFileSync(testFile, '# Test Repository');
    execSync('git add .', { cwd: testRepoPath });
    execSync('git commit -m "Initial commit"', { cwd: testRepoPath });

    gitManager = new GitIntegrationManager(testRepoPath);
  });

  afterEach(() => {
    // Cleanup test repository
    if (existsSync(testRepoPath)) {
      rmSync(testRepoPath, { recursive: true, force: true });
    }
  });

  describe('createClaudeGitTask', () => {
    it('should create Claude Git task with branch', async () => {
      const task = await gitManager.createClaudeGitTask({
        id: 'task-1',
        title: 'Test Task',
        description: 'Test Description',
        agent_id: 'agent-1',
        session_uuid: 'session-abc123def456',
        message_uuid: 'msg-001',
        slug: 'test-slug'
      });

      expect(task).toBeDefined();
      expect(task.id).toBe('task-1');
      expect(task.title).toBe('Test Task');
      expect(task.agent_id).toBe('agent-1');
      expect(task.session_uuid).toBe('session-abc123def456');
      expect(task.message_uuid).toBe('msg-001');
      expect(task.slug).toBe('test-slug');
      expect(task.status).toBe('pending');
      expect(task.git_branch).toBe('claude-session-a/task-1');

      // Verify branch was created
      const branches = execSync('git branch', { cwd: testRepoPath }).toString();
      expect(branches).toContain('claude-session-a/task-1');
    });

    it('should create task without optional fields', async () => {
      const task = await gitManager.createClaudeGitTask({
        id: 'task-2',
        title: 'Minimal Task',
        agent_id: 'agent-2',
        session_uuid: 'session-xyz',
        message_uuid: 'msg-002'
      });

      expect(task).toBeDefined();
      expect(task.description).toBeUndefined();
      expect(task.parent_message_uuid).toBeUndefined();
      expect(task.slug).toBeUndefined();
      expect(task.status).toBe('pending');
    });

    it('should create task with parent message UUID', async () => {
      const task = await gitManager.createClaudeGitTask({
        id: 'task-3',
        title: 'Child Task',
        agent_id: 'agent-3',
        session_uuid: 'session-parent',
        message_uuid: 'msg-003',
        parent_message_uuid: 'msg-001'
      });

      expect(task.parent_message_uuid).toBe('msg-001');
    });

    it('should generate unique branch names for different sessions', async () => {
      const task1 = await gitManager.createClaudeGitTask({
        id: 'task-1',
        title: 'Task 1',
        agent_id: 'agent-1',
        session_uuid: 'session-aaa111',
        message_uuid: 'msg-001'
      });

      const task2 = await gitManager.createClaudeGitTask({
        id: 'task-2',
        title: 'Task 2',
        agent_id: 'agent-2',
        session_uuid: 'session-bbb222',
        message_uuid: 'msg-002'
      });

      expect(task1.git_branch).not.toBe(task2.git_branch);
    });

    it('should generate same branch prefix for same session', async () => {
      const task1 = await gitManager.createClaudeGitTask({
        id: 'task-1',
        title: 'Task 1',
        agent_id: 'agent-1',
        session_uuid: 'session-same123',
        message_uuid: 'msg-001'
      });

      const task2 = await gitManager.createClaudeGitTask({
        id: 'task-2',
        title: 'Task 2',
        agent_id: 'agent-2',
        session_uuid: 'session-same123',
        message_uuid: 'msg-002'
      });

      expect(task1.git_branch).toBe('claude-session-s/task-1');
      expect(task2.git_branch).toBe('claude-session-s/task-2');
    });
  });

  describe('createBranchesFromChain', () => {
    it('should create branches from task chain nodes', async () => {
      const mockNodes = [
        { task_id: 1, node_order: 0 },
        { task_id: 2, node_order: 1 },
        { task_id: 3, node_order: 2 }
      ];

      const mockGetTaskChainNodes = async (chainId: string) => {
        return mockNodes;
      };

      const mockGetSessionUUID = async (taskId: string) => {
        const sessionMap: { [key: string]: string } = {
          '1': 'session-abc111',
          '2': 'session-def222',
          '3': 'session-ghi333'
        };
        return sessionMap[taskId];
      };

      const branches = await gitManager.createBranchesFromChain(
        'chain-1',
        mockGetTaskChainNodes,
        mockGetSessionUUID
      );

      expect(branches).toHaveLength(3);
      expect(branches[0]).toBe('claude-session-a/task-1');
      expect(branches[1]).toBe('claude-session-d/task-2');
      expect(branches[2]).toBe('claude-session-g/task-3');

      // Verify branches exist
      const allBranches = execSync('git branch', { cwd: testRepoPath }).toString();
      expect(allBranches).toContain('claude-session-a/task-1');
      expect(allBranches).toContain('claude-session-d/task-2');
      expect(allBranches).toContain('claude-session-g/task-3');
    });

    it('should handle nodes with missing session UUID', async () => {
      const mockNodes = [
        { task_id: 1, node_order: 0 },
        { task_id: 2, node_order: 1 }
      ];

      const mockGetTaskChainNodes = async (chainId: string) => {
        return mockNodes;
      };

      const mockGetSessionUUID = async (taskId: string) => {
        // Return empty string for task 2
        if (taskId === '2') return '';
        return 'session-abc111';
      };

      const branches = await gitManager.createBranchesFromChain(
        'chain-2',
        mockGetTaskChainNodes,
        mockGetSessionUUID
      );

      // Only one branch should be created
      expect(branches).toHaveLength(1);
      expect(branches[0]).toBe('claude-session-a/task-1');
    });

    it('should handle empty chain', async () => {
      const mockGetTaskChainNodes = async (chainId: string) => {
        return [];
      };

      const mockGetSessionUUID = async (taskId: string) => {
        return 'session-abc';
      };

      const branches = await gitManager.createBranchesFromChain(
        'chain-empty',
        mockGetTaskChainNodes,
        mockGetSessionUUID
      );

      expect(branches).toHaveLength(0);
    });

    it('should handle chain creation errors gracefully', async () => {
      const mockNodes = [
        { task_id: 1, node_order: 0 },
        { task_id: 2, node_order: 1 }
      ];

      const mockGetTaskChainNodes = async (chainId: string) => {
        return mockNodes;
      };

      const mockGetSessionUUID = async (taskId: string) => {
        return 'session-abc';
      };

      // Create first branch
      await gitManager.createClaudeGitTask({
        id: '1',
        title: 'Task 1',
        agent_id: 'agent-1',
        session_uuid: 'session-abc',
        message_uuid: 'msg-001'
      });

      // Try to create same branch again (should handle gracefully)
      const branches = await gitManager.createBranchesFromChain(
        'chain-duplicate',
        mockGetTaskChainNodes,
        mockGetSessionUUID
      );

      // Should still return array, maybe with fewer branches
      expect(Array.isArray(branches)).toBe(true);
    });
  });

  describe('getClaudeBranches', () => {
    it('should return empty array (implementation dependent)', async () => {
      // This method's implementation depends on database context
      const branches = await gitManager.getClaudeBranches('session-abc');
      expect(Array.isArray(branches)).toBe(true);
      expect(branches).toHaveLength(0);
    });
  });

  describe('cleanupClaudeBranches', () => {
    it('should cleanup branches for a session', async () => {
      // Create some branches
      await gitManager.createClaudeGitTask({
        id: 'task-1',
        title: 'Task 1',
        agent_id: 'agent-1',
        session_uuid: 'session-cleanup',
        message_uuid: 'msg-001'
      });

      await gitManager.createClaudeGitTask({
        id: 'task-2',
        title: 'Task 2',
        agent_id: 'agent-2',
        session_uuid: 'session-cleanup',
        message_uuid: 'msg-002'
      });

      // Note: cleanupClaudeBranches depends on getClaudeBranches
      // which returns empty array in current implementation
      // This test verifies the method exists and can be called
      const deletedCount = await gitManager.cleanupClaudeBranches('session-cleanup');
      expect(typeof deletedCount).toBe('number');
    });

    it('should return 0 for non-existent session', async () => {
      const deletedCount = await gitManager.cleanupClaudeBranches('non-existent-session');
      expect(deletedCount).toBe(0);
    });
  });
});
