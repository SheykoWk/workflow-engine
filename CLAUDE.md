# CLAUDE.md — Workflow Engine

Context and guidance for working on this codebase.

---

## What This Project Is

A **distributed workflow orchestration engine** written in Go. The goal is to execute long-running, fault-tolerant processes across services — inspired by Temporal and AWS Step Functions.

Core problems it solves:
- Multi-step process coordination across services
- Reliable retries and failure recovery
- State persistence and replayability
- Execution traceability

**Status: early-stage alpha.** API and data layers are working. The execution engine (dequeue → execute → update state) is not yet implemented.

---

## Architecture

Hexagonal (ports & adapters). Dependencies flow inward — domain has no external imports.

```
cmd/api/                        → entrypoint, wires everything together
internal/domain/                → core types (no external dependencies)
internal/app/                   → use cases, orchestration (placeholder today)
internal/interfaces/http/       → HTTP adapter (handlers, router, middleware)
internal/infrastructure/db/     → PostgreSQL adapter (repositories, migrations)
internal/auth/                  → API key generation, hashing, middleware
docs/                           → generated Swagger/OpenAPI specs
```

### Multi-tenancy

Everything is scoped to a **Project**. A project is the tenant root. Auth happens via a Bearer API key which is resolved to a `project_id` and injected into the request context. All queries are project-scoped.

### Authentication Flow

1. `POST /projects` creates a project + a default API key (full key returned once, never stored)
2. All protected routes require `Authorization: Bearer wf_<prefix>.<secret>`
3. Middleware extracts prefix → looks up key hash by prefix → bcrypt compare → injects `project_id` into context

API key format: `wf_<9-byte-base64url-prefix>.<32-byte-base64url-secret>`

---

## Domain Model

| Entity | Purpose |
|---|---|
| `Project` | Tenant root |
| `APIKey` | Auth credential scoped to a project |
| `Workflow` | Blueprint definition with ordered steps |
| `WorkflowStep` | One step in a blueprint (type + JSONB config) |
| `WorkflowRun` | A single execution instance of a Workflow |
| `StepRun` | A single attempt at executing one WorkflowStep |

`StepRun.attempt` tracks retries. `(workflow_run_id, workflow_step_id, attempt)` is unique — this is how idempotency is enforced at the DB level.

### Status Enums

`WorkflowRun.status`: `pending | running | succeeded | failed | cancelled`
`StepRun.status`: `pending | running | succeeded | failed | skipped | cancelled`

These are custom Go types in `internal/domain/run_status.go` with `database/sql` scanner support.

---

## Current API Surface

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | No | Liveness probe |
| POST | `/projects` | No | Create project + default API key |
| GET | `/projects` | Yes | Get authenticated project |
| GET | `/workflows` | Yes | List workflows for project |
| GET | `/swagger/*` | No | Swagger UI |

**Missing (planned next):**
- `POST /workflows` — create workflow + steps
- `POST /workflows/:id/runs` — trigger a workflow run
- `GET /workflows/:id/runs` — list runs
- `GET /runs/:id` — get run detail + step results

---

## Database

PostgreSQL. Migrations live in `internal/infrastructure/db/migrations/` as paired `.up.sql` / `.down.sql` files. Applied with `./internal/infrastructure/db/migrate-up.sh`.

| Migration | Creates |
|---|---|
| 000001 | `projects`, `api_keys` |
| 000002 | `workflows`, `workflow_steps` |
| 000003 | `workflow_runs`, `step_runs` |
| 000004 | Unique index on `api_keys.key_prefix` |

All queries use prepared statements. Repositories follow the pattern: interface in `internal/app/` (not yet formalized), concrete implementation in `internal/infrastructure/db/`.

---

## Tech Stack

| Layer | Current | Planned |
|---|---|---|
| Language | Go 1.26 | — |
| DB | PostgreSQL + pgx/v5 | — |
| Auth | bcrypt API keys | — |
| HTTP | `net/http` stdlib | — |
| Docs | swaggo/swag | — |
| Messaging | — | Kafka / SQS |
| Caching/Locks | — | Redis |
| Worker RPC | — | gRPC |
| Tracing | — | OpenTelemetry |
| Metrics | — | Prometheus + Grafana |

---

## Current Status & What's Missing

### Working
- Project + API key provisioning
- Bearer auth middleware with bcrypt key verification
- Workflow and step schema (creation not yet exposed via API)
- DB connection pooling, graceful shutdown, migration system

### Not Yet Implemented

#### 1. Workflow / Step Creation API
`POST /workflows` needs to accept a workflow definition with steps array, validate step order continuity, and persist atomically in a transaction.

#### 2. Run Trigger + Execution Engine
This is the core missing piece. High-level flow:
```
POST /workflows/:id/runs
  → insert workflow_run (status=pending)
  → insert step_runs for each step (status=pending)
  → dispatch to executor (goroutine or worker queue)

Executor:
  → dequeue pending step_run
  → mark running
  → execute step (HTTP call, script, delay, etc.)
  → mark succeeded/failed
  → advance to next step or fail the run
```

#### 3. Retry / Backoff Logic
Schema already supports `attempt` tracking. Need: max attempt config per step, exponential backoff, and a scheduler that re-enqueues failed steps.

#### 4. Worker Distribution
Today everything would run in-process. For scale: workers poll a queue (Redis BRPOP or Kafka topic), claim a step run, execute, report back.

#### 5. Step Types
`workflow_steps.step_type` is a discriminator. Needs executors per type:
- `http_request` — call an external URL
- `delay` — wait N seconds
- `script` — run inline code or subprocess
- `condition` — branch based on previous output

#### 6. Observability
- Structured logging (replace `log.Printf` with `slog`)
- Request/response logging middleware
- OpenTelemetry tracing
- Prometheus metrics (runs/sec, latency p95, retry rate)

---

## Suggested Next Steps (in order)

1. **Workflow creation endpoint** — `POST /workflows` with steps, validation, atomic insert
2. **Run trigger** — `POST /workflows/:id/runs`, create `workflow_run` + `step_runs`, return run ID
3. **In-process executor** — goroutine-based, supports `http_request` and `delay` step types
4. **Run status polling** — `GET /runs/:id` with step-level detail
5. **Retry logic** — configurable max attempts + exponential backoff per step
6. **Structured logging** — swap to `slog`, add request logging middleware
7. **Worker abstraction** — extract executor behind an interface so it can be swapped for a distributed queue later
8. **OpenTelemetry** — add trace spans around run + step execution

---

## Conventions

- **Error wrapping**: always `fmt.Errorf("context: %w", err)`
- **Sentinel errors**: `ErrXxxNotFound` in each repository for clean 404 handling
- **HTTP responses**: use `respond.JSON(w, status, payload)` and `respond.Error(w, status, message)` from `internal/interfaces/http/respond.go`
- **Auth context**: inject with `auth.WithProjectID(ctx, id)`, read with `auth.ProjectIDFromContext(ctx)`
- **Slug generation**: lowercase alphanumeric + hyphens, with random suffix to prevent collisions
- **JSON decoding**: always use `DisallowUnknownFields()` on decoders
- **SQL**: always parameterized (`$1`, `$2`, ...); never string-interpolated queries

---

## Local Setup

```bash
# 1. Install deps
go mod download

# 2. Create DB
psql -U postgres -h localhost -c 'CREATE DATABASE "workflow-engine";'

# 3. Configure env
export DATABASE_URL='postgresql://postgres:yourpassword@localhost:5432/workflow-engine'

# 4. Run migrations
./internal/infrastructure/db/migrate-up.sh

# 5. Run server
go run ./cmd/api
# → http://localhost:8080
# → http://localhost:8080/swagger/index.html
```

Override listen address: `export HTTP_ADDR=:3000`

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `HTTP_ADDR` | No | `:8080` | HTTP listen address |
