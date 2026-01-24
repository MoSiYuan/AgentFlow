#!/bin/bash
# pnpm node_modules resolver for Node.js v24

cd /Users/jiangxiaolong/work/project/AgentFlow/nodejs

# Find all pnpm packages
PNPM_PACKAGES=$(find node_modules/.pnpm -maxdepth 1 -name "node_modules" -type d)

# Build NODE_PATH
export NODE_PATH=""
for pkg in $PNPM_PACKAGES; do
    if [ -z "$NODE_PATH" ]; then
        NODE_PATH="$pkg"
    else
        NODE_PATH="$NODE_PATH:$pkg"
    fi
done

# Also add workspace packages
export NODE_PATH="$NODE_PATH:packages/master/node_modules:packages/worker/node_modules:packages/skill/node_modules"

echo "NODE_PATH=$NODE_PATH"
exec "$@"
