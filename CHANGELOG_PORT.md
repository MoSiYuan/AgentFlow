# Port Change Update Log

## v2.0.0 (2026-01-23)

### Breaking Change: Port Migration

**Summary**: Default port changed from `8848` to `6767`

#### Impact
- **Master Server**: Now runs on port 6767 by default
- **API Endpoints**: All API calls must use new port
- **Environment Variables**: `AGENTFLOW_MASTER_URL` default changed
- **Documentation**: All examples updated to use port 6767

#### Migration Required

**Yes** - All users must update their configurations:

1. **Environment Variables**
   ```bash
   export AGENTFLOW_MASTER_URL="http://localhost:6767"
   ```

2. **Application Code**
   ```typescript
   master_url: 'http://localhost:6767'
   ```

3. **Docker Configuration**
   ```yaml
   ports:
     - "6767:6767"
   ```

4. **Firewall Rules**
   - Allow port 6767/tcp
   - Remove port 8848/tcp (if no longer needed)

#### Files Changed

- **50 files** updated across the codebase
- **0 files** still reference old port 8848
- Complete list in [PORT_CHANGE.md](PORT_CHANGE.md)

#### Documentation

- ✅ All documentation updated
- ✅ All examples updated
- ✅ All deployment configs updated
- ✅ Migration guide created: [PORT_MIGRATION_GUIDE.md](PORT_MIGRATION_GUIDE.md)

#### Rollback

To rollback to port 8848, see the [Migration Guide](PORT_MIGRATION_GUIDE.md#回滚方案).

#### Support

For issues or questions:
- Documentation: [PORT_MIGRATION_GUIDE.md](PORT_MIGRATION_GUIDE.md)
- GitHub Issues: https://github.com/MoSiYuan/AgentFlow/issues

---

## Why This Change?

1. **Better Memorability**: 6767 is easier to remember (repeating digits)
2. **Fewer Conflicts**: Less likely to conflict with other services
3. **Consistency**: Aligns with industry best practices for ephemeral ports
4. **IANA Compliance**: Uses port in the dynamic/private range (49152-65535)

## Testing

All automated tests have been updated to use the new port:
- ✅ Unit tests
- ✅ Integration tests
- ✅ Worker tests
- ✅ Orchestration tests

## Next Steps

1. ✅ Update your environment variables
2. ✅ Update application code
3. ✅ Update Docker/K8s configurations
4. ✅ Restart services
5. ✅ Verify connectivity: `curl http://localhost:6767/health`

---

**AgentFlow Version**: 2.0.0
**Release Date**: 2026-01-23
**Status**: Production Ready
