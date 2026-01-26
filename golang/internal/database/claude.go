package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ClaudeMapping Claude ID 映射
type ClaudeMapping struct {
	ID                int        `json:"id"`
	TaskID            string     `json:"task_id"`
	SessionUUID       string     `json:"session_uuid"`
	MessageUUID       string     `json:"message_uuid"`
	ParentMessageUUID *string   `json:"parent_message_uuid,omitempty"`
	Slug              string     `json:"slug"`
	Source            string     `json:"source"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateClaudeMapping 创建 Claude ID 映射
func (d *Database) CreateClaudeMapping(mapping *ClaudeMapping) error {
	query := `
		INSERT INTO claude_mappings
		(task_id, session_uuid, message_uuid, parent_message_uuid, slug, source)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query,
		mapping.TaskID,
		mapping.SessionUUID,
		mapping.MessageUUID,
		mapping.ParentMessageUUID,
		mapping.Slug,
		mapping.Source,
	)

	if err != nil {
		return fmt.Errorf("failed to create claude mapping: %w", err)
	}

	return nil
}

// GetClaudeMappingByTaskID 通过 AgentFlow Task ID 获取 Claude 信息
func (d *Database) GetClaudeMappingByTaskID(taskID string) (*ClaudeMapping, error) {
	query := `
		SELECT id, task_id, session_uuid, message_uuid,
		       parent_message_uuid, slug, source, created_at, updated_at
		FROM claude_mappings
		WHERE task_id = ?
	`

	row := d.db.QueryRow(query, taskID)

	var mapping ClaudeMapping
	var parentMsg sql.NullString

	err := row.Scan(
		&mapping.ID,
		&mapping.TaskID,
		&mapping.SessionUUID,
		&mapping.MessageUUID,
		&parentMsg,
		&mapping.Slug,
		&mapping.Source,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no claude mapping found for task_id: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get claude mapping: %w", err)
	}

	if parentMsg.Valid {
		mapping.ParentMessageUUID = &parentMsg.String
	}

	return &mapping, nil
}

// GetClaudeMappingByMessageUUID 通过 Claude Message UUID 获取 AgentFlow Task
func (d *Database) GetClaudeMappingByMessageUUID(messageUUID string) (*ClaudeMapping, error) {
	query := `
		SELECT id, task_id, session_uuid, message_uuid,
		       parent_message_uuid, slug, source, created_at, updated_at
		FROM claude_mappings
		WHERE message_uuid = ?
	`

	row := d.db.QueryRow(query, messageUUID)

	var mapping ClaudeMapping
	var parentMsg sql.NullString

	err := row.Scan(
		&mapping.ID,
		&mapping.TaskID,
		&mapping.SessionUUID,
		&mapping.MessageUUID,
		&parentMsg,
		&mapping.Slug,
		&mapping.Source,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no claude mapping found for message_uuid: %s", messageUUID)
		}
		return nil, fmt.Errorf("failed to get claude mapping: %w", err)
	}

	if parentMsg.Valid {
		mapping.ParentMessageUUID = &parentMsg.String
	}

	return &mapping, nil
}

// GetTaskChainBySession 获取同一个 Session 下的所有任务（支持任务链）
func (d *Database) GetTaskChainBySession(sessionUUID string) ([]ClaudeMapping, error) {
	query := `
		SELECT id, task_id, session_uuid, message_uuid,
		       parent_message_uuid, slug, source, created_at, updated_at
		FROM claude_mappings
		WHERE session_uuid = ?
		ORDER BY created_at ASC
	`

	rows, err := d.db.Query(query, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task chain: %w", err)
	}
	defer rows.Close()

	var chain []ClaudeMapping
	for rows.Next() {
		var mapping ClaudeMapping
		var parentMsg sql.NullString

		err := rows.Scan(
			&mapping.ID,
			&mapping.TaskID,
			&mapping.SessionUUID,
			&mapping.MessageUUID,
			&parentMsg,
			&mapping.Slug,
			&mapping.Source,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}

		if parentMsg.Valid {
			mapping.ParentMessageUUID = &parentMsg.String
		}

		chain = append(chain, mapping)
	}

	if len(chain) == 0 {
		return nil, fmt.Errorf("no claude mappings found for session: %s", sessionUUID)
	}

	return chain, nil
}

// GetClaudeMappingBySlug 通过 Slug 获取映射
func (d *Database) GetClaudeMappingBySlug(slug string) (*ClaudeMapping, error) {
	query := `
		SELECT id, task_id, session_uuid, message_uuid,
		       parent_message_uuid, slug, source, created_at, updated_at
		FROM claude_mappings
		WHERE slug = ?
	`

	row := d.db.QueryRow(query, slug)

	var mapping ClaudeMapping
	var parentMsg sql.NullString

	err := row.Scan(
		&mapping.ID,
		&mapping.TaskID,
		&mapping.SessionUUID,
		&mapping.MessageUUID,
		&parentMsg,
		&mapping.Slug,
		&mapping.Source,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no claude mapping found for slug: %s", slug)
		}
		return nil, fmt.Errorf("failed to get claude mapping by slug: %w", err)
	}

	if parentMsg.Valid {
		mapping.ParentMessageUUID = &parentMsg.String
	}

	return &mapping, nil
}

// UpdateClaudeMappingSlug 更新映射的 Slug
func (d *Database) UpdateClaudeMappingSlug(taskID string, slug string) error {
	query := `
		UPDATE claude_mappings
		SET slug = ?, updated_at = CURRENT_TIMESTAMP
		WHERE task_id = ?
	`

	_, err := d.db.Exec(query, slug, taskID)
	if err != nil {
		return fmt.Errorf("failed to update claude mapping slug: %w", err)
	}

	return nil
}

// DeleteClaudeMapping 删除 Claude 映射
func (d *Database) DeleteClaudeMapping(taskID string) error {
	query := `DELETE FROM claude_mappings WHERE task_id = ?`

	result, err := d.db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete claude mapping: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no claude mapping found for task_id: %s", taskID)
	}

	return nil
}

// ListClaudeMappingsBySession 列出指定 Session 的所有映射（分页）
func (d *Database) ListClaudeMappingsBySession(sessionUUID string, limit, offset int) ([]ClaudeMapping, error) {
	query := `
		SELECT id, task_id, session_uuid, message_uuid,
		       parent_message_uuid, slug, source, created_at, updated_at
		FROM claude_mappings
		WHERE session_uuid = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`

	rows, err := d.db.Query(query, sessionUUID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list claude mappings: %w", err)
	}
	defer rows.Close()

	var mappings []ClaudeMapping
	for rows.Next() {
		var mapping ClaudeMapping
		var parentMsg sql.NullString

		err := rows.Scan(
			&mapping.ID,
			&mapping.TaskID,
			&mapping.SessionUUID,
			&mapping.MessageUUID,
			&parentMsg,
			&mapping.Slug,
			&mapping.Source,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}

		if parentMsg.Valid {
			mapping.ParentMessageUUID = &parentMsg.String
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}
