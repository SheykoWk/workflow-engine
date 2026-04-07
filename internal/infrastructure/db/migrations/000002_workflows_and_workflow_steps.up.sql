-- Workflow *definition*: reusable blueprint owned by a project (tenant).
-- Runtime *execution* (runs, per-step state, retries) lives in separate tables later.
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT workflows_project_slug_key UNIQUE (project_id, slug)
);

-- Ordered steps belonging to one workflow definition (template graph as a linear sequence for v1).
CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    -- Order within this workflow; stable ordering for linear pipelines (0, 1, 2, …).
    step_index INT NOT NULL,
    name TEXT NOT NULL,
    -- Discriminator for the executor (e.g. http_request, delay, script).
    step_type TEXT NOT NULL,
    -- Definition-only payload: timeouts, URLs, static parameters (not run outputs).
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT workflow_steps_step_index_check CHECK (step_index >= 0),
    CONSTRAINT workflow_steps_workflow_step_index_key UNIQUE (workflow_id, step_index)
);

CREATE INDEX IF NOT EXISTS idx_workflows_project_id ON workflows (project_id);

CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow_id ON workflow_steps (workflow_id);
