//! 任务执行器
//!
//! 集成 Team A 的 TaskExecutor，提供单进程架构的任务执行能力

use crate::claude::ClaudeExecutor;
use agentflow_core::{Task, TaskPriority, TaskStatus};
use anyhow::{Context, Result};
use sqlx::{Pool, Row, Sqlite};
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{debug, error, info};

/// 任务执行器
///
/// 在单进程架构中，Master 直接使用此执行器运行任务，无需远程 Worker
pub struct TaskExecutor {
    /// 数据库连接池
    db: Pool<Sqlite>,
    /// 运行中的任务
    running_tasks: Arc<RwLock<Vec<i64>>>,
    /// 最大并发任务数
    pub max_concurrent_tasks: usize,
    /// Claude CLI 执行器
    claude: ClaudeExecutor,
}

impl Clone for TaskExecutor {
    fn clone(&self) -> Self {
        Self {
            db: self.db.clone(),
            running_tasks: self.running_tasks.clone(),
            max_concurrent_tasks: self.max_concurrent_tasks,
            claude: self.claude.clone(),
        }
    }
}

impl TaskExecutor {
    /// 创建新的任务执行器
    pub fn new(db: Pool<Sqlite>, max_concurrent_tasks: usize) -> Self {
        let claude = ClaudeExecutor::new().unwrap_or_else(|e| {
            info!("创建 ClaudeExecutor 失败: {}, 将使用模拟执行", e);
            ClaudeExecutor::default()
        });

        info!(
            "Claude CLI 可用: {}, 发现 {} 个 Skills",
            claude.is_available(),
            claude.count_available_skills()
        );

        Self {
            db,
            running_tasks: Arc::new(RwLock::new(Vec::new())),
            max_concurrent_tasks,
            claude,
        }
    }

    /// 执行任务（同步执行，返回结果）
    pub async fn execute_task(&self, task_id: i64) -> Result<String> {
        info!("开始执行任务: {}", task_id);

        // 检查并发限制
        {
            let running = self.running_tasks.read().await;
            if running.len() >= self.max_concurrent_tasks {
                return Err(anyhow::anyhow!("已达到最大并发任务数限制"));
            }
        }

        // 获取任务信息
        // 修复: 使用 runtime query 而不是 macro 以避免编译时数据库检查
        let row = sqlx::query(
            r#"
            SELECT
                id, task_id, parent_id, title, description, group_name,
                completion_criteria, status, priority,
                lock_holder, lock_time, result, error, workspace_dir,
                sandboxed, allow_network, max_memory, max_cpu,
                created_by, created_at, started_at, completed_at
            FROM tasks
            WHERE id = ?
            "#
        )
        .bind(task_id)
        .fetch_one(&self.db)
        .await
        .context("任务不存在")?;

        // 从字符串解析状态
        let status_str: String = row.get("status");
        let status = match status_str.as_str() {
            "pending" => TaskStatus::Pending,
            "running" => TaskStatus::Running,
            "completed" => TaskStatus::Completed,
            "failed" => TaskStatus::Failed,
            "blocked" => TaskStatus::Blocked,
            _ => TaskStatus::Pending,
        };

        // 从 i32 转换优先级
        let priority_i32: i32 = row.get("priority");
        let priority = TaskPriority::from_i32(priority_i32);

        let task = Task {
            id: row.get("id"),
            task_id: row.get("task_id"),
            parent_id: row.get("parent_id"),
            title: row.get("title"),
            description: row.get("description"),
            group_name: row.get("group_name"),
            completion_criteria: row.get("completion_criteria"),
            status,
            priority,
            lock_holder: row.get("lock_holder"),
            lock_time: row.get("lock_time"),
            result: row.get("result"),
            error: row.get("error"),
            workspace_dir: row.get("workspace_dir"),
            sandboxed: row.get::<i64, _>("sandboxed") != 0,
            allow_network: row.get::<i64, _>("allow_network") != 0,
            max_memory: row.get("max_memory"),
            max_cpu: row.get("max_cpu"),
            created_by: row.get("created_by"),
            created_at: row.get("created_at"),
            started_at: row.get("started_at"),
            completed_at: row.get("completed_at"),
        };

        // 标记任务为运行中
        self.mark_task_running(task_id).await?;

        // 添加到运行列表
        {
            let mut running = self.running_tasks.write().await;
            running.push(task_id);
        }

        // 执行任务逻辑
        let result = match self.do_execute(&task).await {
            Ok(result) => {
                self.mark_task_completed(task_id, Some(result.clone()), None).await?;
                info!("任务 {} 执行成功", task_id);
                Ok(result)
            }
            Err(e) => {
                let error_msg = e.to_string();
                self.mark_task_completed(task_id, None, Some(&error_msg)).await?;
                error!("任务 {} 执行失败: {}", task_id, error_msg);
                Err(e)
            }
        };

        // 从运行列表移除
        {
            let mut running = self.running_tasks.write().await;
            running.retain(|&id| id != task_id);
        }

        result
    }

    /// 实际执行任务逻辑
    async fn do_execute(&self, task: &Task) -> Result<String> {
        debug!("执行任务: {} - {}", task.task_id, task.title);

        // 构建任务描述
        let desc_str: &str = match task.description {
            Some(ref s) => s.as_str(),
            None => "无描述",
        };

        let description = format!(
            "任务: {}\n\n标题: {}\n描述: {}\n\n分组: {}\n\n请执行此任务并返回结果。",
            task.task_id,
            task.title,
            desc_str,
            &task.group_name
        );

        // 使用 Claude Executor 执行任务
        let result = self.claude.execute(&description).await?;

        debug!("任务执行结果: {}", result);

        Ok(result)
    }

    /// 标记任务为运行中
    async fn mark_task_running(&self, task_id: i64) -> Result<()> {
        let now = chrono::Utc::now();
        // 修复: 使用 runtime query
        sqlx::query(
            r#"
            UPDATE tasks
            SET status = ?,
                started_at = ?,
                lock_holder = ?,
                lock_time = ?
            WHERE id = ?
            "#
        )
        .bind(TaskStatus::Running.to_string())
        .bind(now)
        .bind("master")
        .bind(now)
        .bind(task_id)
        .execute(&self.db)
        .await?;

        Ok(())
    }

    /// 标记任务为完成
    async fn mark_task_completed(
        &self,
        task_id: i64,
        result: Option<String>,
        error: Option<&str>,
    ) -> Result<()> {
        let now = chrono::Utc::now();
        let status = if error.is_some() {
            TaskStatus::Failed
        } else {
            TaskStatus::Completed
        };

        // 修复: 使用 runtime query
        sqlx::query(
            r#"
            UPDATE tasks
            SET status = ?,
                completed_at = ?,
                result = ?,
                error = ?,
                lock_holder = NULL,
                lock_time = NULL
            WHERE id = ?
            "#
        )
        .bind(status.to_string())
        .bind(now)
        .bind(result)
        .bind(error)
        .bind(task_id)
        .execute(&self.db)
        .await?;

        Ok(())
    }

    /// 获取运行中的任务列表
    pub async fn get_running_tasks(&self) -> Vec<i64> {
        self.running_tasks.read().await.clone()
    }

    /// 检查任务是否在运行中
    pub async fn is_task_running(&self, task_id: i64) -> bool {
        let running = self.running_tasks.read().await;
        running.contains(&task_id)
    }

    /// 获取数据库连接池的引用
    /// 修复: 添加公共 getter 以便其他模块可以访问数据库
    pub fn db(&self) -> &Pool<Sqlite> {
        &self.db
    }
}
