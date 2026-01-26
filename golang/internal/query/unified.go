package query

import (
	"fmt"
	"strings"

	"github.com/jiangxiaolong/agentflow-go/internal/database"
)

// TaskQuery represents a unified task query with multiple filter conditions
type TaskQuery struct {
	// AgentFlow conditions
	TaskID *string
	Status *string
	Group  *string

	// Claude conditions
	SessionUUID   *string
	MessageUUID   *string
	Slug          *string
	ParentMessageUUID *string

	// Pagination
	Limit  *int
	Offset *int

	// Sorting
	OrderBy    *string
	OrderDesc  *bool
}

// UnifiedTask represents a task with optional Claude mapping information
type UnifiedTask struct {
	database.Task
	Claude *database.ClaudeMapping `json:"claude,omitempty"`
}

// UnifiedQueryResult represents the query result with metadata
type UnifiedQueryResult struct {
	Tasks      []UnifiedTask `json:"tasks"`
	TotalCount int           `json:"total_count"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// UnifiedQuery provides unified query interface for AgentFlow and Claude data
type UnifiedQuery struct {
	db *database.Database
}

// NewUnifiedQuery creates a new unified query interface
func NewUnifiedQuery(db *database.Database) *UnifiedQuery {
	return &UnifiedQuery{db: db}
}

// QueryTasks executes a unified query with multiple filter conditions
func (uq *UnifiedQuery) QueryTasks(query *TaskQuery) (*UnifiedQueryResult, error) {
	// Build SQL query
	sql, args := uq.buildQuery(query)

	// Execute query
	rows, err := uq.db.GetDB().Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Parse results
	var tasks []UnifiedTask
	taskMap := make(map[string]*UnifiedTask)

	for rows.Next() {
		var task UnifiedTask
		var claudeID *int
		var claudeTaskID *int
		var claudeSessionUUID *string
		var claudeMessageUUID *string
		var claudeParentMessageUUID *string
		var claudeSlug *string
		var claudeSource *string
		var claudeCreatedAt *string
		var claudeUpdatedAt *string

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.GroupName,
			&task.CreatedAt,
			&claudeID,
			&claudeTaskID,
			&claudeSessionUUID,
			&claudeMessageUUID,
			&claudeParentMessageUUID,
			&claudeSlug,
			&claudeSource,
			&claudeCreatedAt,
			&claudeUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Check if task already exists in map
		if _, ok := taskMap[task.ID]; ok {
			continue
		}

		// Add Claude mapping if exists
		if claudeID != nil {
			task.Claude = &database.ClaudeMapping{
				ID:                *claudeID,
				TaskID:            fmt.Sprintf("%d", *claudeTaskID),
				SessionUUID:       *claudeSessionUUID,
				MessageUUID:       *claudeMessageUUID,
				ParentMessageUUID: claudeParentMessageUUID,
				Slug:              *claudeSlug,
				Source:            *claudeSource,
			}
		}

		taskMap[task.ID] = &task
		tasks = append(tasks, task)
	}

	// Get total count
	countSQL, countArgs := uq.buildCountQuery(query)
	var totalCount int
	countRow := uq.db.GetDB().QueryRow(countSQL, countArgs...)
	if err := countRow.Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build result
	limit := 50
	if query.Limit != nil {
		limit = *query.Limit
	}

	offset := 0
	if query.Offset != nil {
		offset = *query.Offset
	}

	result := &UnifiedQueryResult{
		Tasks:      tasks,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}

	return result, nil
}

// buildQuery builds the SQL query based on filter conditions
func (uq *UnifiedQuery) buildQuery(query *TaskQuery) (string, []interface{}) {
	var args []interface{}
	argIndex := 1

	// SELECT clause with LEFT JOIN for Claude mapping
	sql := `
		SELECT DISTINCT
			t.id, t.title, t.description, t.status, t.group_name,
			t.created_at,
			cm.id, cm.task_id, cm.session_uuid, cm.message_uuid,
			cm.parent_message_uuid, cm.slug, cm.source,
			cm.created_at, cm.updated_at
		FROM tasks t
		LEFT JOIN claude_mappings cm ON t.id = cm.task_id
	`

	// WHERE clause
	whereConditions := []string{}

	// AgentFlow conditions
	if query.TaskID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.id = $%d", argIndex))
		args = append(args, *query.TaskID)
		argIndex++
	}

	if query.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, *query.Status)
		argIndex++
	}

	if query.Group != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.group_name = $%d", argIndex))
		args = append(args, *query.Group)
		argIndex++
	}

	// Claude conditions
	if query.SessionUUID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.session_uuid = $%d", argIndex))
		args = append(args, *query.SessionUUID)
		argIndex++
	}

	if query.MessageUUID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.message_uuid = $%d", argIndex))
		args = append(args, *query.MessageUUID)
		argIndex++
	}

	if query.Slug != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.slug = $%d", argIndex))
		args = append(args, *query.Slug)
		argIndex++
	}

	if query.ParentMessageUUID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.parent_message_uuid = $%d", argIndex))
		args = append(args, *query.ParentMessageUUID)
		argIndex++
	}

	// Add WHERE clause if there are conditions
	if len(whereConditions) > 0 {
		sql += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// ORDER BY clause
	if query.OrderBy != nil {
		orderDirection := "ASC"
		if query.OrderDesc != nil && *query.OrderDesc {
			orderDirection = "DESC"
		}
		sql += fmt.Sprintf(" ORDER BY %s %s", *query.OrderBy, orderDirection)
	} else {
		sql += " ORDER BY t.created_at DESC"
	}

	// LIMIT and OFFSET
	limit := 50
	if query.Limit != nil {
		limit = *query.Limit
	}

	offset := 0
	if query.Offset != nil {
		offset = *query.Offset
	}

	sql += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	return sql, args
}

// buildCountQuery builds the COUNT query
func (uq *UnifiedQuery) buildCountQuery(query *TaskQuery) (string, []interface{}) {
	var args []interface{}
	argIndex := 1

	sql := "SELECT COUNT(DISTINCT t.id) FROM tasks t LEFT JOIN claude_mappings cm ON t.id = cm.task_id"

	// WHERE clause (same as buildQuery)
	whereConditions := []string{}

	if query.TaskID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.id = $%d", argIndex))
		args = append(args, *query.TaskID)
		argIndex++
	}

	if query.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, *query.Status)
		argIndex++
	}

	if query.Group != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.group_name = $%d", argIndex))
		args = append(args, *query.Group)
		argIndex++
	}

	if query.SessionUUID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.session_uuid = $%d", argIndex))
		args = append(args, *query.SessionUUID)
		argIndex++
	}

	if query.MessageUUID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.message_uuid = $%d", argIndex))
		args = append(args, *query.MessageUUID)
		argIndex++
	}

	if query.Slug != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.slug = $%d", argIndex))
		args = append(args, *query.Slug)
		argIndex++
	}

	if query.ParentMessageUUID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cm.parent_message_uuid = $%d", argIndex))
		args = append(args, *query.ParentMessageUUID)
		argIndex++
	}

	if len(whereConditions) > 0 {
		sql += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	return sql, args
}

// GetTaskByClaudeMessageUUID retrieves a task by Claude Message UUID
func (uq *UnifiedQuery) GetTaskByClaudeMessageUUID(messageUUID string) (*UnifiedTask, error) {
	query := &TaskQuery{
		MessageUUID: &messageUUID,
		Limit:       intPtr(1),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		return nil, err
	}

	if len(result.Tasks) == 0 {
		return nil, fmt.Errorf("task not found for message UUID: %s", messageUUID)
	}

	return &result.Tasks[0], nil
}

// GetTasksByClaudeSession retrieves all tasks for a Claude session
func (uq *UnifiedQuery) GetTasksByClaudeSession(sessionUUID string, limit, offset int) (*UnifiedQueryResult, error) {
	query := &TaskQuery{
		SessionUUID: &sessionUUID,
		Limit:       &limit,
		Offset:      &offset,
	}

	return uq.QueryTasks(query)
}

// GetTaskBySlug retrieves a task by Claude slug
func (uq *UnifiedQuery) GetTaskBySlug(slug string) (*UnifiedTask, error) {
	query := &TaskQuery{
		Slug:  &slug,
		Limit: intPtr(1),
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		return nil, err
	}

	if len(result.Tasks) == 0 {
		return nil, fmt.Errorf("task not found for slug: %s", slug)
	}

	return &result.Tasks[0], nil
}

// GetTasksWithClaudeInfo retrieves tasks that have Claude mappings
func (uq *UnifiedQuery) GetTasksWithClaudeInfo(limit, offset int) (*UnifiedQueryResult, error) {
	query := &TaskQuery{
		Limit:  &limit,
		Offset: &offset,
	}

	result, err := uq.QueryTasks(query)
	if err != nil {
		return nil, err
	}

	// Filter only tasks with Claude info
	var tasksWithClaude []UnifiedTask
	for _, task := range result.Tasks {
		if task.Claude != nil {
			tasksWithClaude = append(tasksWithClaude, task)
		}
	}

	result.Tasks = tasksWithClaude
	result.TotalCount = len(tasksWithClaude)

	return result, nil
}

// Helper function
func intPtr(i int) *int {
	return &i
}
