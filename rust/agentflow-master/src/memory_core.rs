//! 记忆核心
//!
//! 集成 Team B 的 MemoryCore，提供任务执行过程中的记忆存储和检索

use agentflow_core::{MemoryCategory, MemoryEntry};
use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use serde_json::Value;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{debug, info};

/// 记忆核心
///
/// 提供内存中的键值存储，支持过期时间和分类
pub struct MemoryCore {
    /// 内存存储
    storage: Arc<RwLock<HashMap<String, (MemoryEntry, DateTime<Utc>)>>>,
    /// 默认 TTL（秒）
    default_ttl: i64,
    /// 最大条目数
    max_entries: usize,
}

impl Clone for MemoryCore {
    fn clone(&self) -> Self {
        Self {
            storage: self.storage.clone(),
            default_ttl: self.default_ttl,
            max_entries: self.max_entries,
        }
    }
}

impl MemoryCore {
    /// 创建新的记忆核心
    pub fn new(default_ttl: i64, max_entries: usize) -> Self {
        Self {
            storage: Arc::new(RwLock::new(HashMap::new())),
            default_ttl,
            max_entries,
        }
    }

    /// 设置记忆条目
    pub async fn set(
        &self,
        key: String,
        value: Value,
        category: MemoryCategory,
        task_id: Option<String>,
        ttl: Option<i64>,
    ) -> Result<()> {
        let ttl = ttl.unwrap_or(self.default_ttl);
        let expires_at = Utc::now().timestamp() + ttl;
        let timestamp = Utc::now().timestamp();

        let entry = MemoryEntry {
            key: key.clone(),
            value,
            expires_at: Some(expires_at),
            category,
            task_id,
            timestamp,
        };

        let mut storage = self.storage.write().await;

        // 检查是否超过最大条目数
        if storage.len() >= self.max_entries && !storage.contains_key(&key) {
            // 清理过期条目
            self.cleanup_expired_internal(&mut storage).await;
        }

        storage.insert(key.clone(), (entry, Utc::now() + chrono::Duration::seconds(ttl)));

        debug!("设置记忆条目: {}, 分类: {:?}", key, category);

        Ok(())
    }

    /// 获取记忆条目
    pub async fn get(&self, key: &str) -> Result<Option<MemoryEntry>> {
        let mut storage = self.storage.write().await;

        if let Some((entry, expires_at)) = storage.get(key) {
            // 检查是否过期
            if *expires_at > Utc::now() {
                return Ok(Some(entry.clone()));
            } else {
                // 过期，删除
                storage.remove(key);
            }
        }

        Ok(None)
    }

    /// 删除记忆条目
    pub async fn delete(&self, key: &str) -> Result<bool> {
        let mut storage = self.storage.write().await;
        Ok(storage.remove(key).is_some())
    }

    /// 搜索记忆条目
    pub async fn search(
        &self,
        query: &str,
        category: Option<MemoryCategory>,
        task_id: Option<String>,
        limit: Option<usize>,
    ) -> Result<Vec<MemoryEntry>> {
        let storage = self.storage.read().await;
        let limit = limit.unwrap_or(100);

        let mut results = Vec::new();

        for (key, (entry, expires_at)) in storage.iter() {
            // 检查是否过期
            if *expires_at <= Utc::now() {
                continue;
            }

            // 检查分类过滤
            if let Some(cat) = category {
                if entry.category != cat {
                    continue;
                }
            }

            // 检查任务过滤
            if let Some(ref tid) = task_id {
                if entry.task_id.as_ref() != Some(tid) {
                    continue;
                }
            }

            // 检查查询匹配
            if !query.is_empty() {
                let key_matches = key.contains(query);
                let value_matches = serde_json::to_string(&entry.value)
                    .unwrap_or_default()
                    .contains(query);

                if !key_matches && !value_matches {
                    continue;
                }
            }

            results.push(entry.clone());

            if results.len() >= limit {
                break;
            }
        }

        debug!("搜索记忆: 查询='{}', 结果数={}", query, results.len());

        Ok(results)
    }

    /// 获取任务相关的所有记忆
    pub async fn get_task_memories(&self, task_id: &str) -> Result<Vec<MemoryEntry>> {
        self.search("", None, Some(task_id.to_string()), None).await
    }

    /// 清理过期的记忆条目
    pub async fn cleanup_expired(&self) -> Result<usize> {
        let mut storage = self.storage.write().await;
        Ok(self.cleanup_expired_internal(&mut storage).await)
    }

    /// 内部清理过期条目
    async fn cleanup_expired_internal(
        &self,
        storage: &mut HashMap<String, (MemoryEntry, DateTime<Utc>)>,
    ) -> usize {
        let now = Utc::now();
        let before = storage.len();

        storage.retain(|_, (_, expires_at)| *expires_at > now);

        let cleaned = before - storage.len();
        if cleaned > 0 {
            debug!("清理了 {} 个过期记忆条目", cleaned);
        }

        cleaned
    }

    /// 获取统计信息
    pub async fn stats(&self) -> Result<MemoryStats> {
        let storage = self.storage.read().await;
        let now = Utc::now();

        let total = storage.len();
        let expired = storage.values().filter(|(_, expires_at)| *expires_at <= now).count();
        let active = total - expired;

        let mut category_counts: HashMap<MemoryCategory, usize> = HashMap::new();

        for (entry, _) in storage.values() {
            *category_counts.entry(entry.category).or_insert(0) += 1;
        }

        Ok(MemoryStats {
            total,
            active,
            expired,
            category_counts,
        })
    }

    /// 清空所有记忆
    pub async fn clear(&self) -> Result<()> {
        let mut storage = self.storage.write().await;
        storage.clear();
        info!("已清空所有记忆条目");
        Ok(())
    }

    /// 创建记忆快照
    pub async fn create_snapshot(&self, worker_id: String) -> Result<MemorySnapshot> {
        let storage = self.storage.read().await;
        let now = Utc::now();

        let entries: Vec<MemoryEntry> = storage
            .values()
            .filter(|(_, expires_at)| *expires_at > now)
            .map(|(entry, _)| entry.clone())
            .collect();

        Ok(MemorySnapshot {
            entries,
            worker_id,
            snapshot_time: now,
        })
    }
}

/// 记忆统计信息
#[derive(Debug, Clone)]
pub struct MemoryStats {
    pub total: usize,
    pub active: usize,
    pub expired: usize,
    pub category_counts: HashMap<MemoryCategory, usize>,
}

/// 记忆快照
#[derive(Debug, Clone)]
pub struct MemorySnapshot {
    pub entries: Vec<MemoryEntry>,
    pub worker_id: String,
    pub snapshot_time: DateTime<Utc>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_memory_set_get() {
        let memory = MemoryCore::new(3600, 1000);

        memory
            .set(
                "test_key".to_string(),
                serde_json::json!({"value": "test"}),
                MemoryCategory::Execution,
                Some("task_123".to_string()),
                None,
            )
            .await
            .unwrap();

        let entry = memory.get("test_key").await.unwrap().unwrap();
        assert_eq!(entry.key, "test_key");
    }

    #[tokio::test]
    async fn test_memory_search() {
        let memory = MemoryCore::new(3600, 1000);

        memory
            .set(
                "key1".to_string(),
                serde_json::json!({"data": "test_value"}),
                MemoryCategory::Context,
                Some("task_123".to_string()),
                None,
            )
            .await
            .unwrap();

        let results = memory
            .search("test_value", None, None, None)
            .await
            .unwrap();
        assert_eq!(results.len(), 1);
    }
}
