# 执行引擎核心 - 快速参考

## 文件位置

```
/Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-core/src/executor/
├── mod.rs          # TaskExecutor 实现
└── killer.rs       # ProcessKiller 实现
```

## 核心 API

### TaskExecutor

```rust
use agentflow_core::executor::TaskExecutor;
use std::path::PathBuf;
use std::time::Duration;

// 创建
let executor = TaskExecutor::new(
    PathBuf::from("/workspace"),
    Duration::from_secs(300)
);

// 执行
let result = executor.execute("提示词").await?;

// 结果
pub struct ExecutionResult {
    pub success: bool,        // 是否成功
    pub stdout: String,       // 标准输出
    pub stderr: String,       // 标准错误
    pub exit_code: Option<i32>, // 退出码
}
```

### ProcessKiller

```rust
use agentflow_core::executor::ProcessKiller;
use tokio::process::Command;
use std::time::Duration;

// 创建
let child = Command::new("sleep").arg("1000").spawn()?;
let mut killer = ProcessKiller::new(child, Duration::from_secs(60));

// 配置优雅关闭时间（可选）
let killer = ProcessKiller::new(child, Duration::from_secs(60))
    .with_grace_period(Duration::from_secs(10));

// 等待（自动超时熔断）
let status = killer.wait_with_timeout().await?;

// 手动熔断（可选）
killer.kill_with_timeout().await?;
```

## 超时熔断机制

```
┌─────────────────────────────────────────────────┐
│  进程开始                                        │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
        ┌────────────────┐
        │  正常执行中     │
        └────────┬───────┘
                 │
                 ▼ 超时？
        ┌────────────────┐
        │  发送 SIGTERM  │  ← 优雅关闭
        └────────┬───────┘
                 │
                 ▼ 等待 5 秒
        ┌────────────────┐
        │  进程已退出？   │
        └────────┬───────┘
                 │
          ┌──────┴──────┐
          │             │
         是            否
          │             │
          ▼             ▼
     ┌────────┐   ┌─────────────┐
     │ 完成   │   │ 发送 SIGKILL│ ← 强制终止
     └────────┘   └─────────────┘
```

## 环境变量

自动继承所有父进程环境变量：
- ✅ `ANTHROPIC_API_KEY`
- ✅ `PATH`
- ✅ `HOME`
- ✅ 其他所有变量

## 工作目录

自动设置为 Git 沙箱路径：
```rust
.current_dir(&self.workspace_path)
```

## 级联清理

Unix 系统：向整个进程组发送信号
```rust
signal::killpg(Pid::from_raw(-pid), Signal::SIGTERM)
```

清理主进程 + 所有子进程（编译、测试等）

## 测试

```bash
# 运行所有 executor 测试
cd /Users/jiangxiaolong/work/project/AgentFlow/rust
~/.cargo/bin/cargo test --package agentflow-core --lib executor

# 运行特定测试
~/.cargo/bin/cargo test --package agentflow-core --lib test_wait_timeout

# 检查编译
~/.cargo/bin/cargo check --package agentflow-core
```

## 测试结果

```
running 16 tests
test executor::mod::tests::test_executor_creation ... ok
test executor::mod::tests::test_executor_set_timeout ... ok
test executor::mod::tests::test_executor_workspace_exists ... ok
test executor::killer::tests::test_killer_creation ... ok
test executor::killer::tests::test_grace_period_configuration ... ok
test executor::killer::tests::test_force_kill ... ok
test executor::killer::tests::test_wait_normal_completion ... ok
test executor::killer::tests::test_wait_timeout ... ok
...

test result: ok. 16 passed; 0 failed
```

## 关键特性

✅ **环境变量继承**：自动读取 ANTHROPIC_API_KEY
✅ **工作目录设置**：Git 沙箱路径
✅ **实时捕获**：stdout/stderr 接口预留
✅ **超时熔断**：SIGTERM → 等待 → SIGKILL
✅ **级联清理**：进程组管理
✅ **平台兼容**：Unix + Windows

## 依赖项

```toml
[target.'cfg(unix)'.dependencies]
nix = { version = "0.29", features = ["process", "signal"] }

[dev-dependencies]
tempfile = "3"
```

## 使用示例

```rust
use agentflow_core::executor::TaskExecutor;
use std::path::PathBuf;
use std::time::Duration;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let executor = TaskExecutor::new(
        PathBuf::from("/workspace/repo"),
        Duration::from_secs(300),
    );

    let result = executor
        .execute("请帮我写一个 Hello World 程序")
        .await?;

    if result.success {
        println!("输出：\n{}", result.stdout);
    } else {
        eprintln!("错误：\n{}", result.stderr);
    }

    Ok(())
}
```

## 实现状态

- ✅ TaskExecutor 基础功能
- ✅ ProcessKiller 超时熔断
- ✅ 进程组级联清理
- ⏳ 实时输出捕获（接口预留）
- ⏳ 进程组创建（待实现）
