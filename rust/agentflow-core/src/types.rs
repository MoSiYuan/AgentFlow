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

impl std::fmt::Display for TaskStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            TaskStatus::Pending => write!(f, "pending"),
            TaskStatus::Running => write!(f, "running"),
            TaskStatus::Completed => write!(f, "completed"),
            TaskStatus::Failed => write!(f, "failed"),
            TaskStatus::Blocked => write!(f, "blocked"),
        }
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
    /// 记忆域（私域/公域）
    #[serde(default = "default_memory_domain")]
    pub domain: MemoryDomain,
    /// 命名空间（用于白名单匹配）
    #[serde(default = "default_namespace")]
    pub namespace: String,
    /// 已推送的目标节点列表
    #[serde(default)]
    pub pushed_targets: Vec<String>,
}

/// 默认记忆域（公域）
fn default_memory_domain() -> MemoryDomain {
    MemoryDomain::Public
}

/// 默认命名空间
fn default_namespace() -> String {
    "default".to_string()
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

// ==================== 推送系统类型 ====================

/// 记忆域分类（私域/公域隔离）
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MemoryDomain {
    /// 私域记忆：加密存储，向量索引只在本地，永不推送到公域
    Private,
    /// 公域记忆：可推送到其他节点，向量索引可共享
    Public,
}

impl Default for MemoryDomain {
    fn default() -> Self {
        Self::Public // 默认为公域，保守策略
    }
}

impl std::fmt::Display for MemoryDomain {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            MemoryDomain::Private => write!(f, "private"),
            MemoryDomain::Public => write!(f, "public"),
        }
    }
}

/// 加密向量（私域记忆内容）
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncryptedVec {
    /// AES-256-GCM 加密的内容
    pub ciphertext: Vec<u8>,
    /// Nonce（用于 GCM 认证）
    pub nonce: [u8; 12],
    /// 认证标签
    pub tag: [u8; 16],
}

/// 记忆负载（推送时的记忆内容）
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryPayload {
    pub memory_id: String,
    pub key: String,
    pub value: serde_json::Value,
    pub category: MemoryCategory,
    pub domain: MemoryDomain,
    pub namespace: String,
    pub timestamp: i64,
    pub expires_at: Option<i64>,
}

/// 推送消息类型
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum PushMessage {
    /// 首次推送
    Create(MemoryPayload),
    /// 更新内容
    Update(MemoryPayload),
    /// 删除（墓碑消息）
    Delete {
        memory_id: String,
        reason: String,
    },
    /// 权限收回（强制删除）
    Revoke {
        memory_id: String,
        timestamp: DateTime<Utc>,
    },
}

/// 回执状态
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ReceiptStatus {
    /// 接收方明确 ACK，并返回存储哈希
    Confirmed,
    /// 已发送，等待 ACK（对方离线）
    Pending,
    /// 明确拒绝（空间不足/权限不足/格式错误）
    Rejected,
    /// 超时未 ACK，可能在网络队列中
    Timeout,
}

impl std::fmt::Display for ReceiptStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ReceiptStatus::Confirmed => write!(f, "confirmed"),
            ReceiptStatus::Pending => write!(f, "pending"),
            ReceiptStatus::Rejected => write!(f, "rejected"),
            ReceiptStatus::Timeout => write!(f, "timeout"),
        }
    }
}

/// 推送回执
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PushReceipt {
    /// 目标节点
    pub target: String,
    /// 回执状态
    pub status: ReceiptStatus,
    /// 接收方存储位置（用于验证）
    pub local_path: Option<String>,
    /// 同步时间
    pub synced_at: Option<DateTime<Utc>>,
    /// 错误信息（如果被拒绝）
    pub error: Option<String>,
    /// 重试次数
    pub retry_count: u32,
}

/// 目标白名单
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TargetWhitelist {
    /// 域名 -> 目标节点列表映射
    pub domains: std::collections::HashMap<String, Vec<String>>,
    /// 记忆命名空间 -> 域绑定规则
    pub memory_boundaries: std::collections::HashMap<String, Vec<String>>,
}

impl TargetWhitelist {
    /// 检查目标是否在白名单中
    pub fn is_allowed(&self, namespace: &str, target: &str) -> bool {
        // 1. 查找命名空间绑定的域
        let domains = self.memory_boundaries.get(namespace);

        if let Some(domains) = domains {
            // 2. 检查目标是否在任意一个允许的域中
            for domain in domains {
                if let Some(targets) = self.domains.get(domain) {
                    if targets.contains(&target.to_string()) {
                        return true;
                    }
                }
            }
            false
        } else {
            // 如果没有明确绑定，默认不允许（安全优先）
            false
        }
    }

    /// 获取命名空间允许的目标列表
    pub fn allowed_targets(&self, namespace: &str) -> Vec<String> {
        let mut targets = Vec::new();

        if let Some(domains) = self.memory_boundaries.get(namespace) {
            for domain in domains {
                if let Some(domain_targets) = self.domains.get(domain) {
                    targets.extend(domain_targets.clone());
                }
            }
        }

        targets.sort();
        targets.dedup();
        targets
    }
}

impl Default for TargetWhitelist {
    fn default() -> Self {
        Self {
            domains: std::collections::HashMap::new(),
            memory_boundaries: std::collections::HashMap::new(),
        }
    }
}

/// 推送配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PushConfig {
    /// 私域命名空间列表
    pub private_namespaces: Vec<String>,
    /// 公域命名空间列表
    pub public_namespaces: Vec<String>,
    /// 目标白名单
    pub target_whitelist: TargetWhitelist,
    /// 私域是否必须加密
    pub require_encryption: bool,
    /// 推送超时时间（秒）
    pub push_timeout: u64,
    /// 最大重试次数
    pub max_retries: u32,
}

impl Default for PushConfig {
    fn default() -> Self {
        Self {
            private_namespaces: vec![
                "diary/*".to_string(),
                "personal/*".to_string(),
                "keys/*".to_string(),
            ],
            public_namespaces: vec![
                "work/*".to_string(),
                "blog-drafts/*".to_string(),
                "open-source/*".to_string(),
            ],
            target_whitelist: TargetWhitelist::default(),
            require_encryption: true,
            push_timeout: 30,
            max_retries: 3,
        }
    }
}

/// 撤回状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RevokeStatus {
    /// 记忆 ID
    pub memory_id: String,
    /// 总共需要撤回的目标数
    pub total_targets: usize,
    /// 已确认撤回的目标数
    pub confirmed: usize,
    /// 待确认的目标数
    pub pending: usize,
    /// 失败的目标数
    pub failed: usize,
    /// 详细状态
    pub details: Vec<RevokeDetail>,
}

/// 撤回详细信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RevokeDetail {
    /// 目标节点
    pub target: String,
    /// 撤回状态
    pub status: ReceiptStatus,
    /// 时间戳
    pub timestamp: DateTime<Utc>,
    /// 错误信息（如果失败）
    pub error: Option<String>,
}
