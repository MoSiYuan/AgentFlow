# AgentFlow 版本路线图

**最后更新**: 2026-01-28
**当前分支**: feature/0.2.1

---

## 📅 版本规划概览

```
v0.2.0 (已完成) → v0.2.1 (进行中) → v0.2.2 (规划中) → v0.3.0 (未来)
    ↓                  ↓                  ↓
  记忆工作流        CLI + 云端联邦      Skill引导强化    生产就绪
```

---

## 🎯 v0.2.0 - 记忆工作流系统 ✅ 已完成

**状态**: 已合并到 main 分支

### 核心特性

- ✅ **MemoryProcessor** - 任务完成后自动整理记忆
- ✅ **简化 MemoryCore** - 关键词匹配替代伪向量嵌入
- ✅ **依赖 Claude CLI** - 完全依赖 Claude 的语义理解
- ✅ **异步非阻塞** - 后台处理，不影响主任务
- ✅ **自动分类存储** - Execution / Result / Error

### 交付物

- `rust/agentflow-core/src/executor/memory_processor.rs`
- `rust/agentflow-core/src/memory/mod.rs` (简化版)
- `docs/MEMORY_WORKFLOW_DESIGN.md`

### 工作流

```
Task Execution → Claude CLI 分析 → 提取关键信息 → 存储到记忆
                                                              ↓
                                        下次任务检索相关记忆
```

---

## 🚀 v0.2.1 - CLI 统一与云端联邦 + Skill 引导增强 (进行中)

**分支**: `feature/0.2.1`
**状态**: Skill 引导增强 ✅ 已完成，CLI/云端联邦规划完成

### ✅ Skill 引导增强系统 (已完成)

**完成日期**: 2026-01-28

**核心改进**: 将 AgentFlow 从"谨慎的客服"转变为"不废话的执行引擎"

#### 实现功能

1. **强系统提示词 (Level 1)** ✅
   - **文件**: [rust/agentflow-core/src/executor/prompt_builder.rs:45](../rust/agentflow-core/src/executor/prompt_builder.rs#L45)
   - **改动**: 完全重写默认系统指令
   - **核心原则**:
     - Action First: 不说 "I suggest"，直接 "I will..."
     - Tool Usage: 优先使用 Bash 执行命令
     - Git Mandatory: 修改后必须 `git add && git commit`
     - Self-Healing: 报错时尝试修复，不直接停止
     - No Questions: 不问"你想运行哪一个？"，直接执行

2. **自动记忆检索 (Level 3)** ✅
   - **文件**: [rust/agentflow-core/src/executor/prompt_builder.rs:298](../rust/agentflow-core/src/executor/prompt_builder.rs#L298)
   - **新方法**: `build_with_memory_search(task, memory_core)`
   - **功能**: 自动检索 Top 3 相关记忆并注入 Prompt
   - **特点**: 异步接口，优雅的错误处理

3. **项目级配置 (Level 2)** ✅
   - **文件**: [rust/agentflow-core/src/executor/prompt_builder.rs:151](../rust/agentflow-core/src/executor/prompt_builder.rs#L151)
   - **配置文件**: [templates/AGENTFLOW.md.example](../templates/AGENTFLOW.md.example)
   - **功能**: 读取 `{workspace}/AGENTFLOW.md` 项目配置
   - **内容**: 构建系统、测试工作流、关键技能、调试策略

#### 四级 Prompt 架构

```
Level 1 [Hardcoded]   → ✅ 强系统提示词 (AgentFlow Execution Engine)
       ↓
Level 2 [Project]      → ✅ AGENTFLOW.md 项目配置
       ↓
Level 3 [Memory]       → ✅ 自动记忆检索 (Top 3)
       ↓
Level 4 [Task]         → ✅ 用户指令
```

#### 交付物

- `rust/agentflow-core/src/executor/prompt_builder.rs` (增强版)
- `templates/AGENTFLOW.md.example` (配置模板，61 行)
- 完整的 Rust Doc 注释
- 向后兼容现有功能

#### 对 v0.2.2 的影响

由于 v0.2.1 已实现核心功能，v0.2.2 可简化为：
- ✅ **移除**: Phase 1 (PromptBuilder 引擎) - 已完成
- ✅ **移除**: Phase 2 (强系统提示词) - 已完成
- ✅ **移除**: Phase 4 (Memory Injection) - 已完成
- 🔄 **保留增强**: Phase 3 (AGENTFLOW.md) - 基础版已完成
- 🆕 **新增**: Git 状态注入 (Level 4)

---

### 4 个并行开发 Team (原始计划)

#### Team A: CLI & 配置层
- 统一 `agentflow` 命令
- 多模式支持 (local/cloud/planner-only)
- TOML 配置系统
- 预留节点 API 端点

**产出**: 4 个规划文档

#### Team B: Memory & Git 集成
- Markdown Chunker（按标题切分）
- Embedder Trait（Claude CLI + GLM-4）
- 文件监听与自动索引
- Git 变化自动回写 Markdown
- Prompt 记忆注入

**产出**: 完整实现计划（6 个阶段）

#### Team C: 云端/边缘联邦
- 节点注册与身份认证
- WebSocket Hello/Heartbeat 协议
- 智谱清言 Webhook 接入
- 5 层安全模型
- 任务路由与分发

**产出**: 4 个规划文档

#### Team D: 打包与文档
- 5 平台交叉编译（Linux/macOS/Windows × AMD64/ARM64）
- 一键安装脚本
- 智谱集成指南
- 完整配置参考
- 测试验证报告

**产出**: 7 个实际文件（3,360 行代码/文档）

### 交付时间线

| 阶段 | 时间 | 依赖 |
|------|------|------|
| Week 1: Team A (CLI) | Days 1-7 | 无 |
| Week 2: Team B/C (Memory + Cloud) | Days 8-14 | Team A |
| Week 2: Team D (Polish) | Days 8-14 | 所有 Team |

### 详细文档

- `docs/plan/v0.2.1_TASK_BREAKDOWN.md` - 总任务拆解
- `docs/plan/v0.2.1迭代计划.md` - 原始迭代计划
- `docs/plan/team-a-*.md` (4个) - Team A 文档
- `docs/plan/team-b-*.md` (1个) - Team B 文档
- `docs/plan/team-c-*.md` (4个) - Team C 文档
- `docs/ZHIPU_INTEGRATION.md` - 智谱集成指南
- `docs/CONFIGURATION.md` - 配置参考

---

## 🎯 v0.2.2 - Skill 引导强化系统 (待调整)

**状态**: ⚠️ 部分功能已在 v0.2.1 实现，需要重新规划

**重要变更**: 由于 v0.2.1 提前实现了 Skill 引导核心功能，v0.2.2 需要重新定位

### 核心目标 (已部分实现)

**原问题**: 默认的 Claude CLI 倾向于做"谨慎的客服"，不符合 AgentFlow "全自动 DevOps" 的场景

**解决方案**: ✅ 已在 v0.2.1 实现核心功能

### v0.2.1 已完成功能

| 功能 | 状态 | 说明 |
|------|------|------|
| **Phase 1**: PromptBuilder 引擎 | ✅ v0.2.1 | 已实现完整架构 |
| **Phase 2**: 强系统提示词 | ✅ v0.2.1 | 5个 Core Directives |
| **Phase 3**: AGENTFLOW.md | ⚠️ v0.2.1 | 基础版已实现 |
| **Phase 4**: Memory Injection | ✅ v0.2.1 | `build_with_memory_search()` |

### v0.2.2 新目标

#### 🆕 新增功能

1. **Git 状态注入 (Level 4)**
   - 获取当前 Git 分支、状态、未提交改动
   - 注入到 Prompt 作为运行时上下文
   - 帮助 Agent 理解当前工作状态

2. **增强 AGENTFLOW.md (Phase 3 完善)**
   - 添加更多配置选项
   - 支持条件规则（基于文件类型、路径等）
   - 提供更丰富的示例

3. **Few-Shot Examples 增强**
   - 改进历史记忆的格式化
   - 添加时间戳和相关性评分
   - 优化 Prompt 结构

4. **完整测试验证**
   - Test Case 1: 强制执行（"跑个测试"）
   - Test Case 2: 遵循项目规则（msbuild vs make）
   - Test Case 3: 基于经验（Shader Cache）

---

### 实施任务清单 (调整后)

#### Phase 1: Git 状态注入 (High Priority)

**文件**: `rust/agentflow-core/src/executor/prompt_builder.rs`

**任务**:
1. 添加 `build_with_git_context()` 方法
2. 调用 `git2` crate 获取状态
3. 格式化 Git 信息注入 Prompt

**示例输出**:
```markdown
## Git 状态
- 分支: feature/new-ui
- 未提交文件: 3 个
- 最近提交: feat: add user authentication (2h ago)
```

#### Phase 2: AGENTFLOW.md 增强 (Medium Priority)

**文件**: `templates/AGENTFLOW.md.example`

**任务**:
1. 添加更多配置示例
2. 支持路径匹配规则
3. 提供常见项目模板

#### Phase 3: Few-Shot Examples 优化 (Medium Priority)

**任务**:
1. 改进记忆格式化（添加时间戳）
2. 相关性评分显示
3. 优化 Token 使用

#### Phase 4: 集成测试 (High Priority)

**文件**: `rust/agentflow-core/tests/skill_guidance_integration.rs`

**任务**:
1. Test Case 1: 强制执行验证
2. Test Case 2: 项目规则遵循验证
3. Test Case 3: 经验复用验证

---

### 旧版任务清单 (已废弃)

#### ~~Phase 1: 实现 PromptBuilder 引擎~~ ✅ v0.2.1

**任务**:
3. 读取 `{workspace_path}/AGENTFLOW.md`
4. 调用 `MemoryCore::search(task, top_k=3)`
5. 获取 Git 状态
6. 按 Level 1-5 顺序拼接字符串

**代码结构**:
```rust
pub struct PromptBuilder {
    base_identity: &'static str,
    workspace: PathBuf,
    memory: Arc<MemoryCore>,
}

impl PromptBuilder {
    pub async fn build(&self, user_task: &str) -> Result<String> {
        // 拼接五层 Prompt
    }
}
```

#### Phase 2: 定义强 System Prompt (Level 1) (High Priority)

**文件**: `rust/agentflow-core/src/llm/prompts.rs` (新建)

**核心要求**:
- **Role**: 声明是 "AgentFlow Execution Engine"，而非 Chat Assistant
- **Directives**:
  - Action First: 不说 "I suggest"，直接 "I will..."
  - Tool Usage: 优先使用 `Bash` 执行命令
  - Git Mandatory: 修改后必须 `git add && git commit`
  - Self-Healing: 错误时尝试修复
  - Protocol: 引入 CoT，要求思考步骤但不输出

#### Phase 3: 设计项目级配置 `AGENTFLOW.md` (Level 2) (High Priority)

**文件**: `templates/AGENTFLOW.md.example`

**内容结构**:
```markdown
# Project Skills

## 1. Build System
- Windows: `msbuild ...`
- Linux: `make ...`

## 2. Testing Workflow
- 不要运行 `make test`
- 正确命令：`./bin/test --suite=unit`

## 3. Critical Skills (Do's and Don'ts)
- DO: 提交前必须跑 `linter`
- DON'T: 不要修改 `third_party/` 目录
- SKILL: Shader 报错需删除 `Cache/` 目录

## 4. Debugging Strategy
- 崩溃时优先检查 `Logs/`
```

#### Phase 4: 历史记忆引导 (Level 3) (Medium Priority)

**逻辑**:
- 当用户下达指令时，`PromptBuilder` 调用检索
- 将检索结果格式化为 "Few-Shot Examples"
- 注入到 Prompt 中

**示例**:
```markdown
## Relevant Historical Experience
[2025-01-10] 优化渲染管线：合并 Draw Call，提升 20% FPS。
关键文件：RenderPass.cpp。
请参考以上经验处理当前任务。
```

---

### 验收标准

#### Test Case 1: 强制执行

**输入**: "跑个测试"

❌ Bad Output: "好的，我可以帮你运行测试，你想运行哪一个？..."

✅ Expected Output:
```bash
./bin/run-tests ...
# 测试完成，正在提交代码
git add tests/
git commit -m "Run tests: all passed"
```

#### Test Case 2: 遵循项目规则

在 `DiveAdstra` 仓库配置 `AGENTFLOW.md`，写入 "Build 命令是 `msbuild`"

**输入**: "编译一下"

✅ Expected Output: 调用 `msbuild`，而不是 `make` 或 `gcc`

#### Test Case 3: 基于经验

历史记录："解决崩溃需删除 Shader Cache"

**输入**: "游戏崩溃了，修一下"

✅ Expected Output: 主动尝试 `rm -rf Cache/`

---

### 开发优先级

1. **High**: Phase 1 (PromptBuilder) & Phase 2 (Base Prompt) - 基础底座
2. **High**: Phase 3 (`AGENTFLOW.md`) - 适配不同项目的关键
3. **Medium**: Phase 4 (Memory Injection) - 锦上添花

---

## 🔮 v0.3.0 - 生产就绪版本 (未来规划)

**预期功能**:
- 完整的云端-边缘协同
- 智谱清言无缝集成
- Skill 引导强化
- 性能监控与优化
- Web UI 控制台

---

## 📊 版本对比

| 特性 | v0.2.0 | v0.2.1 | v0.2.2 | v0.3.0 |
|------|--------|--------|--------|--------|
| **记忆系统** | ✅ 基础 | ✅ Markdown+Git | ✅ Prompt注入 | ✅ 完整 |
| **CLI** | ❌ 仅 binary | ✅ 统一命令 | ✅ Skill引导 | ✅ 完善 |
| **云端联邦** | ❌ 单节点 | ✅ 多节点 | ✅ 多节点 | ✅ 优化 |
| **智谱集成** | ❌ 无 | ✅ Webhook | ✅ Webhook | ✅ 优化 |
| **配置** | ❌ 环境变量 | ✅ TOML | ✅ AGENTFLOW.md | ✅ 完善 |
| **打包** | ❌ 手动 | ✅ 自动化 | ✅ 自动化 | ✅ Docker |
| **文档** | ⚠️ 基础 | ✅ 完整 | ✅ 完整 | ✅ 完善 |

---

## 🚀 实施优先级

### 立即行动 (v0.2.1)
1. **Team A** - CLI & Config（阻塞其他）
2. **Team D** - 打包和文档（可并行）

### 短期 (v0.2.2)
1. **PromptBuilder** 引擎实现
2. **System Prompt** 强化
3. **AGENTFLOW.md** 模板

### 中期 (v0.3.0)
1. 性能监控
2. Web UI
3. Docker 化
4. 插件系统

---

## 📁 相关文档

### 版本规划
- `docs/plan/VERSION_ROADMAP.md` - 本文档
- `docs/plan/v0.2.1_TASK_BREAKDOWN.md` - v0.2.1 详细拆解
- `docs/plan/v0.2.1迭代计划.md` - v0.2.1 原始计划
- `docs/plan/v0.2.2_skill_guidance_system.md` - v0.2.2 详细计划

### Team 文档
- `docs/plan/team-a-*.md` (4个) - Team A 文档
- `docs/plan/team-b-*.md` (1个) - Team B 文档
- `docs/plan/team-c-*.md` (4个) - Team C 文档

### 实现指南
- `docs/MEMORY_WORKFLOW_DESIGN.md` - v0.2.0 设计
- `docs/ZHIPU_INTEGRATION.md` - v0.2.1 智谱集成
- `docs/CONFIGURATION.md` - v0.2.1 配置参考

---

## 🤖 v0.2.3 - 工作流自动化与分布式调度 (规划中)

**分支**: `feature/0.2.3` (待创建)
**状态**: 需求分析完成，迭代计划已制定
**前置版本**: v0.2.1 (Skill 引导系统 ✅)

### 核心理念

**"Rust 是指挥官和情报官，Claude CLI 是执行士兵"**

- **Rust**: 负责环境准备、情报收集、Prompt 组装、工作流控制
- **Claude CLI**: 负责代码生成、文件修改、命令执行

### 版本定位

v0.2.3 是 **工作流自动化版本**，在 v0.2.1 Skill 引导基础上：
- ✅ 自动化 Git 工作流（保护 main 分支）
- ✅ 实现分布式任务调度（DAG）
- ✅ 增强技能清单（标准化命令）
- ⏸️ 暂缓 Tree-sitter 索引（性能优化，非必需）

---

### 三阶段实施计划

#### Phase 1: Git 工作流自动化 (Days 1-3) 🔴 P0

**目标**: 保护 `main` 分支，所有 CLI 执行在隔离的 feature 分支

**核心功能**:
- 自动创建 `agentflow/task-{timestamp}` 分支
- 失败时自动回滚（删除分支）
- 成功时保留或推送分支
- 日志记录完整

**交付物**:
- `rust/agentflow-core/src/git/branch_manager.rs`
- 集成到 TaskExecutor
- 单元测试

---

#### Phase 2: DAG 任务调度器 (Days 4-10) 🔴 P0

**目标**: 实现分布式任务编排，支持依赖管理和并发执行

**核心功能**:
- DAG 数据结构与解析
- HTTP 任务分发 API
- 边缘节点执行器
- 状态机与依赖触发

**交付物**:
- `rust/agentflow-core/src/scheduler/` 模块
  - `dag.rs` - DAG 数据结构
  - `distributor.rs` - 任务分发
  - `coordinator.rs` - 协调器
  - `executor.rs` - 边缘节点执行器
- Cloud Master HTTP 服务
- 集成测试

---

#### Phase 3: 动态技能清单增强 (Days 11-12) 🟡 P1

**目标**: 标准化命令，避免 LLM "发明"命令

**核心功能**:
- 技能定义整合到 `AGENTFLOW.md`
- 技能参数模板展开
- Prompt 注入

**调整**: 不使用独立的 `skills.json`，而是增强 `AGENTFLOW.md`

**交付物**:
- `rust/agentflow-core/src/skills/` 模块
- 增强的 AGENTFLOW.md 模板
- PromptBuilder 集成

---

### 延后功能

#### ❌ Tree-sitter 符号索引

**延后理由**:
- 非核心功能（性能优化）
- 实现成本高（Tree-sitter 集成复杂）
- 中小型项目不需要
- 维护成本高（跟随语言语法更新）

**计划**: 延后到 v0.3.x，触发条件：
- 项目代码 > 10 万行
- CLI 执行时间 > 30 秒
- Token 成本显著上升

---

### 详细文档

- [v0.2.3 迭代计划](plan/v0.2.3_ITERATION_PLAN.md) - 详细实施计划
- [v0.2.3 决策记录](plan/v0.2.3_DECISION_RECORD.md) - 需求分析与决策

---

## 🎯 总结

**AgentFlow 正在快速演进**：

- **v0.2.0** (✅ 已完成): 记忆工作流基础
- **v0.2.1** (🔄 进行中): CLI 统一 + 云端联邦 + **Skill 引导增强 ✅**
- **v0.2.2** (⚠️ 待调整): Git 状态注入 + 完整测试
- **v0.2.3** (📅 规划中): 工作流自动化 + 分布式调度
- **v0.3.0** (🔮 未来): 生产就绪完整版

**关键特性**:
- 🧠 智能记忆系统
- 🌐 云端分布式架构
- 💬 AI 对话集成
- 🛠️ 一键部署
- 🎯 **Skill 引导强化** (v0.2.1 新增 ✅)
  - ✅ 强系统提示词（Action First）
  - ✅ 自动记忆检索（Top 3）
  - ✅ 项目级配置（AGENTFLOW.md）
- 🔧 **工作流自动化** (v0.2.3 计划中)
  - 🔄 Git 分支管理
  - 🔄 DAG 任务调度
  - 🔄 动态技能清单

**最新进展** (2026-01-28):
- ✅ 完成 v0.2.1 Skill 引导增强系统
- ✅ PromptBuilder 升级为四级架构
- ✅ 从"谨慎客服"转变为"执行引擎"
- ✅ 完成 v0.2.3 需求分析与迭代计划
- 📅 v0.2.2 需要重新规划（部分功能已在 v0.2.1 实现）

**开发策略**: 并行 Team，快速迭代，持续交付 🚀
