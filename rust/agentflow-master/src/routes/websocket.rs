//! WebSocket 路由处理
//!
//! 提供实时任务执行进度和状态更新的 WebSocket 连接

use super::AppState;
use axum::{
    extract::{
        Path,
        State,
        ws::{Message, WebSocket, WebSocketUpgrade},
    },
    response::IntoResponse,
};
use futures::{sink::SinkExt, stream::StreamExt};
use serde_json::json;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{debug, error, info};

/// 任务 WebSocket 连接
///
/// GET /ws/task/:id
///
/// 提供任务执行的实时更新
pub async fn task_websocket(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    ws: WebSocketUpgrade,
) -> impl IntoResponse {
    info!("WebSocket 连接请求: 任务 {}", id);

    ws.on_upgrade(move |socket| handle_task_websocket(socket, state, id))
}

/// 处理任务 WebSocket 连接
async fn handle_task_websocket(socket: WebSocket, state: AppState, task_id: i64) {
    let (mut sender, mut receiver) = socket.split();

    // 发送欢迎消息
    let welcome_msg = json!({
        "type": "connected",
        "task_id": task_id,
        "message": "WebSocket 连接已建立"
    });

    if let Err(e) = sender
        .send(Message::Text(welcome_msg.to_string()))
        .await
    {
        error!("发送欢迎消息失败: {}", e);
        return;
    }

    // 克隆 sender 以便在后台任务中使用
    let sender = Arc::new(RwLock::new(sender));

    // 启动任务执行监控
    let monitor_sender = sender.clone();
    let monitor_state = state.clone();
    tokio::spawn(async move {
        if let Err(e) = monitor_task_execution(monitor_state, task_id, monitor_sender).await {
            error!("任务监控失败: {}", e);
        }
    });

    // 处理客户端消息
    while let Some(msg) = receiver.next().await {
        match msg {
            Ok(Message::Text(text)) => {
                debug!("收到客户端消息: {}", text);

                // 解析客户端消息
                if let Ok(client_msg) = serde_json::from_str::<serde_json::Value>(&text) {
                    if let Some(msg_type) = client_msg.get("type").and_then(|v| v.as_str()) {
                        match msg_type {
                            "ping" => {
                                // 响应 ping 消息
                                let pong_msg = json!({
                                    "type": "pong",
                                    "task_id": task_id
                                });

                                let sender = sender.read().await;
                                if let Err(e) = sender.send(Message::Text(pong_msg.to_string())).await {
                                    error!("发送 pong 消息失败: {}", e);
                                    break;
                                }
                            }
                            "subscribe" => {
                                // 订阅任务更新（已经默认订阅）
                                debug!("客户端订阅任务 {} 的更新", task_id);
                            }
                            _ => {
                                debug!("未知消息类型: {}", msg_type);
                            }
                        }
                    }
                }
            }
            Ok(Message::Close(_)) => {
                info!("客户端关闭 WebSocket 连接");
                break;
            }
            Err(e) => {
                error!("WebSocket 错误: {}", e);
                break;
            }
            _ => {}
        }
    }

    info!("WebSocket 连接已关闭: 任务 {}", task_id);
}

/// 监控任务执行并发送实时更新
async fn monitor_task_execution(
    state: AppState,
    task_id: i64,
    sender: Arc<RwLock<futures::stream::SplitSink<WebSocket, Message>>>,
) -> Result<(), anyhow::Error> {
    // 检查任务是否存在
    let task = sqlx::query!(
        r#"
        SELECT id, task_id, status, title, started_at, completed_at
        FROM tasks
        WHERE id = ?
        "#,
        task_id
    )
    .fetch_optional(&state.executor.db)
    .await?;

    if task.is_none() {
        let error_msg = json!({
            "type": "error",
            "task_id": task_id,
            "error": "任务不存在"
        });

        let sender = sender.write().await;
        sender.send(Message::Text(error_msg.to_string())).await?;
        return Err(anyhow::anyhow!("任务不存在"));
    }

    let task = task.unwrap();

    // 发送任务状态
    let status_msg = json!({
        "type": "status",
        "task_id": task_id,
        "status": format!("{:?}", task.status),
        "title": task.title
    });

    {
        let sender = sender.write().await;
        sender.send(Message::Text(status_msg.to_string())).await?;
    }

    // 如果任务正在运行，开始监控
    if task.status.to_string() == "Running" {
        // 模拟任务进度更新
        for progress in 1..=5 {
            tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;

            let progress_msg = json!({
                "type": "progress",
                "task_id": task_id,
                "progress": progress * 20,
                "message": format!("执行进度: {}%", progress * 20)
            });

            let sender = sender.write().await;
            sender.send(Message::Text(progress_msg.to_string())).await?;
        }

        // 检查任务最终状态
        let final_task = sqlx::query!(
            r#"
            SELECT id, status, result, error, completed_at
            FROM tasks
            WHERE id = ?
            "#,
            task_id
        )
        .fetch_optional(&state.executor.db)
        .await?;

        if let Some(final_task) = final_task {
            let completion_msg = if final_task.error.is_some() {
                json!({
                    "type": "failed",
                    "task_id": task_id,
                    "error": final_task.error
                })
            } else {
                json!({
                    "type": "completed",
                    "task_id": task_id,
                    "result": final_task.result
                })
            };

            let sender = sender.write().await;
            sender.send(Message::Text(completion_msg.to_string())).await?;
        }
    } else if task.status.to_string() == "Completed" {
        // 任务已完成
        let completed_msg = json!({
            "type": "completed",
            "task_id": task_id,
            "result": task.result
        });

        let sender = sender.write().await;
        sender.send(Message::Text(completed_msg.to_string())).await?;
    } else if task.status.to_string() == "Failed" {
        // 任务失败
        let failed_msg = json!({
            "type": "failed",
            "task_id": task_id,
            "error": task.error.unwrap_or_else(|| "未知错误".to_string())
        });

        let sender = sender.write().await;
        sender.send(Message::Text(failed_msg.to_string())).await?;
    }

    Ok(())
}
