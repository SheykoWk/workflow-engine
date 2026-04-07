package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSONError is the JSON body for error responses (e.g. 500).
// Used by Swagger / OpenAPI annotations.
type JSONError struct {
	Error string `json:"error" example:"internal server error"`
}

// writeJSON marshals v to JSON first, then writes status and body in one step.
// If Marshal fails, the client never receives a 200 with a truncated body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("httpapi: json marshal: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(b); err != nil {
		log.Printf("httpapi: write body: %v", err)
	}
}

// writeJSONError sends a small JSON object {"error":"..."} with the given status.
func writeJSONError(w http.ResponseWriter, code int, msg string) {
	b, err := json.Marshal(map[string]string{"error": msg})
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if _, err := w.Write(b); err != nil {
		log.Printf("httpapi: write error body: %v", err)
	}
}
