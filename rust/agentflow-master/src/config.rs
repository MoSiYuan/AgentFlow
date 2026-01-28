//! AgentFlow Master 配置管理
//!
//! 负责读取环境变量和配置文件，提供 Master 服务器所需的所有配置项

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::env;

/// Master 服务器配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MasterConfig {
    /// 服务器地址
    pub server_addr: String,
    /// 服务器端口
    pub server_port: u16,
    /// 数据库 URL
    pub database_url: String,
    /// Sandbox 配置
    pub sandbox: SandboxConfig,
    /// Memory 配置
    pub memory: MemoryConfig,
    /// 日志级别
    pub log_level: String,
    /// Worker 心跳超时（秒）
    pub worker_heartbeat_timeout: u64,
    /// 任务执行超时（秒）
    pub task_timeout: u64,
    /// 最大并发任务数
    pub max_concurrent_tasks: usize,
}

/// Sandbox 配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SandboxConfig {
    /// 是否启用沙箱
    pub enabled: bool,
    /// 默认工作目录
    pub default_workspace: String,
    /// 允许网络访问
    pub allow_network: bool,
    /// 最大内存限制
    pub max_memory: Option<String>,
    /// 最大 CPU 限制
    pub max_cpu: Option<i32>,
    /// 超时时间（秒）
    pub timeout: Option<u64>,
}

/// Memory 配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryConfig {
    /// 记忆存储后端 (memory, sqlite, redis)
    pub backend: String,
    /// 数据库连接 URL（用于 sqlite 后端）
    pub database_url: Option<String>,
    /// Redis 连接 URL（用于 redis 后端）
    pub redis_url: Option<String>,
    /// 默认过期时间（秒）
    pub default_ttl: u64,
    /// 最大条目数
    pub max_entries: usize,
    /// 是否启用持久化
    pub enable_persistence: bool,
}

impl Default for MasterConfig {
    fn default() -> Self {
        Self {
            server_addr: "0.0.0.0".to_string(),
            server_port: 6767,
            database_url: "sqlite://agentflow.db".to_string(),
            sandbox: SandboxConfig::default(),
            memory: MemoryConfig::default(),
            log_level: "info".to_string(),
            worker_heartbeat_timeout: 60,
            task_timeout: 300,
            max_concurrent_tasks: 10,
        }
    }
}

impl Default for SandboxConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            default_workspace: "/tmp/agentflow/workspace".to_string(),
            allow_network: false,
            max_memory: Some("1G".to_string()),
            max_cpu: Some(2),
            timeout: Some(300),
        }
    }
}

impl Default for MemoryConfig {
    fn default() -> Self {
        Self {
            backend: "memory".to_string(),
            database_url: Some("sqlite://agentflow_memory.db".to_string()),
            redis_url: None,
            default_ttl: 3600,
            max_entries: 10000,
            enable_persistence: false,
        }
    }
}

impl MasterConfig {
    /// 从环境变量加载配置
    pub fn from_env() -> Result<Self> {
        let mut config = Self::default();

        // 服务器配置
        if let Ok(addr) = env::var("AGENTFLOW_SERVER_ADDR") {
            config.server_addr = addr;
        }
        if let Ok(port) = env::var("AGENTFLOW_SERVER_PORT") {
            config.server_port = port.parse()
                .context("无效的 AGENTFLOW_SERVER_PORT")?;
        }

        // 数据库配置
        if let Ok(db_url) = env::var("AGENTFLOW_DATABASE_URL") {
            config.database_url = db_url;
        }

        // Sandbox 配置
        config.sandbox = SandboxConfig::from_env()?;

        // Memory 配置
        config.memory = MemoryConfig::from_env()?;

        // 日志配置
        if let Ok(log_level) = env::var("AGENTFLOW_LOG_LEVEL") {
            config.log_level = log_level;
        }

        // Worker 心跳超时
        if let Ok(timeout) = env::var("AGENTFLOW_WORKER_HEARTBEAT_TIMEOUT") {
            config.worker_heartbeat_timeout = timeout.parse()
                .context("无效的 AGENTFLOW_WORKER_HEARTBEAT_TIMEOUT")?;
        }

        // 任务超时
        if let Ok(timeout) = env::var("AGENTFLOW_TASK_TIMEOUT") {
            config.task_timeout = timeout.parse()
                .context("无效的 AGENTFLOW_TASK_TIMEOUT")?;
        }

        // 最大并发任务数
        if let Ok(max_tasks) = env::var("AGENTFLOW_MAX_CONCURRENT_TASKS") {
            config.max_concurrent_tasks = max_tasks.parse()
                .context("无效的 AGENTFLOW_MAX_CONCURRENT_TASKS")?;
        }

        Ok(config)
    }

    /// 加载 .env 文件并从环境变量读取配置
    pub fn load() -> Result<Self> {
        // 尝试加载 .env 文件
        let _ = dotenvy::dotenv();

        Self::from_env()
    }

    /// 获取服务器绑定地址
    pub fn bind_address(&self) -> String {
        format!("{}:{}", self.server_addr, self.server_port)
    }

    /// 获取数据库连接池配置
    pub fn pool_options(&self) -> sqlx::sqlite::SqliteConnectOptions {
        sqlx::sqlite::SqliteConnectOptions::new()
            .filename(self.database_url.replace("sqlite://", ""))
            .create_if_missing(true)
    }
}

impl SandboxConfig {
    /// 从环境变量加载 Sandbox 配置
    pub fn from_env() -> Result<Self> {
        let mut config = Self::default();

        if let Ok(enabled) = env::var("AGENTFLOW_SANDBOX_ENABLED") {
            config.enabled = enabled.parse()
                .context("无效的 AGENTFLOW_SANDBOX_ENABLED")?;
        }

        if let Ok(workspace) = env::var("AGENTFLOW_SANDBOX_WORKSPACE") {
            config.default_workspace = workspace;
        }

        if let Ok(allow_network) = env::var("AGENTFLOW_SANDBOX_ALLOW_NETWORK") {
            config.allow_network = allow_network.parse()
                .context("无效的 AGENTFLOW_SANDBOX_ALLOW_NETWORK")?;
        }

        if let Ok(max_memory) = env::var("AGENTFLOW_SANDBOX_MAX_MEMORY") {
            config.max_memory = Some(max_memory);
        }

        if let Ok(max_cpu) = env::var("AGENTFLOW_SANDBOX_MAX_CPU") {
            config.max_cpu = Some(max_cpu.parse()
                .context("无效的 AGENTFLOW_SANDBOX_MAX_CPU")?);
        }

        if let Ok(timeout) = env::var("AGENTFLOW_SANDBOX_TIMEOUT") {
            config.timeout = Some(timeout.parse()
                .context("无效的 AGENTFLOW_SANDBOX_TIMEOUT")?);
        }

        Ok(config)
    }
}

impl MemoryConfig {
    /// 从环境变量加载 Memory 配置
    pub fn from_env() -> Result<Self> {
        let mut config = Self::default();

        if let Ok(backend) = env::var("AGENTFLOW_MEMORY_BACKEND") {
            config.backend = backend;
        }

        if let Ok(db_url) = env::var("AGENTFLOW_MEMORY_DATABASE_URL") {
            config.database_url = Some(db_url);
        }

        if let Ok(redis_url) = env::var("AGENTFLOW_MEMORY_REDIS_URL") {
            config.redis_url = Some(redis_url);
        }

        if let Ok(ttl) = env::var("AGENTFLOW_MEMORY_DEFAULT_TTL") {
            config.default_ttl = ttl.parse()
                .context("无效的 AGENTFLOW_MEMORY_DEFAULT_TTL")?;
        }

        if let Ok(max_entries) = env::var("AGENTFLOW_MEMORY_MAX_ENTRIES") {
            config.max_entries = max_entries.parse()
                .context("无效的 AGENTFLOW_MEMORY_MAX_ENTRIES")?;
        }

        if let Ok(enabled) = env::var("AGENTFLOW_MEMORY_PERSISTENCE") {
            config.enable_persistence = enabled.parse()
                .context("无效的 AGENTFLOW_MEMORY_PERSISTENCE")?;
        }

        Ok(config)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = MasterConfig::default();
        assert_eq!(config.server_addr, "0.0.0.0");
        assert_eq!(config.server_port, 6767);
        assert_eq!(config.bind_address(), "0.0.0.0:6767");
    }

    #[test]
    fn test_bind_address() {
        let config = MasterConfig {
            server_addr: "127.0.0.1".to_string(),
            server_port: 8080,
            ..Default::default()
        };
        assert_eq!(config.bind_address(), "127.0.0.1:8080");
    }
}
