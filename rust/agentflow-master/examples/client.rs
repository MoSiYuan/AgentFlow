//! AgentFlow Master 客户端使用示例
//!
//! 展示如何使用 HTTP API 和 WebSocket 与 Master 服务器交互

use reqwest::Client;
use serde_json::json;
use std::time::Duration;
use tokio_tungstenite::{connect_async, tungstenite::Message};
use futures::{SinkExt, StreamExt};

const SERVER_URL: &str = "http://localhost:6767";
const WS_URL: &str = "ws://localhost:6767/ws/task";

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("🚀 AgentFlow Master 客户端示例\n");

    // 创建 HTTP 客户端
    let client = Client::new();

    // 1. 健康检查
    println!("1️⃣  健康检查");
    let response = client
        .get(&format!("{}/health", SERVER_URL))
        .send()
        .await?;

    if response.status().is_success() {
        let body: serde_json::Value = response.json().await?;
        println!("   ✅ 服务器状态: {}", body["data"]["status"]);
        println!("   ✅ 版本: {}", body["data"]["version"]);
        println!("   ✅ 运行时间: {} 秒\n", body["data"]["uptime"]);
    }

    // 2. 创建任务
    println!("2️⃣  创建任务");
    let task_payload = json!({
        "title": "代码库分析",
        "description": "分析项目结构和依赖关系",
        "group_name": "code-analysis",
        "priority": "high",
        "sandboxed": true,
        "allow_network": false,
        "created_by": "demo_user"
    });

    let response = client
        .post(&format!("{}/api/v1/tasks", SERVER_URL))
        .json(&task_payload)
        .send()
        .await?;

    if response.status().is_success() {
        let body: serde_json::Value = response.json().await?;
        let task_id = body["data"]["id"].as_i64().unwrap();
        let task_uuid = body["data"]["task_id"].as_str().unwrap();
        println!("   ✅ 任务创建成功!");
        println!("   📋 任务 ID: {}", task_id);
        println!("   🆔 任务 UUID: {}", task_uuid);
        println!("   📊 状态: {}\n", body["data"]["status"]);

        // 3. 获取任务详情
        println!("3️⃣  获取任务详情");
        tokio::time::sleep(Duration::from_secs(1)).await;

        let response = client
            .get(&format!("{}/api/v1/tasks/{}", SERVER_URL, task_id))
            .send()
            .await?;

        if response.status().is_success() {
            let body: serde_json::Value = response.json().await?;
            println!("   ✅ 任务详情:");
            println!("   📋 标题: {}", body["data"]["title"]);
            println!("   📝 描述: {:?}", body["data"]["description"]);
            println!("   📊 状态: {}\n", body["data"]["status"]);
        }

        // 4. 执行任务
        println!("4️⃣  执行任务");
        let response = client
            .post(&format!("{}/api/v1/tasks/{}/execute", SERVER_URL, task_id))
            .send()
            .await?;

        if response.status().is_success() {
            println!("   ✅ 任务执行请求已发送");
            println!("   ⏳ 等待执行完成...\n");

            // 等待任务完成
            tokio::time::sleep(Duration::from_secs(3)).await;

            // 检查任务状态
            let response = client
                .get(&format!("{}/api/v1/tasks/{}", SERVER_URL, task_id))
                .send()
                .await?;

            if response.status().is_success() {
                let body: serde_json::Value = response.json().await?;
                let status = body["data"]["status"].as_str().unwrap();
                println!("   ✅ 任务执行完成!");
                println!("   📊 最终状态: {}", status);

                if let Some(result) = body["data"]["result"].as_str() {
                    println!("   📄 执行结果:");
                    for line in result.lines() {
                        println!("      {}", line);
                    }
                }
                println!();
            }
        }

        // 5. 使用 WebSocket 监听任务
        println!("5️⃣  WebSocket 连接示例");
        let ws_url = format!("{}/{}", WS_URL, task_id);
        println!("   🔌 连接到: {}", ws_url);

        match connect_async(&ws_url).await {
            Ok((ws_stream, _)) => {
                println!("   ✅ WebSocket 连接成功!\n");

                let (mut write, mut read) = ws_stream.split();

                // 发送 ping
                let ping_msg = json!({"type": "ping"});
                write.send(Message::Text(ping_msg.to_string())).await?;
                println!("   📤 发送 ping 消息");

                // 接收消息（超时 5 秒）
                let timeout = tokio::time::sleep(Duration::from_secs(5));
                tokio::pin!(timeout);

                loop {
                    tokio::select! {
                        msg = read.next() => {
                            match msg {
                                Some(Ok(Message::Text(text))) => {
                                    if let Ok(data) = serde_json::from_str::<serde_json::Value>(&text) {
                                        println!("   📥 收到消息:");
                                        println!("      类型: {}", data["type"]);
                                        if let Some(msg) = data.get("message") {
                                            println!("      消息: {}", msg);
                                        }
                                    }
                                }
                                Some(Ok(Message::Close(_))) => {
                                    println!("   🔒 WebSocket 连接关闭");
                                    break;
                                }
                                Some(Err(e)) => {
                                    println!("   ❌ WebSocket 错误: {}", e);
                                    break;
                                }
                                _ => {}
                            }
                        }
                        _ = &mut timeout => {
                            println!("   ⏱️  超时，断开连接\n");
                            break;
                        }
                    }
                }
            }
            Err(e) => {
                println!("   ❌ WebSocket 连接失败: {}\n", e);
            }
        }

        // 6. 列出所有任务
        println!("6️⃣  列出所有任务");
        let response = client
            .get(&format!("{}/api/v1/tasks", SERVER_URL))
            .send()
            .await?;

        if response.status().is_success() {
            let body: serde_json::Value = response.json().await?;
            if let Some(tasks) = body["data"].as_array() {
                println!("   ✅ 任务列表 (共 {} 个):", tasks.len());
                for task in tasks.iter().take(5) {
                    println!("   📋 [{}] {} - {}",
                        task["id"],
                        task["title"],
                        task["status"]
                    );
                }
            }
            println!();
        }

        // 7. 记忆搜索
        println!("7️⃣  记忆搜索");
        let response = client
            .get(&format!("{}/api/v1/memory/search?q=分析", SERVER_URL))
            .send()
            .await?;

        if response.status().is_success() {
            let body: serde_json::Value = response.json().await?;
            if let Some(results) = body["data"].as_array() {
                println!("   ✅ 找到 {} 条记忆:", results.len());
                for entry in results.iter().take(3) {
                    println!("   🧠 [{}] {}", entry["category"], entry["key"]);
                }
            } else {
                println!("   ℹ️  没有找到相关记忆");
            }
            println!();
        }

        // 8. 记忆统计
        println!("8️⃣  记忆统计");
        let response = client
            .get(&format!("{}/api/v1/memory/stats", SERVER_URL))
            .send()
            .await?;

        if response.status().is_success() {
            let body: serde_json::Value = response.json().await?;
            println!("   ✅ 记忆统计:");
            println!("   📊 总数: {}", body["data"]["total"]);
            println!("   ✅ 活跃: {}", body["data"]["active"]);
            println!("   ⏰ 过期: {}", body["data"]["expired"]);
            println!();
        }

        // 9. 删除任务
        println!("9️⃣  删除任务");
        let response = client
            .delete(&format!("{}/api/v1/tasks/{}", SERVER_URL, task_id))
            .send()
            .await?;

        if response.status().is_success() {
            let body: serde_json::Value = response.json().await?;
            println!("   ✅ 任务删除成功: {}\n", body["data"]["deleted"]);
        }
    }

    println!("🎉 示例运行完成!");

    Ok(())
}
