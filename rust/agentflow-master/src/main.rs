//! AgentFlow Master 服务器
//!
//! 单进程架构的 Master 服务器，提供 HTTP API 和 WebSocket 支持
//! 集成 TaskExecutor 和 MemoryCore，直接执行任务而无需远程 Worker

mod claude;
mod auth_middleware;
mod config;
mod error;
mod executor;
// mod leader;  // Temporarily disabled due to proto dependency
mod memory_core;
mod routes;
mod webhook;

use anyhow::{Context, Result};
use axum::{
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use tower_http::services::{ServeDir, ServeFile};
use chrono::Utc;
use clap::Parser;
use routes::AppState;
use sqlx::{Pool, Sqlite};
use std::net::SocketAddr;
use tokio::signal;
use tower_http::cors::{Any, CorsLayer};
use tracing::{error, info, warn};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use config::MasterConfig;
use executor::TaskExecutor;
use memory_core::MemoryCore;

/// AgentFlow Master 服务器命令行参数
#[derive(Parser, Debug)]
#[command(name = "agentflow-master")]
#[command(about = "AgentFlow Master Server - 单进程架构的任务调度服务器", long_about = None)]
struct Args {
    /// 配置文件路径
    #[arg(short, long)]
    config: Option<String>,

    /// 服务器绑定地址
    #[arg(short, long, default_value = "0.0.0.0")]
    addr: String,

    /// 服务器端口
    #[arg(short, long, default_value_t = 6767)]
    port: u16,

    /// 日志级别
    #[arg(short, long, default_value = "info")]
    log_level: String,

    /// 运行模式: "master" 或 "leader"
    #[arg(long, default_value = "master")]
    mode: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    // 解析命令行参数
    let args = Args::parse();

    // 初始化日志
    init_tracing(&args.log_level);

    info!("🚀 启动 AgentFlow Master 服务器");

    // 加载配置
    let mut config = MasterConfig::load().context("加载配置失败")?;

    // 命令行参数优先级高于配置文件
    config.server_addr = args.addr;
    config.server_port = args.port;
    config.log_level = args.log_level;

    // 根据模式启动不同的节点
    match args.mode.as_str() {
        "leader" => {
            anyhow::bail!("Leader 模式暂时禁用，因为缺少 proto 模块依赖");
        }
        "master" => {
            info!("🎯 启动 Master 模式");
        }
        _ => {
            anyhow::bail!("未知的运行模式: {}, 只支持 'master'", args.mode);
        }
    }

    info!("📋 配置加载完成");
    info!("   - 服务器地址: {}", config.bind_address());
    info!("   - 数据库: {}", config.database_url);
    info!("   - 最大并发任务: {}", config.max_concurrent_tasks);
    info!("   - 任务超时: {}秒", config.task_timeout);

    // 初始化数据库
    let db = init_database(&config).await.context("初始化数据库失败")?;
    info!("✅ 数据库初始化完成");

    // 创建任务执行器
    let executor = TaskExecutor::new(db.clone(), config.max_concurrent_tasks);
    info!("⚙️  任务执行器已创建");

    // 创建记忆核心
    let memory = MemoryCore::new(config.memory.default_ttl as i64, config.memory.max_entries);
    info!("🧠 记忆核心已创建");

    // 创建 Session 存储
    let session_store = auth_middleware::SessionStore::new();
    info!("🔐 Session 存储已创建");

    // 创建应用状态
    let app_state = AppState {
        executor: executor.clone(),
        memory: memory.clone(),
        auth_config: config.auth.clone(),
        session_store,
        start_time: Utc::now(),
    };

    // 创建路由
    let app = create_app(app_state).await?;

    // 绑定地址
    let addr: SocketAddr = config
        .bind_address()
        .parse()
        .context("无效的服务器地址")?;

    info!("🌐 服务器监听: http://{}", addr);

    // 启动服务器
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .context("绑定端口失败")?;

    // 启动后台清理任务
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(tokio::time::Duration::from_secs(300)); // 每5分钟清理一次
        loop {
            interval.tick().await;
            if let Err(e) = memory.cleanup_expired().await {
                warn!("清理过期记忆失败: {}", e);
            }
        }
    });

    // 启动服务器并等待关闭信号
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
        .context("服务器运行失败")?;

    info!("👋 AgentFlow Master 服务器已关闭");

    Ok(())
}

/// 初始化日志系统
fn init_tracing(log_level: &str) {
    let filter = tracing_subscriber::EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new(log_level));

    tracing_subscriber::registry()
        .with(filter)
        .with(tracing_subscriber::fmt::layer())
        .init();
}

/// 初始化数据库
async fn init_database(config: &MasterConfig) -> Result<Pool<Sqlite>> {
    let pool = sqlx::sqlite::SqlitePoolOptions::new()
        .max_connections(10)
        .connect_with(config.pool_options())
        .await
        .context("连接数据库失败")?;

    // 运行数据库迁移
    sqlx::query(
        r#"
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id TEXT NOT NULL UNIQUE,
            parent_id INTEGER,
            title TEXT NOT NULL,
            description TEXT,
            group_name TEXT NOT NULL DEFAULT 'default',
            completion_criteria TEXT,
            status TEXT NOT NULL DEFAULT 'pending',
            priority INTEGER NOT NULL DEFAULT 1,
            lock_holder TEXT,
            lock_time TEXT,
            result TEXT,
            error TEXT,
            workspace_dir TEXT,
            sandboxed INTEGER NOT NULL DEFAULT 0,
            allow_network INTEGER NOT NULL DEFAULT 0,
            max_memory TEXT,
            max_cpu INTEGER,
            created_by TEXT,
            created_at TEXT NOT NULL,
            started_at TEXT,
            completed_at TEXT,
            FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
        );

        CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
        CREATE INDEX IF NOT EXISTS idx_tasks_group_name ON tasks(group_name);
        CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);
        CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
        "#,
    )
    .execute(&pool)
    .await
    .context("创建表失败")?;

    Ok(pool)
}

/// 创建 Axum 应用
async fn create_app(state: AppState) -> Result<Router> {
    // CORS 配置
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    // React Dashboard 静态文件服务
    let spa_service = ServeDir::new("dashboard/dist")
        .fallback(ServeFile::new("dashboard/dist/index.html"));

    // 创建路由
    let app = Router::new()
        // 登录 API (公开访问)
        .route("/api/v1/login", post(auth_middleware::handle_login))
        // 健康检查 (公开访问)
        .route("/health", get(routes::health::health_check))
        .route("/api/v1/health", get(routes::health::health_check))
        // API 路由 (需要认证)
        .route("/api/v1/tasks", post(routes::tasks::create_task).get(routes::tasks::list_tasks))
        .route(
            "/api/v1/tasks/:id",
            get(routes::tasks::get_task).delete(routes::tasks::delete_task),
        )
        .route("/api/v1/tasks/:id/execute", post(routes::tasks::execute_task))
        .route("/api/v1/tasks/:id/cancel", post(routes::tasks::cancel_task))
        .route(
            "/api/v1/memory/search",
            get(routes::memory::search_memory).post(routes::memory::search_memory),
        )
        .route("/api/v1/memory/:key", get(routes::memory::get_memory).delete(routes::memory::delete_memory))
        .route("/api/v1/memory/stats", get(routes::memory::memory_stats))
        .route("/ws/task/:id", get(routes::websocket::task_websocket))
        // Webhook 路由 (公开访问，有独立验证)
        .merge(webhook::create_routes())
        // 静态文件服务 (公开访问)
        .nest_service("/static", tower_http::services::ServeDir::new("static"))
        // SPA 路由: 所有其他路径都返回 index.html
        .fallback_service(spa_service)
        // 添加认证中间件到 API 路由
        .layer(axum::middleware::from_fn_with_state(
            state.clone(),
            auth_middleware::auth_middleware,
        ))
        // 添加状态
        .with_state(state)
        // 添加 CORS
        .layer(cors)
        // 添加日志中间件
        .layer(tower_http::trace::TraceLayer::new_for_http());

    Ok(app)
}

/// 等待关闭信号
async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("无法监听 Ctrl+C 信号");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("无法监听 TERM 信号")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {
            info!("收到 Ctrl+C 信号");
        },
        _ = terminate => {
            info!("收到 TERM 信号");
        },
    }

    info!("正在优雅关闭...");
}

/// 404 处理器
pub async fn handle_404() -> impl IntoResponse {
    (
        StatusCode::NOT_FOUND,
        Json(serde_json::json!({
            "success": false,
            "error": "接口不存在"
        })),
    )
}
