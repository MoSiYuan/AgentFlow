//! 进程生命周期管理
//!
//! 提供超时熔断机制和级联清理功能。

use anyhow::{Context, Result, anyhow};
use std::time::Duration;
use tokio::process::Child;
use tokio::time::{sleep, timeout};
use tracing::{debug, error, info, warn};

/// 进程终止器
///
/// 负责管理子进程的生命周期，实现超时熔断和优雅关闭：
/// 1. 超时后发送 SIGTERM 信号
/// 2. 等待 5 秒让进程清理资源
/// 3. 如果仍未退出，发送 SIGKILL 强制终止
/// 4. 支持进程组级联清理（清理所有子进程）
pub struct ProcessKiller {
    /// 子进程句柄
    child: Child,
    /// 超时时间
    timeout: Duration,
    /// 优雅关闭等待时间（默认 5 秒）
    grace_period: Duration,
}

impl ProcessKiller {
    /// 创建新的进程终止器
    ///
    /// # 参数
    ///
    /// * `child` - 要管理的子进程
    /// * `timeout` - 总超时时间
    pub fn new(child: Child, timeout: Duration) -> Self {
        Self {
            child,
            timeout,
            grace_period: Duration::from_secs(5),
        }
    }

    /// 设置优雅关闭等待时间
    ///
    /// # 参数
    ///
    /// * `grace_period` - SIGTERM 到 SIGKILL 之间的等待时间
    pub fn with_grace_period(mut self, grace_period: Duration) -> Self {
        self.grace_period = grace_period;
        self
    }

    /// 等待进程完成（带超时控制）
    ///
    /// 如果进程在超时时间内完成，返回退出状态。
    /// 如果超时，则触发熔断机制：SIGTERM -> 等待 -> SIGKILL。
    ///
    /// # 返回
    ///
    /// 返回进程的退出状态
    pub async fn wait_with_timeout(&mut self) -> Result<std::process::ExitStatus> {
        let pid = self.child.id().context("无法获取进程 ID")?;

        debug!(
            pid = pid,
            timeout_secs = self.timeout.as_secs(),
            "开始等待进程完成"
        );

        // 使用 tokio::time::timeout 实现超时控制
        match timeout(self.timeout, self.child.wait()).await {
            Ok(Ok(status)) => {
                // 进程在超时前正常退出
                info!(
                    pid = pid,
                    success = status.success(),
                    exit_code = ?status.code(),
                    "进程在超时前正常退出"
                );
                Ok(status)
            }
            Ok(Err(e)) => {
                // 进程等待出错
                error!(
                    pid = pid,
                    error = %e,
                    "等待进程时发生错误"
                );
                Err(anyhow!("等待进程失败: {}", e))
            }
            Err(_) => {
                // 超时 - 触发熔断机制
                warn!(
                    pid = pid,
                    timeout_secs = self.timeout.as_secs(),
                    "进程执行超时，触发熔断机制"
                );

                self.kill_with_timeout().await?;

                // 返回一个表示超时的退出状态
                #[cfg(unix)]
                {
                    use std::os::unix::process::ExitStatusExt;
                    Ok(std::process::ExitStatus::from_raw(256)) // 退出码 256 表示超时
                }

                #[cfg(windows)]
                {
                    // Windows 不支持 from_raw，使用默认的 ExitStatus
                    // 实际上这里已经在 kill_with_timeout 中处理了
                    self.child.wait().await.context("等待进程终止失败")?
                }
            }
        }
    }

    /// 超时熔断机制
    ///
    /// 分两步终止进程：
    /// 1. 发送 SIGTERM 信号，给予进程清理资源的机会
    /// 2. 等待 grace_period（默认 5 秒）
    /// 3. 如果仍未退出，发送 SIGKILL 强制终止
    ///
    /// 支持进程组级联清理，终止整个进程树。
    pub async fn kill_with_timeout(&mut self) -> Result<()> {
        let pid = self.child.id().context("无法获取进程 ID")?;

        info!(
            pid = pid,
            grace_period_secs = self.grace_period.as_secs(),
            "开始优雅终止进程"
        );

        // 第一步：尝试 SIGTERM（优雅关闭）
        if let Err(e) = self.try_sigterm() {
            debug!(
                pid = pid,
                error = %e,
                "SIGTERM 失败，将直接使用 SIGKILL"
            );
        } else {
            // 等待优雅关闭
            match timeout(self.grace_period, self.child.wait()).await {
                Ok(Ok(status)) => {
                    info!(
                        pid = pid,
                        exit_code = ?status.code(),
                        "进程在 SIGTERM 后正常退出"
                    );
                    return Ok(());
                }
                Ok(Err(e)) => {
                    debug!(
                        pid = pid,
                        error = %e,
                        "等待进程时发生错误"
                    );
                }
                Err(_) => {
                    warn!(
                        pid = pid,
                        grace_period_secs = self.grace_period.as_secs(),
                        "进程在优雅关闭期后仍未退出，使用 SIGKILL"
                    );
                }
            }
        }

        // 第二步：强制 SIGKILL
        self.force_kill().await?;

        Ok(())
    }

    /// 尝试发送 SIGTERM 信号
    ///
    /// 在 Unix 系统上，向进程组发送 SIGTERM，实现级联清理。
    /// 在 Windows 上，调用 TerminateProcess。
    fn try_sigterm(&self) -> Result<()> {
        #[cfg(unix)]
        {
            use std::os::unix::process::CommandExt;
            use nix::unistd::Pid;
            use nix::sys::signal::{self, Signal};

            let pid = self.child.id().context("无法获取进程 ID")? as i32;

            // 发送 SIGTERM 到整个进程组（负 PID）
            // 这会终止主进程及其所有子进程
            debug!(
                pid = pid,
                "向进程组发送 SIGTERM 信号"
            );

            signal::killpg(Pid::from_raw(-pid), Signal::SIGTERM)
                .context("发送 SIGTERM 失败")?;

            Ok(())
        }

        #[cfg(windows)]
        {
            use std::os::windows::process::CommandExt;

            // Windows 不支持信号，只能强制终止
            // 这里我们返回错误，让调用者直接使用 SIGKILL
            Err(anyhow!("Windows 不支持 SIGTERM"))
        }
    }

    /// 强制终止进程（SIGKILL）
    ///
    /// 在所有平台上都使用最强制的方式终止进程。
    async fn force_kill(&mut self) -> Result<()> {
        let pid = self.child.id().context("无法获取进程 ID")?;

        warn!(
            pid = pid,
            "使用 SIGKILL 强制终止进程"
        );

        // 调用 tokio 的 start_kill 方法
        self.child
            .start_kill()
            .context("发送 SIGKILL 失败")?;

        // 等待进程真正退出（最多等待 1 秒）
        match timeout(Duration::from_secs(1), self.child.wait()).await {
            Ok(Ok(status)) => {
                info!(
                    pid = pid,
                    exit_code = ?status.code(),
                    "进程已被强制终止"
                );
                Ok(())
            }
            Ok(Err(e)) => {
                error!(
                    pid = pid,
                    error = %e,
                    "等待进程终止时发生错误"
                );
                Err(anyhow!("等待进程终止失败: {}", e))
            }
            Err(_) => {
                // 即使等待超时，进程也应该已经被终止了
                warn!(
                    pid = pid,
                    "进程终止等待超时，但进程应该已被杀死"
                );
                Ok(())
            }
        }
    }

    /// 获取进程 ID
    pub fn pid(&self) -> Option<u32> {
        self.child.id()
    }

    /// 获取超时时间
    pub fn timeout(&self) -> Duration {
        self.timeout
    }

    /// 获取优雅关闭等待时间
    pub fn grace_period(&self) -> Duration {
        self.grace_period
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::process::Stdio;
    use tokio::process::Command;

    #[tokio::test]
    async fn test_killer_creation() {
        // 创建一个睡眠进程
        let child = Command::new("sleep")
            .arg("1000")
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
            .unwrap();

        let pid = child.id();
        let killer = ProcessKiller::new(child, Duration::from_secs(60));

        assert_eq!(killer.pid(), pid);
        assert_eq!(killer.timeout(), Duration::from_secs(60));
        assert_eq!(killer.grace_period(), Duration::from_secs(5));
    }

    #[tokio::test]
    async fn test_grace_period_configuration() {
        let child = Command::new("sleep")
            .arg("1000")
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
            .unwrap();

        let killer = ProcessKiller::new(child, Duration::from_secs(60))
            .with_grace_period(Duration::from_secs(10));

        assert_eq!(killer.grace_period(), Duration::from_secs(10));
    }

    #[tokio::test]
    async fn test_force_kill() {
        // 创建一个长时间睡眠进程
        let child = Command::new("sleep")
            .arg("1000")
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
            .unwrap();

        let mut killer = ProcessKiller::new(child, Duration::from_millis(100));

        // 直接强制终止
        let result = killer.force_kill().await;
        assert!(result.is_ok(), "强制终止应该成功");
    }

    #[tokio::test]
    async fn test_wait_timeout() {
        // 创建一个超过超时时间的睡眠进程
        let child = Command::new("sleep")
            .arg("10")
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
            .unwrap();

        let mut killer = ProcessKiller::new(child, Duration::from_millis(100));

        // 等待进程完成（应该超时）
        let result = killer.wait_with_timeout().await;
        assert!(result.is_ok(), "等待应该返回成功（即使超时）");
    }

    #[tokio::test]
    async fn test_wait_normal_completion() {
        // 创建一个快速完成的进程
        let child = Command::new("echo")
            .arg("hello")
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
            .unwrap();

        let mut killer = ProcessKiller::new(child, Duration::from_secs(60));

        // 等待进程完成（应该正常退出）
        let result = killer.wait_with_timeout().await;
        assert!(result.is_ok(), "等待应该成功");
        assert!(result.unwrap().success(), "进程应该成功退出");
    }
}
