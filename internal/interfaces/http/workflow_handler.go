package httpapi

import (
	"log"
	"net/http"

	"github.com/SheykoWk/workflow-engine/internal/auth"
	"github.com/SheykoWk/workflow-engine/internal/infrastructure/db"
)

// WorkflowHandler serves HTTP for workflows (definitions under a project).
type WorkflowHandler struct {
	repo *db.WorkflowRepository
}

// NewWorkflowHandler wires the handler with its dependencies.
func NewWorkflowHandler(repo *db.WorkflowRepository) *WorkflowHandler {
	return &WorkflowHandler{repo: repo}
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
	workflows, err := h.repo.GetAllByProjectID(r.Context(), projectID)
	if err != nil {
		log.Printf("workflows: get all: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, workflows)
}
