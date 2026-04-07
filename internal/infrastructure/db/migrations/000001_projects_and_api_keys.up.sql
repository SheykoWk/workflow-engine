-- Multi-tenant boundary: one row per tenant/workspace. API keys always belong
-- to exactly one project (enforced by FK).
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    -- Stable external identifier (URLs, CLI); unique so tenants never collide.
    slug TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT projects_slug_key UNIQUE (slug)
);

-- Hashed secret only: raw key exists once at creation, then bcrypt/argon2 only.
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    -- Human label in dashboard (not the secret).
    name TEXT NOT NULL,
    -- Output of bcrypt/argon2 (TEXT avoids length surprises).
    key_hash TEXT NOT NULL,
    -- Optional: first chars of the key shown at creation (e.g. sk_live_ab…).
    key_prefix TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    CONSTRAINT api_keys_key_hash_key UNIQUE (key_hash)
);

-- UNIQUE (slug) creates a unique index on projects.slug for lookups.

-- FK column: speeds JOINs, “keys for project”, and parent DELETE/UPDATE on children.
CREATE INDEX IF NOT EXISTS idx_api_keys_project_id ON api_keys (project_id);

-- Fast auth: lookup by hash on every authenticated request. UNIQUE (key_hash)
-- already creates a B-tree index; no extra index on key_hash needed.

-- List active keys per tenant without scanning revoked rows.
CREATE INDEX IF NOT EXISTS idx_api_keys_project_id_active ON api_keys (project_id)
    WHERE revoked_at IS NULL;
