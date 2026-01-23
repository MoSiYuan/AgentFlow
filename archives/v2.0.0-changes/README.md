# AgentFlow v2.0.0 Change Documentation

This directory contains detailed documentation about the v2.0.0 restructuring and port migration.

## Quick Links

For users, see the main **[Port Migration Guide](../../PORT_MIGRATION_GUIDE.md)**.

## Documents in This Directory

### 1. **PORT_CHANGE.md**
Technical details about the port migration (8848 → 6767).
- File statistics (50 files updated)
- Before/after comparisons
- Migration impact assessment

### 2. **CHANGELOG_PORT.md**
Version changelog for the port change.
- Breaking change notice
- Migration requirements
- Testing status

### 3. **PORT_UPDATE_SUMMARY.md**
Complete work summary of all updates.
- Code changes (50 files)
- Documentation updates (20+ files)
- Verification results

### 4. **DEPLOYMENT_REORG.md**
Deployment directory reorganization details.
- Old vs new structure
- Docker/K8s configurations
- Migration paths

### 5. **CLEANUP.md**
Project cleanup summary.
- Removed directories (old/, Python SDK)
- Deleted files (Makefile, scripts)
- Archive locations

### 6. **MAKEFILE_CLEANUP.md**
Build system cleanup details.
- Removed Makefile
- Moved to npm/pnpm
- Updated documentation

## Summary of Changes

### Port Migration
- Old Port: 8848
- New Port: 6767
- Files Updated: 50
- Remaining 8848 References: 0 ✅

### Project Restructure
- New `.agentflow/` directory
- New `deployment/` directory
- Documentation reduced by 87%
- Examples updated to Node.js/CLI

### Cleanup
- Removed: old/ (100+ files)
- Removed: golang/sdk/python/
- Removed: Makefile
- Archived: 20+ outdated docs

## For Users

If you're upgrading to v2.0.0, follow these steps:

1. **Read the migration guide**: [PORT_MIGRATION_GUIDE.md](../../PORT_MIGRATION_GUIDE.md)
2. **Update environment variables**:
   ```bash
   export AGENTFLOW_MASTER_URL="http://localhost:6767"
   ```
3. **Update Docker/K8s configs**:
   ```yaml
   ports:
     - "6767:6767"
   ```
4. **Restart services** and verify:
   ```bash
   curl http://localhost:6767/health
   ```

## For Developers

Technical implementation details are available in:
- **Architecture**: [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md)
- **AI Integration**: [docs/AI_INTEGRATION.md](../../docs/AI_INTEGRATION.md)
- **Deployment**: [deployment/README.md](../../deployment/README.md)

## Version History

- v2.0.0 (2026-01-23): Major restructure + Port migration
- v1.0.0: Initial releases

---

**AgentFlow Version**: 2.0.0
**Date**: 2026-01-23
