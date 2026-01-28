# AgentFlow v3 - 记忆整理工作流设计

**日期**: 2026-01-28
**架构**: Master = Worker 单进程 + Claude CLI

---

## 🎯 设计目标

实现完整的任务执行记忆闭环，依赖 Claude CLI 的语义理解能力，自动整理和存储任务执行过程中的关键信息。

---

## 📊 工作流架构

### 完整流程图

```
┌─────────────────────────────────────────────────────────┐
│  用户请求                                                │
│  "帮我写一个快速排序算法"                                 │
└────────────────┬────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────────────┐
│  TaskExecutor.execute()                                  │
│  - 执行 claude 命令                                       │
│  - 捕获 stdout/stderr                                    │
└────────────────┬────────────────────────────────────────┘
                 ↓
        ┌────────┴────────┐
        │   任务执行完成   │
        │  ExitCode: 0    │
        └────────┬────────┘
                 ↓
┌─────────────────────────────────────────────────────────┐
│  MemoryProcessor (记忆整理 Worker)                        │
│  - 分析执行结果                                          │
│  - 提取关键信息                                          │
│  - 生成结构化记忆                                        │
└────────────────┬────────────────────────────────────────┘
                 ↓
        调用 claude 整理记忆
                 ↓
┌─────────────────────────────────────────────────────────┐
│  MemoryCore                                              │
│  - 存储执行记录                                          │
│  - 存储错误信息                                          │
│  - 存储成功经验                                          │
└─────────────────────────────────────────────────────────┘
                 ↓
        下次任务执行时检索相关记忆
```

---

## 🔧 核心组件设计

### 1. TaskExecutor 增强

**文件**: `rust/agentflow-core/src/executor/mod.rs`

```rust
pub struct TaskExecutor {
    workspace_path: PathBuf,
    timeout: Duration,
    memory_processor: Option<MemoryProcessor>,  // 新增
}

impl TaskExecutor {
    pub async fn execute(&self, task: &str) -> Result<ExecutionResult> {
        // 1. 执行 claude 命令
        let result = self.execute_claude(task).await?;

        // 2. 任务完成后，触发记忆整理
        if let Some(ref processor) = self.memory_processor {
            if let Err(e) = processor.process_task_result(task, &result).await {
                warn!("记忆整理失败: {}", e);
            }
        }

        Ok(result)
    }
}
```

### 2. MemoryProcessor (新组件)

**文件**: `rust/agentflow-core/src/executor/memory_processor.rs`

```rust
/// 记忆整理处理器
///
/// 在任务执行完成后，自动调用 Claude CLI 分析结果并整理记忆
pub struct MemoryProcessor {
    memory_core: Arc<MemoryCore>,
    workspace_path: PathBuf,
    timeout: Duration,
}

impl MemoryProcessor {
    /// 处理任务执行结果
    pub async fn process_task_result(
        &self,
        task: &str,
        result: &ExecutionResult,
    ) -> Result<()> {
        // 1. 构建记忆整理 Prompt
        let prompt = self.build_memory_prompt(task, result)?;

        // 2. 调用 Claude CLI 分析
        let analysis = self.analyze_with_claude(&prompt).await?;

        // 3. 解析分析结果
        let memory_entries = self.parse_analysis(&analysis)?;

        // 4. 存储到 MemoryCore
        for entry in memory_entries {
            self.memory_core.index(entry).await?;
        }

        Ok(())
    }

    /// 构建记忆整理 Prompt
    fn build_memory_prompt(&self, task: &str, result: &ExecutionResult) -> Result<String> {
        Ok(format!(
            r#"你是一个记忆整理专家。请分析以下任务执行结果，提取关键信息。

## 任务描述
{}

## 执行结果
退出码: {:?}
成功: {}

## 输出内容
{}

## 错误输出
{}

请按以下 JSON 格式返回分析结果：
{{
  "summary": "任务执行的简要总结（1-2句话）",
  "key_learnings": ["学到的关键点1", "关键点2"],
  "errors": ["遇到的错误1（如有）"],
  "solutions": ["解决方案1（如有）"],
  "tags": ["标签1", "标签2"],
  "category": "Execution|Result|Error"
}}"#,
            task,
            result.exit_code,
            result.success,
            result.stdout,
            result.stderr
        ))
    }

    /// 调用 Claude CLI 分析
    async fn analyze_with_claude(&self, prompt: &str) -> Result<String> {
        let mut cmd = Command::new("claude");
        cmd.arg(prompt)
            .env_clear()
            .envs(std::env::vars())
            .current_dir(&self.workspace_path);

        let output = cmd.output().await?;
        Ok(String::from_utf8_lossy(&output.stdout).to_string())
    }

    /// 解析 Claude 返回的分析结果
    fn parse_analysis(&self, analysis: &str) -> Result<Vec<MemoryEntry>> {
        // 解析 JSON 并生成 MemoryEntry
        let parsed: serde_json::Value = serde_json::from_str(analysis)?;

        let mut entries = Vec::new();

        // 1. 存储执行记录
        if let Some(summary) = parsed.get("summary").and_then(|v| v.as_str()) {
            entries.push(MemoryEntry {
                key: format!("task:{}", uuid::Uuid::new_v4()),
                value: serde_json::json!({"summary": summary}),
                category: MemoryCategory::Execution,
                timestamp: chrono::Utc::now().timestamp(),
                expires_at: None,
                task_id: None,
            });
        }

        // 2. 存储错误信息
        if let Some(errors) = parsed.get("errors").and_then(|v| v.as_array()) {
            if !errors.is_empty() {
                entries.push(MemoryEntry {
                    key: format!("error:{}", uuid::Uuid::new_v4()),
                    value: serde_json::json!({"errors": errors}),
                    category: MemoryCategory::Error,
                    timestamp: chrono::Utc::now().timestamp(),
                    expires_at: None,
                    task_id: None,
                });
            }
        }

        // 3. 存储成功经验
        if let Some(learnings) = parsed.get("key_learnings").and_then(|v| v.as_array()) {
            entries.push(MemoryEntry {
                key: format!("learning:{}", uuid::Uuid::new_v4()),
                value: serde_json::json!({"learnings": learnings}),
                category: MemoryCategory::Result,
                timestamp: chrono::Utc::now().timestamp(),
                expires_at: None,
                task_id: None,
            });
        }

        Ok(entries)
    }
}
```

### 3. MemoryCore 简化

**文件**: `rust/agentflow-core/src/memory/mod.rs`

**改动**:
- ✅ 保留现有的表结构（包括 embedding 字段，以便未来扩展）
- ✅ 简化 `compute_embedding()` 方法（移除伪嵌入，改为简单实现或返回空）
- ✅ 依赖 Claude CLI 的语义理解，不使用外部嵌入模型

```rust
/// 简化版：移除伪嵌入，使用关键词匹配
///
/// 注意：我们不使用外部嵌入模型，而是依赖 Claude CLI 的语义理解能力
/// 未来如果需要，可以实现基于关键词的简单匹配
async fn compute_embedding(&self, _text: &str) -> Result<Vec<u8>> {
    // 返回空向量，表示我们不使用向量嵌入
    // 检索时使用 SQL LIKE 或 FTS5 全文搜索
    Ok(Vec::new())
}

/// 修改检索方法：使用关键词匹配而非向量相似度
pub async fn search(&self, query: &str, limit: usize) -> Result<Vec<MemoryEntry>> {
    // 使用 SQL LIKE 进行简单文本匹配
    let pattern = format!("%{}%", query);

    let rows = sqlx::query(
        r#"
        SELECT key, value, category, task_id, timestamp, expires_at
        FROM memories
        WHERE (value LIKE ?1 OR key LIKE ?1)
          AND (expires_at IS NULL OR expires_at > ?2)
        ORDER BY timestamp DESC
        LIMIT ?3
        "#
    )
    .bind(&pattern)
    .bind(chrono::Utc::now().timestamp())
    .bind(limit as i64)
    .fetch_all(&self.pool)
    .await
    .context("查询记忆条目失败")?;

    // 解析结果...
}
```

---

## 📝 使用示例

### 用户视角

```bash
# 1. 创建任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "实现快速排序",
    "description": "用 Rust 实现快速排序算法"
  }'

# 2. 执行任务
curl -X POST http://localhost:6767/api/v1/tasks/1/execute

# 3. 任务完成后，自动触发记忆整理
# - 分析执行结果
# - 提取关键信息
# - 存储到记忆系统

# 4. 下次执行类似任务时，自动检索相关记忆
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "优化排序算法",
    "description": "优化之前的快速排序实现"
  }'

# PromptBuilder 会自动注入之前的快速排序记忆
```

---

## 🎯 实现优先级

### 第一阶段：核心功能（必须）
1. ✅ 创建 `MemoryProcessor` 结构体
2. ✅ 实现 `process_task_result()` 方法
3. ✅ 集成到 `TaskExecutor` 的执行流程
4. ✅ 简化 `MemoryCore` 的检索逻辑

### 第二阶段：优化（可选）
1. 添加记忆去重逻辑
2. 实现记忆重要性评分
3. 自动过期清理
4. 批量记忆处理

### 第三阶段：增强（未来）
1. SQLite FTS5 全文搜索
2. 记忆可视化界面
3. 记忆导出功能
4. 跨任务记忆关联

---

## 🔒 安全考虑

1. **API Key 隔离**: MemoryProcessor 直接读取环境变量，不通过 Master 传输
2. **沙箱执行**: 记忆整理任务在独立目录执行
3. **超时控制**: 避免记忆整理任务无限期运行
4. **错误处理**: 记忆整理失败不影响主任务结果

---

## 📊 性能影响

| 指标 | 影响 | 说明 |
|------|------|------|
| **任务执行时间** | 几乎无影响 | 记忆整理在后台异步执行 |
| **内存占用** | +5-10MB | Claude CLI 额外进程 |
| **存储空间** | +1-5MB/任务 | 取决于输出大小 |
| **Claude API 调用** | +1次/任务 | 记忆整理分析 |

---

## ✅ 验收标准

1. **功能验收**
   - [ ] 任务执行完成后自动触发记忆整理
   - [ ] 记忆整理能正确提取关键信息
   - [ ] 记忆能正确存储到 MemoryCore
   - [ ] 下次任务能检索到相关记忆

2. **性能验收**
   - [ ] 记忆整理不影响主任务执行时间
   - [ ] 记忆整理任务超时自动终止
   - [ ] 并发任务记忆整理不冲突

3. **安全验收**
   - [ ] API Key 不泄露到日志
   - [ ] 记忆整理任务在沙箱中执行
   - [ ] 错误任务不影响系统稳定性

---

## 🎉 总结

这个设计方案的核心优势：

1. **零外部依赖**: 完全依赖 Claude CLI 的能力，无需额外嵌入模型
2. **自动化闭环**: 任务执行 → 记忆整理 → 记忆存储 → 自动检索
3. **架构简洁**: Master = Worker，单进程完成所有任务
4. **安全可靠**: 沙箱隔离、超时控制、错误处理完善

**下一步**: 开始实现核心组件！
