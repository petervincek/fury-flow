package service

import (
	"context"

	"github.com/petervincek/fury-flow/internal/db"
)

// KanbanAttachmentService provides methods to manage attachments within the Kanban workflow.
// It interacts with the database through the db.Queries instance to perform operations related to attachments.
type KanbanAttachmentService struct {
	q *db.Queries
}

// NewKanbanAttachmentService creates and returns a new instance of KanbanAttachmentService
// using the provided database queries. This service is responsible for handling
// attachment-related operations in the Kanban workflow.
//
// Parameters:
//   - q: Pointer to db.Queries, used for database interactions.
//
// Returns:
//   - Pointer to KanbanAttachmentService.
func NewKanbanAttachmentService(q *db.Queries) *KanbanAttachmentService {
	return &KanbanAttachmentService{
		q: q,
	}
}

// GetKanbanCardAttachments retrieves all attachments associated with a specific Kanban card.
// It takes a context and the card's ID as parameters, and returns a slice of CardAttachment objects
// along with an error if the retrieval fails.
func (kas *KanbanAttachmentService) GetKanbanCardAttachments(ctx context.Context, cardId int32) ([]db.CardAttachment, error) {
	return kas.q.GetKanbanCardAttachments(ctx, cardId)
}

// GetKanbanAttachmentById retrieves a Kanban card attachment by its unique identifier.
// It takes a context for request-scoped values and cancellation, and the attachment ID.
// Returns the corresponding CardAttachment from the database, or an error if not found or on failure.
func (kas *KanbanAttachmentService) GetKanbanAttachmentById(ctx context.Context, attachmentId int32) (db.CardAttachment, error) {
	return kas.q.GetKanbanAttachmentById(ctx, attachmentId)
}

// CreateKanbanAttachment creates a new Kanban card attachment in the database using the provided parameters.
// It returns the created CardAttachment and any error encountered during the operation.
//
// Parameters:
//   - ctx: The context for controlling cancellation and deadlines.
//   - attachment: The parameters required to create a Kanban card attachment.
//
// Returns:
//   - db.CardAttachment: The newly created Kanban card attachment.
//   - error: An error if the creation fails, otherwise nil.
func (kas *KanbanAttachmentService) CreateKanbanAttachment(ctx context.Context, attachment db.CreateKanbanAttachmentParams) (db.CardAttachment, error) {
	return kas.q.CreateKanbanAttachment(ctx, attachment)
}

// DeleteAllKanbanAttachments deletes all Kanban attachments from the data store.
// It delegates the deletion operation to the underlying query service.
// Returns an error if the deletion fails.
func (kas *KanbanAttachmentService) DeleteAllKanbanAttachments(ctx context.Context) error {
	return kas.q.DeleteAllKanbanAttachments(ctx)
}

// DeleteKanbanAttachment deletes a Kanban attachment identified by the given attachmentId.
// It returns an error if the deletion fails.
//
// Parameters:
//
//	ctx - The context for controlling cancellation and deadlines.
//	attachmentId - The unique identifier of the attachment to be deleted.
//
// Returns:
//
//	error - An error if the deletion operation fails, otherwise nil.
func (kas *KanbanAttachmentService) DeleteKanbanAttachment(ctx context.Context, attachmentId int32) error {
	return kas.q.DeleteKanbanAttachment(ctx, attachmentId)
}

// DeleteKanbanAttachmentsForCard deletes all attachments associated with the specified Kanban card.
// It takes a context for request-scoped values and cancellation, and the card's ID.
// Returns an error if the deletion fails.
func (kas *KanbanAttachmentService) DeleteKanbanAttachmentsForCard(ctx context.Context, cardId int32) error {
	return kas.q.DeleteKanbanAttachmentsForCard(ctx, cardId)
}
