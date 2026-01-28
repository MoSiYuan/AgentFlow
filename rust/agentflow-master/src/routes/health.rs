//! 健康检查 API 路由

use agentflow_core::{ApiResponse, HealthResponse};
use axum::{Json, State};
use chrono::Utc;
use tracing::debug;

use super::AppState;

/// 健康检查
///
/// GET /health
/// GET /api/v1/health
pub async fn health_check(State(state): State<AppState>) -> Json<ApiResponse<HealthResponse>> {
    debug!("健康检查请求");

    // 计算运行时间（秒）
    let uptime = {
        let start = state.start_time;
        let now = Utc::now();
        (now - start).num_seconds().max(0) as u64
    };

    // 检查任务执行器状态
    let running_tasks = state.executor.get_running_tasks().await.len();

    // 检查记忆统计
    let memory_stats = state.memory.stats().await.unwrap_or_else(|_| crate::memory_core::MemoryStats {
        total: 0,
        active: 0,
        expired: 0,
        category_counts: std::collections::HashMap::new(),
    });

    let response = HealthResponse {
        status: if running_tasks < state.executor.max_concurrent_tasks {
            "healthy".to_string()
        } else {
            "busy".to_string()
        },
        version: env!("CARGO_PKG_VERSION").to_string(),
        uptime,
        mode: "master".to_string(),
    };

    debug!(
        "健康检查: status={}, uptime={}s, running_tasks={}, memory_entries={}",
        response.status, uptime, running_tasks, memory_stats.total
    );

    Json(ApiResponse {
        success: true,
        data: Some(response),
        error: None,
    })
}
