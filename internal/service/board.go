package service

import (
	"context"
	"errors"

	"github.com/petervincek/fury-flow/internal/db"
)

var (
	ErrorKanbanBoardNotUpdated = errors.New("kanban board not updated")
)

type KanbanBoardService struct {
	q *db.Queries
}

// NewKanbanBoardService creates and returns a new instance of KanbanBoardService
// using the provided database queries. It initializes the service with the given
// db.Queries dependency for board-related operations.
func NewKanbanBoardService(q *db.Queries) *KanbanBoardService {
	return &KanbanBoardService{q: q}
}

// GetKanbanBoards retrieves all Kanban boards from the database.
// It returns a slice of KanbanBoard objects and an error if the operation fails.
//
// Parameters:
//
//	ctx - The context for controlling cancellation and deadlines.
//
// Returns:
//
//	[]db.KanbanBoard - A slice containing the Kanban boards.
//	error - An error if the retrieval fails, otherwise nil.
func (kbs *KanbanBoardService) GetKanbanBoards(ctx context.Context) ([]db.KanbanBoard, error) {
	return kbs.q.GetKanbanBoards(ctx)
}

// GetKanbanBoardById retrieves a Kanban board from the database by its unique identifier.
// It takes a context for request-scoped values and cancellation, and the board's ID as an int32.
// Returns the KanbanBoard object and an error if the retrieval fails.
func (kbs *KanbanBoardService) GetKanbanBoardById(ctx context.Context, boardId int32) (db.KanbanBoard, error) {
	return kbs.q.GetKanbanBoardById(ctx, boardId)
}

// CreateKanbanBoard creates a new Kanban board in the database using the provided parameters.
// It returns the created KanbanBoard and any error encountered during the operation.
//
// Parameters:
//   - ctx: The context for controlling cancellation and deadlines.
//   - board: The parameters required to create a new Kanban board.
//
// Returns:
//   - db.KanbanBoard: The newly created Kanban board.
//   - error: An error if the creation fails, otherwise nil.
func (kbs *KanbanBoardService) CreateKanbanBoard(ctx context.Context, board db.CreateKanbanBoardParams) (db.KanbanBoard, error) {
	return kbs.q.CreateKanbanBoard(ctx, board)
}

// UpdateKanbanBoard updates the details of a Kanban board in the database.
// It takes a context and the parameters required for the update operation.
// Returns an error if the update fails or if no rows were affected.
func (kbs *KanbanBoardService) UpdateKanbanBoard(ctx context.Context, board db.UpdateKanbanBoardParams) error {
	commandTag, err := kbs.q.UpdateKanbanBoard(ctx, board)
	if err != nil {
		return err
	}
	rowsAffected := commandTag.RowsAffected()
	if rowsAffected != 1 {
		return ErrorKanbanBoardNotUpdated
	}
	return nil
}

// DeleteKanbanBoard deletes a Kanban board identified by the given boardId.
// It delegates the deletion operation to the underlying query layer.
// Returns an error if the deletion fails.
func (kbs *KanbanBoardService) DeleteKanbanBoard(ctx context.Context, boardId int32) error {
	return kbs.q.DeleteKanbanBoard(ctx, boardId)
}
