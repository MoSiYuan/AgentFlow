//! 记忆核心系统
//!
//! 实现基于 SQLite 的记忆索引和检索功能
//! 注意：不使用外部嵌入模型，而是依赖 Claude CLI 的语义理解能力

pub mod persistent;
pub mod worker;

pub use worker::{WorkerMemory, WorkerMemoryConfig, WorkerMemoryStats};
pub use persistent::{PersistentMemory, TaskContext};

use crate::types::{MemoryCategory, MemoryEntry};
use anyhow::{Context, Result};
use sqlx::{Row, SqlitePool, sqlite::SqliteConnectOptions};
use std::path::Path;
use std::str;
use tracing::{debug, info, warn};

/// 记忆核心系统
///
/// 提供基于 SQLite 的记忆索引和关键词检索功能
/// 依赖 Claude CLI 的语义理解能力，无需外部嵌入模型
pub struct MemoryCore {
    /// SQLite 连接池
    pool: SqlitePool,
}

impl MemoryCore {
    /// 创建新的记忆核心实例
    ///
    /// # 参数
    /// * `db_path` - SQLite 数据库文件路径
    ///
    /// # 返回
    /// 返回初始化后的 MemoryCore 实例
    ///
    /// # 示例
    /// ```no_run
    /// use agentflow_core::memory::MemoryCore;
    ///
    /// # async fn example() -> anyhow::Result<()> {
    /// let memory = MemoryCore::new("/path/to/memory.db").await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn new(db_path: &str) -> Result<Self> {
        info!("初始化记忆核心系统，数据库路径: {}", db_path);

        // 确保数据库目录存在
        if let Some(parent) = Path::new(db_path).parent() {
            tokio::fs::create_dir_all(parent).await
                .context("创建数据库目录失败")?;
        }

        // 配置 SQLite 连接选项
        let options = SqliteConnectOptions::new()
            .filename(db_path)
            .create_if_missing(true);

        // 创建连接池
        let pool = SqlitePool::connect_with(options).await
            .context("连接 SQLite 数据库失败")?;

        let memory_core = Self { pool };

        // 初始化数据库表结构
        memory_core.init_schema().await?;

        info!("记忆核心系统初始化完成");
        Ok(memory_core)
    }

    /// 初始化数据库表结构
    ///
    /// 创建记忆索引所需的表和索引
    async fn init_schema(&self) -> Result<()> {
        debug!("初始化数据库表结构");

        // 创建记忆表
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS memories (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                key TEXT NOT NULL UNIQUE,
                value TEXT NOT NULL,
                category TEXT NOT NULL,
                task_id TEXT,
                timestamp INTEGER NOT NULL,
                expires_at INTEGER,
                embedding BLOB,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
            "#,
        )
        .execute(&self.pool)
        .await
        .context("创建 memories 表失败")?;

        // 创建索引以提高查询性能
        sqlx::query(
            r#"
            CREATE INDEX IF NOT EXISTS idx_memories_key
            ON memories(key)
            "#,
        )
        .execute(&self.pool)
        .await
        .context("创建 key 索引失败")?;

        sqlx::query(
            r#"
            CREATE INDEX IF NOT EXISTS idx_memories_category
            ON memories(category)
            "#,
        )
        .execute(&self.pool)
        .await
        .context("创建 category 索引失败")?;

        sqlx::query(
            r#"
            CREATE INDEX IF NOT EXISTS idx_memories_task_id
            ON memories(task_id)
            "#,
        )
        .execute(&self.pool)
        .await
        .context("创建 task_id 索引失败")?;

        sqlx::query(
            r#"
            CREATE INDEX IF NOT EXISTS idx_memories_timestamp
            ON memories(timestamp DESC)
            "#,
        )
        .execute(&self.pool)
        .await
        .context("创建 timestamp 索引失败")?;

        debug!("数据库表结构初始化完成");
        Ok(())
    }

    /// 索引新的记忆条目
    ///
    /// 将记忆条目存储到数据库中
    ///
    /// # 参数
    /// * `entry` - 要索引的记忆条目
    ///
    /// # 返回
    /// 成功返回 Ok(())，失败返回错误信息
    ///
    /// # 示例
    /// ```no_run
    /// # use agentflow_core::memory::MemoryCore;
    /// # use agentflow_core::types::{MemoryEntry, MemoryCategory};
    /// # use serde_json::json;
    /// #
    /// # async fn example() -> anyhow::Result<()> {
    /// # let memory = MemoryCore::new("/path/to/memory.db").await?;
    /// let entry = MemoryEntry {
    ///     key: "task_123_result".to_string(),
    ///     value: json!("Task completed successfully"),
    ///     expires_at: None,
    ///     category: MemoryCategory::Result,
    ///     task_id: Some("task_123".to_string()),
    ///     timestamp: chrono::Utc::now().timestamp(),
    /// };
    ///
    /// memory.index(&entry).await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn index(&self, entry: &MemoryEntry) -> Result<()> {
        debug!("索引记忆条目: key={}, category={:?}", entry.key, entry.category);

        // 将 value 序列化为 JSON 字符串
        let value_str = serde_json::to_string(&entry.value)
            .context("序列化记忆值失败")?;

        // 将 category 转换为字符串
        let category_str = format!("{:?}", entry.category);

        // 插入或更新记忆条目
        // 注意：embedding 字段保留但暂不使用（未来可能用于真正的向量检索）
        sqlx::query(
            r#"
            INSERT INTO memories (key, value, category, task_id, timestamp, expires_at, embedding)
            VALUES (?1, ?2, ?3, ?4, ?5, ?6, NULL)
            ON CONFLICT(key) DO UPDATE SET
                value = ?2,
                category = ?3,
                task_id = ?4,
                timestamp = ?5,
                expires_at = ?6
            "#,
        )
        .bind(&entry.key)
        .bind(&value_str)
        .bind(&category_str)
        .bind(&entry.task_id)
        .bind(entry.timestamp)
        .bind(entry.expires_at)
        .execute(&self.pool)
        .await
        .context("插入记忆条目失败")?;

        debug!("记忆条目索引成功: {}", entry.key);
        Ok(())
    }

    /// 关键词检索
    ///
    /// 根据查询文本检索相关的记忆条目（使用 SQL LIKE 匹配）
    ///
    /// # 参数
    /// * `query` - 查询文本
    /// * `limit` - 返回的最大结果数量
    ///
    /// # 返回
    /// 返回按时间戳排序的记忆条目列表
    ///
    /// # 示例
    /// ```no_run
    /// # use agentflow_core::memory::MemoryCore;
    /// #
    /// # async fn example() -> anyhow::Result<()> {
    /// # let memory = MemoryCore::new("/path/to/memory.db").await?;
    /// let results = memory.search("task execution error", 10).await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn search(&self, query: &str, limit: usize) -> Result<Vec<MemoryEntry>> {
        debug!("执行关键词检索: query={}, limit={}", query, limit);

        // 使用 SQL LIKE 进行关键词匹配
        let pattern = format!("%{}%", query);

        let rows = sqlx::query(
            r#"
            SELECT key, value, category, task_id, timestamp, expires_at
            FROM memories
            WHERE (value LIKE ?1 OR key LIKE ?1)
              AND (expires_at IS NULL OR expires_at > ?2)
            ORDER BY timestamp DESC
            LIMIT ?3
            "#
        )
        .bind(&pattern)
        .bind(chrono::Utc::now().timestamp())
        .bind(limit as i64)
        .fetch_all(&self.pool)
        .await
        .context("查询记忆条目失败")?;

        // 解析结果
        let mut results = Vec::new();

        for row in rows {
            let key: String = row.get("key");
            let value_str: String = row.get("value");
            let category_str: String = row.get("category");
            let task_id: Option<String> = row.get("task_id");
            let timestamp: i64 = row.get("timestamp");
            let expires_at: Option<i64> = row.get("expires_at");

            // 解析 value
            let value: serde_json::Value = serde_json::from_str(&value_str)
                .unwrap_or_else(|_| serde_json::Value::String(value_str.clone()));

            // 解析 category
            let category = match category_str.as_str() {
                "Execution" => MemoryCategory::Execution,
                "Context" => MemoryCategory::Context,
                "Result" => MemoryCategory::Result,
                "Error" => MemoryCategory::Error,
                "Checkpoint" => MemoryCategory::Checkpoint,
                _ => MemoryCategory::Context,
            };

            let entry = MemoryEntry {
                key,
                value,
                expires_at,
                category,
                task_id,
                timestamp,
            };

            results.push(entry);
        }

        debug!("关键词检索完成，返回 {} 条结果", results.len());
        Ok(results)
    }

    /// 根据键获取记忆条目
    ///
    /// # 参数
    /// * `key` - 记忆条目的键
    ///
    /// # 返回
    /// 返回找到的记忆条目，如果不存在则返回 None
    pub async fn get(&self, key: &str) -> Result<Option<MemoryEntry>> {
        debug!("获取记忆条目: key={}", key);

        let row = sqlx::query(
            r#"
            SELECT key, value, category, task_id, timestamp, expires_at
            FROM memories
            WHERE key = ?1
              AND (expires_at IS NULL OR expires_at > ?2)
            "#,
        )
        .bind(key)
        .bind(chrono::Utc::now().timestamp())
        .fetch_optional(&self.pool)
        .await
        .context("查询记忆条目失败")?;

        if let Some(row) = row {
            let key: String = row.get("key");
            let value_str: String = row.get("value");
            let category_str: String = row.get("category");
            let task_id: Option<String> = row.get("task_id");
            let timestamp: i64 = row.get("timestamp");
            let expires_at: Option<i64> = row.get("expires_at");

            let value: serde_json::Value = serde_json::from_str(&value_str)
                .unwrap_or_else(|_| serde_json::Value::String(value_str.clone()));

            let category = match category_str.as_str() {
                "Execution" => MemoryCategory::Execution,
                "Context" => MemoryCategory::Context,
                "Result" => MemoryCategory::Result,
                "Error" => MemoryCategory::Error,
                "Checkpoint" => MemoryCategory::Checkpoint,
                _ => MemoryCategory::Context,
            };

            Ok(Some(MemoryEntry {
                key,
                value,
                expires_at,
                category,
                task_id,
                timestamp,
            }))
        } else {
            Ok(None)
        }
    }

    /// 删除记忆条目
    ///
    /// # 参数
    /// * `key` - 要删除的记忆条目的键
    ///
    /// # 返回
    /// 成功返回 Ok(())，失败返回错误信息
    pub async fn delete(&self, key: &str) -> Result<()> {
        debug!("删除记忆条目: key={}", key);

        sqlx::query("DELETE FROM memories WHERE key = ?1")
            .bind(key)
            .execute(&self.pool)
            .await
            .context("删除记忆条目失败")?;

        debug!("记忆条目删除成功: {}", key);
        Ok(())
    }

    /// 清理过期的记忆条目
    ///
    /// # 返回
    /// 返回删除的条目数量
    pub async fn cleanup_expired(&self) -> Result<u64> {
        debug!("清理过期的记忆条目");

        let now = chrono::Utc::now().timestamp();

        let result = sqlx::query("DELETE FROM memories WHERE expires_at IS NOT NULL AND expires_at < ?1")
            .bind(now)
            .execute(&self.pool)
            .await
            .context("清理过期记忆条目失败")?;

        let count = result.rows_affected();
        debug!("清理了 {} 条过期记忆条目", count);
        Ok(count)
    }

    /// 计算文本嵌入向量（已弃用）
    ///
    /// 注意：此方法已弃用，我们不再使用向量嵌入
    /// 而是依赖 Claude CLI 的语义理解能力和关键词匹配
    #[deprecated(note = "使用关键词匹配替代向量嵌入")]
    async fn compute_embedding(&self, _text: &str) -> Result<Vec<u8>> {
        // 返回空向量，表示不使用嵌入
        Ok(Vec::new())
    }

    /// 获取记忆统计信息
    ///
    /// # 返回
    /// 返回记忆库的统计信息
    pub async fn stats(&self) -> Result<MemoryStats> {
        debug!("获取记忆统计信息");

        let total_count: i64 = sqlx::query("SELECT COUNT(*) as count FROM memories")
            .fetch_one(&self.pool)
            .await
            .context("查询记忆总数失败")?
            .get("count");

        let expired_count: i64 = sqlx::query(
            "SELECT COUNT(*) as count FROM memories WHERE expires_at IS NOT NULL AND expires_at < ?1"
        )
        .bind(chrono::Utc::now().timestamp())
        .fetch_one(&self.pool)
        .await
        .context("查询过期记忆数失败")?
        .get("count");

        Ok(MemoryStats {
            total_entries: total_count as usize,
            expired_entries: expired_count as usize,
            active_entries: (total_count - expired_count) as usize,
        })
    }
}

/// 记忆统计信息
#[derive(Debug, Clone)]
pub struct MemoryStats {
    /// 总记忆条目数
    pub total_entries: usize,
    /// 过期条目数
    pub expired_entries: usize,
    /// 有效条目数
    pub active_entries: usize,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::MemoryCategory;
    use serde_json::json;

    #[tokio::test]
    async fn test_memory_core() -> Result<()> {
        let memory = MemoryCore::new(":memory:").await?;

        // 测试索引
        let entry = MemoryEntry {
            key: "test_key".to_string(),
            value: json!("test_value"),
            expires_at: None,
            category: MemoryCategory::Result,
            task_id: Some("task_123".to_string()),
            timestamp: chrono::Utc::now().timestamp(),
        };

        memory.index(&entry).await?;

        // 测试获取
        let retrieved = memory.get("test_key").await?;
        assert!(retrieved.is_some());
        assert_eq!(retrieved.unwrap().key, "test_key");

        // 测试搜索
        let results = memory.search("test", 10).await?;
        assert!(!results.is_empty());

        // 测试删除
        memory.delete("test_key").await?;
        let deleted = memory.get("test_key").await?;
        assert!(deleted.is_none());

        Ok(())
    }
}
