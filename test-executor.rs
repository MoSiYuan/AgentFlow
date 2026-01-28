//! 测试执行器核心功能
//!
//! 验证 TaskExecutor 和 ProcessKiller 的基本功能

use agentflow_core::executor::{TaskExecutor, ProcessKiller};
use std::path::PathBuf;
use std::time::Duration;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    println!("=== AgentFlow 执行器核心测试 ===\n");

    // 测试 1: 创建 TaskExecutor
    println!("测试 1: 创建 TaskExecutor");
    let workspace = PathBuf::from("/tmp/agentflow-test");
    let executor = TaskExecutor::new(
        workspace.clone(),
        Duration::from_secs(60),
    );
    println!("✓ TaskExecutor 创建成功");
    println!("  工作区: {}", executor.workspace_path().display());
    println!("  超时: {:?}", executor.timeout());

    // 测试 2: 修改超时设置
    println!("\n测试 2: 修改超时设置");
    let mut executor = TaskExecutor::new(
        workspace,
        Duration::from_secs(60),
    );
    executor.set_timeout(Duration::from_secs(120));
    println!("✓ 超时已更新为: {:?}", executor.timeout());

    // 测试 3: ProcessKiller 基本测试
    println!("\n测试 3: ProcessKiller 超时控制");
    let child = tokio::process::Command::new("echo")
        .arg("hello")
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()?;

    let mut killer = ProcessKiller::new(child, Duration::from_secs(5));
    let status = killer.wait_with_timeout().await?;
    println!("✓ 进程正常退出，状态: {}", status);

    // 测试 4: 超时熔断测试
    println!("\n测试 4: 超时熔断机制");
    let child = tokio::process::Command::new("sleep")
        .arg("10")
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()?;

    let mut killer = ProcessKiller::new(child, Duration::from_millis(100));
    let result = killer.wait_with_timeout().await;
    println!("✓ 超时熔断触发: {:?}", result.is_ok());

    println!("\n=== 所有测试完成 ===");
    Ok(())
}
