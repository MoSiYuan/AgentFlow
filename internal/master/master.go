package master

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jiangxiaolong/agentflow-go/internal/database"
	"github.com/sirupsen/logrus"
)

type Master struct {
	db           *database.Database
	router       *gin.Engine
	dbPath       string
	workers      map[string]*os.Process
	workersMutex sync.RWMutex
	logger       *logrus.Logger
}

type Config struct {
	DBPath    string
	Host      string
	Port      int
	AutoStart bool // Auto-start local workers
}

func New(cfg *Config) (*Master, error) {
	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	if err := db.Init(); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware(logger))

	m := &Master{
		db:     db,
		router: router,
		dbPath: cfg.DBPath,
		workers: make(map[string]*os.Process),
		logger: logger,
	}

	m.setupRoutes()

	if cfg.AutoStart {
		if err := m.startLocalWorkers(); err != nil {
			logger.WithError(err).Warn("启动本地 workers 失败")
		}
	}

	return m, nil
}

func (m *Master) setupRoutes() {
	api := m.router.Group("/api/v1")
	{
		// Task endpoints
		api.POST("/tasks", m.createTask)
		api.GET("/tasks", m.listTasks)
		api.GET("/tasks/:id", m.getTask)
		api.POST("/tasks/:id/complete", m.completeTask)
		api.POST("/tasks/:id/fail", m.failTask)

		// Worker endpoints
		api.GET("/workers", m.listWorkers)
		api.POST("/workers/:id/heartbeat", m.workerHeartbeat)

		// Task fetching for workers
		api.GET("/tasks/pending", m.fetchPendingTask)
		api.POST("/tasks/:id/lock", m.lockTask)
		api.POST("/tasks/:id/unlock", m.unlockTask)

		// Stats
		api.GET("/stats", m.getStats)
		api.GET("/stats/groups", m.getGroupStats)
	}

	// Health check
	m.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (m *Master) Run(addr string) error {
	m.logger.WithField("addr", addr).Info("启动 Master 服务")
	return m.router.Run(addr)
}

func (m *Master) Close() error {
	m.stopLocalWorkers()
	return m.db.Close()
}

// Task handlers
func (m *Master) createTask(c *gin.Context) {
	var req struct {
		Title               string `json:"title" binding:"required"`
		Description         string `json:"description"`
		GroupName           string `json:"group_name"`
		CompletionCriteria string `json:"completion_criteria"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.GroupName == "" {
		req.GroupName = "default"
	}

	taskID, err := m.db.CreateTask(req.Title, req.Description, req.GroupName, req.CompletionCriteria)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	m.logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"title":   req.Title,
		"group":   req.GroupName,
	}).Info("任务已创建")

	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}

func (m *Master) listTasks(c *gin.Context) {
	status := c.Query("status")
	group := c.Query("group")

	tasks, err := m.db.ListTasks(status, group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (m *Master) getTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := m.db.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (m *Master) completeTask(c *gin.Context) {
	taskID := c.Param("id")

	var req struct {
		WorkerID string `json:"worker_id" binding:"required"`
		Result   string `json:"result"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := m.db.CompleteTask(taskID, req.WorkerID, req.Result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	m.logger.WithFields(logrus.Fields{
		"task_id":   taskID,
		"worker_id": req.WorkerID,
	}).Info("任务已完成")

	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func (m *Master) failTask(c *gin.Context) {
	taskID := c.Param("id")

	var req struct {
		WorkerID string `json:"worker_id" binding:"required"`
		Error    string `json:"error"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := m.db.FailTask(taskID, req.WorkerID, req.Error); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	m.logger.WithFields(logrus.Fields{
		"task_id":   taskID,
		"worker_id": req.WorkerID,
		"error":     req.Error,
	}).Warn("任务失败")

	c.JSON(http.StatusOK, gin.H{"status": "failed"})
}

// Worker handlers
func (m *Master) listWorkers(c *gin.Context) {
	group := c.Query("group")

	workers, err := m.db.ListWorkers(group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workers": workers})
}

func (m *Master) workerHeartbeat(c *gin.Context) {
	workerID := c.Param("id")

	if err := m.db.UpdateWorkerHeartbeat(workerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Task fetching for workers
func (m *Master) fetchPendingTask(c *gin.Context) {
	groupName := c.Query("group")
	if groupName == "" {
		groupName = "default"
	}

	tasks, err := m.db.ListTasks("pending", groupName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(tasks) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (m *Master) lockTask(c *gin.Context) {
	taskID := c.Param("id")

	var req struct {
		WorkerID string `json:"worker_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	locked, err := m.db.LockTask(taskID, req.WorkerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !locked {
		c.JSON(http.StatusConflict, gin.H{"status": "already_locked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "locked"})
}

func (m *Master) unlockTask(c *gin.Context) {
	taskID := c.Param("id")

	var req struct {
		WorkerID string `json:"worker_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := m.db.UnlockTask(taskID, req.WorkerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unlocked"})
}

// Stats
func (m *Master) getStats(c *gin.Context) {
	stats, err := m.db.GetGroupStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (m *Master) getGroupStats(c *gin.Context) {
	stats, err := m.db.GetGroupStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group_stats": stats})
}

// Local worker management
func (m *Master) startLocalWorkers() error {
	groups := detectLocalGroups()

	for _, group := range groups {
		if err := m.startWorkerProcess(group); err != nil {
			m.logger.WithError(err).WithField("group", group).Error("启动 worker 失败")
		}
	}

	return nil
}

func (m *Master) startWorkerProcess(groupName string) error {
	workerID := uuid.New().String()
	cmd := exec.Command("./bin/worker",
		"--master", "http://localhost:8848",
		"--group", groupName,
		"--worker-id", workerID,
	)
	cmd.Dir = "." // TODO: use working directory from config

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 worker 进程失败: %w", err)
	}

	m.workersMutex.Lock()
	m.workers[workerID] = cmd.Process
	m.workersMutex.Unlock()

	m.logger.WithFields(logrus.Fields{
		"worker_id": workerID,
		"group":     groupName,
		"pid":       cmd.Process.Pid,
	}).Info("本地 worker 已启动")

	return nil
}

func (m *Master) stopLocalWorkers() {
	m.workersMutex.Lock()
	defer m.workersMutex.Unlock()

	for workerID, proc := range m.workers {
		if err := proc.Kill(); err != nil {
			m.logger.WithError(err).WithField("worker_id", workerID).Error("停止 worker 失败")
		}
	}

	m.workers = make(map[string]*os.Process)
}

func detectLocalGroups() []string {
	var groups []string

	// Detect OS-based groups
	switch os := runtime.GOOS; os {
	case "windows":
		groups = append(groups, "windows")
	case "linux":
		groups = append(groups, "linux")
	case "darwin":
		groups = append(groups, "darwin")
	}

	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		groups = append(groups, "docker")
	}

	// Check for Kubernetes
	if _, err := os.Stat("/var/run/secrets/kubernetes.io"); err == nil {
		groups = append(groups, "k8s")
	}

	return groups
}

func loggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		entry := logger.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user_agent": c.Request.UserAgent(),
		})

		if c.Writer.Status() >= 500 {
			entry.Error("Server error")
		} else if c.Writer.Status() >= 400 {
			entry.Warn("Client error")
		} else {
			entry.Info("Request completed")
		}
	}
}
