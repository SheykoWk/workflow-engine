package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/SheykoWk/workflow-engine/internal/app"
	"github.com/SheykoWk/workflow-engine/internal/auth"
	"github.com/SheykoWk/workflow-engine/internal/infrastructure/db"
	"github.com/SheykoWk/workflow-engine/internal/interfaces/http/respond"
)

const maxCreateWorkflowBody = 1 << 20 // 1 MiB

// WorkflowHandler serves HTTP for workflows (definitions under a project).
type WorkflowHandler struct {
	svc *app.WorkflowService
}

// NewWorkflowHandler wires the handler with its dependencies.
func NewWorkflowHandler(svc *app.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

// CreateWorkflowRequest is the JSON body for POST /workflows.
type CreateWorkflowRequest struct {
	Name  string                      `json:"name"`
	Steps []CreateWorkflowStepRequest `json:"steps"`
}

// CreateWorkflowStepRequest is one step in the create payload.
type CreateWorkflowStepRequest struct {
	Name string `json:"name" example:"Fetch user"`
	// Type is the executor step kind: delay, http_request.
	Type string `json:"type" enums:"delay http_request" example:"http_request"`
	// Config is step-specific JSON. delay: {"seconds":1}. http_request: {"method":"GET","url":"https://example.com","headers":{},"body":{}}. Optional retry: {"max_attempts":3,"backoff_seconds":5}. Strings may include {{steps.0.output.body.id}}.
	Config json.RawMessage `json:"config" swaggertype:"object"`
}

// CreateWorkflowResponse is returned after creating a workflow definition.
type CreateWorkflowResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	StepsCount int    `json:"steps_count"`
}

// CreateWorkflowRunResponse is returned after POST /workflows/{id}/runs.
type CreateWorkflowRunResponse struct {
	RunID  string `json:"run_id"`
	Status string `json:"status"`
	Steps  int    `json:"steps"`
}

// CreateWorkflow handles POST /workflows (authenticated).
//
//	@Summary		Create workflow
//	@Description	Creates a workflow with ordered steps (`steps[0]`, `steps[1]`, …) for the authenticated project. Each step has `type` and `config` JSON; see `CreateWorkflowStepRequest`. Steps run in order in the worker; later steps can reference earlier outputs via `{{steps.N.output...}}` in config strings.
//	@Tags			workflows
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		CreateWorkflowRequest	true	"Workflow definition"
//	@Success		201		{object}	CreateWorkflowResponse
//	@Failure		400		{object}	JSONError
//	@Failure		401		{object}	JSONError
//	@Failure		500		{object}	JSONError
//	@Router			/workflows [post]
func (h *WorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCreateWorkflowBody)
	var req CreateWorkflowRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	steps := make([]app.CreateWorkflowStepInput, 0, len(req.Steps))
	for _, s := range req.Steps {
		cfg := s.Config
		if len(cfg) == 0 {
			cfg = json.RawMessage(`{}`)
		}
		steps = append(steps, app.CreateWorkflowStepInput{
			Name:     s.Name,
			StepType: s.Type,
			Config:   cfg,
		})
	}

	out, err := h.svc.CreateWorkflow(r.Context(), projectID, app.CreateWorkflowInput{
		Name:  req.Name,
		Steps: steps,
	})
	if err != nil {
		if msg, ok := app.AsValidation(err); ok {
			respond.Error(w, http.StatusBadRequest, msg)
			return
		}
		log.Printf("workflows: create: %v", err)
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respond.JSON(w, http.StatusCreated, CreateWorkflowResponse{
		ID:         out.ID,
		Name:       out.Name,
		StepsCount: out.StepsCount,
	})
}

// CreateWorkflowRun handles POST /workflows/{id}/runs (authenticated).
//
//	@Summary		Start workflow run
//	@Description	Creates a `workflow_run` in status pending and one `step_run` per definition step (pending, attempt 1). The worker process picks up and executes steps; this endpoint only enqueues work.
//	@Tags			workflows
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string	true	"Workflow ID"
//	@Success		201	{object}	CreateWorkflowRunResponse
//	@Failure		400	{object}	JSONError
//	@Failure		401	{object}	JSONError
//	@Failure		404	{object}	JSONError
//	@Failure		500	{object}	JSONError
//	@Router			/workflows/{id}/runs [post]
func (h *WorkflowHandler) CreateWorkflowRun(w http.ResponseWriter, r *http.Request) {
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	wfID := strings.TrimSpace(r.PathValue("id"))
	if wfID == "" {
		respond.Error(w, http.StatusBadRequest, "workflow id is required")
		return
	}

	out, err := h.svc.StartWorkflowRun(r.Context(), projectID, wfID)
	if err != nil {
		if msg, ok := app.AsValidation(err); ok {
			respond.Error(w, http.StatusBadRequest, msg)
			return
		}
		if errors.Is(err, db.ErrWorkflowNotFound) {
			respond.Error(w, http.StatusNotFound, "workflow not found")
			return
		}
		log.Printf("workflows: create run: %v", err)
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respond.JSON(w, http.StatusCreated, CreateWorkflowRunResponse{
		RunID:  out.RunID,
		Status: out.Status,
		Steps:  out.Steps,
	})
}

// GetAllWorkflows handles GET /workflows: list workflows for the authenticated project.
//
//	@Summary		List workflows
//	@Description	Returns workflow definitions for the project from the API key.
//	@Tags			workflows
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{array}	models.Workflow
//	@Failure		401	{object}	JSONError
//	@Failure		500	{object}	JSONError
//	@Router			/workflows [get]
func (h *WorkflowHandler) GetAllWorkflows(w http.ResponseWriter, r *http.Request) {
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	workflows, err := h.svc.ListByProject(r.Context(), projectID)
	if err != nil {
		log.Printf("workflows: get all: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, workflows)
}
