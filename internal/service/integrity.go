package service

import (
	"context"

	"github.com/petervincek/fury-flow/internal/db"
)

// KanbanIntegrityService provides methods to ensure the integrity of Kanban-related operations.
// It interacts with the database through the provided Queries instance.
type KanbanIntegrityService struct {
	q *db.Queries
}

// NewKanbanIntegrityService creates and returns a new instance of KanbanIntegrityService
// using the provided database queries. It is used to perform integrity checks
// related to Kanban operations.
//
// Parameters:
//   - q: Pointer to db.Queries, which provides access to database operations.
//
// Returns:
//   - Pointer to KanbanIntegrityService.
func NewKanbanIntegrityService(q *db.Queries) *KanbanIntegrityService {
	return &KanbanIntegrityService{q: q}
}

// CardBelongsToBoard checks whether a card with the given cardId belongs to the board specified by boardId.
// It returns true if the card is associated with the board, false otherwise.
// An error is returned if the database query fails or if there is an issue during the check.
func (kis *KanbanIntegrityService) CardBelongsToBoard(ctx context.Context,
	boardId int, cardId int) (bool, error) {
	return kis.q.CardBelongsToBoard(ctx, db.CardBelongsToBoardParams{
		BoardID: int32(boardId),
		CardID:  int32(cardId),
	})
}

// CommentBelongsToCard checks whether a given comment belongs to a specific card.
// It takes a context, the card's ID, and the comment's ID as parameters.
// Returns true if the comment is associated with the card, otherwise false.
// An error is returned if the database query fails.
func (kis *KanbanIntegrityService) CommentBelongsToCard(ctx context.Context,
	cardId int, commentId int) (bool, error) {
	return kis.q.CommentBelongsToCard(ctx, db.CommentBelongsToCardParams{
		CardID:    int32(cardId),
		CommentID: int32(commentId),
	})
}

// BoardCardCommentRelationshipExists checks if a relationship exists between a board, card, and comment.
// It takes a context, board ID, card ID, and comment ID as parameters.
// Returns true if the relationship exists, otherwise false. An error is returned if the query fails.
func (kis *KanbanIntegrityService) BoardCardCommentRelationshipExists(ctx context.Context,
	boardId int, cardId int, commentId int) (bool, error) {
	return kis.q.BoardCardCommentRelationshipExists(ctx, db.BoardCardCommentRelationshipExistsParams{
		BoardID:   int32(boardId),
		CardID:    int32(cardId),
		CommentID: int32(commentId),
	})
}
