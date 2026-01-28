#!/bin/bash
# Test script to verify AgentFlow compilation

set -e

echo "==================================================="
echo "AgentFlow Compilation Test Script"
echo "==================================================="
echo ""

# Check if rustc is available
if ! command -v rustc &> /dev/null; then
    echo "Error: rustc not found. Please install Rust toolchain."
    echo "Visit: https://rustup.rs/"
    exit 1
fi

echo "Rust toolchain detected:"
rustc --version
echo ""

# Navigate to rust directory
cd "$(dirname "$0")/rust"

echo "==================================================="
echo "Running cargo check --release..."
echo "==================================================="
echo ""

if cargo check --release 2>&1 | tee /tmp/agentflow_compile.log; then
    echo ""
    echo "==================================================="
    echo "✅ SUCCESS: Compilation completed successfully!"
    echo "==================================================="
    exit 0
else
    echo ""
    echo "==================================================="
    echo "❌ FAILED: Compilation encountered errors"
    echo "==================================================="
    echo ""
    echo "Last 50 lines of error log:"
    tail -50 /tmp/agentflow_compile.log
    exit 1
fi
