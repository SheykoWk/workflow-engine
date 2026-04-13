# Workflow Engine

A distributed workflow orchestration engine in Go — inspired by Temporal and AWS Step Functions.

Designed to execute long-running, fault-tolerant multi-step processes with retries, state persistence, and output chaining between steps.

---

## Quick Start

```bash
# 1. Dependencies
go mod download

# 2. Create database
psql -U postgres -h localhost -c 'CREATE DATABASE "workflow-engine";'

# 3. Environment
cp .env.example .env   # or export DATABASE_URL manually

# 4. Migrations
./internal/infrastructure/db/migrate-up.sh

# 5. Run
go run ./cmd/api
# → http://localhost:8080
# → http://localhost:8080/swagger/index.html
```

Override listen address: `HTTP_ADDR=:3000 go run ./cmd/api`

---

## Environment Variables

| Variable       | Required | Default | Description                        |
|----------------|----------|---------|------------------------------------|
| `DATABASE_URL` | Yes      | —       | PostgreSQL connection string        |
| `HTTP_ADDR`    | No       | `:8080` | HTTP listen address                |

---

## Documentation

| Doc | Description |
|-----|-------------|
| [Architecture](docs/ARCHITECTURE.md) | Hexagonal layout, domain model, executor flow, DB schema |
| [Usage & API](docs/USAGE.md) | Endpoints, auth, step types, curl examples |
| [Adding Endpoints](docs/CONTRIBUTING.md) | Step-by-step guide for adding new routes |
| [Post-MVP Roadmap](docs/POST_MVP.md) | Ideas and checklist for future development |

Swagger UI available at `/swagger/index.html` when running locally.

---

## Tech Stack

| Layer | Current |
|-------|---------|
| Language | Go 1.26 |
| Database | PostgreSQL + pgx/v5 |
| Auth | bcrypt API keys |
| HTTP | `net/http` stdlib |
| Docs | swaggo/swag |

---

## Project Status

MVP complete. The engine can define workflows, trigger runs, and execute steps in-process with retry and output chaining. See [Post-MVP Roadmap](docs/POST_MVP.md) for what comes next.
