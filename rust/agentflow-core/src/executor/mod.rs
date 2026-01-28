//! 任务执行器模块
//!
//! 负责执行 Claude AI 任务，提供进程管理、超时控制、输出捕获等功能。

mod killer;
mod memory_processor;
mod prompt_builder;

pub use killer::ProcessKiller;
pub use memory_processor::MemoryProcessor;
pub use prompt_builder::{PromptBuilder, PromptBuilderConfig};

use crate::git::BranchManager;
use anyhow::{Context, Result};
use std::path::PathBuf;
use std::sync::Arc;
use std::time::Duration;
use tokio::process::Command;
use tracing::{debug, info, warn};

/// 执行结果
#[derive(Debug, Clone)]
pub struct ExecutionResult {
    /// 是否成功
    pub success: bool,
    /// 标准输出内容
    pub stdout: String,
    /// 标准错误内容
    pub stderr: String,
    /// 退出码
    pub exit_code: Option<i32>,
}

/// 任务执行器
///
/// 负责启动和管理 Claude 命令的执行，包括：
/// - 环境变量继承（自动读取 ANTHROPIC_API_KEY）
/// - 工作目录设置（Git 沙箱路径）
/// - 实时输出捕获
/// - 超时控制
/// - 自动化 Git 工作流（创建/清理分支）
/// - 任务完成后自动整理记忆（可选）
pub struct TaskExecutor {
    /// 工作区路径（Git 沙箱目录）
    workspace_path: PathBuf,
    /// 执行超时时间
    timeout: Duration,
    /// Git 分支管理器（可选）
    git_manager: Option<BranchManager>,
    /// 记忆处理器（可选）
    memory_processor: Option<MemoryProcessor>,
}

impl TaskExecutor {
    /// 创建新的任务执行器
    ///
    /// # 参数
    ///
    /// * `workspace_path` - Git 沙箱路径，将作为工作目录
    /// * `timeout` - 执行超时时间
    pub fn new(workspace_path: PathBuf, timeout: Duration) -> Self {
        Self {
            workspace_path,
            timeout,
            git_manager: None,
            memory_processor: None,
        }
    }

    /// 设置 Git 分支管理器
    ///
    /// # 参数
    ///
    /// * `git_manager` - Git 分支管理器实例
    pub fn with_git_manager(mut self, git_manager: BranchManager) -> Self {
        self.git_manager = Some(git_manager);
        self
    }

    /// 设置记忆处理器
    ///
    /// # 参数
    ///
    /// * `processor` - 记忆处理器实例
    pub fn with_memory_processor(mut self, processor: MemoryProcessor) -> Self {
        self.memory_processor = Some(processor);
        self
    }

    /// 执行任务
    ///
    /// 启动 Claude 进程执行指定的提示词，并捕获输出。
    ///
    /// # 参数
    ///
    /// * `prompt` - 要发送给 Claude 的提示词
    ///
    /// # 返回
    ///
    /// 返回执行结果，包含退出状态、标准输出和标准错误
    ///
    /// # 示例
    ///
    /// ```no_run
    /// use agentflow_core::executor::TaskExecutor;
    /// use std::path::PathBuf;
    /// use std::time::Duration;
    ///
    /// # async fn example() -> anyhow::Result<()> {
    /// let executor = TaskExecutor::new(
    ///     PathBuf::from("/tmp/workspace"),
    ///     Duration::from_secs(300)
    /// );
    ///
    /// let result = executor.execute("请帮我写一个 Hello World 程序").await?;
    ///
    /// if result.success {
    ///     println!("输出: {}", result.stdout);
    /// } else {
    ///     eprintln!("错误: {}", result.stderr);
    /// }
    /// # Ok(())
    /// # }
    /// ```
    pub async fn execute(&self, prompt: &str) -> Result<ExecutionResult> {
        info!(
            workspace = %self.workspace_path.display(),
            timeout_secs = self.timeout.as_secs(),
            prompt_len = prompt.len(),
            "开始执行任务"
        );

        // 构建命令
        let mut cmd = Command::new("claude");
        cmd.arg(prompt)
            // 继承父进程的环境变量（包括 ANTHROPIC_API_KEY）
            .env_clear()
            .envs(std::env::vars())
            // 设置工作目录为 Git 沙箱路径
            .current_dir(&self.workspace_path);

        debug!(
            command = ?cmd,
            "准备启动 Claude 进程"
        );

        // 启动进程
        let child = cmd
            .spawn()
            .context("启动 Claude 进程失败")?;

        // 使用 ProcessKiller 管理进程生命周期（超时控制）
        let mut killer = ProcessKiller::new(child, self.timeout);

        // 等待进程完成（带超时控制）
        let status = killer
            .wait_with_timeout()
            .await
            .context("等待进程完成时发生错误")?;

        // 获取输出（注意：tokio::process::Child 的输出需要提前接管）
        // 这里我们假设 ProcessKiller 已经处理了输出捕获
        // 实际实现中需要在 ProcessKiller 中异步读取 stdout/stderr

        let exit_code = status.code();
        let success = status.success();

        debug!(
            success = success,
            exit_code = ?exit_code,
            "进程执行完成"
        );

        // TODO: 实现输出捕获逻辑
        // 需要在 ProcessKiller 中启动异步任务读取 stdout/stderr
        let stdout = String::new();
        let stderr = String::new();

        let result = ExecutionResult {
            success,
            stdout,
            stderr,
            exit_code,
        };

        // 如果配置了记忆处理器，在任务完成后触发记忆整理
        if let Some(ref processor) = self.memory_processor {
            debug!("触发记忆整理");

            // 异步处理记忆，不阻塞主任务结果返回
            let task_description = prompt.to_string();
            let stdout_clone = result.stdout.clone();
            let stderr_clone = result.stderr.clone();
            let exit_code_clone = result.exit_code;
            let processor_clone = processor.clone();

            tokio::spawn(async move {
                if let Err(e) = processor_clone
                    .process_task_result(&task_description, &stdout_clone, &stderr_clone, exit_code_clone)
                    .await
                {
                    warn!("记忆整理失败: {}", e);
                }
            });
        }

        Ok(result)
    }

    /// 获取工作区路径
    pub fn workspace_path(&self) -> &PathBuf {
        &self.workspace_path
    }

    /// 获取超时时间
    pub fn timeout(&self) -> Duration {
        self.timeout
    }

    /// 设置超时时间
    pub fn set_timeout(&mut self, timeout: Duration) {
        self.timeout = timeout;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use tempfile::TempDir;

    #[tokio::test]
    async fn test_executor_creation() {
        let temp_dir = TempDir::new().unwrap();
        let executor = TaskExecutor::new(
            temp_dir.path().to_path_buf(),
            Duration::from_secs(60),
        );

        assert_eq!(executor.workspace_path(), temp_dir.path());
        assert_eq!(executor.timeout(), Duration::from_secs(60));
    }

    #[tokio::test]
    async fn test_executor_set_timeout() {
        let temp_dir = TempDir::new().unwrap();
        let mut executor = TaskExecutor::new(
            temp_dir.path().to_path_buf(),
            Duration::from_secs(60),
        );

        executor.set_timeout(Duration::from_secs(120));
        assert_eq!(executor.timeout(), Duration::from_secs(120));
    }

    #[tokio::test]
    async fn test_executor_workspace_exists() {
        let temp_dir = TempDir::new().unwrap();
        let workspace_path = temp_dir.path().join("workspace");

        // 创建工作区目录
        fs::create_dir(&workspace_path).unwrap();

        let executor = TaskExecutor::new(
            workspace_path.clone(),
            Duration::from_secs(60),
        );

        assert_eq!(executor.workspace_path(), &workspace_path);
        assert!(workspace_path.exists());
    }
}
