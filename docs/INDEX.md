# AgentFlow Documentation

## ⚠️ Important Notice

**Port Change (v2.0.0)**: Default port changed from `8848` to `6767`.
- [Migration Guide](../PORT_MIGRATION_GUIDE.md) - How to update your configuration
- [Port Change Summary](../PORT_CHANGE.md) - Technical details

## Core Documentation

- [README](../README.md) - Project overview and features
- [Skill Usage](SKILL.md) - Command reference and API
- [AI Integration](AI_INTEGRATION.md) - AI guide with examples ⭐
- [Architecture](ARCHITECTURE.md) - System design and components

## Quick Start

```bash
# Install AgentFlow skill
npm install -g @agentflow/skill

# Start Master server (now on port 6767)
cd nodejs
node packages/master/dist/index.js

# Set environment variable
export AGENTFLOW_MASTER_URL="http://localhost:6767"

# Create task
agentflow create "Run tests" -d "npm test"
```

## Archives

Historical documentation and analysis papers are in the [archives/](archives/) directory.

---

**Version**: 2.0.0 | **Last Updated**: 2026-01-23
