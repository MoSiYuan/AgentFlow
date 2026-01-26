# AgentFlow Skill 安装指南

**让 Claude AI 帮你管理任务**

---

## 🎯 这是什么？

AgentFlow 是一个 AI Agent 任务协作系统，可以与 Claude CLI 深度集成，帮你管理复杂的任务链、Git 分支、状态同步等。

---

## ⚡ 一键安装（最简单）

### 告诉 Claude AI：

```
请帮我安装 AgentFlow：
1. 在 ~/.claude/skills/ 目录下创建 agentflow 文件夹
2. 下载 https://raw.githubusercontent.com/jiangxiaolong/AgentFlow/main/.claude/skills/agentflow.md
3. 放到 ~/.claude/skills/agentflow/ 目录
4. 现在我可以使用 /agentflow 命令了
```

### 或者手动安装：

```bash
# 创建目录
mkdir -p ~/.claude/skills/agentflow

# 下载 skill 文件
curl -fsSL https://raw.githubusercontent.com/jiangxiaolong/AgentFlow/main/.claude/skills/agentflow.md \
  -o ~/.claude/skills/agentflow.md

# 完成！现在可以在 Claude 中使用 /agentflow 命令
```

---

## 🚀 使用方法

### 在 Claude CLI 中：

```
/agentflow 创建任务 "测试任务" "echo Hello World"

/agentflow 查看所有任务

/agentflow 创建任务链 测试 构建 部署

/agentflow 同步状态
```

### 或使用 API：

```bash
# 启动 AgentFlow 服务
agentflow-master

# 创建任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"测试任务","description":"echo Hello"}'
```

---

## 📦 系统要求

### 方式 1: 独立可执行文件（推荐）

```bash
# 下载
curl -fsSL https://github.com/jiangxiaolong/AgentFlow/releases/download/v1.0.0/agentflow-master \
  -o agentflow-master
chmod +x agentflow-master

# 运行（无需 Node.js）
./agentflow-master
```

### 方式 2: Docker

```bash
docker run -d \
  --name agentflow \
  -p 6767:6767 \
  -v ~/.claude:/root/.claude \
  jiangxiaolong/agentflow:latest
```

### 方式 3: 从源码

```bash
git clone https://github.com/jiangxiaolong/AgentFlow.git
cd AgentFlow/nodejs
./package.sh
./dist/agentflow-master
```

---

## ✅ 验证安装

```bash
# 检查服务状态
curl http://localhost:6767/health

# 在 Claude CLI 中测试
claude "使用 /agentflow 创建一个测试任务"
```

---

## 🎓 示例

### 创建简单任务

```
/agentflow 创建任务 "运行测试" "npm test"
```

### 创建任务链

```
/agentflow 创建串行任务链:
  - 任务1: 运行测试 "npm test"
  - 任务2: 构建 "npm run build"
  - 任务3: 部署 "npm run deploy"
```

### Git 集成

```
/agentflow 创建 Git 任务:
  - 分支: feature/new-feature
  - 描述: 实现新功能
  - 测试: npm test
```

---

## 🔧 配置

### 默认配置

- **端口**: 6767
- **数据库**: `~/.claude/skills/agentflow/agentflow.db`

### 自定义配置

```bash
# 环境变量
export AGENTFLOW_PORT=8080
export AGENTFLOW_DB_PATH="/custom/path/agentflow.db"
```

---

## 📚 更多信息

- 完整文档: https://github.com/jiangxiaolong/AgentFlow
- API 文档: https://github.com/jiangxiaolong/AgentFlow/wiki/API
- 视频教程: https://github.com/jiangxiaolong/AgentFlow/wiki/Tutorials

---

## 🆘 遇到问题？

### 服务无法启动

```bash
# 检查端口是否被占用
lsof -i :6767

# 使用其他端口
agentflow-master --port 8080
```

### 数据库错误

```bash
# 创建目录
mkdir -p ~/.claude/skills/agentflow

# 检查权限
ls -la ~/.claude/skills/agentflow/
```

### 更多帮助

- GitHub Issues: https://github.com/jiangxiaolong/AgentFlow/issues
- 文档: https://github.com/jiangxiaolong/AgentFlow/wiki

---

**🎉 安装完成！现在可以在 Claude 中使用 /agentflow 命令了！**
