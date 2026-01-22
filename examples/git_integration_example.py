#!/usr/bin/env python3
"""
AgentFlow Git Integration Example

This script demonstrates how to use Git integration for multi-agent
collaborative development with file boundaries and locking.

Usage:
    python git_integration_example.py
"""

import os
import sys
from pathlib import Path

# Add parent directory to path to import agentflow
sys.path.insert(0, str(Path(__file__).parent.parent))

from agentflow.git_integration import (
    GitIntegrationManager,
    GitFileBoundaryManager,
    FileBoundary
)


def example_1_basic_usage():
    """Example 1: Basic Git integration usage"""
    print("\n" + "="*60)
    print("Example 1: Basic Git Integration Usage")
    print("="*60)

    # Initialize Git integration manager
    manager = GitIntegrationManager(
        repo_path=".",  # Current Git repository
        boundary_config=".agentflow/boundaries.example.json"
    )

    print("\n✅ Git Integration Manager initialized")

    # Create an agent task
    print("\n📝 Creating agent task...")
    task = manager.create_agent_task(
        agent_id="agent-frontend",
        task_id="TASK-001",
        description="实现用户登录页面 UI"
    )

    print(f"✅ Task created: {task['task_id']}")
    print(f"   Branch: {task['branch']}")
    print(f"   Boundaries: {len(task['boundaries'])} file boundaries")


def example_2_file_access_control():
    """Example 2: File access control with boundaries"""
    print("\n" + "="*60)
    print("Example 2: File Access Control")
    print("="*60)

    manager = GitIntegrationManager(".", ".agentflow/boundaries.example.json")

    # Test different access scenarios
    test_cases = [
        ("agent-frontend", "src/frontend/App.tsx", "write"),
        ("agent-frontend", "src/backend/auth.go", "write"),
        ("agent-backend", "src/api/types.ts", "write"),
        ("agent-database", "src/database/models.go", "write"),
    ]

    print("\n🔍 Testing file access permissions:")
    for agent_id, file_path, access_type in test_cases:
        allowed, reason = manager.verify_file_access(agent_id, file_path, access_type)
        status = "✅ ALLOWED" if allowed else "❌ DENIED"
        print(f"\n{status}")
        print(f"  Agent: {agent_id}")
        print(f"  File:  {file_path}")
        print(f"  Type:  {access_type}")
        if not allowed:
            print(f"  Reason: {reason}")


def example_3_file_locking():
    """Example 3: File locking mechanism"""
    print("\n" + "="*60)
    print("Example 3: File Locking")
    print("="*60)

    boundary_manager = GitFileBoundaryManager(".agentflow/boundaries.example.json")

    # Test file locking
    file_path = "src/api/types.ts"

    print(f"\n🔒 Testing file locks for: {file_path}")

    # Agent 1 acquires write lock
    print("\n1. Agent-backend tries to acquire write lock...")
    if boundary_manager.acquire_lock("agent-backend", file_path, "write"):
        print("   ✅ Lock acquired successfully")

        # Agent 2 tries to acquire write lock on same file
        print("\n2. Agent-frontend tries to acquire write lock...")
        if not boundary_manager.acquire_lock("agent-frontend", file_path, "write"):
            print("   ❌ Lock denied (already locked by agent-backend)")

        # Agent 2 tries to acquire read lock
        print("\n3. Agent-frontend tries to acquire read lock...")
        if not boundary_manager.acquire_lock("agent-frontend", file_path, "read"):
            print("   ❌ Lock denied (write lock exists)")

        # Release lock
        print("\n4. Agent-backend releases lock...")
        if boundary_manager.release_lock("agent-backend", file_path, "write"):
            print("   ✅ Lock released")

        # Now Agent 2 can acquire lock
        print("\n5. Agent-frontend tries again...")
        if boundary_manager.acquire_lock("agent-frontend", file_path, "write"):
            print("   ✅ Lock acquired successfully")


def example_4_safe_file_operations():
    """Example 4: Safe file operations with automatic locking"""
    print("\n" + "="*60)
    print("Example 4: Safe File Operations")
    print("="*60)

    manager = GitIntegrationManager(".", ".agentflow/boundaries.example.json")

    # Define a file operation
    def write_to_file(file_path):
        with open(file_path, 'w') as f:
            f.write("// Test content\n")
        return True

    # Perform safe operation
    print("\n📝 Performing safe file operation...")
    try:
        manager.safe_file_operation(
            agent_id="agent-backend",
            file_path="test_output.txt",
            operation=write_to_file
        )
        print("✅ File operation completed and committed")
    except PermissionError as e:
        print(f"❌ Operation failed: {e}")


def example_5_conflict_detection():
    """Example 5: Conflict detection"""
    print("\n" + "="*60)
    print("Example 5: Conflict Detection")
    print("="*60)

    manager = GitFileBoundaryManager(".agentflow/boundaries.example.json")

    # Acquire a lock
    file_path = "src/api/user.ts"
    manager.acquire_lock("agent-backend", file_path, "write")

    # Check for conflicts
    print("\n🔍 Checking for conflicts...")
    conflicts = manager.check_conflicts("agent-frontend", [file_path])

    if conflicts:
        print(f"\n⚠️  Found {len(conflicts)} conflicts:")
        for conflict in conflicts:
            print(f"\n  Conflict: {conflict.conflict_id}")
            print(f"  Type: {conflict.conflict_type}")
            print(f"  File: {conflict.file_paths[0]}")
            print(f"  Severity: {conflict.severity}")
            print(f"  Description: {conflict.description}")
    else:
        print("\n✅ No conflicts detected")

    # Clean up
    manager.release_lock("agent-backend", file_path, "write")


def example_6_complete_workflow():
    """Example 6: Complete workflow from task creation to completion"""
    print("\n" + "="*60)
    print("Example 6: Complete Workflow")
    print("="*60)

    manager = GitIntegrationManager(".", ".agentflow/boundaries.example.json")

    agent_id = "agent-frontend"
    task_id = "TASK-002"

    print(f"\n🚀 Starting complete workflow for {agent_id}/{task_id}")

    # Step 1: Create task
    print("\n1️⃣  Creating task...")
    task = manager.create_agent_task(
        agent_id=agent_id,
        task_id=task_id,
        description="实现用户个人中心页面"
    )
    print(f"   ✅ Task created on branch: {task['branch']}")

    # Step 2: Perform file operations (simulated)
    print("\n2️⃣  Performing file operations...")
    # In real scenario, Agent would perform multiple file operations here
    print("   (Simulating Agent work...)")

    # Step 3: Complete and merge
    print("\n3️⃣  Completing task and merging to main...")
    result = manager.complete_agent_task(
        agent_id=agent_id,
        task_id=task_id,
        merge_strategy="squash"
    )

    if result['status'] == 'completed':
        print(f"   ✅ {result['message']}")
    elif result['status'] == 'conflict':
        print(f"   ⚠️  {result['message']}")
        print(f"   Conflicts: {result['conflicts']}")


def main():
    """Run all examples"""
    print("\n" + "="*60)
    print("AgentFlow Git Integration Examples")
    print("="*60)

    try:
        example_1_basic_usage()
        example_2_file_access_control()
        example_3_file_locking()
        example_4_safe_file_operations()
        example_5_conflict_detection()
        # example_6_complete_workflow()  # Skip to avoid actual Git operations

        print("\n" + "="*60)
        print("✅ All examples completed successfully!")
        print("="*60)

    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    main()
