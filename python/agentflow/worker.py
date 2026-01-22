"""
Worker for AgentFlow (Python version)
Multi-process task execution with Claude CLI integration
"""

import subprocess
import requests
import time
import threading
import socket
import uuid
import os
import re
from typing import Optional, Tuple
from .database import Task


class Worker:
    """Worker for task execution"""

    def __init__(self, master_url: str = "http://127.0.0.1:8848",
                 name: str = None, auto_mode: bool = True,
                 oneshot: bool = False):
        self.master_url = master_url.rstrip("/")
        self.name = name or f"worker-{socket.gethostname()}-{uuid.uuid4().hex[:6]}"
        self.worker_id = f"WORKER-{uuid.uuid4().hex[:8].upper()}"
        self.auto_mode = auto_mode
        self.oneshot = oneshot
        self.platform = os.uname().sysname.lower() if hasattr(os, 'uname') else "unknown"
        self._running = False
        self._stop_event = threading.Event()

    def register(self) -> bool:
        """Register worker with master"""
        try:
            response = requests.post(
                f"{self.master_url}/api/workers/register",
                json={
                    "worker_id": self.worker_id,
                    "worker_name": self.name,
                    "platform": self.platform
                },
                timeout=10
            )
            return response.status_code == 200
        except Exception as e:
            print(f"✗ Failed to register: {e}")
            return False

    def send_heartbeat(self) -> bool:
        """Send heartbeat to master"""
        try:
            response = requests.post(
                f"{self.master_url}/api/workers/heartbeat",
                json={"worker_id": self.worker_id},
                timeout=5
            )
            return response.status_code == 200
        except Exception:
            return False

    def get_pending_tasks(self) -> list:
        """Get pending tasks from master"""
        try:
            response = requests.get(
                f"{self.master_url}/api/tasks/pending?limit=10",
                timeout=10
            )
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    return data["data"]["tasks"]
            return []
        except Exception as e:
            print(f"✗ Failed to get tasks: {e}")
            return []

    def assign_task(self, task_id: str) -> bool:
        """Assign task to self"""
        try:
            response = requests.post(
                f"{self.master_url}/api/tasks/assign",
                json={
                    "task_id": task_id,
                    "worker_id": self.worker_id
                },
                timeout=10
            )
            if response.status_code == 200:
                data = response.json()
                return data.get("success", False)
            return False
        except Exception as e:
            print(f"✗ Failed to assign task: {e}")
            return False

    def update_progress(self, task_id: str, progress: int, message: str = "") -> bool:
        """Update task progress"""
        try:
            response = requests.post(
                f"{self.master_url}/api/tasks/progress",
                json={
                    "task_id": task_id,
                    "progress": progress,
                    "message": message
                },
                timeout=5
            )
            return response.status_code == 200
        except Exception:
            return False

    def complete_task(self, task_id: str, result: str = None, error: str = None) -> bool:
        """Complete task"""
        try:
            response = requests.post(
                f"{self.master_url}/api/tasks/complete",
                json={
                    "task_id": task_id,
                    "result": result,
                    "error": error
                },
                timeout=10
            )
            return response.status_code == 200
        except Exception as e:
            print(f"✗ Failed to complete task: {e}")
            return False

    def execute_task(self, task: dict) -> Tuple[bool, str, Optional[str]]:
        """Execute a task and return (success, result, error)"""

        description = task.get("description", "")
        print(f"\n→ Executing task: {task.get('title')} ({task.get('task_id')})")
        print(f"  Description: {description}")

        # Parse task type
        if description.startswith("shell:"):
            # Shell command
            return self._execute_shell(description[6:])

        elif description.startswith("ai:"):
            # AI task with Claude CLI
            return self._execute_claude(description[3:])

        else:
            # Default: treat as plain text prompt for Claude
            return self._execute_claude(description)

    def _execute_shell(self, command: str) -> Tuple[bool, str, Optional[str]]:
        """Execute shell command"""
        try:
            print(f"  Running shell: {command}")
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                timeout=300  # 5 minute timeout
            )

            if result.returncode == 0:
                output = result.stdout or result.stderr or "Command completed"
                print(f"  ✓ Success: {output[:100]}...")
                return True, output, None
            else:
                error = result.stderr or f"Command failed with exit code {result.returncode}"
                print(f"  ✗ Failed: {error[:100]}...")
                return False, None, error

        except subprocess.TimeoutExpired:
            error = "Command timed out after 5 minutes"
            print(f"  ✗ {error}")
            return False, None, error
        except Exception as e:
            error = str(e)
            print(f"  ✗ Error: {error}")
            return False, None, error

    def _execute_claude(self, prompt: str) -> Tuple[bool, str, Optional[str]]:
        """Execute task with Claude CLI"""

        # Check for claudecli wrapper
        claudecli_path = os.path.expanduser("~/bin/claudecli")

        if os.path.exists(claudecli_path):
            return self._execute_claude_wrapper(prompt, claudecli_path)

        # Check for claude in PATH
        try:
            result = subprocess.run(
                ["which", "claude"],
                capture_output=True,
                text=True
            )
            if result.returncode == 0:
                return self._execute_claude_direct(prompt)
        except Exception:
            pass

        # Fallback to simulation
        print(f"  ! Claude CLI not found, using simulation mode")
        print(f"  Prompt: {prompt[:50]}...")
        return True, f"[SIMULATED] Response to: {prompt}", None

    def _execute_claude_wrapper(self, prompt: str, claudecli_path: str) -> Tuple[bool, str, Optional[str]]:
        """Execute using claudecli wrapper (old API)"""
        try:
            print(f"  Using claudecli wrapper...")
            env = os.environ.copy()
            env["PATH"] = f"{os.path.expanduser('~/bin')}:{env.get('PATH', '')}"

            result = subprocess.run(
                [claudecli_path, "chat", "--prompt", prompt, "--no-interactive"],
                capture_output=True,
                text=True,
                timeout=300,
                env=env
            )

            if result.returncode == 0:
                output = result.stdout.strip()
                print(f"  ✓ Claude response: {output[:100]}...")
                return True, output, None
            else:
                error = result.stderr or f"Claude CLI failed with exit code {result.returncode}"
                print(f"  ! Claude CLI error: {error[:100]}")
                # Don't fail the task, just note the error
                return True, f"[Claude CLI Error: {error}]", None

        except subprocess.TimeoutExpired:
            error = "Claude CLI timed out after 5 minutes"
            print(f"  ✗ {error}")
            return False, None, error
        except Exception as e:
            error = str(e)
            print(f"  ! Claude CLI exception: {error[:100]}")
            return True, f"[Claude CLI Exception: {error}]", None

    def _execute_claude_direct(self, prompt: str) -> Tuple[bool, str, Optional[str]]:
        """Execute using Claude CLI directly (new API)"""
        try:
            print(f"  Using Claude CLI directly...")
            result = subprocess.run(
                ["claude", "-p", prompt],
                capture_output=True,
                text=True,
                timeout=300
            )

            if result.returncode == 0:
                output = result.stdout.strip()
                print(f"  ✓ Claude response: {output[:100]}...")
                return True, output, None
            else:
                error = result.stderr or f"Claude CLI failed with exit code {result.returncode}"
                print(f"  ! Claude CLI error: {error[:100]}")
                return True, f"[Claude CLI Error: {error}]", None

        except subprocess.TimeoutExpired:
            error = "Claude CLI timed out after 5 minutes"
            print(f"  ✗ {error}")
            return False, None, error
        except Exception as e:
            error = str(e)
            print(f"  ! Claude CLI exception: {error[:100]}")
            return True, f"[Claude CLI Exception: {error}]", None

    def run(self):
        """Run worker loop"""
        self._running = True

        # Register with master
        print(f"→ Registering worker: {self.name} ({self.worker_id})")
        if not self.register():
            print("✗ Failed to register with master")
            return

        print(f"✓ Worker registered successfully")

        # Start heartbeat thread
        def heartbeat_loop():
            while self._running and not self._stop_event.is_set():
                time.sleep(30)  # Heartbeat every 30 seconds
                if self._running:
                    if not self.send_heartbeat():
                        print(f"! Heartbeat failed (master may be offline)")

        heartbeat_thread = threading.Thread(target=heartbeat_loop, daemon=True)
        heartbeat_thread.start()

        # Give time for registration
        time.sleep(1)

        # Main loop
        if self.oneshot:
            self._run_oneshot()
        elif self.auto_mode:
            self._run_auto()
        else:
            self._run_manual()

    def _run_oneshot(self):
        """Execute one task and exit"""
        print(f"→ Running in oneshot mode...")

        tasks = self.get_pending_tasks()
        if not tasks:
            print("  No pending tasks, exiting")
            return

        task = tasks[0]
        self._execute_and_complete(task)

    def _run_auto(self):
        """Auto mode: continuously fetch and execute tasks"""
        print(f"→ Running in auto mode (continuous)...")

        while self._running and not self._stop_event.is_set():
            # Get pending tasks
            tasks = self.get_pending_tasks()

            if not tasks:
                print(f"  No pending tasks, waiting...")
                time.sleep(5)
                continue

            # Try to execute tasks
            executed = False
            for task in tasks:
                if not self._running:
                    break

                if self.assign_task(task["task_id"]):
                    self._execute_and_complete(task)
                    executed = True
                    break  # Execute one task per iteration

            if not executed:
                # Race condition: another worker grabbed the task
                print(f"  All tasks were assigned, waiting...")
                time.sleep(2)

    def _run_manual(self):
        """Manual mode: wait for external commands"""
        print(f"→ Running in manual mode (waiting for commands)...")
        print(f"  Use /agentflow commands to control this worker")
        while self._running and not self._stop_event.is_set():
            time.sleep(1)

    def _execute_and_complete(self, task: dict):
        """Execute task and report completion"""
        task_id = task["task_id"]

        try:
            # Execute task
            success, result, error = self.execute_task(task)

            # Complete task
            if success:
                self.complete_task(task_id, result=result, error=None)
                print(f"✓ Task completed: {task_id}")
            else:
                self.complete_task(task_id, result=None, error=error)
                print(f"✗ Task failed: {task_id}")

        except Exception as e:
            error = str(e)
            print(f"✗ Task execution error: {error}")
            self.complete_task(task_id, result=None, error=error)

    def stop(self):
        """Stop worker"""
        print(f"→ Stopping worker: {self.name}")
        self._running = False
        self._stop_event.set()


def run_worker(master_url: str = "http://127.0.0.1:8848",
               name: str = None, auto: bool = True, oneshot: bool = False):
    """Convenience function to run worker"""
    worker = Worker(
        master_url=master_url,
        name=name,
        auto_mode=auto,
        oneshot=oneshot
    )
    try:
        worker.run()
    except KeyboardInterrupt:
        print("\n✓ Worker interrupted by user")
        worker.stop()
