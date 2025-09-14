package service

import (
	"context"

	"github.com/petervincek/fury-flow/internal/db"
)

// KanbanCommentService provides methods to manage comments within a Kanban workflow.
// It interacts with the database through the db.Queries instance to perform operations
// such as creating, retrieving, updating, and deleting comments associated with Kanban items.
type KanbanCommentService struct {
	q *db.Queries
}

// NewKanbanCommentService creates a new instance of KanbanCommentService using the provided database queries.
// It returns a pointer to the initialized KanbanCommentService.
func NewKanbanCommentService(q *db.Queries) *KanbanCommentService {
	return &KanbanCommentService{
		q: q,
	}
}

// GetKanbanCardComments retrieves all comments associated with a specific Kanban card.
// It takes a context and the card's ID as parameters, and returns a slice of CardComment objects
// along with an error if the operation fails.
func (kcs *KanbanCommentService) GetKanbanCardComments(ctx context.Context, cardId int32) ([]db.CardComment, error) {
	return kcs.q.GetKanbanCardComments(ctx, cardId)
}

// GetKanbanCommentById retrieves a Kanban card comment by its unique identifier.
// It takes a context for request-scoped values and cancellation, and the comment's ID.
// Returns the corresponding CardComment from the database, or an error if not found.
func (kcs *KanbanCommentService) GetKanbanCommentById(ctx context.Context, commentId int32) (db.CardComment, error) {
	return kcs.q.GetKanbanCommentById(ctx, commentId)
}

// CreateKanbanComment creates a new comment on a Kanban card using the provided parameters.
// It returns the created CardComment and any error encountered during the operation.
//
// Parameters:
//   - ctx: The context for controlling cancellation and deadlines.
//   - comment: The parameters required to create a Kanban comment.
//
// Returns:
//   - db.CardComment: The newly created Kanban card comment.
//   - error: An error if the creation fails, otherwise nil.
func (kcs *KanbanCommentService) CreateKanbanComment(ctx context.Context, comment db.CreateKanbanCommentParams) (db.CardComment, error) {
	return kcs.q.CreateKanbanComment(ctx, comment)
}

// DeleteAllKanbanComments deletes all kanban comments from the data store.
// It delegates the deletion operation to the underlying query service.
// Returns an error if the deletion fails.
func (kcs *KanbanCommentService) DeleteAllKanbanComments(ctx context.Context) error {
	return kcs.q.DeleteAllKanbanComments(ctx)
}

// DeleteKanbanComment deletes a Kanban comment identified by the given commentId.
// It delegates the deletion operation to the underlying query layer.
// Returns an error if the deletion fails.
//
// Parameters:
//
//	ctx - The context for controlling cancellation and deadlines.
//	commentId - The unique identifier of the Kanban comment to be deleted.
//
// Returns:
//
//	error - An error if the deletion operation fails, otherwise nil.
func (kcs *KanbanCommentService) DeleteKanbanComment(ctx context.Context, commentId int32) error {
	return kcs.q.DeleteKanbanComment(ctx, commentId)
}

// DeleteKanbanCommentsForCard deletes all comments associated with the specified Kanban card.
// It takes a context for cancellation and timeout control, and the cardId of the Kanban card.
// Returns an error if the deletion fails.
func (kcs *KanbanCommentService) DeleteKanbanCommentsForCard(ctx context.Context, cardId int32) error {
	return kcs.q.DeleteKanbanCommentsForCard(ctx, cardId)
}
