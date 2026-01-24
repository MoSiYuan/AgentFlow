# AgentFlow 归档文档

本目录包含 AgentFlow 项目的历史文档、测试报告和迁移指南。

## 📁 目录结构

### reports/
测试报告和验证报告
- `FINAL_TEST_REPORT.md` - 最终测试报告
- `FINAL_GO_REPORT.md` - Go 版本最终报告
- `GO_VERSION_REPORT.md` - Go 版本详细报告
- `TEST_REPORT.md` - 基础测试报告

### comparisons/
版本和实现对比
- `COMPILATION_COMPARISON.md` - 编译对比
- `DEPLOYMENT_COMPARISON.md` - 部署对比
- `NODEJS20_BUILD_REPORT.md` - Node.js 20 构建报告
- `NODEJS20_SUCCESS_REPORT.md` - Node.js 20 成功报告
- `NODEJS22_TEST_REPORT.md` - Node.js 22 测试报告（不兼容）

### database/
数据库相关文档
- `SQLITE_STATUS_REPORT.md` - SQLite 状态报告
- `SQLITE_UPDATE_GUIDE.md` - SQLite 更新指南

### deployment/
部署相关文档
- `DEPLOYMENT.md` - 部署说明
- `GETTING_STARTED.md` - 快速开始指南

### migration/
迁移和版本升级
- `PORT_MIGRATION_GUIDE.md` - 端口迁移指南（8848 → 6767）

### nodejs/
Node.js 实现相关
- `LOCAL_EXECUTION.md` - 本地执行文档

### testing/
测试相关
- `TEST_GO_WORKER.md` - Go Worker 测试

### guides/
设置和配置指南
- `NODEJS_SETUP_GUIDE.md` - Node.js 环境设置

### historical/
历史文档和架构说明
- `AI_*.md` - AI 集成相关文档
- `API.md` - API 文档
- `ARCHITECTURE_SIMPLIFICATION.md` - 架构简化
- `WORKER_GROUPS.md` - Worker 组说明
- `agent-*.md` - 各种机制说明
- 其他历史文档

## 📋 文档索引

### 按时间排序

| 日期 | 文档 | 说明 |
|------|------|------|
| 2026-01-24 | NODEJS_GUIDE.md | 当前使用指南（外部） |
| 2026-01-24 | GO_VERSION_GUIDE.md | Go 版本指南（外部） |
| 2026-01-24 | INDEX.md | 文档索引（外部） |
| 2026-01-23 | NODEJS20_SUCCESS_REPORT.md | Node.js 20 成功报告 |
| 2026-01-23 | NODEJS22_TEST_REPORT.md | Node.js 22 测试（失败） |
| 2026-01-23 | NODEJS20_BUILD_REPORT.md | Node.js 20 构建报告 |
| 2026-01-23 | PORT_MIGRATION_GUIDE.md | 端口迁移指南 |
| 2026-01-22 | Various | 开发过程文档 |

### 按主题排序

**架构与设计:**
- ARCHITECTURE_SIMPLIFICATION.md
- unified-architecture.md
- API.md

**AI 集成:**
- AI_CLAUDE.md
- AI_DEPLOYMENT.md
- AI_QUICKSTART.md
- AI_README.md
- real-ai-proof.md

**部署:**
- DEPLOYMENT.md
- deployment-scripts.md
- mobile-development-guide.md

**开发:**
- GETTING_STARTED.md
- installation.md
- contributing.md

**Node.js:**
- LOCAL_EXECUTION.md
- nodejs-executive-summary.md
- nodejs-local-summary.md
- nodejs-version-analysis.md

**测试:**
- TEST_REPORT.md
- TEST_GO_WORKER.md
- worker-integration-test-results.md

**比较:**
- nodejs-vs-go-summary.md
- nodejs-vs-python-local.md
- go-master-nodejs-worker.md

## ⚠️ 文档状态说明

### 仍相关但已归档
这些文档描述了开发过程中的重要决策和实现细节，保留用于历史参考：
- Node.js 22 测试结果（展示兼容性问题）
- 端口迁移指南（v1.x → v2.0）
- 数据库升级指南

### 过时文档
以下文档可能已过时，仅供参考：
- Python SDK 相关（已移除）
- 旧部署脚本（已被新方法替代）
- 临时性配置指南

## 🔍 查找历史文档

如果您在寻找特定的历史文档：

1. **测试报告** → `reports/`
2. **版本比较** → `comparisons/`
3. **设置指南** → `guides/`
4. **架构文档** → `historical/`
5. **Node.js 开发** → `nodejs/`

## 📖 相关文档

当前活跃的文档请查看项目根目录：
- **[README.md](../../README.md)** - 主文档
- **[CHANGELOG.md](../../CHANGELOG.md)** - 变更日志
- **[CLI 指南](../../AGENTFLOW_CLI_GUIDE.md)** - CLI 使用
- **[文档索引](../INDEX.md)** - 完整索引

---

**最后更新**: 2026-01-24
**状态**: 归档完成，仅供历史参考
