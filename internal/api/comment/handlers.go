package comment

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/petervincek/fury-flow/internal/db"
	"github.com/petervincek/fury-flow/internal/logging"
	"github.com/petervincek/fury-flow/internal/service"
	"github.com/petervincek/fury-flow/internal/utils/monad"
	"go.uber.org/zap"
)

const (
	REQUESTED_COMMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG = "requested comment does not belong to request card"
	KANBAN_COMMENT_VALIDATION_ERROR_MSG                 = "kanban comment validation error"
)

var logger = logging.GetLogger()
var validate = validator.New()

// CommentHandler provides HTTP handlers for managing comments within the Kanban board system.
// It interacts with KanbanBoardService, KanbanCardService, and KanbanCommentService to perform
// operations related to boards, cards, and comments respectively.
type CommentHandler struct {
	boardService   *service.KanbanBoardService
	cardService    *service.KanbanCardService
	commentService *service.KanbanCommentService
}

// New creates and returns a new instance of CommentHandler with the provided
// KanbanBoardService, KanbanCardService, and KanbanCommentService dependencies.
func New(boardService *service.KanbanBoardService, cardService *service.KanbanCardService, commentService *service.KanbanCommentService) *CommentHandler {
	return &CommentHandler{
		boardService:   boardService,
		cardService:    cardService,
		commentService: commentService,
	}
}

// GetCommentsForCard handles the HTTP request to retrieve comments for a specific card within a board.
// It performs the following steps:
//  1. Parses the board ID from the request context.
//  2. Parses the card ID from the request context.
//  3. Checks if the specified board exists.
//  4. Checks if the specified card exists within the board.
//  5. Retrieves the comments associated with the card.
//  6. Returns the comments as a JSON response with HTTP 200 status on success.
//  7. Returns appropriate error responses if any step fails.
//
// Expected request parameters:
//   - Board ID (from route or query)
//   - Card ID (from route or query)
//
// Returns:
//   - 200 OK with the list of comments in JSON format if successful.
//   - 400/404/500 error responses in case of invalid input, not found, or internal errors.
func (ch *CommentHandler) GetCommentsForCard(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// try to get the comments for a given card (card id)
	comments, err := ch.commentService.GetKanbanCardComments(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error("error while getting card's comments", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(TransformComments(comments))
}

// GetCommentById handles the HTTP request to retrieve a specific comment by its ID.
// It performs the following steps:
//  1. Parses the board, card, and comment IDs from the request context.
//  2. Checks the existence of the specified board and card.
//  3. Retrieves the comment by its ID from the comment service.
//  4. Ensures the comment belongs to the specified card.
//  5. Returns the comment as a JSON response if found, or an appropriate error response.
//
// Returns:
//   - 200 OK with the comment data if successful.
//   - 400 Bad Request if the comment does not belong to the card.
//   - 404 Not Found if the comment does not exist.
//   - 500 Internal Server Error for unexpected errors.
func (ch *CommentHandler) GetCommentById(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// try to parse comment id
	commentIdResult := common.ParseCommentId(ctx)
	if commentIdResult.IsLeft() {
		return commentIdResult.Left()
	}
	commentId := commentIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// try to get the comment (by comment id)
	comment, err := ch.commentService.GetKanbanCommentById(ctx.Context(), int32(commentId))
	if err != nil {
		logger.Error("error while getting card comment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.COMMENT_ID_PARAM, commentId))
		if errors.Is(err, sql.ErrNoRows) {
			return ctx.Status(fiber.StatusNotFound).JSON(common.NewErrorMsg(fmt.Sprintf("no kanban comment for commentId: %d", commentId), nil))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}

	// check that the comment belong to the card
	if comment.CardID != int32(cardId) {
		logger.Error(REQUESTED_COMMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.COMMENT_ID_PARAM, commentId))
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewErrorMsg(REQUESTED_COMMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(TransformComment(comment))
}

// CreateComment handles the creation of a new comment for a specific card on a board.
// It performs the following steps:
//  1. Parses and validates the board ID from the request context.
//  2. Parses and validates the card ID from the request context.
//  3. Parses and validates the comment data from the request body.
//  4. Checks the existence of the specified board.
//  5. Checks the existence of the specified card within the board.
//  6. Creates the comment using the comment service.
//  7. Returns the created comment as a JSON response with HTTP 201 status.
//
// If any step fails, it returns an appropriate error response.
func (ch *CommentHandler) CreateComment(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// try to parse the comment
	// parse and validate the card from the request body
	comment := &Comment{}
	resultCheck := parseAndValidateComment(ctx, comment)
	if resultCheck.IsLeft() {
		return resultCheck.Left()
	}

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// try to create card's comment
	createdComment, err := ch.commentService.CreateKanbanComment(ctx.Context(), db.CreateKanbanCommentParams{
		CardID:      int32(comment.CardId),
		CommentText: comment.CommentText,
		CommentBy:   comment.CommentBy,
		CreatedAt:   pgtype.Timestamptz{Time: comment.CreatedAt, Valid: true},
	})
	if err != nil {
		logger.Error("error while creating new comment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}

	return ctx.Status(fiber.StatusCreated).JSON(TransformComment(createdComment))
}

// DeleteCommentsForCard handles the HTTP request to delete all comments associated with a specific card.
// It parses the board and card IDs from the request context, checks for their existence,
// and then deletes all comments linked to the specified card. Returns appropriate HTTP status codes
// based on the outcome of each operation.
func (ch *CommentHandler) DeleteCommentsForCard(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// try to delete all comments associated with the given card
	err := ch.commentService.DeleteKanbanCommentsForCard(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error("error while deleting all card's comments", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// DeleteComment handles the deletion of a comment from a kanban card.
// It performs the following steps:
//  1. Parses the board, card, and comment IDs from the request context.
//  2. Checks the existence of the specified board and card.
//  3. Retrieves the comment by its ID and verifies that it belongs to the specified card.
//  4. Deletes the comment if all checks pass.
//
// Returns appropriate HTTP status codes and error messages for invalid input,
// non-existent resources, or internal errors.
func (ch *CommentHandler) DeleteComment(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// try to parse comment id
	commentIdResult := common.ParseCommentId(ctx)
	if commentIdResult.IsLeft() {
		return commentIdResult.Left()
	}
	commentId := commentIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// try to get the comments for a given card (card id)
	comment, err := ch.commentService.GetKanbanCommentById(ctx.Context(), int32(commentId))
	if err != nil {
		logger.Error("error while getting card comment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.COMMENT_ID_PARAM, commentId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}

	// check that the comment belong to the card
	if comment.CardID != int32(cardId) {
		logger.Error(REQUESTED_COMMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.COMMENT_ID_PARAM, commentId))
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewErrorMsg(REQUESTED_COMMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, nil))
	}

	// try to delete that one specific/requested comment
	err = ch.commentService.DeleteKanbanComment(ctx.Context(), int32(commentId))
	if err != nil {
		logger.Error("error while deleting card comment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.COMMENT_ID_PARAM, commentId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// parseAndValidateComment parses the request body into the provided Comment struct and validates its fields.
// It returns a monad.Result containing an empty struct on success, or an error result with a JSON error response
// if parsing or validation fails. Validation errors are reported with detailed field and tag information.
// The function logs errors using the global logger and sets the appropriate HTTP status code in the response.
//
// Parameters:
//   - ctx: Fiber context containing the request and response.
//   - comment: Pointer to the Comment struct to be populated and validated.
//
// Returns:
//   - monad.Result[struct{}]: Success result if parsing and validation succeed, or error result with JSON error response.
func parseAndValidateComment(ctx *fiber.Ctx, comment *Comment) monad.Result[struct{}] {
	// parse and validate the comment from the request body
	if err := ctx.BodyParser(comment); err != nil {
		logger.Error(fmt.Sprintf("error while parsing comment entity, err: %v", err),
			zap.String("type", fmt.Sprintf("%T", err)), zap.Error(err))
		e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("unable to parse comment entity", nil))
		return monad.NewResultError[struct{}](&e)
	}
	//validate the card
	if err := validate.Struct(comment); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			logger.Error(fmt.Sprintf("Unexpected validation error type, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
			e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_COMMENT_VALIDATION_ERROR_MSG, []string{"invalid input"}))
			return monad.NewResultError[struct{}](&e)
		}
		errMsgs := make([]string, len(errs))
		for idx, validationErr := range errs {
			errMsgs[idx] = fmt.Sprintf("Field '%s' failed on the '%s' tag", validationErr.Field(), validationErr.Tag())
		}
		e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_COMMENT_VALIDATION_ERROR_MSG, errMsgs))
		return monad.NewResultError[struct{}](&e)
	}
	return monad.NewResult(&struct{}{})
}
