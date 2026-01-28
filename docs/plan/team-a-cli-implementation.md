# Team A CLI Implementation Plan - AgentFlow v0.2.1

**Date**: 2026-01-28
**Team**: Team A (CLI & Configuration Layer)
**Branch**: feature/0.2.1
**Status**: Research & Planning Phase

---

## 📋 Executive Summary

This document outlines the implementation plan for **AgentFlow v0.2.1's unified CLI interface with multi-mode support**. The goal is to transform the current "manual `agentflow-master` binary" approach into a polished, user-friendly command-line experience supporting three distinct operational modes:

- **Local Mode**: Master=Worker, direct Claude CLI execution (current default behavior)
- **Cloud Mode**: Distributed planner with node registration and task routing
- **Planner-Only Mode**: Lightweight planning without direct worker execution

---

## 🎯 Objectives

### Primary Goals
1. ✅ Create unified `agentflow` CLI entry point
2. ✅ Implement multi-mode configuration system (local/cloud/planner-only)
3. ✅ Add mock API endpoints for cloud/edge federation
4. ✅ Maintain backward compatibility with existing v3 architecture

### Non-Goals
- Actual node registration implementation (Team C responsibility)
- Real webhook handling (Team C responsibility)
- Memory/Git integration (Team B responsibility)
- Packaging and installation scripts (Team D responsibility)

---

## 🏗️ Current Architecture Analysis

### Existing Structure

```
rust/
├── agentflow-core/          # Shared library
│   ├── src/
│   │   ├── types.rs         # Core type definitions ✅
│   │   ├── database.rs      # Database utilities ✅
│   │   ├── memory/          # Memory system (Team B)
│   │   ├── executor/        # Task executor ✅
│   │   ├── sandbox/         # Sandbox utilities ✅
│   │   └── lib.rs
│   └── Cargo.toml
├── agentflow-master/        # HTTP server
│   ├── src/
│   │   ├── main.rs          # Entry point with basic clap ✅
│   │   ├── config.rs        # Environment-based config ✅
│   │   ├── executor.rs      # Simple task executor ✅
│   │   ├── routes/          # API routes ✅
│   │   ├── error.rs         # Error handling ✅
│   │   └── memory_core.rs   # Memory core ✅
│   └── Cargo.toml           # clap already in dependencies ✅
└── Cargo.toml               # Workspace config
```

### Key Findings

1. **`clap` already available**: clap 4.5 with derive features is in workspace dependencies
2. **Existing config system**: `config.rs` provides environment-based configuration
3. **No TOML support yet**: Need to add `toml` crate dependency
4. **Simple CLI structure**: Current `main.rs` has basic `Args` struct, not subcommands
5. **Routes organized**: Good separation of concerns in `routes/` module

---

## 📐 Design Architecture

### 1. CLI Structure

#### Command Hierarchy

```
agentflow
├── server [mode]           # Start server (default: local)
│   ├── --mode <MODE>       # local | cloud | planner-only
│   ├── --port <PORT>       # Override config port
│   └── --config <PATH>     # Custom config file
├── task
│   ├── create <title> [desc] [options]
│   ├── run <id>            # Execute with streaming output
│   ├── list [filters]      # List all tasks
│   ├── show <id>           # Show task details
│   └── cancel <id>         # Cancel running task
├── memory
│   ├── search <query>      # Search memory store
│   └── stats               # Memory statistics
├── node
│   ├── list                # List registered nodes (mock for now)
│   └── status <id>         # Node status (mock for now)
├── config
│   ├── show                # Display current config
│   ├── validate            # Validate config file
│   └── init                # Generate default config
└── info                    # System information
```

#### Rust Type Definition

```rust
use clap::{Parser, Subcommand};
use clap_derive::Subcommand;

#[derive(Parser, Debug)]
#[command(name = "agentflow")]
#[command(about = "AgentFlow - AI Agent Task Collaboration System", long_about = None)]
#[command(version = env!("CARGO_PKG_VERSION"))]
struct Cli {
    /// Configuration file path (default: ~/.agentflow/config.toml or ./agentflow.toml)
    #[arg(short, long, global = true)]
    config: Option<String>,

    /// Verbose output
    #[arg(short, long, global = true)]
    verbose: bool,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    /// Start AgentFlow server
    Server {
        /// Server mode: local, cloud, or planner-only
        #[arg(short, long, default_value = "local")]
        mode: String,

        /// Server port (overrides config)
        #[arg(short, long)]
        port: Option<u16>,

        /// Bind address (overrides config)
        #[arg(short, long)]
        addr: Option<String>,
    },

    /// Task management
    Task {
        #[command(subcommand)]
        action: TaskCommands,
    },

    /// Memory operations
    Memory {
        #[command(subcommand)]
        action: MemoryCommands,
    },

    /// Node management
    Node {
        #[command(subcommand)]
        action: NodeCommands,
    },

    /// Configuration operations
    Config {
        #[command(subcommand)]
        action: ConfigCommands,
    },

    /// System information
    Info,
}

#[derive(Subcommand, Debug)]
enum TaskCommands {
    /// Create a new task
    Create {
        /// Task title
        title: String,

        /// Task description
        #[arg(short, long)]
        description: Option<String>,

        /// Task group name
        #[arg(short, long)]
        group: Option<String>,

        /// Workspace directory
        #[arg(short, long)]
        workspace: Option<String>,

        /// Enable sandbox
        #[arg(long)]
        sandboxed: bool,

        /// Allow network access
        #[arg(long)]
        allow_network: bool,
    },

    /// Execute a task with streaming output
    Run {
        /// Task ID
        id: i64,

        /// Wait for completion (default: true)
        #[arg(short, long, default_value_t = true)]
        wait: bool,

        /// Stream output in real-time
        #[arg(short, long)]
        stream: bool,
    },

    /// List all tasks
    List {
        /// Filter by status
        #[arg(short, long)]
        status: Option<String>,

        /// Filter by group
        #[arg(short, long)]
        group: Option<String>,

        /// Maximum number of tasks to show
        #[arg(short, long, default_value_t = 20)]
        limit: usize,

        /// Output format: text, json, table
        #[arg(short, long, default_value = "text")]
        format: String,
    },

    /// Show task details
    Show {
        /// Task ID
        id: i64,
    },

    /// Cancel a running task
    Cancel {
        /// Task ID
        id: i64,
    },
}

#[derive(Subcommand, Debug)]
enum MemoryCommands {
    /// Search memory store
    Search {
        /// Search query
        query: String,

        /// Maximum results
        #[arg(short, long, default_value_t = 10)]
        limit: usize,

        /// Output format: text, json
        #[arg(short, long, default_value = "text")]
        format: String,
    },

    /// Show memory statistics
    Stats {
        /// Output format: text, json
        #[arg(short, long, default_value = "text")]
        format: String,
    },
}

#[derive(Subcommand, Debug)]
enum NodeCommands {
    /// List all registered nodes
    List {
        /// Output format: text, json, table
        #[arg(short, long, default_value = "text")]
        format: String,
    },

    /// Show node status
    Status {
        /// Node ID
        id: String,

        /// Output format: text, json
        #[arg(short, long, default_value = "text")]
        format: String,
    },
}

#[derive(Subcommand, Debug)]
enum ConfigCommands {
    /// Show current configuration
    Show {
        /// Show sensitive values
        #[arg(long)]
        reveal: bool,

        /// Output format: text, json, toml
        #[arg(short, long, default_value = "text")]
        format: String,
    },

    /// Validate configuration file
    Validate {
        /// Configuration file path
        file: Option<String>,
    },

    /// Generate default configuration
    Init {
        /// Output path
        #[arg(short, long)]
        output: Option<String>,

        /// Overwrite existing file
        #[arg(long)]
        force: bool,
    },
}
```

### 2. Configuration System

#### Config File Locations (Priority Order)

1. `--config <PATH>` (command-line override)
2. `./agentflow.toml` (project-specific)
3. `~/.agentflow/config.toml` (user-specific)
4. Built-in defaults

#### TOML Configuration Structure

```toml
# agentflow.toml.example

[server]
# Server mode: "local" | "cloud" | "planner-only"
mode = "local"
# Bind address
host = "0.0.0.0"
# Server port
port = 6767
# Log level: "trace" | "debug" | "info" | "warn" | "error"
log_level = "info"

[database]
# Database URL (SQLite)
url = "sqlite://agentflow.db"
# Maximum connections
max_connections = 10

[cli]
# CLI worker configuration (local mode)

[cli.worker_safe]
# Command to execute for worker tasks
command = "claude"
# Arguments to pass
args = [
    "-p",
    "--dangerously-skip-permissions",
    "--allowedTools", "Read,Edit,Bash(git *),Write,Glob"
]
# Working directory (default: current directory)
working_dir = null

[cli.planner]
# Command for lightweight planning (cloud/planner-only mode)
command = "claude"
# Arguments for planning
args = [
    "-p",
    "--model", "claude-3-haiku"
]

[sandbox]
# Sandbox configuration
enabled = true
# Default workspace directory
default_workspace = "/tmp/agentflow/workspace"
# Whitelist of allowed directories (empty = all allowed)
whitelist_dirs = [
    # "/home/user/repos",
    # "/Users/user/repos"
]
# Enable symlink protection
enable_symlink_protection = true
# Maximum task execution time (seconds)
max_task_seconds = 1800
# Allow network access by default
allow_network = false
# Maximum memory limit (e.g., "1G", "512M")
max_memory = "1G"
# Maximum CPU cores
max_cpu = 2

[memory]
# Memory backend: "memory" | "sqlite" | "redis"
backend = "memory"
# Default TTL for memory entries (seconds)
default_ttl = 3600
# Maximum number of entries
max_entries = 10000
# Enable persistence
enable_persistence = false
# Database URL (for sqlite backend)
database_url = "sqlite://agentflow_memory.db"
# Redis URL (for redis backend)
redis_url = null

[executor]
# Task executor configuration
max_concurrent_tasks = 10
# Task timeout (seconds)
task_timeout = 300
# Worker heartbeat timeout (seconds)
worker_heartbeat_timeout = 60

[webhook]
# Webhook configuration (cloud mode)
# Webhook secret for authentication
secret = "change-me-in-production"
# Rate limit per IP (requests per minute)
rate_limit_per_ip = 10
# Allowed actions (empty = all allowed)
allowed_actions = [
    "run_test",
    "benchmark",
    "status"
]

[cloud]
# Cloud mode configuration
# Node discovery endpoint
discovery_endpoint = null
# Enable node registration
enable_registration = true
# Heartbeat interval (seconds)
heartbeat_interval = 30
# Node timeout (seconds)
node_timeout = 120
```

#### Rust Configuration Structures

```rust
// agentflow-core/src/config/mod.rs

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::path::{Path, PathBuf};

/// Main configuration structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub database: DatabaseConfig,
    pub cli: CliConfig,
    pub sandbox: SandboxConfig,
    pub memory: MemoryConfig,
    pub executor: ExecutorConfig,
    #[serde(default)]
    pub webhook: WebhookConfig,
    #[serde(default)]
    pub cloud: CloudConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerConfig {
    /// Server mode: local, cloud, or planner-only
    #[serde(default = "default_server_mode")]
    pub mode: ServerMode,

    /// Bind address
    #[serde(default = "default_server_host")]
    pub host: String,

    /// Server port
    #[serde(default = "default_server_port")]
    pub port: u16,

    /// Log level
    #[serde(default = "default_log_level")]
    pub log_level: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ServerMode {
    Local,
    Cloud,
    PlannerOnly,
}

impl Default for ServerMode {
    fn default() -> Self {
        ServerMode::Local
    }
}

impl std::str::FromStr for ServerMode {
    type Err = anyhow::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "local" => Ok(ServerMode::Local),
            "cloud" => Ok(ServerMode::Cloud),
            "planner-only" | "planner_only" => Ok(ServerMode::PlannerOnly),
            _ => Err(anyhow::anyhow!("Invalid server mode: {}", s)),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CliConfig {
    #[serde(default)]
    pub worker_safe: CommandConfig,
    #[serde(default)]
    pub planner: CommandConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CommandConfig {
    #[serde(default = "default_command")]
    pub command: String,

    #[serde(default)]
    pub args: Vec<String>,

    pub working_dir: Option<PathBuf>,
}

impl Default for CommandConfig {
    fn default() -> Self {
        Self {
            command: default_command(),
            args: Vec::new(),
            working_dir: None,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DatabaseConfig {
    #[serde(default = "default_database_url")]
    pub url: String,

    #[serde(default = "default_max_connections")]
    pub max_connections: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SandboxConfig {
    #[serde(default = "default_sandbox_enabled")]
    pub enabled: bool,

    #[serde(default = "default_workspace")]
    pub default_workspace: PathBuf,

    #[serde(default)]
    pub whitelist_dirs: Vec<PathBuf>,

    #[serde(default = "default_symlink_protection")]
    pub enable_symlink_protection: bool,

    #[serde(default = "default_max_task_seconds")]
    pub max_task_seconds: u64,

    #[serde(default)]
    pub allow_network: bool,

    pub max_memory: Option<String>,

    pub max_cpu: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryConfig {
    #[serde(default = "default_memory_backend")]
    pub backend: String,

    #[serde(default = "default_memory_ttl")]
    pub default_ttl: u64,

    #[serde(default = "default_memory_entries")]
    pub max_entries: usize,

    #[serde(default)]
    pub enable_persistence: bool,

    pub database_url: Option<String>,

    pub redis_url: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutorConfig {
    #[serde(default = "default_concurrent_tasks")]
    pub max_concurrent_tasks: usize,

    #[serde(default = "default_task_timeout")]
    pub task_timeout: u64,

    #[serde(default = "default_heartbeat_timeout")]
    pub worker_heartbeat_timeout: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct WebhookConfig {
    pub secret: Option<String>,

    #[serde(default = "default_rate_limit")]
    pub rate_limit_per_ip: usize,

    #[serde(default)]
    pub allowed_actions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct CloudConfig {
    pub discovery_endpoint: Option<String>,

    #[serde(default = "default_true")]
    pub enable_registration: bool,

    #[serde(default = "default_heartbeat_interval")]
    pub heartbeat_interval: u64,

    #[serde(default = "default_node_timeout")]
    pub node_timeout: u64,
}

// Default value functions
fn default_server_mode() -> ServerMode { ServerMode::Local }
fn default_server_host() -> String { "0.0.0.0".to_string() }
fn default_server_port() -> u16 { 6767 }
fn default_log_level() -> String { "info".to_string() }
fn default_command() -> String { "claude".to_string() }
fn default_database_url() -> String { "sqlite://agentflow.db".to_string() }
fn default_max_connections() -> u32 { 10 }
fn default_sandbox_enabled() -> bool { true }
fn default_workspace() -> PathBuf { PathBuf::from("/tmp/agentflow/workspace") }
fn default_symlink_protection() -> bool { true }
fn default_max_task_seconds() -> u64 { 1800 }
fn default_memory_backend() -> String { "memory".to_string() }
fn default_memory_ttl() -> u64 { 3600 }
fn default_memory_entries() -> usize { 10000 }
fn default_concurrent_tasks() -> usize { 10 }
fn default_task_timeout() -> u64 { 300 }
fn default_heartbeat_timeout() -> u64 { 60 }
fn default_rate_limit() -> usize { 10 }
fn default_true() -> bool { true }
fn default_heartbeat_interval() -> u64 { 30 }
fn default_node_timeout() -> u64 { 120 }

impl Default for Config {
    fn default() -> Self {
        Self {
            server: ServerConfig {
                mode: default_server_mode(),
                host: default_server_host(),
                port: default_server_port(),
                log_level: default_log_level(),
            },
            database: DatabaseConfig {
                url: default_database_url(),
                max_connections: default_max_connections(),
            },
            cli: CliConfig::default(),
            sandbox: SandboxConfig {
                enabled: default_sandbox_enabled(),
                default_workspace: default_workspace(),
                whitelist_dirs: Vec::new(),
                enable_symlink_protection: default_symlink_protection(),
                max_task_seconds: default_max_task_seconds(),
                allow_network: false,
                max_memory: Some("1G".to_string()),
                max_cpu: Some(2),
            },
            memory: MemoryConfig {
                backend: default_memory_backend(),
                default_ttl: default_memory_ttl(),
                max_entries: default_memory_entries(),
                enable_persistence: false,
                database_url: Some("sqlite://agentflow_memory.db".to_string()),
                redis_url: None,
            },
            executor: ExecutorConfig {
                max_concurrent_tasks: default_concurrent_tasks(),
                task_timeout: default_task_timeout(),
                worker_heartbeat_timeout: default_heartbeat_timeout(),
            },
            webhook: WebhookConfig::default(),
            cloud: CloudConfig::default(),
        }
    }
}

impl Config {
    /// Load configuration from file with fallback to defaults
    pub fn load(path: Option<&Path>) -> Result<Self> {
        // Try multiple locations in order
        let config_paths = [
            path.map(PathBuf::from),
            Some(PathBuf::from("./agentflow.toml")),
            Self::user_config_path()?,
        ];

        for config_path in config_paths.into_iter().flatten() {
            if config_path.exists() {
                let content = std::fs::read_to_string(&config_path)
                    .with_context(|| format!("Failed to read config file: {}", config_path.display()))?;

                let mut config: Config = toml::from_str(&content)
                    .with_context(|| format!("Failed to parse config file: {}", config_path.display()))?;

                // Apply environment variable overrides
                config.apply_env_overrides();

                return Ok(config);
            }
        }

        // No config file found, use defaults
        Ok(Config::default())
    }

    /// Get user config path (~/.agentflow/config.toml)
    fn user_config_path() -> Result<Option<PathBuf>> {
        let home = dirs::home_dir()
            .ok_or_else(|| anyhow::anyhow!("Cannot determine home directory"))?;

        let config_dir = home.join(".agentflow");
        Ok(Some(config_dir.join("config.toml")))
    }

    /// Apply environment variable overrides
    fn apply_env_overrides(&mut self) {
        // Server overrides
        if let Ok(mode) = std::env::var("AGENTFLOW_SERVER_MODE") {
            if let Ok(mode) = mode.parse() {
                self.server.mode = mode;
            }
        }
        if let Ok(port) = std::env::var("AGENTFLOW_SERVER_PORT") {
            if let Ok(port) = port.parse() {
                self.server.port = port;
            }
        }
        if let Ok(host) = std::env::var("AGENTFLOW_SERVER_HOST") {
            self.server.host = host;
        }

        // Database overrides
        if let Ok(url) = std::env::var("AGENTFLOW_DATABASE_URL") {
            self.database.url = url;
        }

        // Log level override
        if let Ok(level) = std::env::var("AGENTFLOW_LOG_LEVEL") {
            self.server.log_level = level;
        }

        // Webhook secret override
        if let Ok(secret) = std::env::var("AGENTFLOW_WEBHOOK_SECRET") {
            self.webhook.secret = Some(secret);
        }
    }

    /// Save configuration to file
    pub fn save(&self, path: &Path) -> Result<()> {
        // Create parent directory if needed
        if let Some(parent) = path.parent() {
            std::fs::create_dir_all(parent)
                .with_context(|| format!("Failed to create config directory: {}", parent.display()))?;
        }

        let content = toml::to_string_pretty(self)
            .context("Failed to serialize config")?;

        std::fs::write(path, content)
            .with_context(|| format!("Failed to write config file: {}", path.display()))?;

        Ok(())
    }

    /// Validate configuration
    pub fn validate(&self) -> Result<()> {
        // Validate server mode
        match self.server.mode {
            ServerMode::Local | ServerMode::Cloud | ServerMode::PlannerOnly => {},
        }

        // Validate port range
        if self.server.port == 0 || self.server.port > 65535 {
            anyhow::bail!("Invalid port number: {}", self.server.port);
        }

        // Validate workspace exists if sandbox enabled
        if self.sandbox.enabled && !self.sandbox.default_workspace.exists() {
            anyhow::bail!("Sandbox workspace does not exist: {}",
                self.sandbox.default_workspace.display());
        }

        // Validate database URL
        if !self.database.url.starts_with("sqlite://") {
            anyhow::bail!("Only SQLite databases are supported");
        }

        Ok(())
    }

    /// Get bind address
    pub fn bind_address(&self) -> String {
        format!("{}:{}", self.server.host, self.server.port)
    }

    /// Convert to legacy MasterConfig for backward compatibility
    pub fn to_master_config(&self) -> crate::config::MasterConfig {
        crate::config::MasterConfig {
            server_addr: self.server.host.clone(),
            server_port: self.server.port,
            database_url: self.database.url.clone(),
            sandbox: crate::config::SandboxConfig {
                enabled: self.sandbox.enabled,
                default_workspace: self.sandbox.default_workspace
                    .to_string_lossy().to_string(),
                allow_network: self.sandbox.allow_network,
                max_memory: self.sandbox.max_memory.clone(),
                max_cpu: self.sandbox.max_cpu,
                timeout: Some(self.sandbox.max_task_seconds),
            },
            memory: crate::config::MemoryConfig {
                backend: self.memory.backend.clone(),
                database_url: self.memory.database_url.clone(),
                redis_url: self.memory.redis_url.clone(),
                default_ttl: self.memory.default_ttl,
                max_entries: self.memory.max_entries,
                enable_persistence: self.memory.enable_persistence,
            },
            log_level: self.server.log_level.clone(),
            worker_heartbeat_timeout: self.executor.worker_heartbeat_timeout,
            task_timeout: self.executor.task_timeout,
            max_concurrent_tasks: self.executor.max_concurrent_tasks,
        }
    }
}
```

### 3. API Endpoints (Mock Implementation)

#### New Routes

```rust
// agentflow-master/src/routes/nodes.rs (NEW)

use axum::{Json, State};
use serde_json::json;
use super::AppState;

/// List all registered nodes (mock for now)
///
/// GET /api/v1/nodes
pub async fn list_nodes(State(_state): State<AppState>) -> Json<serde_json::Value> {
    Json(json!({
        "success": true,
        "data": {
            "nodes": [],
            "total": 0,
            "mode": "local"
        },
        "message": "Node listing not yet implemented - Team C will implement in v0.2.1"
    }))
}

/// Get cluster status (mock for now)
///
/// GET /api/v1/cluster/status
pub async fn cluster_status(State(_state): State<AppState>) -> Json<serde_json::Value> {
    Json(json!({
        "success": true,
        "data": {
            "status": "local",
            "nodes": {
                "total": 0,
                "active": 0,
                "inactive": 0
            },
            "tasks": {
                "running": 0,
                "pending": 0,
                "completed": 0
            }
        },
        "message": "Cluster status not yet implemented - Team C will implement in v0.2.1"
    }))
}

/// Get current configuration
///
/// GET /api/v1/config
pub async fn get_config(State(state): State<AppState>) -> Json<serde_json::Value> {
    // Return sanitized config (hide secrets)
    Json(json!({
        "success": true,
        "data": {
            "server": {
                "mode": "local",
                "host": "0.0.0.0",
                "port": 6767
            },
            "executor": {
                "max_concurrent_tasks": state.executor.max_concurrent_tasks
            },
            "memory": {
                "backend": "memory",
                "max_entries": 10000
            }
        }
    }))
}
```

#### Route Integration

```rust
// In agentflow-master/src/routes/mod.rs

pub mod nodes; // NEW

pub fn create_routes() -> Router<AppState> {
    Router::new()
        // ... existing routes ...
        // Node management (mock)
        .route("/api/v1/nodes", get(nodes::list_nodes))
        .route("/api/v1/cluster/status", get(nodes::cluster_status))
        .route("/api/v1/config", get(nodes::get_config))
}
```

### 4. CLI Command Implementation

```rust
// agentflow-master/src/cli.rs (NEW)

use anyhow::Result;
use clap::Parser;
use reqwest::Client;
use serde_json::json;
use std::path::PathBuf;

use crate::config::{Config, ServerMode};

/// AgentFlow CLI - AI Agent Task Collaboration System
#[derive(Parser, Debug)]
#[command(name = "agentflow")]
#[command(about = "AgentFlow - AI Agent Task Collaboration System", long_about = None)]
#[command(version = env!("CARGO_PKG_VERSION"))]
pub struct Cli {
    /// Configuration file path
    #[arg(short, long, global = true)]
    pub config: Option<PathBuf>,

    /// Verbose output
    #[arg(short, long, global = true)]
    pub verbose: bool,

    #[command(subcommand)]
    pub command: Commands,
}

#[derive(clap::Subcommand, Debug)]
pub enum Commands {
    /// Start AgentFlow server
    Server {
        /// Server mode: local, cloud, or planner-only
        #[arg(short, long, default_value = "local")]
        mode: String,

        /// Server port
        #[arg(short, long)]
        port: Option<u16>,

        /// Bind address
        #[arg(short, long)]
        addr: Option<String>,
    },

    /// Task management
    Task {
        #[command(subcommand)]
        action: TaskCommands,
    },

    /// Memory operations
    Memory {
        #[command(subcommand)]
        action: MemoryCommands,
    },

    /// Node management
    Node {
        #[command(subcommand)]
        action: NodeCommands,
    },

    /// Configuration operations
    Config {
        #[command(subcommand)]
        action: ConfigCommands,
    },

    /// System information
    Info,
}

#[derive(clap::Subcommand, Debug)]
pub enum TaskCommands {
    Create {
        title: String,
        #[arg(short, long)]
        description: Option<String>,
        #[arg(short, long)]
        group: Option<String>,
        #[arg(short, long)]
        workspace: Option<String>,
        #[arg(long)]
        sandboxed: bool,
        #[arg(long)]
        allow_network: bool,
    },
    Run {
        id: i64,
        #[arg(short, long, default_value_t = true)]
        wait: bool,
        #[arg(short, long)]
        stream: bool,
    },
    List {
        #[arg(short, long)]
        status: Option<String>,
        #[arg(short, long)]
        group: Option<String>,
        #[arg(short, long, default_value_t = 20)]
        limit: usize,
        #[arg(short, long, default_value = "text")]
        format: String,
    },
    Show {
        id: i64,
    },
    Cancel {
        id: i64,
    },
}

#[derive(clap::Subcommand, Debug)]
pub enum MemoryCommands {
    Search {
        query: String,
        #[arg(short, long, default_value_t = 10)]
        limit: usize,
        #[arg(short, long, default_value = "text")]
        format: String,
    },
    Stats {
        #[arg(short, long, default_value = "text")]
        format: String,
    },
}

#[derive(clap::Subcommand, Debug)]
pub enum NodeCommands {
    List {
        #[arg(short, long, default_value = "text")]
        format: String,
    },
    Status {
        id: String,
        #[arg(short, long, default_value = "text")]
        format: String,
    },
}

#[derive(clap::Subcommand, Debug)]
pub enum ConfigCommands {
    Show {
        #[arg(long)]
        reveal: bool,
        #[arg(short, long, default_value = "text")]
        format: String,
    },
    Validate {
        file: Option<PathBuf>,
    },
    Init {
        #[arg(short, long)]
        output: Option<PathBuf>,
        #[arg(long)]
        force: bool,
    },
}

impl Cli {
    /// Execute CLI command
    pub async fn execute(self) -> Result<()> {
        match self.command {
            Commands::Server { mode, port, addr } => {
                self.start_server(mode, port, addr).await
            },
            Commands::Task { action } => {
                self.handle_task(action).await
            },
            Commands::Memory { action } => {
                self.handle_memory(action).await
            },
            Commands::Node { action } => {
                self.handle_node(action).await
            },
            Commands::Config { action } => {
                self.handle_config(action).await
            },
            Commands::Info => {
                self.show_info().await
            },
        }
    }

    /// Start server (delegates to main.rs)
    async fn start_server(&self, mode: String, port: Option<u16>, addr: Option<String>) -> Result<()> {
        // Load config
        let config_path = self.config.as_deref();
        let mut config = Config::load(config_path)?;

        // Apply CLI overrides
        if let Ok(mode) = mode.parse() {
            config.server.mode = mode;
        }
        if let Some(port) = port {
            config.server.port = port;
        }
        if let Some(addr) = addr {
            config.server.host = addr;
        }

        // Validate config
        config.validate()?;

        // Delegate to server startup logic in main.rs
        // This is a placeholder - actual implementation in main.rs
        println!("Starting AgentFlow server in {} mode on {}",
            config.server.mode.as_str(),
            config.bind_address()
        );

        Ok(())
    }

    /// Handle task commands
    async fn handle_task(&self, action: TaskCommands) -> Result<()> {
        let client = Client::new();
        let base_url = "http://127.0.0.1:6767";

        match action {
            TaskCommands::Create { title, description, group, workspace, sandboxed, allow_network } => {
                let req = json!({
                    "title": title,
                    "description": description,
                    "group_name": group,
                    "workspace_dir": workspace,
                    "sandboxed": sandboxed,
                    "allow_network": allow_network
                });

                let resp = client.post(&format!("{}/api/v1/tasks", base_url))
                    .json(&req)
                    .send()
                    .await?;

                let result: serde_json::Value = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result)?);
            },
            TaskCommands::Run { id, wait, stream } => {
                println!("Running task {} (wait={}, stream={})", id, wait, stream);
                // TODO: Implement streaming execution
            },
            TaskCommands::List { status, group, limit, format } => {
                let resp = client.get(&format!("{}/api/v1/tasks", base_url))
                    .send()
                    .await?;

                let result: serde_json::Value = resp.json().await?;

                match format.as_str() {
                    "json" => println!("{}", serde_json::to_string_pretty(&result)?),
                    "table" => {
                        // TODO: Implement table formatting
                        println!("{}", serde_json::to_string_pretty(&result)?);
                    },
                    _ => println!("{}", serde_json::to_string_pretty(&result)?),
                }
            },
            TaskCommands::Show { id } => {
                let resp = client.get(&format!("{}/api/v1/tasks/{}", base_url, id))
                    .send()
                    .await?;

                let result: serde_json::Value = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result)?);
            },
            TaskCommands::Cancel { id } => {
                let resp = client.post(&format!("{}/api/v1/tasks/{}/cancel", base_url, id))
                    .send()
                    .await?;

                let result: serde_json::Value = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result)?);
            },
        }

        Ok(())
    }

    /// Handle memory commands
    async fn handle_memory(&self, action: MemoryCommands) -> Result<()> {
        let client = Client::new();
        let base_url = "http://127.0.0.1:6767";

        match action {
            MemoryCommands::Search { query, limit, format } => {
                let resp = client.get(&format!("{}/api/v1/memory/search?q={}&limit={}", base_url, query, limit))
                    .send()
                    .await?;

                let result: serde_json::Value = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result)?);
            },
            MemoryCommands::Stats { format } => {
                let resp = client.get(&format!("{}/api/v1/memory/stats", base_url))
                    .send()
                    .await?;

                let result: serde_json::Value = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result)?);
            },
        }

        Ok(())
    }

    /// Handle node commands
    async fn handle_node(&self, action: NodeCommands) -> Result<()> {
        let client = Client::new();
        let base_url = "http://127.0.0.1:6767";

        match action {
            NodeCommands::List { format } => {
                let resp = client.get(&format!("{}/api/v1/nodes", base_url))
                    .send()
                    .await?;

                let result: serde_json::Value = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result)?);
            },
            NodeCommands::Status { id, format } => {
                println!("Node status for {} (not yet implemented)", id);
            },
        }

        Ok(())
    }

    /// Handle config commands
    async fn handle_config(&self, action: ConfigCommands) -> Result<()> {
        match action {
            ConfigCommands::Show { reveal, format } => {
                let config_path = self.config.as_deref();
                let config = Config::load(config_path)?;

                match format.as_str() {
                    "json" => println!("{}", serde_json::to_string_pretty(&config)?),
                    "toml" => println!("{}", toml::to_string_pretty(&config)?),
                    _ => println!("{}", toml::to_string_pretty(&config)?),
                }
            },
            ConfigCommands::Validate { file } => {
                let path = file.or(self.config.clone()).unwrap_or_else(|| PathBuf::from("./agentflow.toml"));
                let config = Config::load(Some(&path))?;
                config.validate()?;
                println!("✅ Configuration file is valid: {}", path.display());
            },
            ConfigCommands::Init { output, force } => {
                let output_path = output.unwrap_or_else(|| PathBuf::from("./agentflow.toml"));

                if output_path.exists() && !force {
                    anyhow::bail!("Config file already exists: {}. Use --force to overwrite.",
                        output_path.display());
                }

                let config = Config::default();
                config.save(&output_path)?;
                println!("✅ Configuration file created: {}", output_path.display());
            },
        }

        Ok(())
    }

    /// Show system information
    async fn show_info(&self) -> Result<()> {
        println!("AgentFlow v{}", env!("CARGO_PKG_VERSION"));
        println!("Rust Edition v3");
        println!();
        println!("Configuration:");
        println!("  Config file: {:?}", self.config);
        println!("  Verbose: {}", self.verbose);
        println!();
        println!("For more information, run: agentflow --help");

        Ok(())
    }
}

impl ServerMode {
    pub fn as_str(&self) -> &'static str {
        match self {
            ServerMode::Local => "local",
            ServerMode::Cloud => "cloud",
            ServerMode::PlannerOnly => "planner-only",
        }
    }
}
```

---

## 🔄 Migration Strategy

### Phase 1: Add Dependencies

```toml
# rust/Cargo.toml (workspace dependencies)

toml = "0.8"
dirs = "5.0"

# rust/agentflow-core/Cargo.toml

toml = { workspace = true }
dirs = { workspace = true }

# rust/agentflow-master/Cargo.toml

# Already has:
# - clap (workspace)
# Need to add:
reqwest = { workspace = true }  # For CLI HTTP client
```

### Phase 2: Create Configuration Module

1. Create `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/config/mod.rs`
2. Implement TOML-based configuration system
3. Add migration layer for backward compatibility with existing `config.rs`

### Phase 3: Implement CLI

1. Create `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/cli.rs`
2. Modify `main.rs` to use CLI parser
3. Keep existing server startup logic, just change entry point

### Phase 4: Add Mock Endpoints

1. Create `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/nodes.rs`
2. Add routes to `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/mod.rs`

### Phase 5: Testing

1. Test all CLI commands
2. Verify backward compatibility
3. Test configuration loading
4. Test mock API endpoints

---

## 📦 Deliverables

### Code Files

1. **`/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/config/mod.rs`**
   - New configuration module with TOML support
   - Multi-mode configuration structures
   - Configuration validation

2. **`/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/cli.rs`**
   - Unified CLI entry point
   - All subcommands implemented
   - HTTP client for API calls

3. **`/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/nodes.rs`**
   - Mock node listing endpoint
   - Mock cluster status endpoint
   - Configuration endpoint

4. **Modified Files**:
   - `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/main.rs`
   - `/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/routes/mod.rs`
   - `/Users/jiangxiaolong/work/project/AgentFlow/rust/Cargo.toml`

### Documentation

5. **`/Users/jiangxiaolong/work/project/AgentFlow/agentflow.toml.example`**
   - Complete example configuration file
   - Inline documentation for all options

6. **Updated README.md sections**
   - CLI usage examples
   - Configuration guide
   - Multi-mode setup instructions

### Tests

7. **Unit Tests**
   - Configuration loading tests
   - CLI parser tests
   - Configuration validation tests

---

## 🧪 Testing Strategy

### Unit Tests

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_config_default() {
        let config = Config::default();
        assert_eq!(config.server.mode, ServerMode::Local);
        assert_eq!(config.server.port, 6767);
    }

    #[test]
    fn test_config_from_file() {
        let config = Config::load(Some(Path::new("tests/fixtures/config.toml"))).unwrap();
        assert_eq!(config.server.mode, ServerMode::Cloud);
    }

    #[test]
    fn test_config_validation() {
        let mut config = Config::default();
        config.server.port = 0;
        assert!(config.validate().is_err());
    }

    #[test]
    fn test_server_mode_parsing() {
        assert_eq!("local".parse::<ServerMode>().unwrap(), ServerMode::Local);
        assert_eq!("cloud".parse::<ServerMode>().unwrap(), ServerMode::Cloud);
        assert!("invalid".parse::<ServerMode>().is_err());
    }
}
```

### Integration Tests

1. **Configuration Loading**
   - Test loading from various locations
   - Test environment variable overrides
   - Test validation

2. **CLI Commands**
   - Test all subcommands
   - Test API interaction
   - Test output formatting

3. **Server Startup**
   - Test all three modes
   - Test port binding
   - Test graceful shutdown

---

## ⚠️ Risks & Mitigation

### Risk 1: Breaking Changes

**Risk**: New CLI might break existing workflows
**Mitigation**:
- Keep `agentflow-master` binary working as before
- Provide migration guide
- Support both interfaces during transition

### Risk 2: Configuration Complexity

**Risk**: TOML configuration might be too complex
**Mitigation**:
- Provide sensible defaults
- Create `agentflow config init` command
- Document all options thoroughly

### Risk 3: Dependency Conflicts

**Risk**: Adding `toml` and `dirs` might conflict
**Mitigation**:
- Use workspace dependencies
- Pin versions in Cargo.toml
- Test on multiple platforms

### Risk 4: Mock Endpoints Confusion

**Risk**: Users might think mock endpoints are real
**Mitigation**:
- Clear documentation that endpoints are placeholders
- Return descriptive error messages
- Add "not yet implemented" notes

---

## 📅 Timeline

### Week 1 (Current)

**Days 1-2**: Research & Planning ✅
- Analyze existing codebase
- Design architecture
- Create implementation plan (this document)

**Days 3-4**: Configuration System
- Implement `config/mod.rs`
- Add TOML dependency
- Create migration layer

**Days 5-7**: CLI Implementation
- Implement `cli.rs`
- Modify `main.rs`
- Add HTTP client commands

### Week 2

**Days 1-2**: API Endpoints
- Create `routes/nodes.rs`
- Add mock endpoints
- Test API responses

**Days 3-4**: Testing
- Unit tests
- Integration tests
- Manual testing

**Days 5-7**: Documentation
- Create `agentflow.toml.example`
- Update README
- Write usage guide

---

## 🎯 Success Criteria

### Functional Requirements

- ✅ `agentflow server local` starts local mode server
- ✅ `agentflow server cloud` starts cloud mode server (with mock endpoints)
- ✅ `agentflow task create "test"` creates task via API
- ✅ `agentflow task list` lists tasks
- ✅ `agentflow config init` generates default config
- ✅ `agentflow config validate` validates configuration

### Non-Functional Requirements

- ✅ Zero breaking changes to existing functionality
- ✅ All tests passing
- ✅ Clean compilation without warnings
- ✅ Complete documentation
- ✅ Backward compatibility maintained

### Quality Metrics

- Code coverage: >80%
- Documentation coverage: 100%
- Performance: <100ms CLI response time
- Security: No sensitive data in logs

---

## 📚 References

### Existing Documentation

- `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/v0.2.1迭代计划.md` - Overall iteration plan
- `/Users/jiangxiaolong/work/project/AgentFlow/docs/plan/v0.2.1_TASK_BREAKDOWN.md` - Task breakdown
- `/Users/jiangxiaolong/work/project/AgentFlow/docs/TEAM_A_IMPLEMENTATION_REPORT.md` - Previous Team A work

### External Resources

- [Clap Documentation](https://docs.rs/clap/latest/clap/)
- [TOML crate](https://docs.rs/toml/latest/toml/)
- [Dirs crate](https://docs.rs/dirs/latest/dirs/)

---

## 🤝 Integration with Other Teams

### Team B (Memory & Git)

- **Dependency**: Team B may need configuration for memory backend
- **Handoff**: Provide `memory` section in config
- **Timeline**: Week 2, after config system is stable

### Team C (Cloud/Edge)

- **Dependency**: Team C will implement real node endpoints
- **Handoff**: Mock endpoints with clear documentation
- **Timeline**: Week 2-3, replace mock implementations

### Team D (Packaging)

- **Dependency**: Team D needs `agentflow.toml.example` for installation
- **Handoff**: Complete configuration example
- **Timeline**: Week 2, provide final config template

---

## 📝 Notes

### Design Decisions

1. **Why TOML over JSON/YAML?**
   - TOML is more readable for configuration
   - Rust ecosystem has excellent TOML support
   - Similar to Cargo.toml format users know

2. **Why separate `agentflow-core` config?**
   - Keeps configuration logic separate from server
   - Allows reuse in future worker binaries
   - Cleaner separation of concerns

3. **Why mock endpoints instead of waiting?**
   - Unblocks other teams
   - Provides API contract upfront
   - Allows CLI testing without full implementation

4. **Why keep `agentflow-master` binary?**
   - Backward compatibility
   - Allows gradual migration
   - Supports existing deployments

### Future Enhancements

1. **Shell Completion**: Add completions for bash/zsh/fish
2. **Config Profiles**: Support multiple configuration profiles
3. **Interactive Mode**: Add `agentflow shell` for interactive CLI
4. **Output Formatting**: Support table, JSON, CSV output formats
5. **Plugin System**: Allow custom CLI commands via plugins

---

## ✅ Checklist

### Planning Phase (Current)
- [x] Analyze existing codebase
- [x] Design CLI structure
- [x] Design configuration system
- [x] Create implementation plan
- [ ] Review with team

### Implementation Phase (Next)
- [ ] Add `toml` and `dirs` dependencies
- [ ] Create `config/mod.rs`
- [ ] Implement `cli.rs`
- [ ] Modify `main.rs`
- [ ] Create `routes/nodes.rs`
- [ ] Create `agentflow.toml.example`
- [ ] Write unit tests
- [ ] Write integration tests

### Documentation Phase
- [ ] Update README.md
- [ ] Create CLI usage guide
- [ ] Document configuration options
- [ ] Add examples

### Testing Phase
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual testing
- [ ] Performance testing

---

**End of Implementation Plan**

Next steps:
1. Review and approve this plan
2. Start with Phase 1: Add dependencies
3. Begin implementation of configuration system
4. Progress through phases as outlined
