package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/SheykoWk/workflow-engine/internal/infrastructure/db"
)

// APIKeyMiddleware returns middleware that validates Authorization: Bearer wf_...,
// loads the key row by prefix, verifies bcrypt, and sets project ID on the request context.
func APIKeyMiddleware(keys *db.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := ParseBearerAPIKey(r.Header.Get("Authorization"))
			if !ok {
				writeAuthJSON(w, http.StatusUnauthorized, `{"error":"missing or invalid authorization"}`)
				return
			}
			dbPrefix, fullKey, ok := SplitWFAPIKey(raw)
			if !ok {
				writeAuthJSON(w, http.StatusUnauthorized, `{"error":"invalid API key format"}`)
				return
			}
			projectID, keyHash, err := keys.FindActiveByPrefix(r.Context(), dbPrefix)
			if err != nil {
				if !errors.Is(err, db.ErrAPIKeyNotFound) {
					log.Printf("auth: lookup key: %v", err)
				}
				writeAuthJSON(w, http.StatusUnauthorized, `{"error":"unauthorized"}`)
				return
			}
			if !CompareAPIKey(keyHash, fullKey) {
				writeAuthJSON(w, http.StatusUnauthorized, `{"error":"unauthorized"}`)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithProjectID(r.Context(), projectID)))
		})
	}
}

func writeAuthJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}
