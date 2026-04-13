# Adding New Endpoints

A practical guide to adding a new route to the API. Follow this pattern consistently.

---

## Anatomy of an Endpoint

Every endpoint touches 4 layers in this order (outer → inner):

```
router.go  →  handler  →  app service  →  repository
```

---

## Step-by-Step Example

We'll add `GET /workflows/{id}` — fetch a single workflow by ID.

---

### 1. Add the DB query (repository)

`internal/infrastructure/db/workflow_repository.go`

```go
// GetByID returns a workflow if it belongs to the given project.
func (r *WorkflowRepository) GetByID(ctx context.Context, projectID, workflowID string) (*models.Workflow, error) {
    const q = `
        SELECT id, project_id, name, slug, description, created_at, updated_at
        FROM workflows
        WHERE id = $1 AND project_id = $2
    `
    var w models.Workflow
    var desc sql.NullString
    err := r.db.QueryRowContext(ctx, q, workflowID, projectID).Scan(
        &w.ID, &w.ProjectID, &w.Name, &w.Slug, &desc, &w.CreatedAt, &w.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrWorkflowNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("workflow_repository: get by id: %w", err)
    }
    if desc.Valid {
        s := desc.String
        w.Description = &s
    }
    return &w, nil
}
```

Rules:
- Always use parameterized queries (`$1`, `$2`). Never interpolate strings into SQL.
- Return sentinel errors (`ErrWorkflowNotFound`) for "not found" cases so the handler can map them to 404.
- Wrap errors with `fmt.Errorf("repository_name: operation: %w", err)`.

---

### 2. Add the use case (app layer)

`internal/app/workflow.go`

```go
// GetWorkflow returns a workflow by ID for a given project.
func (s *WorkflowService) GetWorkflow(ctx context.Context, projectID, workflowID string) (*models.Workflow, error) {
    if strings.TrimSpace(workflowID) == "" {
        return nil, &ValidationError{Msg: "workflow id is required"}
    }
    return s.repo.GetByID(ctx, projectID, workflowID)
}
```

The app layer is where validation and business rules live. Keep it free of HTTP and SQL concerns.

---

### 3. Add the handler

`internal/interfaces/http/workflow_handler.go`

Add the Swag annotations and handler function:

```go
// GetWorkflow godoc
//
//	@Summary		Get workflow
//	@Description	Returns a single workflow definition by ID.
//	@Tags			workflows
//	@Produce		json
//	@Param			id	path		string	true	"Workflow ID"
//	@Success		200	{object}	models.Workflow
//	@Failure		400	{object}	JSONError
//	@Failure		401	{object}	JSONError
//	@Failure		404	{object}	JSONError
//	@Failure		500	{object}	JSONError
//	@Security		ApiKeyAuth
//	@Router			/workflows/{id} [get]
func (h *WorkflowHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
    projectID, ok := auth.ProjectIDFromContext(r.Context())
    if !ok {
        respond.Error(w, http.StatusUnauthorized, "unauthorized")
        return
    }

    workflowID := r.PathValue("id")

    wf, err := h.svc.GetWorkflow(r.Context(), projectID, workflowID)
    if errors.Is(err, db.ErrWorkflowNotFound) {
        respond.Error(w, http.StatusNotFound, "workflow not found")
        return
    }
    if msg, ok := app.AsValidation(err); ok {
        respond.Error(w, http.StatusBadRequest, msg)
        return
    }
    if err != nil {
        log.Printf("workflows: get: %v", err)
        respond.Error(w, http.StatusInternalServerError, "internal server error")
        return
    }

    respond.JSON(w, http.StatusOK, wf)
}
```

Handler rules:
- Always extract `projectID` from context for protected routes — never trust a client-supplied project ID.
- Map sentinel errors to the correct HTTP status (`ErrWorkflowNotFound` → 404).
- Log unexpected errors at `log.Printf` level before returning 500. Never expose internal details in the response.
- Use `respond.JSON` / `respond.Error` — never write directly to `w`.

---

### 4. Register the route

`internal/interfaces/http/router.go`

```go
protected.HandleFunc("GET /workflows/{id}", workflows.GetWorkflow)
```

Route registration rules:
- Protected routes (auth required) go on the `protected` mux.
- Public routes go on the `mux` directly.
- Use Go 1.22+ method+path pattern: `"METHOD /path/{param}"`.

---

### 5. Regenerate Swagger

```bash
/Users/sahidayala/go/bin/swag init -g cmd/api/main.go -o docs
```

Or via go generate:

```bash
go generate ./cmd/api/...
```

---

## Checklist for Every New Endpoint

- [ ] SQL query uses `$N` parameters — no string interpolation
- [ ] Repository returns a sentinel error for "not found"
- [ ] App layer validates input and returns `*ValidationError` for bad input
- [ ] Handler reads `projectID` from context (never from request body/params)
- [ ] Handler maps every error type to the correct HTTP status
- [ ] Swag annotations added with all response codes documented
- [ ] Route registered on the correct mux (`protected` vs public)
- [ ] Swagger regenerated and `docs/` committed

---

## Common Patterns

### Reading path params

```go
id := r.PathValue("id")  // Go 1.22+ stdlib
```

### Decoding a JSON request body

```go
dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
dec.DisallowUnknownFields()
var req MyRequest
if err := dec.Decode(&req); err != nil {
    respond.Error(w, http.StatusBadRequest, "invalid request body")
    return
}
```

### Responding with JSON

```go
respond.JSON(w, http.StatusOK, payload)    // 200 with body
respond.JSON(w, http.StatusCreated, payload) // 201 with body
respond.Error(w, http.StatusNotFound, "not found") // error shape
```

### Injecting and reading project ID

```go
// Middleware injects it:
r.WithContext(auth.WithProjectID(r.Context(), projectID))

// Handler reads it:
projectID, ok := auth.ProjectIDFromContext(r.Context())
```

---

## Adding a New Step Type

To add a new executor step type (e.g., `script`):

1. Add a case in `internal/app/executor/executor.go` → `executeStep` switch
2. Implement the executor function in a new file `internal/app/executor/script.go`
3. Add the type to the `enum` in the `CreateWorkflowStepRequest` Swag annotation
4. Document the `config` shape in `USAGE.md`
5. Add `isNonRetriable` handling in `retry.go` if needed
