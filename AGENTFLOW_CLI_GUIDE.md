# AgentFlow 简化命令使用指南

## 快速开始

### 1. 安装 CLI 工具（可选）

```bash
# 方式1: 创建全局命令（推荐）
sudo cp /tmp/agentflow-cli.sh /usr/local/bin/agentflow
sudo chmod +x /usr/local/bin/agentflow

# 方式2: 使用别名
echo 'alias agentflow="export PATH=\"/opt/homebrew/opt/node@20/bin:$PATH\" && node /Users/jiangxiaolong/work/project/AgentFlow/nodejs/packages/cli/dist/index.js"' >> ~/.zshrc
source ~/.zshrc

# 方式3: 直接使用完整路径
node /Users/jiangxiaolong/work/project/AgentFlow/nodejs/packages/cli/dist/index.js
```

### 2. 基本使用

#### ✅ 最简单的执行命令

```bash
# 执行单个命令
agentflow run "echo Hello World"

# 查看帮助
agentflow run --help
```

#### ✅ 指定任务标题

```bash
agentflow run "ls -la" --title "列出文件"
```

#### ✅ 使用自定义数据库

```bash
agentflow run "npm test" --db ./my-tasks.db
```

#### ✅ 保持 Master 和 Worker 运行

```bash
# 不自动关闭，可以持续执行任务
agentflow run "echo test" --no-shutdown
```

## 命令对比

### ❌ 之前：需要手动管理多个组件

```bash
# 终端1: 启动 Master
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
node nodejs/packages/master/dist/index.js --port 6767 --db data/agentflow.db

# 终端2: 启动 Worker
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
node nodejs/packages/worker/dist/index.js

# 终端3: 创建任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"测试","description":"echo hello"}'
```

### ✅ 现在：一行命令搞定

```bash
agentflow run "echo hello"
```

## 命令说明

### agentflow run

**自动执行任务（推荐）**

```bash
agentflow run [command] [options]

参数:
  command           要执行的命令（必填）

选项:
  --title <title>   任务标题
  --master-host     Master 主机地址（默认：localhost）
  --master-port     Master 端口（默认：6767）
  --db <path>       数据库路径
  --no-shutdown     完成后不关闭 Master 和 Worker
  --group <group>   Worker 组名（默认：cli）
```

### agentflow master

**启动 Master 服务器**

```bash
agentflow master [options]

选项:
  --host <host>    绑定地址（默认：0.0.0.0）
  --port <port>    端口（默认：6767）
  --db <path>      数据库路径
  --auto-shutdown  启用自动关闭
  --ws-enabled     启用 WebSocket（默认：true）
  --ws-port <port> WebSocket 端口（默认：8849）
```

### agentflow worker

**启动 Worker**

```bash
agentflow worker [options]

选项:
  --master <url>   Master URL（默认：http://localhost:6767）
  --id <id>        Worker ID
  --group <group>  组名（默认：default）
  --mode <mode>    模式：auto|manual|oneshot（默认：auto）
```

## 使用示例

### 示例 1: 快速测试

```bash
# 简单命令
agentflow run "echo 'Hello AgentFlow'"

# 输出：
# 🚀 Executing: echo 'Hello AgentFlow'
# ✓ Worker registered
# ✓ Task 1 completed
# ✅ Execution complete!
```

### 示例 2: 代码分析任务

```bash
# 复杂任务会自动调用 Claude CLI
agentflow run "分析以下代码的性能：
function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n-1) + fibonacci(n-2);
}" --title "性能分析"
```

### 示例 3: 批量执行

```bash
# 先启动服务（不自动关闭）
agentflow run "echo first" --no-shutdown

# 然后可以快速执行多个任务
agentflow run "npm test"
agentflow run "npm run build"
agentflow run "echo done"
```

### 示例 4: 自定义数据库

```bash
# 使用持久化数据库
agentflow run "echo test" --db ~/my-project/tasks.db
```

## 高级用法

### 1. 在脚本中使用

```bash
#!/bin/bash
# deploy.sh

echo "🚀 开始部署..."

# 运行测试
agentflow run "npm test" --title "运行测试"

# 构建
agentflow run "npm run build" --title "构建项目"

# 部署
agentflow run "./deploy.sh" --title "部署到生产"

echo "✅ 部署完成！"
```

### 2. CI/CD 集成

```yaml
# .github/workflows/test.yml
name: Test

on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Setup Node.js
        uses: actions/setup-node@v2
        with:
          node-version: '20'

      - name: Run tests with AgentFlow
        run: |
          agentflow run "npm test" --title "单元测试"
          agentflow run "npm run integration:test" --title "集成测试"
```

### 3. 定时任务

```bash
# crontab -e
# 每小时运行健康检查
0 * * * * agentflow run "curl -f http://localhost:3000/health || echo 'Health check failed'" --title "健康检查"
```

## 原理说明

`agentflow run` 命令实际上封装了以下操作：

1. ✅ 自动启动 Master 服务器
2. ✅ 自动启动 Worker
3. ✅ 创建任务
4. ✅ Worker 执行任务
5. ✅ 等待完成
6. ✅ 返回结果
7. ✅ 自动清理（可选）

这样你只需要关注**要执行什么**，而不用关心**如何执行**。

## 故障排除

### 问题1: 命令找不到

```bash
# 检查别名是否设置
alias agentflow

# 如果没有，重新添加
echo 'alias agentflow="node /Users/jiangxiaolong/work/project/AgentFlow/nodejs/packages/cli/dist/index.js"' >> ~/.zshrc
source ~/.zshrc
```

### 问题2: Node.js 版本错误

```bash
# 确保 Node.js 20
export PATH="/opt/homebrew/opt/node@20/bin:$PATH"
node --version  # 应该显示 v20.x.x
```

### 问题3: Claude CLI 未找到

```bash
# 安装 Claude CLI
npm install -g @anthropic-ai/claude-code

# 验证安装
claude --version
```

## 性能对比

| 方式 | 命令数 | 时间 | 复杂度 |
|------|--------|------|--------|
| **传统方式** | 3+ | ~30秒 | 高（需要3个终端） |
| **CLI方式** | 1 | ~5秒 | 低（一个命令） |

## 总结

**推荐使用场景：**

- ✅ 快速测试单个命令
- ✅ 执行脚本中的任务
- ✅ CI/CD 集成
- ✅ 定时任务
- ✅ 开发调试

**不推荐使用场景：**

- ❌ 长期运行的服务（使用 `agentflow master` + `agentflow worker`）
- ❌ 多用户协作（需要独立的 Master/Worker）
- ❌ 高并发场景（使用 Go 版本）

现在你只需要一行命令就能执行任何任务！🎉
