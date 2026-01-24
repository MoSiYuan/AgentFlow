# AgentFlow Changelog

All notable changes to AgentFlow will be documented in this file.

## [2.1.0] - 2026-01-24

### Added - Node.js 20 LTS Support
- **Full Node.js 20 compatibility** - Tested with v20.19.6 LTS
- **better-sqlite3 v12.6.2 integration** - Rebuild for Node.js 20 ABI
- **LocalExecutor package** - Automatic Master/Worker lifecycle management
- **CLI `run` command** - One-line task execution
- **Comprehensive documentation**
  - [CLI Guide](AGENTFLOW_CLI_GUIDE.md) - Complete CLI usage guide
  - [Node.js Guide](docs/NODEJS_GUIDE.md) - Development and deployment guide
  - [Go Guide](docs/GO_VERSION_GUIDE.md) - Go implementation details
  - [Documentation Index](docs/INDEX.md) - Complete documentation navigation

### Fixed
- **Worker JSON Parse Error** - Fixed 204 No Content handling in task polling
- **Worker Heartbeat Error** - Silently ignore ECONNREFUSED during shutdown
- **Task ID Format Inconsistency** - Unified ID format across all APIs
- **Route Order Issue** - Fixed `/api/v1/tasks/pending` route precedence

### Changed
- **Master main entry point** - Added automatic server startup when run directly
- **Worker main entry point** - Added worker registration and graceful shutdown
- **Documentation structure** - Reduced from 20+ to 4 core files
- **Archived 15 historical reports** - Moved to `docs/archive/`

### Performance
- **Startup time**: ~2s (cold start), ~1s (warm start)
- **Task execution**: Immediate for simple commands
- **Memory usage**: ~80MB (Master + Worker)
- **Zero configuration**: Works out of the box

### Documentation
- Created unified README with 3 usage options
- Added comprehensive CLI guide with examples
- Added Node.js development guide
- Added Go version guide
- Created documentation index
- Archived all historical reports

## [2.0.0] - 2026-01-23

### Breaking Change - Port Migration
- **Default port changed from 8848 to 6767**
- All API endpoints, configuration files, and deployment manifests updated
- See [Port Migration Guide](docs/archive/PORT_MIGRATION_GUIDE.md) for migration instructions

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
- - Main README simplified by 83%
- - Created SKILL.md quick reference
- - Archived 20+ outdated documents

### Removed
- **`old/` directory** - Old Go implementation removed
- **`golang/sdk/python/`** - Python SDK removed (use Node.js instead)
- **`Makefile`** - Use npm/pnpm commands instead
- **Python examples** - Replaced with Node.js/CLI examples
- **Deployment scripts** - Moved to archives/old-scripts/

### Deprecated
- None

## [1.0.0] - 2026-01-20

### Initial Release
- Go implementation of Master and Worker
- Node.js implementation of Master and Worker
- 100% API-compatible dual-language support
- RESTful API for task management
- SQLite database integration
- WebSocket support for real-time updates
- Claude CLI integration for complex tasks
- Comprehensive deployment guides
- Multi-platform support (macOS, Linux, Windows)
