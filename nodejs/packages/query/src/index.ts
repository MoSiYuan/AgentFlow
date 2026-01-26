/**
 * Unified Query Interface for AgentFlow Node.js
 *
 * Provides multi-dimensional filtering for AgentFlow and Claude data
 */

export interface TaskQuery {
  // AgentFlow conditions
  task_id?: string;
  status?: string;
  group?: string;

  // Claude conditions
  session_uuid?: string;
  message_uuid?: string;
  slug?: string;
  parent_message_uuid?: string;

  // Pagination
  limit?: number;
  offset?: number;

  // Sorting
  order_by?: string;
  order_desc?: boolean;
}

export interface ClaudeMapping {
  id?: number;
  task_id: string;
  session_uuid: string;
  message_uuid: string;
  parent_message_uuid?: string;
  slug?: string;
  source?: string;
  created_at?: string;
  updated_at?: string;
}

export interface UnifiedTask {
  id: string;
  title: string;
  description?: string;
  status: string;
  group?: string;
  agent_id?: string;
  created_at: string;
  updated_at: string;
  claude?: ClaudeMapping;
}

export interface UnifiedQueryResult {
  tasks: UnifiedTask[];
  total_count: number;
  limit: number;
  offset: number;
}

/**
 * UnifiedQuery provides unified query interface for AgentFlow and Claude data
 */
export class UnifiedQuery {
  constructor(private db: any) {}

  /**
   * Query tasks with multiple filter conditions
   */
  queryTasks(query: TaskQuery): UnifiedQueryResult {
    const { sql, args } = this.buildQuery(query);

    // Execute query
    const rows = this.db.GetDB().prepare(sql).all(...args);

    // Parse results
    const tasks: UnifiedTask[] = [];
    const taskMap = new Map<string, UnifiedTask>();

    for (const row of rows) {
      // Check if task already exists in map
      if (taskMap.has(row.id)) {
        continue;
      }

      const task: UnifiedTask = {
        id: row.id,
        title: row.title,
        description: row.description,
        status: row.status,
        group: row.group_name,
        agent_id: row.agent_id,
        created_at: row.created_at,
        updated_at: row.updated_at
      };

      // Add Claude mapping if exists
      if (row.claude_id) {
        task.claude = {
          id: row.claude_id,
          task_id: String(row.claude_task_id),
          session_uuid: row.claude_session_uuid,
          message_uuid: row.claude_message_uuid,
          parent_message_uuid: row.claude_parent_message_uuid,
          slug: row.claude_slug,
          source: row.claude_source,
          created_at: row.claude_created_at,
          updated_at: row.claude_updated_at
        };
      }

      taskMap.set(row.id, task);
      tasks.push(task);
    }

    // Get total count
    const { sql: countSQL, args: countArgs } = this.buildCountQuery(query);
    const countResult = this.db.GetDB().prepare(countSQL).get(...countArgs);
    const totalCount = countResult.count;

    // Build result
    const limit = query.limit || 50;
    const offset = query.offset || 0;

    return {
      tasks,
      total_count: totalCount,
      limit,
      offset
    };
  }

  /**
   * Build SQL query based on filter conditions
   */
  private buildQuery(query: TaskQuery): { sql: string; args: any[] } {
    const args: any[] = [];
    let argIndex = 1;

    // SELECT clause with LEFT JOIN for Claude mapping
    let sql = `
      SELECT DISTINCT
        t.id, t.title, t.description, t.status, t.group_name,
        t.agent_id, t.created_at, t.updated_at,
        cm.id as claude_id, cm.task_id as claude_task_id,
        cm.session_uuid as claude_session_uuid,
        cm.message_uuid as claude_message_uuid,
        cm.parent_message_uuid as claude_parent_message_uuid,
        cm.slug as claude_slug, cm.source as claude_source,
        cm.created_at as claude_created_at,
        cm.updated_at as claude_updated_at
      FROM tasks t
      LEFT JOIN claude_mappings cm ON t.id = cm.task_id
    `;

    // WHERE clause
    const whereConditions: string[] = [];

    // AgentFlow conditions
    if (query.task_id) {
      whereConditions.push(`t.id = ?${argIndex++}`);
      args.push(query.task_id);
    }

    if (query.status) {
      whereConditions.push(`t.status = ?${argIndex++}`);
      args.push(query.status);
    }

    if (query.group) {
      whereConditions.push(`t.group_name = ?${argIndex++}`);
      args.push(query.group);
    }

    // Claude conditions
    if (query.session_uuid) {
      whereConditions.push(`cm.session_uuid = ?${argIndex++}`);
      args.push(query.session_uuid);
    }

    if (query.message_uuid) {
      whereConditions.push(`cm.message_uuid = ?${argIndex++}`);
      args.push(query.message_uuid);
    }

    if (query.slug) {
      whereConditions.push(`cm.slug = ?${argIndex++}`);
      args.push(query.slug);
    }

    if (query.parent_message_uuid) {
      whereConditions.push(`cm.parent_message_uuid = ?${argIndex++}`);
      args.push(query.parent_message_uuid);
    }

    // Add WHERE clause if there are conditions
    if (whereConditions.length > 0) {
      sql += ' WHERE ' + whereConditions.join(' AND ');
    }

    // ORDER BY clause
    if (query.order_by) {
      const orderDirection = query.order_desc ? 'DESC' : 'ASC';
      sql += ` ORDER BY ${query.order_by} ${orderDirection}`;
    } else {
      sql += ' ORDER BY t.created_at DESC';
    }

    // LIMIT and OFFSET
    const limit = query.limit || 50;
    const offset = query.offset || 0;

    sql += ` LIMIT ${limit} OFFSET ${offset}`;

    // Replace ?1, ?2, ... with ? for better-sqlite3
    sql = sql.replace(/\?\d+/g, '?');

    return { sql, args };
  }

  /**
   * Build COUNT query
   */
  private buildCountQuery(query: TaskQuery): { sql: string; args: any[] } {
    const args: any[] = [];
    let argIndex = 1;

    let sql = `SELECT COUNT(DISTINCT t.id) as count FROM tasks t LEFT JOIN claude_mappings cm ON t.id = cm.task_id`;

    // WHERE clause (same as buildQuery)
    const whereConditions: string[] = [];

    if (query.task_id) {
      whereConditions.push(`t.id = ?${argIndex++}`);
      args.push(query.task_id);
    }

    if (query.status) {
      whereConditions.push(`t.status = ?${argIndex++}`);
      args.push(query.status);
    }

    if (query.group) {
      whereConditions.push(`t.group_name = ?${argIndex++}`);
      args.push(query.group);
    }

    if (query.session_uuid) {
      whereConditions.push(`cm.session_uuid = ?${argIndex++}`);
      args.push(query.session_uuid);
    }

    if (query.message_uuid) {
      whereConditions.push(`cm.message_uuid = ?${argIndex++}`);
      args.push(query.message_uuid);
    }

    if (query.slug) {
      whereConditions.push(`cm.slug = ?${argIndex++}`);
      args.push(query.slug);
    }

    if (query.parent_message_uuid) {
      whereConditions.push(`cm.parent_message_uuid = ?${argIndex++}`);
      args.push(query.parent_message_uuid);
    }

    if (whereConditions.length > 0) {
      sql += ' WHERE ' + whereConditions.join(' AND ');
    }

    // Replace ?1, ?2, ... with ? for better-sqlite3
    sql = sql.replace(/\?\d+/g, '?');

    return { sql, args };
  }

  /**
   * Get task by Claude Message UUID
   */
  getTaskByClaudeMessageUUID(messageUUID: string): UnifiedTask | null {
    const query: TaskQuery = {
      message_uuid: messageUUID,
      limit: 1
    };

    const result = this.queryTasks(query);
    if (result.tasks.length === 0) {
      return null;
    }

    return result.tasks[0];
  }

  /**
   * Get tasks by Claude session UUID
   */
  getTasksByClaudeSession(sessionUUID: string, limit = 50, offset = 0): UnifiedQueryResult {
    const query: TaskQuery = {
      session_uuid: sessionUUID,
      limit,
      offset
    };

    return this.queryTasks(query);
  }

  /**
   * Get task by Claude slug
   */
  getTaskBySlug(slug: string): UnifiedTask | null {
    const query: TaskQuery = {
      slug: slug,
      limit: 1
    };

    const result = this.queryTasks(query);
    if (result.tasks.length === 0) {
      return null;
    }

    return result.tasks[0];
  }

  /**
   * Get tasks with Claude information
   */
  getTasksWithClaudeInfo(limit = 50, offset = 0): UnifiedQueryResult {
    const query: TaskQuery = {
      limit,
      offset
    };

    const result = this.queryTasks(query);

    // Filter only tasks with Claude info
    const tasksWithClaude = result.tasks.filter(task => task.claude !== undefined);

    return {
      ...result,
      tasks: tasksWithClaude,
      total_count: tasksWithClaude.length
    };
  }

  /**
   * Get tasks by multiple filters (convenience method)
   */
  getTasksByFilters(filters: {
    status?: string;
    group?: string;
    session_uuid?: string;
    limit?: number;
    offset?: number;
  }): UnifiedQueryResult {
    const query: TaskQuery = {
      status: filters.status,
      group: filters.group,
      session_uuid: filters.session_uuid,
      limit: filters.limit,
      offset: filters.offset
    };

    return this.queryTasks(query);
  }
}

export default UnifiedQuery;
