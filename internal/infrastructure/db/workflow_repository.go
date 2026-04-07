package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SheykoWk/workflow-engine/internal/infrastructure/db/models"
)

// WorkflowRepository loads and persists workflows using database/sql.
type WorkflowRepository struct {
	db *sql.DB
}

// NewWorkflowRepository returns a repository backed by db.
func NewWorkflowRepository(db *sql.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// GetAllByProjectID returns workflow definitions for one tenant, newest first.
func (r *WorkflowRepository) GetAllByProjectID(ctx context.Context, projectID string) ([]models.Workflow, error) {
	const q = `
		SELECT id, project_id, name, slug, description, created_at, updated_at
		FROM workflows
		WHERE project_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("workflow_repository: query: %w", err)
	}
	defer rows.Close()

	list := make([]models.Workflow, 0)
	for rows.Next() {
		var w models.Workflow
		var desc sql.NullString
		if err := rows.Scan(
			&w.ID,
			&w.ProjectID,
			&w.Name,
			&w.Slug,
			&desc,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("workflow_repository: scan row: %w", err)
		}
		if desc.Valid {
			s := desc.String
			w.Description = &s
		}
		list = append(list, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("workflow_repository: iterate rows: %w", err)
	}
	return list, nil
}
