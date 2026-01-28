//! 记忆管理 API 路由

use super::AppState;
use agentflow_core::{ApiResponse, MemoryCategory, MemoryEntry};
use axum::{
    extract::{Path, Query, State},
    Json,
};
use serde::Deserialize;
use tracing::{debug, info};

/// 记忆搜索查询参数
#[derive(Debug, Deserialize)]
pub struct MemorySearchQuery {
    /// 搜索关键词
    pub q: Option<String>,
    /// 记忆分类
    pub category: Option<String>,
    /// 任务 ID
    pub task_id: Option<String>,
    /// 结果限制
    pub limit: Option<usize>,
}

/// 搜索记忆
///
/// GET/POST /api/v1/memory/search
pub async fn search_memory(
    State(state): State<AppState>,
    Query(params): Query<MemorySearchQuery>,
) -> Result<Json<ApiResponse<Vec<MemoryEntry>>>, ApiError> {
    let query = params.q.unwrap_or_default();
    let category = params.category.and_then(|c| match c.as_str() {
        "execution" => Some(MemoryCategory::Execution),
        "context" => Some(MemoryCategory::Context),
        "result" => Some(MemoryCategory::Result),
        "error" => Some(MemoryCategory::Error),
        "checkpoint" => Some(MemoryCategory::Checkpoint),
        _ => None,
    });

    info!(
        "搜索记忆: query='{}', category={:?}, task_id={:?}",
        query, category, params.task_id
    );

    let results = state
        .memory
        .search(&query, category, params.task_id, params.limit)
        .await
        .map_err(|e| ApiError::internal(format!("搜索记忆失败: {}", e)))?;

    debug!("找到 {} 条记忆", results.len());

    Ok(Json(ApiResponse {
        success: true,
        data: Some(results),
        error: None,
    }))
}

/// 获取指定记忆
///
/// GET /api/v1/memory/:key
pub async fn get_memory(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> Result<Json<ApiResponse<MemoryEntry>>, ApiError> {
    debug!("获取记忆: {}", key);

    let entry = state
        .memory
        .get(&key)
        .await
        .map_err(|e| ApiError::internal(format!("获取记忆失败: {}", e)))?
        .ok_or_else(|| ApiError::not_found(format!("记忆 '{}' 不存在或已过期", key)))?;

    Ok(Json(ApiResponse {
        success: true,
        data: Some(entry),
        error: None,
    }))
}

/// 删除指定记忆
///
/// DELETE /api/v1/memory/:key
pub async fn delete_memory(
    State(state): State<AppState>,
    Path(key): Path<String>,
) -> Result<Json<ApiResponse<serde_json::Value>>, ApiError> {
    info!("删除记忆: {}", key);

    let deleted = state
        .memory
        .delete(&key)
        .await
        .map_err(|e| ApiError::internal(format!("删除记忆失败: {}", e)))?;

    if !deleted {
        return Err(ApiError::not_found(format!("记忆 '{}' 不存在", key)));
    }

    Ok(Json(ApiResponse {
        success: true,
        data: Some(serde_json::json!({"key": key, "deleted": true})),
        error: None,
    }))
}

/// 获取记忆统计信息
///
/// GET /api/v1/memory/stats
pub async fn memory_stats(
    State(state): State<AppState>,
) -> Result<Json<ApiResponse<serde_json::Value>>, ApiError> {
    debug!("获取记忆统计信息");

    let stats = state
        .memory
        .stats()
        .await
        .map_err(|e| ApiError::internal(format!("获取统计信息失败: {}", e)))?;

    let category_counts: serde_json::Value = stats
        .category_counts
        .into_iter()
        .map(|(k, v)| {
            let category_str = match k {
                MemoryCategory::Execution => "execution",
                MemoryCategory::Context => "context",
                MemoryCategory::Result => "result",
                MemoryCategory::Error => "error",
                MemoryCategory::Checkpoint => "checkpoint",
            };
            (category_str.to_string(), v)
        })
        .collect();

    let stats_json = serde_json::json!({
        "total": stats.total,
        "active": stats.active,
        "expired": stats.expired,
        "category_counts": category_counts
    });

    Ok(Json(ApiResponse {
        success: true,
        data: Some(stats_json),
        error: None,
    }))
}

/// API 错误类型
pub struct ApiError {
    status: axum::http::StatusCode,
    message: String,
}

impl ApiError {
    fn not_found(message: impl Into<String>) -> Self {
        Self {
            status: axum::http::StatusCode::NOT_FOUND,
            message: message.into(),
        }
    }

    fn internal(message: impl Into<String>) -> Self {
        Self {
            status: axum::http::StatusCode::INTERNAL_SERVER_ERROR,
            message: message.into(),
        }
    }
}

impl axum::response::IntoResponse for ApiError {
    fn into_response(self) -> axum::response::Response {
        let body = serde_json::json!({
            "success": false,
            "error": self.message,
        });

        (self.status, Json(body)).into_response()
    }
}
