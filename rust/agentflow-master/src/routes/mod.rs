//! API 路由模块
//!
//! 定义所有 HTTP 和 WebSocket 路由处理函数

pub mod tasks;
pub mod memory;
pub mod websocket;
pub mod health;

use axum::{
    routing::{get, post},
    Router,
};

use crate::{AppState, executor::TaskExecutor, memory_core::MemoryCore};

/// 创建 API 路由
pub fn create_routes() -> Router<AppState> {
    Router::new()
        // 健康检查
        .route("/health", get(health::health_check))
        .route("/api/v1/health", get(health::health_check))
        // 任务管理 API
        .route("/api/v1/tasks", post(tasks::create_task).get(tasks::list_tasks))
        .route(
            "/api/v1/tasks/:id",
            get(tasks::get_task).delete(tasks::delete_task),
        )
        .route("/api/v1/tasks/:id/execute", post(tasks::execute_task))
        .route("/api/v1/tasks/:id/cancel", post(tasks::cancel_task))
        // 记忆管理 API
        .route(
            "/api/v1/memory/search",
            get(memory::search_memory).post(memory::search_memory),
        )
        .route("/api/v1/memory/:key", get(memory::get_memory).delete(memory::delete_memory))
        .route("/api/v1/memory/stats", get(memory::memory_stats))
        // WebSocket
        .route("/ws/task/:id", get(websocket::task_websocket))
}

/// 应用状态
#[derive(Clone)]
pub struct AppState {
    /// 任务执行器
    pub executor: TaskExecutor,
    /// 记忆核心
    pub memory: MemoryCore,
    /// 服务器启动时间
    pub start_time: chrono::DateTime<chrono::Utc>,
}
