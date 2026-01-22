package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	id            string
	db            *database.Database
	masterURL     string
	groupName     string
	workerType    string
	client        *http.Client
	httpExecutor  *HTTPExecutor
	claudeExecutor *ClaudeExecutor
	logger        *logrus.Logger
	stopCh        chan struct{}
}

type Config struct {
	ID        string
	MasterURL string
	DBPath    string
	GroupName string
}

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	GroupName   string `json:"group_name"`
}

type LockResponse struct {
	Status string `json:"status"`
}

type PendingTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

func New(cfg *Config) (*Worker, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetOutput(os.Stdout)

	// Detect group name if not provided
	groupName := cfg.GroupName
	if groupName == "" {
		groupName = detectGroup()
	}

	// Setup database for local mode
	var db *database.Database
	if cfg.DBPath != "" {
		var err error
		db, err = database.NewDatabase(cfg.DBPath)
		if err != nil {
			return nil, fmt.Errorf("创建数据库失败: %w", err)
		}
	}

	workerID := cfg.ID
	if workerID == "" {
		workerID = fmt.Sprintf("%s-%d", groupName, time.Now().Unix())
	}

	workerType := "remote"
	if db != nil {
		workerType = "local"
	}

	// Initialize executors with priority order:
	// 1. HTTP Executor (fastest, if Claude server available)
	// 2. Claude CLI Executor (direct claudecli calls)
	// 3. Local execution (shell commands fallback)
	httpExecutor := NewHTTPExecutor("", workerID, logger)
	claudeExecutor := NewClaudeExecutor(logger)

	return &Worker{
		id:             workerID,
		db:             db,
		masterURL:      cfg.MasterURL,
		groupName:      groupName,
		workerType:     workerType,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		httpExecutor:   httpExecutor,
		claudeExecutor: claudeExecutor,
		logger:         logger,
		stopCh:         make(chan struct{}),
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	// Register with master
	if err := w.register(); err != nil {
		return fmt.Errorf("注册失败: %w", err)
	}

	// Start heartbeat
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Start task polling
	pollTicker := time.NewTicker(5 * time.Second)
	defer pollTicker.Stop()

	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"group":     w.groupName,
		"type":      w.workerType,
	}).Info("Worker 已启动")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("收到停止信号")
			return nil

		case <-w.stopCh:
			w.logger.Info("收到停止信号")
			return nil

		case <-heartbeatTicker.C:
			if err := w.sendHeartbeat(); err != nil {
				w.logger.WithError(err).Warn("发送心跳失败")
			}

		case <-pollTicker.C:
			if err := w.pollAndExecute(); err != nil {
				w.logger.WithError(err).Error("任务执行失败")
			}
		}
	}
}

func (w *Worker) Stop() {
	close(w.stopCh)
}

// RunOneShot executes one task and exits (one-shot mode)
func (w *Worker) RunOneShot(ctx context.Context) (string, error) {
	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"group":     w.groupName,
		"type":      w.workerType,
	}).Info("Starting one-shot worker")

	// Register with master
	if err := w.register(); err != nil {
		return "", fmt.Errorf("注册失败: %w", err)
	}

	// Start heartbeat in background
	heartbeatCtx, heartbeatCancel := context.WithCancel(ctx)
	defer heartbeatCancel()

	go w.heartbeatLoop(heartbeatCtx)

	// Give some time for registration
	time.Sleep(1 * time.Second)

	// Try to get and execute a task
	return w.executeOneTask(ctx)
}

// heartbeatLoop runs heartbeat until context is cancelled
func (w *Worker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.sendHeartbeat(); err != nil {
				w.logger.WithError(err).Warn("发送心跳失败")
			}
		}
	}
}

// executeOneTask tries to fetch and execute one task
func (w *Worker) executeOneTask(ctx context.Context) (string, error) {
	// Local mode: read from database
	if w.db != nil {
		return w.executeOneLocalTask(ctx)
	}

	// Remote mode: fetch from master API
	return w.executeOneRemoteTask(ctx)
}

// executeOneLocalTask executes one task from local database
func (w *Worker) executeOneLocalTask(ctx context.Context) (string, error) {
	tasks, err := w.db.ListTasks("pending", w.groupName)
	if err != nil {
		return "", fmt.Errorf("获取任务失败: %w", err)
	}

	if len(tasks) == 0 {
		w.logger.Info("No pending tasks")
		return "", nil
	}

	// Try to lock first task
	task := tasks[0]
	locked, err := w.db.LockTask(task.ID, w.id)
	if err != nil {
		return "", fmt.Errorf("锁定任务失败: %w", err)
	}

	if !locked {
		w.logger.Info("No available tasks (all locked)")
		return "", nil
	}

	w.logger.WithField("task_id", task.ID).Info("开始执行任务 (OneShot)")

	// Execute task
	result, err := w.executeTask(&task)

	// Complete or fail task
	if err != nil {
		if failErr := w.db.FailTask(task.ID, w.id, err.Error()); failErr != nil {
			w.logger.WithError(failErr).Error("标记任务失败错误")
		}
		return "", err
	}

	if completeErr := w.db.CompleteTask(task.ID, w.id, result); completeErr != nil {
		w.logger.WithError(completeErr).Error("完成任务错误")
	}

	w.logger.WithField("task_id", task.ID).Info("任务已完成 (OneShot)")
	return result, nil
}

// executeOneRemoteTask executes one task from remote master
func (w *Worker) executeOneRemoteTask(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/api/v1/tasks/pending?group=%s", w.masterURL, w.groupName)

	resp, err := w.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取任务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		w.logger.Info("No pending tasks")
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取任务失败: %s", resp.Status)
	}

	var pendingResp PendingTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&pendingResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(pendingResp.Tasks) == 0 {
		w.logger.Info("No pending tasks")
		return "", nil
	}

	// Try to lock first task
	task := pendingResp.Tasks[0]
	lockURL := fmt.Sprintf("%s/api/v1/tasks/%s/lock", w.masterURL, task.ID)

	lockData := map[string]string{"worker_id": w.id}
	lockBody, _ := json.Marshal(lockData)

	lockResp, err := http.Post(lockURL, "application/json", bytes.NewBuffer(lockBody))
	if err != nil {
		return "", fmt.Errorf("锁定任务失败: %w", err)
	}
	defer lockResp.Body.Close()

	if lockResp.StatusCode == http.StatusConflict {
		w.logger.Info("No available tasks (all locked)")
		return "", nil
	}

	if lockResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("锁定任务失败: %s", lockResp.Status)
	}

	w.logger.WithField("task_id", task.ID).Info("开始执行任务 (OneShot)")

	// Fetch task details
	taskDetail, err := w.fetchTaskDetail(task.ID)
	if err != nil {
		w.unlockTask(task.ID)
		return "", fmt.Errorf("获取任务详情失败: %w", err)
	}

	// Execute task
	result, err := w.executeTaskDetail(taskDetail)

	// Complete or fail task
	if err != nil {
		w.failTask(task.ID, err.Error())
		return "", fmt.Errorf("任务执行失败: %w", err)
	}

	w.completeTask(task.ID, result)

	w.logger.WithField("task_id", task.ID).Info("任务已完成 (OneShot)")
	return result, nil
}

func (w *Worker) register() error {
	// For local workers, register directly in database
	if w.db != nil {
		capabilities := w.getCapabilities()
		if err := w.db.RegisterWorker(w.id, w.groupName, w.workerType, capabilities); err != nil {
			return err
		}
	}

	// Send heartbeat to register
	return w.sendHeartbeat()
}

func (w *Worker) sendHeartbeat() error {
	url := fmt.Sprintf("%s/api/v1/workers/%s/heartbeat", w.masterURL, w.id)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("心跳失败: %s", resp.Status)
	}

	return nil
}

func (w *Worker) pollAndExecute() error {
	// Local mode: read directly from database
	if w.db != nil {
		return w.executeLocalTask()
	}

	// Remote mode: fetch from master API
	return w.executeRemoteTask()
}

func (w *Worker) executeLocalTask() error {
	tasks, err := w.db.ListTasks("pending", w.groupName)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	// Try to lock first task
	task := tasks[0]
	locked, err := w.db.LockTask(task.ID, w.id)
	if err != nil {
		return fmt.Errorf("锁定任务失败: %w", err)
	}

	if !locked {
		// Task already locked by another worker
		return nil
	}

	w.logger.WithField("task_id", task.ID).Info("开始执行任务")

	// Execute task
	result, err := w.executeTask(&task)

	// Complete or fail task
	if err != nil {
		if failErr := w.db.FailTask(task.ID, w.id, err.Error()); failErr != nil {
			w.logger.WithError(failErr).Error("标记任务失败错误")
		}
		return err
	}

	if completeErr := w.db.CompleteTask(task.ID, w.id, result); completeErr != nil {
		w.logger.WithError(completeErr).Error("完成任务错误")
	}

	w.logger.WithField("task_id", task.ID).Info("任务已完成")
	return nil
}

func (w *Worker) executeRemoteTask() error {
	url := fmt.Sprintf("%s/api/v1/tasks/pending?group=%s", w.masterURL, w.groupName)

	resp, err := w.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// No tasks available
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取任务失败: %s", resp.Status)
	}

	var pendingResp PendingTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&pendingResp); err != nil {
		return err
	}

	if len(pendingResp.Tasks) == 0 {
		return nil
	}

	// Try to lock first task
	task := pendingResp.Tasks[0]
	lockURL := fmt.Sprintf("%s/api/v1/tasks/%s/lock", w.masterURL, task.ID)

	lockData := map[string]string{"worker_id": w.id}
	lockBody, _ := json.Marshal(lockData)

	lockResp, err := http.Post(lockURL, "application/json", bytes.NewBuffer(lockBody))
	if err != nil {
		return err
	}
	defer lockResp.Body.Close()

	if lockResp.StatusCode == http.StatusConflict {
		// Task already locked
		return nil
	}

	if lockResp.StatusCode != http.StatusOK {
		return fmt.Errorf("锁定任务失败: %s", lockResp.Status)
	}

	w.logger.WithField("task_id", task.ID).Info("开始执行任务")

	// Fetch task details
	taskDetail, err := w.fetchTaskDetail(task.ID)
	if err != nil {
		w.unlockTask(task.ID)
		return err
	}

	// Execute task
	result, err := w.executeTaskDetail(taskDetail)

	// Complete or fail task
	if err != nil {
		w.failTask(task.ID, err.Error())
		return err
	}

	w.completeTask(task.ID, result)

	w.logger.WithField("task_id", task.ID).Info("任务已完成")
	return nil
}

func (w *Worker) fetchTaskDetail(taskID string) (*Task, error) {
	url := fmt.Sprintf("%s/api/v1/tasks/%s", w.masterURL, taskID)

	resp, err := w.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取任务详情失败: %s", resp.Status)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (w *Worker) unlockTask(taskID string) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/unlock", w.masterURL, taskID)

	data := map[string]string{"worker_id": w.id}
	body, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (w *Worker) completeTask(taskID, result string) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/complete", w.masterURL, taskID)

	data := map[string]string{
		"worker_id": w.id,
		"result":    result,
	}
	body, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (w *Worker) failTask(taskID, errorMsg string) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/fail", w.masterURL, taskID)

	data := map[string]string{
		"worker_id": w.id,
		"error":     errorMsg,
	}
	body, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (w *Worker) executeTask(task *database.Task) (string, error) {
	// Convert database.Task to worker.Task
	workerTask := Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		GroupName:   task.GroupName,
	}

	return w.executeTaskWithPriority(workerTask)
}

func (w *Worker) executeTaskDetail(task *Task) (string, error) {
	return w.executeTaskWithPriority(*task)
}

func (w *Worker) executeShellCommand(command string) (string, error) {
	w.logger.WithField("command", command).Info("执行命令")

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("命令执行失败: %w: %s", err, string(output))
	}

	return string(output), nil
}

func (w *Worker) executeScript(scriptPath string) (string, error) {
	w.logger.WithField("script", scriptPath).Info("执行脚本")

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if strings.HasSuffix(scriptPath, ".ps1") {
			cmd = exec.Command("powershell", "-File", scriptPath)
		} else {
			cmd = exec.Command("cmd", "/c", scriptPath)
		}
	} else {
		cmd = exec.Command(scriptPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("脚本执行失败: %w: %s", err, string(output))
	}

	return string(output), nil
}

func (w *Worker) getCapabilities() string {
	capabilities := map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"cpu_num": runtime.NumCPU(),
	}

	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		capabilities["docker"] = true
	}

	// Check for Kubernetes
	if _, err := os.Stat("/var/run/secrets/kubernetes.io"); err == nil {
		capabilities["kubernetes"] = true
	}

	data, _ := json.Marshal(capabilities)
	return string(data)
}

func detectGroup() string {
	// Check environment variables
	if group := os.Getenv("WORKER_GROUP"); group != "" {
		return group
	}

	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}

	// Check for Kubernetes
	if _, err := os.Stat("/var/run/secrets/kubernetes.io"); err == nil {
		return "k8s"
	}

	// Default to OS-based group
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	default:
		return "default"
	}
}

// Helper function for logging task execution
func logToFile(taskID, workerID, message string) error {
	// TODO: Implement proper file logging
	return nil
}

// Helper function to capture both stdout and stderr
func captureOutput(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Helper function to stream output in real-time
func streamOutput(ctx context.Context, cmd *exec.Cmd, output io.Writer) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Stream output
	go io.Copy(output, stdout)
	go io.Copy(output, stderr)

	return cmd.Wait()
}

// executeTaskWithPriority executes task using 3-tier priority system:
// Priority 1: HTTP Executor (Claude server) - fastest if available
// Priority 2: Claude CLI Executor - direct claudecli calls
// Priority 3: Local execution (shell commands) - fallback
func (w *Worker) executeTaskWithPriority(task Task) (string, error) {
	ctx := context.Background()

	description := strings.TrimSpace(task.Description)

	// Check for special command prefixes
	if strings.HasPrefix(description, "shell:") {
		return w.executeShellCommand(strings.TrimPrefix(description, "shell:"))
	}

	if strings.HasPrefix(description, "script:") {
		scriptPath := strings.TrimPrefix(description, "script:")
		return w.executeScript(scriptPath)
	}

	// Priority 1: Try HTTP Executor (Claude Server)
	if w.httpExecutor.CheckServerAvailable() {
		w.logger.WithField("task_id", task.ID).Info("Priority 1: Using HTTP Executor")
		output, err := w.httpExecutor.ExecuteTask(ctx, task)
		if err == nil {
			return output, nil
		}
		w.logger.WithError(err).Warn("HTTP Executor failed, trying next priority")
	}

	// Priority 2: Try Claude CLI Executor
	if w.claudeExecutor.IsEnabled() {
		w.logger.WithField("task_id", task.ID).Info("Priority 2: Using Claude CLI Executor")
		output, err := w.claudeExecutor.ExecuteTask(ctx, task)
		if err == nil {
			return output, nil
		}
		w.logger.WithError(err).Warn("Claude CLI Executor failed, trying next priority")
	}

	// Priority 3: Fallback to local shell execution
	w.logger.WithField("task_id", task.ID).Info("Priority 3: Using local shell execution")
	return w.executeShellCommand(description)
}
