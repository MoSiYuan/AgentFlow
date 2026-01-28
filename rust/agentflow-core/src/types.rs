//! AgentFlow 核心类型定义

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// 任务状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "task_status", rename_all = "lowercase")]
pub enum TaskStatus {
    Pending,
    Running,
    Completed,
    Failed,
    Blocked,
}

impl Default for TaskStatus {
    fn default() -> Self {
        Self::Pending
    }
}

/// 任务优先级
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize, sqlx::Type)]
#[serde(rename_all = "lowercase")]
#[sqlx(type_name = "task_priority", rename_all = "lowercase")]
pub enum TaskPriority {
    Low,
    Medium,
    High,
}

impl TaskPriority {
    pub fn as_i32(&self) -> i32 {
        match self {
            Self::Low => 0,
            Self::Medium => 1,
            Self::High => 2,
        }
    }

    pub fn from_i32(value: i32) -> Self {
        match value {
            0 => Self::Low,
            1 => Self::Medium,
            2 => Self::High,
            _ => Self::Low,
        }
    }
}

impl Default for TaskPriority {
    fn default() -> Self {
        Self::Medium
    }
}

/// 任务
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Task {
    pub id: i64,
    pub task_id: String,
    pub parent_id: Option<i64>,
    pub title: String,
    pub description: Option<String>,
    pub group_name: String,
    pub completion_criteria: Option<String>,
    pub status: TaskStatus,
    pub priority: TaskPriority,
    pub lock_holder: Option<String>,
    pub lock_time: Option<DateTime<Utc>>,
    pub result: Option<String>,
    pub error: Option<String>,
    pub workspace_dir: Option<String>,
    pub sandboxed: bool,
    pub allow_network: bool,
    pub max_memory: Option<String>,
    pub max_cpu: Option<i32>,
    pub created_by: Option<String>,
    pub created_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
}

/// 创建任务请求
#[derive(Debug, Deserialize)]
pub struct CreateTaskRequest {
    pub parent_id: Option<i64>,
    pub title: String,
    pub description: Option<String>,
    pub group_name: Option<String>,
    pub completion_criteria: Option<String>,
    pub priority: Option<String>,
    pub workspace_dir: Option<String>,
    pub sandboxed: Option<bool>,
    pub allow_network: Option<bool>,
    pub max_memory: Option<String>,
    pub max_cpu: Option<i32>,
    pub created_by: Option<String>,
}

/// 任务过滤器
#[derive(Debug, Default, Clone)]
pub struct TaskFilter {
    pub status: Option<TaskStatus>,
    pub group_name: Option<String>,
    pub parent_id: Option<i64>,
    pub limit: Option<usize>,
    pub offset: Option<usize>,
}

/// Worker 状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum WorkerStatus {
    Active,
    Inactive,
}

impl Default for WorkerStatus {
    fn default() -> Self {
        Self::Active
    }
}

/// Worker
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Worker {
    pub id: String,
    pub group_name: String,
    pub r#type: WorkerType,
    pub capabilities: Vec<String>,
    pub status: WorkerStatus,
    pub last_heartbeat: DateTime<Utc>,
    pub created_at: DateTime<Utc>,
}

/// Worker 类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum WorkerType {
    Local,
    Remote,
}

impl Default for WorkerType {
    fn default() -> Self {
        Self::Local
    }
}

/// Worker 注册请求
#[derive(Debug, Deserialize)]
pub struct WorkerRegistration {
    pub worker_id: String,
    pub worker_name: Option<String>,
    pub group_name: String,
    pub platform: Option<String>,
    pub capabilities: Vec<String>,
}

/// 健康检查响应
#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: String,
    pub uptime: u64,
    pub mode: String,
}

/// API 响应
#[derive(Debug, Serialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<T>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

/// 记忆条目分类
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MemoryCategory {
    Execution,
    Context,
    Result,
    Error,
    Checkpoint,
}

/// 记忆条目
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryEntry {
    pub key: String,
    pub value: serde_json::Value,
    pub expires_at: Option<i64>,
    pub category: MemoryCategory,
    pub task_id: Option<String>,
    pub timestamp: i64,
}

/// 记忆快照
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemorySnapshot {
    pub entries: Vec<MemoryEntry>,
    pub worker_id: String,
    pub snapshot_time: DateTime<Utc>,
}

/// 经验总结类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "summary_type", rename_all = "snake_case")]
pub enum ExperienceType {
    SuccessPattern,
    FailurePattern,
    Optimization,
    BestPractice,
}

/// 经验总结
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExperienceSummary {
    pub id: Option<i64>,
    pub summary_type: ExperienceType,
    pub pattern_description: String,
    pub context: Option<serde_json::Value>,
    pub confidence_score: f64,
    pub usage_count: i32,
    pub success_count: i32,
    pub created_at: Option<DateTime<Utc>>,
    pub last_used_at: Option<DateTime<Utc>>,
}
