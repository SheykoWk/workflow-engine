# Post-MVP Roadmap

Ideas and improvements beyond the current MVP. Grouped by category and priority.

---

## API Completeness

- [ ] `GET /workflows/{id}` — fetch single workflow definition
- [ ] `GET /workflows/{id}/runs` — list runs for a workflow (with status filter + pagination)
- [ ] `GET /runs/{id}` — get run detail: status, started_at, completed_at, step-level timeline
- [ ] `GET /runs/{id}/steps` — list all step_runs with their outputs and attempt history
- [ ] `POST /runs/{id}/cancel` — cancel a pending or running workflow run
- [ ] `POST /runs/{id}/retry` — re-trigger a failed run from scratch
- [ ] `DELETE /workflows/{id}` — soft-delete a workflow (prevent new runs, keep history)
- [ ] `GET /projects/api-keys` — list active API keys for the project
- [ ] `DELETE /projects/api-keys/{id}` — revoke an API key

---

## Observability

- [ ] Structured logging — replace `log.Printf` with `slog` (key-value pairs, JSON output)
- [ ] Request/response logging middleware — method, path, status, latency
- [ ] Correlation ID propagation — generate or forward `X-Request-ID` through the full execution chain
- [ ] OpenTelemetry tracing — spans for run creation, step execution, DB queries
- [ ] Prometheus metrics — `workflow_runs_total`, `step_runs_total`, `step_duration_seconds`, `retry_total`
- [ ] Grafana dashboard — latency p50/p95, error rate, retry rate, queue depth

---

## Executor Improvements

- [ ] Parallel step execution — allow steps at the same `step_index` to run concurrently
- [ ] Step timeout — configurable per-step max execution duration
- [ ] Conditional steps — skip a step based on output from a previous step
- [ ] `script` step type — run an inline shell command or embedded JS/Lua
- [ ] `wait_for_event` step type — pause execution until an external webhook call resumes the run
- [ ] Dead letter queue — move permanently failed runs to a separate table for inspection
- [ ] Executor health metrics — steps processed/sec, queue depth, lag

---

## Reliability & Scaling

- [ ] Distributed worker locking — use `SELECT ... FOR UPDATE SKIP LOCKED` to prevent duplicate execution across multiple worker instances
- [ ] Horizontal worker scaling — run `cmd/worker` as multiple replicas against the same DB
- [ ] Redis-backed queue — move step dispatch off poll loop onto a push queue (lower latency)
- [ ] Kafka / SQS integration — event-driven execution triggered by external messages
- [ ] Graceful worker drain — on SIGTERM, finish the current step before exiting
- [ ] Circuit breaker on `http_request` steps — stop hammering a repeatedly failing external service

---

## Developer Experience

- [ ] `GET /runs/{id}/replay` — re-run a workflow from a specific step, reusing previous outputs
- [ ] Workflow versioning — allow updating a workflow definition without breaking in-progress runs
- [ ] Dry-run mode — validate and simulate a workflow run without persisting or executing
- [ ] OpenAPI SDK generation — auto-generate a typed Go/TypeScript client from the Swagger spec
- [ ] CLI tool — `wfctl run trigger <workflow-id>`, `wfctl runs list`, `wfctl runs logs <run-id>`

---

## Security & Operations

- [ ] API key scopes — read-only vs read-write keys per project
- [ ] Rate limiting — per-project request rate limits on the public API
- [ ] `POST /projects` rate limit — prevent API key brute-force via project creation
- [ ] Secrets management — encrypt sensitive step config values at rest (e.g., API keys inside `http_request` headers)
- [ ] Audit log — record who triggered runs, when, and with what input
- [ ] mTLS between `cmd/api` and `cmd/worker` for internal traffic

---

## Testing

- [ ] Integration test suite — spin up a real PostgreSQL instance and exercise the full stack end-to-end
- [ ] Executor unit tests — mock `StepRunRepository` and test state transitions
- [ ] HTTP handler tests — `httptest.NewRecorder` + table-driven tests per endpoint
- [ ] Load test — `k6` or `hey` baseline for throughput and latency under concurrent runs
- [ ] Chaos testing — kill the worker mid-execution and verify the run resumes correctly

---

## Infrastructure

- [ ] Dockerfile — multi-stage build for `cmd/api` and `cmd/worker`
- [ ] Docker Compose — local dev with PostgreSQL, API, and worker in one command
- [ ] GitHub Actions CI — build + vet + test on every PR
- [ ] Kubernetes manifests — Deployment + Service + HPA for the API and worker
- [ ] Helm chart — parameterized deployment for multiple environments
- [ ] Database connection pooling via PgBouncer — reduce connection overhead at scale

---

## Priority Order (suggested)

| Priority | Item |
|----------|------|
| 1 | `GET /runs/{id}` + step timeline (needed to verify execution) |
| 2 | Structured logging with `slog` |
| 3 | `SELECT ... FOR UPDATE SKIP LOCKED` for safe multi-worker execution |
| 4 | Integration test suite |
| 5 | Prometheus metrics + Grafana dashboard |
| 6 | Dockerfile + Docker Compose |
| 7 | Parallel steps + step timeout |
| 8 | `wait_for_event` step type |
| 9 | GitHub Actions CI |
| 10 | Kubernetes / Helm |
