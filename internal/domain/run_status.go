package domain

import (
	"database/sql/driver"
	"fmt"
)

// WorkflowStatus is the lifecycle state of a workflow_runs row (matches DB CHECK).
type WorkflowStatus string

const (
	WorkflowPending   WorkflowStatus = "pending"
	WorkflowRunning   WorkflowStatus = "running"
	WorkflowSucceeded WorkflowStatus = "succeeded"
	WorkflowFailed    WorkflowStatus = "failed"
	WorkflowCancelled WorkflowStatus = "cancelled"
)

func (s *WorkflowStatus) Scan(src any) error {
	if src == nil {
		*s = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		*s = WorkflowStatus(v)
	case []byte:
		*s = WorkflowStatus(v)
	default:
		return fmt.Errorf("domain.WorkflowStatus: cannot scan %T", src)
	}
	return nil
}

func (s WorkflowStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// StepRunStatus is the lifecycle state of a step_runs row (matches DB CHECK).
type StepRunStatus string

const (
	StepPending   StepRunStatus = "pending"
	StepRunning   StepRunStatus = "running"
	StepSucceeded StepRunStatus = "succeeded"
	StepFailed    StepRunStatus = "failed"
	StepSkipped   StepRunStatus = "skipped"
	StepCancelled StepRunStatus = "cancelled"
)

func (s *StepRunStatus) Scan(src any) error {
	if src == nil {
		*s = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		*s = StepRunStatus(v)
	case []byte:
		*s = StepRunStatus(v)
	default:
		return fmt.Errorf("domain.StepRunStatus: cannot scan %T", src)
	}
	return nil
}

func (s StepRunStatus) Value() (driver.Value, error) {
	return string(s), nil
}
