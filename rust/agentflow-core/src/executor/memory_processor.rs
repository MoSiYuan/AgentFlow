//! 记忆整理处理器
//!
//! 在任务执行完成后，自动调用 Claude CLI 分析结果并整理记忆

use crate::memory::MemoryCore;
use crate::types::{MemoryCategory, MemoryEntry};
use anyhow::{Context, Result};
use serde_json::json;
use std::path::PathBuf;
use std::process::Stdio;
use std::sync::Arc;
use tokio::process::Command;
use tracing::{debug, info, warn};

/// 记忆整理处理器
///
/// 负责在任务执行完成后，调用 Claude CLI 分析结果并提取关键信息存储到记忆系统
#[derive(Clone)]
pub struct MemoryProcessor {
    /// 记忆核心
    memory_core: Arc<MemoryCore>,
    /// 工作目录
    workspace_path: PathBuf,
    /// 分析超时时间（秒）
    timeout: u64,
}

impl MemoryProcessor {
    /// 创建新的记忆处理器
    ///
    /// # 参数
    /// * `memory_core` - 记忆核心实例
    /// * `workspace_path` - 工作目录路径
    /// * `timeout` - 分析超时时间（秒）
    pub fn new(memory_core: MemoryCore, workspace_path: PathBuf, timeout: u64) -> Self {
        Self {
            memory_core: Arc::new(memory_core),
            workspace_path,
            timeout,
        }
    }

    /// 处理任务执行结果
    ///
    /// 调用 Claude CLI 分析任务执行结果，提取关键信息并存储到记忆系统
    ///
    /// # 参数
    /// * `task` - 任务描述
    /// * `stdout` - 标准输出
    /// * `stderr` - 标准错误输出
    /// * `exit_code` - 退出码
    ///
    /// # 返回
    /// 成功返回 Ok(())，失败返回错误信息
    pub async fn process_task_result(
        &self,
        task: &str,
        stdout: &str,
        stderr: &str,
        exit_code: Option<i32>,
    ) -> Result<()> {
        info!("开始整理任务记忆: task={}", task);

        // 1. 构建记忆整理 Prompt
        let prompt = self.build_memory_prompt(task, stdout, stderr, exit_code)?;

        debug!("记忆整理 Prompt 长度: {} 字符", prompt.len());

        // 2. 调用 Claude CLI 分析
        let analysis = tokio::time::timeout(
            std::time::Duration::from_secs(self.timeout),
            self.analyze_with_claude(&prompt),
        )
        .await
        .context("记忆整理超时")?
        .context("调用 Claude CLI 分析失败")?;

        debug!("Claude 分析结果长度: {} 字符", analysis.len());

        // 3. 解析分析结果
        let memory_entries = self.parse_analysis(task, &analysis)?;

        debug!("解析出 {} 条记忆条目", memory_entries.len());

        // 4. 存储到 MemoryCore
        for entry in &memory_entries {
            self.memory_core.index(entry).await.with_context(|| {
                format!("存储记忆条目失败: key={}", entry.key)
            })?;
        }

        info!(
            "任务记忆整理完成: 生成了 {} 条记忆",
            memory_entries.len()
        );

        Ok(())
    }

    /// 构建记忆整理 Prompt
    ///
    /// # 参数
    /// * `task` - 任务描述
    /// * `stdout` - 标准输出
    /// * `stderr` - 标准错误输出
    /// * `exit_code` - 退出码
    fn build_memory_prompt(
        &self,
        task: &str,
        stdout: &str,
        stderr: &str,
        exit_code: Option<i32>,
    ) -> Result<String> {
        let success = exit_code.map_or(false, |code| code == 0);

        // 截断过长的输出（避免超过 Claude 的 token 限制）
        let max_output_len = 10000;
        let stdout_truncated = if stdout.len() > max_output_len {
            format!("{}...[已截断]", &stdout[..max_output_len])
        } else {
            stdout.to_string()
        };

        let stderr_truncated = if stderr.len() > max_output_len {
            format!("{}...[已截断]", &stderr[..max_output_len])
        } else {
            stderr.to_string()
        };

        Ok(format!(
            r#"你是一个记忆整理专家。请分析以下任务执行结果，提取关键信息。

## 任务描述
{}

## 执行结果
- 退出码: {:?}
- 成功: {}

## 标准输出
{}

## 错误输出
{}

请按以下 JSON 格式返回分析结果（必须返回有效的 JSON）：
{{
  "summary": "任务执行的简要总结（1-2句话）",
  "key_learnings": ["学到的关键点1", "关键点2"],
  "errors": ["遇到的错误1（如有）"],
  "solutions": ["解决方案1（如有）"],
  "tags": ["标签1", "标签2"],
  "category": "Execution|Result|Error"
}}

注意：
1. 如果任务失败，重点关注错误和解决方案
2. 如果任务成功，重点关注学习经验和成果
3. tags 应该包含相关的技术关键词
4. 只返回 JSON，不要包含其他内容"#,
            task, exit_code, success, stdout_truncated, stderr_truncated
        ))
    }

    /// 调用 Claude CLI 分析
    ///
    /// # 参数
    /// * `prompt` - 分析 Prompt
    ///
    /// # 返回
    /// 返回 Claude 的分析结果
    async fn analyze_with_claude(&self, prompt: &str) -> Result<String> {
        debug!("调用 Claude CLI 进行记忆分析");

        let output = Command::new("claude")
            .arg(prompt)
            .env_clear()
            .envs(std::env::vars()) // 继承环境变量（包括 ANTHROPIC_API_KEY）
            .current_dir(&self.workspace_path)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .output()
            .await
            .context("启动 Claude CLI 失败")?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            warn!("Claude CLI 执行失败: {}", stderr);
            return Err(anyhow::anyhow!("Claude CLI 执行失败: {}", stderr));
        }

        let result = String::from_utf8_lossy(&output.stdout).to_string();
        debug!("Claude CLI 分析完成，输出长度: {} 字符", result.len());

        Ok(result)
    }

    /// 解析 Claude 返回的分析结果
    ///
    /// # 参数
    /// * `task` - 任务描述
    /// * `analysis` - Claude 返回的分析结果
    ///
    /// # 返回
    /// 返回解析出的记忆条目列表
    fn parse_analysis(&self, task: &str, analysis: &str) -> Result<Vec<MemoryEntry>> {
        debug!("解析 Claude 分析结果");

        // 尝试提取 JSON（Claude 可能会在 JSON 前后添加额外文本）
        let json_str = self.extract_json(analysis)?;

        // 解析 JSON
        let parsed: serde_json::Value = serde_json::from_str(json_str)
            .with_context(|| format!("解析 JSON 失败: {}", json_str))?;

        let mut entries = Vec::new();
        let timestamp = chrono::Utc::now().timestamp();

        // 1. 存储执行记录（总是存在）
        if let Some(summary) = parsed.get("summary").and_then(|v| v.as_str()) {
            let tags = parsed.get("tags").and_then(|v| v.as_array());
            let tags_vec = tags.map(|arr| {
                arr.iter()
                    .filter_map(|v| v.as_str().map(|s| s.to_string()))
                    .collect::<Vec<_>>()
            });

            entries.push(MemoryEntry {
                key: format!("execution:{}:{}", task, uuid::Uuid::new_v4()),
                value: json!({
                    "summary": summary,
                    "task": task,
                    "tags": tags_vec.unwrap_or_default()
                }),
                category: MemoryCategory::Execution,
                timestamp,
                expires_at: None,
                task_id: None,
            });
        }

        // 2. 存储错误信息（如果有）
        if let Some(errors) = parsed.get("errors").and_then(|v| v.as_array()) {
            if !errors.is_empty() {
                let solutions = parsed.get("solutions").and_then(|v| v.as_array());

                entries.push(MemoryEntry {
                    key: format!("error:{}:{}", task, uuid::Uuid::new_v4()),
                    value: json!({
                        "errors": errors,
                        "solutions": solutions,
                        "task": task
                    }),
                    category: MemoryCategory::Error,
                    timestamp,
                    expires_at: None,
                    task_id: None,
                });
            }
        }

        // 3. 存储成功经验（如果有）
        if let Some(learnings) = parsed.get("key_learnings").and_then(|v| v.as_array()) {
            if !learnings.is_empty() {
                entries.push(MemoryEntry {
                    key: format!("learning:{}:{}", task, uuid::Uuid::new_v4()),
                    value: json!({
                        "learnings": learnings,
                        "task": task
                    }),
                    category: MemoryCategory::Result,
                    timestamp,
                    expires_at: None,
                    task_id: None,
                });
            }
        }

        Ok(entries)
    }

    /// 从文本中提取 JSON
    ///
    /// Claude 可能会在 JSON 前后添加解释性文本，需要提取纯 JSON
    ///
    /// # 参数
    /// * `text` - 包含 JSON 的文本
    ///
    /// # 返回
    /// 返回提取出的 JSON 字符串
    fn extract_json<'a>(&self, text: &'a str) -> Result<&'a str> {
        // 查找第一个 { 和最后一个 }
        let start = text
            .find('{')
            .context("未找到 JSON 开始标记")?;
        let end = text
            .rfind('}')
            .context("未找到 JSON 结束标记")?;

        if start >= end {
            return Err(anyhow::anyhow!("JSON 标记顺序错误"));
        }

        Ok(&text[start..=end])
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_json() {
        let processor = MemoryProcessor::new(
            // memory_core 参数会在实际测试时创建
            unimplemented!(),
            PathBuf::from("/tmp"),
            60,
        );

        // 测试纯 JSON
        let json = r#"{"key": "value"}"#;
        assert_eq!(processor.extract_json(json).unwrap(), json);

        // 测试带前后文本的 JSON
        let text = r#"一些解释文本
        {
          "key": "value"
        }
        更多文本"#;
        assert_eq!(
            processor.extract_json(text).unwrap(),
            "{\n          \"key\": \"value\"\n        }"
        );
    }
}
