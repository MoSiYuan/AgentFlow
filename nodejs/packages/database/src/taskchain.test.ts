/**
 * Task Chain Unit Tests
 */

import { AgentFlowDatabase } from './index';

describe('Task Chain', () => {
  let db: AgentFlowDatabase;

  beforeEach(() => {
    db = new AgentFlowDatabase(':memory:');
    db.init();
  });

  afterEach(() => {
    db.close();
  });

  describe('createTaskChain', () => {
    it('should create a sequential task chain', () => {
      // First create some test tasks
      const task1 = db.createTask('Task 1', 'Description 1', 'default', '');
      const task2 = db.createTask('Task 2', 'Description 2', 'default', '');
      const task3 = db.createTask('Task 3', 'Description 3', 'default', '');

      const result = db.createTaskChain({
        id: 'chain-1',
        session_uuid: 'test-session-123',
        root_message_uuid: 'msg-root-001',
        chain_type: 'sequential',
        status: 'pending'
      });

      expect(result).toBe(true);

      // Verify chain was created
      const chain = db.getTaskChain('chain-1');
      expect(chain).not.toBeNull();
      expect(chain?.id).toBe('chain-1');
      expect(chain?.session_uuid).toBe('test-session-123');
      expect(chain?.chain_type).toBe('sequential');
      expect(chain?.status).toBe('pending');
    });

    it('should create a parallel task chain', () => {
      const result = db.createTaskChain({
        id: 'chain-2',
        session_uuid: 'test-session-456',
        root_message_uuid: 'msg-root-002',
        chain_type: 'parallel',
        status: 'pending'
      });

      expect(result).toBe(true);

      const chain = db.getTaskChain('chain-2');
      expect(chain?.chain_type).toBe('parallel');
    });

    it('should create a tree task chain', () => {
      const result = db.createTaskChain({
        id: 'chain-3',
        session_uuid: 'test-session-789',
        root_message_uuid: 'msg-root-003',
        chain_type: 'tree',
        status: 'pending'
      });

      expect(result).toBe(true);

      const chain = db.getTaskChain('chain-3');
      expect(chain?.chain_type).toBe('tree');
    });

    it('should handle duplicate chain ID', () => {
      db.createTaskChain({
        id: 'chain-dup',
        session_uuid: 'test-session-dup',
        root_message_uuid: 'msg-root-dup',
        chain_type: 'sequential'
      });

      const result = db.createTaskChain({
        id: 'chain-dup', // duplicate
        session_uuid: 'test-session-dup-2',
        root_message_uuid: 'msg-root-dup-2',
        chain_type: 'sequential'
      });

      expect(result).toBe(false);
    });
  });

  describe('createTaskChainNode', () => {
    it('should create a task chain node', () => {
      // Create a task first
      const taskID = db.createTask('Test Task', 'Description', 'default', '');

      // Create a chain
      db.createTaskChain({
        id: 'chain-node-1',
        session_uuid: 'test-session-node',
        root_message_uuid: 'msg-root-node',
        chain_type: 'sequential'
      });

      const result = db.createTaskChainNode({
        chain_id: 'chain-node-1',
        task_id: parseInt(taskID),
        node_order: 0
      });

      expect(result).toBe(true);

      // Verify node was created
      const nodes = db.getTaskChainNodes('chain-node-1');
      expect(nodes).toHaveLength(1);
      expect(nodes[0].task_id).toBe(parseInt(taskID));
      expect(nodes[0].node_order).toBe(0);
    });

    it('should create node with parent', () => {
      const task1 = db.createTask('Task 1', 'Desc 1', 'default', '');
      const task2 = db.createTask('Task 2', 'Desc 2', 'default', '');

      db.createTaskChain({
        id: 'chain-node-2',
        session_uuid: 'test-session-parent',
        root_message_uuid: 'msg-root-parent',
        chain_type: 'tree'
      });

      // Create parent node
      db.createTaskChainNode({
        chain_id: 'chain-node-2',
        task_id: parseInt(task1),
        node_order: 0
      });

      // Get the parent node ID
      const nodes = db.getTaskChainNodes('chain-node-2');
      const parentID = nodes[0].id;

      // Create child node
      const result = db.createTaskChainNode({
        chain_id: 'chain-node-2',
        task_id: parseInt(task2),
        parent_node_id: parentID,
        node_order: 1
      });

      expect(result).toBe(true);

      const childNodes = db.getTaskChainNodes('chain-node-2');
      expect(childNodes).toHaveLength(2);
      expect(childNodes[1].parent_node_id).toBe(parentID);
    });

    it('should handle non-existent chain', () => {
      const result = db.createTaskChainNode({
        chain_id: 'non-existent-chain',
        task_id: 999,
        node_order: 0
      });

      expect(result).toBe(false);
    });
  });

  describe('getTaskChain', () => {
    it('should retrieve chain by ID', () => {
      db.createTaskChain({
        id: 'chain-get-1',
        session_uuid: 'test-session-get',
        root_message_uuid: 'msg-root-get',
        chain_type: 'sequential'
      });

      const chain = db.getTaskChain('chain-get-1');
      expect(chain).not.toBeNull();
      expect(chain?.id).toBe('chain-get-1');
      expect(chain?.session_uuid).toBe('test-session-get');
    });

    it('should return null for non-existent chain', () => {
      const chain = db.getTaskChain('non-existent');
      expect(chain).toBeNull();
    });
  });

  describe('getTaskChainNodes', () => {
    beforeEach(() => {
      // Create a chain with multiple nodes
      db.createTaskChain({
        id: 'chain-nodes-test',
        session_uuid: 'test-session-nodes',
        root_message_uuid: 'msg-root-nodes',
        chain_type: 'sequential'
      });

      // Create tasks
      for (let i = 1; i <= 3; i++) {
        const taskID = db.createTask(`Task ${i}`, `Description ${i}`, 'default', '');
        db.createTaskChainNode({
          chain_id: 'chain-nodes-test',
          task_id: parseInt(taskID),
          node_order: i - 1
        });
      }
    });

    it('should retrieve all nodes for a chain', () => {
      const nodes = db.getTaskChainNodes('chain-nodes-test');
      expect(nodes).toHaveLength(3);
    });

    it('should maintain node order', () => {
      const nodes = db.getTaskChainNodes('chain-nodes-test');
      expect(nodes[0].node_order).toBe(0);
      expect(nodes[1].node_order).toBe(1);
      expect(nodes[2].node_order).toBe(2);
    });

    it('should return empty array for non-existent chain', () => {
      const nodes = db.getTaskChainNodes('non-existent');
      expect(nodes).toHaveLength(0);
    });
  });

  describe('getTaskChainsBySession', () => {
    beforeEach(() => {
      // Create multiple chains for the same session
      db.createTaskChain({
        id: 'chain-session-1',
        session_uuid: 'test-session-multi',
        root_message_uuid: 'msg-root-1',
        chain_type: 'sequential'
      });

      db.createTaskChain({
        id: 'chain-session-2',
        session_uuid: 'test-session-multi',
        root_message_uuid: 'msg-root-2',
        chain_type: 'parallel'
      });

      // Create a chain for different session
      db.createTaskChain({
        id: 'chain-session-3',
        session_uuid: 'test-session-other',
        root_message_uuid: 'msg-root-3',
        chain_type: 'sequential'
      });
    });

    it('should retrieve all chains for a session', () => {
      const chains = db.getTaskChainsBySession('test-session-multi');
      expect(chains).toHaveLength(2);
    });

    it('should not return chains from other sessions', () => {
      const chains = db.getTaskChainsBySession('test-session-multi');
      const chainIDs = chains.map(c => c.id);
      expect(chainIDs).toContain('chain-session-1');
      expect(chainIDs).toContain('chain-session-2');
      expect(chainIDs).not.toContain('chain-session-3');
    });

    it('should return empty array for non-existent session', () => {
      const chains = db.getTaskChainsBySession('non-existent');
      expect(chains).toHaveLength(0);
    });
  });

  describe('updateTaskChainStatus', () => {
    it('should update chain status to running', () => {
      db.createTaskChain({
        id: 'chain-status-1',
        session_uuid: 'test-session-status',
        root_message_uuid: 'msg-root-status',
        chain_type: 'sequential',
        status: 'pending'
      });

      const result = db.updateTaskChainStatus('chain-status-1', 'running');
      expect(result).toBe(true);

      const chain = db.getTaskChain('chain-status-1');
      expect(chain?.status).toBe('running');
      expect(chain?.started_at).not.toBeNull();
    });

    it('should update chain status to completed', () => {
      db.createTaskChain({
        id: 'chain-status-2',
        session_uuid: 'test-session-status-2',
        root_message_uuid: 'msg-root-status-2',
        chain_type: 'sequential',
        status: 'running'
      });

      const result = db.updateTaskChainStatus('chain-status-2', 'completed');
      expect(result).toBe(true);

      const chain = db.getTaskChain('chain-status-2');
      expect(chain?.status).toBe('completed');
      expect(chain?.completed_at).not.toBeNull();
    });

    it('should update chain status to failed', () => {
      db.createTaskChain({
        id: 'chain-status-3',
        session_uuid: 'test-session-status-3',
        root_message_uuid: 'msg-root-status-3',
        chain_type: 'sequential'
      });

      const result = db.updateTaskChainStatus('chain-status-3', 'failed');
      expect(result).toBe(true);

      const chain = db.getTaskChain('chain-status-3');
      expect(chain?.status).toBe('failed');
    });

    it('should return false for non-existent chain', () => {
      const result = db.updateTaskChainStatus('non-existent', 'running');
      expect(result).toBe(false);
    });
  });

  describe('deleteTaskChain', () => {
    it('should delete a chain', () => {
      db.createTaskChain({
        id: 'chain-delete-1',
        session_uuid: 'test-session-delete',
        root_message_uuid: 'msg-root-delete',
        chain_type: 'sequential'
      });

      const result = db.deleteTaskChain('chain-delete-1');
      expect(result).toBe(true);

      const chain = db.getTaskChain('chain-delete-1');
      expect(chain).toBeNull();
    });

    it('should cascade delete nodes', () => {
      // Create chain with nodes
      db.createTaskChain({
        id: 'chain-delete-2',
        session_uuid: 'test-session-delete-2',
        root_message_uuid: 'msg-root-delete-2',
        chain_type: 'sequential'
      });

      const taskID = db.createTask('Task', 'Description', 'default', '');
      db.createTaskChainNode({
        chain_id: 'chain-delete-2',
        task_id: parseInt(taskID),
        node_order: 0
      });

      // Delete chain
      db.deleteTaskChain('chain-delete-2');

      // Verify nodes are also deleted
      const nodes = db.getTaskChainNodes('chain-delete-2');
      expect(nodes).toHaveLength(0);
    });

    it('should return false for non-existent chain', () => {
      const result = db.deleteTaskChain('non-existent');
      expect(result).toBe(false);
    });
  });

  describe('Integration Tests', () => {
    it('should create complete sequential chain with nodes', () => {
      const sessionUUID = 'test-integration-sequential';
      const taskIDs: number[] = [];

      // Create tasks
      for (let i = 1; i <= 3; i++) {
        const id = db.createTask(`Sequential Task ${i}`, `Description ${i}`, 'default', '');
        taskIDs.push(parseInt(id));
      }

      // Create chain
      db.createTaskChain({
        id: 'integration-chain-1',
        session_uuid: sessionUUID,
        root_message_uuid: 'msg-root-integration-1',
        chain_type: 'sequential'
      });

      // Create nodes in order
      taskIDs.forEach((taskID, index) => {
        db.createTaskChainNode({
          chain_id: 'integration-chain-1',
          task_id: taskID,
          node_order: index
        });
      });

      // Verify
      const chain = db.getTaskChain('integration-chain-1');
      expect(chain?.chain_type).toBe('sequential');

      const nodes = db.getTaskChainNodes('integration-chain-1');
      expect(nodes).toHaveLength(3);
      expect(nodes[0].node_order).toBe(0);
      expect(nodes[1].node_order).toBe(1);
      expect(nodes[2].node_order).toBe(2);
    });

    it('should create complete parallel chain with nodes', () => {
      const sessionUUID = 'test-integration-parallel';
      const taskIDs: number[] = [];

      // Create tasks
      for (let i = 1; i <= 3; i++) {
        const id = db.createTask(`Parallel Task ${i}`, `Description ${i}`, 'default', '');
        taskIDs.push(parseInt(id));
      }

      // Create chain
      db.createTaskChain({
        id: 'integration-chain-2',
        session_uuid: sessionUUID,
        root_message_uuid: 'msg-root-integration-2',
        chain_type: 'parallel'
      });

      // Create nodes (all with order 0 for parallel)
      taskIDs.forEach((taskID) => {
        db.createTaskChainNode({
          chain_id: 'integration-chain-2',
          task_id: taskID,
          node_order: 0
        });
      });

      // Verify
      const chain = db.getTaskChain('integration-chain-2');
      expect(chain?.chain_type).toBe('parallel');

      const nodes = db.getTaskChainNodes('integration-chain-2');
      expect(nodes).toHaveLength(3);
      nodes.forEach(node => {
        expect(node.node_order).toBe(0);
      });
    });

    it('should create complete tree chain with parent-child relationships', () => {
      const sessionUUID = 'test-integration-tree';
      const taskIDs: number[] = [];

      // Create tasks
      for (let i = 1; i <= 3; i++) {
        const id = db.createTask(`Tree Task ${i}`, `Description ${i}`, 'default', '');
        taskIDs.push(parseInt(id));
      }

      // Create chain
      db.createTaskChain({
        id: 'integration-chain-3',
        session_uuid: sessionUUID,
        root_message_uuid: 'msg-root-integration-3',
        chain_type: 'tree'
      });

      // Create root node
      db.createTaskChainNode({
        chain_id: 'integration-chain-3',
        task_id: taskIDs[0],
        node_order: 0
      });

      const nodes = db.getTaskChainNodes('integration-chain-3');
      const rootNodeID = nodes[0].id;

      // Create child nodes
      db.createTaskChainNode({
        chain_id: 'integration-chain-3',
        task_id: taskIDs[1],
        parent_node_id: rootNodeID,
        node_order: 1
      });

      db.createTaskChainNode({
        chain_id: 'integration-chain-3',
        task_id: taskIDs[2],
        parent_node_id: rootNodeID,
        node_order: 2
      });

      // Verify tree structure
      const allNodes = db.getTaskChainNodes('integration-chain-3');
      expect(allNodes).toHaveLength(3);

      const rootNode = allNodes.find(n => n.id === rootNodeID);
      expect(rootNode?.parent_node_id).toBeNull();

      const childNodes = allNodes.filter(n => n.parent_node_id === rootNodeID);
      expect(childNodes).toHaveLength(2);
    });
  });
});
