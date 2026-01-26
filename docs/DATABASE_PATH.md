# AgentFlow 数据库路径配置

## 默认路径

AgentFlow 数据库文件默认位于：

```
~/.claude/skills/agentflow/agentflow.db
```

这个位置是 AgentFlow skill 的工作目录。

## Skill 部署结构

```
~/.claude/skills/agentflow/
├── agentflow.db           # AgentFlow 数据库 ⭐
├── bin/                   # 可执行文件
│   ├── master            # Master 服务器
│   └── worker            # Worker 进程
├── scripts/               # 工具脚本
├── config/                # 配置文件
└── README.md              # 说明文档
```

## 快速开始

### 使用默认配置

```bash
# Go 版本
./bin/master
./bin/worker

# Node.js 版本
cd nodejs/packages/master
npm start
```

### 自定义路径

```bash
# 通过环境变量
export MASTER_DB_PATH="/custom/path/agentflow.db"

# 通过命令行参数
./bin/master --db /custom/path/agentflow.db

# 通过配置文件
cat > ~/.claude/agentflow.yaml <<EOF
master:
  db_path: "~/.claude/skills/agentflow/agentflow.db"
  port: 8848
EOF
./bin/master --config ~/.claude/agentflow.yaml
```

## 验证数据库位置

```bash
# 检查数据库文件
ls -lh ~/.claude/skills/agentflow/agentflow.db

# 查看 skill 目录结构
ls -la ~/.claude/skills/agentflow/

# 查看数据库内容
sqlite3 ~/.claude/skills/agentflow/agentflow.db "SELECT COUNT(*) FROM tasks;"
```

## 相关文档

- [详细配置说明](./DATABASE_LOCATION.md)
- [架构设计](../ARCHITECTURE.md)
- [API 文档](./API.md)
