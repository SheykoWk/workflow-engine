package httpapi

import "github.com/SheykoWk/workflow-engine/internal/infrastructure/db/models"

// Swag-linked type anchors so swag can resolve models.* in handler comments without "unused import".
var _ = []any{
	(*[]models.Workflow)(nil),
	(*[]models.Project)(nil),
}
