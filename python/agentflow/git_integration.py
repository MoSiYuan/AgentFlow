"""
AgentFlow Git Integration Module

Provides Git integration, file boundary control, and file locking
for multi-agent collaborative development.

Features:
- Git branch management per agent
- File boundary enforcement
- Read/write file locking
- Conflict detection and resolution
"""

import os
import json
import subprocess
import hashlib
from pathlib import Path
from typing import List, Dict, Optional, Set
from dataclasses import dataclass, asdict
from datetime import datetime
import threading


@dataclass
class FileBoundary:
    """File boundary definition for an agent"""
    file_pattern: str      # e.g., "src/controllers/*.go"
    access_type: str       # "exclusive" | "shared" | "readonly"
    line_range: str = ""   # e.g., "1-100" (optional)
    description: str = ""


@dataclass
class FileLock:
    """File lock information"""
    file_path: str
    agent_id: str
    lock_type: str         # "read" | "write"
    acquired_at: datetime
    status: str           # "active" | "released"


@dataclass
class GitConflict:
    """Git conflict information"""
    conflict_id: str
    task_id: str
    agent_id: str
    conflict_type: str    # "file_locked" | "boundary_overlap" | "merge_conflict"
    file_paths: List[str]
    severity: str         # "low" | "medium" | "high" | "critical"
    description: str
    status: str          # "pending" | "resolving" | "resolved"


class GitFileBoundaryManager:
    """
    Manages file boundaries and permissions for agents

    Ensures each agent can only operate within its designated file boundaries.
    """

    def __init__(self, config_path: str = ".agentflow/boundaries.json"):
        self.config_path = Path(config_path)
        self.boundaries: Dict[str, List[FileBoundary]] = {}
        self.locks: Dict[str, FileLock] = {}
        self.lock_mutex = threading.Lock()
        self._load_boundaries()

    def _load_boundaries(self):
        """Load file boundaries from config file"""
        if self.config_path.exists():
            with open(self.config_path, 'r') as f:
                data = json.load(f)
                for agent_id, boundaries in data.items():
                    self.boundaries[agent_id] = [
                        FileBoundary(**b) for b in boundaries
                    ]
        else:
            # Create default configuration
            self._create_default_config()

    def _create_default_config(self):
        """Create default boundary configuration"""
        default_boundaries = {
            "frontend": [
                FileBoundary(
                    file_pattern="src/frontend/**/*",
                    access_type="exclusive",
                    description="Frontend agent can exclusively modify frontend files"
                ),
                FileBoundary(
                    file_pattern="src/api/**/*",
                    access_type="readonly",
                    description="Frontend agent can read API files"
                )
            ],
            "backend": [
                FileBoundary(
                    file_pattern="src/backend/**/*",
                    access_type="exclusive",
                    description="Backend agent can exclusively modify backend files"
                ),
                FileBoundary(
                    file_pattern="src/api/**/*",
                    access_type="shared",
                    description="Backend agent shares API files"
                )
            ],
            "database": [
                FileBoundary(
                    file_pattern="src/database/**/*",
                    access_type="exclusive",
                    description="Database agent exclusively manages database files"
                )
            ]
        }

        self.boundaries = default_boundaries
        self._save_boundaries()

    def _save_boundaries(self):
        """Save boundaries to config file"""
        self.config_path.parent.mkdir(parents=True, exist_ok=True)

        data = {}
        for agent_id, boundaries in self.boundaries.items():
            data[agent_id] = [asdict(b) for b in boundaries]

        with open(self.config_path, 'w') as f:
            json.dump(data, f, indent=2)

    def get_agent_boundaries(self, agent_id: str) -> List[FileBoundary]:
        """Get file boundaries for an agent"""
        return self.boundaries.get(agent_id, [])

    def can_access_file(self, agent_id: str, file_path: str,
                        access_type: str = "write") -> bool:
        """
        Check if an agent can access a file

        Args:
            agent_id: Agent identifier
            file_path: File path to check
            access_type: Type of access ("read" | "write")

        Returns:
            True if agent can access the file
        """
        boundaries = self.get_agent_boundaries(agent_id)
        if not boundaries:
            return False  # No boundaries defined = no access

        file_path = Path(file_path)

        for boundary in boundaries:
            # Check if file matches pattern
            pattern = Path(boundary.file_pattern)

            # Simple pattern matching (can be enhanced with fnmatch)
            if self._matches_pattern(file_path, pattern):
                # Check access type
                if boundary.access_type == "exclusive":
                    return True
                elif boundary.access_type == "shared":
                    return True
                elif boundary.access_type == "readonly" and access_type == "read":
                    return True

        return False

    def _matches_pattern(self, file_path: Path, pattern: Path) -> bool:
        """Check if file path matches pattern"""
        try:
            file_path = file_path.resolve()
            pattern_parts = pattern.parts

            # Handle wildcards
            for i, part in enumerate(pattern_parts):
                if part == "**":
                    return True  # Matches any subdirectory
                elif part.startswith("*") or part.endswith("*"):
                    # Simple wildcard matching
                    if len(file_path.parts) > i:
                        if part.replace("*", "") in file_path.parts[i]:
                            continue
                elif len(file_path.parts) > i and file_path.parts[i] == part:
                    continue
                else:
                    return False

            return True
        except (ValueError, IndexError):
            return False

    def acquire_lock(self, agent_id: str, file_path: str,
                    lock_type: str = "write") -> bool:
        """
        Acquire a file lock

        Args:
            agent_id: Agent requesting the lock
            file_path: File to lock
            lock_type: Type of lock ("read" | "write")

        Returns:
            True if lock acquired successfully
        """
        with self.lock_mutex:
            lock_key = f"{file_path}:{lock_type}"

            # Check if lock already exists
            if lock_key in self.locks:
                existing_lock = self.locks[lock_key]
                if existing_lock.status == "active":
                    # Write locks are exclusive
                    if lock_type == "write" or existing_lock.lock_type == "write":
                        return False
                    # Multiple read locks allowed

            # Create lock
            lock = FileLock(
                file_path=file_path,
                agent_id=agent_id,
                lock_type=lock_type,
                acquired_at=datetime.now(),
                status="active"
            )

            self.locks[lock_key] = lock
            return True

    def release_lock(self, agent_id: str, file_path: str,
                    lock_type: str = "write") -> bool:
        """Release a file lock"""
        with self.lock_mutex:
            lock_key = f"{file_path}:{lock_type}"

            if lock_key in self.locks:
                lock = self.locks[lock_key]
                if lock.agent_id == agent_id:
                    lock.status = "released"
                    del self.locks[lock_key]
                    return True

            return False

    def get_active_locks(self, file_path: Optional[str] = None) -> List[FileLock]:
        """Get all active locks"""
        with self.lock_mutex:
            if file_path:
                return [
                    lock for lock in self.locks.values()
                    if lock.file_path == file_path and lock.status == "active"
                ]
            return [
                lock for lock in self.locks.values()
                if lock.status == "active"
            ]

    def check_conflicts(self, agent_id: str, file_paths: List[str]) -> List[GitConflict]:
        """
        Check for potential conflicts with existing locks

        Args:
            agent_id: Agent requesting access
            file_paths: List of files to access

        Returns:
            List of detected conflicts
        """
        conflicts = []

        for file_path in file_paths:
            active_locks = self.get_active_locks(file_path)

            for lock in active_locks:
                if lock.agent_id != agent_id:
                    # Write-write conflict
                    if lock.lock_type == "write":
                        conflicts.append(GitConflict(
                            conflict_id=self._generate_conflict_id(),
                            task_id="",
                            agent_id=agent_id,
                            conflict_type="file_locked",
                            file_paths=[file_path],
                            severity="high",
                            description=f"File {file_path} is locked by {lock.agent_id}",
                            status="pending"
                        ))

        return conflicts

    def _generate_conflict_id(self) -> str:
        """Generate unique conflict ID"""
        return hashlib.md5(
            f"{datetime.now().isoformat()}".encode()
        ).hexdigest()[:8]


class GitBranchManager:
    """
    Manages Git branches for agents

    Each agent gets its own branch for isolated development.
    """

    def __init__(self, repo_path: str = "."):
        self.repo_path = Path(repo_path)
        if not (self.repo_path / ".git").exists():
            raise ValueError(f"Not a Git repository: {repo_path}")

    def create_agent_branch(self, agent_id: str, task_id: str) -> str:
        """
        Create a new branch for an agent task

        Args:
            agent_id: Agent identifier
            task_id: Task identifier

        Returns:
            Branch name
        """
        branch_name = f"agent-{agent_id}/task-{task_id}"

        # Create branch from main/master
        try:
            subprocess.run(
                ["git", "checkout", "-b", branch_name],
                cwd=self.repo_path,
                check=True,
                capture_output=True
            )
            return branch_name
        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Failed to create branch: {e.stderr.decode()}")

    def delete_agent_branch(self, branch_name: str, force: bool = False):
        """Delete an agent branch"""
        try:
            args = ["git", "branch", "-d", branch_name]
            if force:
                args[2] = "-D"

            subprocess.run(
                args,
                cwd=self.repo_path,
                check=True,
                capture_output=True
            )
        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Failed to delete branch: {e.stderr.decode()}")

    def switch_branch(self, branch_name: str):
        """Switch to a branch"""
        try:
            subprocess.run(
                ["git", "checkout", branch_name],
                cwd=self.repo_path,
                check=True,
                capture_output=True
            )
        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Failed to switch branch: {e.stderr.decode()}")

    def commit_changes(self, agent_id: str, message: str) -> str:
        """
        Commit changes on agent's branch

        Args:
            agent_id: Agent identifier
            message: Commit message

        Returns:
            Commit hash
        """
        # Add changes
        try:
            subprocess.run(
                ["git", "add", "."],
                cwd=self.repo_path,
                check=True,
                capture_output=True
            )
        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Failed to add files: {e.stderr.decode()}")

        # Commit
        commit_message = f"[{agent_id}] {message}"
        try:
            result = subprocess.run(
                ["git", "commit", "-m", commit_message],
                cwd=self.repo_path,
                check=True,
                capture_output=True,
                text=True
            )

            # Extract commit hash
            output = result.stdout
            for line in output.split('\n'):
                if line.startswith('['):
                    # Extract hash from bracket notation
                    parts = line.split(' ')
                    if len(parts) > 1:
                        return parts[1].strip(']')

            # Alternative: get latest commit hash
            result = subprocess.run(
                ["git", "rev-parse", "HEAD"],
                cwd=self.repo_path,
                check=True,
                capture_output=True,
                text=True
            )
            return result.stdout.strip()

        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Failed to commit: {e.stderr.decode()}")

    def merge_to_main(self, branch_name: str, strategy: str = "merge") -> bool:
        """
        Merge agent branch to main

        Args:
            branch_name: Branch to merge
            strategy: Merge strategy ("merge" | "squash" | "rebase")

        Returns:
            True if merge successful, False otherwise
        """
        # Switch to main first
        try:
            self.switch_branch("main")
        except RuntimeError:
            try:
                self.switch_branch("master")
            except RuntimeError:
                raise RuntimeError("No main or master branch found")

        # Merge
        try:
            if strategy == "merge":
                subprocess.run(
                    ["git", "merge", branch_name],
                    cwd=self.repo_path,
                    check=True,
                    capture_output=True
                )
            elif strategy == "squash":
                subprocess.run(
                    ["git", "merge", "--squash", branch_name],
                    cwd=self.repo_path,
                    check=True,
                    capture_output=True
                )
                subprocess.run(
                    ["git", "commit", "-m", f"Squashed merge of {branch_name}"],
                    cwd=self.repo_path,
                    check=True,
                    capture_output=True
                )
            elif strategy == "rebase":
                subprocess.run(
                    ["git", "rebase", branch_name],
                    cwd=self.repo_path,
                    check=True,
                    capture_output=True
                )

            return True

        except subprocess.CalledProcessError as e:
            # Merge conflict
            return False

    def get_merge_conflicts(self) -> List[str]:
        """Get list of files with merge conflicts"""
        try:
            result = subprocess.run(
                ["git", "diff", "--name-only", "--diff-filter=U"],
                cwd=self.repo_path,
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                conflicts = result.stdout.strip().split('\n')
                return [c for c in conflicts if c]

            return []

        except subprocess.CalledProcessError:
            return []


class GitIntegrationManager:
    """
    High-level Git integration manager for AgentFlow

    Combines file boundaries, locking, and branch management.
    """

    def __init__(self, repo_path: str = ".",
                 boundary_config: str = ".agentflow/boundaries.json"):
        self.boundary_manager = GitFileBoundaryManager(boundary_config)
        self.branch_manager = GitBranchManager(repo_path)

    def create_agent_task(self, agent_id: str, task_id: str,
                         description: str) -> dict:
        """
        Create a new task for an agent with Git integration

        Args:
            agent_id: Agent identifier
            task_id: Task identifier
            description: Task description

        Returns:
            Task information dictionary
        """
        # Create agent branch
        branch_name = self.branch_manager.create_agent_branch(agent_id, task_id)

        return {
            "agent_id": agent_id,
            "task_id": task_id,
            "branch": branch_name,
            "description": description,
            "status": "in_progress",
            "boundaries": [asdict(b) for b in self.boundary_manager.get_agent_boundaries(agent_id)]
        }

    def verify_file_access(self, agent_id: str, file_path: str,
                          access_type: str = "write") -> tuple[bool, Optional[str]]:
        """
        Verify if agent can access a file

        Returns:
            (allowed, reason) tuple
        """
        if not self.boundary_manager.can_access_file(agent_id, file_path, access_type):
            return False, f"Agent {agent_id} not authorized to access {file_path}"

        # Check locks
        conflicts = self.boundary_manager.check_conflicts(agent_id, [file_path])
        if conflicts and access_type == "write":
            return False, f"File locked: {conflicts[0].description}"

        return True, None

    def safe_file_operation(self, agent_id: str, file_path: str,
                           operation: callable, *args, **kwargs):
        """
        Perform a safe file operation with locking and boundary checks

        Args:
            agent_id: Agent identifier
            file_path: File path
            operation: Function to execute
            *args, **kwargs: Arguments for operation

        Returns:
            Operation result
        """
        # Check boundaries
        allowed, reason = self.verify_file_access(agent_id, file_path, "write")
        if not allowed:
            raise PermissionError(reason)

        # Acquire lock
        if not self.boundary_manager.acquire_lock(agent_id, file_path, "write"):
            raise PermissionError(f"Could not acquire lock for {file_path}")

        try:
            # Perform operation
            result = operation(file_path, *args, **kwargs)

            # Commit changes
            self.branch_manager.commit_changes(
                agent_id,
                f"Modified {file_path}"
            )

            return result

        finally:
            # Release lock
            self.boundary_manager.release_lock(agent_id, file_path, "write")

    def complete_agent_task(self, agent_id: str, task_id: str,
                           merge_strategy: str = "squash") -> dict:
        """
        Complete an agent task and merge to main

        Args:
            agent_id: Agent identifier
            task_id: Task identifier
            merge_strategy: Merge strategy

        Returns:
            Result dictionary
        """
        branch_name = f"agent-{agent_id}/task-{task_id}"

        # Check for conflicts
        merge_success = self.branch_manager.merge_to_main(branch_name, merge_strategy)

        if not merge_success:
            conflicts = self.branch_manager.get_merge_conflicts()
            return {
                "status": "conflict",
                "branch": branch_name,
                "conflicts": conflicts,
                "message": f"Merge conflicts detected in {len(conflicts)} files"
            }

        # Clean up
        self.branch_manager.delete_agent_branch(branch_name, force=True)

        return {
            "status": "completed",
            "branch": branch_name,
            "message": f"Task {task_id} completed and merged successfully"
        }
