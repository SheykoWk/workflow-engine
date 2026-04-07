-- One lookup prefix per key for Bearer auth (wf_<prefix>.<secret> stores prefix only in key_prefix).
CREATE UNIQUE INDEX IF NOT EXISTS api_keys_key_prefix_unique
    ON api_keys (key_prefix)
    WHERE key_prefix IS NOT NULL;
