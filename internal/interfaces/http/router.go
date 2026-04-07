package httpapi

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/SheykoWk/workflow-engine/internal/auth"
	"github.com/SheykoWk/workflow-engine/internal/infrastructure/db"
)

// NewRouter returns the HTTP handler tree for the API. It uses net/http only
// (no external router). Register routes on a ServeMux and return it as http.Handler.
func NewRouter(apiKeys *db.APIKeyRepository, workflows *WorkflowHandler, projects *ProjectHandler) http.Handler {
	protected := http.NewServeMux()
	protected.HandleFunc("GET /workflows", workflows.GetAllWorkflows)
	protected.HandleFunc("GET /projects", projects.GetCurrentProject)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /projects", projects.CreateProject)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/", auth.APIKeyMiddleware(apiKeys)(protected))
	return mux
}

// health liveness probe.
//
//	@Summary		Health check
//	@Description	Returns plain text ok when the process is up.
//	@Tags			system
//	@Produce		plain
//	@Success		200	{string}	string	"ok"
//	@Router			/health [get]
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
