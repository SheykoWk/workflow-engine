-- Baseline 000003 already defines output; this migration is safe on fresh and legacy DBs.
ALTER TABLE step_runs
ADD COLUMN IF NOT EXISTS output JSONB;
