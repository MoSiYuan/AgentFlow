# AgentFlow for AI Agents

> Claude Parallel Development System - 专为 AI Agent 设计的分布式任务执行系统

## 为什么选择 AgentFlow？

### 对 AI Agent 友好

- **极简 API**：3 个命令即可开始（init, add, list）
- **智能任务分解**：自动将大任务分解为子任务
- **边界安全**：沙箱执行、路径限制、只读模式
- **实时反馈**：任务状态实时更新

### 强大的功能

- ✅ **任务层次结构**：主任务自动分解为子任务
- ✅ **Worker 组**：按环境/能力分组（Linux, Windows, Docker, K8s）
- ✅ **双重模式**：本地（直连 DB）和远程（HTTP API）
- ✅ **安全控制**：沙箱、资源限制、审计日志
- ✅ **自动重试**：任务失败自动重试
- ✅ **并发执行**：多 Worker 并行处理

## 30 秒上手

```bash
# 启动
make build && ./bin/agentflow init agentflow.db && ./bin/agentflow master

# 创建任务
./bin/agentflow add "实现登录功能" --desc "task:implement:login" --group local

# 等待完成
./bin/agentflow list
```

## AI 任务类型

### 1. 智能任务（推荐）

```bash
# AI 自动分解任务
cpds add "实现用户注册" --desc "task:implement:registration"
```

自动分解为：
- 设计数据模型
- 实现 API 端点
- 编写测试
- 更新文档

### 2. Shell 命令

```bash
cpds add "运行测试" --desc "shell:go test ./..." --group linux
```

### 3. 脚本执行

```bash
cpds add "部署应用" --desc "script:./deploy.sh" --group production
```

### 4. 文件操作

```bash
cpds add "更新配置" --desc "file:write:config.json:{\"key\":\"value\"}"
```

## 部署模式

### 本地模式（推荐给 AI）

```bash
# 一键启动
./bin/agentflow master --db agentflow.db

# Worker 自动启动，直接访问 SQLite
# 无网络开销，性能最佳
```

### 云端模式（分布式）

```bash
# 服务器
./bin/agentflow master --db /data/agentflow.db --host 0.0.0.0

# 本地 AI
export MASTER_URL=http://server.com:8848
./bin/agentflow worker
```

## 安全特性

### 沙箱执行

```bash
export SANDBOXED=true
./bin/agentflow worker
# 所有命令在 Docker 容器中执行
```

### 路径限制

```bash
export WORKSPACE_DIR=/tmp/workspace
export RESTRICT_PATH=true
./bin/agentflow worker
# 只能操作指定目录
```

### 只读模式

```bash
export READ_ONLY=true
./bin/agentflow worker
# 只能执行查询，不能修改
```

### 资源限制

```bash
export MAX_MEMORY=512M
export MAX_CPU=1
./bin/agentflow worker
# 限制 CPU 和内存使用
```

## Claude Code 集成示例

```typescript
// Claude Code Skill
async function developFeature(feature: string) {
  // 创建任务
  await exec('cpds', [
    'add',
    `实现${feature}`,
    '--desc', `task:implement:${feature}`,
    '--group', 'local'
  ]);

  // 轮询状态
  while (true) {
    const { stdout } = await exec('cpds', ['list', '--status', 'running']);
    if (!stdout.includes(`实现${feature}`)) {
      break;
    }
    await sleep(2000);
  }

  // 获取结果
  const { stdout } = await exec('cpds', ['list', '--status', 'completed']);
  console.log(stdout);
}
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MASTER_URL` | Master 地址 | http://localhost:8848 |
| `WORKER_GROUP` | Worker 组 | 自动检测 |
| `WORKSPACE_DIR` | 工作目录 | /tmp/cpds-workspace |
| `SANDBOXED` | 沙箱模式 | false |
| `RESTRICT_PATH` | 路径限制 | false |
| `READ_ONLY` | 只读模式 | false |
| `MAX_MEMORY` | 最大内存 | 512M |
| `MAX_CPU` | 最大 CPU | 1 |

## 文档

- [快速开始](AI_QUICKSTART.md) - 3 分钟上手
- [部署指南](AI_DEPLOYMENT.md) - 详细部署说明
- [API 文档](API.md) - 完整 API 参考
- [架构设计](ARCHITECTURE.md) - 系统架构

## 贡献

欢迎贡献！请查看 [贡献指南](CONTRIBUTING.md)

## 许可证

[MIT License](LICENSE)

---

**Made with ❤️ for AI Agents**
