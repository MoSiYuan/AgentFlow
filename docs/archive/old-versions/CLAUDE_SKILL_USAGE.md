# AgentFlow + Claude CLI 实战指南

## 核心理念

**让 Claude 主动调用 AgentFlow**，而不是 AgentFlow 被动调用 Claude。

### 为什么？

1. **Claude CLI 的优势**：理解意图、规划任务、分析结果
2. **AgentFlow 的优势**：执行、调度、持久化、分布式
3. **完美互补**：Claude = 大脑，AgentFlow = 执行者

## 快速开始

### 1. 安装 AgentFlow Skill

Skill 文件已创建：`.claude/skills/agentflow.md`

启动 Claude CLI 时会自动加载此 skill。

### 2. 验证 Skill 加载

```bash
$ claude

Claude: Hi! I'm ready to help. I have access to the following skills:
  - agentflow: Execute and manage distributed tasks

User: /agentflow help
Claude: [显示 AgentFlow 使用帮助]
```

## 实际使用场景

### 场景 1：CI/CD 流水线

**对话示例**：

```bash
You: 我需要运行完整的 CI 流程：lint、test、build，然后部署到 staging

Claude: 我将使用 AgentFlow 来管理这些任务。

[AgentFlow 正在执行...]

✅ 任务创建成功！

**任务队列**:
1. ⏳ Lint (pending) - priority: high
2. ⏳ Test (pending) - priority: high
3. ⏳ Build (pending) - priority: medium (depends on: 1,2)
4. ⏳ Deploy Staging (pending) - priority: medium (depends on: 3)

**执行策略**:
- Lint 和 Test 并行执行
- Build 在两者通过后执行
- Deploy 在 Build 成功后执行

[等待任务完成...]

✅ CI 流程完成！

**结果**:
1. ✅ Lint - 通过 (2.3s)
   - 0 errors, 3 warnings

2. ✅ Test - 通过 (15.7s)
   - 128 tests passed
   - Coverage: 87%

3. ✅ Build - 成功 (23.1s)
   - Output: dist/app.js (1.2MB)

4. ✅ Deploy Staging - 完成 (45.2s)
   - URL: https://staging.example.com
   - Commit: abc123def

**查看详情**: agentflow status TASK-00000004
```

**发生了什么**：

1. Claude 理解你的需求
2. 调用 AgentFlow skill 创建任务
3. 自动设置依赖关系和优先级
4. 监控执行状态
5. 汇总结果并提供链接

### 场景 2：批量数据处理

**对话示例**：

```bash
You: 处理 data/ 目录下的所有 CSV 文件，每个文件跑一个分析任务，总共 100 个文件

Claude: 我将为每个文件创建一个任务并使用 AgentFlow 并行执行。

[扫描文件...]
发现 100 个 CSV 文件

[AgentFlow 正在创建任务...]

✅ 已创建 100 个任务

**任务分布**:
- Group: batch-processing
- Priority: medium
- Max concurrent: 5 (每个 Worker)

**Worker 配置**:
- Worker 1: localhost - 20 tasks
- Worker 2: localhost - 20 tasks
- Worker 3: remote-1 - 30 tasks
- Worker 4: remote-2 - 30 tasks

[监控进度...]

📊 实时进度:

████████████████████░░░░░░░░ 80/100 (80%)

✅ 已完成: 75
⏳ 运行中: 5
⏳ 待处理: 15
❌ 失败: 5

**失败任务**:
- data/corrupt_1.csv - 文件损坏
- data/corrupt_2.csv - 文件损坏
- data/incomplete_3.csv - 数据不完整
- data/empty_4.csv - 空文件
- data/invalid_5.csv - 格式错误

**建议**:
1. 修复或移除损坏的文件
2. 重新运行失败的任务
3. 对其余文件进行验证

You: 好的，重新运行那 5 个失败的任务

Claude: [使用 AgentFlow 重新执行失败任务...]

✅ 重新执行完成

**结果**:
- 3 个任务成功（修复后）
- 2 个任务失败（文件确实损坏）

**最终统计**:
- 成功: 78/100 (78%)
- 失败: 2/100 (2%)
- 跳过: 20/100 (20% - 空文件)
```

### 场景 3：多环境部署

**对话示例**：

```bash
You: 部署新版本到所有环境：dev → staging → production

Claude: 我将按顺序部署到各个环境，每个环境部署前运行验证测试。

[AgentFlow 创建部署任务...]

**部署计划**:

1. ✅ Dev Environment
   - Task: TASK-00000001
   - Tests: npm run test:dev
   - Deploy: ./deploy.sh dev
   - URL: https://dev.example.com

2. ⏳ Staging Environment (depends on: 1)
   - Task: TASK-00000002
   - Tests: npm run test:staging
   - Deploy: ./deploy.sh staging
   - URL: https://staging.example.com

3. ⏳ Production Environment (depends on: 2)
   - Task: TASK-00000003
   - Tests: npm run test:prod
   - Deploy: ./deploy.sh production
   - URL: https://example.com

**执行策略**:
- 串行部署（按环境顺序）
- 每个环境部署前运行测试
- 测试失败则停止后续部署
- Production 使用最高优先级

[开始执行...]

✅ Dev 部署完成！
- Tests: ✅ 通过 (12.3s)
- Deploy: ✅ 成功 (8.1s)
- URL: https://dev.example.com
- Health check: ✅ 正常

→ 开始 Staging 部署...

✅ Staging 部署完成！
- Tests: ✅ 通过 (15.7s)
- Deploy: ✅ 成功 (10.2s)
- URL: https://staging.example.com
- Health check: ✅ 正常
- Smoke tests: ✅ 通过

→ 开始 Production 部署...

⚠️ Production 测试失败！

**错误**: 2 tests failed
- test/api.js: User authentication
- test/payment.js: Payment processing

**已停止**: Production 部署已取消

**建议**:
1. 修复失败的测试
2. 在 Staging 环境验证修复
3. 重新执行 Production 部署

You: 看看测试失败的日志

Claude: [使用 AgentFlow 获取日志...]

**测试日志** (TASK-00000003):

\`\`\`
FAIL test/api.js (123ms)
  User authentication
    ✓ Should authenticate with valid credentials
    ✗ Should reject invalid credentials
      Error: Expected 401 but got 200

FAIL test/payment.js (456ms)
  Payment processing
    ✓ Should process valid payment
    ✗ Should handle insufficient funds
      Error: Timeout after 30000ms

Stack trace:
  at Payment.process (src/payment.js:45)
    ...
\`\`\`

**分析**:
1. 认证测试 - 生产环境可能使用了测试配置
2. 支付测试 - 超时，可能是 API 响应慢

**修复建议**:
1. 检查生产环境配置文件
2. 增加支付测试超时时间
3. 添加更好的错误处理

You: 好的，修复这些问题并重新部署到 production

Claude: [修复代码...]

[使用 AgentFlow 重新部署...]

✅ Production 部署完成！
- Tests: ✅ 通过 (18.2s)
- Deploy: ✅ 成功 (12.3s)
- URL: https://example.com
- Health check: ✅ 正常
- Smoke tests: ✅ 通过
```

### 场景 4：定时任务管理

**对话示例**：

```bash
You: 设置每天凌晨 2 点运行数据库备份，每周日晚上运行性能分析

Claude: 我将创建这些定期任务并使用 AgentFlow 管理。

**创建定期任务**:

1. 每日数据库备份
   \`\`\`bash
   agentflow create \
     --title "Daily DB Backup" \
     --description "./scripts/backup.sh" \
     --priority medium \
     --group maintenance
   \`\`\`

2. 每周性能分析
   \`\`\`bash
   agentflow create \
     --title "Weekly Performance Analysis" \
     --description "./scripts/performance-analysis.sh" \
     --priority low \
     --group maintenance
   \`\`\`

**Cron 配置**:

需要添加到 crontab:

\`\`\`cron
# AgentFlow 定期任务
0 2 * * * agentflow run "./scripts/backup.sh" --title "Daily DB Backup" --group maintenance
0 22 * * 0 agentflow run "./scripts/performance-analysis.sh" --title "Weekly Performance Analysis" --group maintenance
\`\`\`

**是否要自动添加到 crontab?**

You: 是的，添加

Claude: [添加到 crontab...]

✅ Cron 任务已创建！

**当前 crontab**:
\`\`\`
0 2 * * * agentflow run "./scripts/backup.sh" --title "Daily DB Backup" --group maintenance
0 22 * * 0 agentflow run "./scripts/performance-analysis.sh" --title "Weekly Performance Analysis" --group maintenance
\`\`\`

**下次执行时间**:
- Daily DB Backup: 明天 02:00
- Weekly Performance Analysis: 本周日 22:00

**查看任务历史**: agentflow list --group maintenance
```

### 场景 5：故障排查和修复

**对话示例**：

```bash
You: 生产环境出问题了，用户反馈登录失败

Claude: 我将使用 AgentFlow 快速诊断和修复问题。

**诊断步骤**:

1. 检查日志
   \`\`\`bash
   agentflow run "journalctl -u auth-service -n 100" \
     --title "Check Auth Logs" \
     --priority high \
     --group production
   \`\`\`

2. 检查服务状态
   \`\`\`bash
   agentflow run "systemctl status auth-service" \
     --title "Check Service Status" \
     --priority high \
     --group production
   \`\`\`

3. 重启服务（如果需要）
   \`\`\`bash
   agentflow run "systemctl restart auth-service" \
     --title "Restart Auth Service" \
     --priority high \
     --group production
   \`\`\`

[并行执行诊断任务...]

✅ 诊断完成！

**发现的问题**:

1. **Auth Service 状态**: ❌ 失败 (exit code 1)
   - Service: auth-service
   - Status: failed (Result: exit-code)
   - Uptime: 0s

2. **日志分析**: ✅ 完成
   \`\`\`
   ERROR: Database connection failed
   ERROR: connect ECONNREFUSED 127.0.0.1:5432
   ERROR: Failed to start authentication service
   \`\`\`

**根本原因**: PostgreSQL 数据库未运行

**修复方案**:

1. 启动数据库
   \`\`\`bash
   agentflow run "systemctl start postgresql" \
     --title "Start PostgreSQL" \
     --priority critical \
     --group production
   \`\`\`

2. 验证数据库
   \`\`\`bash
   agentflow run "pg_isready -h localhost" \
     --title "Verify DB Connection" \
     --priority high \
     --group production
   \`\`\`

3. 重启 Auth Service
   \`\`\`bash
   agentflow run "systemctl restart auth-service" \
     --title "Restart Auth Service" \
     --priority high \
     --group production
   \`\`\`

[执行修复任务...]

✅ 修复完成！

**结果**:
1. ✅ PostgreSQL - 已启动 (2.1s)
2. ✅ DB Connection - 正常 (0.3s)
3. ✅ Auth Service - 运行中 (3.7s)

**验证服务**:
\`\`\`bash
agentflow run "curl https://api.example.com/health" \
  --title "Health Check" \
  --priority high \
  --group production
\`\`\`

✅ 服务已恢复！用户登录功能正常。

**后续建议**:
1. 调查 PostgreSQL 为什么停止运行
2. 添加数据库监控告警
3. 设置自动重启策略
```

## 高级用法

### 智能任务分解

```bash
You: 实现用户认证功能，包括注册、登录、密码重置

Claude: 我将把这个大任务分解为多个小任务，使用 AgentFlow 管理。

**任务分解**:

1. 设计数据模型 (30min)
   \`\`\`bash
   agentflow create \
     --title "Design Auth Data Model" \
     --description "Create User schema with email, password_hash, etc." \
     --priority high
   \`\`\`

2. 实现注册 API (1hr)
   \`\`\`bash
   agentflow create \
     --title "Implement Registration API" \
     --description "POST /api/auth/register endpoint" \
     --priority high
   \`\`\`

3. 实现登录 API (1hr)
   \`\`\`bash
   agentflow create \
     --title "Implement Login API" \
     --description "POST /api/auth/login with JWT tokens" \
     --priority high
   \`\`\`

4. 实现密码重置 (1.5hr)
   \`\`\`bash
   agentflow create \
     --title "Implement Password Reset" \
     --description "POST /api/auth/reset with email tokens" \
     --priority medium
   \`\`\`

5. 编写单元测试 (1hr)
   \`\`\`bash
   agentflow create \
     --title "Auth Unit Tests" \
     --description "Test all auth endpoints" \
     --priority medium
   \`\`\`

6. 编写集成测试 (1hr)
   \`\`\`bash
   agentflow create \
     --title "Auth Integration Tests" \
     --description "Test auth flow end-to-end" \
     --priority medium
   \`\`\`

7. 更新 API 文档 (30min)
   \`\`\`bash
   agentflow create \
     --title "Update API Documentation" \
     --description "Document auth endpoints" \
     --priority low
   \`\`\`

**依赖关系**:
- Tasks 2,3,4 依赖 1
- Task 5 依赖 2,3,4
- Task 6 依赖 5
- Task 7 依赖 6

**总预计时间**: ~6.5 hours

**开始执行?**

You: 是的，开始执行

Claude: [使用 AgentFlow 按依赖关系执行任务...]

[执行进度...]

✅ 功能实现完成！

**任务摘要**:
- 设计数据模型: ✅ (28min)
- 注册 API: ✅ (1hr 3min)
- 登录 API: ✅ (1hr 1min)
- 密码重置: ✅ (1hr 27min)
- 单元测试: ✅ (1hr 5min)
- 集成测试: ✅ (1hr 2min)
- API 文档: ✅ (25min)

**总计**: 6小时31分钟

**代码审查**:
需要我进行代码审查吗？

You: 需要

Claude: [审查代码...]

发现以下问题：
1. 密码未使用 bcrypt 加密
2. JWT token 没有过期时间
3. 缺少速率限制
4. 未验证邮箱格式

**创建修复任务**:
\`\`\`bash
agentflow create --title "Fix Security Issues" --description "..."
\`\`\`

✅ 修复任务已创建
```

## 最佳实践

### 1. 任务命名规范

```bash
# ✅ 好的命名
agentflow create \
  --title "Run Unit Tests" \
  --description "npm test"

# ❌ 不好的命名
agentflow create \
  --title "test" \
  --description "npm test"
```

### 2. 合理设置优先级

```bash
# 关键任务
--priority high

# 普通任务
--priority medium

# 后台任务
--priority low
```

### 3. 使用 Worker Groups

```bash
# 生产环境
--group production

# 测试环境
--group testing

# 开发环境
--group development
```

### 4. 监控任务状态

```bash
# 定期检查
agentflow stats

# 查看特定任务
agentflow status TASK-00000001

# 查看日志
agentflow logs TASK-00000001
```

### 5. 错误处理

```bash
# 任务失败后
1. 查看日志
2. 分析原因
3. 修复问题
4. 重新执行
```

## 总结

### 核心优势

1. **智能任务管理**
   - Claude 理解意图并规划任务
   - AgentFlow 负责执行和调度

2. **分布式执行**
   - 多机器并行处理
   - 负载均衡和容错

3. **持久化和监控**
   - 任务状态持久化
   - 实时进度跟踪

4. **生产就绪**
   - 可靠的执行机制
   - 完善的错误处理

### 使用原则

1. **简单任务** → 直接执行
2. **复杂任务** → AgentFlow
3. **需要分析** → Claude + AgentFlow
4. **批量操作** → AgentFlow

### 下一步

1. ✅ 安装 AgentFlow skill
2. ✅ 尝试基本命令
3. ⏳ 集成到你的工作流
4. ⏳ 定制化配置

---

**最后更新**: 2026-01-24
**版本**: 1.0.0
