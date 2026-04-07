package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenSQL opens a *sql.DB pool against PostgreSQL using the pgx driver (stdlib adapter).
func OpenSQL(ctx context.Context, dsn string) (*sql.DB, error) {
	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.PingContext(pingCtx); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	// Pool defaults: cap concurrent DB connections and recycle them so stale
	// connections (e.g. after DB failover) are not kept forever.
	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(30 * time.Minute)

	return pool, nil
}
