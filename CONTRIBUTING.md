# 贡献指南

感谢您对 AgentFlow 项目的关注！我们欢迎各种形式的贡献。

## 🤝 如何贡献

### 报告问题

如果您发现了 Bug 或有功能建议：

1. 搜索 [现有 Issues](https://github.com/MoSiYuan/AgentFlow/issues) 确保问题未被报告
2. 创建新 Issue，使用清晰的标题和详细的描述
3. 提供复现步骤、预期行为和实际行为
4. 附上相关日志、截图或错误信息

### 提交代码

#### 开发环境设置

1. **Fork 项目**
   ```bash
   # 在 GitHub 上 Fork 本仓库
   git clone https://github.com/YOUR_USERNAME/AgentFlow.git
   cd AgentFlow
   ```

2. **安装依赖**
   ```bash
   # Rust 工具链（如果尚未安装）
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

   # Claude CLI（用于测试）
   npm install -g @anthropic-ai/claude-code
   claude auth
   ```

3. **构建项目**
   ```bash
   cd rust
   export SQLX_OFFLINE=true
   cargo build --release
   ```

4. **运行测试**
   ```bash
   cargo test
   ```

#### 代码规范

**Rust 代码**：

- 遵循 [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- 使用 `cargo fmt` 格式化代码
- 使用 `cargo clippy` 检查代码质量
```bash
cargo fmt
cargo clippy -- -D warnings
```

**提交信息**：

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型（type）：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 重构（不是新功能也不是修复）
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具链相关

示例：
```
feat(executor): 添加任务超时控制

实现可选的任务超时机制，防止长时间运行的任务占用资源。

- 添加 task_timeout 配置项
- 实现基于 tokio::time::timeout 的超时控制
- 超时后自动发送 SIGTERM 信号

Closes #123
```

#### Pull Request 流程

1. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   # 或
   git checkout -b fix/your-bug-fix
   ```

2. **进行修改并提交**
   ```bash
   git add .
   git commit -m "feat: 添加你的功能描述"
   ```

3. **同步上游更新**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

4. **推送到你的 Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **创建 Pull Request**
   - 在 GitHub 上打开 Pull Request
   - 填写 PR 模板
   - 关联相关 Issue（如 `Closes #123`）
   - 等待代码审查

#### Pull Request 检查清单

提交 PR 前请确保：

- [ ] 代码通过所有测试 (`cargo test`)
- [ ] 代码通过 clippy 检查 (`cargo clippy`)
- [ ] 代码已格式化 (`cargo fmt`)
- [ ] 添加了必要的测试用例
- [ ] 更新了相关文档
- [ ] 提交信息符合规范
- [ ] PR 描述清晰说明了改动内容

## 📂 项目结构

```
AgentFlow/
├── rust/                          # Rust 实现
│   ├── agentflow-core/           # 核心库
│   │   ├── src/
│   │   │   ├── executor/         # 任务执行引擎
│   │   │   ├── memory/           # 记忆系统
│   │   │   ├── sandbox/          # 安全沙箱
│   │   │   └── types.rs          # 共享类型
│   │   └── Cargo.toml
│   │
│   ├── agentflow-master/         # Master 服务器
│   │   ├── src/
│   │   │   ├── routes/           # API 路由
│   │   │   └── main.rs           # 入口点
│   │   └── Cargo.toml
│   │
│   └── agentflow-helper/         # 辅助工具
│
├── docs/                         # 文档
│   ├── ARCHITECTURE.md           # 架构文档
│   ├── API_REFERENCE.md         # API 参考
│   └── ...
│
├── deployment/                   # 部署相关
│   └── examples/                # 配置示例
│
└── scripts/                      # 脚本
    ├── install.sh               # 安装脚本
    └── ...
```

## 🧪 测试指南

### 运行测试

```bash
# 运行所有测试
cargo test

# 运行特定测试
cargo test test_task_execution

# 显示测试输出
cargo test -- --nocapture

# 运行集成测试
cargo test --test '*'
```

### 添加新测试

在 `rust/agentflow-core/tests/` 或相应模块的 `tests/` 目录下添加测试文件：

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_feature() {
        let result = function_to_test();
        assert_eq!(result, expected_value);
    }
}
```

## 📖 文档贡献

文档是项目的重要组成部分！你可以：

- 修正错别字和语法错误
- 改进现有文档的清晰度
- 添加使用示例
- 翻译文档
- 补充缺失的章节

### 文档规范

- 使用清晰的中文（主 README）或英文（技术文档）
- 提供可运行的代码示例
- 添加适当的 mermaid 图表
- 保持与现有文档风格一致

## 🎯 功能开发建议

### 优先开发的功能

1. **高优先级**
   - 性能优化
   - 安全性增强
   - Bug 修复

2. **中优先级**
   - 新的 AI 模型集成
   - 更多的部署方式
   - 监控和可观测性

3. **低优先级**
   - UI 美化
   - 非核心功能
   - 实验性特性

### 提议新功能前

建议在开发新功能前：
1. 创建 Issue 讨论功能需求
2. 等待维护者反馈
3. 达成共识后再开始开发

## 🌍 社区行为准则

### 我们的承诺

为了营造开放和友好的环境，我们承诺：
- 尊重不同观点和经验
- 使用欢迎和包容的语言
- 优雅地接受建设性批评
- 关注对社区最有利的事情

### 不可接受的行为

- 使用性别化语言或图像
- 人身攻击或政治攻击
- 公开或私下骚扰
- 未经许可发布他人私人信息
- 其他不专业或不适当的行为

## 📜 许可证

通过贡献代码，您同意您的贡献将使用 [MIT License](LICENSE) 进行许可。

## 🙋 获取帮助

如果您有任何问题：

- 查看 [文档](README.md)
- 搜索 [Issues](https://github.com/MoSiYuan/AgentFlow/issues)
- 在 [Discussions](https://github.com/MoSiYuan/AgentFlow/discussions) 提问
- 联系维护者

## ⭐ 成为贡献者

所有贡献者将被添加到 [CONTRIBUTORS.md](CONTRIBUTORS.md) 文件中。

---

**再次感谢您的贡献！** 🎉
