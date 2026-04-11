-- When a step is scheduled for retry, it stays pending until next_run_at.
ALTER TABLE step_runs
ADD COLUMN IF NOT EXISTS next_run_at TIMESTAMPTZ;
