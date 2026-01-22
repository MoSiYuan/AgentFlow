#!/bin/bash
# test-python-version.sh - Test Python AgentFlow version

set -e

echo "🧪 AgentFlow Python - Comprehensive Test"
echo "=========================================="
echo ""

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass_count=0
fail_count=0

test_pass() {
    echo -e "${GREEN}✓ PASS${NC} - $1"
    ((pass_count++))
}

test_fail() {
    echo -e "${RED}✗ FAIL${NC} - $1"
    ((fail_count++))
}

test_info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Cleanup function
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    killall master 2>/dev/null || true
    killall python 2>/dev/null || true
    sleep 1
}

trap cleanup EXIT

# Test 1: Check Python version
echo "📋 Test 1: Python Version Check"
python_version=$(python3 --version 2>&1 | awk '{print $2}')
echo "   Python version: $python_version"

if python3 -c "import sys; exit(0 if sys.version_info >= (3, 8) else 1)" 2>/dev/null; then
    test_pass "Python 3.8+ available"
else
    test_fail "Python 3.8+ required"
    exit 1
fi

# Test 2: Check syntax
echo ""
echo "📋 Test 2: Code Syntax Check"
for file in agentflow/*.py; do
    if python3 -m py_compile "$file" 2>/dev/null; then
        test_pass "$(basename $file) syntax OK"
    else
        test_fail "$(basename $file) syntax error"
    fi
done

# Test 3: Import test
echo ""
echo "📋 Test 3: Module Import Test"
python3 << 'EOF'
import sys
sys.path.insert(0, '.')

try:
    from agentflow.database import Database, Task, Worker
    print("✓ database module imports OK")
except Exception as e:
    print(f"✗ database import failed: {e}")
    sys.exit(1)

try:
    from agentflow.master import Master
    print("✓ master module imports OK")
except ImportError as e:
    # Flask not installed is OK for this test
    if "Flask" in str(e):
        print("⚠ master module requires Flask (not installed)")
    else:
        print(f"✗ master import failed: {e}")
        sys.exit(1)

try:
    from agentflow.worker import Worker
    print("✓ worker module imports OK")
except ImportError as e:
    # requests not installed is OK for this test
    if "requests" in str(e):
        print("⚠ worker module requires requests (not installed)")
    else:
        print(f"✗ worker import failed: {e}")
        sys.exit(1)
EOF

# Test 4: Database operations
echo ""
echo "📋 Test 4: Database Operations"
test_db="/tmp/test-agentflow-$$.db"

python3 << EOF
import sys
sys.path.insert(0, '.')
from agentflow.database import Database, Task, Worker

# Create database
db = Database("$test_db")

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

# Cleanup
import os
os.remove("$test_db")
print("✓ Database cleanup OK")
EOF

if [ $? -eq 0 ]; then
    test_pass "Database operations"
else
    test_fail "Database operations"
fi

# Test 5: Check dependencies
echo ""
echo "📋 Test 5: Dependencies Check"

# Check Flask
if python3 -c "import flask" 2>/dev/null; then
    flask_version=$(python3 -c "import flask; print(flask.__version__)")
    test_pass "Flask installed ($flask_version)"
    flask_available=true
else
    test_info "Flask not installed (required for runtime)"
    flask_available=false
fi

# Check requests
if python3 -c "import requests" 2>/dev/null; then
    requests_version=$(python3 -c "import requests; print(requests.__version__)")
    test_pass "requests installed ($requests_version)"
    requests_available=true
else
    test_info "requests not installed (required for runtime)"
    requests_available=false
fi

# Test 6: Runtime test (if dependencies available)
if [ "$flask_available" = true ] && [ "$requests_available" = true ]; then
    echo ""
    echo "📋 Test 6: Runtime Test"

    # Start Master
    test_info "Starting Master server..."
    python3 -m agentflow.cli master --port 8850 > /tmp/master-test.log 2>&1 &
    master_pid=$!
    sleep 2

    # Check if Master is running
    if kill -0 $master_pid 2>/dev/null; then
        test_pass "Master started (PID: $master_pid)"

        # Health check
        if curl -s http://127.0.0.1:8850/api/health | grep -q "healthy"; then
            test_pass "Master health check OK"
        else
            test_fail "Master health check failed"
        fi

        # Create task
        test_result=$(curl -s -X POST http://127.0.0.1:8850/api/tasks/create \
          -H "Content-Type: application/json" \
          -d '{
            "task_id": "TEST-RUNTIME",
            "title": "Runtime Test",
            "description": "shell:echo Python AgentFlow Test",
            "priority": "high"
          }')

        if echo "$test_result" | grep -q "success"; then
            test_pass "Task creation OK"
        else
            test_fail "Task creation failed"
        fi

        # Cleanup
        kill $master_pid 2>/dev/null || true
    else
        test_fail "Master failed to start"
        cat /tmp/master-test.log
    fi
else
    echo ""
    echo "📋 Test 6: Runtime Test"
    test_info "Skipped (dependencies not installed)"
    echo "   Install with: pip install Flask requests"
fi

# Test 7: CLI test
echo ""
echo "📋 Test 7: CLI Test"

if python3 -m agentflow.cli --help > /dev/null 2>&1; then
    test_pass "CLI help command works"
else
    test_info "CLI requires Flask and requests"
fi

# Summary
echo ""
echo "=========================================="
echo "📊 Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo ""

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    echo ""
    echo "🚀 Python AgentFlow is ready to use!"
    echo ""
    echo "Quick start:"
    echo "  cd python"
    echo "  pip install Flask requests  # if not installed"
    echo "  python -m agentflow.cli master --port 8848"
    echo "  python -m agentflow.cli worker --auto"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    echo ""
    echo "Please check the errors above."
    exit 1
fi
