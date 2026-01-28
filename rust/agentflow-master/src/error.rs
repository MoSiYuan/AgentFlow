//! AgentFlow Master 错误处理
//!
//! 定义所有 Master 服务器相关的错误类型

use axum::{
    http::StatusCode,
    response::{IntoResponse, Json, Response},
};
use serde_json::json;
use thiserror::Error;

/// Master 服务器错误类型
#[derive(Error, Debug)]
pub enum MasterError {
    /// 数据库错误
    #[error("数据库错误: {0}")]
    Database(#[from] sqlx::Error),

    /// 任务未找到
    #[error("任务未找到: {0}")]
    TaskNotFound(i64),

    /// 任务执行错误
    #[error("任务执行错误: {0}")]
    TaskExecution(String),

    /// Worker 未找到
    #[error("Worker 未找到: {0}")]
    WorkerNotFound(String),

    /// 无效请求
    #[error("无效请求: {0}")]
    InvalidRequest(String),

    /// 内部错误
    #[error("内部错误: {0}")]
    Internal(String),

    /// 记忆错误
    #[error("记忆错误: {0}")]
    Memory(String),

    /// 配置错误
    #[error("配置错误: {0}")]
    Config(#[from] anyhow::Error),
}

impl IntoResponse for MasterError {
    fn into_response(self) -> Response {
        let (status, error_message) = match &self {
            MasterError::TaskNotFound(_) => (StatusCode::NOT_FOUND, self.to_string()),
            MasterError::WorkerNotFound(_) => (StatusCode::NOT_FOUND, self.to_string()),
            MasterError::InvalidRequest(_) => (StatusCode::BAD_REQUEST, self.to_string()),
            MasterError::TaskExecution(_) => (StatusCode::INTERNAL_SERVER_ERROR, self.to_string()),
            MasterError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, self.to_string()),
            MasterError::Internal(_) => (StatusCode::INTERNAL_SERVER_ERROR, self.to_string()),
            MasterError::Memory(_) => (StatusCode::INTERNAL_SERVER_ERROR, self.to_string()),
            MasterError::Config(_) => (StatusCode::INTERNAL_SERVER_ERROR, self.to_string()),
        };

        let body = json!({
            "success": false,
            "error": error_message,
        });

        (status, Json(body)).into_response()
    }
}

/// API 错误响应
#[derive(Debug)]
pub struct ApiError {
    pub status: StatusCode,
    pub message: String,
}

impl ApiError {
    pub fn new(status: StatusCode, message: impl Into<String>) -> Self {
        Self {
            status,
            message: message.into(),
        }
    }

    pub fn bad_request(message: impl Into<String>) -> Self {
        Self::new(StatusCode::BAD_REQUEST, message)
    }

    pub fn not_found(message: impl Into<String>) -> Self {
        Self::new(StatusCode::NOT_FOUND, message)
    }

    pub fn internal(message: impl Into<String>) -> Self {
        Self::new(StatusCode::INTERNAL_SERVER_ERROR, message)
    }
}

impl IntoResponse for ApiError {
    fn into_response(self) -> Response {
        let body = json!({
            "success": false,
            "error": self.message,
        });

        (self.status, Json(body)).into_response()
    }
}

impl From<anyhow::Error> for ApiError {
    fn from(err: anyhow::Error) -> Self {
        Self::internal(err.to_string())
    }
}

impl From<sqlx::Error> for ApiError {
    fn from(err: sqlx::Error) -> Self {
        Self::internal(err.to_string())
    }
}

impl From<serde_json::Error> for ApiError {
    fn from(err: serde_json::Error) -> Self {
        Self::internal(err.to_string())
    }
}
