// Package respond writes JSON HTTP responses (success and error bodies).
package respond

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON writes a JSON response with the given status. The body is marshaled in full before writing.
func JSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("respond: json marshal: %v", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(b); err != nil {
		log.Printf("respond: write body: %v", err)
	}
}

// Error writes {"error":"..."} with the given HTTP status.
func Error(w http.ResponseWriter, code int, msg string) {
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
		log.Printf("respond: write error body: %v", err)
	}
}
