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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
	"github.com/sirupsen/logrus"
)

// AIWorker 扩展 Worker，支持子任务创建和安全控制
type AIWorker struct {
	*Worker
	workspaceDir string
	restrictPath bool
	readOnly     bool
	sandboxed    bool
}

// AIWorkerConfig AI Worker 配置
type AIWorkerConfig struct {
	*Config
	WorkspaceDir string
	RestrictPath bool
	ReadOnly     bool
	Sandboxed    bool
}

// NewAIWorker 创建支持 AI 的 Worker
func NewAIWorker(cfg *AIWorkerConfig) (*AIWorker, error) {
	baseWorker, err := New(cfg.Config)
	if err != nil {
		return nil, err
	}

	// 设置工作目录
	workspaceDir := cfg.WorkspaceDir
	if workspaceDir == "" {
		workspaceDir = os.Getenv("WORKSPACE_DIR")
		if workspaceDir == "" {
			workspaceDir = filepath.Join(os.TempDir(), "agentflow-workspace")
		}
	}

	// 创建工作目录
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}

	restrictPath := cfg.RestrictPath
	if !restrictPath {
		restrictPath = os.Getenv("RESTRICT_PATH") == "true"
	}

	readOnly := cfg.ReadOnly
	if !readOnly {
		readOnly = os.Getenv("READ_ONLY") == "true"
	}

	sandboxed := cfg.Sandboxed
	if !sandboxed {
		sandboxed = os.Getenv("SANDBOXED") == "true"
	}

	return &AIWorker{
		Worker:       baseWorker,
		workspaceDir: workspaceDir,
		restrictPath: restrictPath,
		readOnly:     readOnly,
		sandboxed:    sandboxed,
	}, nil
}

// ExecuteAITask 执行 AI 任务，支持创建子任务
func (w *AIWorker) ExecuteAITask(ctx context.Context, task *database.Task) error {
	w.logger.WithFields(logrus.Fields{
		"task_id": task.ID,
		"title":   task.Title,
	}).Info("开始执行 AI 任务")

	// 解析任务描述
	taskType, params := w.parseTaskDescription(task.Description)

	switch taskType {
	case "task": // AI 子任务
		return w.executeSubTask(ctx, task, params)

	case "shell":
		return w.executeShellTask(ctx, task, params)

	case "script":
		return w.executeScriptTask(ctx, task, params)

	case "file":
		return w.executeFileTask(ctx, task, params)

	default:
		// 默认作为 shell 命令执行
		return w.executeShellTask(ctx, task, task.Description)
	}
}

// parseTaskDescription 解析任务描述
func (w *AIWorker) parseTaskDescription(desc string) (taskType string, params string) {
	parts := strings.SplitN(desc, ":", 2)
	if len(parts) < 2 {
		return "shell", desc
	}
	return parts[0], parts[1]
}

// executeSubTask 执行子任务（创建新的任务）
func (w *AIWorker) executeSubTask(ctx context.Context, parentTask *database.Task, params string) error {
	// 解析子任务参数
	subParams := strings.SplitN(params, ":", 2)
	if len(subParams) < 2 {
		return fmt.Errorf("无效的子任务格式")
	}

	subTaskType := subParams[0]
	subTaskDesc := subParams[1]

	// 根据子任务类型创建具体任务
	var subTasks []struct {
		Title       string
		Description string
		Group       string
	}

	switch subTaskType {
	case "implement":
		// 实现功能：分解为多个子任务
		subTasks = []struct {
			Title       string
			Description string
			Group       string
		}{
			{
				Title:       fmt.Sprintf("设计数据模型: %s", subTaskDesc),
				Description: "task:design_model:" + subTaskDesc,
				Group:       w.groupName,
			},
			{
				Title:       fmt.Sprintf("实现核心逻辑: %s", subTaskDesc),
				Description: "task:implement_core:" + subTaskDesc,
				Group:       w.groupName,
			},
			{
				Title:       fmt.Sprintf("编写测试: %s", subTaskDesc),
				Description: "task:write_tests:" + subTaskDesc,
				Group:       w.groupName,
			},
		}

	case "test":
		subTasks = []struct {
			Title       string
			Description string
			Group       string
		}{
			{
				Title:       "运行单元测试",
				Description: "shell:go test ./... -v",
				Group:       w.groupName,
			},
			{
				Title:       "运行集成测试",
				Description: "shell:go test ./tests/integration/... -v",
				Group:       w.groupName,
			},
		}

	default:
		// 创建单个子任务
		subTasks = []struct {
			Title       string
			Description string
			Group       string
		}{
			{
				Title:       subTaskDesc,
				Description: params,
				Group:       w.groupName,
			},
		}
	}

	// 批量创建子任务
	var childTaskIDs []string
	for _, st := range subTasks {
		taskID, err := w.createChildTask(parentTask.ID, st.Title, st.Description, st.Group)
		if err != nil {
			w.logger.WithError(err).Error("创建子任务失败")
			continue
		}
		childTaskIDs = append(childTaskIDs, taskID)
		w.logger.WithFields(logrus.Fields{
			"child_task_id": taskID,
			"title":         st.Title,
		}).Info("子任务已创建")
	}

	// 等待所有子任务完成
	if err := w.waitForChildTasks(ctx, childTaskIDs); err != nil {
		return fmt.Errorf("等待子任务完成失败: %w", err)
	}

	// 聚合子任务结果
	return w.aggregateChildResults(parentTask.ID, childTaskIDs)
}

// createChildTask 创建子任务
func (w *AIWorker) createChildTask(parentID string, title, description, group string) (string, error) {
	if w.db != nil {
		// 本地模式：直接创建
		return w.db.CreateTask(title, description, group, "")
	}

	// 远程模式：通过 API 创建
	url := fmt.Sprintf("%s/api/v1/tasks", w.masterURL)

	data := map[string]interface{}{
		"title":       title,
		"description": description,
		"group_name":  group,
		"parent_id":   parentID,
	}
	body, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("创建任务失败: %s", resp.Status)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	return result["task_id"], nil
}

// waitForChildTasks 等待子任务完成
func (w *AIWorker) waitForChildTasks(ctx context.Context, childTaskIDs []string) error {
	w.logger.WithField("count", len(childTaskIDs)).Info("等待子任务完成")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	pendingTasks := make(map[string]bool)
	for _, id := range childTaskIDs {
		pendingTasks[id] = true
	}

	for len(pendingTasks) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for taskID := range pendingTasks {
				status, err := w.getTaskStatus(taskID)
				if err != nil {
					w.logger.WithError(err).WithField("task_id", taskID).Warn("获取任务状态失败")
					continue
				}

				if status == "completed" || status == "failed" {
					delete(pendingTasks, taskID)
					w.logger.WithFields(logrus.Fields{
						"task_id": taskID,
						"status":  status,
					}).Info("子任务已完成")
				}
			}
		}
	}

	return nil
}

// getTaskStatus 获取任务状态
func (w *AIWorker) getTaskStatus(taskID string) (string, error) {
	if w.db != nil {
		task, err := w.db.GetTask(taskID)
		if err != nil {
			return "", err
		}
		return task.Status, nil
	}

	url := fmt.Sprintf("%s/api/v1/tasks/%s", w.masterURL, taskID)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取任务失败: %s", resp.Status)
	}

	var task map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&task)
	return task["status"].(string), nil
}

// aggregateChildResults 聚合子任务结果
func (w *AIWorker) aggregateChildResults(parentID string, childTaskIDs []string) error {
	var results []string
	var errors []string

	for _, taskID := range childTaskIDs {
		if w.db != nil {
			task, err := w.db.GetTask(taskID)
			if err != nil {
				errors = append(errors, err.Error())
				continue
			}

			if task.Status == "completed" {
				results = append(results, task.Result)
			} else if task.Status == "failed" {
				errors = append(errors, task.Error)
			}
		}
	}

	// 保存聚合结果
	aggregateResult := strings.Join(results, "\n\n")
	if len(errors) > 0 {
		aggregateResult += "\n\n错误:\n" + strings.Join(errors, "\n")
	}

	w.logger.WithField("result_length", len(aggregateResult)).Info("子任务结果已聚合")
	return nil
}

// executeShellTask 执行 Shell 任务（带安全控制）
func (w *AIWorker) executeShellTask(ctx context.Context, task *database.Task, command string) error {
	w.logger.WithField("command", command).Info("执行 Shell 命令")

	// 只读检查
	if w.readOnly {
		return fmt.Errorf("只读模式，不允许执行命令")
	}

	// 沙箱检查
	if w.sandboxed {
		return w.executeInSandbox(ctx, command)
	}

	// 设置工作目录
	workDir := w.workspaceDir
	if task.WorkspaceDir != "" {
		workDir = task.WorkspaceDir
	}

	// 路径限制检查
	if w.restrictPath {
		if !filepath.IsAbs(workDir) {
			return fmt.Errorf("工作目录必须是绝对路径")
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	cmd.Dir = workDir

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("命令执行失败: %w: %s", err, string(output))
	}

	w.logger.WithField("output_length", len(output)).Info("命令执行成功")
	return nil
}

// executeInSandbox 在沙箱中执行命令
func (w *AIWorker) executeInSandbox(ctx context.Context, command string) error {
	w.logger.Info("在沙箱中执行命令")

	// 使用 Docker 沙箱
	dockerCmd := fmt.Sprintf("docker run --rm -v %s:/workspace -w /workspace alpine sh -c %q",
		w.workspaceDir, command)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", dockerCmd)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", dockerCmd)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("沙箱执行失败: %w: %s", err, string(output))
	}

	w.logger.Info("沙箱执行成功")
	return nil
}

// executeScriptTask 执行脚本任务
func (w *AIWorker) executeScriptTask(ctx context.Context, task *database.Task, scriptPath string) error {
	w.logger.WithField("script", scriptPath).Info("执行脚本")

	if w.readOnly {
		return fmt.Errorf("只读模式，不允许执行脚本")
	}

	// 检查脚本路径
	if w.restrictPath {
		absPath, err := filepath.Abs(scriptPath)
		if err != nil {
			return fmt.Errorf("无效的脚本路径: %w", err)
		}

		// 确保脚本在工作目录内
		workDir := w.workspaceDir
		if task.WorkspaceDir != "" {
			workDir = task.WorkspaceDir
		}

		if !strings.HasPrefix(absPath, workDir) {
			return fmt.Errorf("脚本必须在工作目录内")
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if strings.HasSuffix(scriptPath, ".ps1") {
			cmd = exec.CommandContext(ctx, "powershell", "-File", scriptPath)
		} else {
			cmd = exec.CommandContext(ctx, "cmd", "/c", scriptPath)
		}
	} else {
		cmd = exec.CommandContext(ctx, scriptPath)
	}

	// 设置工作目录
	workDir := w.workspaceDir
	if task.WorkspaceDir != "" {
		workDir = task.WorkspaceDir
	}
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("脚本执行失败: %w: %s", err, string(output))
	}

	w.logger.Info("脚本执行成功")
	return nil
}

// executeFileTask 执行文件操作任务
func (w *AIWorker) executeFileTask(ctx context.Context, task *database.Task, params string) error {
	parts := strings.SplitN(params, ":", 3)
	if len(parts) < 2 {
		return fmt.Errorf("无效的文件操作格式")
	}

	operation := parts[0]
	targetPath := parts[1]

	// 构建完整路径
	fullPath := filepath.Join(w.workspaceDir, targetPath)

	switch operation {
	case "write":
		if len(parts) < 3 {
			return fmt.Errorf("write 操作需要内容")
		}
		content := parts[2]
		return w.writeFile(fullPath, content)

	case "read":
		return w.readFile(fullPath)

	case "delete":
		if w.readOnly {
			return fmt.Errorf("只读模式，不允许删除文件")
		}
		return w.deleteFile(fullPath)

	default:
		return fmt.Errorf("不支持的文件操作: %s", operation)
	}
}

// writeFile 写文件
func (w *AIWorker) writeFile(path, content string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 写文件
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("写文件失败: %w", err)
	}

	w.logger.WithField("path", path).Info("文件已写入")
	return nil
}

// readFile 读文件
func (w *AIWorker) readFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读文件失败: %w", err)
	}

	w.logger.WithFields(logrus.Fields{
		"path":           path,
		"content_length": len(content),
	}).Info("文件已读取")
	return nil
}

// deleteFile 删除文件
func (w *AIWorker) deleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	w.logger.WithField("path", path).Info("文件已删除")
	return nil
}

// StreamOutput 流式输出（用于实时日志）
func (w *AIWorker) StreamOutput(ctx context.Context, cmd *exec.Cmd, output io.Writer) error {
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

	// 流式输出
	go io.Copy(output, stdout)
	go io.Copy(output, stderr)

	return cmd.Wait()
}
