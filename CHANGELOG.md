# AgentFlow Changelog

All notable changes to AgentFlow will be documented in this file.

## [2.0.0] - 2026-01-23

### Breaking Change - Port Migration
- **Default port changed from 8848 to 6767**
- All API endpoints, configuration files, and deployment manifests updated
- See [Port Migration Guide](PORT_MIGRATION_GUIDE.md) for migration instructions

### Added
- **`.agentflow/` directory** - Centralized configuration and templates
  - Agent templates (developer, tester, reviewer)
  - Skill definitions (git-operations, testing)
  - Workflow templates (feature-development)
  - Usage examples and workspace rules

- **New CLI commands**
  - `agentflow init` - Initialize AgentFlow in project directory
  - `agentflow info` - Check installation status
  - `agentflow update` - Update templates (coming soon)

- **Deployment directory** (`deployment/`)
  - Node.js Docker configurations
  - Kubernetes manifests
  - Comprehensive deployment guide

- **Port migration guides**
  - Complete migration steps
  - Technical details
  - Rollback instructions

### Changed
- **Master server** - Now runs on port 6767 by default
- **Documentation** - Reduced from 30+ to 4 core files (87% reduction)
  - Main README simplified by 83%
  - Created SKILL.md quick reference
  - Archived 20+ outdated documents

### Removed
- **`old/` directory** - Old Go implementation removed
- **`golang/sdk/python/`** - Python SDK removed (use Node.js instead)
- **`Makefile`** - Use npm/pnpm commands instead
- **Python examples** - Replaced with Node.js/CLI examples
- **Deployment scripts** - Moved to archives/old-scripts/

### Deprecated
- None

### Security
- No security issues addressed in this release

### Fixed
- None

### Migration Required
Yes - See [Port Migration Guide](PORT_MIGRATION_GUIDE.md)

**Upgrade Instructions**:
```bash
# 1. Update environment variable
export AGENTFLOW_MASTER_URL="http://localhost:6767"

# 2. Update application code
master_url: 'http://localhost:6767'

# 3. Update Docker/K8s configs
ports: ["6767:6767"]

# 4. Restart services
```

## [1.0.0] - Previous Versions

For earlier versions, see GitHub commit history.

---

**Version Format**: Major.Minor.Patch
**Migration Guides**: See [docs/](docs/) for detailed guides
**Support**: [GitHub Issues](https://github.com/MoSiYuan/AgentFlow/issues)
