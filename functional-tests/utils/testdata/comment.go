package testdata

import (
	"fmt"

	"github.com/petervincek/fury-flow/internal/api/comment"
)

const (
	POST_KANBAN_COMMENT_URL_TEMPLATE        = "%s/boards/%d/cards/%d/comments"
	GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE = POST_KANBAN_COMMENT_URL_TEMPLATE
	GET_KANBAN_COMMENT_URL_TEMPLATE         = "%s/boards/%d/cards/%d/comments/%d"
	DELETE_KANBAN_COMMENTS_URL_TEMPLATE     = GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE
	DELETE_KANBAN_COMMENT_URL_TEMPLATE      = GET_KANBAN_COMMENT_URL_TEMPLATE
)

// CreateKanbanComment creates a new comment on a Kanban card identified by boardId and cardId.
// It sends a POST request with the provided kanbanComment data and returns the created comment
// wrapped in a ResponseResult, or an error if the request fails.
//
// Parameters:
//   - boardId: the ID of the Kanban board.
//   - cardId: the ID of the Kanban card.
//   - kanbanComment: the comment to be added.
//
// Returns:
//   - ResponseResult[comment.Comment]: the result containing the created comment.
//   - error: an error if the request fails.
func (td *TestData) CreateKanbanComment(boardId int, cardId int,
	kanbanComment comment.Comment) (ResponseResult[comment.Comment], error) {
	return MakePostRequest[comment.Comment, comment.Comment](fmt.Sprintf(
		POST_KANBAN_COMMENT_URL_TEMPLATE, td.GetAppUrl(), boardId, cardId), kanbanComment)
}

// GetKanbanCardComments retrieves the list of comments for a specific Kanban card on a given board.
// It takes the board ID and card ID as parameters and returns a ResponseResult containing a slice of comment.Comment,
// along with an error if the request fails.
func (td *TestData) GetKanbanCardComments(boardId int, cardId int) (ResponseResult[[]comment.Comment], error) {
	return MakeGetRequest[[]comment.Comment](fmt.Sprintf(GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE,
		td.GetAppUrl(), boardId, cardId))
}

// GetKanbanCardComment retrieves a specific comment from a Kanban card on the given board.
// It takes the board ID, card ID, and comment ID as parameters, and returns a ResponseResult
// containing the comment data, or an error if the request fails.
func (td *TestData) GetKanbanCardComment(boardId int, cardId int, commentId int) (ResponseResult[comment.Comment], error) {
	return MakeGetRequest[comment.Comment](fmt.Sprintf(GET_KANBAN_COMMENT_URL_TEMPLATE,
		td.GetAppUrl(), boardId, cardId, commentId))
}

// DeleteKanbanCardComments deletes all comments associated with a specific Kanban card on a given board.
// It takes the board ID and card ID as parameters and returns a ResponseResult with an empty struct
// if successful, or an error if the request fails.
func (td *TestData) DeleteKanbanCardComments(boardId int, cardId int) (ResponseResult[struct{}], error) {
	return MakeDeleteRequest[struct{}](fmt.Sprintf(DELETE_KANBAN_COMMENTS_URL_TEMPLATE,
		td.GetAppUrl(), boardId, cardId))
}

// DeleteKanbanCardComment deletes a comment from a Kanban card identified by the given boardId, cardId, and commentId.
// It sends a DELETE request to the appropriate API endpoint and returns a ResponseResult with an empty struct on success.
// If an error occurs during the request, it returns the error.
func (td *TestData) DeleteKanbanCardComment(boardId int, cardId int, commentId int) (ResponseResult[struct{}], error) {
	return MakeDeleteRequest[struct{}](fmt.Sprintf(DELETE_KANBAN_COMMENT_URL_TEMPLATE,
		td.GetAppUrl(), boardId, cardId, commentId))
}
