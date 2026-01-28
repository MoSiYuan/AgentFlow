//! 健康检查 API 路由

use agentflow_core::{ApiResponse, HealthResponse};
use axum::{extract::State, Json};
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
    // 修复: chrono 0.4+ 中，DateTime 的差值是 TimeDelta，使用 num_seconds()
    let uptime = {
        let start = state.start_time;
        let now = Utc::now();
        let duration = now.signed_duration_since(start);
        duration.num_seconds().max(0) as u64
    };

    // 检查任务执行器状态
    // 修复: 明确类型注解以解决类型推断问题
    let running_tasks: Vec<i64> = state.executor.get_running_tasks().await;
    let running_count = running_tasks.len();

    // 检查记忆统计
    // 修复: 明确类型注解以解决类型推断问题
    let memory_stats = match state.memory.stats().await {
        Ok(stats) => stats,
        Err(_) => crate::memory_core::MemoryStats {
            total: 0,
            active: 0,
            expired: 0,
            category_counts: std::collections::HashMap::new(),
        },
    };

    let response = HealthResponse {
        status: if running_count < state.executor.max_concurrent_tasks {
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
        response.status, uptime, running_count, memory_stats.total
    );

    Json(ApiResponse {
        success: true,
        data: Some(response),
        error: None,
    })
}
