"""
Master server for AgentFlow (Python version)
RESTful API for task management and worker coordination
"""

from flask import Flask, request, jsonify
import threading
import time
import uuid
from typing import Dict, Any
from .database import Database, Task, Worker


class Master:
    """Master server for task coordination"""

    def __init__(self, host: str = "0.0.0.0", port: int = 8848,
                 db_path: str = ".claude/cpds-manager/agentflow-python.db",
                 auto_shutdown: bool = False):
        self.host = host
        self.port = port
        self.db = Database(db_path)
        self.auto_shutdown = auto_shutdown
        self.app = Flask(__name__)
        self.setup_routes()
        self.start_time = time.time()
        self._shutdown_event = threading.Event()

    def setup_routes(self):
        """Setup Flask routes"""

        @self.app.route("/api/health", methods=["GET"])
        def health_check():
            """Health check endpoint"""
            return jsonify({
                "status": "healthy",
                "mode": "standalone" if self.auto_shutdown else "cloud"
            })

        @self.app.route("/api/status", methods=["GET"])
        def get_status():
            """Get system status"""
            status = self.db.get_system_status()
            uptime = int(time.time() - self.start_time)
            return jsonify({
                "success": True,
                "data": {
                    **status,
                    "uptime_seconds": uptime
                }
            })

        # Worker management
        @self.app.route("/api/workers/register", methods=["POST"])
        def register_worker():
            """Register a new worker"""
            data = request.get_json()

            required = ["worker_id", "worker_name", "platform"]
            if not all(k in data for k in required):
                return jsonify({
                    "success": False,
                    "error": "Missing required fields"
                }), 400

            worker = Worker(
                worker_id=data["worker_id"],
                worker_name=data["worker_name"],
                platform=data["platform"]
            )

            if self.db.register_worker(worker):
                return jsonify({
                    "success": True,
                    "data": {
                        "worker_id": worker.worker_id,
                        "message": "Worker registered successfully"
                    }
                })
            else:
                return jsonify({
                    "success": False,
                    "error": "Failed to register worker"
                }), 500

        @self.app.route("/api/workers/heartbeat", methods=["POST"])
        def worker_heartbeat():
            """Update worker heartbeat"""
            data = request.get_json()
            worker_id = data.get("worker_id")

            if not worker_id:
                return jsonify({
                    "success": False,
                    "error": "worker_id is required"
                }), 400

            if self.db.update_worker_heartbeat(worker_id):
                return jsonify({
                    "success": True,
                    "data": {"message": "Heartbeat received"}
                })
            else:
                return jsonify({
                    "success": False,
                    "error": "Failed to update heartbeat"
                }), 500

        @self.app.route("/api/workers", methods=["GET"])
        def get_workers():
            """Get all online workers"""
            workers = self.db.get_online_workers()
            return jsonify({
                "success": True,
                "data": {
                    "workers": [w.__dict__ for w in workers],
                    "count": len(workers)
                }
            })

        # Task management
        @self.app.route("/api/tasks/create", methods=["POST"])
        def create_task():
            """Create a new task"""
            data = request.get_json()

            task_id = data.get("task_id") or f"TASK-{uuid.uuid4().hex[:8].upper()}"
            title = data.get("title", "Untitled Task")
            description = data.get("description", "")
            priority = data.get("priority", "medium")

            task = Task(
                task_id=task_id,
                title=title,
                description=description,
                priority=priority
            )

            if self.db.create_task(task):
                return jsonify({
                    "success": True,
                    "data": {
                        "task_id": task.task_id,
                        "message": "Task created successfully"
                    }
                })
            else:
                return jsonify({
                    "success": False,
                    "error": "Failed to create task"
                }), 500

        @self.app.route("/api/tasks/pending", methods=["GET"])
        def get_pending_tasks():
            """Get pending tasks"""
            limit = request.args.get("limit", 10, type=int)
            tasks = self.db.get_pending_tasks(limit)
            return jsonify({
                "success": True,
                "data": {
                    "tasks": [t.__dict__ for t in tasks],
                    "count": len(tasks)
                }
            })

        @self.app.route("/api/tasks/running", methods=["GET"])
        def get_running_tasks():
            """Get running tasks"""
            tasks = self.db.get_running_tasks()
            return jsonify({
                "success": True,
                "data": {
                    "tasks": [t.__dict__ for t in tasks],
                    "count": len(tasks)
                }
            })

        @self.app.route("/api/tasks/completed", methods=["GET"])
        def get_completed_tasks():
            """Get completed tasks"""
            limit = request.args.get("limit", 100, type=int)
            tasks = self.db.get_completed_tasks(limit)
            return jsonify({
                "success": True,
                "data": {
                    "tasks": [t.__dict__ for t in tasks],
                    "count": len(tasks)
                }
            })

        @self.app.route("/api/tasks/<task_id>", methods=["GET"])
        def get_task(task_id: str):
            """Get task by ID"""
            task = self.db.get_task(task_id)
            if task:
                return jsonify({
                    "success": True,
                    "data": task.__dict__
                })
            else:
                return jsonify({
                    "success": False,
                    "error": "Task not found"
                }), 404

        @self.app.route("/api/tasks/assign", methods=["POST"])
        def assign_task():
            """Assign task to worker"""
            data = request.get_json()
            task_id = data.get("task_id")
            worker_id = data.get("worker_id")

            if not task_id or not worker_id:
                return jsonify({
                    "success": False,
                    "error": "task_id and worker_id are required"
                }), 400

            # Get task first to verify it's pending
            task = self.db.get_task(task_id)
            if not task or task.status != "pending":
                return jsonify({
                    "success": False,
                    "error": "Task not found or not pending"
                }), 404

            if self.db.update_task_status(task_id, "in_progress", worker_id):
                return jsonify({
                    "success": True,
                    "data": {
                        "task_id": task_id,
                        "worker_id": worker_id,
                        "message": "Task assigned successfully"
                    }
                })
            else:
                return jsonify({
                    "success": False,
                    "error": "Failed to assign task"
                }), 500

        @self.app.route("/api/tasks/progress", methods=["POST"])
        def update_progress():
            """Update task progress"""
            data = request.get_json()
            task_id = data.get("task_id")
            progress = data.get("progress", 0)
            message = data.get("message", "")

            if not task_id:
                return jsonify({
                    "success": False,
                    "error": "task_id is required"
                }), 400

            if self.db.update_task_progress(task_id, progress, message):
                return jsonify({
                    "success": True,
                    "data": {"message": "Progress updated"}
                })
            else:
                return jsonify({
                    "success": False,
                    "error": "Failed to update progress"
                }), 500

        @self.app.route("/api/tasks/complete", methods=["POST"])
        def complete_task():
            """Complete task"""
            data = request.get_json()
            task_id = data.get("task_id")
            result = data.get("result")
            error = data.get("error")

            if not task_id:
                return jsonify({
                    "success": False,
                    "error": "task_id is required"
                }), 400

            if self.db.complete_task(task_id, result, error):
                return jsonify({
                    "success": True,
                    "data": {"message": "Task completed"}
                })
            else:
                return jsonify({
                    "success": False,
                    "error": "Failed to complete task"
                }), 500

    def run(self, debug: bool = False):
        """Run the master server"""

        # Start heartbeat cleanup thread
        def cleanup_loop():
            while not self._shutdown_event.is_set():
                time.sleep(60)  # Check every minute
                if not self._shutdown_event.is_set():
                    self.db.cleanup_offline_workers()

                    # Auto-shutdown check
                    if self.auto_shutdown:
                        status = self.db.get_system_status()
                        if status["pending_tasks"] == 0 and status["in_progress_tasks"] == 0:
                            print("✓ All tasks completed, shutting down...")
                            self._shutdown_event.set()
                            import os
                            os._exit(0)

        cleanup_thread = threading.Thread(target=cleanup_loop, daemon=True)
        cleanup_thread.start()

        # Run Flask app
        print(f"✓ Master server starting on {self.host}:{self.port}")
        print(f"  Mode: {'standalone (auto-shutdown)' if self.auto_shutdown else 'cloud'}")
        print(f"  Database: {self.db.db_path}")
        self.app.run(host=self.host, port=self.port, debug=debug, threaded=True)

    def shutdown(self):
        """Shutdown the server"""
        self._shutdown_event.set()
        self.db.close()


def run_master(host: str = "0.0.0.0", port: int = 8848,
               db_path: str = ".claude/cpds-manager/agentflow-python.db",
               auto_shutdown: bool = False):
    """Convenience function to run master"""
    master = Master(host=host, port=port, db_path=db_path, auto_shutdown=auto_shutdown)
    master.run()
