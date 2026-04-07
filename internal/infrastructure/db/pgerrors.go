package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation reports whether err is a PostgreSQL unique_violation (23505).
func IsUniqueViolation(err error) bool {
	var pe *pgconn.PgError
	return errors.As(err, &pe) && pe.Code == "23505"
}
