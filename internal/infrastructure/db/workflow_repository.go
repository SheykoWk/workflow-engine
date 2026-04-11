package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SheykoWk/workflow-engine/internal/infrastructure/db/models"
)

// ErrWorkflowNotFound means the workflow id is missing or not visible to this project.
var ErrWorkflowNotFound = errors.New("workflow_repository: workflow not found")

// WorkflowStepInsert is one row for workflow_steps inside CreateWithSteps.
type WorkflowStepInsert struct {
	Name     string
	StepType string
	Config   []byte
}

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

// CreateWithSteps inserts a workflow and its steps in one transaction.
func (r *WorkflowRepository) CreateWithSteps(ctx context.Context, projectID, name, slug string, steps []WorkflowStepInsert) (workflowID string, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("workflow_repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const insertWF = `
		INSERT INTO workflows (project_id, name, slug)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	if err := tx.QueryRowContext(ctx, insertWF, projectID, name, slug).Scan(&workflowID); err != nil {
		return "", fmt.Errorf("workflow_repository: insert workflow: %w", err)
	}

	const insertStep = `
		INSERT INTO workflow_steps (workflow_id, step_index, name, step_type, config)
		VALUES ($1, $2, $3, $4, $5)
	`
	for i, st := range steps {
		if _, err := tx.ExecContext(ctx, insertStep, workflowID, i, st.Name, st.StepType, st.Config); err != nil {
			return "", fmt.Errorf("workflow_repository: insert step %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("workflow_repository: commit: %w", err)
	}
	return workflowID, nil
}

// CreateWorkflowRunWithStepRuns creates a workflow_run (pending) and one step_run per definition step,
// all in one transaction. Verifies workflows.project_id matches projectID.
func (r *WorkflowRepository) CreateWorkflowRunWithStepRuns(ctx context.Context, projectID, workflowID string) (runID string, stepCount int, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, fmt.Errorf("workflow_repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var wfProject string
	err = tx.QueryRowContext(ctx, `SELECT project_id FROM workflows WHERE id = $1`, workflowID).Scan(&wfProject)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, ErrWorkflowNotFound
	}
	if err != nil {
		return "", 0, fmt.Errorf("workflow_repository: load workflow: %w", err)
	}
	if wfProject != projectID {
		return "", 0, ErrWorkflowNotFound
	}

	const insertRun = `
		INSERT INTO workflow_runs (workflow_id, status, input)
		VALUES ($1, 'pending', '{}'::jsonb)
		RETURNING id
	`
	if err := tx.QueryRowContext(ctx, insertRun, workflowID).Scan(&runID); err != nil {
		return "", 0, fmt.Errorf("workflow_repository: insert workflow_run: %w", err)
	}

	const listSteps = `
		SELECT id FROM workflow_steps WHERE workflow_id = $1 ORDER BY step_index ASC
	`
	rows, err := tx.QueryContext(ctx, listSteps, workflowID)
	if err != nil {
		return "", 0, fmt.Errorf("workflow_repository: list workflow_steps: %w", err)
	}
	var stepIDs []string
	for rows.Next() {
		var stepID string
		if err := rows.Scan(&stepID); err != nil {
			rows.Close()
			return "", 0, fmt.Errorf("workflow_repository: scan step id: %w", err)
		}
		stepIDs = append(stepIDs, stepID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return "", 0, fmt.Errorf("workflow_repository: iterate steps: %w", err)
	}
	rows.Close()

	const insertStepRun = `
		INSERT INTO step_runs (workflow_run_id, workflow_step_id, attempt, status, input)
		VALUES ($1, $2, 1, 'pending', '{}'::jsonb)
	`
	for _, stepID := range stepIDs {
		if _, err := tx.ExecContext(ctx, insertStepRun, runID, stepID); err != nil {
			return "", 0, fmt.Errorf("workflow_repository: insert step_run: %w", err)
		}
		stepCount++
	}

	if err := tx.Commit(); err != nil {
		return "", 0, fmt.Errorf("workflow_repository: commit run: %w", err)
	}
	return runID, stepCount, nil
}
