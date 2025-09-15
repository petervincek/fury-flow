package testdata

import (
	"fmt"

	"github.com/petervincek/fury-flow/internal/api/attachment"
)

const (
	POST_KANBAN_ATTACHMENT_URL_TEMPLATE        = "%s/boards/%d/cards/%d/attachments"
	GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE = POST_KANBAN_ATTACHMENT_URL_TEMPLATE
	GET_KANBAN_ATTACHMENT_URL_TEMPLATE         = "%s/boards/%d/cards/%d/attachments/%d"
	DELETE_KANBAN_ATTACHMENTS_URL_TEMPLATE     = GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE
	DELETE_KANBAN_ATTACHMENT_URL_TEMPLATE      = GET_KANBAN_ATTACHMENT_URL_TEMPLATE
)

// CreateKanbanAttachment creates a new attachment for a specified Kanban card on a board.
// It sends a POST request to the Kanban attachment endpoint with the provided attachment data.
// Parameters:
//   - boardId: the ID of the Kanban board.
//   - cardId: the ID of the Kanban card.
//   - kanbanAttachment: the attachment data to be added.
//
// Returns:
//   - ResponseResult containing the created attachment.
//   - error if the request fails.
func (td *TestData) CreateKanbanAttachment(boardId int, cardId int,
	kanbanAttachment attachment.Attachment) (ResponseResult[attachment.Attachment], error) {
	return MakePostRequest[attachment.Attachment, attachment.Attachment](fmt.Sprintf(
		POST_KANBAN_ATTACHMENT_URL_TEMPLATE, td.GetAppUrl(), boardId, cardId), kanbanAttachment)
}

// GetKanbanCardAttachments retrieves the list of attachments for a specific Kanban card on a given board.
// It takes the board ID and card ID as parameters and returns a ResponseResult containing a slice of Attachment objects,
// along with an error if the request fails.
func (td *TestData) GetKanbanCardAttachments(boardId int, cardId int) (ResponseResult[[]attachment.Attachment], error) {
	return MakeGetRequest[[]attachment.Attachment](fmt.Sprintf(GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE,
		td.GetAppUrl(), boardId, cardId))
}

// GetKanbanCardAttachment retrieves an attachment associated with a specific comment on a Kanban card.
// It requires the board ID, card ID, and comment ID to construct the request.
// Returns a ResponseResult containing the Attachment and an error if the request fails.
func (td *TestData) GetKanbanCardAttachment(boardId int, cardId int, commentId int) (ResponseResult[attachment.Attachment], error) {
	return MakeGetRequest[attachment.Attachment](fmt.Sprintf(GET_KANBAN_ATTACHMENT_URL_TEMPLATE,
		td.GetAppUrl(), boardId, cardId, commentId))
}

// DeleteKanbanCardAttachments deletes all attachments from a specified Kanban card.
// It takes the board ID and card ID as parameters and returns a ResponseResult with an empty struct
// and an error if the operation fails.
func (td *TestData) DeleteKanbanCardAttachments(boardId int, cardId int) (ResponseResult[struct{}], error) {
	return MakeDeleteRequest[struct{}](fmt.Sprintf(DELETE_KANBAN_ATTACHMENTS_URL_TEMPLATE,
		td.GetAppUrl(), boardId, cardId))
}

// DeleteKanbanCardAttachment deletes an attachment from a kanban card comment.
// It takes the board ID, card ID, and comment ID as parameters and returns a ResponseResult
// containing an empty struct if successful, or an error if the request fails.
func (td *TestData) DeleteKanbanCardAttachment(boardId int, cardId int, commentId int) (ResponseResult[struct{}], error) {
	return MakeDeleteRequest[struct{}](fmt.Sprintf(DELETE_KANBAN_ATTACHMENT_URL_TEMPLATE,
		td.GetAppUrl(), boardId, cardId, commentId))
}
