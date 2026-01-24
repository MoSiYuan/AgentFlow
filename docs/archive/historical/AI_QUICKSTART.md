# AgentFlow AI Agent 快速使用指南

> 3 分钟上手 AgentFlow（面向 Claude Code）

## 最简使用（本地模式）

```bash
# 1. 一键启动
git clone https://github.com/jiangxiaolong/agentflow-go.git
cd agentflow-go
make build && ./bin/cpds init agentflow.db && ./bin/cpds master --db agentflow.db

# 2. Claude Code 创建任务
cpds add "实现用户登录" --desc "task:implement:login" --group local

# 3. 查看结果
cpds list --status completed
```

## 任务类型速查

| 类型 | 格式 | 说明 |
|------|------|------|
| AI 任务 | `task:implement:feature` | AI 自动分解并执行 |
| Shell | `shell:go test ./...` | 执行命令 |
| 脚本 | `script:./deploy.sh` | 运行脚本 |
| 文件 | `file:write:README.md:content` | 文件操作 |

## 常用命令

```bash
# 创建任务
cpds add "标题" --desc "任务描述" --group local

# 查看所有任务
cpds list

# 查看特定状态
cpds list --status running
cpds list --status pending

# 查看 Worker
cpds workers
```

## 云端部署（分布式）

```bash
# 服务器端
./bin/cpds master --db /data/agentflow.db --host 0.0.0.0

# 本地 Worker
export MASTER_URL=http://your-server.com:6767
./bin/cpds worker
```

## 安全模式

```bash
# 沙箱模式
export SANDBOXED=true
./bin/cpds worker

# 只读模式
export READ_ONLY=true
./bin/cpds worker

# 限制工作目录
export WORKSPACE_DIR=/tmp/workspace
export RESTRICT_PATH=true
./bin/cpds worker
```

## Claude Code 集成

```python
# Claude Code 可以直接调用
import subprocess

def create_and_wait(title, description):
    # 创建任务
    subprocess.run(['cpds', 'add', title,
                   '--desc', description,
                   '--group', 'local'])

    # 等待完成
    while True:
        result = subprocess.run(['cpds', 'list', '--status', 'running'],
                              capture_output=True)
        if title not in result.stdout.decode():
            break
        time.sleep(2)

    # 获取结果
    result = subprocess.run(['cpds', 'list', '--status', 'completed'],
                          capture_output=True)
    print(result.stdout.decode())
```

## 更多信息

- [完整部署指南](AI_DEPLOYMENT.md)
- [API 文档](API.md)
- [架构设计](ARCHITECTURE.md)
