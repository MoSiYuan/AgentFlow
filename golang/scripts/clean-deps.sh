#!/bin/bash
# clean-deps.sh - Clean unused Go dependencies

set -e

echo "🧹 Cleaning Go dependencies..."
echo ""

# Check if we're in the golang directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found"
    echo "   Please run this script from the golang/ directory"
    exit 1
fi

# Backup current go.mod
echo "📦 Backing up go.mod..."
cp go.mod go.mod.backup

# Clean dependencies
echo "🔧 Running go mod tidy..."
go mod tidy

# Show what changed
echo ""
echo "📊 Changes:"
if command -v diff &> /dev/null; then
    diff go.mod.backup go.mod || true
fi

# Clean up backup
echo ""
echo "🧹 Cleaning up..."
rm go.mod.backup

echo ""
echo "✅ Dependencies cleaned successfully!"
echo ""
echo "📝 Summary:"
echo "   - Removed unused dependencies"
echo "   - Added missing dependencies"
echo "   - Updated indirect dependencies"
