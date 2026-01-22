"""
Database module for AgentFlow
SQLite-based task and worker management
"""

import sqlite3
import json
import time
from datetime import datetime
from typing import List, Dict, Optional, Any
from dataclasses import dataclass, asdict


@dataclass
class Task:
    """Task data model"""
    task_id: str
    title: str
    description: str
    priority: str
    status: str = "pending"
    assigned_to: Optional[str] = None
    progress: int = 0
    result: Optional[str] = None
    error: Optional[str] = None
    created_at: str = None
    updated_at: str = None
    completed_at: Optional[str] = None

    def __post_init__(self):
        if self.created_at is None:
            self.created_at = datetime.now().isoformat()
        if self.updated_at is None:
            self.updated_at = datetime.now().isoformat()


@dataclass
class Worker:
    """Worker data model"""
    worker_id: str
    worker_name: str
    platform: str
    status: str = "online"
    last_heartbeat: str = None
    registered_at: str = None

    def __post_init__(self):
        if self.last_heartbeat is None:
            self.last_heartbeat = datetime.now().isoformat()
        if self.registered_at is None:
            self.registered_at = datetime.now().isoformat()


class Database:
    """SQLite database for AgentFlow"""

    def __init__(self, db_path: str = ".claude/cpds-manager/agentflow-python.db"):
        self.db_path = db_path
        self.conn = None
        self.connect()
        self.init_tables()

    def connect(self):
        """Connect to database"""
        import os
        os.makedirs(os.path.dirname(self.db_path), exist_ok=True)
        self.conn = sqlite3.connect(self.db_path, check_same_thread=False)
        self.conn.row_factory = sqlite3.Row

    def init_tables(self):
        """Initialize database tables"""
        cursor = self.conn.cursor()

        # Tasks table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS tasks (
                task_id TEXT PRIMARY KEY,
                title TEXT NOT NULL,
                description TEXT NOT NULL,
                priority TEXT DEFAULT 'medium',
                status TEXT DEFAULT 'pending',
                assigned_to TEXT,
                progress INTEGER DEFAULT 0,
                result TEXT,
                error TEXT,
                created_at TEXT,
                updated_at TEXT,
                completed_at TEXT
            )
        """)

        # Workers table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS workers (
                worker_id TEXT PRIMARY KEY,
                worker_name TEXT NOT NULL,
                platform TEXT NOT NULL,
                status TEXT DEFAULT 'online',
                last_heartbeat TEXT,
                registered_at TEXT
            )
        """)

        # Index for faster queries
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_tasks_status
            ON tasks(status)
        """)

        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_tasks_priority
            ON tasks(priority, created_at)
        """)

        self.conn.commit()

    # Task operations
    def create_task(self, task: Task) -> bool:
        """Create a new task"""
        try:
            cursor = self.conn.cursor()
            cursor.execute("""
                INSERT INTO tasks (
                    task_id, title, description, priority, status,
                    assigned_to, progress, result, error,
                    created_at, updated_at, completed_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                task.task_id, task.title, task.description, task.priority,
                task.status, task.assigned_to, task.progress, task.result,
                task.error, task.created_at, task.updated_at, task.completed_at
            ))
            self.conn.commit()
            return True
        except sqlite3.IntegrityError:
            return False

    def get_task(self, task_id: str) -> Optional[Task]:
        """Get task by ID"""
        cursor = self.conn.cursor()
        cursor.execute("SELECT * FROM tasks WHERE task_id = ?", (task_id,))
        row = cursor.fetchone()
        if row:
            return Task(**dict(row))
        return None

    def get_pending_tasks(self, limit: int = 10) -> List[Task]:
        """Get pending tasks ordered by priority"""
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT * FROM tasks
            WHERE status = 'pending'
            ORDER BY
                CASE priority
                    WHEN 'high' THEN 1
                    WHEN 'medium' THEN 2
                    WHEN 'low' THEN 3
                END,
                created_at ASC
            LIMIT ?
        """, (limit,))
        return [Task(**dict(row)) for row in cursor.fetchall()]

    def get_running_tasks(self) -> List[Task]:
        """Get running tasks"""
        cursor = self.conn.cursor()
        cursor.execute("SELECT * FROM tasks WHERE status = 'in_progress'")
        return [Task(**dict(row)) for row in cursor.fetchall()]

    def get_completed_tasks(self, limit: int = 100) -> List[Task]:
        """Get completed tasks"""
        cursor = self.conn.cursor()
        cursor.execute("""
            SELECT * FROM tasks
            WHERE status = 'completed'
            ORDER BY completed_at DESC
            LIMIT ?
        """, (limit,))
        return [Task(**dict(row)) for row in cursor.fetchall()]

    def update_task_status(self, task_id: str, status: str, assigned_to: str = None) -> bool:
        """Update task status"""
        try:
            cursor = self.conn.cursor()
            if assigned_to:
                cursor.execute("""
                    UPDATE tasks
                    SET status = ?, assigned_to = ?, updated_at = ?
                    WHERE task_id = ?
                """, (status, assigned_to, datetime.now().isoformat(), task_id))
            else:
                cursor.execute("""
                    UPDATE tasks
                    SET status = ?, updated_at = ?
                    WHERE task_id = ?
                """, (status, datetime.now().isoformat(), task_id))
            self.conn.commit()
            return cursor.rowcount > 0
        except Exception:
            return False

    def update_task_progress(self, task_id: str, progress: int, message: str = None) -> bool:
        """Update task progress"""
        try:
            cursor = self.conn.cursor()
            cursor.execute("""
                UPDATE tasks
                SET progress = ?, updated_at = ?
                WHERE task_id = ?
            """, (progress, datetime.now().isoformat(), task_id))
            self.conn.commit()
            return cursor.rowcount > 0
        except Exception:
            return False

    def complete_task(self, task_id: str, result: str = None, error: str = None) -> bool:
        """Complete task with result or error"""
        try:
            cursor = self.conn.cursor()
            status = "completed" if error is None else "failed"
            cursor.execute("""
                UPDATE tasks
                SET status = ?, result = ?, error = ?,
                    completed_at = ?, updated_at = ?
                WHERE task_id = ?
            """, (status, result, error, datetime.now().isoformat(),
                  datetime.now().isoformat(), task_id))
            self.conn.commit()
            return cursor.rowcount > 0
        except Exception:
            return False

    # Worker operations
    def register_worker(self, worker: Worker) -> bool:
        """Register a new worker"""
        try:
            cursor = self.conn.cursor()
            cursor.execute("""
                INSERT OR REPLACE INTO workers
                (worker_id, worker_name, platform, status, last_heartbeat, registered_at)
                VALUES (?, ?, ?, ?, ?, ?)
            """, (
                worker.worker_id, worker.worker_name, worker.platform,
                worker.status, worker.last_heartbeat, worker.registered_at
            ))
            self.conn.commit()
            return True
        except Exception:
            return False

    def get_worker(self, worker_id: str) -> Optional[Worker]:
        """Get worker by ID"""
        cursor = self.conn.cursor()
        cursor.execute("SELECT * FROM workers WHERE worker_id = ?", (worker_id,))
        row = cursor.fetchone()
        if row:
            return Worker(**dict(row))
        return None

    def get_online_workers(self) -> List[Worker]:
        """Get all online workers (heartbeat within 5 minutes)"""
        cursor = self.conn.cursor()
        five_min_ago = (datetime.now().timestamp() - 300) * 1000
        cursor.execute("""
            SELECT * FROM workers
            WHERE status = 'online'
            AND datetime(last_heartbeat) > datetime(?)
        """, (datetime.fromtimestamp(five_min_ago).isoformat(),))
        return [Worker(**dict(row)) for row in cursor.fetchall()]

    def update_worker_heartbeat(self, worker_id: str) -> bool:
        """Update worker heartbeat"""
        try:
            cursor = self.conn.cursor()
            cursor.execute("""
                UPDATE workers
                SET last_heartbeat = ?, status = 'online'
                WHERE worker_id = ?
            """, (datetime.now().isoformat(), worker_id))
            self.conn.commit()
            return cursor.rowcount > 0
        except Exception:
            return False

    def cleanup_offline_workers(self, timeout_seconds: int = 300):
        """Mark workers as offline if no heartbeat for timeout seconds"""
        cursor = self.conn.cursor()
        timeout_time = datetime.fromtimestamp(
            datetime.now().timestamp() - timeout_seconds
        ).isoformat()
        cursor.execute("""
            UPDATE workers
            SET status = 'offline'
            WHERE last_heartbeat < ?
        """, (timeout_time,))
        self.conn.commit()

    # System status
    def get_system_status(self) -> Dict[str, int]:
        """Get system status statistics"""
        cursor = self.conn.cursor()

        cursor.execute("SELECT COUNT(*) FROM tasks WHERE status = 'pending'")
        pending = cursor.fetchone()[0]

        cursor.execute("SELECT COUNT(*) FROM tasks WHERE status = 'in_progress'")
        in_progress = cursor.fetchone()[0]

        cursor.execute("SELECT COUNT(*) FROM tasks WHERE status = 'completed'")
        completed = cursor.fetchone()[0]

        cursor.execute("SELECT COUNT(*) FROM workers WHERE status = 'online'")
        online_workers = cursor.fetchone()[0]

        return {
            "pending_tasks": pending,
            "in_progress_tasks": in_progress,
            "completed_tasks": completed,
            "online_workers": online_workers
        }

    def close(self):
        """Close database connection"""
        if self.conn:
            self.conn.close()
