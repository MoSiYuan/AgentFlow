/**
 * AgentFlow Memory System - SQLite Schema
 *
 * 记忆系统数据库表结构（SQLite 格式）
 * 包括：工作记忆、长期记忆、会话记忆
 */

export const memorySchemaSQL = `
  -- =====================================================
  -- 1. 工作记忆表 (Working Memory)
  -- =====================================================

  -- 任务上下文表（存储任务执行的中间状态）
  CREATE TABLE IF NOT EXISTS task_context (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER UNIQUE NOT NULL,
    context_data TEXT NOT NULL,
    variables TEXT,
    dependencies TEXT,
    execution_history TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
  );

  -- =====================================================
  -- 2. 长期记忆表 (Long-term Memory)
  -- =====================================================

  -- 任务历史表（用于分析和学习）
  CREATE TABLE IF NOT EXISTS task_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    group_name TEXT,
    status TEXT,
    priority TEXT,
    result TEXT,
    duration_ms INTEGER,
    worker_id TEXT,
    agent_id TEXT,
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    metadata TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL
  );

  -- 经验总结表（从执行历史中提取的知识）
  CREATE TABLE IF NOT EXISTS experience_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    summary_type TEXT NOT NULL CHECK(summary_type IN ('success_pattern', 'failure_pattern', 'optimization', 'best_practice')),
    pattern_description TEXT NOT NULL,
    context TEXT,
    confidence_score REAL DEFAULT 0.5,
    usage_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,
    expires_at DATETIME
  );

  -- 技能使用统计表
  CREATE TABLE IF NOT EXISTS skill_usage_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    skill_name TEXT NOT NULL,
    task_id INTEGER,
    agent_id TEXT,
    execution_count INTEGER DEFAULT 1,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    avg_duration_ms INTEGER,
    last_used_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(skill_name, task_id, agent_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
  );

  -- =====================================================
  -- 3. 会话记忆表 (Conversational Memory)
  -- =====================================================

  -- 会话表
  CREATE TABLE IF NOT EXISTS conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT UNIQUE NOT NULL,
    agent_id TEXT,
    user_id TEXT,
    title TEXT,
    context_data TEXT NOT NULL,
    preferences TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME
  );

  -- 消息历史表
  CREATE TABLE IF NOT EXISTS conversation_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    message_id TEXT UNIQUE NOT NULL,
    parent_message_id TEXT,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool')),
    content TEXT NOT NULL,
    metadata TEXT,
    tokens_used INTEGER,
    model_version TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES conversations(session_id) ON DELETE CASCADE
  );

  -- =====================================================
  -- 索引创建
  -- =====================================================

  -- 工作记忆索引
  CREATE INDEX IF NOT EXISTS idx_task_context_task_id ON task_context(task_id);
  CREATE INDEX IF NOT EXISTS idx_task_context_updated_at ON task_context(updated_at);

  -- 长期记忆索引
  CREATE INDEX IF NOT EXISTS idx_task_history_status ON task_history(status);
  CREATE INDEX IF NOT EXISTS idx_task_history_group_name ON task_history(group_name);
  CREATE INDEX IF NOT EXISTS idx_task_history_completed_at ON task_history(completed_at);
  CREATE INDEX IF NOT EXISTS idx_task_history_worker_id ON task_history(worker_id);
  CREATE INDEX IF NOT EXISTS idx_task_history_agent_id ON task_history(agent_id);

  CREATE INDEX IF NOT EXISTS idx_experience_summaries_type ON experience_summaries(summary_type);
  CREATE INDEX IF NOT EXISTS idx_experience_summaries_confidence ON experience_summaries(confidence_score);
  CREATE INDEX IF NOT EXISTS idx_experience_summaries_usage_count ON experience_summaries(usage_count);
  CREATE INDEX IF NOT EXISTS idx_experience_summaries_expires_at ON experience_summaries(expires_at);

  CREATE INDEX IF NOT EXISTS idx_skill_usage_stats_skill_name ON skill_usage_stats(skill_name);
  CREATE INDEX IF NOT EXISTS idx_skill_usage_stats_last_used ON skill_usage_stats(last_used_at);

  -- 会话记忆索引
  CREATE INDEX IF NOT EXISTS idx_conversations_session_id ON conversations(session_id);
  CREATE INDEX IF NOT EXISTS idx_conversations_agent_id ON conversations(agent_id);
  CREATE INDEX IF NOT EXISTS idx_conversations_expires_at ON conversations(expires_at);

  CREATE INDEX IF NOT EXISTS idx_conversation_messages_session_id ON conversation_messages(session_id);
  CREATE INDEX IF NOT EXISTS idx_conversation_messages_parent_id ON conversation_messages(parent_message_id);
  CREATE INDEX IF NOT EXISTS idx_conversation_messages_created_at ON conversation_messages(created_at);

  -- =====================================================
  -- 视图 - 便于查询
  -- =====================================================

  -- 任务摘要视图
  CREATE VIEW IF NOT EXISTS v_task_summary AS
  SELECT
    t.id AS task_id,
    t.task_id AS task_display_id,
    t.title,
    t.status,
    t.priority,
    COUNT(ch.id) AS checkpoint_count,
    tc.context_data,
    tc.created_at,
    t.completed_at
  FROM tasks t
  LEFT JOIN task_context tc ON t.id = tc.task_id
  LEFT JOIN task_checkpoints ch ON t.id = ch.task_id
  GROUP BY t.id;

  -- 活跃会话视图
  CREATE VIEW IF NOT EXISTS v_active_conversations AS
  SELECT
    c.session_id,
    c.agent_id,
    c.title,
    COUNT(cm.id) AS message_count,
    c.updated_at,
    c.expires_at
  FROM conversations c
  LEFT JOIN conversation_messages cm ON c.session_id = cm.session_id
  WHERE c.expires_at > datetime('now') OR c.expires_at IS NULL
  GROUP BY c.session_id
  ORDER BY c.updated_at DESC;

  -- 技能性能统计视图
  CREATE VIEW IF NOT EXISTS v_skill_performance AS
  SELECT
    skill_name,
    SUM(execution_count) AS total_executions,
    SUM(success_count) AS total_success,
    SUM(failure_count) AS total_failures,
    AVG(avg_duration_ms) AS avg_duration_ms,
    MAX(last_used_at) AS last_used_at
  FROM skill_usage_stats
  GROUP BY skill_name
  ORDER BY total_executions DESC;
`;

export const memoryTriggersSQL = `
  -- 任务完成时自动记录到历史
  CREATE TRIGGER IF NOT EXISTS after_task_completion
  AFTER UPDATE OF status ON tasks
  WHEN NEW.status IN ('completed', 'failed') AND OLD.status NOT IN ('completed', 'failed')
  BEGIN
    INSERT INTO task_history (
      task_id, title, description, group_name, status, priority,
      created_at, completed_at
    ) VALUES (
      NEW.id, NEW.title, NEW.description, NEW.group_name,
      NEW.status, NEW.priority, NEW.created_at, NEW.updated_at
    );
  END;
`;

export const memoryInitialDataSQL = `
  -- 插入一些默认的经验总结
  INSERT OR IGNORE INTO experience_summaries (
    summary_type, pattern_description, context, confidence_score
  ) VALUES
  ('best_practice', 'Always include clear task descriptions with specific goals and expected outcomes',
   '{"domain": "task_management"}', 0.9),
  ('optimization', 'Use parallel workers for independent tasks to reduce total execution time',
   '{"domain": "performance", "condition": "no_dependencies"}', 0.85);
`;
