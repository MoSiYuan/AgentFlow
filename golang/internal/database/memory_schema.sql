-- =====================================================
-- AgentFlow Memory System Database Schema
-- =====================================================
-- 这个文件定义了 AgentFlow 记忆系统的数据库表结构
-- 支持 MySQL 8.0+
-- =====================================================

-- =====================================================
-- 1. 工作记忆表 (Working Memory)
-- =====================================================

-- 任务上下文表（存储任务执行的中间状态）
CREATE TABLE IF NOT EXISTS task_context (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id INT UNIQUE NOT NULL,
    context_data JSON NOT NULL COMMENT '任务上下文数据',
    variables JSON COMMENT '任务变量',
    dependencies JSON COMMENT '依赖追踪',
    execution_history JSON COMMENT '执行历史',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    INDEX idx_task_id (task_id),
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务上下文表';

-- =====================================================
-- 2. 长期记忆表 (Long-term Memory)
-- =====================================================

-- 任务历史表（用于分析和学习）
CREATE TABLE IF NOT EXISTS task_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id INT NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    group_name VARCHAR(100),
    status VARCHAR(50),
    priority VARCHAR(20),
    result TEXT COMMENT '执行结果摘要',
    duration_ms INT COMMENT '执行时长(毫秒)',
    worker_id VARCHAR(255),
    agent_id VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    metadata JSON COMMENT '额外元数据',
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL,
    INDEX idx_status (status),
    INDEX idx_group_name (group_name),
    INDEX idx_completed_at (completed_at),
    INDEX idx_worker_id (worker_id),
    INDEX idx_agent_id (agent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务历史表';

-- 经验总结表（从执行历史中提取的知识）
CREATE TABLE IF NOT EXISTS experience_summaries (
    id INT AUTO_INCREMENT PRIMARY KEY,
    summary_type ENUM('success_pattern', 'failure_pattern', 'optimization', 'best_practice') NOT NULL,
    pattern_description TEXT NOT NULL COMMENT '模式描述',
    context JSON COMMENT '适用场景',
    confidence_score REAL DEFAULT 0.5 COMMENT '置信度(0-1)',
    usage_count INT DEFAULT 0 COMMENT '使用次数',
    success_count INT DEFAULT 0 COMMENT '成功次数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    INDEX idx_type (summary_type),
    INDEX idx_confidence (confidence_score),
    INDEX idx_usage_count (usage_count),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='经验总结表';

-- 技能使用统计表
CREATE TABLE IF NOT EXISTS skill_usage_stats (
    id INT AUTO_INCREMENT PRIMARY KEY,
    skill_name VARCHAR(255) NOT NULL,
    task_id INT,
    agent_id VARCHAR(255),
    execution_count INT DEFAULT 1 COMMENT '执行次数',
    success_count INT DEFAULT 0 COMMENT '成功次数',
    failure_count INT DEFAULT 0 COMMENT '失败次数',
    avg_duration_ms INT COMMENT '平均执行时长(毫秒)',
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_skill_task_agent (skill_name, task_id, agent_id),
    INDEX idx_skill_name (skill_name),
    INDEX idx_last_used (last_used_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技能使用统计表';

-- =====================================================
-- 3. 会话记忆表 (Conversational Memory)
-- =====================================================

-- 会话表
CREATE TABLE IF NOT EXISTS conversations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    agent_id VARCHAR(255),
    user_id VARCHAR(255),
    title VARCHAR(500),
    context_data JSON NOT NULL COMMENT '会话上下文',
    preferences JSON COMMENT '用户偏好',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL COMMENT '会话过期时间',
    INDEX idx_session_id (session_id),
    INDEX_idx_agent_id (agent_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- 消息历史表
CREATE TABLE IF NOT EXISTS conversation_messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    message_id VARCHAR(255) UNIQUE NOT NULL,
    parent_message_id VARCHAR(255) COMMENT '父消息ID',
    role ENUM('user', 'assistant', 'system', 'tool') NOT NULL,
    content TEXT NOT NULL,
    metadata JSON COMMENT '额外信息',
    tokens_used INT COMMENT '使用的token数',
    model_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES conversations(session_id) ON DELETE CASCADE,
    INDEX idx_session_id (session_id),
    INDEX idx_parent_id (parent_message_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息历史表';

-- =====================================================
-- 4. 向量存储表 (用于语义检索 - 可选)
-- =====================================================

-- 记忆嵌入表（存储语义向量）
-- 注意：如果使用 sqlite-vec 扩展，需要安装该扩展
CREATE TABLE IF NOT EXISTS memory_embeddings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    memory_type ENUM('task', 'conversation', 'experience', 'skill', 'agent') NOT NULL,
    memory_id INT NOT NULL,
    embedding BLOB NOT NULL COMMENT '向量数据(二进制)',
    embedding_model VARCHAR(100) NOT NULL COMMENT '使用的嵌入模型',
    dimension INT NOT NULL COMMENT '向量维度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_memory_type (memory_type),
    INDEX idx_memory_id (memory_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='记忆嵌入表';

-- =====================================================
-- 5. 维护和清理任务
-- =====================================================

-- 创建存储过程：清理过期数据
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS CleanupExpiredMemories()
BEGIN
    -- 删除过期的检查点（超过 7 天）
    DELETE FROM checkpoints
    WHERE expires_at < NOW() - INTERVAL 7 DAY;

    -- 删除过期的会话（超过 30 天未活动）
    DELETE FROM conversations
    WHERE expires_at < NOW()
    AND updated_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

    -- 删除过期的经验总结（低置信度且超过 90 天未使用）
    DELETE FROM experience_summaries
    WHERE confidence_score < 0.3
    AND (last_used_at < DATE_SUB(NOW(), INTERVAL 90 DAY) OR last_used_at IS NULL);

    -- 优化表
    OPTIMIZE TABLE task_context;
    OPTIMIZE TABLE task_history;
    OPTIMIZE TABLE conversations;
    OPTIMIZE TABLE experience_summaries;

    SELECT CONCAT('Cleanup completed at ', NOW()) AS result;
END //

DELIMITER ;

-- 创建事件：每天凌晨 2 点执行清理
SET GLOBAL event_scheduler = ON;

CREATE EVENT IF NOT EXISTS cleanup_memory_event
ON SCHEDULE EVERY 1 DAY
STARTS (TIMESTAMP(CURRENT_DATE) + INTERVAL 1 DAY + INTERVAL 2 HOUR)
DO CALL CleanupExpiredMemories();

-- =====================================================
-- 6. 视图 - 便于查询
-- =====================================================

-- 任务摘要视图
CREATE OR REPLACE VIEW v_task_summary AS
SELECT
    t.id AS task_id,
    t.task_id AS task_display_id,
    t.title,
    t.status,
    t.priority,
    COUNT(ch.id) AS checkpoint_count,
    tc.context_data->>'$.variables' AS task_variables,
    tc.context_data->>'$.dependencies' AS task_dependencies,
    t.created_at,
    t.completed_at
FROM tasks t
LEFT JOIN task_context tc ON t.id = tc.task_id
LEFT JOIN checkpoints ch ON t.id = ch.task_id
GROUP BY t.id;

-- 活跃会话视图
CREATE OR REPLACE VIEW v_active_conversations AS
SELECT
    c.session_id,
    c.agent_id,
    c.title,
    COUNT(cm.id) AS message_count,
    c.updated_at,
    c.expires_at
FROM conversations c
LEFT JOIN conversation_messages cm ON c.session_id = cm.session_id
WHERE c.expires_at > NOW() OR c.expires_at IS NULL
GROUP BY c.session_id
ORDER BY c.updated_at DESC;

-- 技能性能统计视图
CREATE OR REPLACE VIEW v_skill_performance AS
SELECT
    skill_name,
    SUM(execution_count) AS total_executions,
    SUM(success_count) AS total_success,
    SUM(failure_count) AS total_failures,
    AVG(avg_duration_ms) AS avg_duration_ms,
    MAX(last_used_at) AS last_used_at,
    CASE
        WHEN AVG(avg_duration_ms) IS NULL THEN 0
        ELSE AVG(avg_duration_ms)
    END AS avg_duration_ms_calc
FROM skill_usage_stats
GROUP BY skill_name
ORDER BY total_executions DESC;

-- =====================================================
-- 7. 触发器 - 自动维护
-- =====================================================

-- 任务完成时自动记录到历史
DELIMITER //

CREATE TRIGGER IF NOT EXISTS after_task_completion
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
    -- 当任务状态变为 completed 或 failed 时，记录到历史
    IF NEW.status IN ('completed', 'failed') AND OLD.status NOT IN ('completed', 'failed') THEN
        INSERT INTO task_history (
            task_id, title, description, group_name,
            status, priority, created_at, completed_at
        ) VALUES (
            NEW.id, NEW.title, NEW.description, NEW.group_name,
            NEW.status, NEW.priority, NEW.created_at, NEW.updated_at
        );
    END IF;
END //

DELIMITER ;

-- =====================================================
-- 8. 索引优化建议
-- =====================================================
--
-- 对于频繁查询的场景，建议添加以下复合索引：
--
-- ALTER TABLE task_history ADD INDEX idx_status_completed (status, completed_at);
-- ALTER TABLE skill_usage_stats ADD INDEX idx_skill_last_used (skill_name, last_used_at);
-- ALTER TABLE conversation_messages ADD INDEX idx_session_created (session_id, created_at);
--
-- 对于大数据量场景，考虑分区表：
-- ALTER TABLE task_history PARTITION BY RANGE (YEAR(created_at)) (...)
--

-- =====================================================
-- 9. 初始化数据
-- =====================================================

-- 插入一些默认的经验总结
INSERT INTO experience_summaries (
    summary_type, pattern_description, context, confidence_score
) VALUES
('best_practice', 'Always include clear task descriptions with specific goals and expected outcomes',
 '{"domain": "task_management"}', 0.9),
('optimization', 'Use parallel workers for independent tasks to reduce total execution time',
 '{"domain": "performance", "condition": "no_dependencies"}', 0.85);

-- =====================================================
-- 10. 权限设置（如果使用）
-- =====================================================
--
-- GRANT SELECT, INSERT, UPDATE, DELETE ON agentflow.* TO 'agentflow_app'@'%';
-- GRANT SELECT, EXECUTE ON PROCEDURE agentflow.CleanupExpiredMemories TO 'agentflow_app'@'%';
--
