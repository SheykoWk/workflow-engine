#!/usr/bin/env bash
# Apply all SQL migrations in migrations/ (next to this script) to PostgreSQL.
# Usage (from anywhere):
#   export DATABASE_URL='postgresql://postgres:pass@localhost:5432/workflow-engine'
#   ./internal/infrastructure/db/migrate-up.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/migrations"

if [[ -z "${DATABASE_URL:-}" ]]; then
	echo "Error: DATABASE_URL is not set (PostgreSQL connection string)." >&2
	echo "Example: export DATABASE_URL='postgresql://postgres:pass@localhost:5432/workflow-engine'" >&2
	exit 1
fi

found=0
while IFS= read -r f; do
	[[ -z "$f" ]] && continue
	found=1
	echo "==> Applying $f"
	psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done < <(find "$MIGRATIONS_DIR" -maxdepth 1 -name '*.up.sql' | sort)

if [[ "$found" -eq 0 ]]; then
	echo "Error: no *.up.sql files found in $MIGRATIONS_DIR" >&2
	exit 1
fi

echo "Done: all migrations applied."
