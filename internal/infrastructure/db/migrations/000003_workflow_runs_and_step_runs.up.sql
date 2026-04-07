-- One row per *invocation* of a workflow definition (a single “job” / saga instance).
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows (id) ON DELETE RESTRICT,
    -- Lifecycle of the whole run (scheduler + worker update this).
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
    -- Trigger payload and final aggregate result (execution data, not the blueprint).
    input JSONB NOT NULL DEFAULT '{}',
    output JSONB,
    error_message TEXT,
    -- Optional tracing id (logs, webhooks, support).
    correlation_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

-- One row per *attempt* to execute a definition step within a given run.
CREATE TABLE IF NOT EXISTS step_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_run_id UUID NOT NULL REFERENCES workflow_runs (id) ON DELETE CASCADE,
    workflow_step_id UUID NOT NULL REFERENCES workflow_steps (id) ON DELETE RESTRICT,
    -- Retry audit trail: attempt 1 fails → insert or advance to attempt 2 (see UNIQUE below).
    attempt INT NOT NULL DEFAULT 1 CHECK (attempt >= 1),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'skipped', 'cancelled')),
    input JSONB NOT NULL DEFAULT '{}',
    output JSONB,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT step_runs_run_step_attempt_key UNIQUE (workflow_run_id, workflow_step_id, attempt)
);

-- List/filter runs for a definition (dashboards, “re-run this workflow”).
CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id ON workflow_runs (workflow_id);

-- Recent runs per definition (newest first) without sorting full table.
CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id_created_at ON workflow_runs (workflow_id, created_at DESC);

-- Scheduler: pick work that is not finished (tune predicates to match your dispatcher).
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status_created_at ON workflow_runs (status, created_at)
    WHERE status IN ('pending', 'running');

-- All step attempts for a run (ordered execution, UI timeline).
CREATE INDEX IF NOT EXISTS idx_step_runs_workflow_run_id ON step_runs (workflow_run_id);

-- Worker: next pending step for a run (optional; depends on how you dequeue).
CREATE INDEX IF NOT EXISTS idx_step_runs_run_status ON step_runs (workflow_run_id, status)
    WHERE status IN ('pending', 'running');
