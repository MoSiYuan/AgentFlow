#!/bin/bash
# Simple test for Python AgentFlow

echo "🧪 Testing Python AgentFlow..."
echo ""

# Test 1: Python version
echo "1. Python Version:"
python3 --version
echo ""

# Test 2: Code syntax
echo "2. Checking syntax..."
for file in agentflow/*.py; do
    echo -n "   $file: "
    if python3 -m py_compile "$file" 2>/dev/null; then
        echo "✓ OK"
    else
        echo "✗ FAILED"
        exit 1
    fi
done
echo ""

# Test 3: Database module
echo "3. Testing database module..."
python3 << 'EOF'
import sys
sys.path.insert(0, '.')
from agentflow.database import Database, Task, Worker

# Create test database
db = Database("/tmp/test-agentflow.db")

# Test task
task = Task(task_id="T1", title="Test", description="shell:echo test", priority="high")
assert db.create_task(task)
print("   ✓ Task creation OK")

# Test worker
worker = Worker(worker_id="W1", worker_name="test", platform="linux")
assert db.register_worker(worker)
print("   ✓ Worker registration OK")

# Test system status
status = db.get_system_status()
print(f"   ✓ System status: {status}")

db.close()

import os
os.remove("/tmp/test-agentflow.db")
print("   ✓ Database test PASSED")
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ All core tests PASSED!"
    echo ""
    echo "Note: Flask/requests not installed, skipping runtime tests."
    echo "Install with: pip install Flask requests"
else
    echo ""
    echo "❌ Tests FAILED"
    exit 1
fi
