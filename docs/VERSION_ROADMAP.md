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

## 🚀 v0.2.1 - CLI 统一与云端联邦 (进行中)

**分支**: `feature/0.2.1`
**状态**: 规划完成，准备实施

### 4 个并行开发 Team

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

## 🎯 v0.2.2 - Skill 引导强化系统 (规划中)

**状态**: 概念设计完成

### 核心目标

**问题**: 默认的 Claude CLI 倾向于做"谨慎的客服"，不符合 AgentFlow "全自动 DevOps" 的场景

**解决方案**: 构建四级 Prompt 构造器，强制 Agent 变成"不废话、猛冲猛打、懂项目规则的高级工程师"

---

### 核心架构：四级 Prompt 构造器

**文件**: `rust/agentflow-core/src/llm/prompt_builder.rs` (新建)

```
Level 1 [Hardcoded]   → 核心身份与行为准则
       ↓
Level 2 [Project]      → 项目专属技能集 (AGENTFLOW.md)
       ↓
Level 3 [Memory]       → 相关历史经验 (SQLite 检索)
       ↓
Level 4 [Context]      → 运行时上下文 (Git 状态)
       ↓
Level 5 [Task]         → 用户指令
```

---

### 实施任务清单

#### Phase 1: 实现 PromptBuilder 引擎 (High Priority)

**任务**:
1. 定义结构体 `PromptBuilder`
2. 实现 `build(task: &str) -> Result<String>` 方法
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

## 🎯 总结

**AgentFlow 正在快速演进**：

- **v0.2.0** (已完成): 记忆工作流基础
- **v0.2.1** (进行中): CLI 统一 + 云端联邦
- **v0.2.2** (规划中): Skill 引导强化
- **v0.3.0** (未来): 生产就绪完整版

**关键特性**:
- 🧠 智能记忆系统
- 🌐 云端分布式架构
- 💬 AI 对话集成
- 🛠️ 一键部署
- 🎯 Skill 引导强化

**开发策略**: 并行 Team，快速迭代，持续交付 🚀
