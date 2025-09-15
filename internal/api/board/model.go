package board

import (
	"time"

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

// Builder is a struct that provides functionality to construct and manage a Board instance.
type Builder struct {
	board Board
}

// NewBuilder creates and returns a new instance of Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// SetBoardId sets the BoardId field of the board to the specified value.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetBoardId(boardId int) *Builder {
	b.board.BoardId = boardId
	return b
}

// SetBoardName sets the name of the board in the Builder and returns the Builder for chaining.
func (b *Builder) SetBoardName(boardName string) *Builder {
	b.board.BoardName = boardName
	return b
}

// SetDescription sets the description of the board and returns the Builder for chaining.
func (b *Builder) SetDescription(description string) *Builder {
	b.board.Description = description
	return b
}

// SetCreatedBy sets the CreatedBy field of the board to the specified value.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCreatedBy(createdBy string) *Builder {
	b.board.CreatedBy = createdBy
	return b
}

// SetCreatedAt sets the CreatedAt field of the board to the specified time.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCreatedAt(createdAt time.Time) *Builder {
	b.board.CreatedAt = createdAt
	return b
}

// SetUpdatedBy sets the UpdatedBy field of the board to the specified value.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetUpdatedBy(updatedBy string) *Builder {
	b.board.UpdatedBy = updatedBy
	return b
}

// SetUpdatedAt sets the UpdatedAt field of the board to the specified time.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetUpdatedAt(updatedAt time.Time) *Builder {
	b.board.UpdatedAt = updatedAt
	return b
}

// Build validates the board structure and returns the constructed Board.
// If validation fails, it returns an empty Board and the validation error.
func (b *Builder) Build() (Board, error) {
	err := validate.Struct(b.board)
	if err != nil {
		return Board{}, err
	}
	return b.board, nil
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
