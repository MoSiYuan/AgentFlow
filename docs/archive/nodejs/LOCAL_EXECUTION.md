# 本地执行指南

AgentFlow 支持一键自动执行：自动启动服务、执行任务、完成后自动关闭。

## 快速开始

### 最简单的方式

```bash
# 构建项目
cd nodejs && npm run build && cd ..

# 执行任务
agentflow run "npm test" "npm run build"
```

### 或使用独立脚本

```bash
node local-executor.js "npm test" "npm run build"
```

## 工作流程

```
1. 启动服务   → 自动启动 Master
2. 注册节点   → 自动连接 Worker
3. 创建任务   → 提交你的任务
4. 执行监控   → 实时显示进度
5. 自动关闭   → 完成后清理
```

## 使用示例

### 测试和构建

```bash
agentflow run "npm test" "npm run build"
```

输出：
```
AgentFlow 本地执行

  正在启动服务...
  ✓ 服务已就绪
  正在注册工作节点...
  ✓ 工作节点已就绪
  正在创建任务...
  ✓ npm test
  ✓ npm run build

  正在创建任务...
  共创建 2 个任务
  正在执行任务...
  进度: 2/2 完成, 0 失败
  ✓ 所有任务已完成
  正在关闭服务...
  ✓ 执行完成
```

### 并行测试

```bash
agentflow run \
  "cd frontend && npm test" \
  "cd backend && npm test" \
  "cd api && npm test"
```

### 开发工作流

```bash
agentflow run \
  "npm run format" \
  "npm run lint" \
  "npm test" \
  "npm run build"
```

## 配置文件

在项目根目录创建 `.agentflow/config.json`：

```json
{
  "master": {
    "path": "nodejs/packages/master/dist/index.js"
  },
  "worker": {
    "group": "local"
  }
}
```

## 特性

- ✅ **零配置** - 默认配置即可使用
- ✅ **自动管理** - 启动、执行、关闭全自动
- ✅ **实时进度** - 显示任务执行状态
- ✅ **并行执行** - 多个任务同时运行
- ✅ **自动清理** - 完成后自动关闭服务

## 故障排除

### Master 启动失败

```bash
# 检查是否已构建
ls -l nodejs/packages/master/dist/index.js

# 重新构建
cd nodejs && npm run build
```

### 端口被占用

```bash
# 查找占用进程
lsof -i :6767

# 终止进程
kill -9 <PID>
```

### 任务执行失败

查看 Worker 日志排查问题。

## 相关文档

- [架构说明](docs/ARCHITECTURE.md)
- [CLI 使用说明](docs/SKILL.md)
- [部署指南](deployment/)

---

**版本**: 2.0.0 | **更新日期**: 2026-01-23
