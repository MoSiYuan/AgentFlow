#!/usr/bin/env python3
"""
Complete CPDS Autonomous Test Script

Creates tasks, starts workers, and monitors the execution
"""

import requests
import time
import json
import sys
import subprocess

MASTER_URL = "http://localhost:8848"


def create_tasks(num_tasks=10):
    """Create autonomous story tasks"""
    print(f"📝 Creating {num_tasks} tasks...")

    for i in range(1, num_tasks + 1):
        task_id = f"STORY-AUTO-{i:03d}"
        payload = {
            "task_id": task_id,
            "title": f"Agent {i} 自主创作克苏鲁故事",
            "description": "你是专业的克苏鲁神话作家。请自主选择题材并创作500-1000字故事。流程：1) GET /api/topics 查询已选题材；2) 选择独特题材；3) POST /api/topics/register 注册；4) 创作。洛夫克拉夫特风格。",
            "priority": "high",
            "tags": ["creative-writing", "lovecraft", "autonomous"]
        }

        response = requests.post(
            f"{MASTER_URL}/api/tasks/create",
            json=payload,
            timeout=5
        )

        if response.status_code == 200:
            print(f"  ✅ Created {task_id}")
        else:
            print(f"  ❌ Failed to create {task_id}: {response.text}")

        time.sleep(0.1)  # Small delay to avoid DB locks

    print(f"\n✅ Task creation complete\n")


def run_workers():
    """Run the autonomous worker test script"""
    print("🤖 Starting autonomous workers...\n")

    # Run the worker test script
    script_path = "/Users/jiangxiaolong/work/project/game/AdStar/cpds-go/docs/cpds-test/test_autonomous_workers.py"
    result = subprocess.run(
        [sys.executable, script_path],
        capture_output=False
    )

    return result.returncode == 0


def show_results():
    """Display final results"""
    print("\n" + "=" * 60)
    print("📊 FINAL RESULTS")
    print("=" * 60)

    # Check topics
    response = requests.get(f"{MASTER_URL}/api/topics", timeout=5)
    if response.status_code == 200:
        data = response.json()
        topics = data.get("topics", [])
        print(f"\n🎯 Selected Topics: {len(topics)}")
        for i, topic in enumerate(topics, 1):
            worker_short = topic.get('worker_id', '')[:8]
            print(f"  {i:2d}. {topic['topic']}")
            print(f"      Worker: {worker_short}...")

    # Check completed tasks
    response = requests.get(f"{MASTER_URL}/api/tasks/completed", timeout=5)
    if response.status_code == 200:
        data = response.json()
        tasks = data.get("data", {}).get("tasks", [])
        print(f"\n✅ Completed Tasks: {len(tasks)}")
        for task in tasks:
            task_id = task['task_id']
            output = task.get('output', '')
            # Extract first line of story
            first_line = output.split('\n')[0][:60] if output else 'No output'
            print(f"  - {task_id}: {first_line}...")

    # Check pending tasks
    response = requests.get(f"{MASTER_URL}/api/tasks/pending", timeout=5)
    if response.status_code == 200:
        data = response.json()
        tasks = data.get("data", {}).get("tasks", [])
        if tasks:
            print(f"\n⏳ Pending Tasks: {len(tasks)}")
            for task in tasks:
                print(f"  - {task['task_id']}")

    print("\n" + "=" * 60)


def main():
    """Main test orchestration"""
    print("=" * 60)
    print("🚀 CPDS Autonomous Test - Complete Workflow")
    print("=" * 60)
    print(f"🌐 Master: {MASTER_URL}")
    print("=" * 60)

    # Step 1: Check Master is running
    try:
        response = requests.get(f"{MASTER_URL}/api/health", timeout=5)
        if response.status_code != 200:
            print("❌ Master is not healthy")
            return 1
        print("✅ Master is running\n")
    except Exception as e:
        print(f"❌ Cannot connect to Master: {e}")
        return 1

    # Step 2: Create tasks
    create_tasks(10)

    # Step 3: Run workers
    success = run_workers()

    # Step 4: Wait for completion
    print("\n⏳ Waiting for final cleanup...")
    time.sleep(3)

    # Step 5: Show results
    show_results()

    print(f"\n{'=' * 60}")
    if success:
        print("✨ Test completed successfully!")
    else:
        print("⚠️  Test completed with some issues")
    print(f"{'=' * 60}\n")

    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())
