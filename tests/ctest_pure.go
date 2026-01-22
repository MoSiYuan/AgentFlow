package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	mrand "math/rand"
	"strings"
	"time"
)

// 简化的 Task 结构
type Task struct {
	ID          int
	Title       string
	Description string
	Status      string
	Result      string
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║          AgentFlow 克苏鲁神话故事实战测试（纯标准库版）              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. 初始化
	dbPath := "ctest_pure.db"
	storyDir := "./ctest_stories"

	os.Remove(dbPath)
	os.RemoveAll(storyDir)
	os.MkdirAll(storyDir, 0755)

	fmt.Println("📦 1. 初始化环境...")
	fmt.Println("✅ 测试目录已创建")
	fmt.Println()

	// 2. 创建 Worker
	workerID := fmt.Sprintf("worker-%d", time.Now().Unix())
	fmt.Printf("🤖 2. Worker ID: %s\n", workerID)
	fmt.Println()

	// 3. 创建 10 个故事任务
	fmt.Println("📝 3. 创建 10 个故事生成任务...")
	fmt.Println()

	var tasks []Task
	stories := []string{
		"沉睡之城", "深海召唤", "古老低语", "星空之下", "暗影之地",
		"遗忘之书", "梦魇边缘", "虚空凝视", "时间之河", "永恒迷宫",
	}

	for i, title := range stories {
		desc := fmt.Sprintf("write_story:%s:cthulhu_mythos:%03d", title, i+1)
		task := Task{
			ID:          i + 1,
			Title:       title,
			Description: desc,
			Status:      "pending",
		}
		tasks = append(tasks, task)
		fmt.Printf("  ✅ 任务 #%d: %s\n", task.ID, task.Title)
	}

	fmt.Println()
	fmt.Printf("✅ 已创建 %d 个故事生成任务\n", len(tasks))
	fmt.Println()

	// 4. 执行故事生成任务
	fmt.Println("⏳ 4. 执行故事生成任务...")
	fmt.Println()

	for _, task := range tasks {
		fmt.Printf("  🔨 执行任务 #%d: %s\n", task.ID, task.Title)

		// 解析任务描述
		parts := strings.Split(task.Description, ":")
		if len(parts) >= 4 && parts[0] == "write_story" {
			title := parts[1]
			storyType := parts[2]

			// 生成故事
			story := generateCthulhuStory(title, storyType, workerID)

			// 保存文件
			filename := filepath.Join(storyDir, fmt.Sprintf("story_%d.md", task.ID))
			content := fmt.Sprintf("# %s\n\n", title)
			content += fmt.Sprintf("**类型**: %s\n", storyType)
			content += fmt.Sprintf("**作者**: %s\n", workerID)
			content += fmt.Sprintf("**创建时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
			content += "---\n\n"
			content += story
			content += "\n\n---\n\n"
			content += "## 评审区\n\n"
			content += "*（评审将添加到此处）*\n"

			err := ioutil.WriteFile(filename, []byte(content), 0644)
			if err != nil {
				fmt.Printf("     ❌ 写入文件失败: %v\n", err)
				continue
			}

			task.Status = "completed"
			task.Result = fmt.Sprintf("故事已保存到 %s", filename)
			wordCount := len(strings.Split(story, " "))
			fmt.Printf("     ✅ 故事已生成: %d 词\n", wordCount)
		}
	}

	fmt.Println()
	fmt.Println("📊 5. 故事生成完成统计")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	total := len(tasks)
	storyCompleted := 0
	for _, t := range tasks {
		if t.Status == "completed" {
			storyCompleted++
		}
	}

	fmt.Printf("总任务数: %d\n", total)
	fmt.Printf("已完成: %d\n", storyCompleted)
	fmt.Printf("成功率: %.0f%%\n", float64(storyCompleted)/float64(total)*100)
	fmt.Println()

	// 5. 创建评审任务
	fmt.Println("💬 6. 创建评审任务...")
	fmt.Println()

	var reviewTasks []Task
	// 检查所有任务的状态，不管状态如何都创建评审（因为故事文件已生成）
	for _, task := range tasks {
		// 每个故事创建 2 个评审
		for i := 1; i <= 2; i++ {
			reviewTask := Task{
				ID:          len(reviewTasks) + len(tasks) + 1,
				Title:       fmt.Sprintf("评审:%s #%d", task.Title, i),
				Description: fmt.Sprintf("review_story:%d:%s:review", task.ID, task.Title),
				Status:      "pending",
			}
			reviewTasks = append(reviewTasks, reviewTask)
			fmt.Printf("  ✅ 创建评审任务: %s\n", reviewTask.Title)
		}
	}

	fmt.Println()
	fmt.Printf("✅ 已创建 %d 个评审任务\n", len(reviewTasks))
	fmt.Println()

	// 6. 执行评审任务
	fmt.Println("⏳ 7. 执行评审任务...")
	fmt.Println()

	for _, task := range reviewTasks {
		fmt.Printf("  🔨 执行评审: %s\n", task.Title)

		// 解析评审任务
		parts := strings.Split(task.Description, ":")
		if len(parts) >= 3 && parts[0] == "review_story" {
			targetTaskID := parts[1]
			targetTitle := parts[2]

			// 读取目标故事文件
			storyFile := filepath.Join(storyDir, fmt.Sprintf("story_%s.md", targetTaskID))
			content, err := ioutil.ReadFile(storyFile)
			if err != nil {
				fmt.Printf("     ❌ 读取故事失败: %v\n", err)
				continue
			}

			// 生成评审
			review := generateReview(targetTitle, workerID)

			// 追加到文件
			reviewSection := fmt.Sprintf("\n\n### 评审者: %s\n", workerID)
			reviewSection += fmt.Sprintf("**评审时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
			reviewSection += review
			reviewSection += "\n\n---\n"

			updatedContent := strings.ReplaceAll(
				string(content),
				"*（评审将添加到此处）*",
				reviewSection+"*（评审将添加到此处）*",
			)

			err = ioutil.WriteFile(storyFile, []byte(updatedContent), 0644)
			if err != nil {
				fmt.Printf("     ❌ 写入评审失败: %v\n", err)
				continue
			}

			task.Status = "completed"
			fmt.Printf("     ✅ 评审已添加到 story_%s.md\n", targetTaskID)
		}
	}

	fmt.Println()
	fmt.Println("📊 8. 最终统计")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	totalTasks := len(tasks) + len(reviewTasks)
	completedCount := 0
	for _, t := range tasks {
		if t.Status == "completed" {
			completedCount++
		}
	}
	for _, t := range reviewTasks {
		if t.Status == "completed" {
			completedCount++
		}
	}

	fmt.Printf("总任务数: %d\n", totalTasks)
	fmt.Printf("已完成: %d\n", completedCount)
	fmt.Printf("成功率: %.0f%%\n", float64(completedCount)/float64(totalTasks)*100)
	fmt.Printf("故事: %d 个\n", len(tasks))
	fmt.Printf("评审: %d 个\n", len(reviewTasks))
	fmt.Println()

	// 9. 展示示例故事
	fmt.Println("📖 9. 示例故事和评审")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	files, _ := filepath.Glob(filepath.Join(storyDir, "story_*.md"))
	if len(files) > 0 {
		// 读取第一个文件
		content, _ := ioutil.ReadFile(files[0])
		lines := strings.Split(string(content), "\n")

		// 显示前 35 行
		fmt.Printf("📄 %s\n", filepath.Base(files[0]))
		fmt.Println()
		for i, line := range lines {
			if i >= 35 {
				break
			}
			fmt.Println(line)
		}
		if len(lines) > 35 {
			fmt.Println("...")
		}
	}

	fmt.Println()
	fmt.Println("✅ 测试完成！")
	fmt.Println()
	fmt.Printf("📁 故事保存在: %s/\n", storyDir)
	fmt.Printf("📊 共生成 %d 个故事，%d 个评审\n", len(tasks), len(reviewTasks))
	fmt.Println()
}

// 生成克苏鲁神话故事
func generateCthulhuStory(title, storyType, workerID string) string {
	locations := []string{
		"古老的图书馆", "深海废墟", "被遗忘的墓穴", "迷雾笼罩的村庄",
		"南极冰原", "太平洋岛屿", "梦境边缘", "时间裂缝",
	}
	creatures := []string{
		"克苏鲁", "奈亚拉托提普", "犹格·索托斯", "莎布·尼古拉丝",
		"修格斯", "深潜者", "米·戈", "廷达罗斯猎犬",
	}
	artifacts := []string{
		"necronomicon 死灵之书", "黄色印记", "克苏鲁雕像", "古老卷轴",
		"星之石", "梦境护符", "深渊之钥", "时空水晶",
	}
	feelings := []string{
		"无法名状的恐惧", "疯狂的呓语", "理智的崩塌", "深海的呼唤",
		"星空的凝视", "时间的错乱", "梦境侵蚀", "虚空低语",
	}

	mrand.Seed(time.Now().UnixNano())

	location := locations[mrand.Intn(len(locations))]
	creature := creatures[mrand.Intn(len(creatures))]
	artifact := artifacts[mrand.Intn(len(artifacts))]
	feeling := feelings[mrand.Intn(len(feelings))]

	story := fmt.Sprintf(`
在%s的深处，%s正静静沉睡。

调查员发现了%s。那是一个注定改变命运的瞬间。

%s充斥着整个空间。%s开始从虚空中浮现，扭曲着现实与梦境的边界。

"这不可能..."调查员喃喃自语，但理智正在崩塌。

%s传来低沉的回应，那是不属于这个世界的语言，却在脑海中直接形成概念。

%s开始显现，世界正在重新定义。

也许，这从来不是人类的世界。我们只是%s的临时居所。

当%s第一次看到%s时，调查员知道生活永远不会回到从前。

那个命运之夜，%s中的%s发出了召唤。

%s和%s交织在一起，形成了一个无法逃脱的漩涡。

%s不仅是禁忌，更是%s的诅咒。

"我们必须阻止它..."但一切都太迟了。

%s已经苏醒。末日即将来临。

*(此故事由 %s 在%s创建)*
`,
		location, creature, artifact, feeling, creature,
		location, feeling, creature,
		creature,
		creature, location, creature,
		feeling, creature,
		artifact, creature,
		feeling, creature,
		creature,
		workerID, time.Now().Format("2006-01-02 15:04:05"),
	)

	return story
}

// 生成评审
func generateReview(title, reviewerID string) string {
	templates := []string{
		`
**评分**: ⭐⭐⭐⭐☆

**优点**:
- 氛围营造出色，克苏鲁神话的恐惧感深入人心
- 诡异意象运用得当，"无法名状的恐惧"表达精准
- 节奏把握良好，层层递进
- 结局设计令人印象深刻

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
- 文笔老练，构思精巧

**改进建议**:
- 无明显瑕疵，已是佳作

**总体评价**: 这是一个注定要被载入%s档案的作品。
`,

		`
**评分**: ⭐⭐⭐☆☆

**优点**:
- 情节设定有趣
- 神话元素运用准确
- 整体结构完整

**改进建议**:
- 部分描写略显仓促
- 可以增加更多细节来增强代入感
- 人物动机可以更清晰

**总体评价**: 基础扎实，有潜力成为更好的作品。建议继续打磨细节。
`,
	}

	mrand.Seed(time.Now().UnixNano())
	template := templates[mrand.Intn(len(templates))]

	if strings.Contains(template, "%s") {
		return fmt.Sprintf(template, title)
	}
	return template
}
