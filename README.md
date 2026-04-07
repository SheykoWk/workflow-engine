# workflow-engine

Go API service (hexagonal layout) with PostgreSQL. This document covers how to set up your environment, apply database migrations, and run the server.

## Prerequisites

- **Go** 1.26 or newer (see `go.mod`).
- **PostgreSQL** running locally or reachable over the network.
- **`psql`** and **`bash`** on your `PATH` (used by the migration script).

## 1. Clone and dependencies

```bash
git clone <repository-url>
cd workflow-engine
go mod download
```

## 2. Create the database

Create an empty database for the app (name is up to you; examples below use `workflow-engine`).

Using `psql` as a superuser:

```bash
psql -U postgres -h localhost -c 'CREATE DATABASE "workflow-engine";'
```

If your database name uses only letters, numbers, and underscores, you can omit the double quotes.

## 3. Configure `DATABASE_URL`

The migration script reads **`DATABASE_URL`** (PostgreSQL connection URI).

Example:

```bash
export DATABASE_URL='postgresql://postgres:yourpassword@localhost:5432/workflow-engine'
```

Adjust user, password, host, port, and database name to match your setup.

## 4. Run migrations

Migrations live under `internal/infrastructure/db/migrations/` (paired `.up.sql` / `.down.sql` files). They are written to be **re-runnable** in simple setups (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).

From the repository root:

```bash
chmod +x internal/infrastructure/db/migrate-up.sh   # first time only, if needed
./internal/infrastructure/db/migrate-up.sh
```

The script applies every `*.up.sql` file in sorted order. It stops on the first error (`ON_ERROR_STOP` in `psql`).

To tear down schema in reverse dependency order, run the `.down.sql` files manually (newest first), or extend the repo with a `migrate-down` script if you need it regularly.

## 5. Run the API

```bash
go run ./cmd/api
```

By default the server listens on **`:8080`**. Override with:

```bash
export HTTP_ADDR=:3000
go run ./cmd/api
```

Health check:

```bash
curl -s http://127.0.0.1:8080/health
```

## Environment variables

| Variable        | Required for migrations | Description                                      |
|----------------|-------------------------|--------------------------------------------------|
| `DATABASE_URL` | Yes                     | PostgreSQL connection string for `migrate-up.sh` |
| `HTTP_ADDR`    | No                      | Listen address for the API (default `:8080`)      |

## Project layout (short)

| Path | Purpose |
|------|---------|
| `cmd/api` | Application entrypoint |
| `internal/domain` | Core domain (no framework imports) |
| `internal/app` | Use cases and orchestration |
| `internal/interfaces/http` | HTTP adapter (`net/http`) |
| `internal/infrastructure` | DB, external systems; SQL under `internal/infrastructure/db/migrations/` |
