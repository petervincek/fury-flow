package board

import (
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/petervincek/fury-flow/internal/db"
)

// Board represents a board entity with an ID, name, description, and auditable fields.
// BoardName and Description are required and must be at least 3 characters long.
// AuditableFields provides common auditing metadata such as timestamps and user information.
type Board struct {
	BoardId     int    `json:"boardId"`
	BoardName   string `json:"boardName" validate:"required,min=3"`
	Description string `json:"description" validate:"required,min=3"`
	common.AuditableFields
}

// TransformBoard converts a db.KanbanBoard object to a Board object,
// mapping relevant fields and transforming types as necessary.
// It also initializes the AuditableFields with creation and update metadata.
func TransformBoard(kb db.KanbanBoard) Board {
	return Board{
		BoardId:     int(kb.BoardID),
		BoardName:   kb.BoardName,
		Description: kb.Description.String,
		AuditableFields: common.AuditableFields{
			CreatedAt: kb.CreatedAt.Time,
			CreatedBy: kb.CreatedBy,
			UpdatedAt: kb.UpdatedAt.Time,
			UpdatedBy: kb.UpdatedBy.String,
		},
	}
}

// TransformBoards converts a slice of db.KanbanBoard objects to a slice of Board objects
// by applying the TransformBoard function to each element.
// It returns a new slice containing the transformed Board objects.
func TransformBoards(kbs []db.KanbanBoard) []Board {
	boards := make([]Board, len(kbs))
	for idx, kb := range kbs {
		boards[idx] = TransformBoard(kb)
	}
	return boards
}
