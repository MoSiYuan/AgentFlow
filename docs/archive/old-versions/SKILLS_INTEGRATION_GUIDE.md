# AgentFlow Skills Integration Guide

## 概述

AgentFlow Worker 现在支持自动发现和使用宿主机上的所有 Claude Skills，使 Claude CLI 在执行任务时能够直接使用这些 skills。

## 功能特性

✅ **自动发现**: 自动扫描以下 Skills 目录：
- `~/.claude/skills/` - 默认 Claude skills 目录
- `.claude/skills/` - 项目特定 skills
- `/usr/local/share/claude/skills/` - 全局 skills
- 通过环境变量指定的自定义路径

✅ **环境变量配置**: 支持通过环境变量配置：
- `AGENTFLOW_SKILLS_ENABLED` - 启用 skills 集成（默认: true）
- `AGENTFLOW_SKILLS_PATHS` - 额外的 skills 目录（用 `:` 分隔）
- `AGENTFLOW_SKILLS_AUTO_DISCOVER` - 自动发现（默认: true）

✅ **能力报告**: Worker 注册时会报告可用的 skills 数量

✅ **无缝集成**: Claude CLI 通过环境变量 `CLAUDE_SKILLS_PATH` 接收 skills 路径

## 使用方法

### 方法 1: 使用默认配置

```bash
# 直接启动 Worker（自动发现默认 skills 目录）
node nodejs/packages/worker/dist/index.js
```

Worker 会自动发现并使用：
- `~/.claude/skills/`
- `.claude/skills/` (项目目录下)
- `/usr/local/share/claude/skills/`

### 方法 2: 指定自定义 Skills 目录

```bash
# 设置自定义 skills 路径
export AGENTFLOW_SKILLS_PATHS="/opt/skills:/usr/local/share/my-skills:~/custom-skills"

# 启动 Worker
node nodejs/packages/worker/dist/index.js
```

### 方法 3: 禁用 Skills 集成

```bash
# 禁用 skills 集成
export AGENTFLOW_SKILLS_ENABLED=false

# 启动 Worker
node nodejs/packages/worker/dist/index.js
```

## 创建自定义 Skill

### Skill 目录结构

```
.claude/skills/
└── my-skill/
    └── SKILL.md
```

### SKILL.md 示例

```markdown
---
name: my-skill
description: "Description of what this skill does"
---

# My Skill

Use this skill to...

## Usage

\`\`\`bash
command-to-run
\`\`\`

## Examples

Example 1:
\`\`\`
your code here
\`\`\`
```

## 验证 Skills 集成

### 1. 运行测试脚本

```bash
node test-skills-simple.js
```

### 2. 检查 Worker 日志

启动 Worker 时会输出：
```
✓ Worker with 5 skills in 2 directories
```

### 3. 查看 Worker 能力

Worker 注册时的 capabilities 会包含：
```
shell, typescript, javascript, claude-cli, claude-skills, skills:5
```

### 4. 创建测试任务

```bash
# 启动 Master
./bin/master --port 6767

# 在另一个终端启动 Worker
AGENTFLOW_SKILLS_ENABLED=true node nodejs/packages/worker/dist/index.js

# 创建一个使用 Claude CLI 的任务
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Skills",
    "description": "Use the /github skill to list recent commits",
    "group_name": "default"
  }'
```

## 环境变量参考

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `AGENTFLOW_SKILLS_ENABLED` | boolean | true | 是否启用 skills 集成 |
| `AGENTFLOW_SKILLS_PATHS` | string | - | 额外的 skills 目录（用 `:` 分隔） |
| `AGENTFLOW_SKILLS_AUTO_DISCOVER` | boolean | true | 是否自动发现 skills 目录 |

## 配置文件方式

除了环境变量，也可以在代码中配置：

```typescript
const { Worker } = require('@agentflow/worker');

const worker = new Worker({
  master_url: 'http://localhost:6767',
  group_name: 'default',
  skills_enabled: true,
  skills_auto_discover: true,
  skills_paths: [
    '/opt/skills',
    '~/custom-skills'
  ]
});

worker.start();
```

## 常见问题

### Q: Claude CLI 没有加载我的 skills？

A: 确保：
1. Skills 目录下有 `SKILL.md` 文件（注意大小写）
2. Skills 目录路径正确
3. Worker 启动时显示了 "Found X skills in Y directories"

### Q: 如何查看发现了哪些 skills？

A: 查看 Worker 启动日志，或运行测试脚本：
```bash
node test-skills-simple.js
```

### Q: 可以在 Docker 中使用吗？

A: 可以！需要挂载 skills 目录到容器中：
```yaml
version: '3.8'
services:
  agentflow-worker:
    image: agentflow-worker:latest
    environment:
      - AGENTFLOW_SKILLS_ENABLED=true
      - AGENTFLOW_SKILLS_PATHS=/app/skills:/custom/skills
    volumes:
      - ~/.claude/skills:/app/skills:ro
      - ./custom-skills:/custom/skills:ro
```

### Q: Skills 优先级是什么？

A: Claude CLI 会按照以下优先级加载 skills：
1. 环境变量 `CLAUDE_SKILLS_PATH` 中的目录（按顺序）
2. 默认的 `~/.claude/skills/`

## 下一步

- 查看 [Claude Skills 文档](https://code.claude.com/docs/en/skills) 了解如何创建 skills
- 查看 [ClawdBot Skills 示例](https://github.com/clawdbot/clawdbot/tree/master/skills) 获取灵感
- 集成你的自定义 skills 到 AgentFlow

## 相关文件

- Worker 实现: `nodejs/packages/worker/src/index.ts`
- 类型定义: `nodejs/packages/shared/src/types.ts`
- 测试脚本: `test-skills-simple.js`
