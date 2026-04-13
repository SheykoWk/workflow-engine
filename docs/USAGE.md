# Usage & API Reference

Base URL: `http://localhost:8080`  
Swagger UI: `http://localhost:8080/swagger/index.html`

---

## Authentication

All endpoints except `POST /projects` and `GET /health` require a Bearer API key.

```
Authorization: Bearer wf_<prefix>.<secret>
```

In Swagger UI you can paste the key directly (without `Bearer `) in the Authorize dialog.

---

## Endpoints

### Health

```
GET /health
```

Returns `ok`. No auth required. Use for liveness probes.

---

### Projects

#### Create project

```
POST /projects
```

Creates a project and a default API key. **The full key is returned once — save it immediately.**

```bash
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "My Project"}'
```

Response `201`:
```json
{
  "project_id": "a1b2c3d4-...",
  "api_key": "wf_aBcDeFgHiJ.xYzSecretToken..."
}
```

---

#### Get current project

```
GET /projects
```

Returns the project associated with the API key.

```bash
curl http://localhost:8080/projects \
  -H "Authorization: Bearer wf_aBcDeFgHiJ.xYzSecretToken..."
```

---

### Workflows

#### Create workflow

```
POST /workflows
```

Defines a workflow with an ordered list of steps. Steps execute sequentially in `step_index` order.

```bash
curl -X POST http://localhost:8080/workflows \
  -H "Authorization: Bearer wf_..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Notify on signup",
    "steps": [
      {
        "name": "fetch-user",
        "type": "http_request",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/users/123"
        }
      },
      {
        "name": "send-email",
        "type": "http_request",
        "config": {
          "method": "POST",
          "url": "https://api.example.com/emails",
          "body": {
            "to": "{{steps.0.output.body.email}}",
            "subject": "Welcome!"
          }
        }
      }
    ]
  }'
```

Response `201`:
```json
{
  "id": "wf-uuid-...",
  "name": "Notify on signup",
  "steps_count": 2
}
```

---

#### List workflows

```
GET /workflows
```

Returns all workflow definitions for the authenticated project.

```bash
curl http://localhost:8080/workflows \
  -H "Authorization: Bearer wf_..."
```

---

#### Trigger a workflow run

```
POST /workflows/{id}/runs
```

Creates a `workflow_run` in status `pending` and one `step_run` per step. The executor picks it up within 1 second.

```bash
curl -X POST http://localhost:8080/workflows/wf-uuid-.../runs \
  -H "Authorization: Bearer wf_..."
```

Response `201`:
```json
{
  "run_id": "run-uuid-...",
  "status": "pending",
  "steps": 2
}
```

---

## Step Types

### `delay`

Waits a fixed number of seconds before proceeding.

```json
{
  "name": "wait-5s",
  "type": "delay",
  "config": {
    "seconds": 5
  }
}
```

---

### `http_request`

Performs an HTTP call. Supports `GET`, `POST`, `PUT`, `DELETE`.

```json
{
  "name": "call-api",
  "type": "http_request",
  "config": {
    "method": "POST",
    "url": "https://api.example.com/notify",
    "headers": {
      "X-API-Key": "secret"
    },
    "body": {
      "message": "hello"
    }
  }
}
```

The step output is always:
```json
{
  "status_code": 200,
  "body": { ... }
}
```

`body` is decoded JSON when the response is valid JSON, otherwise a plain string.

**Success:** HTTP 2xx  
**Failure:** non-2xx — retried unless it's a 4xx (except 429)

---

## Retry Policy

Add a `retry` block to any step config:

```json
{
  "method": "POST",
  "url": "https://api.example.com/webhook",
  "retry": {
    "max_attempts": 3,
    "backoff_seconds": 10
  }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `max_attempts` | 1 | Total attempts including the first |
| `backoff_seconds` | 0 | Base delay. Actual delay = `base * 2^(attempt-1)` ±20% jitter |

**Non-retriable:** HTTP 4xx (except 429), unknown step type, invalid config, context cancellation.

---

## Output Chaining

Use `{{steps.N.output.path}}` in any string value inside a step's `config` to reference the output of a previous step.

`N` is the zero-based `step_index`. `path` is a dot-separated JSON path into that step's output object.

```json
{
  "name": "step-2",
  "type": "http_request",
  "config": {
    "method": "POST",
    "url": "https://api.example.com/items/{{steps.0.output.body.id}}/notify",
    "body": {
      "status": "{{steps.1.output.body.status}}"
    }
  }
}
```

Rules:
- Only outputs from steps with a lower `step_index` than the current step are available.
- Missing paths resolve to an empty string (no error).
- The interpolated config must still be valid JSON — if it isn't, the step fails.

---

## Full Example: Two-Step Workflow

```bash
# 1. Create project
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "demo"}' | jq .

# Save api_key from response, e.g.: wf_abc.xyz
export TOKEN="wf_abc.xyz"

# 2. Create workflow
curl -X POST http://localhost:8080/workflows \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo-flow",
    "steps": [
      {
        "name": "wait",
        "type": "delay",
        "config": {"seconds": 2}
      },
      {
        "name": "ping",
        "type": "http_request",
        "config": {
          "method": "GET",
          "url": "https://httpbin.org/get"
        }
      }
    ]
  }' | jq .

# Save id from response
export WF_ID="wf-uuid-here"

# 3. Trigger run
curl -X POST http://localhost:8080/workflows/$WF_ID/runs \
  -H "Authorization: Bearer $TOKEN" | jq .
```

After triggering, watch the server logs — the executor will log each step transition within 1-2 seconds.
