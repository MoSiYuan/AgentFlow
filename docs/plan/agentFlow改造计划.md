收到。根据你的最新调整，方案进行了核心优化：
1.  **架构极简**：云端和本地使用同一个 Rust 二进制程序（`agentflow`），通过启动参数区分模式（`--mode local` / `--mode cloud`）。
2.  **Worker 语言**：回归 **Node.js**，利用其生态简便性管理 CLI 进程。
3.  **安全机制**：基于 RSA 公钥 MD5 的节点认证，完全去中心化鉴权。
4.  **打包方案**：GitHub 仓库直接包含编译好的二进制文件，克隆即用。
这是更新后的完整开发计划书。
---
# 项目开发计划书：AgentFlow v2.0 (单二进制个人云)
## 1. 项目愿景与技术约束
**目标**：一个**克隆即用**的分布式个人云系统，通过单一 Rust 二进制文件管理云端和本地，Node.js 负责执行层。
**核心原则**：
- **单一二进制**：云端 Master 和本地 Master 是同一个 Rust 程序，减少维护成本。
- **零依赖部署**：仓库自带编译好的 `agentflow` 二进制文件，用户无需安装 Rust 工具链。
- **Node.js Worker**：利用 Node.js 强大的进程管理能力运行 Claude CLI。
- **去中心化鉴权**：各节点通过 `md5(public_key)` 互认，不依赖第三方 CA。
- **隐私隔离**：Master 不接触 `ANTHROPIC_API_KEY`，由 Worker 直接读取本地环境变量。
---
## 2. 项目结构与打包规范 (先决条件)
在编写代码前，必须先确立以下仓库结构，确保“克隆即用”。
```
agentflow/
├── bin/                    # [打包输出] 预编译二进制
│   ├── agentflow-linux-amd64
│   ├── agentflow-darwin-arm64
│   ├── agentflow-windows-amd64.exe
│   └── agentflow-linux-arm64
├── worker/                 # [Node.js Worker]
│   ├── package.json
│   ├── src/
│   │   └── index.js        # 主入口
│   └── lib/
│       └── claude.js       # Claude CLI 封装
├── scripts/                # [安装/配置脚本]
│   ├── setup.sh            # Linux/Mac 自动配置
│   └── setup.ps1           # Windows 自动配置
├── rust-core/              # [Rust 源码]
│   ├── Cargo.toml
│   └── src/
├── docker/                 # [可选] Docker 部署
│   └── Dockerfile
├── README.md               # 包含 "Quick Start: Clone & Run"
└── AGENTFLOW_PLAN.md       # 本计划书
```
**用户使用流程**：
1.  `git clone https://github.com/user/agentflow.git`
2.  `cd agentflow`
3.  `./scripts/setup.sh` (自动生成密钥对，选择模式)
4.  `./bin/agentflow --mode local` (启动 Master)
5.  `cd worker && npm install && npm start` (启动 Worker)
---
## 3. 第一阶段：安全协议与核心定义 (并行前置条件)
### Task 1.1: 密钥生成与认证协议 (安全基石)
**目标**: 实现基于公钥指纹的节点互认。
**技术细节**:
1.  **算法**: 使用 `Ed25519` (Rust `ed25519-dalek` 库)。
2.  **凭证生成**:
    - 首次启动时，`agentflow` 生成 `keypair.pk8` (私钥) 和 `keypair.pub` (公钥)。
    - 计算指纹: `fingerprint = md5(public_key_bytes)` (Hex String).
3.  **握手流程**:
    - Edge Master 连接 Cloud Master 时，发送 Hello 包: `{ node_id, fingerprint }`.
    - Cloud Master 检查本地 `trusted_nodes.db`:
        - 若指纹存在 -> 允许连接。
        - 若指纹不存在 -> 拒绝，并提示在云端控制台添加该节点指纹。
### Task 1.2: 通用二进制模式设计 (Rust 架构)
**目标**: 一个程序，两种模式。
**实现逻辑**:
- **CLI 参数**:
    - `agentflow --mode local`: 启动本地服务 (开启 Sandbox, Git管理, Worker调度).
    - `agentflow --mode cloud`: 启动云服务 (开启 WebSocket 网关, 路由转发, Zhipu Webhook).
- **Feature Flags**:
    - 利用 Rust `cfg` 属性或运行时 `Config` 结构体区分加载的模块。
    - Cloud 模式下禁用文件系统直接操作，仅保留网络转发能力。
### Task 1.3: 通信协议定义 (JSON over WebSocket)
**消息格式** (全链路通用):
```json
{
  "type": "auth|task|result|heartbeat|sync",
  "payload": { ... },
  "timestamp": 1710000000
}
```
---
## 4. 第二阶段：并行开发轨道 (同时执行)
### Track A: Rust Core (核心大脑)
**负责**: 实现 `agentflow` 二进制文件的核心逻辑。
**Task A.1: 基础设施与数据库**
- 使用 `sqlx` + `SQLite`。
- 实现表结构: `nodes`, `plans`, `tasks`, `artifacts` (Git Commits).
- 实现 `Config` 模块，读取 `agentflow.toml`。
**Task A.2: 通信层**
- 实现 WebSocket Server (模式: cloud) 和 WebSocket Client (模式: local)。
- 实现 `Authenticator`，校验 MD5(PubKey)。
**Task A.3: 调度器**
- **Local 模式**: 直接管理本地 Node.js Worker 进程 (或 Socket 连接)。
- **Cloud 模式**: 根据负载策略，将任务路由给指定的 Edge Node ID。
**Task A.4: Git 沙箱**
- 封装 `git2` 库。
- 实现 `safe_execute(command)`: `git checkout` -> `run` -> `git commit` -> `reset` (可选)。
- 确保所有命令运行在 `WORKSPACE_ROOT` 之下。
### Track B: Node.js Worker (手脚与感官)
**负责**: 运行 Claude CLI，处理敏感信息。
**Task B.1: 项目初始化**
- `npm init -y`.
- 依赖: `ws` (WebSocket客户端), `dotenv`, `shelljs`.
**Task B.2: 进程管理与安全**
- **环境隔离**: Worker 启动时，自动加载 `.env` 中的 `ANTHROPIC_API_KEY`。**严禁**将该 Key 打印到 Stdout 或发给 Master。
- **执行器**: 使用 `child_process.spawn` 执行 `claude` 命令。
- **输出流处理**: 实时捕获 `claude` 的 Stdout/Stderr，并通过 WebSocket 发送给 Master。
**Task B.3: Master 通信模块**
- 连接本地 `agentflow` (通常是 `ws://localhost:8080`)。
- 发送 `Register` 消息。
- 接收 `Task` 消息，解析命令，执行。
### Track C: 部署与打包工程 (交付物流)
**负责**: 确保用户拿到手就能跑。
**Task C.1: 交叉编译脚本**
- 编写 GitHub Actions (`.github/workflows/release.yml`)。
- 触发条件: Tag Push.
- 任务:
    - 编译 Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64).
    - 上传二进制文件到 `bin/` 目录。
**Task C.2: 自动化安装脚本**
- `setup.sh`:
    - 检测 OS 架构。
    - 下载对应的 `bin/agentflow`。
    - 运行 `agentflow --gen-key` (生成密钥对)。
    - 输出指纹，提示用户在云端添加。
---
## 5. 核心安全与交互细节 (开发注意事项)
### 5.1 Node.js Worker 的安全红线
- **规则**: 绝不通过 Master 转发 API Key。
- **实现**: Worker 内部逻辑 `process.env.ANTHROPIC_API_KEY`。Master 只下发“思考任务”，Worker 自行“消化”。
### 5.2 Master 间通信安全
- **规则**: Cloud Master 只认指纹，不认 IP。
- **实现**: 新的 Edge Master 连接时，Cloud Master 在日志中打印:
  > `Unknown node connected. Fingerprint: a1b2c3d4... Please add to trusted list.`
- 用户操作: 登录云端 Web 界面，输入指纹，点击 "Trust"。
### 5.3 记忆同步
- **Local Master**: 维护详细的 SQLite 记忆库。
- **Cloud Master**: 维护全局索引。
- **同步**: Local Master 定期将记忆的 `Vector Hash` 和 `NodeID` 推送给 Cloud Master。Cloud Master 仅存储索引，不存储内容。
---
## 6. 交付与验收标准
1.  **打包验收**:
    - 在一台全新的 Linux 服务器上执行 `git clone && ./scripts/setup.sh`。
    - 能成功启动 `agentflow --mode cloud`。
    - 在一台 Windows 机器上执行同样步骤，能启动 `agentflow --mode local`。
    - 两者能通过指纹认证建立连接。
2.  **功能验收**:
    - 用户配置好 Claude Key 后。
    - 通过 Cloud Master API 下发任务。
    - Windows 上的 Node.js Worker 能拉取代码并在沙箱中执行 Claude。
    - 执行结果（日志、Git Commit）能回传到 Cloud Master。
3.  **代码规范**:
    - Rust: `cargo fmt`, `cargo clippy`。
    - Node.js: `eslint` (standard)。
---
**指令提示**:
Agent 请按照以下顺序执行：
1.  **初始化仓库结构**: 创建目录，生成 `package.json` 和 `Cargo.toml`。
2.  **Track C (打包) 先行**: 即使代码还没写，先写好 GitHub Actions 的编译配置，确保我们能产出二进制文件。
3.  **Track A & B 并行**: 开发 Rust 核心和 Node.js Worker。
4.  **集成测试**: 模拟 Cloud 和 Local 的通信。
**开始！**
