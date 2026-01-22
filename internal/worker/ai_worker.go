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

	case "write_story": // 生成故事
		return w.executeWriteStoryTask(ctx, task, params)

	case "review_story": // 评审故事
		return w.executeReviewStoryTask(ctx, task, params)

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
func (w *AIWorker) createChildTask(parentID int, title, description, group string) (string, error) {
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
func (w *AIWorker) aggregateChildResults(parentID int, childTaskIDs []string) error {
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
		"path":          path,
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

// executeWriteStoryTask 执行故事生成任务
func (w *AIWorker) executeWriteStoryTask(ctx context.Context, task *database.Task, params string) error {
	// 解析参数: write_story:TITLE:TYPE:INDEX
	parts := strings.Split(params, ":")
	if len(parts) < 4 {
		return fmt.Errorf("无效的故事任务格式，需要: TITLE:TYPE:INDEX")
	}

	title := parts[0]
	storyType := parts[1]
	index := parts[2]

	w.logger.WithFields(logrus.Fields{
		"title":      title,
		"type":       storyType,
		"index":      index,
		"worker_id":  w.id,
	}).Info("生成克苏鲁神话故事")

	// 生成故事内容
	story := w.generateCthulhuStory(title, storyType, w.id)

	// 保存到公共区域
	storyDir := filepath.Join(w.workspaceDir, "cthulhu_stories")
	if err := os.MkdirAll(storyDir, 0755); err != nil {
		return fmt.Errorf("创建故事目录失败: %w", err)
	}

	storyFile := filepath.Join(storyDir, fmt.Sprintf("story_%s.md", task.ID))

	// 写入故事
	content := fmt.Sprintf("# %s\n\n", title)
	content += fmt.Sprintf("**类型**: %s\n", storyType)
	content += fmt.Sprintf("**作者**: Worker %s\n", w.id)
	content += fmt.Sprintf("**创建时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	content += "---\n\n"
	content += story
	content += "\n\n---\n\n"
	content += "## 评审区\n\n"
	content += "*（评审将添加到此处）*\n"

	if err := os.WriteFile(storyFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入故事文件失败: %w", err)
	}

	w.logger.WithFields(logrus.Fields{
		"file":     storyFile,
		"words":    len(story),
	}).Info("故事已生成")

	return nil
}

// executeReviewStoryTask 执行故事评审任务
func (w *AIWorker) executeReviewStoryTask(ctx context.Context, task *database.Task, params string) error {
	// 解析参数: review_story:TASK_ID:TITLE:TYPE
	parts := strings.Split(params, ":")
	if len(parts) < 3 {
		return fmt.Errorf("无效的评审任务格式")
	}

	targetTaskID := parts[0]
	targetTitle := parts[1]

	w.logger.WithFields(logrus.Fields{
		"target_task": targetTaskID,
		"target_title": targetTitle,
		"reviewer":    w.id,
	}).Info("评审故事")

	// 查找目标故事文件
	storyDir := filepath.Join(w.workspaceDir, "cthulhu_stories")
	storyFile := filepath.Join(storyDir, fmt.Sprintf("story_%s.md", targetTaskID))

	// 读取原故事
	storyContent, err := os.ReadFile(storyFile)
	if err != nil {
		return fmt.Errorf("读取故事失败: %w", err)
	}

	// 生成评审
	review := w.generateStoryReview(string(storyContent), targetTitle, w.id)

	// 追加到原文件末尾（在评审区）
	reviewSection := fmt.Sprintf("\n\n### 评审者: Worker %s\n", w.id)
	reviewSection += fmt.Sprintf("**评审时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	reviewSection += review
	reviewSection += "\n\n---\n"

	// 替换 "*（评审将添加到此处）*"
	updatedContent := strings.ReplaceAll(
		string(storyContent),
		"*（评审将添加到此处）*",
		reviewSection+"*（评审将添加到此处）*",
	)

	if err := os.WriteFile(storyFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("写入评审失败: %w", err)
	}

	w.logger.WithField("review_file", storyFile).Info("评审已完成")
	return nil
}

// generateCthulhuStory 生成克苏鲁神话故事
func (w *AIWorker) generateCthulhuStory(title, storyType, workerID string) string {
	// 克苏鲁神话元素库
	elements := map[string][]string{
		"location": {
			"古老的图书馆", "深海废墟", "被遗忘的墓穴", "迷雾笼罩的村庄",
			"南极冰原", "太平洋岛屿", "梦境边缘", "时间裂缝",
		},
		"creature": {
			"克苏鲁", "奈亚拉托提普", "犹格·索托斯", "莎布·尼古拉丝",
			"修格斯", "深潜者", "米·戈", "廷达罗斯猎犬",
		},
		"artifact": {
			"necronomicon 死灵之书", "黄色印记", "克苏鲁雕像", "古老卷轴",
			"星之石", "梦境护符", "深渊之钥", "时空水晶",
		},
		"feeling": {
			"无法名状的恐惧", "疯狂的呓语", "理智的崩塌", "深海的呼唤",
			"星空的凝视", "时间的错乱", "梦境侵蚀", "虚空低语",
		},
	}

	stories := []string{
		fmt.Sprintf(`
在%s的深处，%s正静静沉睡。

%s发现了%s。那是一个注定改变命运的瞬间。

%s充斥着整个空间。%s开始从虚空中浮现，扭曲着现实与梦境的边界。

"这不可能..."%s喃喃自语，但理智正在崩塌。

%s传来低沉的回应，那是不属于这个世界的语言，却在脑海中直接形成概念。

%s开始显现，世界正在重新定义。
`,
			randomChoice(elements["location"]),
			randomChoice(elements["creature"]),
			randomChoice(elements["creature"]),
			randomChoice(elements["artifact"]),
			randomChoice(elements["feeling"]),
			randomChoice(elements["creature"]),
			"调查员",
			randomChoice(elements["location"]),
			randomChoice(elements["feeling"]),
		),

		fmt.Sprintf(`
%s的传说早已被遗忘，直到那个命运的夜晚。

%s在%s中被发现，散发着%s的气息。

%s无法抗拒%s的诱惑。那个声音，如此古老，又如此熟悉。

"%s"_%s在梦中重复着这句话。

%s开始显现。理智的防线正在瓦解，%s的到来预示着一个新时代。

也许，这从来不是人类的世界。我们只是%s的临时居所。
`,
			title,
			randomChoice(elements["artifact"]),
			randomChoice(elements["location"]),
			randomChoice(elements["feeling"]),
			"好奇心",
			randomChoice(elements["creature"]),
			randomChoice(elements["feeling"]),
			"梦境者",
			randomChoice(elements["creature"]),
			randomChoice(elements["location"]),
			randomChoice(elements["creature"]),
		),

		fmt.Sprintf(`
当%s第一次看到%s时，%s知道生活永远不会回到从前。

那个%s夜晚，%s中的%s发出了召唤。

%s和%s交织在一起，形成了一个无法逃脱的漩涡。

%s不仅是禁忌，更是%s的诅咒。

"我们必须阻止它..."但%s知道，一切都太迟了。

%s已经苏醒。%s即将来临。
`,
			"主角",
			randomChoice(elements["creature"]),
			"他",
			randomChoice(elements["feeling"]),
			"月圆",
			randomChoice(elements["location"]),
			randomChoice(elements["artifact"]),
			randomChoice(elements["creature"]),
			randomChoice(elements["feeling"]),
			randomChoice(elements["artifact"]),
			randomChoice(elements["creature"]),
			"他",
			randomChoice(elements["creature"]),
			"末日",
		),
	}

	// 随机选择一个故事模板
	story := stories[int(time.Now().Unix())%len(stories)]

	// 添加 Worker 标识
	story += fmt.Sprintf("\n\n*(此故事由 Worker %s 在%s创建)*\n",
		workerID, time.Now().Format("2006-01-02 15:04:05"))

	return story
}

// generateStoryReview 生成故事评审
func (w *AIWorker) generateStoryReview(storyContent, title, reviewerID string) string {
	reviewTemplates := []string{
		`
**评分**: ⭐⭐⭐⭐☆

**优点**:
- 氛围营造出色，克苏鲁神话的恐惧感深入人心
- 诡异意象运用得当，"无法名状的恐惧"表达精准
- 节奏把握良好，层层递进

**改进建议**:
- 可以增加更多环境描写
- 人物心理刻画可以更深入

**总体评价**: 一篇优秀的克苏鲁神话短文，成功传达了 cosmic horror 的核心精髓。
`,

		`
**评分**: ⭐⭐⭐⭐⭐

**优点**:
- 完美诠释了洛夫克拉夫特的风格
- 古老低语和疯狂呓语的使用恰到好处
- 结尾震撼，留下深刻印象

**改进建议**:
- 无明显瑕疵，已是佳作

**总体评价**: 这是一个注定要被载入%s档案的作品。文笔老练，构思精巧。
`,

		`
**评分**: ⭐⭐⭐☆☆

**优点**:
- 情节设定有趣
- 神话元素运用准确

**改进建议**:
- 部分描写略显仓促
- 可以增加更多细节来增强代入感
- 人物动机可以更清晰

**总体评价**: 基础扎实，有潜力成为更好的作品。建议继续打磨细节。
`,

		`
**评分**: ⭐⭐⭐⭐☆

**优点**:
- 氛围营造极具张力
- 对未知的恐惧表达到位
- 文笔流畅，阅读体验良好

**改进建议**:
- 可以考虑增加一个反转
- 结尾可以更加开放，留给读者想象空间

**总体评价**: 一篇质量上乘的克苏鲁神话作品，展现了作者对 horror 写作的深刻理解。
`,
	}

	// 随机选择评审模板
	template := reviewTemplates[int(time.Now().Unix()+int64(len(reviewerID)))%len(reviewTemplates)]

	if strings.Contains(template, "%s") {
		return fmt.Sprintf(template, title)
	}
	return template
}

// randomChoice 从切片中随机选择
func randomChoice(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[int(time.Now().Unix())%len(items)]
}

