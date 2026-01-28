# 项目开发计划书：AgentFlow v5 (Pure CLI Orchestration)
## 1. 核心架构确认
**原则**：
- **零 API 调用**：Rust 代码中不出现 `reqwest::post` 到 LLM Provider 的逻辑。
- **进程池化**：所有的 AI 思考都通过 `tokio::process::Command` 拉起外部 CLI 进程完成。
- **透传模式**：Rust 负责 IO 转发、状态管理和沙箱隔离。AI 能力完全取决于 CLI 里面的配置（无论是 Anthropic 还是 GLM）。
**双角色 CLI**：
1.  **云端 CLI (The Planner)**：快速解析指令，输出 JSON。
2.  **本地 CLI (The Worker)**：执行代码，运行测试，输出 Diff。
---
## 2. 第一阶段：通用 CLI 执行器
这是整个系统的核心引擎，必须能稳定地“榨干” CLI 的性能。
### Task 1.1: 异步进程封装
**文件**: `rust-core/src/executor/cli_runner.rs`
**需求细节**:
- **交互方式**：使用 `Stdio::piped()`。
- **输入**：将 System Prompt 和 User Task 写入子进程的 `Stdin`。
- **输出流处理**：
    - **实时流**：监听 `Stdout`，逐行解析。
    - **JSON 提取**：Claude/CLI 可能会在 Markdown 代码块中返回 JSON。需要一个简单的状态机来提取 `<thinking>...</thinking>` 或 ```json ... ``` 中的内容。
- **环境隔离**：
    - `env("ANTHROPIC_API_KEY", ...)` -> 继承父进程环境。
    - `current_dir("/workspace/repo")` -> 强制切换工作目录。
**代码逻辑**:
```rust
pub struct CliRunner {
    command: String, // e.g., "claude" or "glm-cli"
    args: Vec<String>,
}
impl CliRunner {
    pub async fn execute(&self, prompt: &str) -> Result<String, ExecutionError> {
        let mut child = Command::new(&self.command)
            .args(&self.args)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            // 关键：确保 Worker 在沙箱中运行
            .current_dir(self.workspace_path) 
            .spawn()?;
        
        // 写入 Prompt ...
        // 读取 Output ...
    }
}
```
### Task 1.2: 流式响应截获与转发
**需求**:
- 用户在手机上通过智谱 Webhook 发送任务后，需要实时看到进度。
- Rust 需要将 CLI 的 Stdout 逐行通过 WebSocket 推送给前端。
- 实现 `Stream<Item = String>` trait，方便与其他 Tokio 组件组合。
---
## 3. 第二阶段：云端 Master 的 CLI 调度
云端 Master 接收到 Webhook 后，不再做简单的正则匹配，而是**拉起一个 CLI 进程做“意图识别”**。
### Task 2.1: 云端意图识别
**文件**: `rust-core/src/cloud/planner.rs`
**逻辑**:
1.  Webhook 接收文本 `text`。
2.  构建 Prompt：`"你是一个调度器。请将用户指令转换为 JSON Task。用户指令：{{text}}"`。
3.  **关键点**：云端 Master 调用 `CliRunner`，启动一个轻量级的 CLI 进程（例如指定使用 `claude-3-haiku` 或你的 GLM 快速模式）。
4.  等待 CLI 返回 JSON。
5.  解析 JSON，得到 `{ target_node, repo, command }`。
**Prompt 配置**:
云端 CLI 的 Prompt 应该极其简短，追求速度：
```text
Role: Task Dispatcher.
Input: {{user_text}}
Output JSON: { "node": "home_win", "repo": "diveadstra", "cmd": "run benchmark dx12" }
Do not output markdown. JSON only.
```
---
## 4. 第三阶段：本地 Master 的执行流
本地 Master 接收到云端发来的 Task 后，拉起**全功能的 CLI**（Claude Skill + GLM Code Plans）。
### Task 3.1: 沙箱执行上下文
**文件**: `rust-core/src/local/sandbox.rs`
**需求**:
- **Pre-hook**: 执行前，运行 `git status`。
- **Inject Prompt**: 将云端传来的 `cmd` 和本地的 Git Context 拼接成完整 Prompt。
    ```text
    [Context from Cloud]
    Task: Run benchmark dx12
    [Local Context]
    Git Diff: ...
    Current File: render_loop.cpp
    
    [Request]
    Please perform the task using Claude Skills / GLM Code Plans.
    ```
- **Execution**: 调用 `CliRunner`。这里的 CLI 配置为使用你的最强模型（`claude-3.5-sonnet` 或 `glm-4-code-plans-max`）。
### Task 3.2: 结果回写
**需求**:
- CLI 结束后，检查工作目录是否有文件变更。
- 若有，执行 `git commit -m "AgentFlow: Executed remote task"`.
- 将 Commit Hash 回传给云端 Master，再由云端 Master 推送给手机端显示。
---
## 5. 第四阶段：部署与配置
因为依赖 CLI，部署脚本需要自动检查并引导用户配置 CLI。
### Task 4.1: 环境检查脚本
**文件**: `scripts/check_env.sh`
**功能**:
- 检查 `claude` (或你的 GLM CLI 封装) 是否在 PATH 中。
- 检查是否配置了 API Key（通过尝试运行 `claude --version` 或简单 echo 测试）。
- 提示用户：`请确保本地 CLI 已登录。AgentFlow 不会触碰你的 Key。`
### Task 4.2: 模型配置映射
**文件**: `agentflow.toml`
```toml
[cli]
# 定义不同的任务使用不同的 CLI 配置
[cli.planner] # 云端用的
command = "claude"
args = ["-m", "claude-3-haiku", "--no-prompt"]
[cli.worker]  # 本地用的
command =claude" # 如果是 GLM 封装，这里改成 "glm-cli"
args = ["-m", "claude-3.5-sonnet", "--dangerously-skip-permissions"] # 或 GLM Code Plans 参数
```
---
## 6. 并行开发任务分配
### Team A: 进程与流 (Rust Core)
- 实现 `CliRunner` (Task 1.1).
- 实现非阻塞的 Stdout 逐行读取器。
- 实现 `Stream` 到 WebSocket 的桥接。
### Team B: 业务逻辑 (Orchestration)
- 实现云端 Webhook -> CLI Planner 的流程 (Task 2.1).
- 实现本地 Task -> CLI Worker -> Git Commit 的流程 (Task 3.1).
### Team C: 部署与配置
- 编写 `check_env.sh`.
- 更新 `README.md`，说明 CLI 依赖。
---
## 7. 验收标准
1.  **云端测试**：
    - 向 Webhook 发送 "hello"。
    - 观察云端服务器日志，看到 `Spawned process: claude -m haiku ...`。
    - 返回正确的 JSON 解析结果。
2.  **本地测试**：
    - 云端下发任务。
    - 本地看到 `Spawned process: claude -m sonnet ...`。
    - 任务完成后，本地 Git 仓库自动新增了 Commit。
3.  **安全性**：
    - `ps aux | grep agentflow` 只能看到 AgentFlow 进程和 CLI 进程。
    - CLI 进程的环境变量中包含 API Key，但 AgentFlow 的 Rust 代码中从未出现该 Key 的明文。
---
**指令提示**:
Agent 请注意，核心难点在于 **异步进程通信的稳定性**。
请务必处理：
1.  子进程崩溃（非 0 退出码）的情况，并捕获 Stderr 中的错误信息。
2.  Stdout 流可能包含 ANSI 颜色代码，需要清洗后再做 JSON 解析。
3.  不要让 CLI 进程变成僵尸进程。
保持 CLI 架构的纯粹性，开始执行！
