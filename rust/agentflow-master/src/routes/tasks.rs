//! 任务管理 API 路由

use super::AppState;
use agentflow_core::{
    ApiResponse, CreateTaskRequest, Task, TaskPriority, TaskStatus,
};
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{sse::{Event, Sse}, IntoResponse, Json},
    Json as JsonExtractor,
};
use futures::stream::Stream;
use serde_json::json;
use sqlx::Row;
use std::convert::Infallible;
use tracing::{debug, error, info};

// 辅助函数: 从数据库行转换为 Task
fn row_to_task(row: &sqlx::sqlite::SqliteRow) -> Result<Task, anyhow::Error> {
    use agentflow_core::{TaskStatus, TaskPriority};
    use sqlx::Row;

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

    Ok(Task {
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
    })
}

/// 创建任务
///
/// POST /api/v1/tasks
pub async fn create_task(
    State(state): State<AppState>,
    JsonExtractor(req): JsonExtractor<CreateTaskRequest>,
) -> Result<Json<ApiResponse<Task>>, ApiError> {
    info!("创建任务: {}", req.title);

    let task_id = uuid::Uuid::new_v4().to_string();
    let priority = req
        .priority
        .and_then(|p| match p.as_str() {
            "low" => Some(TaskPriority::Low),
            "medium" => Some(TaskPriority::Medium),
            "high" => Some(TaskPriority::High),
            _ => None,
        })
        .unwrap_or_default();

    let group_name = req.group_name.unwrap_or_else(|| "default".to_string());
    let now = chrono::Utc::now();

    // 插入任务到数据库
    // 修复: 使用 runtime query
    let result = sqlx::query(
        r#"
        INSERT INTO tasks (
            task_id, parent_id, title, description, group_name,
            completion_criteria, status, priority, workspace_dir,
            sandboxed, allow_network, max_memory, max_cpu, created_by,
            created_at, started_at, completed_at
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL)
        "#
    )
    .bind(&task_id)
    .bind(req.parent_id)
    .bind(&req.title)
    .bind(&req.description)
    .bind(&group_name)
    .bind(&req.completion_criteria)
    .bind(TaskStatus::Pending.to_string())
    .bind(priority.as_i32())
    .bind(&req.workspace_dir)
    .bind(req.sandboxed.unwrap_or(false) as i64)
    .bind(req.allow_network.unwrap_or(false) as i64)
    .bind(&req.max_memory)
    .bind(req.max_cpu)
    .bind(&req.created_by)
    .bind(now)
    .fetch_one(state.executor.db())
    .await
    .map_err(|e| {
        error!("创建任务失败: {}", e);
        ApiError::internal(format!("创建任务失败: {}", e))
    })?;

    // 修复: 从 last_insert_rowid 获取插入的 ID，而不是从 result.id
    let inserted_id = sqlx::query("SELECT last_insert_rowid() as id")
        .fetch_one(state.executor.db())
        .await
        .map_err(|e| {
            error!("获取插入 ID 失败: {}", e);
            ApiError::internal(format!("获取插入 ID 失败: {}", e))
        })?;
    let id: i64 = inserted_id.get("id");

    let task = Task {
        id,
        task_id,
        parent_id: req.parent_id,
        title: req.title,
        description: req.description,
        group_name,
        completion_criteria: req.completion_criteria,
        status: TaskStatus::Pending,
        priority,
        lock_holder: None,
        lock_time: None,
        result: None,
        error: None,
        workspace_dir: req.workspace_dir,
        sandboxed: req.sandboxed.unwrap_or(false),
        allow_network: req.allow_network.unwrap_or(false),
        max_memory: req.max_memory,
        max_cpu: req.max_cpu,
        created_by: req.created_by,
        created_at: now,
        started_at: None,
        completed_at: None,
    };

    info!("任务创建成功: ID={}, task_id={}", task.id, task.task_id);

    Ok(Json(ApiResponse {
        success: true,
        data: Some(task),
        error: None,
    }))
}

/// 获取任务详情
///
/// GET /api/v1/tasks/:id
pub async fn get_task(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<Json<ApiResponse<Task>>, ApiError> {
    debug!("获取任务: {}", id);

    // 修复: 使用 runtime query
    let row = sqlx::query(
        r#"
        SELECT
            id, task_id, parent_id, title, description, group_name,
            completion_criteria, status, priority, lock_holder, lock_time, result, error,
            workspace_dir, sandboxed, allow_network, max_memory, max_cpu,
            created_by, created_at, started_at, completed_at
        FROM tasks
        WHERE id = ?
        "#
    )
    .bind(id)
    .fetch_one(state.executor.db())
    .await
    .map_err(|e| {
        error!("获取任务失败: {}", e);
        ApiError::not_found(format!("任务 {} 不存在", id))
    })?;

    let task = row_to_task(&row).map_err(|e| {
        error!("解析任务失败: {}", e);
        ApiError::internal(format!("解析任务失败: {}", e))
    })?;

    Ok(Json(ApiResponse {
        success: true,
        data: Some(task),
        error: None,
    }))
}

/// 列出所有任务
///
/// GET /api/v1/tasks
pub async fn list_tasks(
    State(state): State<AppState>,
) -> Result<Json<ApiResponse<Vec<Task>>>, ApiError> {
    debug!("列出所有任务");

    // 修复: 使用 runtime query
    let rows = sqlx::query(
        r#"
        SELECT
            id, task_id, parent_id, title, description, group_name,
            completion_criteria, status, priority, lock_holder, lock_time, result, error,
            workspace_dir, sandboxed, allow_network, max_memory, max_cpu,
            created_by, created_at, started_at, completed_at
        FROM tasks
        ORDER BY created_at DESC
        LIMIT 100
        "#
    )
    .fetch_all(state.executor.db())
    .await
    .map_err(|e| {
        error!("列出任务失败: {}", e);
        ApiError::internal(format!("列出任务失败: {}", e))
    })?;

    let tasks: Result<Vec<Task>, _> = rows.iter().map(row_to_task).collect();
    let tasks = tasks.map_err(|e| {
        error!("解析任务列表失败: {}", e);
        ApiError::internal(format!("解析任务列表失败: {}", e))
    })?;

    Ok(Json(ApiResponse {
        success: true,
        data: Some(tasks),
        error: None,
    }))
}

/// 执行任务（SSE 流式输出）
///
/// POST /api/v1/tasks/:id/execute
pub async fn execute_task(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Sse<impl Stream<Item = Result<Event, anyhow::Error>>> {
    info!("执行任务（SSE）: {}", id);

    // 创建 SSE 流
    let stream = async_stream::try_stream! {
        // 发送开始事件
        yield Event::default()
            .json_data(json!({
                "type": "start",
                "task_id": id,
                "message": "开始执行任务"
            }))?;

        // 执行任务
        match state.executor.execute_task(id).await {
            Ok(result) => {
                yield Event::default()
                    .json_data(json!({
                        "type": "progress",
                        "task_id": id,
                        "message": "任务执行中"
                    }))?;

                yield Event::default()
                    .json_data(json!({
                        "type": "complete",
                        "task_id": id,
                        "result": result
                    }))?;
            }
            Err(e) => {
                yield Event::default()
                    .json_data(json!({
                        "type": "error",
                        "task_id": id,
                        "error": e.to_string()
                    }))?;
            }
        }
    };

    Sse::new(stream).keep_alive(
        axum::response::sse::KeepAlive::new()
            .interval(std::time::Duration::from_secs(1))
            .text("keepalive"),
    )
}

/// 取消任务
///
/// POST /api/v1/tasks/:id/cancel
pub async fn cancel_task(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<Json<ApiResponse<serde_json::Value>>, ApiError> {
    info!("取消任务: {}", id);

    // 检查任务是否存在
    // 修复: 使用 runtime query
    let row = sqlx::query(
        r#"
        SELECT
            id, task_id, parent_id, title, description, group_name,
            completion_criteria, status, priority, lock_holder, lock_time, result, error,
            workspace_dir, sandboxed, allow_network, max_memory, max_cpu,
            created_by, created_at, started_at, completed_at
        FROM tasks
        WHERE id = ?
        "#
    )
    .bind(id)
    .fetch_optional(state.executor.db())
    .await
    .map_err(|e| {
        error!("查询任务失败: {}", e);
        ApiError::internal(format!("查询任务失败: {}", e))
    })?;

    if row.is_none() {
        return Err(ApiError::not_found(format!("任务 {} 不存在", id)));
    }

    let task = row_to_task(&row.unwrap()).map_err(|e| {
        error!("解析任务失败: {}", e);
        ApiError::internal(format!("解析任务失败: {}", e))
    })?;

    // 只能取消运行中的任务
    if task.status != TaskStatus::Running {
        return Err(ApiError::bad_request(format!(
            "任务状态为 {:?}，无法取消",
            task.status
        )));
    }

    // 更新任务状态为已取消（使用 Blocked 状态表示）
    let now = chrono::Utc::now();
    // 修复: 使用 runtime query
    sqlx::query(
        r#"
        UPDATE tasks
        SET status = ?,
            completed_at = ?,
            lock_holder = NULL,
            lock_time = NULL
        WHERE id = ?
        "#
    )
    .bind(TaskStatus::Blocked.to_string())
    .bind(now)
    .bind(id)
    .execute(state.executor.db())
    .await
    .map_err(|e| {
        error!("取消任务失败: {}", e);
        ApiError::internal(format!("取消任务失败: {}", e))
    })?;

    info!("任务 {} 已取消", id);

    Ok(Json(ApiResponse {
        success: true,
        data: Some(json!({"task_id": id, "cancelled": true})),
        error: None,
    }))
}

/// 删除任务
///
/// DELETE /api/v1/tasks/:id
pub async fn delete_task(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<Json<ApiResponse<serde_json::Value>>, ApiError> {
    info!("删除任务: {}", id);

    // 修复: 使用 runtime query
    let result = sqlx::query("DELETE FROM tasks WHERE id = ?")
        .bind(id)
        .execute(state.executor.db())
        .await
        .map_err(|e| {
            error!("删除任务失败: {}", e);
            ApiError::internal(format!("删除任务失败: {}", e))
        })?;

    if result.rows_affected() == 0 {
        return Err(ApiError::not_found(format!("任务 {} 不存在", id)));
    }

    info!("任务 {} 已删除", id);

    Ok(Json(ApiResponse {
        success: true,
        data: Some(json!({"task_id": id, "deleted": true})),
        error: None,
    }))
}

/// API 错误类型
pub struct ApiError {
    status: StatusCode,
    message: String,
}

impl ApiError {
    fn bad_request(message: impl Into<String>) -> Self {
        Self {
            status: StatusCode::BAD_REQUEST,
            message: message.into(),
        }
    }

    fn not_found(message: impl Into<String>) -> Self {
        Self {
            status: StatusCode::NOT_FOUND,
            message: message.into(),
        }
    }

    fn internal(message: impl Into<String>) -> Self {
        Self {
            status: StatusCode::INTERNAL_SERVER_ERROR,
            message: message.into(),
        }
    }
}

impl IntoResponse for ApiError {
    fn into_response(self) -> axum::response::Response {
        let body = json!({
            "success": false,
            "error": self.message,
        });

        (self.status, Json(body)).into_response()
    }
}
