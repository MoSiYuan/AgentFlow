//! AgentFlow Master Server
//!
//! 主服务器，提供 HTTP API 和 WebSocket 服务。

pub mod broadcast;
pub mod auth_middleware;
pub mod claude;
pub mod config;
pub mod error;
pub mod executor;
// pub mod grpc;  // Temporarily disabled due to compilation issues
// pub mod leader;  // Temporarily disabled due to proto dependency
pub mod memory_core;
// pub mod monitor;  // Temporarily disabled due to proto dependency
// pub mod proto;  // Temporarily disabled due to compilation issues
pub mod routes;
// pub mod scheduler;  // Temporarily disabled due to proto dependency
pub mod webhook;
pub mod websocket;

// 新增：分布式执行模块
pub mod agent_comm;
pub mod distributed_lock;
pub mod worker_registry;

pub use broadcast::Broadcaster;
pub use claude::ClaudeExecutor;
pub use config::*;
pub use error::*;
pub use executor::*;
// pub use grpc::*;  // Temporarily disabled
// pub use leader::{LeaderNode, raft::RaftNode};  // Temporarily disabled
pub use memory_core::*;
// pub use monitor::*;  // Temporarily disabled
pub use routes::AppState;
// pub use scheduler::{TaskScheduler, dependency::TaskDependencyGraph, queue::PriorityTaskQueue};  // Temporarily disabled
pub use agent_comm::AgentCommunication;
pub use distributed_lock::DistributedLock;
pub use worker_registry::WorkerRegistry;