//! Prompt 构建器模块
//!
//! 提供完整的 Prompt 构建功能，包括系统指令、记忆检索和任务整合

use crate::memory::MemoryCore;
use crate::types::{MemoryCategory, MemoryEntry};
use anyhow::{Context, Result};
use serde_json::json;
use std::path::PathBuf;
use tracing::{debug, warn};

/// Prompt 构建器配置
#[derive(Debug, Clone)]
pub struct PromptBuilderConfig {
    /// 系统指令
    pub system_instruction: String,

    /// 最大 token 数量（默认 200K）
    pub max_tokens: usize,

    /// 是否包含记忆上下文
    pub include_memory: bool,

    /// 记忆条目最大数量
    pub max_memory_entries: usize,

    /// 是否包含任务元数据
    pub include_metadata: bool,
}

impl Default for PromptBuilderConfig {
    fn default() -> Self {
        Self {
            system_instruction: Self::default_system_instruction(),
            max_tokens: 200_000,
            include_memory: true,
            max_memory_entries: 10,
            include_metadata: true,
        }
    }
}

impl PromptBuilderConfig {
    /// 创建默认系统指令
    fn default_system_instruction() -> String {
        r#"## Role
你是 AgentFlow Execution Engine，而非 Chat Assistant。

## Core Directives
1. **Action First**: 不说 "I suggest"，直接 "I will..."
2. **Tool Usage**: 优先使用 Bash 执行命令，而非生成代码片段
3. **Git Mandatory**: 任何实质性代码修改后，必须执行 `git add && git commit`
4. **Self-Healing**: 命令报错时，尝试分析并修复，不要直接停止
5. **No Questions**: 不要问"你想运行哪一个？"，直接执行最合理的操作

## Protocol
- 在行动前先思考步骤（但不输出思考过程）
- 执行后简要报告结果
- 失败时主动尝试替代方案

## Important
- 用户给你的任务是明确的指令，不是讨论的话题
- 默认情况下，直接执行；只有遇到真正危险的操作时才询问"#.to_string()
    }

    /// 创建自定义配置
    ///
    /// # 参数
    /// * `system_instruction` - 自定义系统指令
    /// * `max_tokens` - 最大 token 数量
    pub fn new(system_instruction: String, max_tokens: usize) -> Self {
        Self {
            system_instruction,
            max_tokens,
            ..Default::default()
        }
    }
}

/// Prompt 构建器
///
/// 负责构建完整的 Prompt，包括系统指令、记忆上下文和任务描述
pub struct PromptBuilder {
    /// 构建器配置
    config: PromptBuilderConfig,
    /// 工作目录路径
    workspace_path: PathBuf,
}

impl PromptBuilder {
    /// 创建新的 Prompt 构建器
    ///
    /// # 示例
    /// ```
    /// use agentflow_core::executor::PromptBuilder;
    ///
    /// let builder = PromptBuilder::new();
    /// ```
    pub fn new() -> Self {
        Self {
            config: PromptBuilderConfig::default(),
            workspace_path: std::env::current_dir().unwrap_or_else(|_| PathBuf::from(".")),
        }
    }

    /// 使用指定工作目录创建 Prompt 构建器
    ///
    /// # 参数
    /// * `path` - 工作目录路径
    ///
    /// # 示例
    /// ```
    /// use agentflow_core::executor::PromptBuilder;
    /// use std::path::PathBuf;
    ///
    /// let builder = PromptBuilder::with_workspace(PathBuf::from("/path/to/project"));
    /// ```
    pub fn with_workspace(path: PathBuf) -> Self {
        Self {
            config: PromptBuilderConfig::default(),
            workspace_path: path,
        }
    }

    /// 使用自定义配置创建 Prompt 构建器
    ///
    /// # 参数
    /// * `config` - 构建器配置
    ///
    /// # 示例
    /// ```
    /// use agentflow_core::executor::{PromptBuilder, PromptBuilderConfig};
    ///
    /// let config = PromptBuilderConfig::default();
    /// let builder = PromptBuilder::with_config(config);
    /// ```
    pub fn with_config(config: PromptBuilderConfig) -> Self {
        Self {
            config,
            workspace_path: std::env::current_dir().unwrap_or_else(|_| PathBuf::from(".")),
        }
    }

    /// 读取项目配置文件
    ///
    /// 如果 {workspace}/AGENTFLOW.md 存在，读取其内容
    ///
    /// # 返回
    /// * `Some(String)` - 配置文件内容
    /// * `None` - 配置文件不存在或读取失败
    fn load_project_config(&self) -> Option<String> {
        let config_path = self.workspace_path.join("AGENTFLOW.md");

        if config_path.exists() {
            std::fs::read_to_string(&config_path).ok()
        } else {
            None
        }
    }

    /// 构建完整的 Prompt
    ///
    /// 将系统指令、记忆上下文和任务描述组合成完整的 Prompt
    ///
    /// # 参数
    /// * `task` - 任务描述
    /// * `memories` - 相关记忆条目（可选）
    ///
    /// # 返回
    /// 返回构建好的完整 Prompt 字符串
    ///
    /// # 示例
    /// ```no_run
    /// # use agentflow_core::executor::PromptBuilder;
    /// # use agentflow_core::types::{MemoryEntry, MemoryCategory};
    /// # use serde_json::json;
    /// #
    /// # fn example() {
    /// let builder = PromptBuilder::new();
    ///
    /// let task = "实现一个用户登录功能";
    /// let memories = vec![
    ///     MemoryEntry {
    ///         key: "previous_login_impl".to_string(),
    ///         value: json!("之前使用 JWT 实现的登录"),
    ///         expires_at: None,
    ///         category: MemoryCategory::Context,
    ///         task_id: None,
    ///         timestamp: 0,
    ///     }
    /// ];
    ///
    /// let prompt = builder.build(task, &memories);
    /// println!("{}", prompt);
    /// # }
    /// ```
    pub fn build(&self, task: &str, memories: &[MemoryEntry]) -> String {
        debug!("构建 Prompt，任务长度: {}, 记忆条目数: {}", task.len(), memories.len());

        let mut parts = Vec::new();

        // 1. 添加系统指令
        parts.push(self.build_system_section());

        // 2. 添加项目配置（如果存在）
        if let Some(project_config) = self.load_project_config() {
            parts.push(self.build_project_config_section(&project_config));
        }

        // 3. 添加记忆上下文（如果有）
        if self.config.include_memory && !memories.is_empty() {
            let memory_section = self.build_memory_section(memories);
            parts.push(memory_section);
        }

        // 4. 添加任务描述
        parts.push(self.build_task_section(task));

        // 5. 组合所有部分
        let full_prompt = parts.join("\n\n");

        debug!("Prompt 构建完成，总长度: {}", full_prompt.len());

        // 6. 检查并截断（如果超过 token 限制）
        self.truncate_if_needed(&full_prompt)
    }

    /// 构建带自动记忆检索的 Prompt
    ///
    /// 自动调用 MemoryCore::search() 查找相关记忆，无需手动提供记忆条目
    ///
    /// # 参数
    /// * `task` - 任务描述
    /// * `memory_core` - 记忆核心引用
    ///
    /// # 返回
    /// 返回包含相关记忆的完整 Prompt
    ///
    /// # 示例
    /// ```no_run
    /// # use agentflow_core::executor::PromptBuilder;
    /// # use agentflow_core::memory::MemoryCore;
    /// #
    /// # async fn example() -> anyhow::Result<()> {
    /// let builder = PromptBuilder::new();
    /// let memory_core = MemoryCore::new("/path/to/memory.db").await?;
    ///
    /// let task = "实现一个用户登录功能";
    /// let prompt = builder.build_with_memory_search(task, &memory_core).await;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn build_with_memory_search(
        &self,
        task: &str,
        memory_core: &crate::memory::MemoryCore,
    ) -> String {
        debug!("执行自动记忆检索，任务: {}", task);

        // 1. 搜索相关记忆（top 3）
        let memories = match memory_core.search(task, 3).await {
            Ok(m) => {
                debug!("成功检索到 {} 条相关记忆", m.len());
                m
            },
            Err(e) => {
                warn!("记忆检索失败: {}, 使用空记忆", e);
                Vec::new()
            }
        };

        // 2. 调用现有的 build 方法
        self.build(task, &memories)
    }

    /// 构建带元数据的 Prompt
    ///
    /// # 参数
    /// * `task` - 任务描述
    /// * `memories` - 相关记忆条目
    /// * `metadata` - 任务元数据
    ///
    /// # 返回
    /// 返回包含元数据的完整 Prompt
    pub fn build_with_metadata(
        &self,
        task: &str,
        memories: &[MemoryEntry],
        metadata: &serde_json::Value,
    ) -> String {
        debug!("构建带元数据的 Prompt");

        let mut parts = Vec::new();

        // 1. 添加系统指令
        parts.push(self.build_system_section());

        // 2. 添加项目配置（如果存在）
        if let Some(project_config) = self.load_project_config() {
            parts.push(self.build_project_config_section(&project_config));
        }

        // 3. 添加元数据（如果启用）
        if self.config.include_metadata {
            let metadata_section = self.build_metadata_section(metadata);
            parts.push(metadata_section);
        }

        // 4. 添加记忆上下文
        if self.config.include_memory && !memories.is_empty() {
            let memory_section = self.build_memory_section(memories);
            parts.push(memory_section);
        }

        // 5. 添加任务描述
        parts.push(self.build_task_section(task));

        // 6. 组合并处理长度限制
        let full_prompt = parts.join("\n\n");
        self.truncate_if_needed(&full_prompt)
    }

    /// 构建带自动记忆检索的 Prompt
    ///
    /// 自动调用 MemoryCore::search() 查找相关记忆
    ///
    /// # 参数
    /// * `task` - 任务描述
    /// * `memory_core` - 记忆核心引用
    ///
    /// # 返回
    /// 返回包含相关记忆的完整 Prompt
    ///
    /// # 示例
    /// ```no_run
    /// # use agentflow_core::executor::PromptBuilder;
    /// # use agentflow_core::memory::MemoryCore;
    /// #
    /// # async fn example() {
    /// # let builder = PromptBuilder::new();
    /// # let memory_core = MemoryCore::new("memory.db").await.unwrap();
    /// let task = "修复登录 bug";
    /// let prompt = builder.build_with_memory_search(&task, &memory_core).await;
    /// # }
    /// ```
    pub async fn build_with_memory_search(
        &self,
        task: &str,
        memory_core: &MemoryCore,
    ) -> String {
        debug!("构建带自动记忆检索的 Prompt: task={}", task);

        // 搜索相关记忆（top 3）
        let memories = match memory_core.search(task, 3).await {
            Ok(m) => {
                debug!("检索到 {} 条相关记忆", m.len());
                m
            }
            Err(e) => {
                warn!("记忆检索失败: {}, 使用空记忆", e);
                Vec::new()
            }
        };

        // 调用现有的 build 方法
        self.build(task, &memories)
    }

    /// 构建系统指令部分
    fn build_system_section(&self) -> String {
        format!(
            "{}\n{}",
            "## 系统指令".to_string(),
            self.config.system_instruction
        )
    }

    /// 构建项目配置部分
    ///
    /// # 参数
    /// * `project_config` - 项目配置内容
    ///
    /// # 返回
    /// 格式化的项目配置字符串
    fn build_project_config_section(&self, project_config: &str) -> String {
        format!(
            "## 项目专属配置\n\n{}",
            project_config
        )
    }

    /// 构建记忆上下文部分
    ///
    /// # 参数
    /// * `memories` - 记忆条目列表
    ///
    /// # 返回
    /// 格式化的记忆上下文字符串
    fn build_memory_section(&self, memories: &[MemoryEntry]) -> String {
        // 限制记忆条目数量
        let limited_memories = memories
            .iter()
            .take(self.config.max_memory_entries)
            .collect::<Vec<_>>();

        if limited_memories.is_empty() {
            return String::new();
        }

        let mut memory_text = String::from("## 相关上下文\n\n");

        for (index, memory) in limited_memories.iter().enumerate() {
            let category_label = match memory.category {
                MemoryCategory::Execution => "执行记录",
                MemoryCategory::Context => "上下文",
                MemoryCategory::Result => "结果",
                MemoryCategory::Error => "错误",
                MemoryCategory::Checkpoint => "检查点",
            };

            memory_text.push_str(&format!(
                "### {} [{}]\n",
                category_label, index + 1
            ));

            if let Some(task_id) = &memory.task_id {
                memory_text.push_str(&format!("- **任务ID**: {}\n", task_id));
            }

            // 格式化记忆值
            let value_str = self.format_memory_value(&memory.value);
            memory_text.push_str(&format!("- **内容**: {}\n\n", value_str));
        }

        memory_text
    }

    /// 构建任务描述部分
    ///
    /// # 参数
    /// * `task` - 任务描述
    ///
    /// # 返回
    /// 格式化的任务描述字符串
    fn build_task_section(&self, task: &str) -> String {
        format!(
            "## 当前任务\n\n{}",
            task
        )
    }

    /// 构建元数据部分
    ///
    /// # 参数
    /// * `metadata` - 元数据 JSON 值
    ///
    /// # 返回
    /// 格式化的元数据字符串
    fn build_metadata_section(&self, metadata: &serde_json::Value) -> String {
        if metadata.is_null() {
            return String::new();
        }

        let metadata_str = serde_json::to_string_pretty(metadata)
            .unwrap_or_else(|_| metadata.to_string());

        format!(
            "## 任务元数据\n\n```json\n{}\n```",
            metadata_str
        )
    }

    /// 格式化记忆值
    ///
    /// # 参数
    /// * `value` - JSON 值
    ///
    /// # 返回
    /// 格式化后的字符串
    fn format_memory_value(&self, value: &serde_json::Value) -> String {
        match value {
            serde_json::Value::String(s) => s.clone(),
            serde_json::Value::Array(arr) => {
                if arr.len() == 1 {
                    self.format_memory_value(&arr[0])
                } else {
                    serde_json::to_string_pretty(value)
                        .unwrap_or_else(|_| value.to_string())
                }
            }
            serde_json::Value::Object(obj) => {
                serde_json::to_string_pretty(value)
                    .unwrap_or_else(|_| value.to_string())
            }
            _ => value.to_string(),
        }
    }

    /// 根据需要截断 Prompt
    ///
    /// 如果 Prompt 超过最大 token 限制，则截断记忆部分
    ///
    /// # 参数
    /// * `prompt` - 完整的 Prompt 字符串
    ///
    /// # 返回
    /// 截断后的 Prompt（如果需要）
    fn truncate_if_needed(&self, prompt: &str) -> String {
        // 估算 token 数量（简化估算：1 token ≈ 4 字符）
        let estimated_tokens = prompt.len() / 4;

        if estimated_tokens <= self.config.max_tokens {
            return prompt.to_string();
        }

        warn!(
            "Prompt 超过 token 限制 (估计: {} > 最大: {})，需要截断",
            estimated_tokens, self.config.max_tokens
        );

        // 截断策略：保留系统指令和任务，减少记忆条目
        let parts: Vec<&str> = prompt.split("\n\n## ").collect();

        if parts.len() < 3 {
            // 如果结构简单，直接截断
            let max_chars = self.config.max_tokens * 4;
            return prompt.chars().take(max_chars).collect();
        }

        // 保留系统指令和任务，移除记忆部分
        let system_part = parts.get(0).unwrap_or(&"");
        let task_part = parts.last().unwrap_or(&"");

        // 尝试保留部分记忆
        let truncated = format!(
            "{}\n\n## {}",
            system_part,
            task_part
        );

        // 如果仍然太长，进一步截断
        let max_chars = self.config.max_tokens * 4;
        if truncated.len() > max_chars {
            truncated.chars().take(max_chars).collect()
        } else {
            truncated
        }
    }

    /// 估算 Prompt 的 token 数量
    ///
    /// # 参数
    /// * `prompt` - Prompt 字符串
    ///
    /// # 返回
    /// 估算的 token 数量
    pub fn estimate_tokens(&self, prompt: &str) -> usize {
        // 简化估算：1 token ≈ 4 字符（英文），中文约 2 字符/token
        let char_count = prompt.chars().count();
        char_count / 3
    }

    /// 验证 Prompt 是否在限制内
    ///
    /// # 参数
    /// * `prompt` - Prompt 字符串
    ///
    /// # 返回
    /// - Ok(()) 如果在限制内
    /// - Err 如果超过限制
    pub fn validate_length(&self, prompt: &str) -> Result<()> {
        let estimated = self.estimate_tokens(prompt);

        if estimated > self.config.max_tokens {
            anyhow::bail!(
                "Prompt 超过 token 限制: {} > {}",
                estimated,
                self.config.max_tokens
            );
        }

        Ok(())
    }

    /// 获取配置引用
    pub fn config(&self) -> &PromptBuilderConfig {
        &self.config
    }

    /// 更新配置
    pub fn set_config(&mut self, config: PromptBuilderConfig) -> &mut Self {
        self.config = config;
        self
    }
}

impl Default for PromptBuilder {
    fn default() -> Self {
        Self::new()
    }
}

/// 快速构建 Prompt 的便捷函数
///
/// # 参数
/// * `task` - 任务描述
/// * `memories` - 记忆条目（可选）
///
/// # 返回
/// 构建好的 Prompt
///
/// # 示例
/// ```no_run
/// # use agentflow_core::executor::build_prompt;
/// # use agentflow_core::types::MemoryEntry;
/// #
/// # fn example() {
/// let prompt = build_prompt("实现登录功能", &[]);
/// println!("{}", prompt);
/// # }
/// ```
pub fn build_prompt(task: &str, memories: &[MemoryEntry]) -> String {
    let builder = PromptBuilder::new();
    builder.build(task, memories)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::MemoryCategory;
    use serde_json::json;

    #[test]
    fn test_prompt_builder_basic() {
        let builder = PromptBuilder::new();
        let task = "实现一个简单的计算器";
        let memories = vec![];

        let prompt = builder.build(task, &memories);

        assert!(prompt.contains("系统指令"));
        assert!(prompt.contains("当前任务"));
        assert!(prompt.contains("计算器"));
    }

    #[test]
    fn test_prompt_builder_with_memories() {
        let builder = PromptBuilder::new();
        let task = "修复登录 bug";
        let memories = vec![
            MemoryEntry {
                key: "login_error".to_string(),
                value: json!("用户登录时遇到 500 错误"),
                expires_at: None,
                category: MemoryCategory::Error,
                task_id: Some("task_123".to_string()),
                timestamp: 0,
            },
        ];

        let prompt = builder.build(task, &memories);

        assert!(prompt.contains("相关上下文"));
        assert!(prompt.contains("错误"));
        assert!(prompt.contains("500 错误"));
    }

    #[test]
    fn test_prompt_builder_with_metadata() {
        let builder = PromptBuilder::new();
        let task = "部署应用";
        let memories = vec![];
        let metadata = json!({
            "environment": "production",
            "version": "1.0.0"
        });

        let prompt = builder.build_with_metadata(task, &memories, &metadata);

        assert!(prompt.contains("任务元数据"));
        assert!(prompt.contains("production"));
        assert!(prompt.contains("1.0.0"));
    }

    #[test]
    fn test_prompt_truncation() {
        let config = PromptBuilderConfig {
            system_instruction: "System instruction".to_string(),
            max_tokens: 100, // 很小的限制
            include_memory: true,
            max_memory_entries: 10,
            include_metadata: true,
        };

        let builder = PromptBuilder::with_config(config);

        // 创建一个很长的任务
        let long_task = "A".repeat(1000);
        let memories = vec![];

        let prompt = builder.build(&long_task, &memories);

        // 验证 prompt 被截断
        assert!(prompt.len() < 1000); // 应该被截断
    }

    #[test]
    fn test_token_estimation() {
        let builder = PromptBuilder::new();
        let short_prompt = "Hello, world!";
        let long_prompt = "A".repeat(1000);

        let short_tokens = builder.estimate_tokens(short_prompt);
        let long_tokens = builder.estimate_tokens(&long_prompt);

        assert!(long_tokens > short_tokens);
    }

    #[test]
    fn test_validate_length() {
        let config = PromptBuilderConfig {
            system_instruction: "System".to_string(),
            max_tokens: 10,
            include_memory: false,
            max_memory_entries: 0,
            include_metadata: false,
        };

        let builder = PromptBuilder::with_config(config);

        let short_prompt = "Short";
        assert!(builder.validate_length(short_prompt).is_ok());

        let long_prompt = "A".repeat(1000);
        assert!(builder.validate_length(&long_prompt).is_err());
    }

    #[test]
    fn test_convenience_function() {
        let prompt = build_prompt("测试任务", &[]);
        assert!(prompt.contains("测试任务"));
    }

    #[test]
    fn test_format_memory_value() {
        let builder = PromptBuilder::new();

        let string_value = json!("simple string");
        assert_eq!(builder.format_memory_value(&string_value), "simple string");

        let object_value = json!({"key": "value"});
        let formatted = builder.format_memory_value(&object_value);
        assert!(formatted.contains("key"));
        assert!(formatted.contains("value"));
    }
}
