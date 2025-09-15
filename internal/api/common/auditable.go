package common

import "time"

// AuditableFields represents common audit fields for tracking creation and modification metadata.
// It includes timestamps and user identifiers for both creation and last update events.
type AuditableFields struct {
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy string    `json:"updatedBy"`
}
