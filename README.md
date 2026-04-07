# Distributed Workflow Engine

## 🧠 Overview

This project is a **distributed workflow orchestration engine** designed to execute long-running, fault-tolerant processes across services.

It is inspired by systems like Temporal and AWS Step Functions, with a focus on:

* Reliability under failure
* Idempotent execution
* Distributed coordination
* Observability

---

## 🎯 Problem Statement

In distributed systems, coordinating multi-step processes is complex due to:

* Partial failures (a step fails mid-execution)
* Retry handling
* State consistency across services
* Lack of visibility into execution

This project aims to solve:

* Workflow orchestration across distributed workers
* Reliable retries and failure recovery
* State persistence and replayability
* Execution traceability

---

## 🏗️ High-Level Architecture

Core components:

* **API Layer** → receives workflow execution requests
* **Workflow Engine** → orchestrates execution and state transitions
* **Workers** → execute individual steps
* **State Store (PostgreSQL)** → persists workflow state
* **Event Bus (future)** → decouples execution and enables scaling

---

## ⚙️ Current Status

🚧 Early-stage implementation

Currently implemented:

* Basic API service in Go
* Hexagonal architecture structure
* PostgreSQL integration
* Migration system

Planned:

* Workflow execution engine
* Retry + backoff strategies
* Distributed workers
* Event-driven execution

---

## 🧩 Project Goals

* Build a **production-grade workflow engine**
* Support **idempotent and resilient execution**
* Enable **horizontal scaling via workers**
* Provide **observability and debugging tools**
* Simulate real-world distributed failures

---

## 🛠️ Tech Stack

**Current:**

* Go (core service)
* PostgreSQL (state persistence)

**Planned:**

* Kafka / SQS (event-driven execution)
* Redis (caching, locks)
* gRPC (worker communication)
* OpenTelemetry (tracing)
* Prometheus + Grafana (metrics)

---

## 🔥 Technical Challenges

This project intentionally tackles hard problems:

* Exactly-once vs at-least-once execution
* Idempotency guarantees
* Distributed state consistency
* Failure recovery (worker crashes, retries)
* Long-running workflows
* Event ordering and duplication

---

## 📊 Metrics (to be implemented)

* Workflows executed/sec
* Execution latency (p95)
* Retry rate
* Failure recovery success rate
* System throughput

---

## 🧪 Failure Scenarios (planned)

* Worker crashes during execution
* Duplicate step execution
* Database downtime
* Network partitions

---

## 📂 Project Structure

| Path                                    | Purpose                                |
| --------------------------------------- | -------------------------------------- |
| `cmd/api`                               | Application entrypoint                 |
| `internal/domain`                       | Core domain (no external dependencies) |
| `internal/app`                          | Use cases and orchestration            |
| `internal/interfaces/http`              | HTTP adapter                           |
| `internal/infrastructure`               | DB and external integrations           |
| `internal/infrastructure/db/migrations` | SQL migrations                         |

---

## ⚙️ Setup & Local Development

### Prerequisites

* Go 1.26+
* PostgreSQL
* `psql` and `bash`

---

### 1. Clone & install dependencies

```bash
git clone <repository-url>
cd workflow-engine
go mod download
```

---

### 2. Create database

```bash
psql -U postgres -h localhost -c 'CREATE DATABASE "workflow-engine";'
```

---

### 3. Configure environment

```bash
export DATABASE_URL='postgresql://postgres:yourpassword@localhost:5432/workflow-engine'
```

---

### 4. Run migrations

```bash
chmod +x internal/infrastructure/db/migrate-up.sh
./internal/infrastructure/db/migrate-up.sh
```

---

### 5. Run API

```bash
go run ./cmd/api
```

Default:

```
http://localhost:8080
```

Health check:

```bash
curl http://127.0.0.1:8080/health
```

---

## 🌱 Future Work

* Workflow definition DSL
* Distributed worker system
* Retry and compensation logic (Saga pattern)
* Event-driven execution
* Observability (tracing + metrics)
* Load testing and benchmarking

---

## 💬 Pitch

Built a distributed workflow orchestration engine to handle long-running processes with retries, fault tolerance, and state persistence, focusing on reliability and observability in distributed systems.
