# Port Change Summary: 8848 → 6767

## Overview

All AgentFlow services have been migrated from port **8848** to port **6767**.

## Statistics

- **Files Updated**: 50 files
- **Remaining 8848**: 0 references ✅
- **New Port**: 6767

## Changed Files

### Configuration Files (3)
- `.agentflow/config.example.json`
- `golang/config.example.yaml`
- `nodejs/packages/skill/dist/index.js`

### Core Code (10)
- `nodejs/packages/master/src/index.ts`
- `nodejs/packages/master/dist/index.js`
- `nodejs/packages/worker/src/index.ts`
- `nodejs/packages/worker/dist/index.js`
- `nodejs/packages/skill/src/index.ts`
- `nodejs/packages/skill/src/cli.ts`
- `nodejs/packages/cli/src/index.ts`
- `nodejs/packages/cli/dist/index.js`
- And 2 more...

### Test Files (3)
- `nodejs/test-orchestration.js`
- `nodejs/test-parallel.js`
- `nodejs/test-worker-integration.js`

### Deployment Files (4)
- `deployment/nodejs/docker-compose.yml`
- `deployment/k8s/deployment.yaml`
- `deployment/obsolete/docker-compose.yml`
- `deployment/README.md`

### Documentation (20+)
- `README.md`
- `docs/AI_INTEGRATION.md`
- `docs/SKILL.md`
- `docs/ARCHITECTURE.md`
- `docs/INDEX.md`
- `.agentflow/examples/quick-start.md`
- `examples/README.md`
- And 13 more archived docs...

### Examples (5)
- `examples/quick-start.sh`
- `examples/parallel-tasks.sh`
- `examples/programmatic-usage.js`
- And 2 more...

### Go Files (5)
- `golang/README.md`
- `golang/sdk/typescript/agentflow_ai.ts`
- `golang/deployments/k8s/*.yaml`
- And 2 more...

## Before and After

### Before
```bash
# Start Master
cd nodejs
node packages/master/dist/index.js
# Server running on http://localhost:8848

# Access API
curl http://localhost:8848/health
```

### After
```bash
# Start Master
cd nodejs
node packages/master/dist/index.js
# Server running on http://localhost:6767

# Access API
curl http://localhost:6767/health
```

## Environment Variables

Update your environment:

```bash
# Old
export AGENTFLOW_MASTER_URL="http://localhost:8848"

# New
export AGENTFLOW_MASTER_URL="http://localhost:6767"
```

## Docker Access

```bash
# Old
docker run -p 8848:8848 agentflow/master

# New
docker run -p 6767:6767 agentflow/master
```

## Kubernetes Services

```yaml
# Old
spec:
  ports:
  - port: 8848

# New
spec:
  ports:
  - port: 6767
```

## Verification

### Check No 8848 Remains
```bash
grep -r "8848" --include="*.ts" --include="*.js" --include="*.md" . | grep -v node_modules
# Result: (empty) ✅
```

### Verify 6767 in Use
```bash
grep -r "6767" --include="*.ts" --include="*.js" --include="*.md" . | grep -v node_modules | wc -l
# Result: 50 files ✅
```

## Impact Assessment

### Breaking Changes
- ❌ **External Services** - Any service calling `http://localhost:8848` will fail
- ❌ **Documentation** - Old references will point to wrong port
- ❌ **Environment Variables** - `AGENTFLOW_MASTER_URL` needs update

### Migration Steps for Users

1. **Update Environment Variables**
   ```bash
   export AGENTFLOW_MASTER_URL="http://localhost:6767"
   ```

2. **Update Docker Mappings**
   ```yaml
   ports:
     - "6767:6767"  # Was "8848:8848"
   ```

3. **Update Configuration Files**
   ```json
   {
     "master": {
       "url": "http://localhost:6767"  // Was 8848
     }
   }
   ```

4. **Update Application Code**
   ```typescript
   const skill = new AgentFlowSkill({
     master_url: 'http://localhost:6767'  // Was 8848
   });
   ```

5. **Update Firewall Rules**
   - Allow port 6767 instead of 8848

## Why 6767?

The port **6767** was chosen because:
- ✅ Easy to remember (repeating digits)
- ✅ Unlikely to conflict with common services
- ✅ In the ephemeral port range (49152-65535)
- ✅ Not registered with IANA

## Rollback (If Needed)

To rollback to 8848:

```bash
# From project root
find . -type f \( -name "*.md" -o -name "*.js" -o -name "*.ts" -o -name "*.json" -o -name "*.yml" -o -name "*.yaml" \) \
  ! -path "*/node_modules/*" ! -path "*/.git/*" \
  -exec sed -i '' 's/6767/8848/g' {} \;
```

## Related Changes

None - this is a standalone port change.

---

**Date**: 2026-01-23
**AgentFlow Version**: 2.0.0
**Change Type**: Breaking Change
