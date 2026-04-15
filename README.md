# Distributed Workflow Orchestration Engine

A self-hosted distributed workflow orchestration engine built in Go for executing long-running, fault-tolerant multi-step processes with durable state persistence, retry semantics, exponential backoff, and inter-step output chaining.

Inspired by Temporal and AWS Step Functions — designed as a lightweight, infrastructure-minimal alternative that can run with only PostgreSQL.

---

## Why This Exists

Modern distributed systems frequently require orchestration of multi-step workflows such as:

- payment processing pipelines
- onboarding automations
- asynchronous SaaS provisioning
- multi-service approval chains
- retryable integration workflows

Most teams solve this repeatedly inside application code, creating:

- duplicated retry logic
- fragile failure recovery
- inconsistent orchestration behavior
- no centralized execution visibility

This engine solves that problem by providing a reusable orchestration layer.

---

## Core Capabilities

### Durable Workflow Execution
Workflow state is persisted in PostgreSQL and survives crashes or restarts.

### Retry with Exponential Backoff + Jitter
Automatic retry handling with:
- exponential backoff
- ±20% jitter
- capped retry windows

### Multi-Step Dependency Sequencing
Steps execute only when all prior dependencies succeed.

### Multi-Tenant Isolation
Every workflow, run, and API request is tenant-scoped.

### Inter-Step Output Chaining
Downstream steps can dynamically consume outputs from previous steps:

```json
{
  "url": "{{steps.1.output.data.url}}"
}
````

### Idempotent Execution Guarantees

Duplicate step attempts prevented via DB-level unique constraints.

### Distributed Worker Architecture

Worker processes run independently from API server.

---

## Architecture Highlights

This project follows **Hexagonal Architecture (Ports & Adapters)**:

* Domain layer isolated from infrastructure
* PostgreSQL persistence adapter
* HTTP adapter for API interface
* Executor adapter for workflow processing

### Key Design Principles:

* Domain-first modeling
* Infrastructure independence
* Atomic transactional writes
* Explicit retry classification
* Stateless worker scalability model

---

## Tech Stack

| Layer        | Technology             |
| ------------ | ---------------------- |
| Language     | Go 1.26                |
| Database     | PostgreSQL             |
| DB Driver    | pgx/v5                 |
| Auth         | bcrypt API keys        |
| HTTP         | net/http stdlib        |
| API Docs     | Swagger (swaggo)       |
| Architecture | Hexagonal Architecture |

---

## Current Supported Step Types

### 1. HTTP Request Step

Supports:

* method validation
* custom headers
* request body
* timeout control
* response parsing

### 2. Delay Step

Supports:

* cancellable timed waits
* delayed orchestration scheduling

---

## Project Maturity

### Current Level:

Senior-level distributed systems MVP

Implemented:

* persistent workflow state machine
* retry engine
* workflow sequencing
* output chaining
* tenant isolation
* distributed worker binary

---

## Known Current Limitations

Not yet implemented:

* distributed row locking (`FOR UPDATE SKIP LOCKED`)
* parallel step execution concurrency
* saga compensation / rollback flows
* conditional branching steps
* structured logging / tracing
* Prometheus metrics

These are documented in roadmap as production-grade upgrades.

---

## Quick Start

### 1. Install dependencies

```bash
go mod download
```

### 2. Create database

```bash
psql -U postgres -h localhost -c 'CREATE DATABASE "workflow-engine";'
```

### 3. Configure environment

```bash
cp .env.example .env
```

### 4. Run migrations

```bash
./internal/infrastructure/db/migrate-up.sh
```

### 5. Start API server

```bash
go run ./cmd/api
```

API available at:

```bash
http://localhost:8080
```

Swagger UI:

```bash
http://localhost:8080/swagger/index.html
```

---

## Running Worker Process

Start worker executor separately:

```bash
go run ./cmd/worker
```

This process polls pending workflow steps and executes them independently.

---

## Environment Variables

| Variable     | Required | Default | Description                  |
| ------------ | -------- | ------- | ---------------------------- |
| DATABASE_URL | Yes      | —       | PostgreSQL connection string |
| HTTP_ADDR    | No       | :8080   | HTTP listen address          |

---

## Example Workflow Execution

Example:

1. Create workflow definition
2. Trigger workflow run
3. Worker picks pending steps
4. Executes sequentially
5. Retries failures automatically
6. Marks workflow complete

---

## Example Use Cases

This engine is ideal for:

* SaaS tenant provisioning
* payment retry orchestration
* background integration pipelines
* async approval workflows
* compliance automation processes

---

## Documentation

| Document                             | Description                               |
| ------------------------------------ | ----------------------------------------- |
| [Architecture](docs/ARCHITECTURE.md) | Internal architecture and execution model |
| [Usage & API](docs/USAGE.md)         | Endpoints and examples                    |
| [Contributing](docs/CONTRIBUTING.md) | How to extend the engine                  |
| [Post-MVP Roadmap](docs/POST_MVP.md) | Planned production-grade enhancements     |

---

## Roadmap to Production Grade

Next major upgrades:

### Critical

* Add `SELECT FOR UPDATE SKIP LOCKED`
* Prevent multi-worker race conditions

### Scalability

* Parallel executor concurrency
* Worker pool scheduling

### Observability

* Structured logging (slog)
* Prometheus metrics
* OpenTelemetry tracing

### Workflow Features

* Conditional branching
* Saga compensation handlers
* Script execution steps

---

## Engineering Highlights

This project demonstrates:

* distributed systems orchestration design
* persistent workflow state machines
* retry semantics engineering
* idempotent execution guarantees
* scalable worker separation patterns

---

## Why This Project Matters

This is not a CRUD application.

It is infrastructure software:
a reusable orchestration engine solving a real distributed systems problem.

It demonstrates senior-level backend engineering in:

* systems design
* fault tolerance
* execution reliability
* architectural separation

---

## Author

Built as a distributed systems engineering project focused on resilient orchestration patterns in modern backend architectures.

Sahid Kick  
Senior Backend Engineer

GitHub: https://github.com/SheykoWk