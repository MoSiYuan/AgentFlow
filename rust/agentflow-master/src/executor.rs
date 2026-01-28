//! 任务执行器
//!
//! 集成 Team A 的 TaskExecutor，提供单进程架构的任务执行能力

use agentflow_core::{Task, TaskStatus};
use anyhow::{Context, Result};
use sqlx::{Pool, Sqlite};
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
    max_concurrent_tasks: usize,
}

impl Clone for TaskExecutor {
    fn clone(&self) -> Self {
        Self {
            db: self.db.clone(),
            running_tasks: self.running_tasks.clone(),
            max_concurrent_tasks: self.max_concurrent_tasks,
        }
    }
}

impl TaskExecutor {
    /// 创建新的任务执行器
    pub fn new(db: Pool<Sqlite>, max_concurrent_tasks: usize) -> Self {
        Self {
            db,
            running_tasks: Arc::new(RwLock::new(Vec::new())),
            max_concurrent_tasks,
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
        let task = sqlx::query_as!(
            Task,
            r#"
            SELECT
                id, task_id, parent_id, title, description, group_name,
                completion_criteria, status as "status: TaskStatus", priority as "priority: TaskPriority",
                lock_holder, lock_time, result, error, workspace_dir,
                sandboxed, allow_network, max_memory, max_cpu,
                created_by, created_at, started_at, completed_at
            FROM tasks
            WHERE id = ?
            "#,
            task_id
        )
        .fetch_one(&self.db)
        .await
        .context("任务不存在")?;

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

        // 模拟任务执行
        // 在实际实现中，这里会调用 AI Agent 或执行具体的任务逻辑
        tokio::time::sleep(tokio::time::Duration::from_secs(2)).await;

        let result = format!(
            "任务 {} 执行完成\n标题: {}\n描述: {:?}\n分组: {}",
            task.task_id,
            task.title,
            task.description,
            task.group_name
        );

        Ok(result)
    }

    /// 标记任务为运行中
    async fn mark_task_running(&self, task_id: i64) -> Result<()> {
        let now = chrono::Utc::now();
        sqlx::query!(
            r#"
            UPDATE tasks
            SET status = ?,
                started_at = ?,
                lock_holder = ?,
                lock_time = ?
            WHERE id = ?
            "#,
            TaskStatus::Running,
            now,
            "master",
            now,
            task_id
        )
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

        sqlx::query!(
            r#"
            UPDATE tasks
            SET status = ?,
                completed_at = ?,
                result = ?,
                error = ?,
                lock_holder = NULL,
                lock_time = NULL
            WHERE id = ?
            "#,
            status,
            now,
            result,
            error,
            task_id
        )
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
}
