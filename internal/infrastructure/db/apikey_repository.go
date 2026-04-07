package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrAPIKeyNotFound is returned when no active key matches the prefix.
var ErrAPIKeyNotFound = errors.New("api_key_repository: key not found")

// APIKeyRepository persists and looks up API keys.
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository returns a repository backed by db.
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// FindActiveByPrefix returns project_id and key_hash for a non-revoked key with the given prefix.
func (r *APIKeyRepository) FindActiveByPrefix(ctx context.Context, keyPrefix string) (projectID, keyHash string, err error) {
	const q = `
		SELECT project_id, key_hash
		FROM api_keys
		WHERE key_prefix = $1 AND revoked_at IS NULL
		LIMIT 1
	`
	err = r.db.QueryRowContext(ctx, q, keyPrefix).Scan(&projectID, &keyHash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrAPIKeyNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("api_key_repository: find by prefix: %w", err)
	}
	return projectID, keyHash, nil
}

// RevokeKey sets revoked_at for a key belonging to projectID (optional admin use).
func (r *APIKeyRepository) RevokeKey(ctx context.Context, projectID, keyID string) error {
	const q = `
		UPDATE api_keys
		SET revoked_at = now()
		WHERE id = $1 AND project_id = $2 AND revoked_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q, keyID, projectID)
	if err != nil {
		return fmt.Errorf("api_key_repository: revoke: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}
