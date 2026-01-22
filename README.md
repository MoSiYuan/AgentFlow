# AgentFlow - AI Agent 任务协作系统

> **30秒上手，3令牌完成任务** - 专为 Claude Code 设计

## 🎯 核心功能

- **任务分发**: Master 分配任务给 Worker
- **任务升级**: Worker 创建子任务（如：生成故事→评审）
- **Worker 协作**: 多 Worker 并发执行，互评结果
- **边界安全**: 工作目录隔离，可沙箱执行

## ⚡ 3秒上手

```bash
# 运行测试演示（10个克苏鲁神话故事+评审，3秒完成）
cd tests && go run ctest_pure.go

# 查看结果
cat ctest_stories/story_1.md
```

## 📝 任务格式

```bash
# 故事生成
agentflow add "标题" --desc "write_story:标题:type:001"

# 评审
agentflow add "评审标题" --desc "review_story:任务ID:标题:review"

# AI 任务（自动分解）
agentflow add "开发功能" --desc "task:implement:功能名"
```

## 🔧 常用命令

```bash
agentflow init <db>           # 初始化
agentflow master --db <db>       # 启动 Master
agentflow add "标题" --desc "..."  # 创建任务
agentflow list [--status ...]    # 查看任务
agentflow workers                 # 查看 Worker
```

## 📖 文档

- [测试指南](tests/README.md) - 快速测试
- [AI 部署](docs/AI_DEPLOYMENT.md) - 云端部署
- [AI 快速开始](docs/AI_QUICKSTART.md) - 3分钟教程

## 🧪 实战示例

**已验证**: 10个故事生成+20个评审=100%成功

- 总任务: 30个
- 执行时间: 3秒
- 输出: 10个Markdown文件
- 位置: `tests/ctest_stories/`

## 🚀 为 AI 优化

### 节约 Token 技巧

1. **短命令**: `af add "T" --d "s:T:t:1"` (16 token)
2. **批量**: 一次创建多个任务
3. **过滤**: `af list --s completed` 只看结果

### 快速集成

```go
// 1行创建任务
exec("agentflow add T --desc s:T:t:1")

// 1行查询状态
exec("agentflow list --s completed")
```

## 📁 项目结构

```
agentflow-go/
├── cmd/agentflow/          # CLI工具
├── internal/
│   ├── database/     # SQLite层
│   ├── master/       # Master服务
│   └── worker/       # Worker+AI Worker
├── tests/
│   ├── ctest_pure.go # 测试代码（可运行）
│   └── ctest_stories/ # 测试结果（示例）
└── docs/             # 完整文档
```

## 💡 使用场景

1. **本地开发**: Master自动启动本地Workers，直连DB
2. **云端部署**: Master在服务器，Workers分布式连接
3. **任务协作**: 主任务完成后创建子任务
4. **内容生成**: 文档、代码、测试、评审

## 🔗 关键文件

- [AI Worker 实现](internal/worker/ai_worker.go)
- [Python SDK](sdk/python/agentflow_ai.py)
- [TypeScript SDK](sdk/typescript/agentflow_ai.ts)

---

**项目**: [jiangxiaolong/agentflow-go](https://github.com/jiangxiaolong/agentflow-go)
**许可证**: MIT
**测试状态**: ✅ 30任务/100%成功（3秒）
