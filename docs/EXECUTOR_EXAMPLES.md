# 执行引擎核心 - 代码示例集合

## 目录

1. [基础示例](#基础示例)
2. [高级用法](#高级用法)
3. [错误处理](#错误处理)
4. [测试示例](#测试示例)
5. [集成示例](#集成示例)

---

## 基础示例

### 示例 1: 创建并使用 TaskExecutor

```rust
use agentflow_core::executor::TaskExecutor;
use std::path::PathBuf;
use std::time::Duration;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // 创建执行器
    let executor = TaskExecutor::new(
        PathBuf::from("/workspace/my-project"),
        Duration::from_secs(300), // 5分钟超时
    );

    // 执行任务
    let result = executor.execute("请帮我写一个冒泡排序函数").await?;

    // 处理结果
    if result.success {
        println!("✓ 任务成功");
        println!("输出:\n{}", result.stdout);
    } else {
        eprintln!("✗ 任务失败");
        eprintln!("错误:\n{}", result.stderr);
        eprintln!("退出码: {:?}", result.exit_code);
    }

    Ok(())
}
```

### 示例 2: 动态调整超时时间

```rust
use agentflow_core::executor::TaskExecutor;
use std::path::PathBuf;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let workspace = PathBuf::from("/workspace");
    let mut executor = TaskExecutor::new(
        workspace,
        Duration::from_secs(60), // 初始超时 60 秒
    );

    // 根据任务复杂度调整超时
    let task_prompt = "复杂的重构任务...";
    if task_prompt.len() > 100 {
        executor.set_timeout(Duration::from_secs(600)); // 10 分钟
    }

    let result = executor.execute(task_prompt).await?;
    Ok(())
}
```

### 示例 3: 检查工作区路径

```rust
use agentflow_core::executor::TaskExecutor;
use std::path::PathBuf;

fn main() {
    let workspace = PathBuf::from("/workspace/repo");
    let executor = TaskExecutor::new(
        workspace.clone(),
        Duration::from_secs(300),
    );

    // 获取并验证工作区路径
    let workspace_path = executor.workspace_path();
    println!("工作区路径: {}", workspace_path.display());
    println!("是否存在: {}", workspace_path.exists());
    println!("是否绝对路径: {}", workspace_path.is_absolute());
}
```

---

## 高级用法

### 示例 4: 直接使用 ProcessKiller

```rust
use agentflow_core::executor::ProcessKiller;
use tokio::process::Command;
use std::time::Duration;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // 启动进程
    let child = Command::new("claude")
        .arg("长时间运行的任务...")
        .spawn()?;

    // 创建进程管理器，10 分钟超时
    let mut killer = ProcessKiller::new(child, Duration::from_secs(600));

    // 等待进程完成（自动处理超时）
    let status = killer.wait_with_timeout().await?;

    println!("进程退出状态: {}", status);

    Ok(())
}
```

### 示例 5: 自定义优雅关闭时间

```rust
use agentflow_core::executor::ProcessKiller;
use tokio::process::Command;
use std::time::Duration;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let child = Command::new("claude")
        .arg("需要清理资源的任务...")
        .spawn()?;

    // 设置 15 秒的优雅关闭时间（默认 5 秒）
    let mut killer = ProcessKiller::new(child, Duration::from_secs(300))
        .with_grace_period(Duration::from_secs(15));

    let status = killer.wait_with_timeout().await?;
    Ok(())
}
```

### 示例 6: 获取进程信息

```rust
use agentflow_core::executor::ProcessKiller;
use tokio::process::Command;
use std::time::Duration;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let child = Command::new("sleep")
        .arg("100")
        .spawn()?;

    let killer = ProcessKiller::new(child, Duration::from_secs(60));

    // 获取进程 ID
    if let Some(pid) = killer.pid() {
        println!("进程 PID: {}", pid);
    }

    // 获取超时配置
    println!("超时时间: {:?}", killer.timeout());
    println!("优雅关闭时间: {:?}", killer.grace_period());

    Ok(())
}
```

---

## 错误处理

### 示例 7: 处理执行错误

```rust
use agentflow_core::executor::TaskExecutor;
use std::path::PathBuf;

#[tokio::main]
async fn main() {
    let executor = TaskExecutor::new(
        PathBuf::from("/workspace"),
        Duration::from_secs(60),
    );

    match executor.execute("任务提示词").await {
        Ok(result) => {
            if result.success {
                println!("成功: {}", result.stdout);
            } else {
                eprintln!("失败: {}", result.stderr);
            }
        }
        Err(e) => {
            // 处理各种错误
            if let Some(io_err) = e.downcast_ref::<std::io::Error>() {
                eprintln!("IO 错误: {}", io_err);
            } else {
                eprintln!("其他错误: {}", e);
            }
        }
    }
}
```

### 示例 8: 处理超时错误

```rust
use agentflow_core::executor::ProcessKiller;
use tokio::process::Command;
use std::time::Duration;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let child = Command::new("sleep")
        .arg("1000")
        .spawn()?;

    let mut killer = ProcessKiller::new(child, Duration::from_millis(100));

    match killer.wait_with_timeout().await {
        Ok(status) => {
            if let Some(code) = status.code() {
                if code == 256 {
                    println!("任务超时被终止");
                } else {
                    println!("任务退出，退出码: {}", code);
                }
            }
        }
        Err(e) => {
            eprintln!("等待进程时出错: {}", e);
        }
    }

    Ok(())
}
```

### 示例 9: 错误上下文

```rust
use agentflow_core::executor::TaskExecutor;
use anyhow::{Context, Result};

#[tokio::main]
async fn main() -> Result<()> {
    let workspace = std::path::PathBuf::from("/workspace");

    let executor = TaskExecutor::new(
        workspace.clone(),
        std::time::Duration::from_secs(60),
    );

    let result = executor
        .execute("任务提示词")
        .context(format!(
            "在工作区 {} 执行任务失败",
            workspace.display()
        ))?;

    Ok(())
}
```

---

## 测试示例

### 示例 10: 单元测试 TaskExecutor

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use agentflow_core::executor::TaskExecutor;
    use tempfile::TempDir;

    #[tokio::test]
    async fn test_executor_with_temp_workspace() {
        // 创建临时工作区
        let temp_dir = TempDir::new().unwrap();
        let executor = TaskExecutor::new(
            temp_dir.path().to_path_buf(),
            Duration::from_secs(60),
        );

        // 验证工作区
        assert!(executor.workspace_path().exists());
        assert_eq!(executor.timeout(), Duration::from_secs(60));
    }
}
```

### 示例 11: 测试进程终止

```rust
#[tokio::test]
async fn test_process_termination() {
    use agentflow_core::executor::ProcessKiller;

    // 创建长时间运行的进程
    let child = Command::new("sleep")
        .arg("1000")
        .stdout(std::process::Stdio::null())
        .spawn()
        .unwrap();

    // 设置短超时
    let mut killer = ProcessKiller::new(child, Duration::from_millis(100));

    // 等待超时并终止
    let result = killer.wait_with_timeout().await;
    assert!(result.is_ok());
}
```

### 示例 12: 集成测试

```rust
#[tokio::test]
async fn integration_test() {
    use agentflow_core::executor::TaskExecutor;
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let executor = TaskExecutor::new(
        temp_dir.path().to_path_buf(),
        Duration::from_secs(10),
    );

    // 执行简单任务
    let result = executor.execute("echo hello").await;

    // 验证结果
    assert!(result.is_ok());
    let exec_result = result.unwrap();
    assert!(exec_result.success);
}
```

---

## 集成示例

### 示例 13: 与数据库集成

```rust
use agentflow_core::executor::TaskExecutor;
use agentflow_core::database::create_pool;
use std::path::PathBuf;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // 创建数据库连接池
    let pool = create_pool(Path::new("/data/agentflow.db")).await?;

    // 创建执行器
    let executor = TaskExecutor::new(
        PathBuf::from("/workspace/repo"),
        Duration::from_secs(300),
    );

    // 执行任务
    let result = executor.execute("任务提示词").await?;

    // 保存结果到数据库
    // TODO: 实现数据库保存逻辑

    Ok(())
}
```

### 示例 14: 与记忆系统集成

```rust
use agentflow_core::executor::TaskExecutor;
use agentflow_core::types::MemoryEntry;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let executor = TaskExecutor::new(
        PathBuf::from("/workspace"),
        Duration::from_secs(300),
    );

    // 准备相关记忆
    let memories = vec![
        MemoryEntry {
            key: "context".to_string(),
            value: serde_json::json!("项目上下文信息"),
            expires_at: None,
            category: MemoryCategory::Context,
            task_id: None,
            timestamp: chrono::Utc::now().timestamp(),
        },
    ];

    // 构建包含记忆的提示词
    let prompt = format!(
        "背景: {}\n\n任务: {}",
        memories[0].value, "具体任务"
    );

    let result = executor.execute(&prompt).await?;
    Ok(())
}
```

### 示例 15: 批量任务执行

```rust
use agentflow_core::executor::TaskExecutor;
use futures::stream::{self, StreamExt};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let executor = TaskExecutor::new(
        PathBuf::from("/workspace"),
        Duration::from_secs(300),
    );

    let tasks = vec![
        "任务 1",
        "任务 2",
        "任务 3",
    ];

    // 并行执行任务
    let results = stream::iter(tasks)
        .map(|task| {
            let executor = &executor;
            async move {
                executor.execute(task).await
            }
        })
        .buffer_unordered(3) // 并发 3 个任务
        .collect::<Vec<_>>()
        .await;

    // 处理结果
    for (i, result) in results.into_iter().enumerate() {
        match result {
            Ok(r) => println!("任务 {} 成功: {}", i + 1, r.success),
            Err(e) => eprintln!("任务 {} 失败: {}", i + 1, e),
        }
    }

    Ok(())
}
```

---

## 最佳实践

### ✅ DO（推荐做法）

1. **始终设置合理的超时时间**
   ```rust
   let executor = TaskExecutor::new(
       workspace,
       Duration::from_secs(300), // 5 分钟
   );
   ```

2. **使用临时目录进行测试**
   ```rust
   let temp_dir = TempDir::new()?;
   let executor = TaskExecutor::new(temp_dir.path().to_path_buf(), timeout);
   ```

3. **处理所有可能的错误**
   ```rust
   match executor.execute(prompt).await {
       Ok(result) => { /* 处理结果 */ }
       Err(e) => { /* 处理错误 */ }
   }
   ```

4. **使用 context() 添加错误上下文**
   ```rust
   .context("执行任务时出错")?
   ```

### ❌ DON'T（不推荐做法）

1. **不要使用硬编码路径**
   ```rust
   // ❌ 不好
   let executor = TaskExecutor::new(
       PathBuf::from("/home/user/project"),
       timeout,
   );

   // ✅ 好
   let workspace = std::env::var("WORKSPACE_DIR")?;
   let executor = TaskExecutor::new(
       PathBuf::from(workspace),
       timeout,
   );
   ```

2. **不要忽略错误**
   ```rust
   // ❌ 不好
   let _ = executor.execute(prompt).await;

   // ✅ 好
   let result = executor.execute(prompt).await?;
   ```

3. **不要设置过长的超时**
   ```rust
   // ❌ 不好
   Duration::from_secs(86400) // 24 小时

   // ✅ 好
   Duration::from_secs(600) // 10 分钟
   ```

---

## 性能优化建议

### 1. 复用执行器实例

```rust
// ❌ 不好：每次创建新实例
for task in tasks {
    let executor = TaskExecutor::new(workspace, timeout);
    executor.execute(task).await?;
}

// ✅ 好：复用实例
let executor = TaskExecutor::new(workspace, timeout);
for task in tasks {
    executor.execute(task).await?;
}
```

### 2. 使用适当的超时时间

```rust
// 根据任务复杂度设置不同的超时
let timeout = match task.complexity {
    Complexity::Low => Duration::from_secs(60),
    Complexity::Medium => Duration::from_secs(300),
    Complexity::High => Duration::from_secs(600),
};
```

### 3. 并行执行独立任务

```rust
use futures::future::join_all;

let futures = tasks.iter()
    .map(|task| executor.execute(task))
    .collect::<Vec<_>>();

let results = join_all(futures).await;
```

---

## 常见问题

### Q1: 如何判断任务是超时还是正常结束？

```rust
match result.exit_code {
    Some(256) => println!("任务超时"),
    Some(0) => println!("任务成功"),
    Some(code) => println!("任务退出，退出码: {}", code),
    None => println!("任务被信号终止"),
}
```

### Q2: 如何清理僵尸进程？

使用 ProcessKiller 的进程组清理功能：

```rust
let mut killer = ProcessKiller::new(child, timeout);
// 自动清理所有子进程
killer.kill_with_timeout().await?;
```

### Q3: 如何限制并发任务数？

```rust
use futures::stream::{StreamExt, self};

let results = stream::iter(tasks)
    .map(|task| executor.execute(task))
    .buffer_unordered(3) // 最多 3 个并发
    .collect::<Vec<_>>()
    .await;
```

---

**更多示例和用法，请参考源代码中的文档注释。**
