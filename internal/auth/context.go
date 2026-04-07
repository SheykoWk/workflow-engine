package auth

import "context"

type ctxKey int

const projectIDKey ctxKey = iota

// WithProjectID returns a child context carrying the authenticated project ID.
func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, projectIDKey, projectID)
}

// ProjectIDFromContext returns the project ID set by API key middleware.
func ProjectIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(projectIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}
