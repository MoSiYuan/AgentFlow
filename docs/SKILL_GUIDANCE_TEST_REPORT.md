# Skill 引导系统 - 集成测试报告

**日期**: 2026-01-28
**版本**: v0.2.1
**状态**: 实现完成，待测试验证

---

## 📋 测试概述

本文档定义了 Skill 引导系统的集成测试用例，用于验证 AgentFlow 是否从"谨慎的客服"成功转变为"不废话的执行引擎"。

---

## 🎯 测试目标

### 核心验证点

1. **强系统提示词生效**
   - Agent 不再询问"你想运行哪一个？"
   - Agent 直接执行最合理的操作
   - Agent 优先使用 Bash 而非生成代码片段

2. **自动记忆检索工作**
   - 自动检索相关历史经验（Top 3）
   - 记忆正确注入到 Prompt
   - 记忆格式清晰易读

3. **项目级配置生效**
   - AGENTFLOW.md 被正确读取
   - 项目特定规则被遵循
   - 配置不存在时不影响功能

---

## 🧪 测试用例

### Test Case 1: 强制执行验证

**目标**: 验证 Agent 不再询问，直接执行

**测试步骤**:

1. **准备测试环境**
   ```bash
   cd /path/to/test/workspace
   export ANTHROPIC_API_KEY=your_key
   ```

2. **创建简单测试场景**
   - 创建一个可运行的测试脚本
   - 确保 Agent 有权限执行

3. **执行测试**
   ```bash
   # 启动 AgentFlow
   agentflow task create "跑个测试"

   # 观察输出
   ```

**预期结果** ✅:
```
✓ Agent 直接执行测试命令
✓ 输出: "I will run the tests..."
✗ 不应该输出: "你想运行哪一个测试？"
✗ 不应该输出: "我可以帮你运行测试..."
```

**验收标准**:
- ✅ Agent 使用 `I will...` 而非 `I suggest...`
- ✅ Agent 直接调用 Bash 执行测试
- ✅ 不询问用户确认

**实际结果**: ⏳ 待测试

---

### Test Case 2: 项目规则遵循验证

**目标**: 验证 Agent 遵循 AGENTFLOW.md 中定义的规则

**测试步骤**:

1. **创建测试项目**
   ```bash
   mkdir test_project && cd test_project
   ```

2. **创建 AGENTFLOW.md**
   ```markdown
   # Test Project Configuration

   ## Build System
   - 使用 `test-build.sh` 脚本进行编译
   - **不要**使用 `make` 或 `cargo build`

   ## Critical Rules
   - 编译前必须先清理缓存
   - 不要修改 `generated/` 目录
   ```

3. **执行测试**
   ```bash
   agentflow task create "编译一下项目"
   ```

**预期结果** ✅:
```
✓ Agent 读取 AGENTFLOW.md
✓ Agent 调用 `./test-build.sh` 而非 `make`
✓ Agent 先执行清理缓存
✗ 不应该调用 `make build`
✗ 不应该使用 `cargo build`
```

**验收标准**:
- ✅ Prompt 中包含"项目专属配置"
- ✅ 执行的命令符合 AGENTFLOW.md 定义
- ✅ 不违反明确禁止的操作

**实际结果**: ⏳ 待测试

---

### Test Case 3: 历史经验复用验证

**目标**: 验证 Agent 能够检索并利用历史经验

**测试步骤**:

1. **准备记忆数据**
   ```bash
   # 手动插入历史记忆
   sqlite3 memory.db "
   INSERT INTO memories (key, value, category, timestamp)
   VALUES (
     'shader_cache_fix',
     '{\"summary\": \"解决 Shader 编译错误\",\"
       solution\": \"删除 Cache/ 目录\",\"
       date\": \"2026-01-27\"}',
     'Result',
     strftime('%s', 'now', '-1 day')
   );
   "
   ```

2. **创建触发场景**
   ```bash
   # 制造一个 Shader 错误
   mkdir Cache && echo "corrupted" > Cache/shader.bin
   ```

3. **执行测试**
   ```bash
   agentflow task create "游戏崩溃了，修一下"
   ```

**预期结果** ✅:
```
✓ Agent 检索到相关记忆（Shader Cache 问题）
✓ Prompt 中包含历史经验
✓ Agent 主动尝试删除 Cache/ 目录
✗ 不应该重新编译整个项目
✗ 不应该询问用户如何处理
```

**验收标准**:
- ✅ `build_with_memory_search()` 被调用
- ✅ 检索到 Top 3 相关记忆
- ✅ Agent 基于历史经验采取行动
- ✅ 记忆信息在 Prompt 中清晰展示

**实际结果**: ⏳ 待测试

---

## 🔍 深度测试

### Test Case 4: Git 强制提交验证

**目标**: 验证 Git Mandatory 指令生效

**测试步骤**:
1. 修改一个代码文件
2. 执行任务: "修复一个 bug"
3. 观察是否自动提交

**预期结果** ✅:
```
✓ Agent 修复 bug 后执行 `git add`
✓ Agent 执行 `git commit -m "..."`
✗ 不应该只修改代码不提交
```

---

### Test Case 5: Self-Healing 验证

**目标**: 验证错误时自动修复

**测试步骤**:
1. 创建一个会失败的命令
2. 执行任务
3. 观察错误处理

**预期结果** ✅:
```
✓ 命令失败后 Agent 尝试修复
✓ Agent 分析错误原因
✓ Agent 尝试替代方案
✗ 不应该直接停止并报错
```

---

## 📊 测试矩阵

| 测试用例 | 优先级 | 复杂度 | 预计时间 | 状态 |
|---------|--------|--------|----------|------|
| Test Case 1: 强制执行 | 🔴 High | 🟢 Low | 10 min | ⏳ 待测 |
| Test Case 2: 项目规则 | 🔴 High | 🟡 Medium | 20 min | ⏳ 待测 |
| Test Case 3: 历史经验 | 🟡 Medium | 🟡 Medium | 15 min | ⏳ 待测 |
| Test Case 4: Git 提交 | 🟡 Medium | 🟢 Low | 10 min | ⏳ 待测 |
| Test Case 5: Self-Healing | 🟢 Low | 🔴 High | 30 min | ⏳ 待测 |

---

## 🛠️ 测试环境要求

### 必需组件

- ✅ Rust 编译器（1.70+）
- ✅ AgentFlow 二进制文件（v0.2.1+）
- ✅ SQLite3（用于手动验证记忆）
- ✅ ANTHROPIC_API_KEY 环境变量

### 可选组件

- Git（用于 Test Case 4）
- 测试项目（用于 Test Case 2）

---

## 📝 测试执行指南

### 快速测试（5 分钟）

```bash
# 1. 验证 PromptBuilder 编译
cd rust/agentflow-core
cargo test prompt_builder

# 2. 验证系统提示词
grep "AgentFlow Execution Engine" src/executor/prompt_builder.rs

# 3. 验证自动记忆检索
grep "build_with_memory_search" src/executor/prompt_builder.rs

# 4. 验证项目配置读取
grep "load_project_config" src/executor/prompt_builder.rs
```

### 完整测试（1 小时）

按照上述 5 个测试用例依次执行，记录结果。

---

## 🐛 已知问题

### 当前限制

1. **Cargo 不可用**
   - 环境: 当前开发环境未安装 cargo
   - 影响: 无法编译测试
   - 解决方案: 需要先安装 Rust 工具链

2. **缺少集成测试框架**
   - 状态: 单元测试存在，但集成测试未实现
   - 计划: v0.2.2 添加完整集成测试

---

## ✅ 成功标准

### 最小可行性

- ✅ [ ] Test Case 1 通过（强制执行）
- ✅ [ ] Test Case 2 通过（项目规则）

### 完整功能

- ✅ [ ] 所有 5 个测试用例通过
- ✅ [ ] 性能测试通过（Prompt 构建 < 100ms）
- ✅ [ ] 内存使用正常（< 150MB）

---

## 📅 测试计划

| 阶段 | 时间 | 负责人 | 状态 |
|------|------|--------|------|
| 环境准备 | 2026-01-28 | - | ⏳ 进行中 |
| 快速验证 | 2026-01-28 | - | ⏳ 待开始 |
| 完整测试 | Week of 2026-01-29 | - | 📅 计划中 |
| 问题修复 | As needed | - | 📅 计划中 |

---

## 🎯 下一步行动

1. **立即** (Today)
   - [ ] 安装 Rust 工具链
   - [ ] 编译 AgentFlow
   - [ ] 运行快速验证

2. **本周** (Week of 2026-01-29)
   - [ ] 执行完整测试用例
   - [ ] 记录测试结果
   - [ ] 修复发现的问题

3. **下周** (Week of 2026-02-05)
   - [ ] 编写集成测试框架
   - [ ] 添加性能基准测试
   - [ ] 准备 v0.2.2 规划

---

## 📞 联系方式

如有问题或建议，请：
- 提交 GitHub Issue
- 查看 [docs/VERSION_ROADMAP.md](VERSION_ROADMAP.md)
- 参考 [docs/plan/v0.2.2迭代计划.md](plan/v0.2.2迭代计划.md)

---

**最后更新**: 2026-01-28
**文档版本**: 1.0
