#!/bin/bash
# AgentFlow Python - Quick Test (without pip install)
# This tests the code structure without running the server

echo "🧪 AgentFlow Python - Code Structure Test"
echo ""

# Test 1: Check Python syntax
echo "📋 Test 1: Checking Python syntax..."
python3 -m py_compile agentflow/__init__.py && echo "✓ __init__.py syntax OK"
python3 -m py_compile agentflow/database.py && echo "✓ database.py syntax OK"
python3 -m py_compile agentflow/master.py && echo "✓ master.py syntax OK"
python3 -m py_compile agentflow/worker.py && echo "✓ worker.py syntax OK"
python3 -m py_compile agentflow/cli.py && echo "✓ cli.py syntax OK"

# Test 2: Check database module
echo ""
echo "📋 Test 2: Testing database module..."
python3 << 'EOF'
import sys
sys.path.insert(0, '.')
from agentflow.database import Database, Task, Worker

# Create test database
db = Database("/tmp/test-agentflow.db")

# Test task creation
task = Task(
    task_id="TEST-1",
    title="Test Task",
    description="shell:echo test",
    priority="high"
)
assert db.create_task(task), "Failed to create task"
print("✓ Task creation OK")

# Test get task
retrieved = db.get_task("TEST-1")
assert retrieved is not None, "Failed to retrieve task"
assert retrieved.task_id == "TEST-1", "Task ID mismatch"
print("✓ Task retrieval OK")

# Test worker registration
worker = Worker(
    worker_id="WORKER-1",
    worker_name="test-worker",
    platform="linux"
)
assert db.register_worker(worker), "Failed to register worker"
print("✓ Worker registration OK")

# Test system status
status = db.get_system_status()
assert status["pending_tasks"] == 1, "Pending tasks count mismatch"
print("✓ System status OK")

db.close()
print("✓ All database tests passed")
EOF

# Test 3: Check imports
echo ""
echo "📋 Test 3: Testing module imports..."
python3 << 'EOF'
import sys
sys.path.insert(0, '.')

# Test that modules can be imported (even without Flask/requests)
try:
    from agentflow.database import Database, Task, Worker
    print("✓ Database module imports OK")
except Exception as e:
    print(f"✗ Database import failed: {e}")
    sys.exit(1)

# Test data classes
task = Task(task_id="T", title="T", description="D")
worker = Worker(worker_id="W", worker_name="W", platform="L")
print("✓ Data classes OK")
EOF

echo ""
echo "✅ All code structure tests passed!"
echo ""
echo "⚠️  Note: Full runtime tests require Flask and requests"
echo "   Install with: pip install Flask requests"
echo "   Or use: python3 -m venv venv && source venv/bin/activate && pip install Flask requests"
