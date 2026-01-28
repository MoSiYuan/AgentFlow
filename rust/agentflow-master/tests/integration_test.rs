//! AgentFlow Master 集成测试
//!
//! 测试 Master 服务器的核心功能

use reqwest::{Client, Response};
use serde_json::json;

const SERVER_URL: &str = "http://localhost:6767";

#[tokio::test]
async fn test_health_check() {
    let client = Client::new();
    let response = client
        .get(&format!("{}/health", SERVER_URL))
        .send()
        .await
        .expect("健康检查失败");

    assert!(response.status().is_success());

    let body: serde_json::Value = response.json().await.expect("解析响应失败");
    assert_eq!(body["success"], true);
    assert_eq!(body["data"]["status"], "healthy");
    assert_eq!(body["data"]["mode"], "master");
}

#[tokio::test]
async fn test_create_task() {
    let client = Client::new();

    let task_payload = json!({
        "title": "集成测试任务",
        "description": "这是一个集成测试任务",
        "group_name": "test",
        "priority": "high",
        "sandboxed": true,
        "allow_network": false
    });

    let response = client
        .post(&format!("{}/api/v1/tasks", SERVER_URL))
        .json(&task_payload)
        .send()
        .await
        .expect("创建任务失败");

    assert!(response.status().is_success());

    let body: serde_json::Value = response.json().await.expect("解析响应失败");
    assert_eq!(body["success"], true);
    assert_eq!(body["data"]["title"], "集成测试任务");
    assert_eq!(body["data"]["status"], "pending");
}

#[tokio::test]
async fn test_get_task() {
    let client = Client::new();

    // 先创建一个任务
    let task_payload = json!({
        "title": "获取测试任务",
        "description": "测试获取任务详情"
    });

    let create_response = client
        .post(&format!("{}/api/v1/tasks", SERVER_URL))
        .json(&task_payload)
        .send()
        .await
        .expect("创建任务失败");

    let task: serde_json::Value = create_response.json().await.expect("解析失败");
    let task_id = task["data"]["id"].as_i64().unwrap();

    // 获取任务详情
    let get_response = client
        .get(&format!("{}/api/v1/tasks/{}", SERVER_URL, task_id))
        .send()
        .await
        .expect("获取任务失败");

    assert!(get_response.status().is_success());

    let body: serde_json::Value = get_response.json().await.expect("解析失败");
    assert_eq!(body["success"], true);
    assert_eq!(body["data"]["id"], task_id);
}

#[tokio::test]
async fn test_list_tasks() {
    let client = Client::new();

    let response = client
        .get(&format!("{}/api/v1/tasks", SERVER_URL))
        .send()
        .await
        .expect("列出任务失败");

    assert!(response.status().is_success());

    let body: serde_json::Value = response.json().await.expect("解析失败");
    assert_eq!(body["success"], true);
    assert!(body["data"].is_array());
}

#[tokio::test]
async fn test_execute_task() {
    let client = Client::new();

    // 创建任务
    let task_payload = json!({
        "title": "执行测试任务",
        "description": "测试任务执行"
    });

    let create_response = client
        .post(&format!("{}/api/v1/tasks", SERVER_URL))
        .json(&task_payload)
        .send()
        .await
        .expect("创建任务失败");

    let task: serde_json::Value = create_response.json().await.expect("解析失败");
    let task_id = task["data"]["id"].as_i64().unwrap();

    // 执行任务（SSE）
    let execute_response = client
        .post(&format!("{}/api/v1/tasks/{}/execute", SERVER_URL, task_id))
        .send()
        .await
        .expect("执行任务失败");

    assert!(execute_response.status().is_success());

    // 验证任务状态
    tokio::time::sleep(tokio::time::Duration::from_secs(3)).await;

    let get_response = client
        .get(&format!("{}/api/v1/tasks/{}", SERVER_URL, task_id))
        .send()
        .await
        .expect("获取任务失败");

    let body: serde_json::Value = get_response.json().await.expect("解析失败");
    assert_eq!(body["data"]["status"], "completed");
}

#[tokio::test]
async fn test_memory_search() {
    let client = Client::new();

    let response = client
        .get(&format!("{}/api/v1/memory/search?q=test", SERVER_URL))
        .send()
        .await
        .expect("搜索记忆失败");

    assert!(response.status().is_success());

    let body: serde_json::Value = response.json().await.expect("解析失败");
    assert_eq!(body["success"], true);
    assert!(body["data"].is_array());
}

#[tokio::test]
async fn test_memory_stats() {
    let client = Client::new();

    let response = client
        .get(&format!("{}/api/v1/memory/stats", SERVER_URL))
        .send()
        .await
        .expect("获取统计失败");

    assert!(response.status().is_success());

    let body: serde_json::Value = response.json().await.expect("解析失败");
    assert_eq!(body["success"], true);
    assert!(body["data"]["total"].is_number());
    assert!(body["data"]["active"].is_number());
}
